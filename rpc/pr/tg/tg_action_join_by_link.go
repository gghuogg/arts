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
	im.RegisterAction(protobuf.Action_JOIN_BY_LINK, &tgJoinByLink{})
}

type tgJoinByLink struct{}

func (t *tgJoinByLink) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetJoinByLinkDetail()
	join := p.PrGetSyncMap(rpc.TG_JOIN_BY_LINK)
	value, ok := p.Handler.GetServerMap().Load(details.Detail.Key)
	if ok {
		ip := value.(string)
		existingData, _ := join.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.JoinByLinkDetail)
			tmp[details.Detail.Key] = details
			join.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.JoinByLinkDetail)
			tmp[details.Detail.Key] = details
			join.Store(ip, tmp)
		}
	}
	return
}
