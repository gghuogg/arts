package telegram

import (
	"arthas/etcd"
	"arthas/protobuf"
	"github.com/go-kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
)

func (m *Manager) etcdAccountLoginState(phoneNum string, isLogin bool, proxyUrl string, userId int64, username string, firstName string, lastName string) {
	m.ServiceWithUserJid(phoneNum)
	m.UserJidProxyIp(phoneNum, proxyUrl)
	m.OnlineAccount(phoneNum)

	userJid, _ := strconv.ParseUint(phoneNum, 10, 64)

	req := etcd.LeaseReq{ResChan: make(chan etcd.LeaseRes)}
	m.leaseChan <- req
	res := <-req.ResChan
	lease := res.Result

	leaseGrant, err := lease.Grant(m.context, 45)
	if err != nil {
		level.Error(m.logger).Log("msg", "leaseGrant err", "err", err)
	}

	etcdvalue := &protobuf.AccountLogin{
		IsLogin:    isLogin,
		GrpcServer: m.ListenAddress,
		UserJid:    userJid,
		UserID:     userId,
		UserName:   username,
		FirstName:  firstName,
		LastName:   lastName,
	}
	bvalue, _ := proto.Marshal(etcdvalue)
	put := etcd.PutReq{
		Key:     "/services/" + m.env + "/" + m.ns + "/telegram" + "/" + strconv.FormatUint(userJid, 10),
		Value:   string(bvalue),
		Options: []clientv3.OpOption{clientv3.WithLease(leaseGrant.ID)},
		ResChan: make(chan etcd.PutRes),
	}
	m.putChan <- put

	keepRespChan, err := lease.KeepAlive(m.context, leaseGrant.ID)
	if err != nil {
		level.Error(m.logger).Log("msg", "lease keepAlive err", "err", err)
	}

	go func() {
		for {
			select {
			case resp := <-keepRespChan:
				if resp == nil {
					level.Info(m.logger).Log("msg", "lease keepAlive fail")
					return
				}
			}
		}
	}()
}

func (m *Manager) OnlineAccount(phoneNum string) {
	getOnline := etcd.GetReq{
		Key:     "/service/" + m.env + "/" + m.ns + "/online/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	m.getChan <- getOnline
	onlineResult := <-getOnline.ResChan
	if onlineResult.Result.Kvs != nil {
		for _, onlineValue := range onlineResult.Result.Kvs {
			curIp := strings.Split(string(onlineValue.Key), "/")[5]
			if curIp != m.ListenAddress {
				delReq := etcd.DeleteReq{
					Key:     "/service/" + m.env + "/" + m.ns + "/online/" + curIp + "/" + phoneNum,
					Options: nil,
					ResChan: make(chan etcd.DeleteRes),
				}
				m.deleteChan <- delReq
			}
		}
	}

	putReq := etcd.PutReq{
		Key:     "/service/" + m.env + "/" + m.ns + "/online/" + m.ListenAddress + "/" + phoneNum,
		Value:   "",
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	m.putChan <- putReq
}

func (m *Manager) ServiceWithUserJid(phoneNum string) {
	putReq := etcd.PutReq{
		Key:     "/service/" + m.ListenAddress + "/" + phoneNum,
		Value:   "",
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	m.putChan <- putReq
}

func (m *Manager) UserJidProxyIp(phoneNum, proxyUrl string) {
	putReq := etcd.PutReq{
		Key:     "/service/proxyaddr/" + phoneNum,
		Value:   proxyUrl,
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	m.putChan <- putReq
}
