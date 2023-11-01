package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
)

// 图片，文件，视频，音频都改成用  Action_SEND_FILE
func init() {
	im, err := rpc.GetPR(consts.PRTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SEND_VIDEO, &tgSendVideoMsg{})
}

type tgSendVideoMsg struct{}

func (t *tgSendVideoMsg) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendFileDetail().GetDetails()
	sendFile := p.PrGetSyncMap(rpc.TG_SEND_VIDEO)
	if len(details) > 0 {
		for _, detail := range details {
			for k, v := range detail.SendTgData {
				value, ok := p.Handler.GetServerMap().Load(k)
				ip := value.(string)
				if ok {
					existingData, _ := sendFile.Load(ip)
					if existingData != nil {
						list := existingData.([]map[uint64]*protobuf.UintTgFileDetailValue)
						tmp := make(map[uint64]*protobuf.UintTgFileDetailValue)
						tmp[k] = v
						list = append(list, tmp)
						sendFile.Store(ip, list)

					} else {
						list := make([]map[uint64]*protobuf.UintTgFileDetailValue, 0)
						tmp := make(map[uint64]*protobuf.UintTgFileDetailValue)
						tmp[k] = v
						list = append(list, tmp)
						sendFile.Store(ip, list)
					}
				}
			}
		}
	}
	return
}
