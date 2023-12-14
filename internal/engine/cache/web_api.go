package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/vearne/consul-cache/internal/biz"
	"github.com/vearne/consul-cache/internal/config"
	"github.com/vearne/consul-cache/internal/consts"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/mappool"
	"github.com/vearne/consul-cache/internal/middleware"
	"github.com/vearne/consul-cache/internal/model"
	"github.com/vearne/consul-cache/internal/resource"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type WebAPIServer struct {
	Server *http.Server
}

func NewWebServer(ginHandler *gin.Engine) *WebAPIServer {
	zlog.Info("[init]WebServer")
	worker := &WebAPIServer{}
	cacheConfig := config.GetConsulCacheOpts()
	worker.Server = &http.Server{
		Addr:           cacheConfig.Web.ListenAddress,
		Handler:        ginHandler,
		ReadTimeout:    10 * time.Minute,
		WriteTimeout:   10 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}
	return worker
}

func (worker *WebAPIServer) Start() {
	zlog.Info("[start]WebAPIServer")
	err := worker.Server.ListenAndServe()
	if err != nil {
		zlog.Error("worker.Server.ListenAndServe", zap.Error(err))
	}
}

func (worker *WebAPIServer) Stop() {
	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := worker.Server.Shutdown(cxt)
	if err != nil {
		zlog.Error("shutdown error", zap.Error(err))
	}
	zlog.Info("[end]WebAPIServer exit")
}

func NewRouter() *gin.Engine {
	r := gin.Default()
	g := r.Group("/v1")
	g.Use(middleware.Metric())
	g.Use(middleware.GlobalRateLimit())
	g.Use(middleware.ConcurrentReq())

	// /v1/health/service/tv-proxy?dc=baoding-dingxing&passing=true&stale&index=1000&wait=5s
	g.GET("/health/service/:svcName", healthService)

	r.GET("/version", func(c *gin.Context) {
		data := map[string]string{}
		data["version"] = consts.Version
		data["buildTime"] = consts.BuildTime
		data["gitTag"] = consts.GitTag
		data["upTime"] = consts.UpTime
		c.JSON(http.StatusOK, data)
	})

	return r
}

type HSParam struct {
	SvcName string
	DC      string
	Index   uint64
	Wait    time.Duration
	Stale   bool
	Tags    []string
}

func parseHSParam(c *gin.Context) (*HSParam, error) {
	// 实际只需要处理svcName、dc、index、wait 3个参数
	// wait 如果不传，默认是5分钟
	// index 如果不传，默认是0
	var ok bool
	var param HSParam
	param.Stale = false

	param.SvcName = c.Param("svcName")
	param.DC, ok = c.GetQuery("dc")
	if !ok {
		return nil, errors.New("parameter [dc] is missing")
	}
	idxStr := c.DefaultQuery("index", "0")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		return nil, errors.New("invalid index time")
	}
	param.Index = uint64(idx)

	waitStr := c.DefaultQuery("wait", "5m")
	d, err := time.ParseDuration(waitStr)
	if err != nil {
		return nil, errors.New("invalid wait time")
	}
	param.Wait = d
	_, ok = c.GetQuery("stale")
	if ok {
		param.Stale = true
	}

	param.Tags = c.QueryArray("tag")
	return &param, nil
}

func healthService(c *gin.Context) {
	param, err := parseHSParam(c)
	if err != nil {
		zlog.Error("healthService, parseHSParam", zap.Error(err))
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	zlog.Debug("healthService",
		zap.String("dc", param.DC),
		zap.String("service", param.SvcName),
		zap.Uint64("index", param.Index),
		zap.String("wait", param.Wait.String()),
		zap.Bool("stale", param.Stale),
	)
	defaultExpiration := config.GetConsulCacheOpts().Cache.Expiration + time.Second*time.Duration(rand.Intn(60))
	var state *model.ServiceState

	// 1.如果localCache中有数据，那么假定localcache所拥有的就是最新的数据
	key := fmt.Sprintf(consts.StatekeyFormat, param.DC, param.SvcName)
	val, ok := resource.SeviceStateCache.Get(key)
	if ok {
		zlog.Debug("SeviceStateCache.Get, cache hit")
		state = val.(*model.ServiceState)
	} else {
		zlog.Debug("SeviceStateCache.Get, cache miss")
		// 尝试从Redis加载一次
		state, err = biz.ReloadFromRedis(param.DC, param.SvcName, defaultExpiration)
		if err != nil {
			zlog.Error("biz.ReloadFromRedis", zap.Error(err))
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// 2. 比较index值
	// 2.1 直接返回
	if param.Index < state.Index {
		masquerading(c, state.Index, param.Stale)
		c.JSON(http.StatusOK, filterWithTag(state.Data, param.Tags))
		return
	}

	// 2.2 进行等待
	//  或者wait超时
	//  或者change event到达
	timer := time.NewTimer(param.Wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		zlog.Debug("wait timeout")
	case <-resource.ServiceStateWait.GetOrCreateItem(param.DC, param.SvcName):
		zlog.Debug("change event comming",
			zap.String("dc", param.DC), zap.String("service", param.SvcName))
	}

	val, ok = resource.SeviceStateCache.Get(key)
	if ok {
		zlog.Debug("SeviceStateCache.Get, cache hit")
		state = val.(*model.ServiceState)
	} else {
		zlog.Debug("SeviceStateCache.Get, cache miss")
		// 尝试从Redis加载一次
		state, err = biz.ReloadFromRedis(param.DC, param.SvcName, defaultExpiration)
		if err != nil {
			zlog.Error("biz.ReloadFromRedis", zap.Error(err))
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	masquerading(c, state.Index, param.Stale)
	c.JSON(http.StatusOK, filterWithTag(state.Data, param.Tags))
}

func filterWithTag(entrys []consulapi.ServiceEntry, tags []string) []consulapi.ServiceEntry {
	if len(tags) <= 0 {
		return entrys
	}
	result := make([]consulapi.ServiceEntry, 0)
	for _, entry := range entrys {
		if entryContainTags(entry.Service.Tags, tags) {
			result = append(result, entry)
		}
	}

	return result
}

func entryContainTags(entryTags, tags []string) bool {
	tagMap := mappool.Get()
	defer mappool.Put(tagMap)

	if len(entryTags) < len(tags) {
		return false
	}

	for _, etag := range entryTags {
		tagMap[etag] = struct{}{}
	}

	for _, tag := range tags {
		_, ok := tagMap[tag]
		if !ok {
			return false
		}
	}

	return true
}

func masquerading(c *gin.Context, index uint64, stale bool) {
	if stale {
		c.Header("X-Consul-Effective-Consistency", "stale")
	} else {
		c.Header("X-Consul-Effective-Consistency", "leader")
	}

	c.Header("X-Consul-Index", strconv.Itoa(int(index)))
	c.Header("X-Consul-Knownleader", "true")
	c.Header("X-Consul-Lastcontact", "5")
}
