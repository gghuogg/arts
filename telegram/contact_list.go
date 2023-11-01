package telegram

import (
	"arthas/callback"
	"arthas/protobuf"
	"context"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gotd/td/telegram"
	"strconv"
)

type ContactListDetail struct {
	Account uint64
	ResChan chan ContactListRes
}

type ContactListRes struct {
	AccountResult *protobuf.ResponseMessage
	Users         []ContactUserDetail
}

type ContactUserDetail struct {
	TgId      int64  `json:"tgId"          description:"tg id"`
	Username  string `json:"username"      description:"账号号码"`
	FirstName string `json:"firstName"     description:"First Name"`
	LastName  string `json:"lastName"      description:"Last Name"`
	Phone     string `json:"phone"         description:"手机号"`
	Photo     string `json:"photo"         description:"账号头像"`
	Comment   string `json:"comment"       description:"备注"`
}

func (m *Manager) GetContactList(ctx context.Context, c *telegram.Client, opts ContactListDetail) error {
	users := make([]ContactUserDetail, 0)
	tgContactsCallbacks := make([]callback.TgContactsCallback, 0)

	contacts, err := c.API().ContactsGetContacts(ctx, 0)
	if err != nil {
		level.Error(m.logger).Log("msg", "get contacts fail", err)
		accountResult := &protobuf.ResponseMessage{
			ActionResult: protobuf.ActionResult_ALL_FAIL, Data: gjson.MustEncode(users),
			Comment: strconv.FormatUint(opts.Account, 10) + "get contacts fail...",
		}
		opts.ResChan <- ContactListRes{AccountResult: accountResult}
		return err
	}
	modified, _ := contacts.AsModified()

	for _, user := range modified.Users {
		t, b := user.AsNotEmpty()
		if !b {
			continue
		}
		users = append(users, ContactUserDetail{
			TgId:      t.ID,
			Username:  t.Username,
			FirstName: t.FirstName,
			LastName:  t.LastName,
			Phone:     t.Phone,
		})

		tgContactsCallbacks = append(tgContactsCallbacks, callback.TgContactsCallback{
			TgId:      t.ID,
			Username:  t.Username,
			FirstName: t.FirstName,
			LastName:  t.LastName,
			Phone:     t.Phone,
			Type:      1,
		})
	}
	accountResult := &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(users)}
	opts.ResChan <- ContactListRes{AccountResult: accountResult}

	return nil
}
