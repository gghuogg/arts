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
	im.RegisterAction(protobuf.Action_INVITE_TO_CHANNEL, &tgInviteToChannelAction{})
}

type tgInviteToChannelAction struct{}

func (t *tgInviteToChannelAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetInviteToChannelDetail().GetDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Key, 10))
	if ok {
		a := telegram.InviteToChannelDetail{
			Invite:  details.Key,
			Channel: in.GetInviteToChannelDetail().Channel,
			Invited: details.Values,
			Res:     make(chan *protobuf.ResponseMessage),
		}
		h.GetTgInviteToChannelChan() <- a

		result = <-a.Res

	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Key)
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "Account not exist,will be not continue login"}
	}

	return
}
