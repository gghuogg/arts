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
	im.RegisterAction(protobuf.Action_LOGOUT, &whatsLogoutAction{})
}

type whatsLogoutAction struct{}

func (t *whatsLogoutAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	wsLogout := p.PrGetSyncMap(rpc.WS_LOGOUT)
	details := in.GetOrdinaryAction().GetLoginDetail()
	for userJid, ld := range details {
		value, ok := p.Handler.GetServerMap().Load(userJid)
		ip := value.(string)
		if ok {
			existingData, _ := wsLogout.Load(ip)
			if existingData != nil {
				tmp := existingData.(map[uint64]*protobuf.LoginDetail)
				tmp[userJid] = ld
				wsLogout.Store(ip, tmp)
			} else {
				tmp := make(map[uint64]*protobuf.LoginDetail)
				tmp[userJid] = ld
				wsLogout.Store(ip, tmp)
			}
		}
	}
	return
}
