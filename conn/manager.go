package conn

import (
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type Options struct {
}

type Manager struct {
	logger  log.Logger
	context context.Context
}

func New(logger log.Logger, ctx context.Context, o *Options) *Manager {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	m := &Manager{
		logger:  logger,
		context: ctx,
	}
	return m
}

func (m *Manager) Run() error {
	level.Info(m.logger).Log("msg", "Running to conn manager.")
	for {
		select {
		case <-m.context.Done():
			return nil
		}
	}
}

func (m *Manager) ApplyConfig() error {
	return nil
}
