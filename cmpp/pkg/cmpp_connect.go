package pkg

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"go.uber.org/zap"
	_log "log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	cmpp "github.com/bigwhite/gocmpp"
	cmpputils "github.com/bigwhite/gocmpp/utils"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/utils/cron_cache"
	"mock-cmpp-stress-test/utils/log"
)

var Clients = make(map[string]*CmppClientManager)

// =====================CmppClient=====================
func (cm *CmppClientManager) Init(cfg *config.CmppClientConfig, addr string, account config.CmppAccount) error {
	v := GetVersion(cfg.Version)
	if v == InvalidVersion {
		return errors.New("invalid cmpp version")
	}
	cm.Client = cmpp.NewClient(v)
	cm.Addr = addr
	cm.Version = v
	cm.UserName = account.Username
	cm.Password = account.Password
	cm.Timeout = time.Duration(cfg.TimeOut) * time.Second
	cm.ActiveTestInterval = time.Duration(cfg.ActiveTestInterval) * time.Millisecond
	cm.SpId = account.SpID
	cm.SpCode = account.SpCode

	if cm.Timeout > defaultTimeout {
		cm.Timeout = defaultTimeout
	}
	cm.Ctx, cm.cancel = context.WithCancel(context.Background())
	cm.Cmpp2SubmitChan = make(chan *cmpp.Cmpp2SubmitReqPkt, 500)
	cm.Cmpp3SubmitChan = make(chan *cmpp.Cmpp3SubmitReqPkt, 500)
	return nil
}

func (cm *CmppClientManager) Connect() error {
	if cm.Connected {
		return nil
	}
	err := cm.Client.Connect(cm.Addr, cm.UserName, cm.Password, cm.Timeout)
	if err != nil {
		log.Logger.Error("[CmppClient][Connect] Error",
			zap.String("Addr", cm.Addr),
			zap.String("UserName", cm.UserName),
			zap.Error(err))
		return err
	}
	cm.Connected = true
	log.Logger.Info("[CmppClient][Connect] Success.", zap.String("Addr", cm.Addr), zap.String("UserName", cm.UserName), zap.String("Password", cm.Password))
	go cm.KeepAlive()
	go cm.StartSubmit()
	go cm.StartClientReceive()
	return nil
}

func (cm *CmppClientManager) Disconnect() {
	cm.cancel()
	cm.Client.Disconnect()
	log.Logger.Info("[CmppClient][Disconnect] Success", zap.String("Addr", cm.Addr), zap.String("UserName", cm.UserName), zap.String("Password", cm.Password))
}

func (cm *CmppClientManager) Cmpp2ConnRsp(resp *cmpp.Cmpp2ConnRspPkt) error {
	if resp.Status == 0 {
		log.Logger.Info("[CmppClient][Cmpp2ConnRsp] Success", zap.Uint32("SeqId", resp.SeqId), zap.Any("Version", resp.Version))
	} else {
		log.Logger.Info("[CmppClient][Cmpp2ConnRsp] Error", zap.Uint32("SeqId", resp.SeqId), zap.Any("Version", resp.Version))
	}
	return nil
}

func (cm *CmppClientManager) Cmpp3ConnRsp(resp *cmpp.Cmpp3ConnRspPkt) error {
	if resp.Status == 0 {
		log.Logger.Info("[CmppClient][Cmpp3ConnRsp] Success", zap.Uint32("SeqId", resp.SeqId), zap.Any("Version", resp.Version))
	} else {
		log.Logger.Info("[CmppClient][Cmpp3ConnRsp] Error", zap.Uint32("SeqId", resp.SeqId), zap.Any("Version", resp.Version))
	}
	return nil
}

func (cm *CmppClientManager) ReceivePkg(pkg interface{}) error {
	switch p := pkg.(type) {
	case *cmpp.CmppActiveTestReqPkt:
		return cm.CmppActiveTestReq(p) // 收到来自服务端的心跳检测包
	case *cmpp.CmppActiveTestRspPkt:
		return cm.CmppActiveTestRsp(p) // 收到服务端回复的心跳检测包
	case *cmpp.Cmpp2ConnRspPkt: // 服务端连接回包
		return cm.Cmpp2ConnRsp(p)
	case *cmpp.Cmpp3ConnRspPkt: // 服务端连接回包
		return cm.Cmpp3ConnRsp(p)
	case *cmpp.Cmpp2SubmitRspPkt:
		return cm.Cmpp2SubmitResp(p)
	case *cmpp.Cmpp3SubmitRspPkt:
		return cm.Cmpp3SubmitResp(p)

	case *cmpp.Cmpp2DeliverReqPkt:
		return cm.Cmpp2DeliverReq(p)
	case *cmpp.Cmpp3DeliverReqPkt:
		return cm.Cmpp3DeliverReq(p)

	default:
		typeErr := errors.New("unhandled pkg type")
		log.Logger.Error("[CmppClient][ReceivePkgs] Error",
			zap.Error(typeErr),
			zap.Any("pkg.Type", p))
	}
	return nil
}

