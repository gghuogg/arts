package etcd

import (
	"fmt"
	"testing"
)

func TestTxnInsert(t *testing.T) {
	initEtcd()
	putChan := manager.TxnInsertChan()
	putReq := TxnInsertReq{
		Key:     "ping",
		Value:   "pong",
		Options: nil,
		ResChan: make(chan TxnInsertRes)}
	putChan <- putReq
	res := <-putReq.ResChan
	panicErr(res.Err)
	fmt.Println(res.Result)
}
