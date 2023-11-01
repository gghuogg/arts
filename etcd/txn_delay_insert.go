package etcd

import (
	"fmt"
	"github.com/go-kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
)

type TxnDelayInsertReq struct {
	Key     string
	Value   string
	Options []clientv3.OpOption
}

func (m *Manager) TxnDelayInsertChan() chan TxnDelayInsertReq {
	return m.txnDelayInsertChan
}

func (m *Manager) updateDelayInsertKvs(req TxnDelayInsertReq) {
	m.DelayInsertKvs.Store(req.Key, req.Value)
}

func (m *Manager) txnDelayInsert() {
	var insertList = make([]clientv3.Op, 0)
	m.DelayInsertKvs.Range(func(key, value any) bool {
		put := clientv3.OpPut(fmt.Sprintf("%v", key), fmt.Sprintf("%v", value))
		insertList = append(insertList, put)
		return true
	})
	m.DelayInsertKvs = sync.Map{}
	if len(insertList) == 0 {
		return
	}
	_, err := m.client.Txn(m.context).
		If().
		Then(insertList...).
		Commit()
	if err != nil {
		level.Error(m.logger).Log("msg", "Failed put to etcd.", "err", err)
	}
}
