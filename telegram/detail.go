package telegram

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"context"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"google.golang.org/protobuf/proto"
	"strconv"
)

type AccountDetail struct {
	PhoneNumber string
	AppId       uint64
	AppHash     string
}

type LoginDetail struct {
	PhoneNumber string
	ProxyUrl    string
	LoginId     string
	ResChan     chan LoginDetailRes
}
type LoginDetailRes struct {
	LoginStatus protobuf.AccountStatus
	LoginId     string
	Comment     string
}

const (
	appID   = 19827791
	appHash = "c3d8d33877358f1edef69007967e1825"
)

func (m *Manager) syncAccountDetail(detail AccountDetail) {
	if detail.AppId == 0 && detail.AppHash == "" {
		detail.AppId = appID
		detail.AppHash = appHash
	}

	a := &AccountDetail{
		PhoneNumber: detail.PhoneNumber,
		AppId:       detail.AppId,
		AppHash:     detail.AppHash,
	}
	m.accountsSync.Store(detail.PhoneNumber, a)

	//保存etcd
	_ = m.insertDetail(detail)
}

func (m *Manager) insertDetail(detail AccountDetail) (err error) {
	protocDetail := protobuf.TgAccountDetail{
		PhoneNumber: detail.PhoneNumber,
		AppHash:     detail.AppHash,
		AppId:       detail.AppId,
	}
	protoData, err := proto.Marshal(&protocDetail)
	if err != nil {
		level.Error(m.logger).Log("msg", "Failed insert  to etcd.", "err", err)
		return
	}
	key := fmt.Sprintf("/arthas/tg_accountdetail/%v", detail.PhoneNumber)
	reqModel := etcd.TxnInsertReq{
		Key:     key,
		Value:   string(protoData),
		Options: nil,
	}
	m.txnInsertChan <- reqModel
	return
}

type SyncContact struct {
	Account  uint64
	Contacts []uint64
	ResChan  chan *protobuf.ResponseMessage
}

func (m *Manager) AddContact(ctx context.Context, c *telegram.Client, opts SyncContact) error {

	tgContactsCallbacks := make([]callback.TgContactsCallback, 0)
	importList := make([]tg.InputPhoneContact, 0)

	for _, contact := range opts.Contacts {
		phone := strconv.FormatUint(contact, 10)
		importInfo := tg.InputPhoneContact{
			Phone: phone,
		}

		importList = append(importList, importInfo)
	}

	contacts, err := c.API().ContactsImportContacts(ctx, importList)
	if err != nil {
		level.Error(m.logger).Log("msg", "import contact fail", "err", err)
		return err
	}

	for _, user := range contacts.Users {
		t := user.(*tg.User)
		level.Info(m.logger).Log("msg", "user information ", "info", t)

		contactsMsg := callback.TgContactsCallback{
			TgId:      t.ID,
			Username:  t.Username,
			FirstName: t.FirstName,
			LastName:  t.LastName,
			Phone:     t.Phone,
			Type:      1,
		}
		tgContactsCallbacks = append(tgContactsCallbacks, contactsMsg)
	}

	if len(tgContactsCallbacks) > 0 {
		var syncMap = make(map[uint64][]callback.TgContactsCallback)
		syncMap[opts.Account] = tgContactsCallbacks
		m.contactsSyncCallback(syncMap)
	}
	return nil
	//})
}
func (m *Manager) RunWithAuth(ctx context.Context, client *telegram.Client, f func(ctx context.Context) error) error {
	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("not authorized. please login first")
		}

		//logger.From(ctx).Info("Authorized",
		//	zap.Int64("id", status.User.ID),
		//	zap.String("username", status.User.Username))

		return f(ctx)
	})
}
