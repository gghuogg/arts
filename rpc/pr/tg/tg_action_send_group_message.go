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
	im.RegisterAction(protobuf.Action_SEND_GROUP_MESSAGE, &tgSendGroupMessage{})
}

type tgSendGroupMessage struct{}

func (t *tgSendGroupMessage) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	tgSendGroupMsg := p.PrGetSyncMap(rpc.TG_SEND_GROUP_MSG)
	details := in.GetSendGroupMessageDetail().GetDetails()
	for _, detail := range details {
		for userJid, sd := range detail.SendData {
			value, ok := p.Handler.GetServerMap().Load(userJid)
			ip := value.(string)
			if ok {
				existingData, _ := tgSendGroupMsg.Load(ip)
				if existingData != nil {
					list := existingData.([]map[uint64]*protobuf.StringKeyStringvalue)
					tmp := make(map[uint64]*protobuf.StringKeyStringvalue)
					tmp[userJid] = sd
					list = append(list, tmp)
					tgSendGroupMsg.Store(ip, tmp)
				} else {
					list := make([]map[uint64]*protobuf.StringKeyStringvalue, 0)
					tmp := make(map[uint64]*protobuf.StringKeyStringvalue)
					tmp[userJid] = sd
					list = append(list, tmp)
					tgSendGroupMsg.Store(ip, list)
				}

			}
		}
	}
	return
}
