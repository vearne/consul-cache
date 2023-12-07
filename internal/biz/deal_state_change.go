package biz

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	jsoniter "github.com/json-iterator/go"
	"github.com/vearne/consul-cache/internal/consts"
	"github.com/vearne/consul-cache/internal/informer"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/model"
	"github.com/vearne/consul-cache/internal/resource"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func DealStateChange(sc *informer.StateChange) error {
	// 1) 写redis data
	ctx := context.Background()
	key := fmt.Sprintf(consts.DatakeyFormat, sc.DC, sc.Service, sc.NewState.Index)
	value, err := jsoniter.MarshalToString(sc.NewState.ServiceEntrys)
	if err != nil {
		zlog.Error("jsoniter.MarshalToString", zap.Error(err))
		return err
	}
	_, err = resource.RedisConsulClient.Set(ctx, key, value, 0).Result()
	if err != nil {
		zlog.Error("RedisConsulClient.Set", zap.Error(err))
		return err
	}
	// 2) 写redis index
	key = fmt.Sprintf(consts.IndexkeyFormat, sc.DC, sc.Service)
	_, err = resource.RedisConsulClient.Set(ctx, key,
		strconv.Itoa(int(sc.NewState.Index)),
		0).Result()
	if err != nil {
		zlog.Error("RedisConsulClient.Set", zap.Error(err))
		return err
	}
	// 3) 向MQ发送一条消息
	buf, _ := jsoniter.Marshal(model.StateChangeMsg{
		DC:      sc.DC,
		Service: sc.Service,
		Index:   sc.NewState.Index,
	})
	msg := primitive.NewMessage(resource.StateTopic, buf)
	result, err := resource.StateMQProducer.SendSync(ctx, msg)
	zlog.Info("send message", zap.String("msg", string(buf)),
		zap.Int("status", int(result.Status)), zap.String("msgId", result.MsgID))
	if err != nil {
		zlog.Error("send message", zap.String("msg", string(buf)), zap.Error(err))
		return err
	}
	// 4) redis清理上一个版本
	time.AfterFunc(time.Second*5, func() {
		key = fmt.Sprintf(consts.DatakeyFormat, sc.DC, sc.Service, sc.LastIndex)
		_, err = resource.RedisConsulClient.Del(ctx, key).Result()
		if err != nil {
			zlog.Error("RedisConsulClient.Del", zap.String("key", key), zap.Error(err))
		}
	})
	return nil
}
