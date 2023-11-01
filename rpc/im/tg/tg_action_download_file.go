package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/telegram"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetIM(consts.IMTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_DOWNLOAD_FILE, &tgDownloadFileAction{})
}

type tgDownloadFileAction struct{}

func (t *tgDownloadFileAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetGetDownLoadFileDetail().GetDownloadFile()

	if len(detail) > 0 {
		for k, v := range detail {
			sendCode := telegram.DownChatFileDetail{Account: k, ChatId: v.GetChatId(), MessageId: v.GetMessageId(), ResChan: make(chan telegram.DownChatFileRes)}
			h.GetTgDownloadFileChan() <- sendCode
			res := <-sendCode.ResChan
			result = res.ResMsg
		}
	} else {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "send protobuf msg is nil or having err "}

	}
	return
}
