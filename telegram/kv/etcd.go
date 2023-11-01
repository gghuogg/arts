package kv

import (
	"arthas/etcd"
	"context"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("key not found")

type KV interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	WithNs(ns string)
}

type Etcd struct {
	ctx     context.Context
	ns      string
	getChan chan etcd.GetReq
	putChan chan etcd.PutReq
	delChan chan etcd.DeleteReq
}

type EtcdOptions struct {
	Ctx     context.Context
	NS      string `validate:"required"`
	GetChan chan etcd.GetReq
	PutChan chan etcd.PutReq
	DelChan chan etcd.DeleteReq
}

func NewEtcd(opts EtcdOptions) (KV, error) {

	return &Etcd{
		ctx:     opts.Ctx,
		ns:      opts.NS,
		getChan: opts.GetChan,
		putChan: opts.PutChan,
		delChan: opts.DelChan,
	}, nil
}

func (e *Etcd) WithNs(ns string) {
	e.ns = ns
}

func (e *Etcd) Get(key string) ([]byte, error) {
	var val []byte
	getChan := e.getChan
	getReq := etcd.GetReq{Key: e.newKey(key), ResChan: make(chan etcd.GetRes)}

	getChan <- getReq
	res := <-getReq.ResChan
	if res.Err != nil || res.Result.Count == 0 {
		return nil, ErrNotFound
	}
	val = res.Result.Kvs[0].Value
	return val, nil
}

func (e *Etcd) Set(key string, val []byte) error {
	putChan := e.putChan
	putReq := etcd.PutReq{
		Key:     e.newKey(key),
		Value:   string(val),
		Options: nil,
		ResChan: make(chan etcd.PutRes)}
	putChan <- putReq
	res := <-putReq.ResChan
	return res.Err
}

// Delete removes a key from the bucket. If the key does not exist then nothing is done and a nil error is returned
func (e *Etcd) Delete(key string) error {
	delChan := e.delChan
	delReq := etcd.DeleteReq{
		Key:     e.newKey(key),
		Options: nil,
		ResChan: make(chan etcd.DeleteRes),
	}
	delChan <- delReq
	res := <-delReq.ResChan
	return res.Err
}

func (e *Etcd) newKey(key string) string {
	return fmt.Sprintf("/tg/%s/%s/json", e.ns, key)
	//return fmt.Sprintf("/tg/%s/%s", e.ns, key)

}
