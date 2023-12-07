package informer

import (
	"github.com/fatih/structs"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/model"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

type ServiceWatcher struct {
	ServiceName  string
	DC           string
	Token        string
	Plan         *watch.Plan // 每个service的单独监控
	ServiceState *State      // 最新状态
	ConsulAddr   string
}

func NewServiceWatcher(service string, c *model.ConsulInfo, ch chan<- *StateChange) (*ServiceWatcher, error) {
	zlog.Info("NewServiceWatcher", zap.String("dc", c.DC), zap.String("service", service))

	var w ServiceWatcher
	var err error
	w.ServiceName = service
	w.DC = c.DC
	// 初始状态
	w.ServiceState = &State{ServiceEntrys: make([]*consulapi.ServiceEntry, 0), T: time.Now(), Index: 0}

	N := len(c.Addresses)
	w.ConsulAddr = c.Addresses[rand.Intn(N)]

	param := PlanParam{
		Type:        "service",
		Service:     service,
		PassingOnly: true,
		DC:          w.DC,
		Token:       c.Token,
		Stale:       false,
	}

	w.Plan, err = watch.Parse(structs.Map(&param))
	if err != nil {
		return nil, err
	}

	// FIXME 如果ServiceState被修改，但实际上访问Redis或者MQ失败，可能会导致最新的State永远无法被推送
	w.Plan.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		case []*consulapi.ServiceEntry:
			// 配置发生了变更
			if idx != w.ServiceState.Index {
				newState := State{ServiceEntrys: d, T: time.Now(), Index: idx}
				ch <- &StateChange{
					NewState:  newState,
					DC:        w.DC,
					Service:   w.ServiceName,
					LastIndex: w.ServiceState.Index,
				}
				// 修改当前状态为最新状态
				w.ServiceState = &newState
			}
		default:
			zlog.Error("unknown data type,", zap.Any("data", data))
		}
	}

	return &w, nil
}

func (w *ServiceWatcher) Run() error {
	zlog.Info("ServiceWatcher, Run()", zap.String("dc", w.DC), zap.String("service", w.ServiceName))
	err := w.Plan.Run(w.ConsulAddr)
	if err != nil {
		zlog.Error("ServiceWatcher", zap.String("dc", w.DC),
			zap.String("service", w.ServiceName), zap.Error(err))
		return err
	}
	return nil
}

func (w *ServiceWatcher) Stop() {
	zlog.Info("ServiceWatcher, Stop()", zap.String("dc", w.DC), zap.String("service", w.ServiceName))
	w.Plan.Stop()
}

type PlanParam struct {
	Type        string   `structs:"type"`
	Service     string   `structs:"service"`
	PassingOnly bool     `structs:"passingonly"`
	DC          string   `structs:"datacenter,omitempty"`
	Token       string   `structs:"token,omitempty"`
	Tag         []string `structs:"tag,omitempty"`
	Stale       bool     `structs:"stale,omitempty"`
}
