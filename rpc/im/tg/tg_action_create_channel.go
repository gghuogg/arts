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
	im.RegisterAction(protobuf.Action_CREATE_CHANNEL, &tgCreateChannelAction{})
}

type tgCreateChannelAction struct{}

func (t *tgCreateChannelAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetCreateChannelDetail().GetDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Key, 10))
	if ok {
		g := telegram.CreateChannelDetail{
			Creator:            details.Key,
			ChannelTitle:       in.GetCreateChannelDetail().ChannelTitle,
			ChannelDescription: in.GetCreateChannelDetail().ChannelDescription,
			ChannelUserName:    in.GetCreateChannelDetail().ChannelUserName,
			ChannelMember:      details.Values,
			IsChannel:          in.GetCreateChannelDetail().IsChannel,
			IsSuperGroup:       in.GetCreateChannelDetail().IsSuperGroup,
			Res:                make(chan *protobuf.ResponseMessage),
		}
		h.GetTgCreateChannel() <- g
		result = <-g.Res

	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Key)
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "Account not exist,will be not continue login"}
	}

	return
}
