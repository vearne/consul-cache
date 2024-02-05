package resource

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/redis/go-redis/v9"
	"github.com/vearne/consul-cache/internal/config"
	"github.com/vearne/consul-cache/internal/coolcache"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/model"
	"time"
)

var (
	// {dc}:{service} -> *ServiceState
	SeviceStateCache   *coolcache.CoolCache
	StateReadOnlyRedis redis.Cmdable
	StateMQConsumer    rocketmq.PushConsumer

	// 用于控制block query
	ServiceStateWait *model.StateWait

	// 并发执行的请求数
	ConcurrentReq int64
)

func InitConsulCacheResource() {
	zlog.Info("initServiceCache")
	initCache()
	zlog.Info("InitFetchRedis")
	initStateRedis()
	zlog.Info("initRocketMQConsumer")
	initRocketMQConsumer()
	zlog.Info("initStateWait")
	initStateWait()
	zlog.Info("initPromtheus")
	initPromtheus()
	zlog.Info("initRateLimiter")
	initRateLimiter()
}

func initCache() {
	SeviceStateCache = coolcache.NewCoolCache(100, 5*time.Minute, 10*time.Minute)
}

func initStateRedis() {
	StateReadOnlyRedis = InitSingleRedis(config.GetConsulCacheOpts().RedisConsul, "redis-consul")
}

func initRocketMQConsumer() {
	mq := config.GetConsulCacheOpts().StateMQ
	StateTopic = mq.Topic
	StateMQConsumer = initMQConsumer(mq.Consumer.GroupID, mq.NameServers)
}

func initStateWait() {
	ServiceStateWait = model.NewStateWait()
}
