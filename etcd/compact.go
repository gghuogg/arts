package etcd

import clientv3 "go.etcd.io/etcd/client/v3"

type CompactReq struct {
	Rev     int64
	Options []clientv3.CompactOption
	ResChan chan CompactRes
}

type CompactRes struct {
	Err    error
	Result *clientv3.CompactResponse
}

func (m *Manager) CompactChan() chan CompactReq {
	return m.compactChan
}

func (m *Manager) compact(req CompactReq) {
	res := CompactRes{}
	res.Result, res.Err = m.client.Compact(m.context, req.Rev, req.Options...)
	req.ResChan <- res
}
