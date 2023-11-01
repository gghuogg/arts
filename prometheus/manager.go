package prometheus

import (
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type (
	Manager struct {
		addr    string
		logger  log.Logger
		context context.Context
		cfg     api.Config
		client  api.Client
		v1api   v1.API
	}
	Options struct {
		Address        string
		addr           string
		Port           string
		Username       string
		Password       string
		ListenAndServe string
	}
)

func New(logger log.Logger, ctx context.Context, o *Options) *Manager {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	cfg := api.Config{
		Address: o.Address,
	}
	m := &Manager{
		context: ctx,
		logger:  logger,
		cfg:     cfg,
	}
	return m
}

func (m *Manager) ApplyConfig() (err error) {
	client, err := api.NewClient(m.cfg)
	m.client = client
	m.v1api = v1.NewAPI(client)
	result, _ := m.v1api.Targets(context.Background())

	if err != nil {
		return err
	}

	fmt.Printf("Result: %v\n", result)
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)
	return nil

}

func (m *Manager) Run() error {
	level.Info(m.logger).Log("msg", "Running to prometheus manager.")
	for {
		select {
		case <-m.context.Done():
			return nil
		}
	}
}
