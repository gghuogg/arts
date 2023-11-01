package etcd

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
)

func TestDelete(t *testing.T) {
	initEtcd()
	delChan := manager.DeleteChan()
	delReq := DeleteReq{
		Key:     "/service",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan DeleteRes),
	}
	delChan <- delReq
	res := <-delReq.ResChan
	panicErr(res.Err)
	fmt.Println(res.Result)
}
