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
	im.RegisterAction(protobuf.Action_SEND_VIDEO, &tgSendVideoAction{})
}

type tgSendVideoAction struct{}

func (t *tgSendVideoAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendFileDetail().GetDetails()
	if len(details) > 0 {
		for _, detail := range details {
			for k, v := range detail.SendTgData {
				sendData := telegram.SendVideoDetail{
					Sender:   k,
					Receiver: v.Key,
				}
				for _, f := range v.Value {
					sendData.FileType = append(sendData.FileType, telegram.SendFileType{
						FileType:  f.GetFileType(),
						SendType:  f.GetSendType(),
						FileBytes: f.GetFileByte(),
						Name:      f.GetName(),
						Path:      f.GetPath(),
					})
				}
				h.GetTgSendVideoChan() <- sendData
			}
		}
	}
	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
