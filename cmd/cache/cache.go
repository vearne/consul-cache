package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vearne/consul-cache/internal/config"
	"github.com/vearne/consul-cache/internal/consts"
	cc "github.com/vearne/consul-cache/internal/engine/cache"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/resource"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
)

var (
	// config file path
	cfgFile     string
	versionFlag bool
)

func init() {
	flag.StringVar(&cfgFile, "config", "", "config file")
	flag.BoolVar(&versionFlag, "version", false, "Show version")
}

func main() {
	flag.Parse()

	if versionFlag {
		fmt.Println("service: consul-cache")
		fmt.Println("Version", consts.Version)
		fmt.Println("BuildTime", consts.BuildTime)
		fmt.Println("GitTag", consts.GitTag)
		return
	}

	config.ReadConfig("cache", cfgFile)
	config.InitConsulCacheConfig()

	zlog.InitLogger(&config.GetConsulCacheOpts().Logger)
	resource.InitConsulCacheResource()

	// pprof
	go func() {
		err := http.ListenAndServe(config.GetConsulCacheOpts().Web.PprofAddress, nil)
		zlog.Error("pprof server", zap.Error(err))
	}()

	// 添加Prometheus的相关监控
	// /metrics
	go func() {
		r := gin.Default()
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
		err := r.Run(config.GetConsulCacheOpts().Web.PrometheusAddress)
		zlog.Error("metrics server", zap.Error(err))
	}()

	app := wm.NewApp()
	app.AddWorker(cc.NewMsgWorker())
	app.AddWorker(cc.NewWebServer(cc.NewRouter()))
	app.Run()
}
