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
	im.RegisterAction(protobuf.Action_SYNC_APP_INFO, &tgSyncAppInfo{})
}

type tgSyncAppInfo struct{}

func (t *tgSyncAppInfo) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSyncAppAction().GetAppData()
	tgSyncInfo := p.PrGetSyncMap(rpc.TG_SYNC_INFO)
	ips, err := p.GetServiceListByName(in.GetType())
	if len(ips) > 0 {
		for phoneNum, ad := range details {
			ip := ips[len(ips)-1]
			if err != nil {
				level.Error(p.GetLogger()).Log("msg", "get service list fail", "err", err)
			}
			existingData, _ := tgSyncInfo.Load(ip)
			if existingData != nil {
				tmp := existingData.(map[uint64]*protobuf.AppData)
				tmp[phoneNum] = ad
				tgSyncInfo.Store(ip, tmp)
			} else {
				tmp := make(map[uint64]*protobuf.AppData)
				tmp[phoneNum] = ad
				tgSyncInfo.Store(ip, tmp)
			}

		}
	} else {
		err = gerror.Wrap(err, "No services available ")
		return nil, err
	}

	return
}
