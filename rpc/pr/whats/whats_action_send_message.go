package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetPR(consts.PRWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SEND_MESSAGE, &whatsSendMsgAction{})
}

type whatsSendMsgAction struct{}

func (t *whatsSendMsgAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendmessageDetail().GetDetails()
	wsSendMsg := p.PrGetSyncMap(rpc.WS_SEND_MSG)
	for _, detail := range details {
		for userJid, sd := range detail.SendData {
			value, ok := p.Handler.GetServerMap().Load(userJid)
			ip := value.(string)
			if ok {
				existingData, _ := wsSendMsg.Load(ip)
				if existingData != nil {
					list := existingData.([]map[uint64]*protobuf.UintkeyStringvalue)
					tmp := make(map[uint64]*protobuf.UintkeyStringvalue)
					tmp[userJid] = sd
					list = append(list, tmp)
					wsSendMsg.Store(ip, list)
				} else {
					list := make([]map[uint64]*protobuf.UintkeyStringvalue, 0)
					tmp := make(map[uint64]*protobuf.UintkeyStringvalue)
					tmp[userJid] = sd
					list = append(list, tmp)
					wsSendMsg.Store(ip, list)
				}

			}
		}
	}
	return
}
