package server

import (
	"go.uber.org/zap"
	"mock-cmpp-stress-test/config"
)

type CmppServer struct {
	cfg    *config.CmppServerConfig
	Logger *zap.Logger
}

func (s *CmppServer) Init(logger *zap.Logger) {
	s.cfg = config.ConfigObj.ServerConfig
	s.Logger = logger
}
