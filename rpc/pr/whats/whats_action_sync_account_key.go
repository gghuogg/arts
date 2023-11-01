package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func init() {
	im, err := rpc.GetPR(consts.PRWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SYNC_ACCOUNT_KEY, &whatsSyncAccountKeyAction{})
}

type whatsSyncAccountKeyAction struct{}

func (t *whatsSyncAccountKeyAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSyncAccountKeyAction().KeyData
	wsSyncKey := p.PrGetSyncMap(rpc.WS_SYNC_KEY)
	for userJid, kd := range details {
		ips, err := p.GetServiceListByName(in.GetType())
		ip := ips[len(ips)-1]
		if err != nil {
			level.Error(p.GetLogger()).Log("msg", "get service list fail", "err", err)
		}
		existingData, _ := wsSyncKey.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.KeyData)
			tmp[userJid] = kd
			wsSyncKey.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.KeyData)
			tmp[userJid] = kd
			wsSyncKey.Store(ip, tmp)
		}
	}
	return
}
