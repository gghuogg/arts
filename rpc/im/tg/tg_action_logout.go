package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/telegram"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/errors/gerror"
	"strconv"
)

func init() {
	im, err := rpc.GetIM(consts.IMTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_LOGOUT, &tgLogoutAction{})
}

type tgLogoutAction struct{}

func (t *tgLogoutAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetLogoutAction().GetLogoutDetail()
	var logoutSuccess int
	for userJid, detail := range details {
		value, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(userJid, 10))
		if ok {
			account := value.(*telegram.Account)
			if account.IsLogin == false {
				continue
			}
			logoutSuccess++
			d := telegram.TgLogoutDetail{
				UserJid:  userJid,
				ProxyUrl: detail.ProxyUrl,
				ResChan:  make(chan telegram.LogoutRes),
			}
			h.GetTgLogoutChan() <- d
			logoutRes := <-d.ResChan
			if logoutRes.LogoutStatus == protobuf.AccountStatus_SUCCESS {
				account.IsLogin = false
				h.GetTgAccountsSync().Delete(strconv.FormatUint(userJid, 10))
				_ = level.Debug(l).Log("msg", "Account logout success", "userJid", userJid)
				continue
			} else {
				account.IsLogin = false
				_ = level.Debug(l).Log("msg", "Account logout fail ", "userJid", userJid)
				err = gerror.New(logoutRes.Comment)
				return
			}

		} else {
			err = gerror.New("can not find user in server")
			return
		}
	}
	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
