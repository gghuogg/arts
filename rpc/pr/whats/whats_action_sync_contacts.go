package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetPR(consts.PRWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SYNC_ACCOUNT_KEY, &whatsSyncContactsAction{})
}

type whatsSyncContactsAction struct{}

func (t *whatsSyncContactsAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSyncContactDetail().GetDetails()
	syncContact := p.PrGetSyncMap(rpc.WS_SYNC_CONTACT)
	for _, detail := range details {
		value, ok := p.Handler.GetServerMap().Load(detail.Key)
		if ok {
			ip := value.(string)
			existingData, _ := syncContact.Load(ip)
			if existingData != nil {
				tmp := existingData.(map[uint64][]uint64)
				tmp[detail.Key] = detail.Values
				syncContact.Store(ip, tmp)
			} else {
				tmp := make(map[uint64][]uint64)
				tmp[detail.Key] = detail.Values
				syncContact.Store(ip, tmp)
			}
		}
	}
	return
}
