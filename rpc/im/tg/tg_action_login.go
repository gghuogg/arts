package tg

import (
	"arthas/callback"
	"arthas/consts"
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/telegram"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	tg "github.com/gotd/td/telegram"
	"google.golang.org/protobuf/proto"

	"strconv"
)

func init() {
	im, err := rpc.GetIM(consts.IMTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_LOGIN, &tgLoginAction{})
}

type tgLoginAction struct{}

func (t *tgLoginAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetOrdinaryAction().GetLoginDetail()
	for phone, detail := range details {
		phoneNum := strconv.FormatUint(phone, 10)

		get := etcd.GetReq{
			Key:     "/arthas/tg_accountdetail/" + phoneNum,
			Options: nil,
			ResChan: make(chan etcd.GetRes),
		}
		h.EtcdGetChan() <- get
		getResult := <-get.ResChan

		if getResult.Result.Kvs != nil {
			accountDetail := &protobuf.TgAccountDetail{}
			err := proto.Unmarshal(getResult.Result.Kvs[0].Value, accountDetail)
			if err != nil {
				return nil, err
			}

			device := tg.DeviceConfig{}
			if detail.TgDevice == nil {
				device = tg.DeviceConfig{
					DeviceModel:    "Desktop",
					SystemVersion:  "Windows 10",
					AppVersion:     "4.2.4 x64",
					LangCode:       "en",
					SystemLangCode: "en-US",
					LangPack:       "tdesktop",
				}
			} else {
				device = tg.DeviceConfig{
					DeviceModel:    detail.TgDevice.DeviceModel,
					SystemVersion:  detail.TgDevice.SystemVersion,
					AppVersion:     detail.TgDevice.AppVersion,
					SystemLangCode: detail.TgDevice.SystemLangCode,
					LangPack:       detail.TgDevice.LangPack,
					LangCode:       detail.TgDevice.LangCode,
				}
			}

			a := &telegram.Account{
				PhoneNumber: accountDetail.PhoneNumber,
				AppId:       accountDetail.AppId,
				AppHash:     accountDetail.AppHash,
				Device:      device,
			}
			h.GetTgAccountsSync().Store(accountDetail.PhoneNumber, a)
		}

		value, ok := h.GetTgAccountsSync().Load(phoneNum)
		if ok {
			account := value.(*telegram.Account)
			if account.IsLogin {
				_ = level.Info(l).Log("msg", "Account is online,will be not continue login", "userJid", phoneNum)
				return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Account: phoneNum}, nil
			} else {
				_ = level.Debug(l).Log("msg", "Account exist,continue login", "userJid", phoneNum)
				d := telegram.LoginDetail{
					PhoneNumber: phoneNum,
					ProxyUrl:    detail.ProxyUrl,
					LoginId:     detail.LoginId,
					ResChan:     make(chan telegram.LoginDetailRes),
				}
				h.GetTgLoginDetailChan() <- d
				res := <-d.ResChan
				if res.LoginStatus == protobuf.AccountStatus_SUCCESS {
					return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Account: phoneNum}, nil
				} else if res.LoginStatus == protobuf.AccountStatus_NEED_SEND_CODE {
					accountStatus := make(map[string]protobuf.AccountStatus)
					accountStatus[phoneNum] = protobuf.AccountStatus_NEED_SEND_CODE
					return &protobuf.ResponseMessage{
						ActionResult:  protobuf.ActionResult_LOGIN_NEED_CODE,
						AccountStatus: accountStatus,
						LoginId:       res.LoginId,
					}, nil
				} else {
					return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, LoginId: res.LoginId, RespondAccountStatus: res.LoginStatus}, nil
				}
			}
		} else {
			_ = level.Debug(l).Log("msg", "Account not exist,will be not continue login", "phoneNum", phoneNum)
			loginCallbacks := make([]callback.LoginCallback, 0)
			item := callback.LoginCallback{
				UserJid:     phone,
				ProxyUrl:    detail.ProxyUrl,
				LoginStatus: protobuf.AccountStatus_NOT_EXIST,
				Comment:     "Account not exist",
			}
			loginCallbacks = append(loginCallbacks, item)
			h.GetCallbackChan().LoginCallbackChan <- loginCallbacks
			return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Account: phoneNum, Comment: "Account not exist"}, nil
		}
	}
	return
}
