package etcd

import (
	"arthas/protobuf"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"log"
	"testing"
	"time"
)

func TestDelayInsert(t *testing.T) {
	initEtcd()
	protocDetail := protobuf.AccountDetail{
		UserJid:          1,
		PrivateKey:       []byte("PrivateKey"),
		PublicKey:        []byte("PublicKey"),
		PublicMsgKey:     []byte("PublicMsgKey"),
		PrivateMsgKey:    []byte("PrivateMsgKey"),
		Identify:         []byte("Identify"),
		ResumptionSecret: []byte("ResumptionSecret"),
		ClientPayload:    []byte("ClientPayload"),
	}
	protoData, err := proto.Marshal(&protocDetail)
	if err != nil {
		panic(err)
	}
	key := fmt.Sprintf("/arthas/accountdetail/%d", protocDetail.UserJid)
	reqChan := manager.TxnDelayInsertChan()
	reqModel := TxnDelayInsertReq{
		Key:     key,
		Value:   string(protoData),
		Options: nil,
	}
	reqChan <- reqModel
	time.Sleep(11 * time.Second)
	getChan := manager.GetChan()
	getReq := GetReq{Key: key, Options: []clientv3.OpOption{clientv3.WithPrefix()}, ResChan: make(chan GetRes)}
	getChan <- getReq
	res := <-getReq.ResChan
	panicErr(res.Err)
	etcdDetail := &protobuf.AccountDetail{}
	err = proto.Unmarshal(res.Result.Kvs[0].Value, etcdDetail)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}

}
