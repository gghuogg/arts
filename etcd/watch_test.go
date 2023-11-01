package etcd

import (
	"fmt"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	initEtcd()
	time.AfterFunc(30*time.Second, func() {
		cancelFunc()
	})
	watchKVChan := manager.WatchChan()
	reqModel := WatchReq{
		Key:     "ping",
		ResChan: make(chan WatchRes),
	}
	watchKVChan <- reqModel
	for res := range reqModel.ResChan {
		for _, ev := range res.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}
