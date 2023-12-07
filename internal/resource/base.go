package resource

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/vearne/consul-cache/internal/config"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/golib/metric"
	"go.uber.org/zap"
	"math/rand"
	"net"
	"strings"
	"time"
)

func initRedis(conf config.RedisConf) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:            conf.Addr,
		Password:        conf.Password,
		DB:              conf.DB,
		MaxRetries:      conf.MaxRetries,
		MinRetryBackoff: conf.MinRetryBackoff,
		MaxRetryBackoff: conf.MaxRetryBackoff,
		ReadTimeout:     conf.ReadTimeout,
		WriteTimeout:    conf.WriteTimeout,
		PoolSize:        conf.PoolSize,
		PoolTimeout:     conf.PoolTimeout,
		/*
		   咱们公司内部使用的形如: xxx.r.bjdx.qiyi.redis 为高可用域名
		   域名解析在进程内部有缓存，因此需要使用Dialer参数
		*/
		// Dialer creates new network connection and has priority over Network and Addr options.
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			r := &net.Resolver{}
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := r.LookupHost(context.Background(), host)
			if err != nil {
				return nil, err
			}
			ip := ips[rand.Intn(len(ips))]
			netDialer := &net.Dialer{
				Timeout:   conf.DialTimeout,
				KeepAlive: 1 * time.Minute,
			}
			return netDialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
		},
	})
}

func InitSingleRedis(conf config.RedisConf, role string) redis.Cmdable {
	var err error
	client := initRedis(conf)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		zlog.Fatal(fmt.Sprintf("initialize %v error", role), zap.Error(err))
	}
	metric.AddRedis(client, role)
	return client
}

func initMQProducer(groupName string, nameServers string) rocketmq.Producer {
	var err error
	p, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(strings.Split(nameServers, ";"))),
		producer.WithInstanceName(uuid.Must(uuid.NewUUID()).String()),
		producer.WithRetry(2),
		producer.WithGroupName(groupName),
	)
	if err != nil {
		zlog.Fatal("initialize rocketmq error1", zap.Error(err))
	}

	err = p.Start()
	if err != nil {
		zlog.Fatal("initialize rocketmq error2", zap.Error(err))
	}
	return p
}

func initMQConsumer(groupName string, nameServers string) rocketmq.PushConsumer {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(groupName),
		// 指定使用广播模式
		consumer.WithConsumerModel(consumer.BroadCasting),
		consumer.WithInstance(uuid.Must(uuid.NewUUID()).String()),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(strings.Split(nameServers, ";"))),
		// 每次只取回一条数据
		consumer.WithPullBatchSize(1),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
	)
	if err != nil {
		zlog.Fatal("initialize rocketmq error", zap.Error(err))
	}
	return c
}
