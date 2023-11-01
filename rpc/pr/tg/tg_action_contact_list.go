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
	im.RegisterAction(protobuf.Action_CONTACT_LIST, &tgContactList{})
}

type tgContactList struct{}

func (t *tgContactList) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetGetContactList()
	if detail != nil {
		if account := detail.GetAccount(); account != 0 {
			addMember := p.PrGetSyncMap(rpc.TG_GET_CONTACT_LIST)
			value, ok := p.Handler.GetServerMap().Load(account)
			if ok {
				ip := value.(string)
				existingData, _ := addMember.Load(ip)

				if existingData != nil {
					tmp := existingData.(*protobuf.GetContactList)
					tmp = detail
					addMember.Store(ip, tmp)
				} else {
					tmp := &protobuf.GetContactList{}
					tmp.Account = account
					addMember.Store(ip, tmp)
				}
			}

		}
	}
	return
}
