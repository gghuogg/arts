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
	im.RegisterAction(protobuf.Action_SYNC_APP_INFO, &tgSyncAppInfoAction{})
}

type tgSyncAppInfoAction struct{}

func (t *tgSyncAppInfoAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSyncAppAction().GetAppData()
	for phoneNum, detail := range details {
		_, ok := h.GetTgAccountsSync().Load(phoneNum)
		if !ok {
			account := telegram.AccountDetail{
				PhoneNumber: strconv.FormatUint(phoneNum, 10),
				AppId:       detail.AppId,
				AppHash:     detail.AppHash,
			}
			h.GetTgAccountDetailChan() <- account
		} else {
			_ = level.Error(l).Log("msg", "Account is exist,will be not continue sync account", "tg_phoneNum", phoneNum)
		}
	}
	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
