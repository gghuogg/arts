package signal

import "sync"

func (m *Manager) AccountDetailChan() chan AccountDetail {
	return m.accountDetailChan
}

func (m *Manager) LoginDetailChan() chan LoginDetail {
	return m.loginDetailChan
}

func (m *Manager) SyncContactChan() chan SyncContact {
	return m.syncContactChan
}

func (m *Manager) SendMessageChan() chan SendMessageDetail {
	return m.sendMessageChan
}

func (m *Manager) Accounts() map[uint64]*Account {
	return m.accounts
}

func (m *Manager) AccountsSync() *sync.Map {
	return m.accountsSync
}

func (m *Manager) BlockAccountSet() map[uint64]struct{} {
	return m.blockedAccountSet
}

func (m *Manager) SendVCardMessageChan() chan SendVcardMessageDetail {
	return m.sendVcardMessageChan
}

func (m *Manager) LogoutDetailChan() chan LogoutDetail {
	return m.logoutDetailChan
}

func (m *Manager) SetUserHeadImageDetailChan() chan GetUserHeadImageDetail {
	return m.GetUserImageDetailChan
}

func (m *Manager) SendImageFileChan() chan SendImageFileDetail {
	return m.SendImageFileDetailChan
}

func (m *Manager) SendFileChan() chan SendFileDetail {
	return m.SendFileDetailChan
}

func (m *Manager) SendVideoChan() chan SendVideoFileDetail {
	return m.SendVideoDetailChan
}
