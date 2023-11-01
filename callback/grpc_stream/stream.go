package grpc_stream

import (
	"arthas/callback"
	"github.com/go-kit/log/level"
)

var Type = "stream"

type (
	GManager struct {
		*callback.IManager
	}
)

func (m *GManager) ApplyConfig() (err error) {

	m.SetSendFunc(m.sendMessage)
	return nil
}

func (m *GManager) GetManager() *callback.IManager {
	return m.IManager
}

func (m *GManager) Run() error {
	level.Info(m.GetLogger()).Log("msg", "Running to callback stream manager.")
	return m.GetManager().Run()
}

func (m *GManager) sendMessage(topic string, message interface{}) (partition int32, offset int64, err error) {
	level.Info(m.GetLogger()).Log("topic", topic, "message", message)
	return partition, offset, nil
}
