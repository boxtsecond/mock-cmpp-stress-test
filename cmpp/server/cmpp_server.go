package server

import (
	cmpp "github.com/bigwhite/gocmpp"
	"go.uber.org/zap"

	"mock-cmpp-stress-test/config"
	"time"
)

type CmppServer struct {
	cfg    *config.CmppServerConfig
	Logger *zap.Logger
}

var Cmpp2DeliverChan chan *cmpp.Cmpp2DeliverReqPkt
var Cmpp3DeliverChan chan *cmpp.Cmpp3DeliverReqPkt

func (s *CmppServer) Init(logger *zap.Logger) {
	s.cfg = config.ConfigObj.ServerConfig
	s.Logger = logger
}

func (s *CmppServer) Start() {

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

	for {
		select {
		case cmpp2Deliver := <-Cmpp2DeliverChan:
			cmpp2DeliverPkgs = append(cmpp2DeliverPkgs, cmpp2Deliver)
			if len(cmpp2DeliverPkgs) >= 100 {

			}
		case cmpp3Deliver := <-Cmpp3DeliverChan:
			cmpp3DeliverPkgs = append(cmpp3DeliverPkgs, cmpp3Deliver)
			if len(cmpp3DeliverPkgs) >= 100 {

			}
		case <-tk.C:
			if len(cmpp2DeliverPkgs) > 0 {

			}

			if len(cmpp3DeliverPkgs) > 0 {

			}
		}
	}
}
