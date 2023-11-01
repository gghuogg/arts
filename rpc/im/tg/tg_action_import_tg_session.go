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
	im.RegisterAction(protobuf.Action_IMPORT_TG_SESSION, &tgImportSessionAction{})
}

type tgImportSessionAction struct{}

func (t *tgImportSessionAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetImportTgSession()

	sessionDetails := detail.GetSendData()
	if len(sessionDetails) > 0 {
		senDataList := make([]telegram.ImportTgSessionDetail, 0)
		for k, v := range sessionDetails {
			senDataList = append(senDataList, telegram.ImportTgSessionDetail{
				Account: k,
				DC:      v.DC,
				Addr:    v.Addr,
				AuthKey: v.AuthKey,
				AccountDeviceMsg: telegram.AccountDevice{
					AppId:   v.DeviceMsg.AppId,
					AppHash: v.DeviceMsg.AppHash,

					DeviceModel:    v.DeviceMsg.DeviceModel,
					SystemVersion:  v.DeviceMsg.SystemVersion,
					AppVersion:     v.DeviceMsg.AppVersion,
					LangCode:       v.DeviceMsg.LangCode,
					SystemLangCode: v.DeviceMsg.SystemLangCode,
					LangPack:       v.DeviceMsg.LangPack,
				},
			})
		}
		senData := telegram.ImportTgSessionMsg{Sender: detail.GetAccount(), Details: senDataList, ResChan: make(chan telegram.ImportSessionRes)}
		h.GeTgImportSessionChan() <- senData
		res := <-senData.ResChan
		result = res.ResMsg

	} else {
		_ = level.Error(l).Log("msg", "send session is nil", "account", detail.GetAccount())
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "Send session is nil" + strconv.FormatUint(detail.Account, 10)}
	}

	return
}
