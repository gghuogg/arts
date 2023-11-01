package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/signal"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetIM(consts.IMWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SEND_MESSAGE, &whatsSendMessageAction{})
}

type whatsSendMessageAction struct{}

func (t *whatsSendMessageAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendmessageDetail().GetDetails()

	for _, detail := range details {
		for k, v := range detail.SendData {
			sendData := signal.SendMessageDetail{
				Sender:   k,
				Receiver: v.Key,
				SendText: v.Values,
			}
			h.GetWhatsSendMessageChan() <- sendData
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
