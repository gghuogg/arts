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
	im.RegisterAction(protobuf.Action_SEND_VIDEO, &whatsSendVideoMsg{})
}

type whatsSendVideoMsg struct{}

func (t *whatsSendVideoMsg) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendVideoDetail().GetDetails()

	if len(details) > 0 {
		for _, detail := range details {
			for k, v := range detail.SendData {
				addMember := p.PrGetSyncMap(rpc.WS_SEND_VIDEO)
				value, ok := p.Handler.GetServerMap().Load(k)
				if ok {
					ip := value.(string)
					existingData, _ := addMember.Load(ip)
					if existingData != nil {
						tmp := existingData.(map[uint64]*protobuf.UintFileDetailValue)
						tmp[k] = v
						addMember.Store(ip, tmp)
					} else {
						tmp := make(map[uint64]*protobuf.UintFileDetailValue)
						tmp[k] = v
						addMember.Store(ip, tmp)
					}
				}
			}
		}
	}
	return
}
