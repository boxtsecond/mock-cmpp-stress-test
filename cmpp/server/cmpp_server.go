package server

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/pkg"
	"mock-cmpp-stress-test/config"
	"time"
)

var csm pkg.CmppServerManager

type CmppServer struct {
	cfg    *config.CmppServerConfig
	Logger *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *CmppServer) Init(logger *zap.Logger) {
	s.cfg = config.ConfigObj.ServerConfig
	s.Logger = logger
	s.ctx, s.cancel = context.WithCancel(context.Background())
}

// 启动cmpp服务
func (s *CmppServer) Start() error {
	if !s.cfg.Enable {
		return nil
	}

	defer func() {
		csm.Stop()
	}()
	addr := fmt.Sprintf("%s:%d", s.cfg.IP, s.cfg.Port)
	if err := csm.Init(s.cfg.Version, addr); err != nil {
		s.Logger.Error("Cmpp Server Init Error",
			zap.Error(err))
		return err
	}
	go func() {
		if err := csm.Start(); err != nil {
			s.Logger.Error("Cmpp Server Start Error",
				zap.Error(err))
			csm.Stop()
		}
	}()
	go s.StartDeliver()

	s.Logger.Info("Cmpp Server Start Success")
	return nil
}

func (s *CmppServer) Stop() error {
	if !s.cfg.Enable {
		return nil
	}
	// 关闭当前所有连接
	csm.Stop()
	s.Logger.Info("Cmpp Server Stop Success")
	return nil
}

func (s *CmppServer) StartDeliver() {
	cmpp2DeliverPkgs := make([]*pkg.MockCmpp2DeliverPkg, 0)
	cmpp3DeliverPkgs := make([]*pkg.MockCmpp3DeliverPkg, 0)
	tk := time.NewTicker(time.Duration(s.cfg.DeliverInterval) * time.Second)

	defer func() {
		tk.Stop()
		if len(cmpp2DeliverPkgs) > 0 {
			csm.BatchCmpp2Deliver(cmpp2DeliverPkgs)
		}

		if len(cmpp3DeliverPkgs) > 0 {
			csm.BatchCmpp3Deliver(cmpp3DeliverPkgs)
		}
	}()

	for {
		select {
		case cmpp2Deliver := <-pkg.Cmpp2DeliverChan:
			cmpp2DeliverPkgs = append(cmpp2DeliverPkgs, cmpp2Deliver)
			if len(cmpp2DeliverPkgs) >= 500 {
				csm.BatchCmpp2Deliver(cmpp2DeliverPkgs)
				cmpp2DeliverPkgs = cmpp2DeliverPkgs[:0]
			}

		case cmpp3Deliver := <-pkg.Cmpp3DeliverChan:
			cmpp3DeliverPkgs = append(cmpp3DeliverPkgs, cmpp3Deliver)
			if len(cmpp3DeliverPkgs) >= 500 {
				csm.BatchCmpp3Deliver(cmpp3DeliverPkgs)
				cmpp3DeliverPkgs = cmpp3DeliverPkgs[:0]

			}
		case <-tk.C:

			if len(cmpp2DeliverPkgs) > 0 {
				csm.BatchCmpp2Deliver(cmpp2DeliverPkgs)
				cmpp2DeliverPkgs = cmpp2DeliverPkgs[:0]
			}

			if len(cmpp3DeliverPkgs) > 0 {
				csm.BatchCmpp3Deliver(cmpp3DeliverPkgs)
				cmpp3DeliverPkgs = cmpp3DeliverPkgs[:0]
			}
		case <-s.ctx.Done():
			return
		}
	}
}
