package telegram

import (
	"arthas/callback"
	"arthas/protobuf"
	"context"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"regexp"
	"strconv"
)

type CreateChannelDetail struct {
	Creator            uint64
	ChannelTitle       string
	ChannelUserName    string
	ChannelDescription string
	IsChannel          bool
	IsSuperGroup       bool
	ChannelMember      []string
	Res                chan *protobuf.ResponseMessage
}

func (m *Manager) CreateChannel(ctx context.Context, c *telegram.Client, opts *CreateChannelDetail) (result callback.TgCreateChannel, err error) {

	createChannel, err := c.API().ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{Broadcast: opts.IsChannel, Megagroup: opts.IsSuperGroup, Title: opts.ChannelTitle, About: opts.ChannelDescription})
	if err != nil {
		level.Error(m.logger).Log("msg", "CreateChannel err:", "err", err)
		return callback.TgCreateChannel{}, err
	}
	for _, chat := range createChannel.(*tg.Updates).Chats {
		switch chat.(type) {
		case *tg.Channel:
			channel := chat.(*tg.Channel)
			if channel.Title == opts.ChannelTitle {
				if opts.ChannelUserName != "" {
					update, updateNameErr := c.API().ChannelsUpdateUsername(ctx, &tg.ChannelsUpdateUsernameRequest{Channel: channel.AsInput(), Username: opts.ChannelUserName})
					if updateNameErr != nil {
						_, deleteErr := c.API().ChannelsDeleteChannel(ctx, channel.AsInput())
						if deleteErr != nil {
							level.Error(m.logger).Log("msg", "ChannelsDeleteChannel err:", "err", updateNameErr)
						}
						level.Error(m.logger).Log("msg", "ChannelsUpdateUsername err:", "err", updateNameErr)
						return callback.TgCreateChannel{}, updateNameErr
					}
					if update == false {
						_, deleteErr := c.API().ChannelsDeleteChannel(ctx, channel.AsInput())
						if deleteErr != nil {
							level.Error(m.logger).Log("msg", "ChannelsDeleteChannel err:", "err", updateNameErr)
						}
						level.Info(m.logger).Log("msg", "update username")
						return callback.TgCreateChannel{}, fmt.Errorf("the username:%v update fail", opts.ChannelUserName)
					}
				}

				users := make([]tg.InputUserClass, 0)
				for _, member := range opts.ChannelMember {
					peer, getPeerErr := m.getPeer(ctx, c, member)
					if getPeerErr != nil {
						level.Warn(m.logger).Log("msg", "Failed to get peer", getPeerErr)
						return callback.TgCreateChannel{}, getPeerErr
					}

					switch peer.InputPeer().(type) {
					case *tg.InputPeerUser:
						u := peer.InputPeer().(*tg.InputPeerUser)

						tmp := &tg.InputUser{
							UserID:     u.UserID,
							AccessHash: u.AccessHash,
						}

						users = append(users, tmp)
					}

				}

				res, inviteErr := c.API().ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{Channel: channel.AsInput(), Users: users})
				if inviteErr != nil {
					level.Warn(m.logger).Log("msg", "ChannelsInviteToChannel fail", inviteErr)
					return callback.TgCreateChannel{}, inviteErr
				}
				switch res.(type) {
				case *tg.Updates:
					d := res.(*tg.Updates)
					for _, c1 := range d.Chats {
						switch c1.(type) {
						case *tg.Channel:
							d1 := c1.(*tg.Channel)
							tmp := callback.TgCreateChannel{
								ChannelId:       d1.ID,
								AccessHash:      d1.AccessHash,
								ChannelUserName: d1.Username,
								ChannelTitle:    d1.Title,
								Link:            "https://t.me/" + d1.Username,
								ErrInfo:         "",
							}
							result = tmp
						}
					}
				}
			}
		}
	}

	return result, nil
}

