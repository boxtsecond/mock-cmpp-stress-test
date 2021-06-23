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
	statistics.CollectService.Service.AddPackerStatistics("DeliverResp", true)
	return nil
}

func (cm *CmppClientManager) Cmpp3DeliverReq(pkg *cmpp.Cmpp3DeliverReqPkt) error {
	log.Logger.Info("[CmppClient][Cmpp3DeliverReq] Success",
		zap.String("Addr", cm.Addr),
		zap.String("UserName", cm.UserName),
		zap.Any("Pkg", pkg))
	statistics.CollectService.Service.AddPackerStatistics("DeliverResp", true)
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
			if err := conn.SendPkt(pkg, <-conn.SeqId); err != nil {
				log.Logger.Error("[CmppServer][Cmpp2DeliverReq] Failed", zap.Error(err), zap.Uint64("MsgId", pkg.MsgId))
				statistics.CollectService.Service.AddPackerStatistics("Deliver", false)
				return err
			} else {
				log.Logger.Error("[CmppServer][Cmpp2DeliverReq] Success", zap.Uint64("MsgId", pkg.MsgId))
				statistics.CollectService.Service.AddPackerStatistics("Deliver", true)
			}
		}
	} else {
		statistics.CollectService.Service.AddPackerStatistics("Deliver", false)
	}
	return nil
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
				statistics.CollectService.Service.AddPackerStatistics("Deliver", false)
				return err
			} else {
				log.Logger.Error("[CmppServer][Cmpp3DeliverReq] Success", zap.Uint64("MsgId", pkg.MsgId))
				statistics.CollectService.Service.AddPackerStatistics("Deliver", true)
			}
		}

	} else {
		statistics.CollectService.Service.AddPackerStatistics("Deliver", false)
	}
	return nil
}

// =====================CmppServer=====================
