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
	im.RegisterAction(protobuf.Action_LOGIN, &whatsLoginAction{})
}

type whatsLoginAction struct{}

func (t *whatsLoginAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	//name = in.GetType()
	details := in.GetOrdinaryAction().GetLoginDetail()
	wsLogin := p.PrGetSyncMap(rpc.WS_LOGIN)
	for userJid, ld := range details {

		ips, err := p.GetServiceListByName(in.GetType())
		if err != nil {
			level.Error(p.GetLogger()).Log("msg", "get ws service list fail", "err", err)
		}
		value, ok := p.Handler.GetServerMap().Load(userJid)
		if !ok {
			p.GetMu().Lock()
			key := ips[p.GetCurrentWsIPIndex()]
			p.Handler.GetServerMap().Store(userJid, key)

			existingData, _ := wsLogin.Load(key)
			if existingData != nil {
				tmp := existingData.(map[uint64]*protobuf.LoginDetail)
				tmp[userJid] = ld
				wsLogin.Store(key, tmp)
			} else {
				tmp := make(map[uint64]*protobuf.LoginDetail)
				tmp[userJid] = ld
				wsLogin.Store(key, tmp)
			}

			currentWsIPIndex := (p.GetCurrentWsIPIndex() + 1) % len(ips)
			p.SetCurrentWsIPIndex(currentWsIPIndex)
			p.GetMu().Unlock()
		} else {
			key := value.(string)
			existingData, _ := wsLogin.Load(key)
			if existingData != nil {
				tmp := existingData.(map[uint64]*protobuf.LoginDetail)
				tmp[userJid] = ld
				wsLogin.Store(key, tmp)
			} else {
				tmp := make(map[uint64]*protobuf.LoginDetail)
				tmp[userJid] = ld
				wsLogin.Store(key, tmp)
			}

		}
	}
	return
}
