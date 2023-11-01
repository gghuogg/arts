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
	im.RegisterAction(protobuf.Action_SEND_FILE, &whatsSendFileAction{})
}

type whatsSendFileAction struct{}

func (t *whatsSendFileAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendFileDetail().GetDetails()
	if len(details) > 0 {
		for _, detail := range details {
			for k, v := range detail.SendData {
				sendData := signal.SendFileDetail{
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
				h.GetWhatsSendFileChan() <- sendData
			}
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
