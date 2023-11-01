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
	"path/filepath"
	"strconv"
	"time"
)

type SendFileDetail struct {
	Sender   uint64
	Receiver string
	FileType []SendFileType
	//ResChan  chan *protobuf.ResponseMessage
}

type SendFileType struct {
	FileType  string
	SendType  string
	Path      string
	Name      string
	FileBytes []byte
}

func (m *Manager) SendFile(ctx context.Context, c *telegram.Client, opts SendFileDetail) *TgSendMsgError {
	tgFileMsgCallbacks := make([]callback.TgMsgCallback, 0)

	fileNum := len(opts.FileType)
	var fileSuccessNum int = 0

	s := message.NewSender(c.API())

	up := uploader.NewUploader(c.API())
	self, err := c.Self(ctx)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get self", err)
		return &TgSendMsgError{Err: err}
	}
	peer, err := m.getPeer(ctx, c, opts.Receiver)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get peer", err)
		return &TgSendMsgError{Err: err}
	}
	//发送消息前将历史消息标记已读
	_, _ = c.API().MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{Peer: peer.InputPeer()})
	id64, _ := RandInt64(rand.Reader)
	id := int(id64)

	for _, f := range opts.FileType {

		var fi tg.InputFileClass
		if f.SendType == TG_SEND_TYPE_BYTE {
			fi, err = up.Upload(ctx, uploader.NewUpload(f.Name, bytes.NewReader(f.FileBytes), int64(len(f.FileBytes))))
			if err != nil {
				level.Error(m.logger).Log("msg", "upload file is fail")
				tgFileMsgCallbacks = fileErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
		}
		if f.SendType == TG_SEND_TRYPE_URL {
			if f.Path == "" {
				level.Error(m.logger).Log("msg", "file path is nil err")
				tgFileMsgCallbacks = fileErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
			// 通过OSS获取文件切片
			fileByte, err := getFileFromOSSAndConvertToBytes(f.Path)
			if err != nil {
				level.Error(m.logger).Log("msg", "get file from oss path err", "err", err)
				tgFileMsgCallbacks = fileErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
			f.FileBytes = fileByte
			if &f.Name == nil || f.Name == "" {
				f.Name = filepath.Base(f.Path)
			}

			fi, err = up.Upload(ctx, uploader.NewUpload(f.Name, bytes.NewReader(f.FileBytes), int64(len(f.FileBytes))))
			if err != nil {
				level.Error(m.logger).Log("msg", "upload file is err", "err", err)
				tgFileMsgCallbacks = fileErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
				continue
			}
		}

		dialog := s.WithUploader(up).To(peer.InputPeer())

		doc := message.UploadedDocument(fi).Filename(f.Name)

		var media message.MediaOption = doc
		if f.FileType == TG_FILE_TYPE_OGG {
			_ = s.To(peer.InputPeer()).TypingAction().UploadAudio(ctx, 1)
		} else {
			_ = s.To(peer.InputPeer()).TypingAction().UploadDocument(ctx, 1)
		}

		randomValue := generateRandomValue()
		time.Sleep(time.Duration(randomValue) * time.Millisecond)

		var uploadErr error
		var resp tg.UpdatesClass
		if f.FileType == TG_FILE_TYPE_OGG {
			// 发音频
			resp, uploadErr = dialog.Voice(ctx, fi)
		} else if f.FileType == TG_FILE_TYPE_MP3 {
			// 发mp3
			resp, uploadErr = dialog.Media(ctx, message.Audio(fi).Title(f.Name))
		} else {
			// 发文件
			resp, uploadErr = dialog.Media(ctx, media)
		}

		if uploadErr != nil {
			level.Error(m.logger).Log("msg", "send file fail", "err", err)
			tgFileMsgCallbacks = fileErrCallback(tgFileMsgCallbacks, opts, f, id, err.Error())
			continue
			//return &TgSendMsgError{Err: uploadErr, SendMsg: []byte(f.Path), ReqId: id}
		}

		msgCallback := callback.TgMsgCallback{
			Sender:     uint64(self.ID),
			Receiver:   gconv.String(peer.ID()),
			SendTime:   time.Now(),
			ReqId:      gconv.String(resp.(*tg.Updates).Updates[0].(*tg.UpdateMessageID).ID),
			Read:       1,
			Initiator:  uint64(self.ID),
			MsgType:    4,
			Md5:        gmd5.MustEncryptBytes(f.FileBytes),
			SendMsg:    nil,
			FileName:   f.Name,
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
	//})
}

// 文件转成byte

func fileErrCallback(tgFileMsgCallbacks []callback.TgMsgCallback, opts SendFileDetail, f SendFileType, msgId int, errMsg string) []callback.TgMsgCallback {
	msgCallback := callback.TgMsgCallback{
		Sender:     opts.Sender,
		Receiver:   opts.Receiver,
		SendTime:   time.Now(),
		ReqId:      strconv.Itoa(msgId),
		Read:       2,
		Initiator:  opts.Sender,
		MsgType:    4,
		FileName:   f.Name,
		SendMsg:    nil,
		FileType:   f.FileType,
		SendStatus: 2,
		Comment:    errMsg,
	}
	tgFileMsgCallbacks = append(tgFileMsgCallbacks, msgCallback)
	return tgFileMsgCallbacks
}
