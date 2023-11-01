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
	im.RegisterAction(protobuf.Action_SEND_VCARD_MESSAGE, &whatsSendVcardMsgAction{})
}

type whatsSendVcardMsgAction struct{}

func (t *whatsSendVcardMsgAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	//修改
	msg := in.GetSendVcardMessage().GetDetails()
	wsSendVcardMsg := p.PrGetSyncMap(rpc.WS_SEND_VCARD_MSG)

	if msg != nil {
		for _, m := range msg {
			if sendData := m.GetSendData(); sendData != nil {
				for k, v := range sendData {
					value, ok := p.Handler.GetServerMap().Load(k)
					ip := value.(string)
					if ok {
						existingData, _ := wsSendVcardMsg.Load(ip)
						if existingData != nil {
							list := existingData.([]map[uint64]*protobuf.UintSenderVcard)
							tmp := make(map[uint64]*protobuf.UintSenderVcard)
							tmp[k] = v
							list = append(list, tmp)
							wsSendVcardMsg.Store(ip, list)
						} else {
							list := make([]map[uint64]*protobuf.UintSenderVcard, 0)
							tmp := make(map[uint64]*protobuf.UintSenderVcard)
							tmp[k] = v
							list = append(list, tmp)
							wsSendVcardMsg.Store(ip, list)
						}
					}
				}
			}
		}
	}
	return
}
