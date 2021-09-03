package pkg

import (
	"encoding/binary"
	"errors"
	cmpp "github.com/bigwhite/gocmpp"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/statistics"
	"mock-cmpp-stress-test/utils/buf"
	"mock-cmpp-stress-test/utils/log"
	"runtime"
	"time"
)

// =====================CmppClient=====================

func (cm *CmppClientManager) Cmpp2DeliverReq(pkg *cmpp.Cmpp2DeliverReqPkt) error {
	log.Logger.Info("[CmppClient][Cmpp2DeliverReq] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.Any("Pkg", pkg))

	statistics.CollectService.Service.AddPackerStatistics("Client", "Deliver", true)
	statistics.CollectService.Service.AddPackerStatistics("Client", "DeliverResp", true)
	return cm.Client.SendRspPkt(&cmpp.Cmpp2DeliverRspPkt{
		MsgId:  pkg.MsgId,
		Result: 0,
	}, pkg.SeqId)
}

func (cm *CmppClientManager) Cmpp3DeliverReq(pkg *cmpp.Cmpp3DeliverReqPkt) error {
	log.Logger.Info("[CmppClient][Cmpp3DeliverReq] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.Any("Pkg", pkg))
	statistics.CollectService.Service.AddPackerStatistics("Client", "Deliver", true)
	statistics.CollectService.Service.AddPackerStatistics("Client", "DeliverResp", true)
	return cm.Client.SendRspPkt(&cmpp.Cmpp3DeliverRspPkt{
		MsgId:  pkg.MsgId,
		Result: 0,
	}, pkg.SeqId)
}

func (cm *CmppClientManager) BatchCmpp2Submit(pkgs []*cmpp.Cmpp2SubmitReqPkt) {
	if !cm.Connected {
		return
	}
	for _, each := range pkgs {
		cm.Cmpp2SubmitPkg(each)
	}
}

func (cm *CmppClientManager) Cmpp2SubmitPkg(pkg *cmpp.Cmpp2SubmitReqPkt) {
	if !cm.Connected {
		statistics.CollectService.Service.AddPackerStatistics("Client", "Submit", false)
		return
	}
	// 让出 CPU 资源
	runtime.Gosched()
	seqId, sendErr := cm.Client.SendReqPkt(pkg)
	phone := pkg.DestTerminalId[0]
	if sendErr != nil {
		cm.ConnErrCount += 1
		log.Logger.Error("[CmppClient][Cmpp2Submit] Error",
			zap.String("Addr", cm.Addr),
			zap.String("UserName", cm.UserName),
			zap.String("SpId", cm.SpId),
			zap.String("SpCode", cm.SpCode),
			zap.String("Phone", phone),
			zap.Error(sendErr))
		statistics.CollectService.Service.AddPackerStatistics("Client", "Submit", false)
		return
	}

	statistics.CollectService.Service.AddPackerStatistics("Client", "Submit", true)
	log.Logger.Info("[CmppClient][Cmpp2Submit] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.String("SpId", cm.SpId),
		zap.String("SpCode", cm.SpCode),
		zap.String("Phone", phone),
		zap.Any("SeqId", seqId))
}

func (cm *CmppClientManager) BatchCmpp3Submit(pkgs []*cmpp.Cmpp3SubmitReqPkt) {
	for _, each := range pkgs {
		cm.Cmpp3SubmitPkg(each)
	}
}

func (cm *CmppClientManager) Cmpp3SubmitPkg(pkg *cmpp.Cmpp3SubmitReqPkt) {
	if !cm.Connected {
		statistics.CollectService.Service.AddPackerStatistics("Client", "Submit", false)
		return
	}
	// 让出 CPU 资源
	runtime.Gosched()
	seqId, sendErr := cm.Client.SendReqPkt(pkg)
	if sendErr != nil {
		cm.ConnErrCount += 1
		log.Logger.Error("[CmppClient][Cmpp3Submit] Error", zap.Error(sendErr))
		statistics.CollectService.Service.AddPackerStatistics("Client", "Submit", false)
		return
	}
	phone := pkg.DestTerminalId[0]

	log.Logger.Info("[CmppClient][Cmpp3Submit] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.String("SpId", cm.SpId),
		zap.String("SpCode", cm.SpCode),
		zap.String("Phone", phone),
		zap.Any("SeqId", seqId))
	statistics.CollectService.Service.AddPackerStatistics("Client", "Submit", true)
}

// =====================CmppClient=====================

// =====================CmppServer=====================
type MockCmpp2DeliverPkg struct {
	addr string
	p    *cmpp.Cmpp2DeliverReqPkt
}

type MockCmpp3DeliverPkg struct {
	addr string
	p    *cmpp.Cmpp3DeliverReqPkt
}

var Cmpp2DeliverChan = make(chan *MockCmpp2DeliverPkg, 500)
var Cmpp3DeliverChan = make(chan *MockCmpp3DeliverPkg, 500)

func (sm *CmppServerManager) MockCmpp2Deliver(addr, spCode string, msgId uint64, pkg *cmpp.Cmpp2SubmitReqPkt) {
	// 构造一个回执
	stat := "DELIVRD"
	deliverPkg := &cmpp.Cmpp2DeliverReqPkt{
		MsgId:            msgId,
		DestId:           spCode,
		ServiceId:        "",
		TpPid:            0,
		TpUdhi:           0,
		MsgFmt:           0,
		SrcTerminalId:    pkg.DestTerminalId[0],
		RegisterDelivery: 1,
		Reserve:          "",
	}
	submitTime := time.Unix(time.Now().Unix(), 0).Format("0601021504")
	doneTime := submitTime
	msgContent := formatReportMsgContent("V20", msgId, stat, submitTime, doneTime, deliverPkg.SrcTerminalId, uint32(1))

	deliverPkg.MsgContent = msgContent
	deliverPkg.MsgLength = uint8(len(msgContent))

	// 返回状态报告
	sm.SendCmpp2DeliverPkg(deliverPkg, addr)
}

func (sm *CmppServerManager) SendCmpp2DeliverPkg(pkg *cmpp.Cmpp2DeliverReqPkt, addr string) {
	Cmpp2DeliverChan <- &MockCmpp2DeliverPkg{
		addr: addr,
		p:    pkg,
	}
}

func (sm *CmppServerManager) BatchCmpp2Deliver(pkgs []*MockCmpp2DeliverPkg) {
	for _, each := range pkgs {
		go sm.Cmpp2Deliver(each)
	}
}

func (sm *CmppServerManager) MockCmpp3Deliver(addr, spCode string, msgId uint64, pkg *cmpp.Cmpp3SubmitReqPkt) {
	// 构造一个回执
	stat := "DELIVRD"
	deliverPkg := &cmpp.Cmpp3DeliverReqPkt{
		MsgId:            msgId,
		DestId:           spCode,
		ServiceId:        "",
		TpPid:            0,
		TpUdhi:           0,
		MsgFmt:           0,
		SrcTerminalId:    pkg.DestTerminalId[0],
		RegisterDelivery: 1,
	}
	submitTime := time.Unix(time.Now().Unix(), 0).Format("0601021504")
	doneTime := submitTime
	msgContent := formatReportMsgContent("V30", msgId, stat, submitTime, doneTime, deliverPkg.SrcTerminalId, uint32(1))

	deliverPkg.MsgContent = msgContent
	deliverPkg.MsgLength = uint8(len([]rune(msgContent)))

	// 返回状态报告
	go sm.SendCmpp3DeliverPkg(deliverPkg, addr)
}

func (sm *CmppServerManager) SendCmpp3DeliverPkg(pkg *cmpp.Cmpp3DeliverReqPkt, addr string) {
	Cmpp3DeliverChan <- &MockCmpp3DeliverPkg{
		addr: addr,
		p:    pkg,
	}
}

func (sm *CmppServerManager) BatchCmpp3Deliver(pkgs []*MockCmpp3DeliverPkg) {
	for _, each := range pkgs {
		go sm.Cmpp3Deliver(each)
	}
}

// 推送回执给指定连接
func (sm *CmppServerManager) Cmpp2Deliver(pkg *MockCmpp2DeliverPkg) {
	if c, ok := sm.ConnMap.Load(pkg.addr); ok {
		conn := c.(*cmpp.Packet)
		seqId := <-conn.Conn.SeqId
		if err := conn.Conn.SendPkt(pkg.p, seqId); err != nil {
			log.Logger.Error("[CmppServer][Cmpp2DeliverReq] Failed",
				zap.Error(err),
				zap.Uint64("MsgId", pkg.p.MsgId),
				zap.Uint32("SeqId", seqId),
				zap.String("Addr", pkg.addr))
			statistics.CollectService.Service.AddPackerStatistics("Server", "Deliver", false)
			return
		} else {
			log.Logger.Info("[CmppServer][Cmpp2DeliverReq] Success",
				zap.Uint64("MsgId", pkg.p.MsgId),
				zap.Uint32("SeqId", seqId),
				zap.String("Addr", pkg.addr))
			statistics.CollectService.Service.AddPackerStatistics("Server", "Deliver", true)
			return
		}
	} else {
		err := errors.New("can't find connection")
		log.Logger.Error("[CmppServer][Cmpp2DeliverReq] Failed",
			zap.Error(err),
			zap.Uint64("MsgId", pkg.p.MsgId),
			zap.String("Addr", pkg.addr))
		statistics.CollectService.Service.AddPackerStatistics("Server", "Deliver", false)
		return
	}
}

func (sm *CmppServerManager) Cmpp2DeliverResp(pkg *cmpp.Cmpp2DeliverRspPkt, res *cmpp.Response) (bool, error) {
	log.Logger.Info("[CmppServer][Cmpp2DeliverResp] Success", zap.Uint64("MsgId", pkg.MsgId), zap.Uint32("SeqId", pkg.SeqId))
	statistics.CollectService.Service.AddPackerStatistics("Server", "DeliverResp", true)
	return false, nil
}

// 推送回执给指定连接
func (sm *CmppServerManager) Cmpp3Deliver(pkg *MockCmpp3DeliverPkg) {
	if c, ok := sm.ConnMap.Load(pkg.addr); ok {
		conn := c.(*cmpp.Packet)
		seqId := <-conn.Conn.SeqId
		if err := conn.Conn.SendPkt(pkg.p, seqId); err != nil {
			log.Logger.Error("[CmppServer][Cmpp3DeliverReq] Failed", zap.Error(err), zap.Uint64("MsgId", pkg.p.MsgId), zap.Uint32("SeqId", seqId),
				zap.String("Addr", pkg.addr))
			statistics.CollectService.Service.AddPackerStatistics("Server", "Deliver", false)
			return
		} else {
			log.Logger.Info("[CmppServer][Cmpp3DeliverReq] Success", zap.Uint64("MsgId", pkg.p.MsgId), zap.Uint32("SeqId", seqId),
				zap.String("Addr", pkg.addr))
			statistics.CollectService.Service.AddPackerStatistics("Server", "Deliver", true)
			return
		}
	} else {
		err := errors.New("can't find connection")
		log.Logger.Error("[CmppServer][Cmpp3DeliverReq] Failed",
			zap.Error(err),
			zap.Uint64("MsgId", pkg.p.MsgId),
			zap.String("Addr", pkg.addr))
		statistics.CollectService.Service.AddPackerStatistics("Server", "Deliver", false)
		return
	}
}

func (sm *CmppServerManager) Cmpp3DeliverResp(pkg *cmpp.Cmpp3DeliverRspPkt, res *cmpp.Response) (bool, error) {
	log.Logger.Info("[CmppServer][Cmpp3DeliverResp] Success", zap.Uint64("MsgId", pkg.MsgId), zap.Uint32("SeqId", pkg.SeqId))
	statistics.CollectService.Service.AddPackerStatistics("Server", "DeliverResp", true)
	return false, nil
}

// Cmpp2状态报告消息内容
type Cmpp2StatsReportMsgContent struct {
	MsgId          uint64
	Stat           string
	SubmitTime     string
	DoneTime       string
	DestTerminalId string
	SmscSequence   uint32
}

// Cmpp3状态报告消息内容
type Cmpp3StatsReportMsgContent struct {
	MsgId          uint64
	Stat           string
	SubmitTime     string
	DoneTime       string
	DestTerminalId string
	SmscSequence   uint32
}

func (p *Cmpp2StatsReportMsgContent) Encode() (string, error) {
	var pkgLen uint32 = 8 + 7 + 10 + 10 + 21 + 4

	w := buf.NewBufWriter(pkgLen)
	w.WriteInt(p.MsgId, 0, binary.BigEndian)
	w.WriteFixedSizeString(p.Stat, 7)
	w.WriteFixedSizeString(p.SubmitTime, 10)
	w.WriteFixedSizeString(p.DoneTime, 10)
	w.WriteFixedSizeString(p.DestTerminalId, 21)
	w.WriteInt(p.SmscSequence, 0, binary.BigEndian)

	b, err := w.Bytes()
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (p *Cmpp3StatsReportMsgContent) Encode() (string, error) {
	var pkgLen uint32 = 8 + 7 + 10 + 10 + 21 + 4

	w := buf.NewBufWriter(pkgLen)
	w.WriteInt(p.MsgId, 0, binary.BigEndian)
	w.WriteFixedSizeString(p.Stat, 7)
	w.WriteFixedSizeString(p.SubmitTime, 10)
	w.WriteFixedSizeString(p.DoneTime, 10)
	w.WriteFixedSizeString(p.DestTerminalId, 32)
	w.WriteInt(p.SmscSequence, 0, binary.BigEndian)

	b, err := w.Bytes()
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func formatReportMsgContent(version string, msgId uint64, stat string, submitTime string, doneTime, destTerminalId string, smscSeq uint32) string {
	var err error
	var content string
	if version == "V20" || version == "V21" {
		msg := &Cmpp2StatsReportMsgContent{
			MsgId:          msgId,
			Stat:           stat,
			SubmitTime:     submitTime,
			DoneTime:       doneTime,
			DestTerminalId: destTerminalId,
			SmscSequence:   smscSeq,
		}
		content, err = msg.Encode()
	} else {
		msg := &Cmpp3StatsReportMsgContent{
			MsgId:          msgId,
			Stat:           stat,
			SubmitTime:     submitTime,
			DoneTime:       doneTime,
			DestTerminalId: destTerminalId,
			SmscSequence:   smscSeq,
		}
		content, err = msg.Encode()
	}

	if err != nil {
		log.Logger.Error("format report msg content failed",
			zap.String("Error", err.Error()),
			zap.Uint64("MsgId", msgId),
			zap.String("Phone", destTerminalId))
	}

	return content
}

// =====================CmppServer=====================
