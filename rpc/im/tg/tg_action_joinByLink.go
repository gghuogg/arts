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
	im.RegisterAction(protobuf.Action_JOIN_BY_LINK, &tgJoinByLinkAction{})
}

type tgJoinByLinkAction struct{}

func (t *tgJoinByLinkAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetJoinByLinkDetail().GetDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Key, 10))
	if ok {
		j := telegram.JoinByLinkDetail{
			Sender: details.Key,
			Links:  details.Values,
			Res:    make(chan *protobuf.ResponseMessage),
		}
		h.GetTgJoinByLinkChan() <- j

		result = <-j.Res

	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Key)
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "Account not exist,will be not continue login"}
	}

	return
}
