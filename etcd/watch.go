package etcd

import clientv3 "go.etcd.io/etcd/client/v3"

type WatchReq struct {
	Key     string
	Options []clientv3.OpOption
	ResChan chan WatchRes
}

type WatchRes clientv3.WatchResponse

func (m *Manager) WatchChan() chan WatchReq {
	return m.watchChan
}

func (m *Manager) watch(watchReq WatchReq) {
	watchChan := clientv3.NewWatcher(m.client).Watch(m.context, watchReq.Key, watchReq.Options...)
	for res := range watchChan {
		watchReq.ResChan <- WatchRes(res)
	}
}
