package pkg

import (
	"errors"
	cmpp "github.com/bigwhite/gocmpp"
	cmpputils "github.com/bigwhite/gocmpp/utils"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/server"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/utils/log"
	"net"
	"strconv"
	"strings"
)

// =====================CmppClient=====================
// =====================Cmpp2Submit=====================
func (cm *CmppClientManager) GetCmppSubmit2ReqPkg(message *config.CmppMessages) ([]*cmpp.Cmpp2SubmitReqPkt, error) {
	packets := make([]*cmpp.Cmpp2SubmitReqPkt, 0)
	content, err := cmpputils.Utf8ToUcs2(message.Content)
	if err != nil {
		return nil, err
	}

	chunks := cm.SplitLongSms(content)
	var tpUdhi uint8
	if len(chunks) > 1 {
		tpUdhi = 1
	}

	srcId := strings.Join([]string{cm.SpCode, message.Extend}, "")
	if len(srcId) > 21 {
		srcId = srcId[:21]
	}

	for i, chunk := range chunks {
		p := &cmpp.Cmpp2SubmitReqPkt{
			PkTotal:            uint8(len(chunks)),
			PkNumber:           uint8(i + 1),
			RegisteredDelivery: 1,
			MsgLevel:           1,
			ServiceId:          cm.SpId,
			FeeUserType:        2,
			TpUdhi:             tpUdhi,
			FeeTerminalId:      message.Phone,
			MsgFmt:             8,
			MsgSrc:             cm.SpId,
			FeeType:            "02",
			FeeCode:            "10",
			ValidTime:          "151105131555101+",
			AtTime:             "",
			SrcId:              srcId,
			DestUsrTl:          1,
			DestTerminalId:     []string{message.Phone},
			MsgLength:          uint8(len(chunk)),
			MsgContent:         string(chunk),
		}
		packets = append(packets, p)
	}

	return packets, nil
}

func (cm *CmppClientManager) Cmpp2Submit(message *config.CmppMessages) (error, []uint32) {
	pkgs, err := cm.GetCmppSubmit2ReqPkg(message)
	seqIds := make([]uint32, 0)
	if err != nil {
		log.Logger.Error("[CmppClient][GetCmppSubmit2ReqPkg] Error:", zap.Error(err))
		return err, seqIds
	}
	for _, pkg := range pkgs {
		seqId, sendErr := cm.Client.SendReqPkt(pkg)
		if sendErr != nil {
			log.Logger.Error("[CmppClient][Cmpp2Submit] Error:", zap.Error(sendErr))
			return sendErr, seqIds
		}
		seqIds = append(seqIds, seqId)
		setCacheErr := cm.Cache.Set(strconv.Itoa(int(seqId)), "")
		if setCacheErr != nil {
			log.Logger.Error("[CmppClient][Cmpp2Submit][SetCache] Error:", zap.Error(setCacheErr))
		}
	}
	log.Logger.Info("[CmppClient][Cmpp2Submit] Success", zap.String("Addr", cm.Addr), zap.String("UserName", cm.UserName), zap.String("SpId", cm.SpId), zap.String("SpCode", cm.SpCode), zap.String("Phone", message.Phone), zap.Any("SeqIds", seqIds))
	return nil, seqIds
}

func (cm *CmppClientManager) Cmpp2SubmitResp(resp *cmpp.Cmpp2SubmitRspPkt) error {
	key := strconv.Itoa(int(resp.SeqId))
	data := cm.Cache.Get(key)
	defer cm.Cache.Delete(key)

	if data == "" {
		return errors.New("Get Cache Error ")
	}

	// TODO: Add statistics, cache or redis
	if resp.Result == 0 {
		log.Logger.Info("[CmppClient][Cmpp2SubmitResp] Success",
			zap.Uint32("SeqId", resp.SeqId),
			zap.Uint64("MsgId", resp.MsgId))
	} else {
		log.Logger.Info("[CmppClient][Cmpp2SubmitResp] Error",
			zap.Uint32("SeqId", resp.SeqId),
			zap.Uint64("MsgId", resp.MsgId),
			zap.Uint8("ErrorCode", resp.Result))
	}
	return nil
}

// =====================Cmpp2Submit=====================

// =====================Cmpp3Submit=====================
func (cm *CmppClientManager) GetCmppSubmit3ReqPkg(message *config.CmppMessages) ([]*cmpp.Cmpp3SubmitReqPkt, error) {
	packets := make([]*cmpp.Cmpp3SubmitReqPkt, 0)
	content, err := cmpputils.Utf8ToUcs2(message.Content)
	if err != nil {
		return nil, err
	}

	chunks := cm.SplitLongSms(content)
	var tpUdhi uint8
	if len(chunks) > 1 {
		tpUdhi = 1
	}

	srcId := strings.Join([]string{cm.SpCode, message.Extend}, "")
	if len(srcId) > 21 {
		srcId = srcId[:21]
	}

	for i, chunk := range chunks {
		p := &cmpp.Cmpp3SubmitReqPkt{
			PkTotal:            uint8(len(chunks)),
			PkNumber:           uint8(i + 1),
			RegisteredDelivery: 1,
			MsgLevel:           1,
			ServiceId:          cm.SpId,
			FeeUserType:        2,
			FeeTerminalId:      message.Phone,
			FeeTerminalType:    0,
			TpUdhi:             tpUdhi,
			MsgFmt:             8,
			MsgSrc:             cm.SpId,
			FeeType:            "02",
			FeeCode:            "10",
			ValidTime:          "151105131555101+",
			AtTime:             "",
			SrcId:              srcId,
			DestUsrTl:          1,
			DestTerminalId:     []string{message.Phone},
			DestTerminalType:   0,
			MsgLength:          uint8(len(chunk)),
			MsgContent:         string(chunk),
		}
		packets = append(packets, p)
	}

	return packets, nil
}

