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
	im.RegisterAction(protobuf.Action_INVITE_TO_CHANNEL, &tgInviteToChannel{})
}

type tgInviteToChannel struct{}

func (t *tgInviteToChannel) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetInviteToChannelDetail()
	invite := p.PrGetSyncMap(rpc.TG_INVITE_TO_CHANNEL)
	value, ok := p.Handler.GetServerMap().Load(details.Detail.Key)
	if ok {
		ip := value.(string)
		existingData, _ := invite.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.InviteToChannelDetail)
			tmp[details.Detail.Key] = details
			invite.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.InviteToChannelDetail)
			tmp[details.Detail.Key] = details
			invite.Store(ip, tmp)
		}
	}
	return
}
