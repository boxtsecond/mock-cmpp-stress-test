package pkg

import (
	cmpp "github.com/bigwhite/gocmpp"
)

// =====================CmppClient=====================
func (cm *CmppClientManager) CmppTerminateReq(pkg *cmpp.CmppTerminateReqPkt) error {
	err := cm.Client.SendRspPkt(&cmpp.CmppTerminateRspPkt{}, pkg.SeqId)
	if err != nil {
		return err
	}
	cm.Client.Disconnect()
	return nil
}

func (cm *CmppClientManager) CmppTerminateRsp(pkg *cmpp.CmppTerminateRspPkt) error {
	return nil
}
