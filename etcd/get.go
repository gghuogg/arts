package etcd

import clientv3 "go.etcd.io/etcd/client/v3"

type GetReq struct {
	Key     string
	Options []clientv3.OpOption
	ResChan chan GetRes
}

type GetRes struct {
	Err    error
	Result *clientv3.GetResponse
}

func (m *Manager) GetChan() chan GetReq {
	return m.getChan
}

func (m *Manager) get(req GetReq) {
	res := GetRes{}
	res.Result, res.Err = m.client.Get(m.context, req.Key, req.Options...)
	req.ResChan <- res
}
