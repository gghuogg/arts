package telegram

import (
	"arthas/callback"
	"arthas/protobuf"
	"context"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"strconv"
)

type CreateGroupDetail struct {
	Initiator    uint64
	GroupTitle   string
	GroupMembers []string
	Res          chan *protobuf.ResponseMessage
}

func (m *Manager) CreateGroup(ctx context.Context, c *telegram.Client, opts *CreateGroupDetail) (result callback.TgCreateGroup, err error) {

	list := make([]tg.InputUserClass, 0)

	contacts, err := c.API().ContactsGetContacts(ctx, 0)
	if err != nil {
		level.Error(m.logger).Log("msg", "get all contacts fail", "err", err)
		return result, err
	}
	modified, _ := contacts.AsModified()

	t := &tg.User{}
	for _, user := range modified.Users {
		t = user.(*tg.User)
		for _, v := range opts.GroupMembers {
			phone := t.Phone
			if phone == v {
				list = append(list, t.AsInput())
			}
		}

	}

	res, err := c.API().MessagesCreateChat(ctx, &tg.MessagesCreateChatRequest{Flags: t.Flags, Users: list, Title: opts.GroupTitle})
	if err != nil {
		level.Error(m.logger).Log("msg", "create group fail", "err", err)
		return result, err
	}

	info, _ := res.(*tg.Updates).Chats[0].(*tg.Chat)

	m.createGroupCallback(callback.CreateGroupCallback{
		ChatId: info.GetID(),
		Title:  info.GetTitle(),
	})

	return result, nil
}

type AddGroupMemberDetail struct {
	Initiator  uint64
	GroupTitle string
	AddMembers []string
	Res        chan *protobuf.ResponseMessage
}

func (m *Manager) AddGroupMember(ctx context.Context, c *telegram.Client, opts *AddGroupMemberDetail) (result callback.TgAddGroupMember, err error) {

	addGroupMemberCallbacks := make([]callback.AddGroupMemberCallback, 0)

	self, _ := c.Self(ctx)
	chats, err := c.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{OffsetDate: 0, OffsetID: 0, OffsetPeer: self.AsInputPeer(), Hash: 0})
	if err != nil {
		level.Error(m.logger).Log("msg", "get all dialogs fail", "err", err)
		return result, err
	}
	asModified, _ := chats.AsModified()

	var chatID int64
	for _, chat := range asModified.GetChats() {
		switch chat.(type) {
		case *tg.Channel:
		case *tg.Chat:
			t := chat.(*tg.Chat)
			if opts.GroupTitle == t.GetTitle() || opts.GroupTitle == strconv.FormatInt(t.GetID(), 10) {
				chatID = t.GetID()
			}
		}
	}

	contacts, err := c.API().ContactsGetContacts(ctx, 0)
	if err != nil {
		level.Error(m.logger).Log("msg", "get all contacts fail", "err", err)
		return result, err
	}
	modified, _ := contacts.AsModified()

	t := &tg.User{}
	for _, user := range modified.Users {
		t = user.(*tg.User)

		for _, v := range opts.AddMembers {
			phone := t.Phone
			if phone == v || v == strconv.FormatInt(t.GetID(), 10) {
				res, addErr := c.API().MessagesAddChatUser(ctx, &tg.MessagesAddChatUserRequest{UserID: t.AsInput(), ChatID: chatID})
				if addErr != nil {
					level.Error(m.logger).Log("msg", "add group member fail", "err", err)
					return result, addErr
				}
				for _, u := range res.(*tg.Updates).Users {
					member := u.(*tg.User)
					msgCallback := callback.AddGroupMemberCallback{
						MemberId:  member.ID,
						UserName:  member.Username,
						FirstName: member.FirstName,
						LastName:  member.LastName,
						Phone:     member.Phone,
					}
					addGroupMemberCallbacks = append(addGroupMemberCallbacks, msgCallback)
				}
			}
		}

	}
	fmt.Println(addGroupMemberCallbacks)
	if len(addGroupMemberCallbacks) > 0 {
		m.addGroupMemberCallback(addGroupMemberCallbacks)
	}

	return result, nil

}

type GetGroupMembers struct {
	Initiator uint64
	ChatId    int64
	Res       chan *protobuf.ResponseMessage
}

func (m *Manager) GetGroupMembers(ctx context.Context, c *telegram.Client, opts *GetGroupMembers) (result []callback.TgGetGroupMember, err error) {

	result = make([]callback.TgGetGroupMember, 0)
	chats, err := c.API().MessagesGetFullChat(ctx, opts.ChatId)
	if err != nil {
		return result, err
	}

	for _, user := range chats.Users {
		t, b := user.AsNotEmpty()
		if !b {
			continue
		}
		result = append(result, callback.TgGetGroupMember{
			TgId:      t.ID,
			Username:  t.Username,
			FirstName: t.FirstName,
			LastName:  t.LastName,
			Phone:     t.Phone,
		})
	}

	return result, nil
}
