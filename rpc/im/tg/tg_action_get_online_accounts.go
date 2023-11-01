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
	im.RegisterAction(protobuf.Action_GET_ONLINE_ACCOUNTS, &tgGetOnlineAccountsAction{})
}

type tgGetOnlineAccountsAction struct{}

func (t *tgGetOnlineAccountsAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetOnlineAccountsDetail()

	g := telegram.GetOnlineAccountDetail{
		Phone: details.Phone,
		Res:   make(chan *protobuf.ResponseMessage),
	}
	h.GetTgGetOnlineAccountsChan() <- g
	result = <-g.Res

	return
}
