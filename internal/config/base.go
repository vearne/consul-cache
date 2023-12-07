package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var initOnce sync.Once
var gcf atomic.Value

type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"filepath"`
}

type RedisConf struct {
	Addr            string        `mapstructure:"addr"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	MaxRetries      int           `mapstructure:"maxRetries"`
	MinRetryBackoff time.Duration `mapstructure:"minRetryBackoff"`
	MaxRetryBackoff time.Duration `mapstructure:"maxRetryBackoff"`
	DialTimeout     time.Duration `mapstructure:"dialTimeout"`
	ReadTimeout     time.Duration `mapstructure:"readTimeout"`
	WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
	PoolSize        int           `mapstructure:"poolSize"`
	PoolTimeout     time.Duration `mapstructure:"poolTimeout"`
}

func ReadConfig(role string, cfgFile string) {
	viper.SetDefault("RUN_MODE", "test")
	err := viper.BindEnv("RUN_MODE")
	if err != nil {
		log.Println(err)
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("config_files")
		fname := fmt.Sprintf("config.%s.%s", role, viper.GetString("RUN_MODE"))
		viper.SetConfigName(fname)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Println("can't find config file", err)
	}
}

type RocketMQ struct {
	NameServers string              `mapstructure:"NameServers"`
	Topic       string              `mapstructure:"topic"`
	Producer    RocketMQParticipant `mapstructure:"producer"`
	Consumer    RocketMQParticipant `mapstructure:"consumer"`
}

type RocketMQParticipant struct {
	Token   string `mapstructure:"token"`
	GroupID string `mapstructure:"groupId"`
}
