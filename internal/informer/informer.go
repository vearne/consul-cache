package informer

import (
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/model"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Informer struct {
	consuls []model.ConsulInfo
	// 关注那些service
	serviceNames []string

	group    errgroup.Group
	watchers []*ServiceWatcher
}

func NewInformer(consuls []model.ConsulInfo, serviceNames []string) *Informer {
	var informer Informer
	informer.consuls = consuls
	informer.serviceNames = serviceNames
	informer.watchers = make([]*ServiceWatcher, 0)
	return &informer
}

func (informer *Informer) Watch() chan *StateChange {
	ch := make(chan *StateChange, 100)
	for idx := range informer.consuls {
		c := informer.consuls[idx]
		for _, service := range informer.serviceNames {
			watcher, err := NewServiceWatcher(service, &c, ch)
			if err != nil {
				zlog.Error("create ServiceWatcher", zap.String("dc", c.DC),
					zap.String("service", service), zap.Error(err))
				continue
			}

			informer.watchers = append(informer.watchers, watcher)
			informer.group.Go(func() error {
				return watcher.Run()
			})
		}
	}

	return ch
}

func (informer *Informer) Stop() {
	for _, watcher := range informer.watchers {
		watcher.Stop()
	}
	err := informer.group.Wait()
	zlog.Error("Informer.Stop()", zap.Error(err))
}
