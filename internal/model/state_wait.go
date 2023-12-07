package model

import (
	"fmt"
	"github.com/vearne/consul-cache/internal/consts"
	"sync"
)

type StateWait struct {
	m sync.RWMutex
	// {dc}:{svcName} -> chan struct{}
	inner map[string]chan struct{}
}

func NewStateWait() *StateWait {
	var s StateWait
	s.inner = make(map[string]chan struct{})
	return &s
}

func (sw *StateWait) GetOrCreateItem(dc, svcName string) chan struct{} {
	key := fmt.Sprintf(consts.StatekeyFormat, dc, svcName)

	sw.m.RLock()
	ch, ok := sw.inner[key]
	sw.m.RUnlock()
	if ok {
		return ch
	}

	var result chan struct{}

	sw.m.Lock()
	defer sw.m.Unlock()
	// 防止并发
	if result, ok = sw.inner[key]; ok {
		return result
	}
	sw.inner[key] = make(chan struct{})
	return sw.inner[key]
}

func (sw *StateWait) Notify(dc, svcName string) {
	key := fmt.Sprintf(consts.StatekeyFormat, dc, svcName)

	sw.m.Lock()
	defer sw.m.Unlock()
	// 防止并发
	if ch, ok := sw.inner[key]; ok {
		close(ch)
	}
	sw.inner[key] = make(chan struct{})
}
