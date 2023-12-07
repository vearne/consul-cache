package informer

import (
	consulapi "github.com/hashicorp/consul/api"
	"time"
)

type StateChange struct {
	NewState State
	DC       string
	Service  string
	// 上次Index值
	LastIndex uint64
}

type State struct {
	ServiceEntrys []*consulapi.ServiceEntry
	T             time.Time
	Index         uint64
}
