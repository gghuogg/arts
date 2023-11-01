package etcd

import clientv3 "go.etcd.io/etcd/client/v3"

type DeleteReq struct {
	Key     string
	Options []clientv3.OpOption
	ResChan chan DeleteRes
}

type DeleteRes struct {
	Err    error
	Result *clientv3.DeleteResponse
}

func (m *Manager) DeleteChan() chan DeleteReq {
	return m.deleteChan
}

func (m *Manager) delete(req DeleteReq) {
	res := DeleteRes{}
	res.Result, res.Err = m.client.Delete(m.context, req.Key, req.Options...)
	req.ResChan <- res
}
