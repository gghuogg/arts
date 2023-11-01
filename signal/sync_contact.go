package signal

func (m *Manager) syncContact(contact SyncContact) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	value, ok := m.accountsSync.Load(contact.Account)
	if ok {
		a := value.(*Account)
		a.sendSyncContact(contact.Account, contact.Contacts)
	}
}
