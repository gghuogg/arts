package signal

import (
	"arthas/callback"
	"github.com/go-kit/log/level"
)

const (
	ACTIVELY_LOGOUT = 1
	ERROR_LOGOUT    = 2
)

// 退出其它已登录设备
func (m *Manager) logoutCompanionDevice(detail LogoutDetail) {

	value, ok := m.accountsSync.Load(detail.UserJid)
	if !ok {
		level.Info(m.logger).Log("msg", "Account is not exist", "user", detail.UserJid)
		return
	}
	account := value.(*Account)
	//fmt.Println("account:", account)
	if account.IsLogin == false {
		level.Info(m.logger).Log("msg", "Account is offline,will not continue", "user", detail.UserJid)
		return
	}
	account.IsLogin = false

	//account.etcdAccountLoginState()

	err := account.Stream.CloseSend()
	if err != nil {
		level.Info(m.logger).Log("msg", "Account stream is  close", "user", detail.UserJid)
	}
	m.recordConnection(false)
	m.logoutCallback(detail.UserJid, detail.ProxyUrl, ACTIVELY_LOGOUT)

}

func (m *Manager) logoutCallback(userJid uint64, proxy string, logoutType int) {
	logoutCallbacks := make([]callback.LogoutCallback, 0)
	item := callback.LogoutCallback{
		UserJid:    userJid,
		Proxy:      proxy,
		LogoutType: logoutType,
	}
	logoutCallbacks = append(logoutCallbacks, item)

	m.callbackChan.LogoutCallbackChan <- logoutCallbacks
}
