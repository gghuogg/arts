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
	im.RegisterAction(protobuf.Action_SEND_VCARD_MESSAGE, &tgSendMessageVcardMsg{})
}

type tgSendMessageVcardMsg struct{}

func (t *tgSendMessageVcardMsg) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendContactCardDetail().GetDetail()
	tgSendVcardMsg := p.PrGetSyncMap(rpc.TG_SEND_VCARD)
	for _, detail := range details {
		for k, v := range detail.SendData {
			value, ok := p.Handler.GetServerMap().Load(k)
			ip := value.(string)
			if ok {
				existingData, _ := tgSendVcardMsg.Load(ip)
				if existingData != nil {
					list := existingData.([]map[uint64]*protobuf.UintSendContactCard)
					tmp := make(map[uint64]*protobuf.UintSendContactCard)
					tmp[k] = v
					list = append(list, tmp)
					tgSendVcardMsg.Store(ip, list)
				} else {
					list := make([]map[uint64]*protobuf.UintSendContactCard, 0)
					tmp := make(map[uint64]*protobuf.UintSendContactCard)
					tmp[k] = v
					list = append(list, tmp)
					tgSendVcardMsg.Store(ip, list)
				}
			}
		}

	}
	return
}
