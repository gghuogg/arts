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
	im.RegisterAction(protobuf.Action_GET_USER_HEAD_IMAGE, &whatsGetUserHeadImageAction{})
}

type whatsGetUserHeadImageAction struct{}

func (t *whatsGetUserHeadImageAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	msg := in.GetGetUserHeadImage().GetHeadImage()
	getHeadImageMap := p.PrGetSyncMap(rpc.WS_GET_USER_HEAD_IMAGE)
	if msg != nil {
		for userJid, v := range msg {
			value, ok := p.Handler.GetServerMap().Load(userJid)
			if ok {
				ip := value.(string)
				existingData, _ := getHeadImageMap.Load(ip)
				if existingData != nil {
					tmp := existingData.(map[uint64]*protobuf.GetUserHeadImage)
					tmp[userJid] = v
					getHeadImageMap.Store(ip, tmp)
				} else {
					tmp := make(map[uint64]*protobuf.GetUserHeadImage)
					tmp[userJid] = v
					getHeadImageMap.Store(ip, tmp)
				}
			}
		}

	}
	return
}
