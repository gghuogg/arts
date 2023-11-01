package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"
)

type TxnUpdateReq struct {
	Key     string
	Value   string
	Options []clientv3.OpOption
	ResChan chan TxnUpdateRes
}

type TxnUpdateRes struct {
	Err    error
	Result *clientv3.TxnResponse
}

func (m *Manager) TxnUpdateChan() chan TxnUpdateReq {
	return m.txnUpdateChan
}

func (m *Manager) txnUpdate(req TxnUpdateReq) {
	res := TxnUpdateRes{}
	res.Result, res.Err = m.client.Txn(m.context).
		If(clientv3util.KeyExists(req.Key)).
		Then(clientv3.OpPut(req.Key, req.Value, req.Options...), clientv3.OpGet(req.Key)).
		Else(clientv3.OpPut(req.Key, req.Value, req.Options...), clientv3.OpGet(req.Key)).
		Commit()
	req.ResChan <- res
}
