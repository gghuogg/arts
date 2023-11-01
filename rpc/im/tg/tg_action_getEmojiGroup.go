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
	im.RegisterAction(protobuf.Action_GET_EMOJI_GROUP, &tgGetEmojiGroupsAction{})
}

type tgGetEmojiGroupsAction struct{}

func (t *tgGetEmojiGroupsAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetEmojiGroupDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Sender, 10))
	if ok {
		g := telegram.GetEmojiGroupsDetail{
			Sender: details.Sender,
			Res:    make(chan *protobuf.ResponseMessage),
		}
		h.GetTgGetEmojiGroupChan() <- g
		result = <-g.Res

	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Sender)
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL}
	}

	return
}
