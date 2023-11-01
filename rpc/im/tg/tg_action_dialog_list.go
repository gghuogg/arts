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
	im.RegisterAction(protobuf.Action_DIALOG_LIST, &tgGetDialogListAction{})
}

type tgGetDialogListAction struct{}

func (t *tgGetDialogListAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetGetDialogList()
	if detail != nil {
		if account := detail.GetAccount(); account != 0 {
			sendCode := telegram.DialogListDetail{Account: account, ResChan: make(chan telegram.DialogListRes)}
			h.GetTgGetDialogListChan() <- sendCode
			res := <-sendCode.ResChan
			return res.AccountResult, nil
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "send protobuf msg is nil or having err "}

	return
}
