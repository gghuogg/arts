package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetPR(consts.PRTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_LOGOUT, &tgLogoutAction{})
}

type tgLogoutAction struct{}

func (t *tgLogoutAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	tgLogout := p.PrGetSyncMap(rpc.TG_LOGOUT)
	details := in.GetLogoutAction().GetLogoutDetail()
	for userJid, ld := range details {
		value, ok := p.Handler.GetServerMap().Load(userJid)
		ip := value.(string)
		if ok {
			existingData, _ := tgLogout.Load(ip)
			if existingData != nil {
				tmp := existingData.(map[uint64]*protobuf.LogoutDetail)
				tmp[userJid] = ld
				tgLogout.Store(ip, tmp)
			} else {
				tmp := make(map[uint64]*protobuf.LogoutDetail)
				tmp[userJid] = ld
				tgLogout.Store(ip, tmp)
			}
		}
	}
	return
}
