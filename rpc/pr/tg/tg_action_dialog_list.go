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
	im.RegisterAction(protobuf.Action_DIALOG_LIST, &tgDialogList{})
}

type tgDialogList struct{}

func (t *tgDialogList) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetGetDialogList()
	if detail != nil {
		if account := detail.GetAccount(); account != 0 {
			addMember := p.PrGetSyncMap(rpc.TG_GET_DIALOG_LIST)
			value, ok := p.Handler.GetServerMap().Load(account)
			if ok {
				ip := value.(string)
				existingData, _ := addMember.Load(ip)

				if existingData != nil {
					tmp := existingData.(*protobuf.GetDialogList)
					tmp = detail
					addMember.Store(ip, tmp)
				} else {
					tmp := &protobuf.GetDialogList{}
					tmp.Account = account
					addMember.Store(ip, tmp)
				}
			}

		}
	}
	return
}
