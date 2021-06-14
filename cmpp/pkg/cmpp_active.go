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
	 _ ,err := cm.Client.SendReqPkt(&cmpp.CmppActiveTestReqPkt{})
	return err
}


// =====================CmppClient=====================


// =====================CmppServer=====================
// gocmpp都处理了不需要做啥了
// 处理来自客户端的心跳请求
//func (csm *CmppServerManager) CmppActiveTestReq (req *cmpp.Packet, res *cmpp.Response ) (bool, error) {
//	fmt.Println("1111", "收了客户端的心跳检测请求")
//	resp := res.Packer.(*cmpp.CmppActiveTestRspPkt)
//	err := res.Packet.Conn.SendPkt(resp,)
//	return  false ,  err
//}
//
//// 处理客户端回复的心跳包
//func (csm *CmppServerManager) CmppActiveTestRsp (req *cmpp.Packet , res *cmpp.Response) (bool ,error) {
//	fmt.Println("2222", "收到了心跳回复包" , zap.Any(req) , zap.Any(res))
//	return false , nil
//}
// =====================CmppServer=====================


