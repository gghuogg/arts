package signal

import (
	groupRecord "go.mau.fi/libsignal/groups/state/record"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/state/record"
)

func NewSignalProtocolStore(serializer *serialize.Serializer, identityKey *identity.KeyPair, localRegistrationID uint32) *LSignalProtocolStore {

	return &LSignalProtocolStore{
		IdentityKeyStore:  NewIdentityKey(identityKey, localRegistrationID),
		PreKeyStore:       NewPreKey(),
		SessionStore:      NewSession(serializer),
		SignedPreKeyStore: NewSignedPreKey(),
		SenderKeyStore:    NewSenderKey(),
	}
}

type LSignalProtocolStore struct {
	IdentityKeyStore  *IdentityKey
	PreKeyStore       *PreKey
	SessionStore      *Session
	SignedPreKeyStore *SignedPreKey
	SenderKeyStore    *SenderKey
}

func (i *LSignalProtocolStore) GetIdentityKeyPair() *identity.KeyPair {
	return i.IdentityKeyStore.GetIdentityKeyPair()
}

func (i *LSignalProtocolStore) GetLocalRegistrationId() uint32 {
	return i.IdentityKeyStore.GetLocalRegistrationId()
}

func (i *LSignalProtocolStore) SaveIdentity(address *protocol.SignalAddress, identityKey *identity.Key) {
	i.IdentityKeyStore.SaveIdentity(address, identityKey)
}

func (i *LSignalProtocolStore) IsTrustedIdentity(address *protocol.SignalAddress, identityKey *identity.Key) bool {
	return i.IdentityKeyStore.IsTrustedIdentity(address, identityKey)
}

func (i *LSignalProtocolStore) LoadPreKey(preKeyID uint32) *record.PreKey {
	return i.PreKeyStore.LoadPreKey(preKeyID)
}

func (i *LSignalProtocolStore) StorePreKey(preKeyID uint32, preKeyRecord *record.PreKey) {
	i.PreKeyStore.StorePreKey(preKeyID, preKeyRecord)
}

func (i *LSignalProtocolStore) ContainsPreKey(preKeyID uint32) bool {
	return i.PreKeyStore.ContainsPreKey(preKeyID)
}

func (i *LSignalProtocolStore) RemovePreKey(preKeyID uint32) {
	i.PreKeyStore.RemovePreKey(preKeyID)
}

func (i *LSignalProtocolStore) LoadSession(address *protocol.SignalAddress) *record.Session {
	return i.SessionStore.LoadSession(address)
}

func (i *LSignalProtocolStore) GetSubDeviceSessions(name string) []uint32 {
	return i.SessionStore.GetSubDeviceSessions(name)
}

func (i *LSignalProtocolStore) StoreSession(remoteAddress *protocol.SignalAddress, record *record.Session) {
	i.SessionStore.StoreSession(remoteAddress, record)
}

func (i *LSignalProtocolStore) ContainsSession(remoteAddress *protocol.SignalAddress) bool {
	return i.SessionStore.ContainsSession(remoteAddress)
}

func (i *LSignalProtocolStore) DeleteSession(remoteAddress *protocol.SignalAddress) {
	i.SessionStore.DeleteSession(remoteAddress)
}

func (i *LSignalProtocolStore) DeleteAllSessions() {
	// i.sessions = make(map[*protocol.SignalAddress]*record.Session)
	i.SessionStore.DeleteAllSessions()
}

func (i *LSignalProtocolStore) LoadSignedPreKey(signedPreKeyID uint32) *record.SignedPreKey {
	return i.SignedPreKeyStore.LoadSignedPreKey(signedPreKeyID)
}

func (i *LSignalProtocolStore) LoadSignedPreKeys() []*record.SignedPreKey {
	return i.SignedPreKeyStore.LoadSignedPreKeys()
}

func (i *LSignalProtocolStore) StoreSignedPreKey(signedPreKeyID uint32, record *record.SignedPreKey) {
	i.SignedPreKeyStore.StoreSignedPreKey(signedPreKeyID, record)
}

func (i *LSignalProtocolStore) ContainsSignedPreKey(signedPreKeyID uint32) bool {
	return i.SignedPreKeyStore.ContainsSignedPreKey(signedPreKeyID)
}

func (i *LSignalProtocolStore) RemoveSignedPreKey(signedPreKeyID uint32) {
	i.SignedPreKeyStore.RemoveSignedPreKey(signedPreKeyID)
}

func (i *LSignalProtocolStore) StoreSenderKey(senderKeyName *protocol.SenderKeyName, keyRecord *groupRecord.SenderKey) {
	i.SenderKeyStore.StoreSenderKey(senderKeyName, keyRecord)
}

func (i *LSignalProtocolStore) LoadSenderKey(senderKeyName *protocol.SenderKeyName) *groupRecord.SenderKey {
	return i.SenderKeyStore.LoadSenderKey(senderKeyName)
}
