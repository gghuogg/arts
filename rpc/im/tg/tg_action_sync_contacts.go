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
	im.RegisterAction(protobuf.Action_SYNC_CONTACTS, &tgSyncContactsAction{})
}

type tgSyncContactsAction struct{}

func (t *tgSyncContactsAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetSyncContactDetail().GetDetails()
	if len(detail) > 0 {
		for _, detail := range detail {
			value, ok := h.GetTgAccountsSync().Load(strconv.FormatUint(detail.Key, 10))

			if ok {
				account := value.(*telegram.Account)
				if account.IsLogin {
					SyncContact := telegram.SyncContact{
						Account:  detail.Key,
						Contacts: detail.Values,
						ResChan:  make(chan *protobuf.ResponseMessage),
					}
					h.GetTgSyncContactsChan() <- SyncContact
					result = <-SyncContact.ResChan
				} else {
					_ = level.Error(l).Log("msg", "Account is offline", "account", detail.Key)
				}
			} else {
				_ = level.Error(l).Log("msg", "Account is not exist,will be not sync contact", "account", detail.Key)

			}
		}
	} else {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "sync contact msg is null "}
		return
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
