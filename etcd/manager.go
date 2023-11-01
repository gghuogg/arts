package etcd

import (
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

type Options struct {
	EndPoint []string
	Url      string
	Port     string
	Username string
	Password string
	Delay    time.Duration
}

type Manager struct {
	cfg                clientv3.Config
	logger             log.Logger
	context            context.Context
	client             *clientv3.Client
	kv                 clientv3.KV
	delay              time.Duration
	getChan            chan GetReq            //获取kv，无事务
	putChan            chan PutReq            //插入kv，无事务，key存在会覆盖原有value
	deleteChan         chan DeleteReq         //删除kv，无事务
	watchChan          chan WatchReq          //监听kv变化
	compactChan        chan CompactReq        //清理历史kv
	leaseChan          chan LeaseReq          //租约
	txnInsertChan      chan TxnInsertReq      //新增kv，无则新增，Succeeded=true，有则返回的kv，Succeeded=false
	txnUpdateChan      chan TxnUpdateReq      //更新kv，有则修改 Succeeded=true，无则新增 Succeeded=false
	txnDeleteChan      chan TxnDeleteReq      //删除kv，有则删除 Succeeded=true，无则不做操作 Succeeded=false
	txnDelayInsertChan chan TxnDelayInsertReq //延时插入
	sessionChan        chan SessionReq
	DelayInsertKvs     sync.Map
}

type KeyValue struct {
	Key   string
	Value string
}

func New(logger log.Logger, ctx context.Context, o *Options) *Manager {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	cfg := clientv3.Config{
		Context:     ctx,
		Username:    o.Username,
		Password:    o.Password,
		Endpoints:   o.EndPoint,
		DialTimeout: 15 * time.Second,
	}
	m := &Manager{
		context:            ctx,
		logger:             logger,
		cfg:                cfg,
		getChan:            make(chan GetReq),
		putChan:            make(chan PutReq),
		deleteChan:         make(chan DeleteReq),
		txnInsertChan:      make(chan TxnInsertReq),
		txnUpdateChan:      make(chan TxnUpdateReq),
		txnDeleteChan:      make(chan TxnDeleteReq),
		watchChan:          make(chan WatchReq),
		compactChan:        make(chan CompactReq),
		leaseChan:          make(chan LeaseReq),
		txnDelayInsertChan: make(chan TxnDelayInsertReq),
		sessionChan:        make(chan SessionReq),
		delay:              o.Delay,
		DelayInsertKvs:     sync.Map{},
	}
	if o.Delay < 1000 {
		m.delay = 10000
	}
	return m
}

func (m *Manager) ApplyConfig() error {
	var err error
	m.client, err = clientv3.New(m.cfg)
	if err != nil {
		return err
	}
	m.kv = clientv3.NewKV(m.client)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) Run() error {
	level.Info(m.logger).Log("msg", "Running to etcd manager.")
	ticker := time.NewTicker(m.delay)
	for {
		select {
		case getChan := <-m.getChan:
			go m.get(getChan)
		case putChan := <-m.putChan:
			go m.put(putChan)
		case delChan := <-m.deleteChan:
			go m.delete(delChan)
		case txnInsertChan := <-m.txnInsertChan:
			go m.txnInsert(txnInsertChan)
		case txnUpdateChan := <-m.txnUpdateChan:
			go m.txnUpdate(txnUpdateChan)
		case txnDeleteChan := <-m.txnDeleteChan:
			go m.txnDelete(txnDeleteChan)
		case watchChan := <-m.watchChan:
			go m.watch(watchChan)
		case compactChan := <-m.compactChan:
			go m.compact(compactChan)
		case leaseChan := <-m.leaseChan:
			go m.lease(leaseChan)
		case txnDelayInsertChan := <-m.txnDelayInsertChan:
			go m.updateDelayInsertKvs(txnDelayInsertChan)
		case sessionChan := <-m.sessionChan:
			go m.session(sessionChan)
		case <-ticker.C:
			go m.txnDelayInsert()
		case <-m.context.Done():
			return nil
		}
	}
}
