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
	im.RegisterAction(protobuf.Action_GET_EMOJI_GROUP, &tgGetEmojiGroup{})
}

type tgGetEmojiGroup struct{}

func (t *tgGetEmojiGroup) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetGetEmojiGroupDetail()
	getEmoji := p.PrGetSyncMap(rpc.TG_GET_EMOJI_GROUPS)
	value, ok := p.Handler.GetServerMap().Load(details.Sender)
	if ok {
		ip := value.(string)
		existingData, _ := getEmoji.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.GetEmojiGroupsDetail)
			tmp[details.Sender] = details
			getEmoji.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.GetEmojiGroupsDetail)
			tmp[details.Sender] = details
			getEmoji.Store(ip, tmp)
		}
	}
	return
}
