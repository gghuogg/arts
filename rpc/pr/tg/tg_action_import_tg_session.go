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
	im.RegisterAction(protobuf.Action_IMPORT_TG_SESSION, &tgImportTgSession{})
}

type tgImportTgSession struct{}

func (t *tgImportTgSession) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetImportTgSession()
	invite := p.PrGetSyncMap(rpc.TG_IMPORT_SESSION)
	serverList, err := p.GetServiceListByName(in.GetType())
	if err != nil {
		level.Error(p.GetLogger()).Log("get server err", "err", err.Error())
		return
	}
	if len(serverList) > 0 {
		ip := serverList[0]
		existingData, _ := invite.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.ImportTgSessionMsg)
			tmp = detail.GetSendData()
			invite.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.ImportTgSessionMsg)
			tmp = detail.GetSendData()
			invite.Store(ip, tmp)
		}

	} else {
		err = gerror.Wrap(err, "No services available "+err.Error())
		return nil, err
	}
	return
}
