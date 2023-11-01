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
	im.RegisterAction(protobuf.Action_Get_MSG_HISTORY, &tgMsgHistory{})
}

type tgMsgHistory struct{}

func (t *tgMsgHistory) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetGetMsgHistory()
	tgMsgHistory := p.PrGetSyncMap(rpc.TG_GET_MSG_HISTORY)
	value, ok := p.Handler.GetServerMap().Load(detail.Self)
	if ok {
		ip := value.(string)
		existingData, _ := tgMsgHistory.Load(ip)
		if existingData != nil {
			list := existingData.([]*protobuf.GetMsgHistory)
			tmp := &protobuf.GetMsgHistory{
				Self:      detail.Self,
				Other:     detail.Other,
				Limit:     detail.Limit,
				OffsetDat: detail.OffsetDat,
				OffsetID:  detail.OffsetID,
				MaxID:     detail.MaxID,
				MinID:     detail.MinID,
			}
			list = append(list, tmp)
			tgMsgHistory.Store(ip, list)
		} else {
			list := make([]*protobuf.GetMsgHistory, 0)
			tmp := &protobuf.GetMsgHistory{
				Self:      detail.Self,
				Other:     detail.Other,
				Limit:     detail.Limit,
				OffsetDat: detail.OffsetDat,
				OffsetID:  detail.OffsetID,
				MaxID:     detail.MaxID,
				MinID:     detail.MinID,
			}
			list = append(list, tmp)
			tgMsgHistory.Store(ip, list)
		}
	}
	return
}
