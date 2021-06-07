package client

import (
	goCmpp "github.com/bigwhite/gocmpp"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/pkg"
	"mock-cmpp-stress-test/config"
	"net"
	"time"
)

type CmppClient struct {
	cfg    *config.CmppClientConfig
	Logger *zap.Logger
}

func (s *CmppClient) Init(logger *zap.Logger) {
	s.cfg = config.ConfigObj.ClientConfig
	s.Logger = logger
}

func (s *CmppClient) Start() error {
	if !s.cfg.CmppClient.Enable {
		return nil
	}

	version := s.cfg.CmppClient.Version
	client, err := pkg.NewClient(version)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//
	//c := NewCmppClientManager(endpoint, cmppClient)
	//c.login()
	return nil
}

func (s *CmppClient) Stop() error {
	//version := s.cfg.CmppClient.Version
	//cmppClient, err := newClient()
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//
	//c := NewCmppClientManager(endpoint, cmppClient)
	//c.login()
	return nil
}

//func (s *CmppClient) connect() error {
//	version := s.cfg.CmppClient.Version
//
//}

//var cmppClientsMap map[string]*CmppClientManager
//
//type CmppClientManager struct {
//	Accounts  *[]config.CmppAccount
//	Client    *goCmpp.Client
//	PingChan  chan struct{}
//	Connected bool
//	retries   uint8
//	logger    *zap.Logger
//}
//
//func (c *CmppClientManager) Init() {
//	//cfg := config.ConfigObj.ClientConfig
//}
//
//func (c *CmppClientManager) Connect(addr, username, password string, timeout time.Duration) error {
//	conn, err := net.DialTimeout("tcp", addr, timeout)
//	if err != nil {
//		return err
//	}
//
//	c.conn = NewConn(conn, c.v)
//	c.conn.SetState(CONN_CONNECTED)
//
//	var req Packer
//
//	if c.v == pkg.V20 || c.v == pkg.V21 {
//		// login
//		req = &pkg.Cmpp2ConnReqPkg{
//			SrcAddr: username,
//			Secret:  password,
//			Version: c.v,
//		}
//	} else {
//		req = &pkg.Cmpp3ConnReqPkg{
//			SrcAddr: username,
//			Secret:  password,
//			Version: c.v,
//		}
//	}
//
//	// 发送connect包
//	_, err = c.SendReqPkg(req)
//	if err != nil {
//		return err
//	}
//	// 接收connect_resp包
//	p, _, _, err := c.RecvAndUnpackPkg(timeout)
//	if err != nil {
//		return err
//	}
//
//	var ok bool
//	var status uint8
//	if c.v == pkg.V20 || c.v == pkg.V21 {
//		var reqObj *pkg.Cmpp2ConnRespPkg
//		reqObj, ok = p.(*pkg.Cmpp2ConnRespPkg)
//		if !ok {
//			return pkg.ErrConnVersionTooHigh
//		}
//		status = reqObj.Status
//	} else {
//		var reqObj *pkg.Cmpp3ConnRespPkg
//		reqObj, ok = p.(*pkg.Cmpp3ConnRespPkg)
//		if !ok {
//			return pkg.ErrConnVersionTooHigh
//		}
//		status = uint8(reqObj.Status)
//	}
//
//	if status != 0 {
//		return pkg.ConnRespStatusErrMap[status]
//	}
//
//	return nil
//}
//
//func NewConn(conn net.Conn, v pkg.Version) *Conn {
//	seqId, done := newSeqIdGenerator()
//	submitSeqId, submitDone := newSubmitSeqIdGenerator()
//
//	c := &Conn{
//		Conn:        conn,
//		Version:     v,
//		SeqId:       seqId,
//		SubmitSeqId: submitSeqId,
//		done:        done,
//		submitDone:  submitDone,
//		stop:        make(chan struct{}),
//	}
//
//	tcp := c.Conn.(*net.TCPConn)
//	tcp.SetKeepAlive(true)
//
//	return c
//}
