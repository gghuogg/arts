package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/signal"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetIM(consts.IMWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SEND_PHOTO, &whatsSendPhotoAction{})
}

type whatsSendPhotoAction struct{}

func (t *whatsSendPhotoAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendImageFileDetail().GetDetails()
	if len(details) > 0 {
		for _, detail := range details {
			for k, v := range detail.SendData {
				sendData := signal.SendImageFileDetail{
					Sender:   k,
					Receiver: v.Key,
				}
				for _, f := range v.Value {
					sendData.FileType = append(sendData.FileType, signal.SendImageFileType{
						FileType:   f.GetFileType(),
						SendType:   f.GetSendType(),
						Path:       f.GetPath(),
						Name:       f.GetName(),
						ImageBytes: f.GetFileByte(),
					})
				}
				h.GetWhatsSendPhotoChan() <- sendData
			}
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
