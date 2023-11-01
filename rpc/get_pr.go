package rpc

import (
	"github.com/go-kit/log"
	"sync"
)

func (p *ProxyServer) PrGetSyncMap(key string) *sync.Map {
	return p.getSyncMap(key)
}

func (p *ProxyServer) GetLogger() log.Logger {
	return p.logger
}

func (p *ProxyServer) GetMu() *sync.Mutex {
	return &p.mu
}

func (p *ProxyServer) GetCurrentWsIPIndex() int {
	return p.currentWsIPIndex
}

func (p *ProxyServer) SetCurrentWsIPIndex(value int) {
	p.currentWsIPIndex = value
}
