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
	im.RegisterAction(protobuf.Action_LEAVE, &tgLeaveAction{})
}

type tgLeaveAction struct{}

func (t *tgLeaveAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetLeaveDetail().GetDetail()
	_, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(details.Key, 10))
	if ok {
		tmp := telegram.LeaveDetail{
			Sender: details.Key,
			Leaves: details.Values,
			Res:    make(chan *protobuf.ResponseMessage),
		}
		h.GetTgLeaveChan() <- tmp
		result = <-tmp.Res

	} else {
		_ = level.Error(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", details.Key)
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "Account not exist,will be not continue login"}
	}

	return
}
