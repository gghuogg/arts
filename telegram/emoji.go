package telegram

import (
	"arthas/callback"
	"arthas/protobuf"
	"context"
	"github.com/go-kit/log/level"
	"github.com/gotd/td/telegram"
)

type GetEmojiGroupsDetail struct {
	Sender uint64
	Res    chan *protobuf.ResponseMessage
}

func (m *Manager) GetEmojiGroups(ctx context.Context, c *telegram.Client, opts *GetEmojiGroupsDetail) (result []callback.TgGetEmojiGroups, err error) {

	tgGetEmojiGroupsCallBacks := make([]callback.TgGetEmojiGroups, 0)

	groups, err := c.API().MessagesGetEmojiGroups(ctx, 0)
	if err != nil {
		level.Error(m.logger).Log("msg", "MessagesGetEmojiGroups err:", "err", err)
		return nil, err
	}
	modified, b := groups.AsModified()
	if b {
		//fmt.Println(modified.Groups)
		//em := &tg.ReactionEmoji{}
		for _, group := range modified.Groups {

			tgGetEmojiGroupsCallBacks = append(tgGetEmojiGroupsCallBacks, callback.TgGetEmojiGroups{
				Title:       group.Title,
				IconEmojiID: group.IconEmojiID,
				Emoticons:   group.Emoticons,
			})
		}

	}
	//fmt.Println(tgGetEmojiGroupsCallBacks)

	return tgGetEmojiGroupsCallBacks, nil
}
