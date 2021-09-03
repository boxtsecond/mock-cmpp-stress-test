package pkg

import (
	"fmt"
	cmpp "github.com/bigwhite/gocmpp"
	cmpputils "github.com/bigwhite/gocmpp/utils"
	"go.uber.org/zap"
	"math"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/statistics"
	"mock-cmpp-stress-test/utils/log"
	"net"
	"strconv"
	"strings"
	"time"
)

// =====================CmppClient=====================
// =====================Cmpp2Submit=====================

func (cm *CmppClientManager) GetCmppSubmit2ReqPkg(message *config.TextMessages) ([]*cmpp.Cmpp2SubmitReqPkt, error) {
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

func (cm *CmppClientManager) Cmpp2Submit(message *config.TextMessages) {
	pkgs, err := cm.GetCmppSubmit2ReqPkg(message)
	if err != nil {
		log.Logger.Error("[CmppClient][GetCmppSubmit2ReqPkg] Error:", zap.Error(err))
		return
	}

	for _, pkg := range pkgs {
		cm.SendCmpp2SubmitPkg(pkg)
	}

}

func (sm *CmppClientManager) SendCmpp2SubmitPkg(pkg *cmpp.Cmpp2SubmitReqPkt) {
	sm.Cmpp2SubmitChan <- pkg
}

func (cm *CmppClientManager) Cmpp2SubmitResp(resp *cmpp.Cmpp2SubmitRspPkt) error {
	if resp.Result == 0 {
		log.Logger.Info("[CmppClient][Cmpp2SubmitResp] Success",
			zap.String("Addr", cm.Addr),
			zap.String("UserName", cm.UserName),
			zap.Uint32("SeqId", resp.SeqId),
			zap.Uint64("MsgId", resp.MsgId))
		statistics.CollectService.Service.AddPackerStatistics("Client", "SubmitResp", true)
	} else {
		log.Logger.Info("[CmppClient][Cmpp2SubmitResp] Error",
			zap.String("Addr", cm.Addr),
			zap.String("UserName", cm.UserName),
			zap.Uint32("SeqId", resp.SeqId),
			zap.Uint64("MsgId", resp.MsgId),
			zap.Uint8("ErrorCode", resp.Result))
		statistics.CollectService.Service.AddPackerStatistics("Client", "SubmitResp", false)
	}
	return nil
}

// =====================Cmpp2Submit=====================

// =====================Cmpp3Submit=====================
func (cm *CmppClientManager) GetCmppSubmit3ReqPkg(message *config.TextMessages) ([]*cmpp.Cmpp3SubmitReqPkt, error) {
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

func (cm *CmppClientManager) Cmpp3Submit(message *config.TextMessages) {
	pkgs, err := cm.GetCmppSubmit3ReqPkg(message)
	if err != nil {
		log.Logger.Error("[CmppClient][GetCmppSubmit3ReqPkg] Error:", zap.Error(err))
		return
	}
	for _, pkg := range pkgs {
		cm.SendCmpp3SubmitPkg(pkg)
	}

}

func (sm *CmppClientManager) SendCmpp3SubmitPkg(pkg *cmpp.Cmpp3SubmitReqPkt) {
	sm.Cmpp3SubmitChan <- pkg
}

func (cm *CmppClientManager) Cmpp3SubmitResp(resp *cmpp.Cmpp3SubmitRspPkt) error {
	if resp.Result == 0 {
		log.Logger.Info("[CmppClient][Cmpp3SubmitResp] Success", zap.Uint32("SeqId", resp.SeqId), zap.Uint64("MsgId", resp.MsgId))
		statistics.CollectService.Service.AddPackerStatistics("Client", "SubmitResp", true)
	} else {
		log.Logger.Info("[CmppClient][Cmpp3SubmitResp] Error", zap.Uint32("SeqId", resp.SeqId), zap.Uint64("MsgId", resp.MsgId), zap.Uint32("ErrorCode", resp.Result))
		statistics.CollectService.Service.AddPackerStatistics("Client", "SubmitResp", false)
	}
	return nil
}

// =====================Cmpp3Submit=====================
// =====================CmppClient=====================

// =====================CmppServer=====================
func (sm *CmppServerManager) Cmpp2Submit(req *cmpp.Packet, res *cmpp.Response) (bool, error) {
	addr := req.Conn.Conn.RemoteAddr().(*net.TCPAddr).String()

	pkg := req.Packer.(*cmpp.Cmpp2SubmitReqPkt)
	resp := res.Packer.(*cmpp.Cmpp2SubmitRspPkt)
	a, ok := sm.UserMap.Load(addr)
	account := a.(*Conn)
	if !ok {
		log.Logger.Error("[CmppServer][Cmpp2Submit] Error",
			zap.String("Phone", pkg.DestTerminalId[0]),
			zap.String("RemoteAddr", addr))
		statistics.CollectService.Service.AddPackerStatistics("Server", "Submit", false)
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}

	seqId := <-sm.SubmitSeqId
	msgId, err := GetMsgId(account.spId, seqId)
	if err != nil {
		log.Logger.Error("[CmppServer][Cmpp2Submit] GetMsgId Error",
			zap.String("SpId", account.spId),
			zap.Uint32("SeqId", pkg.SeqId),
			zap.Error(err))
		statistics.CollectService.Service.AddPackerStatistics("Server", "Submit", false)
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}

	log.Logger.Info("[CmppServer][Cmpp2Submit] Success",
		zap.String("SpId", account.spId),
		zap.String("Phone", pkg.DestTerminalId[0]),
		zap.Uint16("SeqId", seqId),
		zap.Uint64("MsgId", msgId),
		zap.String("RemoteAddr", addr))
	statistics.CollectService.Service.AddPackerStatistics("Server", "Submit", true)
	statistics.CollectService.Service.AddPackerStatistics("Server", "SubmitResp", true)
	resp.MsgId = msgId
	go sm.MockCmpp2Deliver(addr, account.spCode, msgId, pkg)
	return false, nil
}

func (sm *CmppServerManager) Cmpp3Submit(req *cmpp.Packet, res *cmpp.Response) (bool, error) {
	addr := req.Conn.Conn.RemoteAddr().(*net.TCPAddr).String()

	pkg := req.Packer.(*cmpp.Cmpp3SubmitReqPkt)
	resp := res.Packer.(*cmpp.Cmpp3SubmitRspPkt)
	a, ok := sm.UserMap.Load(addr)
	account := a.(*Conn)
	if !ok {
		log.Logger.Error("[CmppServer][Cmpp3Submit] Error",
			zap.String("Phone", pkg.DestTerminalId[0]),
			zap.String("RemoteAddr", addr))
		statistics.CollectService.Service.AddPackerStatistics("Server", "Submit", false)
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}

	seqId := <-sm.SubmitSeqId
	msgId, err := GetMsgId(account.spId, seqId)
	if err != nil {
		log.Logger.Error("[CmppServer][Cmpp3Submit] GetMsgId Error",
			zap.String("SpId", account.spId),
			zap.Uint32("SeqId", pkg.SeqId),
			zap.Error(err))
		statistics.CollectService.Service.AddPackerStatistics("Server", "Submit", false)
		return false, cmpp.ConnRspStatusErrMap[cmpp.ErrnoConnOthers]
	}
	resp.MsgId = msgId
	log.Logger.Info("[CmppServer][Cmpp3Submit] Success",
		zap.String("SpId", account.spId),
		zap.String("Phone", pkg.DestTerminalId[0]),
		zap.Uint16("SeqId", seqId),
		zap.Uint64("MsgId", msgId),
		zap.String("RemoteAddr", addr))
	statistics.CollectService.Service.AddPackerStatistics("Server", "Submit", true)
	statistics.CollectService.Service.AddPackerStatistics("Server", "SubmitResp", true)
	go sm.MockCmpp3Deliver(addr, account.spCode, msgId, pkg)
	return false, nil
}

// =====================CmppServer=====================

var TpUdhiSeq byte = 0x00

func (cm *CmppClientManager) SplitLongSms(content string) [][]byte {
	smsLength := 140
	smsHeaderLength := 6
	smsBodyLen := smsLength - smsHeaderLength
	contentBytes := []byte(content)
	var chunks [][]byte
	num := 1
	if (len(content)) > 140 {
		num = int(math.Ceil(float64(len(content)) / float64(smsBodyLen)))
	}
	if num == 1 {
		chunks = append(chunks, contentBytes)
		return chunks
	}
	tpUdhiHeader := []byte{0x05, 0x00, 0x03, TpUdhiSeq, byte(num)}
	TpUdhiSeq++

	for i := 0; i < num; i++ {
		chunk := tpUdhiHeader
		chunk = append(chunk, byte(i+1))
		smsBodyLen := smsLength - smsHeaderLength
		offset := i * smsBodyLen
		max := offset + smsBodyLen
		if max > len(content) {
			max = len(content)
		}

		chunk = append(chunk, contentBytes[offset:max]...)
		chunks = append(chunks, chunk)
	}
	return chunks
}

func GetMsgId(spId string, seqId uint16) (uint64, error) {
	now := time.Now()
	month, _ := strconv.ParseInt(fmt.Sprintf("%d", now.Month()), 10, 32)
	day := now.Day()
	hour := now.Hour()
	min := now.Minute()
	sec := now.Second()
	spIdInt, _ := strconv.ParseInt(spId, 10, 32)
	binaryStr := fmt.Sprintf("%04b%05b%05b%06b%06b%022b%016b", month, day, hour, min, sec, spIdInt, seqId)
	msgId, err := strconv.ParseUint(binaryStr, 2, 64)
	if err != nil {
		return 0, err
	}
	return msgId, nil
}
