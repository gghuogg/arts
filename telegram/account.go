package telegram

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/util"
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/gotd/td/telegram"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"strconv"
)

type Account struct {
	logger      log.Logger
	tgm         *Manager
	Stream      protobuf.XmppStream_ConnectClient
	PhoneNumber string
	LastName    string
	FirstName   string
	UserID      int64
	UserName    string
	AppId       uint64
	AppHash     string
	IsLogin     bool
	proxyUrl    string
	Device      telegram.DeviceConfig
}

type GetOnlineAccountDetail struct {
	Phone []string
	Res   chan *protobuf.ResponseMessage
}

func (m *Manager) GetOnlineAccounts(ctx context.Context, opts *GetOnlineAccountDetail) (result []callback.GetServerOnlineAccount, err error) {

	result = make([]callback.GetServerOnlineAccount, 0)
	get := etcd.GetReq{
		Key:     "/services/" + m.env + "/" + m.ns + "/" + m.appName + "/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	m.getChan <- get
	etcdResult := <-get.ResChan

	if etcdResult.Result.Kvs != nil {
		for _, v := range etcdResult.Result.Kvs {
			tmp := &protobuf.AccountLogin{}
			proto.Unmarshal(v.Value, tmp)
			if len(opts.Phone) == 0 {
				result = append(result, callback.GetServerOnlineAccount{
					TgId:      tmp.UserID,
					Username:  tmp.UserName,
					FirstName: tmp.FirstName,
					LastName:  tmp.LastName,
					Phone:     strconv.FormatUint(tmp.UserJid, 10),
				})
			} else {
				if util.InSlice(opts.Phone, strconv.FormatUint(tmp.UserJid, 10)) {
					result = append(result, callback.GetServerOnlineAccount{
						TgId:      tmp.UserID,
						Username:  tmp.UserName,
						FirstName: tmp.FirstName,
						LastName:  tmp.LastName,
						Phone:     strconv.FormatUint(tmp.UserJid, 10),
					})
				}
			}

		}
	} else {
		return result, fmt.Errorf("have no online account")
	}
	fmt.Println(result)
	return result, nil
}
