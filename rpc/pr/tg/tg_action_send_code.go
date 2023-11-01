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
	im.RegisterAction(protobuf.Action_SEND_CODE, &tgSendCode{})
}

type tgSendCode struct{}

func (t *tgSendCode) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetSendCodeDetail().GetDetails()
	if detail != nil {
		if len(detail.GetSendCode()) > 0 {
			for k, v := range detail.SendCode {
				addMember := p.PrGetSyncMap(rpc.TG_SEND_CODE)
				value, ok := p.Handler.GetServerMap().Load(k)
				if ok {
					ip := value.(string)
					existingData, _ := addMember.Load(ip)

					if existingData != nil {
						tmp := existingData.(*protobuf.SendCodeAction)
						codeMap := make(map[uint64]string)
						codeMap[k] = v
						tmp = &protobuf.SendCodeAction{
							SendCode: codeMap,
							LoginId:  detail.GetLoginId(),
						}
						addMember.Store(ip, tmp)
					} else {
						tmp := &protobuf.SendCodeAction{}
						codeMap := make(map[uint64]string)
						codeMap[k] = v
						tmp.SendCode = codeMap
						tmp.LoginId = detail.GetLoginId()
						addMember.Store(ip, tmp)
					}
				}
			}
		}
	}
	return
}
