package etcd

import (
	"go.etcd.io/etcd/client/v3/concurrency"
)

type SessionReq struct {
	ResChan chan SessionRes
}

type SessionRes struct {
	Err    error
	Result *concurrency.Session
}

func (m *Manager) SessionChan() chan SessionReq {
	return m.sessionChan
}

func (m *Manager) session(req SessionReq) {
	res := SessionRes{}

	res.Result, res.Err = concurrency.NewSession(m.client)
	req.ResChan <- res
}
