package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/telegram"
	"github.com/go-kit/log"
	"github.com/iyear/tdl/pkg/utils"
)

func init() {
	im, err := rpc.GetIM(consts.IMTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SEND_FILE, &tgSendFileAction{})
}

type tgSendFileAction struct{}

func (t *tgSendFileAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetSendFileDetail().GetDetails()
	var fileType string
	// 发文件
	sendFileData := telegram.SendFileDetail{}

	// 发图片
	sendPhotoData := telegram.SendImageFileDetail{}

	// 发视频
	sendVideoData := telegram.SendVideoDetail{}

	if len(details) > 0 {
		for _, detail := range details {
			for k, v := range detail.SendTgData {
				sendFileData.Sender = k
				sendFileData.Receiver = v.Key

				sendPhotoData.Sender = k
				sendPhotoData.Receiver = v.Key

				sendVideoData.Sender = k
				sendVideoData.Receiver = v.Key
				for _, f := range v.Value {
					if utils.Media.IsImage(f.FileType) {
						fileType = f.FileType
						sendPhotoData.FileType = append(sendPhotoData.FileType, telegram.SendImageFileType{
							FileType:   f.GetFileType(),
							SendType:   f.GetSendType(),
							Path:       f.GetPath(),
							Name:       f.GetName(),
							ImageBytes: f.GetFileByte(),
						})
					} else if utils.Media.IsVideo(f.FileType) {
						fileType = f.FileType
						sendVideoData.FileType = append(sendVideoData.FileType, telegram.SendFileType{
							FileType:  f.GetFileType(),
							SendType:  f.GetSendType(),
							FileBytes: f.GetFileByte(),
							Name:      f.GetName(),
							Path:      f.GetPath(),
						})
					} else {
						// 其余的全发文档
						fileType = f.FileType
						sendFileData.FileType = append(sendFileData.FileType, telegram.SendFileType{
							FileType:  f.GetFileType(),
							SendType:  f.GetSendType(),
							FileBytes: f.GetFileByte(),
							Name:      f.GetName(),
							Path:      f.GetPath(),
						})
					}
				}
				// 发送
				if utils.Media.IsImage(fileType) {
					h.GetTgSendPhotoChan() <- sendPhotoData

				} else if utils.Media.IsVideo(fileType) {
					h.GetTgSendVideoChan() <- sendVideoData

				} else {
					h.GetTgSendFileChan() <- sendFileData

				}
			}

		}
	} else {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: " send msg is nil "}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
