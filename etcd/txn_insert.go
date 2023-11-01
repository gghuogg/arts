package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"
)

type TxnInsertReq struct {
	Key     string
	Value   string
	Options []clientv3.OpOption
	ResChan chan TxnInsertRes
}

type TxnInsertRes struct {
	Err    error
	Result *clientv3.TxnResponse
}

func (m *Manager) TxnInsertChan() chan TxnInsertReq {
	return m.txnInsertChan
}

func (m *Manager) txnInsert(req TxnInsertReq) {
	res := TxnInsertRes{}
	res.Result, res.Err = m.client.Txn(m.context).
		If(clientv3util.KeyMissing(req.Key)).
		Then(clientv3.OpPut(req.Key, req.Value, req.Options...), clientv3.OpGet(req.Key)).
		Else(clientv3.OpGet(req.Key)).
		Commit()
	req.ResChan <- res
}
