package telegram

import (
	"arthas/callback"
	"bytes"
	"context"
	"crypto/rand"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/utils"
	"io"
	"path/filepath"
	"strconv"
	"time"
)

type SendVideoDetail struct {
	Sender   uint64
	Receiver string
	FileType []SendFileType
	//ResChan  chan *protobuf.ResponseMessage
}

func (m *Manager) SendVideo(ctx context.Context, c *telegram.Client, opts SendVideoDetail) error {
	tgFileMsgCallbacks := make([]callback.TgMsgCallback, 0)

	fileNum := len(opts.FileType)
	var fileSuccessNum int = 0

	s := message.NewSender(c.API())

	self, err := c.Self(ctx)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get self", err)
		return err
	}
	up := uploader.NewUploader(c.API())
	peer, err := m.getPeer(ctx, c, opts.Receiver)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get peer", err)
		return err
	}
	//发送消息前将历史消息标记已读
	_, _ = c.API().MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{Peer: peer.InputPeer()})

	id64, _ := RandInt64(rand.Reader)
	id := int(id64)
	for _, f := range opts.FileType {

		var fi tg.InputFileClass
		fileRead := bytes.NewReader(f.FileBytes)
		if f.SendType == TG_SEND_TYPE_BYTE {
			fi, err = up.Upload(ctx, uploader.NewUpload(f.Name, fileRead, int64(len(f.FileBytes))))
			if err != nil {
				level.Error(m.logger).Log("msg", "upload file is fail")
				tgFileMsgCallbacks = videoErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
		}
		if f.SendType == TG_SEND_TRYPE_URL {
			if f.Path == "" {
				level.Error(m.logger).Log("msg", "file path is nil err")
				tgFileMsgCallbacks = videoErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
			// 通过OSS获取文件切片
			fileByte, err := getFileFromOSSAndConvertToBytes(f.Path)
			if err != nil {
				level.Error(m.logger).Log("msg", "get file from oss path err", "err", err)
				tgFileMsgCallbacks = videoErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
			f.FileBytes = fileByte
			if &f.Name == nil || f.Name == "" {
				f.Name = filepath.Base(f.Path)
			}

			fi, err = up.Upload(ctx, uploader.NewUpload(f.Name, fileRead, int64(len(f.FileBytes))))
			if err != nil {
				level.Error(m.logger).Log("msg", "upload video is err", "err", err)
				tgFileMsgCallbacks = videoErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
		}

		// reset reader
		if _, err = fileRead.Seek(0, io.SeekStart); err != nil {
			level.Error(m.logger).Log("msg", "video seek is err", "err", err)
			tgFileMsgCallbacks = videoErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
			continue
		}

		dialog := s.WithUploader(up).To(peer.InputPeer())
		err := s.To(peer.InputPeer()).TypingAction().UploadVideo(ctx, 1)
		doc := message.UploadedDocument(fi).Filename(f.Name)

		dur, w, h, err := utils.Media.GetMP4Info(fileRead)

		var media message.MediaOption = doc

		randomValue := generateRandomValue()
		time.Sleep(time.Duration(randomValue) * time.Millisecond)

		media = doc.Video().Duration(time.Duration(dur)*time.Second).Resolution(w, h).SupportsStreaming()

		resp, uploadErr := dialog.Media(ctx, media)

		if uploadErr != nil {
			level.Error(m.logger).Log("msg", "send video fail", "err", err)
			tgFileMsgCallbacks = videoErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
			continue
		}

		msgCallback := callback.TgMsgCallback{
			Sender:     uint64(self.ID),
			Receiver:   gconv.String(peer.ID()),
			SendTime:   time.Now(),
			ReqId:      gconv.String(resp.(*tg.Updates).Updates[0].(*tg.UpdateMessageID).ID),
			Read:       1,
			Initiator:  uint64(self.ID),
			MsgType:    3,
			FileName:   f.Name,
			Md5:        gmd5.MustEncryptBytes(f.FileBytes),
			SendMsg:    nil,
			FileType:   f.FileType,
			SendStatus: 1,
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
}

// 文件转成byte

func videoErrCallback(tgFileMsgCallbacks []callback.TgMsgCallback, opts SendVideoDetail, f SendFileType, msgId int, errMsg string) []callback.TgMsgCallback {
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
