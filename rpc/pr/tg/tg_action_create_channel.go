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
	im.RegisterAction(protobuf.Action_CREATE_CHANNEL, &tgCreateChannel{})
}

type tgCreateChannel struct{}

func (t *tgCreateChannel) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetCreateChannelDetail()
	createGroup := p.PrGetSyncMap(rpc.TG_CREATE_CHANNEL)
	value, ok := p.Handler.GetServerMap().Load(details.Detail.Key)
	if ok {
		ip := value.(string)
		existingData, _ := createGroup.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.CreateChannelDetail)
			tmp[details.Detail.Key] = details
			createGroup.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.CreateChannelDetail)
			tmp[details.Detail.Key] = details
			createGroup.Store(ip, tmp)
		}
	}
	return
}
