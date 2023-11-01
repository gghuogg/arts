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
	im.RegisterAction(protobuf.Action_MESSAGES_REACTION, &tgMessagesReaction{})
}

type tgMessagesReaction struct{}

func (t *tgMessagesReaction) Handler(l log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetMessagesReactionDetail()
	messagesReaction := p.PrGetSyncMap(rpc.TG_MESSAGES_REACTION)
	value, ok := p.Handler.GetServerMap().Load(details.Detail.Key)
	if ok {
		ip := value.(string)
		existingData, _ := messagesReaction.Load(ip)
		if existingData != nil {
			tmp := existingData.(map[uint64]*protobuf.MessagesReactionDetail)
			tmp[details.Detail.Key] = details
			messagesReaction.Store(ip, tmp)
		} else {
			tmp := make(map[uint64]*protobuf.MessagesReactionDetail)
			tmp[details.Detail.Key] = details
			messagesReaction.Store(ip, tmp)
		}
	}
	return
}
