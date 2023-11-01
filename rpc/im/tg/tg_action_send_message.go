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
	im.RegisterAction(protobuf.Action_SEND_MESSAGE, &tgSendMessageAction{})
}

type tgSendMessageAction struct{}

func (t *tgSendMessageAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendmessageDetail().GetDetails()

	for _, detail := range details {
		for k, v := range detail.SendTgData {
			sendData := telegram.SendMessageDetail{
				Sender:   k,
				Receiver: v.Key,
				SendText: v.Values,
			}
			h.GetTgSendMessageChan() <- sendData
		}
	}
	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
