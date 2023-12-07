package config

import (
	"github.com/spf13/viper"
	"log"
	"time"
)

type ConsulCacheConfig struct {
	Logger LogConfig `mapstructure:"logger"`

	RedisConsul RedisConf `mapstructure:"redis-consul"`

	StateMQ RocketMQ `mapstructure:"rocketmq-state"`

	Cache struct {
		Expiration time.Duration `mapstructure:"expiration"`
	} `mapstructure:"cache"`

	Web struct {
		Mode              string `mapstructure:"mode"`
		MiddlewareEnabled bool   `mapstructure:"middleware_enabled"`
		ListenAddress     string `mapstructure:"listen_address"`
		PprofAddress      string `mapstructure:"pprof_address"`
		PrometheusAddress string `mapstructure:"prometheus_address"`
	} `mapstructure:"web"`
}

func InitConsulCacheConfig() {
	log.Println("---InitConsulCacheConfig---")
	initOnce.Do(func() {
		var cf = ConsulCacheConfig{}
		err := viper.Unmarshal(&cf)
		if err != nil {
			log.Fatalf("InitConsulCacheConfig:%v \n", err)
		}
		gcf.Store(&cf)
	})
}

func GetConsulCacheOpts() *ConsulCacheConfig {
	return gcf.Load().(*ConsulCacheConfig)
}
