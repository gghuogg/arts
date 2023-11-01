package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/telegram"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"strconv"
)

func init() {
	im, err := rpc.GetIM(consts.IMTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_MESSAGES_REACTION, &tgMessagesReactionAction{})
}

type tgMessagesReactionAction struct{}

func (t *tgMessagesReactionAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetMessagesReactionDetail().GetDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Key, 10))
	if ok {
		list := make([]int, 0)
		for _, uint64Id := range details.Values {
			tmp := int(uint64Id)
			list = append(list, tmp)
		}
		tmp := telegram.MessageReactionDetail{
			Sender:   details.Key,
			Emoticon: in.GetMessagesReactionDetail().Emotion,
			MsgIds:   list,
			Receiver: in.GetMessagesReactionDetail().Receiver,
			Res:      make(chan *protobuf.ResponseMessage),
		}
		h.GetMessagesReactionChan() <- tmp
		result = <-tmp.Res
	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Key)

	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
