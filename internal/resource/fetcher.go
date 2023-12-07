package resource

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/redis/go-redis/v9"
	"github.com/vearne/consul-cache/internal/config"
	zlog "github.com/vearne/consul-cache/internal/log"
)

var (
	RedisConsulClient redis.Cmdable

	StateTopic      string
	StateMQProducer rocketmq.Producer
)

func InitFetcherResource() {
	zlog.Info("InitFetchRedis")
	initFetchRedis()
	zlog.Info("initRocketMQProducer")
	initRocketMQProducer()

}

func initFetchRedis() {
	RedisConsulClient = InitSingleRedis(config.GetFetcherOpts().RedisConsul, "redis-consul")
}

func initRocketMQProducer() {
	mq := config.GetFetcherOpts().StateMQ
	StateTopic = mq.Topic
	StateMQProducer = initMQProducer(mq.Producer.GroupID, mq.NameServers)
}
