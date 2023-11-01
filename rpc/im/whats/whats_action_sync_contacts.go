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
	im.RegisterAction(protobuf.Action_SYNC_CONTACTS, &whatsSyncContactsAction{})
}

type whatsSyncContactsAction struct{}

func (t *whatsSyncContactsAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetSyncContactDetail().GetDetails()
	if len(detail) > 0 {
		for _, detail := range detail {
			value, ok := h.GetAccountsSync().Load(detail.Key)
			if ok {
				account := value.(*signal.Account)
				if account.IsLogin {
					SyncContact := signal.SyncContact{
						Account:  detail.Key,
						Contacts: detail.Values,
					}
					h.GetWhatsSyncContactsChan() <- SyncContact
				} else {
					_ = level.Error(l).Log("msg", "Account is offline.", "account", detail.Key)
				}
			} else {
				_ = level.Error(l).Log("msg", "Account is not exist,will be not sync contact", "account", detail.Key)
			}
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
