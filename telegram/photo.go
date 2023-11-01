package telegram

import (
	"arthas/callback"
	"bytes"
	"context"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type SendImageFileDetail struct {
	Sender   uint64
	Receiver string
	FileType []SendImageFileType
	//ResChan  chan *protobuf.ResponseMessage
}

type SendImageFileType struct {
	FileType   string
	SendType   string
	Path       string
	Name       string
	ImageBytes []byte
}

type TgSendMsgError struct {
	ReqId   int
	Err     error
	SendMsg []byte
}

func (m *Manager) SendPhoto(ctx context.Context, c *telegram.Client, opts SendImageFileDetail) *TgSendMsgError {
	tgFileMsgCallbacks := make([]callback.TgMsgCallback, 0)

	fileNum := len(opts.FileType)
	var fileSuccessNum int = 0

	s := message.NewSender(c.API())
	self, err := c.Self(ctx)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get self", err)
		return &TgSendMsgError{Err: err}
	}
	up := uploader.NewUploader(c.API())
	peer, err := m.getPeer(ctx, c, opts.Receiver)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get peer", err)
		return &TgSendMsgError{Err: err}
	}
	for _, f := range opts.FileType {
		// 判断是否为图片文件
		if f.SendType == TG_SEND_TRYPE_URL {
			if f.Path == "" {
				level.Error(m.logger).Log("msg", "file path is nil")
				tgFileMsgCallbacks = photoErrCallback(tgFileMsgCallbacks, opts, f, 0, err.Error())
				continue
			}
			imageByte, err := getFileFromOSSAndConvertToBytes(f.Path)
			if err != nil {
				level.Error(m.logger).Log("msg", "Invalid OSS path.")
				tgFileMsgCallbacks = photoErrCallback(tgFileMsgCallbacks, opts, f, 0, err.Error())
				continue
			}
			f.ImageBytes = imageByte

			if &f.Name == nil || f.Name == "" {
				f.Name = filepath.Base(f.Path)
			}
		}

		dialog := s.WithUploader(up).To(peer.InputPeer())

		randomValue := generateRandomValue()
		time.Sleep(time.Duration(randomValue) * time.Millisecond)
		err := s.To(peer.InputPeer()).TypingAction().UploadPhoto(ctx, 1)
		if err != nil {
			return nil
		}
		var uploadErr error
		var resp tg.UpdatesClass

		// 发gif
		if f.FileType == TG_FILE_TYPE_GIF {
			fi, err := up.Upload(ctx, uploader.NewUpload(f.Name, bytes.NewReader(f.ImageBytes), int64(len(f.ImageBytes))))
			if err != nil {
				level.Error(m.logger).Log("msg", "upload gif err", "err", err.Error())
				tgFileMsgCallbacks = photoErrCallback(tgFileMsgCallbacks, opts, f, 0, err.Error())
				continue
			} else {
				// 发送gif
				resp, uploadErr = dialog.GIF(ctx, fi)
			}
		} else {
			// 发图片
			resp, uploadErr = dialog.Upload(message.FromBytes(f.Name, f.ImageBytes)).Photo(ctx)
		}
		if uploadErr != nil {
			level.Error(m.logger).Log("msg", "send photo fail", "err", err)
			tgFileMsgCallbacks = photoErrCallback(tgFileMsgCallbacks, opts, f, 0, err.Error())
			continue
		}

		msgCallback := callback.TgMsgCallback{
			Sender:     uint64(self.ID),
			Receiver:   gconv.String(peer.ID()),
			SendTime:   time.Now(),
			ReqId:      gconv.String(resp.(*tg.Updates).Updates[0].(*tg.UpdateMessageID).ID),
			Read:       1,
			Initiator:  uint64(self.ID),
			MsgType:    2,
			FileName:   f.Name,
			Md5:        gmd5.MustEncryptBytes(f.ImageBytes),
			SendMsg:    nil,
			FileType:   f.FileType,
			SendStatus: 1,
			Out:        1,
		}
		tgFileMsgCallbacks = append(tgFileMsgCallbacks, msgCallback)
		fileSuccessNum++
	}

	if len(tgFileMsgCallbacks) > 0 {
		m.sendTgMsgCallback(tgFileMsgCallbacks)
	}
	if fileSuccessNum == fileNum {
		//opts.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	} else if fileSuccessNum == 0 {
		//opts.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL}
	} else {
		//opts.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_PARTIAL_SUCCESS}
	}
	return nil
	//})
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}

	for _, imageExt := range imageExts {
		if ext == imageExt {
			return true
		}
	}

	return false
}

func photoErrCallback(tgFileMsgCallbacks []callback.TgMsgCallback, opts SendImageFileDetail, f SendImageFileType, msgId int, errMsg string) []callback.TgMsgCallback {

	msgCallback := callback.TgMsgCallback{
		Sender:     opts.Sender,
		Receiver:   opts.Receiver,
		SendTime:   time.Now(),
		ReqId:      strconv.Itoa(msgId),
		Read:       2,
		Initiator:  opts.Sender,
		MsgType:    3,
		FileName:   f.Name,
		SendMsg:    nil,
		FileType:   f.FileType,
		SendStatus: 2,
		Comment:    errMsg,
	}
	tgFileMsgCallbacks = append(tgFileMsgCallbacks, msgCallback)
	return tgFileMsgCallbacks
}
