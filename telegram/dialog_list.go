package telegram

import (
	"arthas/protobuf"
	"context"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"strconv"
)

type DialogListDetail struct {
	Account uint64
	ResChan chan DialogListRes
}

type DialogListRes struct {
	AccountResult *protobuf.ResponseMessage
	Elems         []dialogs.Elem
}

type DialogDetail struct {
	TgId             int64  `json:"tgId"      dc:"tg id"`
	Username         string `json:"username"  dc:"username"`
	FirstName        string `json:"firstName" dc:"First Name"`
	LastName         string `json:"lastName"  dc:"Last Name"`
	Phone            string `json:"phone"     dc:"phone"`
	Type             int    `json:"type"      dc:"type"`
	Creator          bool   `json:"creator"   dc:"creator"`
	AdminRightsFlags string `json:"adminRightsFlags" dc:"adminRights Flags"`
}

func (m *Manager) GetDialogList(ctx context.Context, c *telegram.Client, opts DialogListDetail) error {
	result := make([]*DialogDetail, 0)
	dialogs, err := query.GetDialogs(c.API()).BatchSize(100).Collect(ctx)
	manager := peers.Options{Storage: storage.NewPeers(m.kvd)}.Build(c.API())
	for _, d := range dialogs {
		id := utils.Telegram.GetInputPeerID(d.Peer)

		// we can update our access hash state if there is any new peer.
		if err = applyPeers(ctx, manager, d.Entities, id); err != nil {
			level.Error(m.logger).Log("msg", "failed to apply peer updates", "err", err)
		}

		var r *DialogDetail
		switch t := d.Peer.(type) {
		case *tg.InputPeerUser:
			r = processUser(t.UserID, d.Entities)
		case *tg.InputPeerChannel:
			r = processChannel(t.ChannelID, d.Entities)
		case *tg.InputPeerChat:
			r = processChat(t.ChatID, d.Entities)
		}
		// skip unsupported types
		if r == nil {
			continue
		}
		result = append(result, r)
	}
	if err != nil {
		level.Error(m.logger).Log("msg", "get contacts fail", err)
		accountResult := &protobuf.ResponseMessage{
			ActionResult: protobuf.ActionResult_ALL_FAIL, Data: gjson.MustEncode(result),
			Comment: strconv.FormatUint(opts.Account, 10) + "get dialogs fail...",
		}
		opts.ResChan <- DialogListRes{AccountResult: accountResult}
		return err
	}

	accountResult := &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}

	opts.ResChan <- DialogListRes{AccountResult: accountResult}

	return nil
}

func processUser(id int64, entities peer.Entities) *DialogDetail {
	u, ok := entities.User(id)
	if !ok {
		return nil
	}
	return &DialogDetail{
		TgId:      u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Phone:     u.Phone,
		Type:      1,
	}
}

func processChannel(id int64, entities peer.Entities) *DialogDetail {
	c, ok := entities.Channel(id)
	if !ok {
		return nil
	}

	d := &DialogDetail{
		TgId:             c.ID,
		FirstName:        c.Title,
		Username:         c.Username,
		Creator:          c.Creator,
		AdminRightsFlags: c.AdminRights.Flags.String(),
	}

	// channel type
	switch {
	case c.Broadcast:
		d.Type = 3
	case c.Megagroup, c.Gigagroup:
		d.Type = 2
	default:
		d.Type = 2
	}
	return d
}

func processChat(id int64, entities peer.Entities) *DialogDetail {
	c, ok := entities.Chat(id)
	if !ok {
		return nil
	}

	return &DialogDetail{
		TgId:             c.ID,
		FirstName:        c.Title,
		Type:             2,
		Creator:          c.Creator,
		AdminRightsFlags: c.AdminRights.Flags.String(),
	}
}

func applyPeers(ctx context.Context, manager *peers.Manager, entities peer.Entities, id int64) error {
	users := make([]tg.UserClass, 0, 1)
	if user, ok := entities.User(id); ok {
		users = append(users, user)
	}

	chats := make([]tg.ChatClass, 0, 1)
	if chat, ok := entities.Chat(id); ok {
		chats = append(chats, chat)
	}

	if chat, ok := entities.Channel(id); ok {
		chats = append(chats, chat)
	}

	return manager.Apply(ctx, users, chats)
}
