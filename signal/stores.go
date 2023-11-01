package signal

import (
	groupRecord "go.mau.fi/libsignal/groups/state/record"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/state/record"
)

// SessionStore
func NewSession(serializer *serialize.Serializer) *Session {
	return &Session{
		sessions:   make(map[*protocol.SignalAddress]*record.Session),
		serializer: serializer,
	}
}

type Session struct {
	sessions   map[*protocol.SignalAddress]*record.Session
	serializer *serialize.Serializer
}

func (s *Session) LoadSession(address *protocol.SignalAddress) *record.Session {
	if s.ContainsSession(address) {
		return s.sessions[address]
	}
	sessionRecord := record.NewSession(s.serializer.Session, s.serializer.State)
	s.sessions[address] = sessionRecord

	return sessionRecord
}

func (s *Session) GetSubDeviceSessions(name string) []uint32 {
	var deviceIDs []uint32

	for key := range s.sessions {
		if key.Name() == name && key.DeviceID() != 1 {
			deviceIDs = append(deviceIDs, key.DeviceID())
		}
	}

	return deviceIDs
}

func (s *Session) StoreSession(remoteAddress *protocol.SignalAddress, record *record.Session) {
	s.sessions[remoteAddress] = record
}

func (s *Session) ContainsSession(remoteAddress *protocol.SignalAddress) bool {
	_, ok := s.sessions[remoteAddress]
	return ok
}

func (s *Session) DeleteSession(remoteAddress *protocol.SignalAddress) {
	delete(s.sessions, remoteAddress)
}

func (s *Session) DeleteAllSessions() {
	s.sessions = make(map[*protocol.SignalAddress]*record.Session)
}

// IdentityKeyStore
func NewIdentityKey(identityKey *identity.KeyPair, localRegistrationID uint32) *IdentityKey {
	return &IdentityKey{
		trustedKeys:         make(map[*protocol.SignalAddress]*identity.Key),
		identityKeyPair:     identityKey,
		localRegistrationID: localRegistrationID,
	}
}

type IdentityKey struct {
	trustedKeys         map[*protocol.SignalAddress]*identity.Key
	identityKeyPair     *identity.KeyPair
	localRegistrationID uint32
}

func (i *IdentityKey) GetIdentityKeyPair() *identity.KeyPair {
	return i.identityKeyPair
}

func (i *IdentityKey) GetLocalRegistrationId() uint32 {
	return i.localRegistrationID
}

func (i *IdentityKey) SaveIdentity(address *protocol.SignalAddress, identityKey *identity.Key) {
	i.trustedKeys[address] = identityKey
}

func (i *IdentityKey) IsTrustedIdentity(address *protocol.SignalAddress, identityKey *identity.Key) bool {
	trusted := i.trustedKeys[address]
	return trusted == nil || trusted.Fingerprint() == identityKey.Fingerprint()
}

// PreKeyStore
func NewPreKey() *PreKey {
	return &PreKey{
		store: make(map[uint32]*record.PreKey),
	}
}

type PreKey struct {
	store map[uint32]*record.PreKey
}

func (p *PreKey) LoadPreKey(preKeyID uint32) *record.PreKey {
	return p.store[preKeyID]
}

func (p *PreKey) StorePreKey(preKeyID uint32, preKeyRecord *record.PreKey) {
	p.store[preKeyID] = preKeyRecord
}

func (p *PreKey) ContainsPreKey(preKeyID uint32) bool {
	_, ok := p.store[preKeyID]
	return ok
}

func (p *PreKey) RemovePreKey(preKeyID uint32) {
	delete(p.store, preKeyID)
}

// SignedPreKeyStore
func NewSignedPreKey() *SignedPreKey {
	return &SignedPreKey{
		store: make(map[uint32]*record.SignedPreKey),
	}
}

type SignedPreKey struct {
	store map[uint32]*record.SignedPreKey
}

func (s *SignedPreKey) LoadSignedPreKey(signedPreKeyID uint32) *record.SignedPreKey {
	return s.store[signedPreKeyID]
}

func (s *SignedPreKey) LoadSignedPreKeys() []*record.SignedPreKey {
	var preKeys []*record.SignedPreKey

	for _, record := range s.store {
		preKeys = append(preKeys, record)
	}

	return preKeys
}

func (s *SignedPreKey) StoreSignedPreKey(signedPreKeyID uint32, record *record.SignedPreKey) {
	s.store[signedPreKeyID] = record
}

func (s *SignedPreKey) ContainsSignedPreKey(signedPreKeyID uint32) bool {
	_, ok := s.store[signedPreKeyID]
	return ok
}

func (s *SignedPreKey) RemoveSignedPreKey(signedPreKeyID uint32) {
	delete(s.store, signedPreKeyID)
}

// senderKeyStore
func NewSenderKey() *SenderKey {
	return &SenderKey{
		store: make(map[*protocol.SenderKeyName]*groupRecord.SenderKey),
	}
}

type SenderKey struct {
	store map[*protocol.SenderKeyName]*groupRecord.SenderKey
}

func (s *SenderKey) StoreSenderKey(senderKeyName *protocol.SenderKeyName, keyRecord *groupRecord.SenderKey) {
	s.store[senderKeyName] = keyRecord
}

func (s *SenderKey) LoadSenderKey(senderKeyName *protocol.SenderKeyName) *groupRecord.SenderKey {
	return s.store[senderKeyName]
}
