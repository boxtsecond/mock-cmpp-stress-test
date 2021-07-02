package client

import (
	"fmt"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/pkg"
	"mock-cmpp-stress-test/config"
	"strings"
)

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
	errCount := 0

	for _, account := range *s.cfg.Accounts {
		cm := &pkg.CmppClientManager{}
		addr := fmt.Sprintf("%s:%d", account.Ip, account.Port)

		initErr := cm.Init(s.cfg, addr, account)
		if initErr != nil {
			s.Logger.Error("Cmpp Client Init Error",
				zap.String("UserName", account.Username),
				zap.String("Address", addr),
				zap.Error(initErr))
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
		pkg.Clients[key] = cm
	}

	if errCount == 0 {
		s.Logger.Info("Cmpp Client Connect Success")
	}
	return err
}

func (s *CmppClient) Stop() error {
	for _, client := range pkg.Clients {
		client.Disconnect()
	}
	s.Logger.Info("Cmpp Client Stop Success")
	return nil
}
