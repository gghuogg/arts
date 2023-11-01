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
	im.RegisterAction(protobuf.Action_SEND_PHOTO, &tgSendPhotoAction{})
}

type tgSendPhotoAction struct{}

func (t *tgSendPhotoAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendFileDetail().GetDetails()
	for _, detail := range details {
		for k, v := range detail.SendTgData {
			sendData := telegram.SendImageFileDetail{
				Sender:   k,
				Receiver: v.Key,
			}
			for _, f := range v.Value {
				sendData.FileType = append(sendData.FileType, telegram.SendImageFileType{
					FileType:   f.GetFileType(),
					SendType:   f.GetSendType(),
					Path:       f.GetPath(),
					Name:       f.GetName(),
					ImageBytes: f.GetFileByte(),
				})
			}
			h.GetTgSendPhotoChan() <- sendData
		}
	}
	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