// 客户端发送心跳检测请求
func (cm *CmppClientManager) KeepAlive() {
	cm.ConnErrCount = 0
	tk := time.NewTicker(cm.ActiveTestInterval)

	defer func() {
		if err := recover(); err != nil {
			log.Logger.Error("[CmppClient][KeepAlive] panic recover", zap.Any("err", err))
		}
		tk.Stop()
	}()

	for {
		if !cm.Connected {
			return
		}

		err := cm.SendCmppActiveTestReq(&cmpp.CmppActiveTestReqPkt{})
		if err != nil {
			log.Logger.Error("[CmppClient][KeepAlive] Check Alive Error", zap.Error(err), zap.String("UserName", cm.UserName))
			cm.ConnErrCount += 1
		} else {
			cm.ConnErrCount = 0
		}

		select {
		case <-tk.C:
			if cm.ConnErrCount > 3 {
				log.Logger.Error("[CmppClient][KeepAlive] KeepAlive Error", zap.String("UserName", cm.UserName))
				cm.Connected = false
				go cm.Reconnect()
				return
			}

		case <-cm.Ctx.Done():
			return
		}
	}

}

func (cm *CmppClientManager) Reconnect() {
	cm.cancel()
	time.Sleep(100 * time.Millisecond)
	ncm := &CmppClientManager{}
	addrArr := strings.Split(cm.Addr, ":")
	port, _ := strconv.Atoi(addrArr[1])
	naccount := config.CmppAccount{
		Username: cm.UserName,
		Password: cm.Password,
		Ip:       addrArr[0],
		Port:     uint16(port),
		SpID:     cm.SpId,
		SpCode:   cm.SpCode,
	}

	initErr := ncm.Init(config.ConfigObj.ClientConfig, cm.Addr, naccount)
	if initErr != nil {
		log.Logger.Error("Cmpp Client Reconnect Init Error",
			zap.String("UserName", naccount.Username),
			zap.String("Address", cm.Addr),
			zap.Error(initErr))
		return
	}
	log.Logger.Info("Cmpp Client Reconnect Init Success",
		zap.String("UserName", naccount.Username),
		zap.String("Address", cm.Addr))

	err := ncm.Connect()
	if err != nil {
		log.Logger.Error("Cmpp Client Reconnect Error",
			zap.String("UserName", ncm.UserName),
			zap.String("Address", ncm.Addr),
			zap.Error(err))
		ncm.Connected = false
		return
	}

	log.Logger.Error("Cmpp Client Reconnect Success",
		zap.String("UserName", ncm.UserName),
		zap.String("Address", ncm.Addr))

	key := strings.Join([]string{ncm.Addr, ncm.UserName}, "_")
	Clients[key] = ncm
}

