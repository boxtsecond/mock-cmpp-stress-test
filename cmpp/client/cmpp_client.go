package client

import (
	"fmt"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/pkg"
	"mock-cmpp-stress-test/config"
	"net"
	"strings"
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
	errCount := 0

	for _, account := range *s.cfg.Accounts {
		cm := &pkg.CmppClientManager{}
		addr := fmt.Sprintf("%s:%d", account.Ip, account.Port)

		initErr := cm.Init(s.cfg, addr, account)
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
	s.Logger.Info("Cmpp Client Stop Success")
	return nil
}

func (s *CmppClient) ClientReceive(cm *pkg.CmppClientManager) {
	errCount := 0
	for {
		if errCount >= 3 {
			s.Logger.Error("[CmppClient][ReceivePkgs] Error And Reconnect",
				zap.String("UserName", cm.UserName),
				zap.String("Address", cm.Addr),
				zap.Int("errCount", errCount))
			cm.Connected = false
			if err := cm.Connect(); err != nil {
				s.Logger.Error("[CmppClient][ReceivePkgs] Reconnect Error And Return",
					zap.String("UserName", cm.UserName),
					zap.String("Address", cm.Addr))
				return
			}
			go cm.KeepAlive()
			go s.ClientReceive(cm)
			return
		}

		select {
		case <-cm.Ctx.Done():
			return
		default:
			receivePkg, err := cm.Client.RecvAndUnpackPkt(cm.Timeout)
			if err != nil {
				// RecvAndUnpackPkt 长时间没收到包之后会报 timeout，忽略 timeout 错误
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue
				}
				errCount = 3
				s.Logger.Error("[CmppClient][ReceivePkgs] Error",
					zap.String("UserName", cm.UserName),
					zap.String("Address", cm.Addr),
					zap.Error(err))
				continue
			}
			receiveErr := cm.ReceivePkg(receivePkg)
			if receiveErr != nil {
				errCount += 1
				s.Logger.Error("[CmppClient][ReceivePkgs] Error",
					zap.String("UserName", cm.UserName),
					zap.String("Address", cm.Addr),
					zap.Any("Pkg", receivePkg),
					zap.Error(err))
				continue
			}
			errCount = 0
		}
	}
}

func (s *CmppClient) Receive() {
	if len(Clients) == 0 {
		return
	}

	for _, c := range Clients {
		go s.ClientReceive(c)
	}
}
