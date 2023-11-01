package etcd

import (
	"fmt"
	"testing"
)

func TestPut(t *testing.T) {
	initEtcd()
	putChan := manager.PutChan()
	putReq := PutReq{
		Key:     "ping",
		Value:   "pong",
		Options: nil,
		ResChan: make(chan PutRes)}
	putChan <- putReq
	res := <-putReq.ResChan
	panicErr(res.Err)
	fmt.Println(res.Result)
}
