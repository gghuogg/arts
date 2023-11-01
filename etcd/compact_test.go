package etcd

import (
	"fmt"
	"testing"
)

func TestCompact(t *testing.T) {
	initEtcd()
	compactChan := manager.CompactChan()
	reqModel := CompactReq{
		Rev:     1,
		Options: nil,
		ResChan: make(chan CompactRes),
	}
	compactChan <- reqModel
	res := <-reqModel.ResChan
	panicErr(res.Err)
	fmt.Println(res)
}