func (cm *CmppClientManager) StartSubmit() {
	cmpp2SubmitPkgs := make([]*cmpp.Cmpp2SubmitReqPkt, 0)
	cmpp3SubmitPkgs := make([]*cmpp.Cmpp3SubmitReqPkt, 0)
	tk := time.NewTicker(1 * time.Second)

	defer func() {
		tk.Stop()
		if len(cmpp2SubmitPkgs) > 0 {
			cm.BatchCmpp2Submit(cmpp2SubmitPkgs)
		}

		if len(cmpp3SubmitPkgs) > 0 {
			cm.BatchCmpp3Submit(cmpp3SubmitPkgs)
		}
	}()

	for {
		if !cm.Connected {
			return
		}

		select {
		case cmpp2Submit := <-cm.Cmpp2SubmitChan:
			cmpp2SubmitPkgs = append(cmpp2SubmitPkgs, cmpp2Submit)
			if len(cmpp2SubmitPkgs) >= 1000 {
				cm.BatchCmpp2Submit(cmpp2SubmitPkgs)
				cmpp2SubmitPkgs = cmpp2SubmitPkgs[:0]
			}

		case cmpp3Submit := <-cm.Cmpp3SubmitChan:
			cmpp3SubmitPkgs = append(cmpp3SubmitPkgs, cmpp3Submit)
			if len(cmpp3SubmitPkgs) >= 1000 {
				cm.BatchCmpp3Submit(cmpp3SubmitPkgs)
				cmpp3SubmitPkgs = cmpp3SubmitPkgs[:0]

			}
		case <-tk.C:
			if len(cmpp2SubmitPkgs) > 0 {
				cm.BatchCmpp2Submit(cmpp2SubmitPkgs)
				cmpp2SubmitPkgs = cmpp2SubmitPkgs[:0]
			}

			if len(cmpp3SubmitPkgs) > 0 {
				cm.BatchCmpp3Submit(cmpp3SubmitPkgs)
				cmpp3SubmitPkgs = cmpp3SubmitPkgs[:0]
			}
		case <-cm.Ctx.Done():
			return
		}
	}
}

func (cm *CmppClientManager) StartClientReceive() {
	errCount := 0
	for {
		select {
		case <-cm.Ctx.Done():
			return
		default:
			if errCount > 3 {
				log.Logger.Error("[CmppClient][ReceivePkgs] Error And Reconnect",
					zap.String("UserName", cm.UserName),
					zap.String("Address", cm.Addr),
					zap.Int("errCount", errCount))
				cm.Connected = false
				go cm.Reconnect()
				return
			}

			receivePkg, err := cm.Client.RecvAndUnpackPkt(cm.Timeout)
			if err != nil {
				// RecvAndUnpackPkt 长时间没收到包之后会报 timeout，忽略 timeout 错误
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue
				}
				errCount = 4
				log.Logger.Error("[CmppClient][ReceivePkgs] Error",
					zap.String("UserName", cm.UserName),
					zap.String("Address", cm.Addr),
					zap.Error(err))
				continue
			}
			receiveErr := cm.ReceivePkg(receivePkg)
			if receiveErr != nil {
				errCount += 1
				log.Logger.Error("[CmppClient][ReceivePkgs] Error",
					zap.String("UserName", cm.UserName),
					zap.String("Address", cm.Addr),
					zap.Any("Pkg", receivePkg),
					zap.Error(err))
				continue
			}
			errCount = 0
			cm.ConnErrCount = 0
		}
	}
}

// =====================CmppClient=====================

// =====================CmppServer=====================

func newSubmitSeqIdGenerator() (<-chan uint16, chan<- struct{}) {
	seqId := make(chan uint16, 500)
	done := make(chan struct{})

	go func() {
		var i uint16
		for {
			select {
			case seqId <- i:
				i++
			case <-done:
				close(seqId)
				return
			}
		}
	}()

	return seqId, done
}

func (sm *CmppServerManager) Init(version, addr string) error {
	v := GetVersion(version)
	if v == InvalidVersion {
		err := errors.New("invalid cmpp version")
		log.Logger.Error("[CmppServer][GetVersion] Error",
			zap.Error(err))
		return err
	}

	sm.Addr = addr
	sm.Version = v

	cfg := config.ConfigObj.ServerConfig
	sm.heartbeat = time.Duration(cfg.HeartBeat) * time.Second // 每秒心跳检测
	sm.maxNoRespPkgs = int32(cfg.MaxNoRspPkgs)
	sm.ConnMap = &sync.Map{}
	sm.UserMap = &sync.Map{}

	sm.SubmitSeqId, sm.SubmitDone = newSubmitSeqIdGenerator()
	return nil
}

func (sm *CmppServerManager) Start() error {
	// 启动定时
	cron_cache.Start()

	go func() {
		err := cmpp.ListenAndServe(sm.Addr, sm.Version,
			sm.heartbeat,
			sm.maxNoRespPkgs,
			nil,
			cmpp.HandlerFunc(sm.PacketHandler),
		)
		if err != nil {
			log.Logger.Error("[CmppServer][Start] Error",
				zap.Error(err))
			return
		}
	}()

	log.Logger.Info("[CmppServer][Start] Success",
		zap.String("Address", sm.Addr),
		zap.String("Version", sm.Version.String()))
	return nil
}

