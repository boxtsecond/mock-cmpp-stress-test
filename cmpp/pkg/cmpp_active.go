package pkg

import (
	cmpp "github.com/bigwhite/gocmpp"
)

// =====================CmppClient=====================
func (cm *CmppClientManager) CmppActiveTestReq(pkg *cmpp.CmppActiveTestReqPkt) error {
	return cm.Client.SendRspPkt(&cmpp.CmppActiveTestRspPkt{}, pkg.SeqId)
}

func (cm *CmppClientManager) CmppActiveTestRsp(pkg *cmpp.CmppActiveTestRspPkt) error {
	return nil
}

// =====================CmppClient=====================
