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
	im.RegisterAction(protobuf.Action_GET_CHANNEL_MEMBER, &tgGetChannelMember{})
}

type tgGetChannelMember struct{}

func (t *tgGetChannelMember) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetChannelMemberDetail()
	addMember := p.PrGetSyncMap(rpc.TG_GET_CHANNEL_MEMBERS)
	value, ok := p.Handler.GetServerMap().Load(details.Sender)
	if ok {
		ip := value.(string)
		existingData, _ := addMember.Load(ip)
		if existingData != nil {
			tmp := existingData.(*protobuf.GetChannelMemberDetail)
			tmp = details
			addMember.Store(ip, tmp)
		} else {
			tmp := new(protobuf.GetChannelMemberDetail)
			tmp = details
			addMember.Store(ip, tmp)
		}
	}
	return
}
