package fetcher

import (
	"github.com/vearne/consul-cache/internal/biz"
	"github.com/vearne/consul-cache/internal/informer"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/model"
	"go.uber.org/zap"
)

type FetchWorker struct {
	// serviceName -> index
	//svcIndex        map[string]uint64
	svcInfomer      *informer.Informer
	stateChangeChan chan *informer.StateChange
}

func NewFetchWorker(consuls []model.ConsulInfo,
	serviceNames []string) *FetchWorker {
	var worker FetchWorker
	//worker.svcIndex = make(map[string]uint64)
	worker.svcInfomer = informer.NewInformer(consuls, serviceNames)
	return &worker
}

func (w *FetchWorker) Start() {
	zlog.Info("[start]FetchWorker")
	w.stateChangeChan = w.svcInfomer.Watch()

	for sc := range w.stateChangeChan {
		zlog.Info("stateChange",
			zap.String("dc", sc.DC),
			zap.String("service", sc.Service),
			zap.Uint64("LastIndex", sc.LastIndex),
			zap.Uint64("Index", sc.NewState.Index),
			zap.Int("len(ServiceEntrys)", len(sc.NewState.ServiceEntrys)),
		)
		err := biz.DealStateChange(sc)
		if err != nil {
			zlog.Error("deal stateChange error", zap.Error(err))
			// TODO 推送到重试队列
		}
	}
}

func (w *FetchWorker) Stop() {
	zlog.Info("[stop]FetchWorker")
	w.svcInfomer.Stop()
	close(w.stateChangeChan)
}
