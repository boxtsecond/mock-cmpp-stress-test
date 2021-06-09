package pkg

import (
	cmpp "github.com/bigwhite/gocmpp"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/utils/log"
	"net"
)

// =====================CmppClient=====================

func (cm *CmppClientManager) Cmpp2DeliverReq(pkg *cmpp.Cmpp2DeliverReqPkt) error {
	log.Logger.Info("[CmppClient][Cmpp2DeliverReq] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.Any("Pkg", pkg))
	// TODO: 接收回执打点
	return nil
}

func (cm *CmppClientManager) Cmpp3DeliverReq(pkg *cmpp.Cmpp3DeliverReqPkt) error {
	// TODO: 接收回执打点
	log.Logger.Info("[CmppClient][Cmpp3DeliverReq] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.Any("Pkg", pkg))
	return nil
}

// =====================CmppClient=====================

// =====================CmppServer=====================

func (sm *CmppServerManager) Cmpp2Deliver(req *cmpp.Packet, res *cmpp.Response) (bool, error) {
	addr := req.Conn.Conn.RemoteAddr().(*net.TCPAddr).IP.String()

	pkg := req.Packer.(*cmpp.Cmpp2DeliverReqPkt)
	resp := res.Packer.(*cmpp.Cmpp2SubmitRspPkt)

	account, ok := sm.ConnMap[addr]
	if !ok {
		log.Logger.Error("[CmppServer][Cmpp2Deliver] Error",
			zap.String("RemoteAddr", addr))
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}

	msgId, err := GetMsgId(account.spId, pkg.SeqId)
	if err != nil {
		log.Logger.Error("[CmppServer][Cmpp2Submit] GetMsgId Error", zap.Error(err))
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}

	resp.MsgId = msgId
	return false, nil
}

func (sm *CmppServerManager) Cmpp3Deliver(req *cmpp.Packet, res *cmpp.Response) (bool, error) {
	addr := req.Conn.Conn.RemoteAddr().(*net.TCPAddr).IP.String()

	pkg := req.Packer.(*cmpp.Cmpp3SubmitReqPkt)
	resp := res.Packer.(*cmpp.Cmpp3SubmitRspPkt)

	account, ok := sm.ConnMap[addr]
	if !ok {
		log.Logger.Error("[CmppServer][Cmpp2Submit] Error",
			zap.String("Phone", pkg.DestTerminalId[0]),
			zap.String("RemoteAddr", addr))
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}

	msgId, err := GetMsgId(account.spId, pkg.SeqId)
	if err != nil {
		log.Logger.Error("[CmppServer][Cmpp2Submit] GetMsgId Error", zap.Error(err))
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}

	for _, phone := range pkg.DestTerminalId {
		log.Logger.Info("[CmppServer][Cmpp2Submit] Success",
			zap.String("Phone", phone),
			zap.String("MsgId", string(msgId)),
			zap.String("RemoteAddr", addr))
	}

	resp.MsgId = msgId
	return false, nil
}

// =====================CmppServer=====================
