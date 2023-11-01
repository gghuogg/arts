package etcd

//租约
import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

type LeaseReq struct {
	ResChan chan LeaseRes
}

type LeaseRes struct {
	Err    error
	Result clientv3.Lease
}

func (m *Manager) LeaseChan() chan LeaseReq {
	return m.leaseChan
}

func (m *Manager) lease(req LeaseReq) {
	res := LeaseRes{}
	res.Result = clientv3.NewLease(m.client)
	req.ResChan <- res
}
