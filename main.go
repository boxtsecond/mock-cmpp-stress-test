package main

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/client"
	"mock-cmpp-stress-test/cmpp/server"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/statistics"
	"mock-cmpp-stress-test/stress_test_service"
	"mock-cmpp-stress-test/utils/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Service interface {
	Init(log *zap.Logger)
	Start() error
	Stop() error
}

var Services = []Service{
	// 收集数据服务
	new(statistics.Collection),
	// CMPP 服务端
	new(server.CmppServer),
	// CMPP 客户端
	new(client.CmppClient),
	// 压测服务
	new(stress_test_service.StressTest),
}

func Init() error {
	// init log
	if err := config.Init(); err != nil {
		return errors.New(fmt.Sprintf("Load Config Error: %s", err.Error()))
	}
	log.Init(config.ConfigObj.Log)

	return nil
}

func Start() error {
	for _, service := range Services {
		service.Init(log.Logger)
		if err := service.Start(); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func Stop() error {
	for i, service := range Services {
		if err := service.Stop(); err != nil {
			log.Logger.Panic("Stop Failed.",
				zap.Int("Index", i),
				zap.Error(err))
		}
		time.Sleep(time.Second)
	}
	return nil
}

func main() {
	if err := Init(); err != nil {
		log.Logger.Panic("Init Failed.", zap.Error(err))
		return
	}

	if err := Start(); err != nil {
		log.Logger.Panic("Start Failed.", zap.Error(err))
		return
	}

	log.Logger.Info("Start Success.")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit

	log.Logger.Info("Got Signal. Exit.")
	if err := Stop(); err != nil {
		log.Logger.Panic("Stop Error.", zap.Error(err))
		return
	}
}
