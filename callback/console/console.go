package console

import (
	"arthas/callback"
	"github.com/go-kit/log/level"
)

var Type = "console"

type (
	CManager struct {
		*callback.IManager
	}
)

func (m *CManager) ApplyConfig() (err error) {
	m.SetSendFunc(m.sendMessage)
	return nil
}

func (m *CManager) GetManager() *callback.IManager {
	return m.IManager
}

func (m *CManager) Run() error {
	level.Info(m.GetLogger()).Log("msg", "Running to callback console manager.")
	return m.GetManager().Run()
}

func (m *CManager) sendMessage(topic string, message interface{}) (partition int32, offset int64, err error) {
	level.Info(m.GetLogger()).Log("topic", topic, "message", message)
	return partition, offset, nil
}
