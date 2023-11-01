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
	im.RegisterAction(protobuf.Action_SEND_GROUP_MESSAGE, &tgSendGroupMessageAction{})
}

type tgSendGroupMessageAction struct{}

func (t *tgSendGroupMessageAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendGroupMessageDetail().GetDetails()

	for _, detail := range details {
		for k, v := range detail.SendData {
			sendData := telegram.SendGroupMessageDetail{
				Sender:     k,
				GroupTitle: v.Key,
				SendText:   v.Values,
			}
			h.GetTgSendGroupMessageChan() <- sendData
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
