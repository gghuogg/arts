package signal

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-kit/log/level"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func (m *Manager) loginSync(detail LoginDetail) {
	defer func() {
		err := recover()
		if err != nil {
			level.Error(m.logger).Log("loginSync painc", err)
			m.callbackChan.LoginCallbackChan <- []callback.LoginCallback{{UserJid: detail.UserJid, ProxyUrl: detail.ProxyUrl,
				LoginStatus: protobuf.AccountStatus_UNKNOWN, Comment: fmt.Sprintf("%v", err)}}
		}
	}()

	value, ok := m.accountsSync.Load(detail.UserJid)
	if !ok {
		level.Info(m.logger).Log("msg", "Account is not exist", "user", detail.UserJid)
		return
	}
	account := value.(*Account)

	if account.IsLogin == true {
		level.Info(m.logger).Log("msg", "Account is online,will not continue", "user", detail.UserJid)
		return
	}

	conn, err := m.grpcPool.Get(m.context)
	if err != nil {
		level.Info(m.logger).Log("msg", "Did not connect", "err", err)
	}
	//conn, err := grpc.Dial(m.grpcServer, grpc.WithTransportCredentials(insecure.NewCredentials()),
	//	grpc.WithInitialWindowSize(1<<30),
	//	grpc.WithInitialConnWindowSize(1<<30),
	//)
	//if err != nil {
	//	level.Info(m.logger).Log("msg", "Did not connect", "err", err)
	//}
	defer conn.Close()
	loginClientPayload, _ := m.CreateLoginClientPayload(detail.UserJid)
	b, _ := proto.Marshal(loginClientPayload)

	var md metadata.MD
	client := protobuf.NewXmppStreamClient(conn)
	md = metadata.Pairs(
		headerProxyUrl, detail.ProxyUrl,
		headerClientPayload, hex.EncodeToString(b),
		headerPrivateKey, hex.EncodeToString(account.PrivateKey),
		headerResumptionSecret, hex.EncodeToString(account.ResumptionSecret),
	)

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 创建 stream
	stream, err := client.Connect(ctx)
	if err != nil {
		level.Error(m.logger).Log("msg", "could not connect", "err", err)
	}
	account.Stream = stream
	account.logger = m.logger
	account.signal = m
	account.UserJid = detail.UserJid
	account.proxyUrl = detail.ProxyUrl
	account.ClientPayLoad = loginClientPayload
	account.callbackChan = m.callbackChan
	account.Recv()
}

func (m *Manager) GetDetailFromETCD(userJid uint64) (detail *protobuf.AccountDetail, err error) {

	key := fmt.Sprintf("/arthas/accountdetail/%d", userJid)
	get := etcd.GetReq{
		Key:     key,
		Options: nil,
		ResChan: make(chan etcd.GetRes),
	}
	m.getChan <- get
	result := <-get.ResChan

	if result.Result.Kvs == nil || len(result.Result.Kvs) == 0 {
		return nil, errors.New("no IP found for user")
	}

	if result.Result.Kvs != nil {
		proto.Unmarshal(result.Result.Kvs[0].Value, detail)
	}

	return detail, nil
}