func (cm *CmppClientManager) Cmpp3Submit(message *config.CmppMessages) (error, []uint32) {
	pkgs, err := cm.GetCmppSubmit3ReqPkg(message)
	seqIds := make([]uint32, 0)
	if err != nil {
		log.Logger.Error("[CmppClient][GetCmppSubmit3ReqPkg] Error:", zap.Error(err))
		return err, seqIds
	}
	for _, pkg := range pkgs {
		seqId, sendErr := cm.Client.SendReqPkt(pkg)
		if sendErr != nil {
			log.Logger.Error("[CmppClient][Cmpp3Submit] Error:", zap.Error(sendErr))
			return sendErr, seqIds
		}

		seqIds = append(seqIds, seqId)
		setCacheErr := cm.Cache.Set(strconv.Itoa(int(seqId)), "")
		if setCacheErr != nil {
			log.Logger.Error("[CmppClient][Cmpp3Submit][SetCache] Error:", zap.Error(setCacheErr))
		}
	}
	log.Logger.Info("[CmppClient][Cmpp3Submit] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.String("SpId", cm.SpId),
		zap.String("SpCode", cm.SpCode),
		zap.String("Phone", message.Phone),
		zap.Any("SeqIds", seqIds))

	return nil, seqIds
}

func (cm *CmppClientManager) Cmpp3SubmitResp(resp *cmpp.Cmpp3SubmitRspPkt) error {
	key := strconv.Itoa(int(resp.SeqId))
	data := cm.Cache.Get(key)
	defer cm.Cache.Delete(key)

	if data == "" {
		return errors.New("Get Cache Error ")
	}

	// TODO: Add statistics, cache or redis
	if resp.Result == 0 {
		log.Logger.Info("[CmppClient][Cmpp3SubmitResp] Success", zap.Uint32("SeqId", resp.SeqId), zap.Uint64("MsgId", resp.MsgId))
	} else {
		log.Logger.Info("[CmppClient][Cmpp3SubmitResp] Error", zap.Uint32("SeqId", resp.SeqId), zap.Uint64("MsgId", resp.MsgId), zap.Uint32("ErrorCode", resp.Result))
	}
	return nil
}

// =====================Cmpp3Submit=====================
// =====================CmppClient=====================

// =====================CmppServer=====================
func (sm *CmppServerManager) Cmpp2Submit(req *cmpp.Packet, res *cmpp.Response) (bool, error) {
	addr := req.Conn.Conn.RemoteAddr().(*net.TCPAddr).IP.String()

	pkg := req.Packer.(*cmpp.Cmpp2SubmitReqPkt)
	resp := res.Packer.(*cmpp.Cmpp2SubmitRspPkt)

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
		deliverPkg := &cmpp.Cmpp2DeliverReqPkt{
			MsgId:            msgId,
			DestId:           account.spCode,
			ServiceId:        account.spId,
			TpPid:            0,
			TpUdhi:           0,
			MsgFmt:           0,
			SrcTerminalId:    phone,
			RegisterDelivery: 1,
			MsgLength:        uint8(len("DELIVRD")),
			MsgContent:       "DELIVRD",
		}
		log.Logger.Info("[CmppServer][Cmpp2Submit] Success",
			zap.String("Phone", phone),
			zap.String("MsgId", string(msgId)),
			zap.String("RemoteAddr", addr))

		// 返回状态报告
		server.Cmpp2DeliverChan <- deliverPkg
	}

	resp.MsgId = msgId
	return false, nil
}

func (sm *CmppServerManager) Cmpp3Submit(req *cmpp.Packet, res *cmpp.Response) (bool, error) {
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
		deliverPkg := &cmpp.Cmpp3DeliverReqPkt{
			MsgId:            msgId,
			DestId:           account.spCode,
			ServiceId:        account.spId,
			TpPid:            0,
			TpUdhi:           0,
			MsgFmt:           0,
			SrcTerminalId:    phone,
			RegisterDelivery: 1,
			MsgLength:        uint8(len("DELIVRD")),
			MsgContent:       "DELIVRD",
		}
		log.Logger.Info("[CmppServer][Cmpp3Submit] Success",
			zap.String("Phone", phone),
			zap.String("MsgId", string(msgId)),
			zap.String("RemoteAddr", addr))

		// 返回状态报告
		server.Cmpp3DeliverChan <- deliverPkg
	}

	resp.MsgId = msgId
	return false, nil
}

// =====================CmppServer=====================
