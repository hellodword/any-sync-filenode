package acltree

import (
	"github.com/anytypeio/go-anytype-infrastructure-experiments/pkg/acl/account"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/pkg/acl/aclchanges/aclpb"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/util/cid"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/util/keys/asymmetric/encryptionkey"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/util/keys/symmetric"
	"github.com/gogo/protobuf/proto"
	"hash/fnv"
	"time"
)

type MarshalledChange = []byte

type ACLChangeBuilder interface {
	UserAdd(identity string, encryptionKey encryptionkey.PubKey, permissions aclpb.ACLChange_UserPermissions) error
	AddId(id string)      // TODO: this is only for testing
	SetMakeSnapshot(bool) // TODO: who should decide this? probably ACLTree so we can delete it
}

type ChangeBuilder interface {
	ACLChangeBuilder
	AddChangeContent(marshaler proto.Marshaler) // user code should be responsible for making regular snapshots
}

type changeBuilder struct {
	aclState *ACLState
	tree     *Tree
	acc      *account.AccountData

	aclData       *aclpb.ACLChange_ACLData
	changeContent proto.Marshaler
	id            string
	makeSnapshot  bool
	readKey       *symmetric.Key
	readKeyHash   uint64
}

func newChangeBuilder() *changeBuilder {
	return &changeBuilder{}
}

func (c *changeBuilder) Init(state *ACLState, tree *Tree, acc *account.AccountData) {
	c.aclState = state
	c.tree = tree
	c.acc = acc

	c.aclData = &aclpb.ACLChange_ACLData{}
	// setting read key for further encryption etc
	if state.currentReadKeyHash == 0 {
		c.readKey, _ = symmetric.NewRandom()

		hasher := fnv.New64()
		hasher.Write(c.readKey.Bytes())
		c.readKeyHash = hasher.Sum64()
	} else {
		c.readKey = c.aclState.userReadKeys[c.aclState.currentReadKeyHash]
		c.readKeyHash = c.aclState.currentReadKeyHash
	}
}

func (c *changeBuilder) AddId(id string) {
	c.id = id
}

func (c *changeBuilder) SetMakeSnapshot(b bool) {
	c.makeSnapshot = b
}

func (c *changeBuilder) UserAdd(identity string, encryptionKey encryptionkey.PubKey, permissions aclpb.ACLChange_UserPermissions) error {
	var allKeys []*symmetric.Key
	if c.aclState.currentReadKeyHash != 0 {
		for _, key := range c.aclState.userReadKeys {
			allKeys = append(allKeys, key)
		}
	} else {
		allKeys = append(allKeys, c.readKey)
	}

	var encryptedKeys [][]byte
	for _, k := range allKeys {
		res, err := encryptionKey.Encrypt(k.Bytes())
		if err != nil {
			return err
		}

		encryptedKeys = append(encryptedKeys, res)
	}
	rawKey, err := encryptionKey.Raw()
	if err != nil {
		return err
	}
	ch := &aclpb.ACLChange_ACLContentValue{
		Value: &aclpb.ACLChange_ACLContent_Value_UserAdd{
			UserAdd: &aclpb.ACLChange_UserAdd{
				Identity:          identity,
				EncryptionKey:     rawKey,
				EncryptedReadKeys: encryptedKeys,
				Permissions:       permissions,
			},
		},
	}
	c.aclData.AclContent = append(c.aclData.AclContent, ch)
	return nil
}

func (c *changeBuilder) BuildAndApply() (*Change, []byte, error) {
	aclChange := &aclpb.ACLChange{
		TreeHeadIds:        c.tree.Heads(),
		AclHeadIds:         c.tree.ACLHeads(),
		SnapshotBaseId:     c.tree.RootId(),
		AclData:            c.aclData,
		CurrentReadKeyHash: c.readKeyHash,
		Timestamp:          int64(time.Now().Nanosecond()),
		Identity:           c.acc.Identity,
	}
	err := c.aclState.applyChange(aclChange)
	if err != nil {
		return nil, nil, err
	}

	if c.makeSnapshot {
		c.aclData.AclSnapshot = c.aclState.makeSnapshot()
	}

	var marshalled []byte
	if c.changeContent != nil {
		marshalled, err = c.changeContent.Marshal()
		if err != nil {
			return nil, nil, err
		}

		encrypted, err := c.aclState.userReadKeys[c.aclState.currentReadKeyHash].
			Encrypt(marshalled)
		if err != nil {
			return nil, nil, err
		}
		aclChange.ChangesData = encrypted
	}

	fullMarshalledChange, err := proto.Marshal(aclChange)
	if err != nil {
		return nil, nil, err
	}
	signature, err := c.acc.SignKey.Sign(fullMarshalledChange)
	if err != nil {
		return nil, nil, err
	}
	id, err := cid.NewCIDFromBytes(fullMarshalledChange)
	if err != nil {
		return nil, nil, err
	}
	ch := NewChange(id, aclChange)
	ch.DecryptedDocumentChange = marshalled
	ch.Sign = signature

	return ch, fullMarshalledChange, nil
}

func (c *changeBuilder) AddChangeContent(marshaler proto.Marshaler) {
	c.changeContent = marshaler
}
