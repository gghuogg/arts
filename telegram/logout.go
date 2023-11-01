package telegram

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"context"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/gotd/td/telegram"
	"google.golang.org/protobuf/proto"
	"strconv"
)

type TgLogoutDetail struct {
	UserJid  uint64
	ProxyUrl string
	ResChan  chan LogoutRes
}

type LogoutRes struct {
	LogoutStatus protobuf.AccountStatus
	Comment      string
}

func (m *Manager) TgLogout(ctx context.Context, c *telegram.Client, opts TgLogoutDetail) error {

	user, err := c.Self(ctx)
	if err != nil {
		level.Error(m.logger).Log("msg", "get user info fail", "err", err)
		return err
	}

	m.accountsSync.Delete(strconv.FormatUint(opts.UserJid, 10))
	m.clientMap.Delete(strconv.FormatUint(opts.UserJid, 10))
	err = m.OffOnlineAccount(strconv.FormatUint(opts.UserJid, 10))
	if err != nil {
		level.Error(m.logger).Log("msg", "etcd delete user fail", "err", err)
		return err
	}
	// etcd 减少连接数
	m.reduceConnection()

	tgLogoutCallback := callback.TgUserLogoutCallback{
		IsSuccess:     1,
		TgId:          user.ID,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		AccountStatus: 2,
		IsOnline:      2,
		ProxyAddress:  opts.ProxyUrl,
		Comment:       "logout success..",
	}
	m.logoutMsgCallback(tgLogoutCallback)

	opts.ResChan <- LogoutRes{
		LogoutStatus: protobuf.AccountStatus_SUCCESS, Comment: strconv.FormatUint(opts.UserJid, 10) + " user logout success",
	}
	return nil
}

func (m *Manager) OffOnlineAccount(phoneNum string) error {

	delReq := etcd.DeleteReq{
		Key:     "/service/" + m.env + "/" + m.ns + "/online/" + m.ListenAddress + "/" + phoneNum,
		Options: nil,
		ResChan: make(chan etcd.DeleteRes),
	}
	m.deleteChan <- delReq

	delResult := <-delReq.ResChan
	if delResult.Err != nil {
		return delResult.Err
	}

	return nil
}

func (m *Manager) reduceConnection() {
	key := fmt.Sprintf("/service/%v/%v/%v/%v/%v/%v", m.env, m.ns, "count", m.appName, m.vs, m.ListenAddress)
	//记录连接数
	putValue := &protobuf.ServiceInfo{}
	getRes := &protobuf.ServiceInfo{}
	getReq := etcd.GetReq{
		Key:     key,
		Options: nil,
		ResChan: make(chan etcd.GetRes),
	}
	m.getChan <- getReq
	getResult := <-getReq.ResChan
	if getResult.Result.Kvs != nil {
		proto.Unmarshal(getResult.Result.Kvs[0].Value, getRes)
		tmp := &protobuf.ServiceInfo{
			IP:          m.ListenAddress,
			Connections: getRes.Connections - 1,
		}
		putValue = tmp
	}

	putvalueb, _ := proto.Marshal(putValue)

	putReq := etcd.PutReq{
		Key:     key,
		Value:   string(putvalueb),
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	m.putChan <- putReq

}
