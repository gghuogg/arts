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
	im.RegisterAction(protobuf.Action_LEAVE, &tgLeave{})
}

type tgLeave struct{}

func (t *tgLeave) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetLeaveDetail()
	leave := p.PrGetSyncMap(rpc.TG_LEAVE)
	value, ok := p.Handler.GetServerMap().Load(details.Detail.Key)
	if ok {
		ip := value.(string)
		existingData, _ := leave.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.LeaveDetail)
			tmp[details.Detail.Key] = details
			leave.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.LeaveDetail)
			tmp[details.Detail.Key] = details
			leave.Store(ip, tmp)
		}
	}
	return
}
