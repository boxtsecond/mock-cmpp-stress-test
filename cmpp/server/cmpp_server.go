package server

import (
	"fmt"
	cmpp "github.com/bigwhite/gocmpp"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/pkg"
	"mock-cmpp-stress-test/config"
	"time"
)

type CmppServer struct {
	cfg    *config.CmppServerConfig
	Logger *zap.Logger
}

func (s *CmppServer) Init(logger *zap.Logger) {
	s.cfg = config.ConfigObj.ServerConfig
	s.Logger = logger
}
// 连接cmpp客户端
func (s *CmppServer) Start() error {
	if !s.cfg.Enable  {
		return  nil
	}
	csm := pkg.CmppServerManager{}
	addr := fmt.Sprintf( "%s:%d" , s.cfg.IP,s.cfg.Port )
	if err := csm.Init(s.cfg.Version ,addr ) ;err!= nil {
		s.Logger.Error("Cmpp Server Init Error",
			zap.Error(err))
		return err
	}
	go csm.Start()
	s.Logger.Info("Cmpp Server Start Success")
	return nil
}

func (s *CmppServer) Stop() error {

	return nil
}

func (s *CmppServer) StartDeliver() {
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Error("Deliver Worker Error")
		}
	}()
	cmpp2DeliverPkgs := make([]*cmpp.Cmpp2DeliverReqPkt, 0)
	cmpp3DeliverPkgs := make([]*cmpp.Cmpp3DeliverReqPkt, 0)
	tk := time.NewTicker(1 * time.Second)
	defer tk.Stop()
	// csm := pkg.CmppServerManager{}
	for {
		select {
		case cmpp2Deliver := <-pkg.Cmpp2DeliverChan:
			cmpp2DeliverPkgs = append(cmpp2DeliverPkgs, cmpp2Deliver)
			if len(cmpp2DeliverPkgs) >= 100 { // 批次设置为100,推送给指定客户端
				//csm.Cmpp2Deliver(cmpp2DeliverPkgs)
				//cmpp2DeliverPkgs =  make([]*cmpp.Cmpp2DeliverReqPkt, 0)
			}

		case cmpp3Deliver := <-pkg.Cmpp3DeliverChan:
			cmpp3DeliverPkgs = append(cmpp3DeliverPkgs, cmpp3Deliver)
			if len(cmpp3DeliverPkgs) >= 100 {
				//csm.Cmpp2Deliver(cmpp2DeliverPkgs)
				//cmpp2DeliverPkgs =  make([]*cmpp.Cmpp2DeliverReqPkt, 0)

			}
		case <-tk.C:
			if len(cmpp2DeliverPkgs) > 0 {
				//csm.Cmpp2Deliver(cmpp2DeliverPkgs)
			}

			if len(cmpp3DeliverPkgs) > 0 {
				//csm.Cmpp2Deliver(cmpp2DeliverPkgs)
			}
		}
	}
}