func (sm *CmppServerManager) LoginAuthAvailable(account *config.CmppServerAuth, reqTime uint32, username, reqAuthSrc string) (bool, string) {

	authSrc := md5.Sum(bytes.Join([][]byte{[]byte(cmpputils.OctetString(username, 6)),
		make([]byte, 9),
		[]byte(account.Password),
		[]byte(cmpputils.TimeStamp2Str(reqTime))},
		nil))
	if reqAuthSrc != string(authSrc[:]) {
		log.Logger.Error("[CmppServer][LoginAuth] invalid password",
			zap.String("UserName", username))
		return false, ""
	}

	authIsmg := md5.Sum(bytes.Join([][]byte{{byte(0)},
		authSrc[:],
		[]byte(account.Password)},
		nil))
	return true, string(authIsmg[:])
}

func (sm *CmppServerManager) Connect(req *cmpp.Packet, res *cmpp.Response) (bool, error) {
	addr := req.Conn.Conn.RemoteAddr().(*net.TCPAddr).String()
	pkg := req.Packer.(*cmpp.CmppConnReqPkt)
	account := cron_cache.GetAccountInfo(pkg.SrcAddr)
	if account == nil {
		log.Logger.Error("[CmppServer][Connect] Error: invalid username",
			zap.String("UserName", pkg.SrcAddr))
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnInvalidSrcAddr]
	}
	if sm.Version == cmpp.V30 {
		resp := res.Packer.(*cmpp.Cmpp3ConnRspPkt)
		auth, authIsmg := sm.LoginAuthAvailable(account, pkg.Timestamp, pkg.SrcAddr, pkg.AuthSrc)
		if !auth {
			resp.Status = uint32(cmpp.ErrnoConnAuthFailed)
			log.Logger.Error("[CmppServer][Cmpp3Conncet] Error",
				zap.String("UserName", pkg.SrcAddr),
				zap.String("Address", addr))
			return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnAuthFailed]
		} else {
			resp.AuthIsmg = authIsmg
		}
	} else {
		resp := res.Packer.(*cmpp.Cmpp2ConnRspPkt)
		auth, authIsmg := sm.LoginAuthAvailable(account, pkg.Timestamp, pkg.SrcAddr, pkg.AuthSrc)
		if !auth {
			resp.Status = cmpp.ErrnoConnAuthFailed
			log.Logger.Error("[CmppServer][Cmpp2Conncet] Error",
				zap.String("UserName", pkg.SrcAddr),
				zap.String("Address", addr))
			return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnAuthFailed]
		} else {
			resp.AuthIsmg = authIsmg
		}
	}

	sm.ConnMap.Store(addr, req)
	sm.UserMap.Store(addr, &Conn{UserName: account.UserName, password: account.Password, spCode: account.SpCode, spId: account.SpId})

	log.Logger.Info("[CmppServer][Login] Success",
		zap.String("UserName", pkg.SrcAddr),
		zap.String("SpId", account.SpId),
		zap.String("SpCode", account.SpCode),
		zap.String("Address", addr))

	return false, nil
}

func (sm *CmppServerManager) PacketHandler(res *cmpp.Response, pkg *cmpp.Packet, l *_log.Logger) (bool, error) {
	switch p := pkg.Packer.(type) {
	case *cmpp.CmppConnReqPkt: // 处理cmpp连接请求
		return sm.Connect(pkg, res)
	case *cmpp.CmppActiveTestReqPkt:
		return sm.CmppActiveTestReq(p, res)
	case *cmpp.Cmpp2SubmitReqPkt:
		return sm.Cmpp2Submit(pkg, res)
	case *cmpp.Cmpp3SubmitReqPkt:
		return sm.Cmpp3Submit(pkg, res)
	case *cmpp.Cmpp2DeliverRspPkt:
		return sm.Cmpp2DeliverResp(p, res)
	case *cmpp.Cmpp3DeliverRspPkt:
		return sm.Cmpp3DeliverResp(p, res)

	case *cmpp.CmppTerminateReqPkt: // 关闭连接
		return false, nil
	}
	return false, nil
}

func (sm *CmppServerManager) Stop() {
	sm.ConnMap.Range(func(_, conn interface{}) bool {
		c := conn.(*cmpp.Packet)
		c.Conn.Close()
		return true
	})
}

// =====================CmppServer=====================
