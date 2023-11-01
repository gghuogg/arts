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
	im.RegisterAction(protobuf.Action_SEND_VCARD_MESSAGE, &whatsSendVcardMsgAction{})
}

type whatsSendVcardMsgAction struct{}

func (t *whatsSendVcardMsgAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	msg := in.GetSendVcardMessage().GetDetails()
	if msg != nil {
		for _, m := range msg {
			sendData := m.GetSendData()
			if sendData != nil {
				for k, v := range sendData {
					data := signal.SendVcardMessageDetail{
						Sender:   k,
						Receiver: v.GetReceiver(),
					}
					vcardList := make([]*signal.VCardDetail, 0)
					for _, vcard := range v.GetVcards() {
						vcardList = append(vcardList, &signal.VCardDetail{
							Fn:  vcard.GetFn(),
							Tel: vcard.GetTel(),
						})
					}
					data.VCardDetails = vcardList
					h.GetWhatsSendVcardMsgChan() <- data
				}
			}
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
