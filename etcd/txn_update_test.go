package etcd

import (
	"fmt"
	"testing"
)

func TestTxnUpdate(t *testing.T) {
	initEtcd()
	reqChan := manager.TxnUpdateChan()
	req := TxnUpdateReq{
		Key:     "ping",
		Value:   "pong",
		Options: nil,
		ResChan: make(chan TxnUpdateRes)}
	reqChan <- req
	res := <-req.ResChan
	panicErr(res.Err)
	fmt.Println(res.Result)
}
