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
	im.RegisterAction(protobuf.Action_LOGOUT, &whatsLogoutAction{})
}

type whatsLogoutAction struct{}

func (t *whatsLogoutAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetOrdinaryAction().GetLoginDetail()
	var logoutSuccess int
	for userJid, detail := range details {
		value, ok := h.GetAccountsSync().Load(userJid)
		if ok {
			account := value.(*signal.Account)
			if account.IsLogin == false {
				_ = level.Debug(l).Log("msg", "Account is logout", "userJid", userJid)
				continue
			}
			logoutSuccess++
			d := signal.LogoutDetail{
				UserJid:  userJid,
				ProxyUrl: detail.ProxyUrl,
			}
			h.GetLogoutDetailChan() <- d
		}
	}
	if logoutSuccess == len(details) {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
		return
	} else if logoutSuccess > 0 {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
		return
	} else {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL}
		return
	}
}
