package config

import (
	"github.com/spf13/viper"
	"github.com/vearne/consul-cache/internal/model"
	"log"
)

type FetcherConfig struct {
	Logger      LogConfig          `mapstructure:"logger"`
	Consuls     []model.ConsulInfo `mapstructure:"consuls"`
	Services    []string           `mapstructure:"services"`
	RedisConsul RedisConf          `mapstructure:"redis-consul"`

	StateMQ RocketMQ `mapstructure:"rocketmq-state"`
}

func InitFetcherConfig() {
	log.Println("---InitFetcherConfig---")
	initOnce.Do(func() {
		var cf = FetcherConfig{}
		err := viper.Unmarshal(&cf)
		if err != nil {
			log.Fatalf("InitFetcherConfig:%v \n", err)
		}
		gcf.Store(&cf)
	})
}

func GetFetcherOpts() *FetcherConfig {
	return gcf.Load().(*FetcherConfig)
}
