package telegram

import (
	"arthas/callback"
	"arthas/protobuf"
	"bytes"
	"context"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/tmedia"
	"strconv"
)

type DownChatFileDetail struct {
	ChatId    int64
	Account   uint64
	MessageId int64
	ResChan   chan DownChatFileRes
}

type DownChatFileRes struct {
	ResMsg *protobuf.ResponseMessage
}

func (m *Manager) getDownloadFile(ctx context.Context, c *telegram.Client, opts DownChatFileDetail) error {
	fileMsg := callback.TgMsgCallback{}
	peer, err := m.getPeer(ctx, c, strconv.FormatInt(opts.ChatId, 10))

	if err != nil {
		level.Error(m.logger).Log("msg", "get dialog peer err ", "err", err)
		opts.ResChan <- DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}}
		return err
	}

	it := query.Messages(c.API()).
		GetHistory(peer.InputPeer()).OffsetID(int(opts.MessageId + 1)).
		BatchSize(1).Iter()
	// get one message
	if !it.Next(ctx) {
		//return it.Err()
		opts.ResChan <- DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: it.Err().Error()}}
		return err
	}

	message, ok := it.Value().Msg.(*tg.Message)
	if !ok {
		level.Error(m.logger).Log("msg", "get message err:msg is not *tg.Message ", "err", err)
		opts.ResChan <- DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "get message err:msg is not *tg.Message"}}

		return err
	}

	if message.ID != int(opts.MessageId) {
		level.Error(m.logger).Log("msg", "msg may be deleted ", "id", opts.MessageId)
		opts.ResChan <- DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "msg may be deleted"}}
		return err
	}

	var downloadRes DownChatFileRes

	media, _ := tmedia.GetMedia(message)
	if media != nil {
		output := new(bytes.Buffer)
		_, err := downloader.NewDownloader().WithPartSize(512*1024).
			Download(c.API(), media.InputFileLoc).
			WithThreads(4).
			Stream(ctx, output)
		if err == nil {
			fileMsg.SendMsg = output.Bytes()
			fileMsg.FileName = media.Name
			fileMsg.ReqId = strconv.FormatInt(opts.MessageId, 10)
			downloadRes = DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(fileMsg)}}
		} else {
			downloadRes = DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}}
		}
	} else {
		downloadRes = DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "can not get media from  message"}}
	}

	opts.ResChan <- downloadRes
	return nil
}
