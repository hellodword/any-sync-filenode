package config

import (
	commonaccount "github.com/anytypeio/go-anytype-infrastructure-experiments/common/accountservice"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/app"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/app/logger"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/metric"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/net"
	"gopkg.in/yaml.v3"
	"os"
)

const CName = "config"

func NewFromFile(path string) (c *Config, err error) {
	c = &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return
}

type Config struct {
	GrpcServer net.Config           `yaml:"grpcServer"`
	Account    commonaccount.Config `yaml:"account"`
	Mongo      Mongo                `yaml:"mongo"`
	Metric     metric.Config        `yaml:"metric"`
	Log        logger.Config        `yaml:"log"`
}

func (c *Config) Init(a *app.App) (err error) {
	return
}

func (c Config) Name() (name string) {
	return CName
}

func (c Config) GetMongo() Mongo {
	return c.Mongo
}

func (c Config) GetNet() net.Config {
	return c.GrpcServer
}

func (c Config) GetAccount() commonaccount.Config {
	return c.Account
}

func (c Config) GetMetric() metric.Config {
	return c.Metric
}
