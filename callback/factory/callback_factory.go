package factory

import (
	"arthas/callback"
	"arthas/callback/console"
	"arthas/callback/consumer"
	"arthas/callback/grpc_stream"
	"arthas/callback/kafka"
	"context"
	"github.com/go-kit/log"
)

func New(logger log.Logger, ctx context.Context, o *callback.Options) (m callback.ICallback) {
	iManager := callback.DefaultManager(logger, ctx, o)
	switch o.Type {
	case kafka.Type:
		m = &kafka.KManager{IManager: iManager}
	case console.Type:
		m = &console.CManager{IManager: iManager}
	case grpc_stream.Type:
		m = &grpc_stream.GManager{IManager: iManager}
	case consumer.Type:
		m = &consumer.SManager{IManager: iManager}
	}
	return m
}
