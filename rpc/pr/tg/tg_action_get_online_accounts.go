package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func init() {
	im, err := rpc.GetPR(consts.PRTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_GET_ONLINE_ACCOUNTS, &tgGetOnlineAccounts{})
}

type tgGetOnlineAccounts struct{}

func (t *tgGetOnlineAccounts) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetGetOnlineAccountsDetail()
	getOnline := p.PrGetSyncMap(rpc.TG_GET_ONLINE_ACCOUNTS)

	serverList, err := p.GetServiceListByName(in.GetType())
	if err != nil {
		level.Error(p.GetLogger()).Log("get server err", "err", err.Error())
		return
	}
	if len(serverList) > 0 {
		ip := serverList[0]
		phone := detail.Phone
		existingData, _ := getOnline.Load(ip)
		if existingData != nil {
			tmp := existingData.(*protobuf.GetOnlineAccountsDetail)
			tmp.Phone = phone
			getOnline.Store(ip, tmp)
		} else {
			tmp := &protobuf.GetOnlineAccountsDetail{
				Phone: phone,
			}

			getOnline.Store(ip, tmp)
		}
	}
	return
}
