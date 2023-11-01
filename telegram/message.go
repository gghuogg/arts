package telegram

import (
	"arthas/callback"
	"arthas/protobuf"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"strconv"
	"time"
)

type SendMessageDetail struct {
	Sender   uint64
	Receiver string
	SendText []string
	ResChan  chan *protobuf.ResponseMessage
}

func (m *Manager) SendMessage(ctx context.Context, c *telegram.Client, opts SendMessageDetail) error {

	tgTextMsgCallbacks := make([]callback.TgMsgCallback, 0)

	s := message.NewSender(c.API())
	self, err := c.Self(ctx)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get self", err)
		return err
	}
	peer, err := m.getPeer(ctx, c, opts.Receiver)
	if err != nil {
		level.Warn(m.logger).Log("msg", "Failed to get peer", err)
		return err
	}
	//发送消息前将历史消息标记已读(标记与对方的聊天记录就可以了)
	_, _ = c.API().MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{Peer: peer.InputPeer()})
	for _, msg := range opts.SendText {

		id64, _ := RandInt64(rand.Reader)
		id := int(id64)
		_ = s.To(peer.InputPeer()).TypingAction().Typing(ctx)

		randomValue := generateRandomValue()
		time.Sleep(time.Duration(randomValue) * time.Millisecond)

		res, err := s.To(peer.InputPeer()).Text(ctx, msg)
		if err != nil {
			m.sendMessageErrCallback(opts, strconv.Itoa(id), msg, err.Error())
			level.Error(m.logger).Log("msg", "send msg err:", "err", err)
		}
		msgCallback := callback.TgMsgCallback{
			Sender:      uint64(self.ID),
			Receiver:    gconv.String(peer.ID()),
			SendMsg:     []byte(msg),
			Read:        1,
			Initiator:   uint64(self.ID),
			SendStatus:  1,
			MsgType:     1,
			AccountType: 1,
		}
		switch msg := res.(type) {
		case *tg.UpdateShortSentMessage:
			id = msg.ID
			msgCallback.ReqId = strconv.Itoa(id)
			msgCallback.SendTime = gtime.NewFromTimeStamp(int64(msg.Date)).Time
		}

		tgTextMsgCallbacks = append(tgTextMsgCallbacks, msgCallback)

	}

	if len(tgTextMsgCallbacks) > 0 {
		m.sendTgMsgCallback(tgTextMsgCallbacks)
	}

	return nil
	//})
}

type SendGroupMessageDetail struct {
	Sender     uint64
	GroupTitle string
	SendText   []string
	ResChan    chan *protobuf.ResponseMessage
}

func (m *Manager) SendMessageInGroup(ctx context.Context, c *telegram.Client, opts SendGroupMessageDetail) error {

	groupMsgCallbacks := make([]callback.GroupMsgCallback, 0)
	s := message.NewSender(c.API())
	peer, err := m.getPeer(ctx, c, opts.GroupTitle)
	if err != nil {
		return fmt.Errorf("failed to get peer: %w", err)
	}

	for _, msg := range opts.SendText {
		id64, _ := RandInt64(rand.Reader)
		id := int(id64)

		s.To(peer.InputPeer()).TypingAction().Typing(ctx)

		randomValue := generateRandomValue()
		time.Sleep(time.Duration(randomValue) * time.Millisecond)

		res, err := s.To(peer.InputPeer()).Text(ctx, msg)

		if err != nil {
			return err
		}
		switch msg := res.(type) {
		case *tg.UpdateShortSentMessage:
			id = msg.ID
		}

		msgCallback := callback.GroupMsgCallback{
			Sender:    opts.Sender,
			Title:     opts.GroupTitle,
			SendText:  msg,
			SendTime:  time.Now(),
			ReqId:     strconv.Itoa(id),
			Read:      1,
			Initiator: opts.Sender,
		}
		groupMsgCallbacks = append(groupMsgCallbacks, msgCallback)
		//fmt.Println(sendMessage.(*tg.UpdateShortSentMessage))
	}

	fmt.Println(groupMsgCallbacks)
	if len(groupMsgCallbacks) > 0 {
		m.sendGroupMsgCallback(groupMsgCallbacks)
	}

	return nil

}

type GetHistoryDetail struct {
	Self       uint64
	Other      string
	Limit      int
	OffsetDate int
	OffsetID   int
	MaxID      int
	MinID      int
	Res        chan *protobuf.ResponseMessage
}

