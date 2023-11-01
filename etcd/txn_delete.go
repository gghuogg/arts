package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"
)

type TxnDeleteReq struct {
	Key     string
	Options []clientv3.OpOption
	ResChan chan TxnDeleteRes
}

type TxnDeleteRes struct {
	Err    error
	Result *clientv3.TxnResponse
}

func (m *Manager) TxnDeleteChan() chan TxnDeleteReq {
	return m.txnDeleteChan
}

func (m *Manager) txnDelete(req TxnDeleteReq) {
	res := TxnDeleteRes{}
	res.Result, res.Err = m.client.Txn(m.context).
		If(clientv3util.KeyExists(req.Key)).
		Then(clientv3.OpDelete(req.Key, req.Options...)).
		Commit()
	req.ResChan <- res
}