type InviteToChannelDetail struct {
	Invite  uint64
	Channel string
	Invited []string
	Res     chan *protobuf.ResponseMessage
}

func (m *Manager) InviteToChannel(ctx context.Context, c *telegram.Client, opts *InviteToChannelDetail) (result callback.TgInviteChannel, err error) {
	channel := &tg.InputChannel{}
	users := make([]tg.InputUserClass, 0)
	self, _ := c.Self(ctx)
	chats, err := c.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{OffsetDate: 0, OffsetID: 0, OffsetPeer: self.AsInputPeer(), Hash: 0})
	if err != nil {
		level.Error(m.logger).Log("msg", "get all dialogs fail", "err", err)
		return result, err
	}
	asModified, _ := chats.AsModified()

	for _, chat := range asModified.GetChats() {
		switch chat.(type) {
		case *tg.Channel:
			t := chat.(*tg.Channel)
			id, _ := strconv.ParseInt(opts.Channel, 10, 64)
			if t.ID == id || t.Username == opts.Channel {
				channel.ChannelID = t.ID
				channel.AccessHash = t.AccessHash
			}

		case *tg.Chat:
		}
	}

	for _, member := range opts.Invited {
		peer, getPeerErr := m.getPeer(ctx, c, member)
		if getPeerErr != nil {
			level.Warn(m.logger).Log("msg", "Failed to get peer", getPeerErr)
			return result, getPeerErr
		}

		switch peer.InputPeer().(type) {
		case *tg.InputPeerUser:
			u := peer.InputPeer().(*tg.InputPeerUser)

			tmp := &tg.InputUser{
				UserID:     u.UserID,
				AccessHash: u.AccessHash,
			}

			users = append(users, tmp)
		}

	}

	_, inviteErr := c.API().ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{Channel: channel, Users: users})
	if inviteErr != nil {
		level.Warn(m.logger).Log("msg", "ChannelsInviteToChannel fail", inviteErr)
		return result, inviteErr
	}
	//fmt.Println(res)

	return result, nil
}

type JoinByLinkDetail struct {
	Sender uint64
	Links  []string
	Res    chan *protobuf.ResponseMessage
}

func (m *Manager) JoinPublicChannelByLink(ctx context.Context, c *telegram.Client, opts *JoinByLinkDetail) (result []callback.TgJoinPublicChannelByLink, err error) {

	result = make([]callback.TgJoinPublicChannelByLink, 0)
	for _, link := range opts.Links {
		regex := regexp.MustCompile(`(?:t|telegram)\.(?:me|dog)/(joinchat/|\+)?([\w-]+)`)
		if regex.MatchString(link) {
			matches := regex.FindStringSubmatch(link)
			if len(matches) >= 2 {
				username := matches[len(matches)-1]

				resolveUsername, rErr := c.API().ContactsResolveUsername(ctx, username)
				if rErr != nil {
					level.Warn(m.logger).Log("msg", "ContactsResolveUsername fail", err)
					return result, rErr
				}
				for _, v := range resolveUsername.Chats {
					switch v.(type) {
					case *tg.Channel:
						//fmt.Println(v.(*tg.Channel))
						tmp := v.(*tg.Channel)

						res, errJoin := c.API().ChannelsJoinChannel(ctx, tmp.AsInput())
						if errJoin != nil {
							level.Warn(m.logger).Log("msg", "ChannelsJoinChannel fail", err)
							return result, err
						}
						switch res.(type) {
						case *tg.Updates:
							for _, chat := range res.(*tg.Updates).Chats {
								d, b := chat.AsNotEmpty()
								if b {
									switch d.(type) {
									case *tg.Channel:
										ch := chat.(*tg.Channel)
										result = append(result, callback.TgJoinPublicChannelByLink{
											TgId:       ch.ID,
											AccessHash: ch.AccessHash,
											Title:      ch.Title,
											UserName:   link,
										})
									}
								}

							}
						}
					}
				}

			}
		} else {
			return result, fmt.Errorf("link format error")
		}
	}

	fmt.Println("resutl:", result)

	return result, nil
}

