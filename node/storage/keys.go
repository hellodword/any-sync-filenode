package storage

import (
	"fmt"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/pkg/acl/storage"
	"strings"
)

type aclKeys struct {
}

var aclHeadIdKey = []byte("a/headId")
var aclRootIdKey = []byte("a/rootId")

func (a aclKeys) HeadIdKey() []byte {
	return aclHeadIdKey
}

func (a aclKeys) RootIdKey() []byte {
	return aclRootIdKey
}

func (a aclKeys) RawRecordKey(id string) []byte {
	return storage.JoinStringsToBytes("a", id)
}

type treeKeys struct {
	id       string
	prefix   string
	headsKey []byte
}

func newTreeKeys(id string) treeKeys {
	return treeKeys{
		id:       id,
		headsKey: storage.JoinStringsToBytes("t", id, "heads"),
		prefix:   fmt.Sprintf("t/%s", id),
	}
}

func (t treeKeys) HeadsKey() []byte {
	return t.headsKey
}

func (t treeKeys) RawChangeKey(id string) []byte {
	return storage.JoinStringsToBytes("t", t.id, id)
}

func (t treeKeys) isTreeRecordKey(key string) bool {
	return strings.HasPrefix(key, t.prefix)
}

type spaceKeys struct {
	headerKey []byte
}

func newSpaceKeys(spaceId string) spaceKeys {
	return spaceKeys{headerKey: storage.JoinStringsToBytes("s", spaceId)}
}

var (
	spaceIdKey         = []byte("spaceId")
	spaceSettingsIdKey = []byte("spaceSettingsId")
)

func (s spaceKeys) SpaceIdKey() []byte {
	return spaceIdKey
}

func (s spaceKeys) HeaderKey() []byte {
	return s.headerKey
}

func (s spaceKeys) SpaceSettingsIdKey() []byte {
	return spaceSettingsIdKey
}

func isTreeHeadsKey(key string) bool {
	return strings.HasPrefix(key, "t/") && strings.HasSuffix(key, "/heads")
}

func getRootId(key string) string {
	prefixLen := 2 // len("t/")
	suffixLen := 6 // len("/heads")
	rootLen := len(key) - suffixLen - prefixLen
	sBuf := strings.Builder{}
	sBuf.Grow(rootLen)
	for i := prefixLen; i < prefixLen+rootLen; i++ {
		sBuf.WriteByte(key[i])
	}
	return sBuf.String()
}
