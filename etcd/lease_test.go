package etcd

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestLease(t *testing.T) {
	initEtcd()
	leaseChan := manager.LeaseChan()
	reqModel := LeaseReq{ResChan: make(chan LeaseRes)}
	leaseChan <- reqModel
	res := <-reqModel.ResChan
	lease := res.Result
	//声明一个租约，并且设置ttl
	leaseGrant, err := lease.Grant(ctx, 10)
	panicErr(err)
	//自动续约
	keepRespChan, err := lease.KeepAlive(ctx, leaseGrant.ID)
	panicErr(err)

	putChan := manager.PutChan()
	//设置key value
	putReq := PutReq{
		Key:     "ping",
		Value:   "pong",
		Options: []clientv3.OpOption{clientv3.WithLease(leaseGrant.ID)}, //并且绑定租约
		ResChan: make(chan PutRes),
	}
	putChan <- putReq
	putRes := <-putReq.ResChan
	panicErr(putRes.Err)
	//取消续约
	//_, err = lease.Revoke(ctx, leaseGrant.ID)

	go func() {
		//查看续期情况 非必需，帮助观察续租的过程
		for {
			select {
			case resp := <-keepRespChan:
				if resp == nil {
					fmt.Println("租约失效")
					return
				} else {
					fmt.Println("租约成功", resp)
				}
			}
		}
	}()

	for { //持续检测key是否过期
		getChan := manager.GetChan()
		getReq := GetReq{
			Key:     "ping",
			Options: nil,
			ResChan: make(chan GetRes),
		}
		getChan <- getReq
		getRes := <-getReq.ResChan
		panicErr(getRes.Err)
		if getRes.Result.Count == 0 {
			fmt.Println("已经过期")
		} else {
			fmt.Println("没过期", getRes.Result.Kvs)
		}
		time.Sleep(time.Second * 1)
	}
}
