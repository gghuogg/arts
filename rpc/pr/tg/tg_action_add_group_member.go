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
	im.RegisterAction(protobuf.Action_ADD_GROUP_MEMBER, &tgCreateGroupMember{})
}

type tgCreateGroupMember struct{}

func (t *tgCreateGroupMember) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetAddGroupMemberDetail()
	addMember := p.PrGetSyncMap(rpc.TG_ADD_GROUP_MEMBER)
	value, ok := p.Handler.GetServerMap().Load(details.Detail.Key)
	if ok {
		ip := value.(string)
		existingData, _ := addMember.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.AddGroupMemberDetail)
			tmp[details.Detail.Key] = details
			addMember.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.AddGroupMemberDetail)
			tmp[details.Detail.Key] = details
			addMember.Store(ip, tmp)
		}
	}
	return
}
