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
	im.RegisterAction(protobuf.Action_GET_GROUP_MEMBERS, &tgGetGroupMember{})
}

type tgGetGroupMember struct{}

func (t *tgGetGroupMember) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetGroupMembersDetail()
	addMember := p.PrGetSyncMap(rpc.TG_GET_GROUP_MEMBERS)
	value, ok := p.Handler.GetServerMap().Load(details.Account)
	if ok {
		ip := value.(string)
		existingData, _ := addMember.Load(ip)
		if existingData != nil {
			tmp := existingData.(*protobuf.GetGroupMembersDetail)
			tmp = details
			addMember.Store(ip, tmp)
		} else {
			tmp := new(protobuf.GetGroupMembersDetail)
			tmp = details
			addMember.Store(ip, tmp)
		}
	}
	return
}
