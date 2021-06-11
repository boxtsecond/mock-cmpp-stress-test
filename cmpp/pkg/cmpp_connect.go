package pkg

import (
	"bytes"
	"crypto/md5"
	"errors"
	cmpp "github.com/bigwhite/gocmpp"
	cmpputils "github.com/bigwhite/gocmpp/utils"
	"go.uber.org/zap"
	_log "log"
	"mock-cmpp-stress-test/cmpp/cron_cache"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/utils/cache"
	"mock-cmpp-stress-test/utils/log"
	"net"
	"time"
)

// =====================CmppClient=====================
func (cm *CmppClientManager) Init(version, addr, username, password, spId, spCode string, retryTimes uint, timeout time.Duration) error {
	v := GetVersion(version)
	if v == InvalidVersion {
		return errors.New("invalid cmpp version")
	}
	cm.Client = cmpp.NewClient(v)
	cm.Addr = addr
	cm.Version = v
	cm.UserName = username
	cm.Password = password
	cm.Retries = retryTimes
	cm.Timeout = timeout
	cm.SpId = spId
	cm.SpCode = spCode
	cm.Cache = (&cache.Cache{}).New(1e4)
	go cm.Cache.StartRetry()

	if cm.Retries == 0 {
		cm.Retries = defaultRetries
	}

	cm.Timeout = timeout
	if cm.Timeout > defaultTimeout {
		cm.Timeout = defaultTimeout
	}
	return nil
}

func (cm *CmppClientManager) Connect() error {
	if cm.Connected {
		return nil
	}
	err := cm.Client.Connect(cm.Addr, cm.UserName, cm.Password, cm.Timeout)
	if err != nil {
		log.Logger.Error("[CmppClient][Connect] Error",
			zap.Uint("Retries", cm.Retries),
			zap.Error(err))
		if cm.Retries <= 0 {
			return err
		}
		cm.Retries -= 1
		return cm.Connect()
	}
	cm.Connected = true
	log.Logger.Info("[CmppClient][Connect] Success.", zap.String("Addr", cm.Addr), zap.String("UserName", cm.UserName), zap.String("Password", cm.Password))
	return nil
}

func (cm *CmppClientManager) Disconnect() {
	cm.Client.Disconnect()
	cm.Connected = false
	log.Logger.Info("[CmppClient][Disconnect] Success", zap.String("Addr", cm.Addr), zap.String("UserName", cm.UserName), zap.String("Password", cm.Password))
}

func (cm *CmppClientManager) ReceivePkg(pkg interface{}) error {
	switch p := pkg.(type) {
	case *cmpp.CmppActiveTestReqPkt:
		return cm.CmppActiveTestReq(p)
	case *cmpp.CmppActiveTestRspPkt:
		return cm.CmppActiveTestRsp(p)

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
			zap.Error(typeErr))
	}
	return nil
}

// =====================CmppClient=====================

// =====================CmppServer=====================
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
	// 读取账户？？有了缓存是不是其实没用了
	// sm.UserMap =
	//for _, auth := range *cfg.Auths {
	//	sm.UserMap[auth.UserName] = &Conn{
	//		UserName: auth.UserName,
	//		password: auth.Password,
	//		spId:     auth.SpId,
	//		spCode:   auth.SpCode,
	//	}
	//}
	sm.heartbeat = time.Duration(cfg.HeartBeat) * time.Second
	sm.maxNoRespPkgs = int32(cfg.MaxNoRspPkgs)
	sm.ConnMap = make(map[string]*Conn)
	// sm.Server = cmpp.NewServer()

	return nil
}

func (sm *CmppServerManager) Start() error {
	// 启动定时
	cron_cache.Start()
	// 启动端口服务
	err := cmpp.ListenAndServe(sm.Addr, sm.Version,
		sm.heartbeat,
		sm.maxNoRespPkgs,
		nil,
		cmpp.HandlerFunc(sm.PacketHandler),
	)
	if err != nil {
		log.Logger.Error("[CmppServer][Start] Error",
			zap.Error(err))
		return err
	}
	log.Logger.Info("[CmppServer][Start] Success",
		zap.String("Address", sm.Addr),
		zap.String("Version", string(sm.Version)))
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
	addr := req.Conn.Conn.RemoteAddr().(*net.TCPAddr).IP.String()

	pkg := req.Packer.(*cmpp.CmppConnReqPkt)
	account  := cron_cache.GetAccountInfo(pkg.SrcAddr)
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
	sm.ConnMap[addr] = &Conn{
		UserName: pkg.SrcAddr,
		password: account.Password,
		spId:     account.SpId,
		spCode:   account.SpCode,
	}

	log.Logger.Info("[CmppServer][Login] Success",
		zap.String("UserName", pkg.SrcAddr),
		zap.String("Address", addr))

	return false, nil
}

func (sm *CmppServerManager) PacketHandler(res *cmpp.Response, pkg *cmpp.Packet, l *_log.Logger) (bool, error) {
	switch pkg.Packer.(type) {
	case *cmpp.CmppConnReqPkt:
		return sm.Connect(pkg, res)
	case *cmpp.Cmpp2SubmitRspPkt: // 处理cmpp心跳包
		return sm.DealCmppActiveTestReq(pkg ,res)
	case *cmpp.Cmpp2SubmitReqPkt:
		return sm.Cmpp2Submit(pkg, res)
	case *cmpp.Cmpp3SubmitReqPkt:
		return sm.Cmpp3Submit(pkg, res)

	//case *cmpp.CmppActiveTestRespPkg: //
	//	reqObj := req.(*pkg.CmppActiveTestRespPkg)
	//	return dealCmppActiveTestResp(conn, reqObj)
		//case *pkg.CmppTerminateReqPkg:
	//	reqObj := req.(*pkg.CmppTerminateReqPkg)
	//	respObj := resp.(*pkg.CmppTerminateRespPkg)
	//	return dealCmppTerminate(conn, reqObj, respObj)
	//case *pkg.CmppTerminateRespPkg:
	//	reqObj := req.(*pkg.CmppTerminateRespPkg)
	//	return dealCmppTerminateResp(conn, reqObj)
	//

	default:

	}
	return false, nil
}

// =====================CmppServer=====================
