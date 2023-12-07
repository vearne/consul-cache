package cache

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	jsoniter "github.com/json-iterator/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vearne/consul-cache/internal/biz"
	"github.com/vearne/consul-cache/internal/config"
	"github.com/vearne/consul-cache/internal/consts"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/model"
	"github.com/vearne/consul-cache/internal/resource"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const (
	PromLabelSVC = "service"
	PromLabelDC  = "dc"
)

type MsgWorker struct {
	ExitedFlag chan struct{} // 已经退出的标识
	ExitChan   chan struct{}
}

func NewMsgWorker() *MsgWorker {
	var worker MsgWorker
	worker.ExitedFlag = make(chan struct{})
	worker.ExitChan = make(chan struct{})
	return &worker
}

func (w *MsgWorker) Start() {
	err := resource.StateMQConsumer.Subscribe(resource.StateTopic,
		consumer.MessageSelector{Type: consumer.TAG, Expression: "*"},
		DealMessage,
	)
	if err != nil {
		zlog.Error("MsgWorker, StateMQConsumer Subscribe", zap.Error(err))
	}
	err = resource.StateMQConsumer.Start()
	zlog.Info("MsgWorker, StateMQConsumer start")
	if err != nil {
		zlog.Fatal("MsgWorker, StateMQConsumer start", zap.Error(err))
	}

	// ----wait------
	<-w.ExitChan
	zlog.Info("got exit signal from ExitChan")

	err = resource.StateMQConsumer.Shutdown()
	if err != nil {
		zlog.Fatal("MsgWorker, StateMQConsumer shutdown", zap.Error(err))
	}

	close(w.ExitedFlag)
	zlog.Info("MsgWorker exit")
}

func (w *MsgWorker) Stop() {
	close(w.ExitChan)

	<-w.ExitedFlag
	zlog.Info("[end]MsgWorker")
}

func DealMessage(ctx context.Context,
	msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	// 处理消息, 每次只取回一条消息
	if len(msgs) > 1 || len(msgs) < 1 {
		panic("inpossible to happen.")
	}
	msg := msgs[0]
	var c model.StateChangeMsg
	err := jsoniter.Unmarshal(msg.Body, &c)
	if err != nil {
		zlog.Error("jsoniter.Unmarshal", zap.String("body", string(msg.Body)), zap.Error(err))
		return consumer.ConsumeSuccess, nil
	}

	resource.PushMsgTotal.With(prometheus.Labels{
		PromLabelDC:  c.DC,
		PromLabelSVC: c.Service,
	}).Inc()

	key := fmt.Sprintf(consts.StatekeyFormat, c.DC, c.Service)
	val, ok := resource.SeviceStateCache.Get(key)
	if !ok {
		return consumer.ConsumeSuccess, nil
	}

	state := val.(*model.ServiceState)
	if state.Index == c.Index {
		return consumer.ConsumeSuccess, nil
	}

	defaultExpiration := config.GetConsulCacheOpts().Cache.Expiration + time.Second*time.Duration(rand.Intn(60))
	state, err = biz.ReloadFromRedis(c.DC, c.Service, defaultExpiration)
	if err != nil {
		zlog.Error("biz.ReloadFromRedis",
			zap.String("dc", c.DC),
			zap.String("service", c.Service),
			zap.Error(err))
		return consumer.ConsumeRetryLater, nil
	}
	zlog.Info("ReloadFromRedis", zap.String("dc", c.DC),
		zap.String("service", c.Service), zap.Uint64("index", state.Index))

	// notify block query
	resource.ServiceStateWait.Notify(c.DC, c.Service)
	return consumer.ConsumeSuccess, nil
}
