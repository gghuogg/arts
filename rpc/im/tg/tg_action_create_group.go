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
	im.RegisterAction(protobuf.Action_CREATE_GROUP, &tgCreateGroupAction{})
}

type tgCreateGroupAction struct{}

func (t *tgCreateGroupAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetCreateGroupDetail().GetDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Key, 10))
	if ok {
		g := telegram.CreateGroupDetail{
			Initiator:    details.Key,
			GroupTitle:   in.GetCreateGroupDetail().GroupName,
			GroupMembers: details.Values,
			Res:          make(chan *protobuf.ResponseMessage),
		}
		h.GetTgCreateGroupChan() <- g
		result = <-g.Res
	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Key)

	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
