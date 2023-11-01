package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/signal"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func init() {
	im, err := rpc.GetIM(consts.IMWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SYNC_ACCOUNT_KEY, &whatsSyncAccountKeyAction{})
}

type whatsSyncAccountKeyAction struct{}

func (t *whatsSyncAccountKeyAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSyncAccountKeyAction().KeyData
	for userJid, detail := range details {
		_, ok := h.GetAccountsSync().Load(userJid)
		if !ok {
			account := signal.AccountDetail{
				UserJid:          userJid,
				PrivateKey:       detail.Privatekey,
				PrivateMsgKey:    detail.PrivateMsgKey,
				PublicKey:        detail.PublicMsgKey,
				PublicMsgKey:     detail.PublicMsgKey,
				ResumptionSecret: detail.ResumptionSecret,
			}
			h.GetWhatsSyncAccountKeyChan() <- account
		} else {
			_ = level.Debug(l).Log("msg", "Account is exist,will be not continue sync account", "userJid", userJid)
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
