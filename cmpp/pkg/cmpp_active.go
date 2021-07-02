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

// 客户端发送心跳包
func (cm *CmppClientManager) SendCmppActiveTestReq(pkg *cmpp.CmppActiveTestReqPkt) error {
	_, err := cm.Client.SendReqPkt(&cmpp.CmppActiveTestReqPkt{})
	return err
}

// =====================CmppClient=====================

// =====================CmppServer=====================
func (sm *CmppServerManager) CmppActiveTestReq(pkg *cmpp.CmppActiveTestReqPkt, res *cmpp.Response) (bool, error) {
	return false, nil
}

// =====================CmppServer=====================
