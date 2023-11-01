package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetPR(consts.PRTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SEND_MESSAGE, &tgSendMessageAction{})
}

type tgSendMessageAction struct{}

func (t *tgSendMessageAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendmessageDetail().GetDetails()
	tgSendMsg := p.PrGetSyncMap(rpc.TG_SEND_MSG)
	for _, detail := range details {
		for userJid, sd := range detail.SendTgData {
			value, ok := p.Handler.GetServerMap().Load(userJid)
			ip := value.(string)
			if ok {
				existingData, _ := tgSendMsg.Load(ip)
				if existingData != nil {
					list := existingData.([]map[uint64]*protobuf.StringKeyStringvalue)
					tmp := make(map[uint64]*protobuf.StringKeyStringvalue)
					tmp[userJid] = sd
					list = append(list, tmp)
					tgSendMsg.Store(ip, list)
				} else {
					list := make([]map[uint64]*protobuf.StringKeyStringvalue, 0)
					tmp := make(map[uint64]*protobuf.StringKeyStringvalue)
					tmp[userJid] = sd
					list = append(list, tmp)
					tgSendMsg.Store(ip, list)
				}
			}
		}
	}
	return
}
