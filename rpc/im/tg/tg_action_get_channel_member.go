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
	im.RegisterAction(protobuf.Action_GET_CHANNEL_MEMBER, &tgGetChannelMembersAction{})
}

type tgGetChannelMembersAction struct{}

func (t *tgGetChannelMembersAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetChannelMemberDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Sender, 10))
	if ok {
		g := telegram.GetChannelMemberDetail{
			Sender:     details.Sender,
			Channel:    details.Channel,
			SearchType: details.SearchType,
			Offset:     int(details.Offset),
			Limit:      int(details.Limit),
			TopMsgId:   int(details.TopMsgId),
			Res:        make(chan *protobuf.ResponseMessage),
		}
		h.GetTgGetChannelMembersChan() <- g
		result = <-g.Res
	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Sender)
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL}
	}

	return
}
