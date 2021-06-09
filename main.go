package main

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/client"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/utils/log"
	"os"
	"os/signal"
	"syscall"
)

type Service interface {
	//New()
	Init(log *zap.Logger)
	Start() error
	Stop() error
}

var Services = map[string]Service{
	// CMPP 客户端
	"CmppClient": new(client.CmppClient),
	// CMPP 服务端
	//"CmppServer": new(client.CmppClient),
	// redis 服务
	//"Redis": new(client.CmppClient),
}

func Init() error {
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
	}
	return nil
}

func Stop() error {
	for _, service := range Services {
		service.Init(log.Logger)
		if err := service.Stop(); err != nil {
			return err
		}
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
