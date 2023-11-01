package etcd

import (
	"fmt"
	"testing"
)

func TestTxnDelete(t *testing.T) {
	initEtcd()
	delChan := manager.TxnDeleteChan()
	delReq := TxnDeleteReq{
		Key:     "ping",
		Options: nil,
		ResChan: make(chan TxnDeleteRes),
	}
	delChan <- delReq
	res := <-delReq.ResChan
	panicErr(res.Err)
	fmt.Println(res.Result)
}
