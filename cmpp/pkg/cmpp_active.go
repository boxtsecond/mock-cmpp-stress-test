package pkg

import (
	"fmt"
	cmpp "github.com/bigwhite/gocmpp"
)

// =====================CmppClient=====================
func (cm *CmppClientManager) CmppActiveTestReq(pkg *cmpp.CmppActiveTestReqPkt) error {
	return cm.Client.SendRspPkt(&cmpp.CmppActiveTestRspPkt{}, pkg.SeqId)
}

func (cm *CmppClientManager) CmppActiveTestRsp(pkg *cmpp.CmppActiveTestRspPkt) error {
	fmt.Println("111" ,"我收到了心跳包")
	return nil
}

// =====================CmppClient=====================


// =====================CmppServer=====================
// 回复心跳包
func (csm *CmppServerManager) DealCmppActiveTestReq (req *cmpp.Packet, res *cmpp.Response) (bool, error) {
	return false , nil
}
// =====================CmppServer=====================


