package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/telegram"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetIM(consts.IMTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_Get_MSG_HISTORY, &tgGetMsgHistoryAction{})
}

type tgGetMsgHistoryAction struct{}

func (t *tgGetMsgHistoryAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	result = &protobuf.ResponseMessage{}
	detail := in.GetGetMsgHistory()
	if detail != nil {
		tmp := telegram.GetHistoryDetail{
			Self:       detail.Self,
			Other:      detail.Other,
			Limit:      int(detail.Limit),
			OffsetDate: int(detail.OffsetDat),
			OffsetID:   int(detail.OffsetID),
			MaxID:      int(detail.MaxID),
			MinID:      int(detail.MinID),
			Res:        make(chan *protobuf.ResponseMessage),
		}
		h.GetTgMsgHistoryChan() <- tmp
		result = <-tmp.Res

		return result, nil
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "send protobuf msg is nil or having err "}

	return
}
