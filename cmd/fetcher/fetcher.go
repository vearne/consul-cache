package main

import (
	"flag"
	"fmt"
	"github.com/vearne/consul-cache/internal/config"
	"github.com/vearne/consul-cache/internal/consts"
	"github.com/vearne/consul-cache/internal/engine/fetcher"
	zlog "github.com/vearne/consul-cache/internal/log"
	"github.com/vearne/consul-cache/internal/resource"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
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
		fmt.Println("service: consul-fetcher")
		fmt.Println("Version", consts.Version)
		fmt.Println("BuildTime", consts.BuildTime)
		fmt.Println("GitTag", consts.GitTag)
		return
	}

	config.ReadConfig("fetcher", cfgFile)
	config.InitFetcherConfig()

	zlog.InitLogger(&config.GetFetcherOpts().Logger)
	resource.InitFetcherResource()

	cf := config.GetFetcherOpts()
	for _, consul := range cf.Consuls {
		zlog.Info("consul", zap.String("DC", consul.DC), zap.Strings("addrs", consul.Addresses))
	}

	app := wm.NewApp()
	app.AddWorker(fetcher.NewFetchWorker(cf.Consuls, cf.Services))
	app.Run()
}
