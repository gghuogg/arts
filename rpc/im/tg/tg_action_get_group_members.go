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
	im.RegisterAction(protobuf.Action_GET_GROUP_MEMBERS, &tgGetGroupMembersAction{})
}

type tgGetGroupMembersAction struct{}

func (t *tgGetGroupMembersAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetGroupMembersDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Account, 10))
	if ok {
		g := telegram.GetGroupMembers{
			Initiator: details.Account,
			ChatId:    details.ChatId,
			Res:       make(chan *protobuf.ResponseMessage),
		}
		h.GetTgGetGroupMembersChan() <- g
		result = <-g.Res
	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Account)
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL}
	}

	return
}
