package pkg

import (
	cmpp "github.com/bigwhite/gocmpp"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/statistics"
	"strings"
	//"golang.org/x/text/width"
	"mock-cmpp-stress-test/utils/log"
	"strconv"
	//"net"
)

// =====================CmppClient=====================

func (cm *CmppClientManager) Cmpp2DeliverReq(pkg *cmpp.Cmpp2DeliverReqPkt) error {
	log.Logger.Info("[CmppClient][Cmpp2DeliverReq] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.Any("Pkg", pkg))
	statistics.CollectService.Service.AddPackerStatistics("Deliver", true)
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
	statistics.CollectService.Service.AddPackerStatistics("Deliver", true)
	return cm.Client.SendRspPkt(&cmpp.Cmpp3DeliverRspPkt{
		MsgId:  pkg.MsgId,
		Result: 0,
	}, pkg.SeqId)
}

func (cm *CmppClientManager) BatchCmpp2Submit(pkgs []*cmpp.Cmpp2SubmitReqPkt) {
	for _, each := range pkgs {
		cm.Cmpp2SubmitPkg(each)
	}
}

func (cm *CmppClientManager) Cmpp2SubmitPkg(pkg *cmpp.Cmpp2SubmitReqPkt) error {
	seqId, sendErr := cm.Client.SendReqPkt(pkg)
	if sendErr != nil {
		log.Logger.Error("[CmppClient][Cmpp2Submit] Error", zap.Error(sendErr))
		statistics.CollectService.Service.AddPackerStatistics("Submit", false)
		return sendErr
	}
	phone := pkg.DestTerminalId[0]
	setCacheErr := cm.Cache.Set(strconv.Itoa(int(seqId)), strings.Join([]string{cm.Addr, cm.UserName, cm.SpId, cm.SpCode, phone}, ","))
	if setCacheErr != nil {
		log.Logger.Error("[CmppClient][Cmpp2Submit][SetCache] Error:", zap.Error(setCacheErr))
		statistics.CollectService.Service.AddPackerStatistics("Submit", false)
		return setCacheErr
	}
	statistics.CollectService.Service.AddPackerStatistics("Submit", true)
	log.Logger.Info("[CmppClient][Cmpp2Submit] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.String("SpId", cm.SpId),
		zap.String("SpCode", cm.SpCode),
		zap.String("Phone", phone),
		zap.Any("SeqId", seqId))
	return nil
}

func (cm *CmppClientManager) BatchCmpp3Submit(pkgs []*cmpp.Cmpp3SubmitReqPkt) {
	for _, each := range pkgs {
		cm.Cmpp3SubmitPkg(each)
	}
}

func (cm *CmppClientManager) Cmpp3SubmitPkg(pkg *cmpp.Cmpp3SubmitReqPkt) error {
	seqId, sendErr := cm.Client.SendReqPkt(pkg)
	if sendErr != nil {
		log.Logger.Error("[CmppClient][Cmpp3Submit] Error", zap.Error(sendErr))
		statistics.CollectService.Service.AddPackerStatistics("Submit", false)
		return sendErr
	}
	phone := pkg.DestTerminalId[0]
	setCacheErr := cm.Cache.Set(strconv.Itoa(int(seqId)), strings.Join([]string{cm.Addr, cm.UserName, cm.SpId, cm.SpCode, phone}, ","))
	if setCacheErr != nil {
		log.Logger.Error("[CmppClient][Cmpp3Submit][SetCache] Error:", zap.Error(setCacheErr))
		statistics.CollectService.Service.AddPackerStatistics("Submit", false)
		return setCacheErr
	}

	log.Logger.Info("[CmppClient][Cmpp3Submit] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.String("SpId", cm.SpId),
		zap.String("SpCode", cm.SpCode),
		zap.String("Phone", phone),
		zap.Any("SeqId", seqId))
	statistics.CollectService.Service.AddPackerStatistics("Submit", true)
	return nil
}

// =====================CmppClient=====================

// =====================CmppServer=====================
func (sm *CmppServerManager) BatchCmpp2Deliver(pkgs []*cmpp.Cmpp2DeliverReqPkt) {
	for _, each := range pkgs {
		sm.Cmpp2Deliver(each)
	}
}

func (sm *CmppServerManager) BatchCmpp3Deliver(pkgs []*cmpp.Cmpp3DeliverReqPkt) {
	for _, each := range pkgs {
		sm.Cmpp3Deliver(each)
	}
}

// 推送回执给指定连接
func (sm *CmppServerManager) Cmpp2Deliver(pkg *cmpp.Cmpp2DeliverReqPkt) error {
	key := strconv.Itoa(int(pkg.MsgId))
	value := sm.Cache.Get(key)
	addr := strings.Split(value, ",")[0]
	defer sm.Cache.Delete(key)
	if addr != "" {
		if conn, ok := sm.ConnMap[addr]; ok {
			seqId := <-conn.SeqId
			if err := conn.SendPkt(pkg, seqId); err != nil {
				log.Logger.Error("[CmppServer][Cmpp2DeliverReq] Failed", zap.Error(err), zap.Uint64("MsgId", pkg.MsgId), zap.Uint32("SeqId", seqId))
				return err
			} else {
				log.Logger.Info("[CmppServer][Cmpp2DeliverReq] Success", zap.Uint64("MsgId", pkg.MsgId), zap.Uint32("SeqId", seqId))
				return nil
			}
		}
	} else {
		log.Logger.Error("[CmppServer][Cmpp2DeliverReq] Error", zap.Uint64("MsgId", pkg.MsgId))
	}
	return nil
}

func (sm *CmppServerManager) Cmpp2DeliverResp(pkg *cmpp.Cmpp2DeliverRspPkt, res *cmpp.Response) (bool, error) {
	log.Logger.Info("[CmppServer][Cmpp2DeliverResp] Success", zap.Uint64("MsgId", pkg.MsgId), zap.Uint32("SeqId", pkg.SeqId))
	statistics.CollectService.Service.AddPackerStatistics("DeliverResp", true)
	return false, nil
}

// 推送回执给指定连接
func (sm *CmppServerManager) Cmpp3Deliver(pkg *cmpp.Cmpp3DeliverReqPkt) error {
	key := strconv.Itoa(int(pkg.MsgId))
	value := sm.Cache.Get(key)
	addr := strings.Split(value, ",")[0]
	defer sm.Cache.Delete(key)
	if addr != "" {
		if conn, ok := sm.ConnMap[addr]; ok {
			if err := conn.SendPkt(pkg, <-conn.SeqId); err != nil {
				log.Logger.Error("[CmppServer][Cmpp3DeliverReq] Failed", zap.Error(err), zap.Uint64("MsgId", pkg.MsgId))
				return err
			} else {
				log.Logger.Info("[CmppServer][Cmpp3DeliverReq] Success", zap.Uint64("MsgId", pkg.MsgId))
			}
		}
	} else {
		log.Logger.Error("[CmppServer][Cmpp3DeliverReq] Error", zap.Uint64("MsgId", pkg.MsgId))
	}
	return nil
}

func (sm *CmppServerManager) Cmpp3DeliverResp(pkg *cmpp.Cmpp3DeliverRspPkt, res *cmpp.Response) (bool, error) {
	log.Logger.Info("[CmppServer][Cmpp3DeliverResp] Success", zap.Uint64("MsgId", pkg.MsgId), zap.Uint32("SeqId", pkg.SeqId))
	statistics.CollectService.Service.AddPackerStatistics("DeliverResp", true)
	return false, nil
}

// =====================CmppServer=====================
