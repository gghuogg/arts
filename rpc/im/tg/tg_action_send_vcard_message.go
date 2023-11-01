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
	im.RegisterAction(protobuf.Action_SEND_VCARD_MESSAGE, &tgSendVcardMessageAction{})
}

type tgSendVcardMessageAction struct{}

func (t *tgSendVcardMessageAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendContactCardDetail().GetDetail()
	for _, detail := range details {
		for k, v := range detail.SendData {
			contactCardDetail := telegram.ContactCardDetail{
				Sender:   k,
				Receiver: v.Receiver,
			}
			cardList := make([]telegram.ContactCard, 0)
			for _, card := range v.Value {
				cardList = append(cardList, telegram.ContactCard{
					FirstName:   card.FirstName,
					LastName:    card.PhoneNumber,
					PhoneNumber: card.PhoneNumber,
				})
			}
			contactCardDetail.ContactCards = cardList
			h.GetTgSendVcardMsgChan() <- contactCardDetail
		}
	}
	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
