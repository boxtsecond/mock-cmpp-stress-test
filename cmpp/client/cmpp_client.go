package client

import (
	"fmt"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/pkg"
	"mock-cmpp-stress-test/config"
	"net"
	"strings"
	"time"
)

var Clients = make(map[string]*pkg.CmppClientManager, 0)

type CmppClient struct {
	cfg    *config.CmppClientConfig
	Logger *zap.Logger
}

func (s *CmppClient) Init(logger *zap.Logger) {
	s.cfg = config.ConfigObj.ClientConfig
	s.Logger = logger
}

func (s *CmppClient) Start() (err error) {
	if !s.cfg.Enable {
		return nil
	}

	version := s.cfg.Version
	timeout := time.Duration(s.cfg.TimeOut) * time.Second
	errCount := 0

	for _, account := range *s.cfg.Accounts {
		cm := &pkg.CmppClientManager{}
		addr := fmt.Sprintf("%s:%d", account.Ip, account.Port)

		initErr := cm.Init(version, addr, account.Username, account.Password, account.SpID, account.SpCode, s.cfg.Retries, timeout)
		if initErr != nil {
			s.Logger.Error("Cmpp Client Init Error",
				zap.String("UserName", account.Username),
				zap.String("Address", addr),
				zap.Error(err))
			return initErr
		}
		s.Logger.Info("Cmpp Client Init Success",
			zap.String("UserName", account.Username),
			zap.String("Address", addr))
		err = cm.Connect()
		if err != nil {
			s.Logger.Error("Cmpp Client Connect Error",
				zap.String("UserName", account.Username),
				zap.String("Address", addr),
				zap.Error(err))
			errCount += 1
			continue
		}
		key := strings.Join([]string{addr, account.Username}, "_")
		Clients[key] = cm
		go cm.KeepAlive()
	}

	if errCount == 0 {
		s.Logger.Info("Cmpp Client Connect Success")
		s.Receive()
	}
	return err
}

func (s *CmppClient) Stop() error {
	for _, client := range Clients {
		client.Disconnect()
	}
	return nil
}

func (s *CmppClient) Receive() {
	if len(Clients) == 0 {
		return
	}

	for _, c := range Clients {
		go func(cm *pkg.CmppClientManager) {
			errCount := 0
			for {
				receivePkg, err := cm.Client.RecvAndUnpackPkt(cm.Timeout)
				if err != nil {
					errCount += 1
					if e, ok := err.(net.Error); ok && e.Timeout() {
						s.Logger.Error("[CmppClient][ReceivePkgs] Error",
							zap.String("UserName", cm.UserName),
							zap.String("Address", cm.Addr),
							zap.Error(err))
						continue
					}
					if errCount > 3 {
						s.Logger.Error("[CmppClient][ReceivePkgs] Error And Return",
							zap.String("UserName", cm.UserName),
							zap.String("Address", cm.Addr),
							zap.Error(err))
						return
					}
				}
				receiveErr := cm.ReceivePkg(receivePkg)
				if receiveErr != nil {
					s.Logger.Error("[CmppClient][ReceivePkgs] Error",
						zap.String("UserName", cm.UserName),
						zap.String("Address", cm.Addr),
						zap.Any("Pkg", receivePkg),
						zap.Error(err))
				}
			}
		}(c)
	}
}
