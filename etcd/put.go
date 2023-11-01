package etcd

import clientv3 "go.etcd.io/etcd/client/v3"

type PutReq struct {
	Key     string
	Value   string
	Options []clientv3.OpOption
	ResChan chan PutRes
}

type PutRes struct {
	Err    error
	Result *clientv3.PutResponse
}

func (m *Manager) PutChan() chan PutReq {
	return m.putChan
}

func (m *Manager) put(req PutReq) {
	res := PutRes{}
	res.Result, res.Err = m.client.Put(m.context, req.Key, req.Value, req.Options...)
	req.ResChan <- res
}
