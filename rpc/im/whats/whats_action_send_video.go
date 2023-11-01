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
	im.RegisterAction(protobuf.Action_SEND_PHOTO, &whatsSendVideoAction{})
}

type whatsSendVideoAction struct{}

func (t *whatsSendVideoAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendVideoDetail().GetDetails()
	if len(details) > 0 {
		for _, detail := range details {
			for k, v := range detail.SendData {
				sendData := signal.SendVideoFileDetail{
					Sender:   k,
					Receiver: v.Key,
				}
				for _, f := range v.Value {
					sendData.FileType = append(sendData.FileType, signal.SendFileType{
						FileType:  f.GetFileType(),
						SendType:  f.GetSendType(),
						Path:      f.GetPath(),
						Name:      f.GetName(),
						FileBytes: f.GetFileByte(),
					})
				}
				h.GetWhatsSendVideoChan() <- sendData
			}
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
