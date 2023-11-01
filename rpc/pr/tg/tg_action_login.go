package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/errors/gerror"
)

func init() {
	im, err := rpc.GetPR(consts.PRTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_LOGIN, &tgLoginAction{})
}

type tgLoginAction struct{}

func (t *tgLoginAction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetOrdinaryAction().GetLoginDetail()
	tgLogin := p.PrGetSyncMap(rpc.TG_LOGIN)
	for userJid, ld := range details {

		ips, err := p.GetServiceListByName(in.GetType())
		if err != nil {
			level.Error(p.GetLogger()).Log("msg", "get tg service list fail", "err", err)
		}
		value, ok := p.Handler.GetServerMap().Load(userJid)
		if len(ips) > 0 {
			if !ok {
				p.GetMu().Lock()
				key := ips[p.GetCurrentWsIPIndex()]

				p.Handler.GetServerMap().Store(userJid, key)

				existingData, _ := tgLogin.Load(key)
				if existingData != nil {
					tmp := existingData.(map[uint64]*protobuf.LoginDetail)
					tmp[userJid] = ld
					tgLogin.Store(key, tmp)
				} else {
					tmp := make(map[uint64]*protobuf.LoginDetail)
					tmp[userJid] = ld
					tgLogin.Store(key, tmp)
				}
				currentWsIPIndex := (p.GetCurrentWsIPIndex() + 1) % len(ips)
				p.SetCurrentWsIPIndex(currentWsIPIndex)
				p.GetMu().Unlock()
			} else {
				key := value.(string)
				existingData, _ := tgLogin.Load(key)
				if existingData != nil {
					tmp := existingData.(map[uint64]*protobuf.LoginDetail)
					tmp[userJid] = ld
					tgLogin.Store(key, tmp)
				} else {
					tmp := make(map[uint64]*protobuf.LoginDetail)
					tmp[userJid] = ld
					tgLogin.Store(key, tmp)
				}

			}
		} else {
			err = gerror.Wrap(err, "No services available "+err.Error())
			return nil, err
		}

	}
	return
}
