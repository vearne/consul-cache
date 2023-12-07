package model

import (
	consulapi "github.com/hashicorp/consul/api"
)

type StateChangeMsg struct {
	DC      string `json:"dc"`
	Service string `json:"service"`
	Index   uint64 `json:"index"`
}

type ServiceState struct {
	Index uint64
	Data  []consulapi.ServiceEntry
}

type ConsulInfo struct {
	Addresses []string `mapstructure:"addresses"`
	DC        string   `mapstructure:"dc"`
	Token     string   `mapstructure:"token"`
}
