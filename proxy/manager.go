package proxy

import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/net/proxy"
	"net"
	"net/url"
	"time"
)

type Options struct {
	Deadline time.Duration
}

type Manager struct {
	logger          log.Logger
	context         context.Context
	deadline        time.Duration
	proxyDetailChan chan map[uint64]*Raw
	proxies         map[uint64]*Proxy
}

type Raw struct {
	Url   string
	Error chan error
}

func New(logger log.Logger, ctx context.Context, o *Options) *Manager {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	m := &Manager{
		logger:          logger,
		context:         ctx,
		deadline:        o.Deadline,
		proxies:         make(map[uint64]*Proxy),
		proxyDetailChan: make(chan map[uint64]*Raw),
	}
	return m
}

func (m *Manager) ProxyDetailChan() chan map[uint64]*Raw {
	return m.proxyDetailChan
}

func (m *Manager) Run() error {
	level.Info(m.logger).Log("msg", "Running to proxy manager.")
	for {
		select {
		case rawProxy := <-m.proxyDetailChan:
			m.parseProxy(rawProxy)
		case <-m.context.Done():
			return nil
		}
	}
}

func (m *Manager) parseProxy(rawProxy map[uint64]*Raw) {
	for user, raw := range rawProxy {
		u, err := url.Parse(raw.Url)
		if err != nil {
			raw.Error <- err
		}
		p := &Proxy{}
		var dialer proxy.Dialer = new(net.Dialer)
		p.dialer, err = proxy.FromURL(u, dialer)
		if err != nil {
			raw.Error <- err
		}
		m.proxies[user] = p
		if d, ok := p.dialer.(proxy.ContextDialer); ok {
			rawConn, err := d.DialContext(context.Background(), "tcp", "g.whatsapp.net:5222")
			fmt.Println(rawConn)
			fmt.Println(err)
		} else {
		}
		close(raw.Error)
	}
}

func (m *Manager) ApplyConfig() error {
	return nil
}