type LeaveDetail struct {
	Sender uint64
	Leaves []string
	Res    chan *protobuf.ResponseMessage
}

func (m *Manager) Leave(ctx context.Context, c *telegram.Client, opts *LeaveDetail) (result callback.TgLeave, err error) {

	for _, leave := range opts.Leaves {
		peer, peerErr := m.getPeer(ctx, c, leave)
		if peerErr != nil {
			return result, peerErr
		}
		inputChannel := &tg.InputChannel{}
		var chatId int64
		switch peer.InputPeer().(type) {
		case *tg.InputPeerChannel:
			channel := peer.InputPeer().(*tg.InputPeerChannel)
			inputChannel.ChannelID = channel.ChannelID
			inputChannel.AccessHash = channel.AccessHash
		case *tg.InputPeerChat:
			chat := peer.InputPeer().(*tg.InputPeerChat)
			chatId = chat.ChatID
		}

		if chatId != 0 {
			self, sErr := c.Self(ctx)
			if sErr != nil {
				return result, sErr
			}
			_, err1 := c.API().MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{ChatID: chatId, UserID: self.AsInput()})
			if err1 != nil {
				return result, err1
			}

		}

		if inputChannel.ChannelID != 0 {
			_, err2 := c.API().ChannelsLeaveChannel(ctx, inputChannel)
			if err2 != nil {
				return result, err2
			}

		}
	}

	return result, nil
}

type GetChannelMemberDetail struct {
	Sender     uint64
	Channel    string
	Offset     int
	Limit      int
	SearchType string
	TopMsgId   int
	Res        chan *protobuf.ResponseMessage
}

func (m *Manager) GetChannelMember(ctx context.Context, c *telegram.Client, opts *GetChannelMemberDetail) (result []callback.TgGetChannelMember, err error) {

	result = make([]callback.TgGetChannelMember, 0)

	peer, peerErr := m.getPeer(ctx, c, opts.Channel)
	if peerErr != nil {
		return result, peerErr
	}
	inputChannel := &tg.InputChannel{}

	switch peer.InputPeer().(type) {
	case *tg.InputPeerChannel:
		channel := peer.InputPeer().(*tg.InputPeerChannel)
		inputChannel.ChannelID = channel.ChannelID
		inputChannel.AccessHash = channel.AccessHash
	}

	var filter Filter
	switch opts.SearchType {
	case "name":
		filter = &NameFilter{}
	case "recent":
		filter = &RecentFilter{}
	case "admins":
		filter = &AdminsFilter{}
	case "kicked":
		filter = &KickedFilter{}
	case "bots":
		filter = &BotFilter{}
	case "banned":
		filter = &BannedFilter{}
	case "contacts":
		filter = &ContactsFilter{}
	case "mentions":
		if opts.TopMsgId != 0 {
			filter = &MentionsFilter{
				TopMsgID: opts.TopMsgId,
			}
		} else {
			filter = &MentionsFilter{}
		}
	}

	info, err := c.API().ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{Channel: inputChannel,
		Filter: filter.Apply(), Offset: opts.Offset, Limit: opts.Limit, Hash: 0})
	if err != nil {
		return result, err
	}
	//fmt.Println("user:", chats.Users)
	//fmt.Println("chats:", chats.Chats)
	participants, b := info.AsModified()
	if b {
		for _, user := range participants.Users {
			t, b1 := user.AsNotEmpty()
			if !b1 {
				continue
			}
			result = append(result, callback.TgGetChannelMember{
				TgId:      t.ID,
				Username:  t.Username,
				FirstName: t.FirstName,
				LastName:  t.LastName,
				Phone:     t.Phone,
				//Photo:     t.Photo.String(),
			})
		}

	}
	fmt.Println("result:", result)

	return result, nil
}
