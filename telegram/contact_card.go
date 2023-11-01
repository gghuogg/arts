package telegram

import (
	"arthas/callback"
	"context"
	"crypto/rand"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"strconv"
	"time"
)

type ContactCardDetail struct {
	Sender       uint64
	Receiver     string
	ContactCards []ContactCard
}

type ContactCard struct {
	FirstName   string
	LastName    string
	PhoneNumber string
}

func (m *Manager) SendContactCard(ctx context.Context, c *telegram.Client, opts ContactCardDetail) *TgSendMsgError {
	tgCardMsgCallbacks := make([]callback.TgCardMsgCallback, 0)
	//value, ok := m.accountsSync.Load(strconv.FormatUint(opts.Sender, 10))
	//if !ok {
	//	level.Error(m.logger).Log("msg", "TG Account is not exist", "user", strconv.FormatUint(opts.Sender, 10))
	//	return errors.New("account information does not exist")
	//}
	//account := value.(*Account)
	//
	//appId, _ := uint64ToIntSafe(account.AppId)
	//m.kvd.WithNs(strconv.FormatUint(opts.Sender, 10))
	//c, _, err := m.Initialization(ctx, false, appId, account.AppHash, nil, ratelimit.New(rate.Every(time.Millisecond*400), 2))
	//if err != nil {
	//	level.Error(m.logger).Log("msg", "init client fail", "err", err)
	//	return err
	//}
	//return m.RunWithAuth(ctx, c, func(ctx context.Context) error {
	s := message.NewSender(c.API())
	manager := peers.Options{Storage: storage.NewPeers(m.kvd)}.Build(c.API())

	contacts, err := c.API().ContactsGetContacts(ctx, 0)
	if err != nil {
		level.Error(m.logger).Log("msg", "get all contacts fail", "err", err)
		return &TgSendMsgError{Err: err}
	}
	modified, _ := contacts.AsModified()

	t := &tg.User{}
	for _, user := range modified.Users {
		t = user.(*tg.User)
		//发送消息前将历史消息标记已读
		c.API().MessagesReadHistory(ctx, &tg.MessagesReadHistoryRequest{Peer: t.AsInputPeer()})
		phone, _ := t.GetPhone()

		if phone == opts.Receiver {
			userName := t.Username
			peer, err := utils.Telegram.GetInputPeer(ctx, manager, userName)
			if err != nil {
				return &TgSendMsgError{Err: err}
			}

			for _, card := range opts.ContactCards {
				id64, _ := RandInt64(rand.Reader)
				id := int(id64)

				contactCard := tg.InputMediaContact{
					FirstName:   card.FirstName,
					PhoneNumber: card.PhoneNumber,
					LastName:    card.LastName,
				}

				//expectSendMediaAndText(t, &contact, mock, "че с деньгами?", &tg.MessageEntityBold{
				//	Length: utf8.RuneCountInString("че с деньгами?"),
				//})

				_, err := s.To(peer.InputPeer()).Media(ctx, message.Contact(contactCard))
				if err != nil {
					return &TgSendMsgError{Err: err, ReqId: id, SendMsg: gjson.MustEncode(contactCard)}
				}
				msgCallback := callback.TgCardMsgCallback{
					Sender:        opts.Sender,
					Receiver:      opts.Receiver,
					SendTime:      time.Now(),
					Read:          1,
					Initiator:     opts.Sender,
					ReqId:         strconv.Itoa(id),
					MsgType:       3,
					SendMsg:       gjson.MustEncode(contactCard),
					CardFirstName: card.FirstName,
					CardPhone:     card.PhoneNumber,
					CardLastName:  card.LastName,
					SendStatus:    1,
				}
				tgCardMsgCallbacks = append(tgCardMsgCallbacks, msgCallback)
			}
		}
	}
	if len(tgCardMsgCallbacks) > 0 {
		m.sendTgCardMsgCallback(tgCardMsgCallbacks)
	}
	return nil
	//})
}