func (m *Manager) GetHistory(ctx context.Context, c *telegram.Client, opts *GetHistoryDetail) (result []callback.TgMsgCallback, err error) {

	result = make([]callback.TgMsgCallback, 0)
	peer, err := m.getPeer(ctx, c, opts.Other)
	if err != nil {
		return nil, err
	}
	res, err := c.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:       peer.InputPeer(),
		Limit:      opts.Limit,
		OffsetDate: opts.OffsetDate,
		OffsetID:   opts.OffsetID,
		MaxID:      opts.MaxID,
		MinID:      opts.MinID,
	})
	if err != nil {
		level.Error(m.logger).Log("msg", "get message history fail", "err", err)
		return nil, err
	}
	self, err := c.Self(ctx)
	if err != nil {
		return nil, err
	}
	messages, b := res.AsModified()
	if !b {
		return
	}
	for _, messageClass := range messages.GetMessages() {
		switch msg := messageClass.(type) {
		case *tg.Message:
			result = append(result, covertMessage(self, msg))
		}

	}

	return
}

type MessageReactionDetail struct {
	Sender   uint64
	Receiver string
	MsgIds   []int
	Emoticon string
	Res      chan *protobuf.ResponseMessage
}

func (m *Manager) MessageReaction(ctx context.Context, c *telegram.Client, opts *MessageReactionDetail) (result callback.TgMessagesReaction, err error) {
	rs := make([]tg.ReactionClass, 0)

	peer, err := m.getPeer(ctx, c, opts.Receiver)
	if err != nil {
		return result, err
	}

	em := &tg.ReactionEmoji{}
	em.Emoticon = opts.Emoticon
	rs = append(rs, em)
	for _, id := range opts.MsgIds {
		_, reactionErr := c.API().MessagesSendReaction(ctx, &tg.MessagesSendReactionRequest{
			Big:         false,
			AddToRecent: true,
			Peer:        peer.InputPeer(),
			MsgID:       id,
			Reaction:    rs,
		})
		if reactionErr != nil {
			return result, reactionErr
		}
	}

	return result, nil
}

// TODO 先处理文本
func covertMessage(self *tg.User, msg *tg.Message) (item callback.TgMsgCallback) {
	var peerId int64
	switch msg.PeerID.(type) {
	case *tg.PeerUser:
		peerId = msg.PeerID.(*tg.PeerUser).UserID
	case *tg.PeerChat:
		peerId = msg.PeerID.(*tg.PeerChat).ChatID
	case *tg.PeerChannel:
		peerId = msg.PeerID.(*tg.PeerChannel).ChannelID
	}

	item = callback.TgMsgCallback{
		Initiator:  uint64(self.ID),
		ReqId:      gconv.String(msg.ID),
		SendTime:   gtime.NewFromTimeStamp(int64(msg.Date)).Time,
		Read:       1,
		SendStatus: 1,
	}
	if msg.Out {
		//自己发出的
		item.Out = 1
		item.Sender = gconv.Uint64(self.ID)
		item.Receiver = gconv.String(peerId)
	} else {
		item.Out = 2
		item.Sender = gconv.Uint64(peerId)
		item.Receiver = gconv.String(self.ID)
	}
	if msg.Message != "" {
		//文本消息
		item.MsgType = 1
		item.SendMsg = []byte(msg.Message)
	} else if msg.Media != nil {
		//媒体消息
		item.MsgType = 2
		item.SendMsg = nil
	}

	return
}

func (m *Manager) sendMessageErrCallback(opts SendMessageDetail, reqId string, msg string, errMsg string) {
	tgTextMsgCallbacks := make([]callback.TgMsgCallback, 0)
	tgTextMsgCallbacks = append(tgTextMsgCallbacks, callback.TgMsgCallback{
		Sender:     opts.Sender,
		Receiver:   opts.Receiver,
		SendStatus: 2,
		SendTime:   time.Now(),
		Initiator:  opts.Sender,
		ReqId:      reqId,
		SendMsg:    []byte(msg),
		Comment:    errMsg,
		Read:       2,
	})
	if len(tgTextMsgCallbacks) > 0 {
		m.sendTgMsgCallback(tgTextMsgCallbacks)
	}
}
