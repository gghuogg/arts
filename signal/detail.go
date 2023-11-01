package signal

import (
	"arthas/etcd"
	"arthas/protobuf"
	"fmt"
	"github.com/go-kit/log/level"
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/util/bytehelper"
	"google.golang.org/protobuf/proto"
)

type AccountDetail struct {
	UserJid          uint64
	PrivateKey       []byte
	PublicKey        []byte
	PublicMsgKey     []byte
	PrivateMsgKey    []byte
	Identify         []byte
	ResumptionSecret []byte
	ClientPayload    []byte
}

type LoginDetail struct {
	UserJid  uint64
	ProxyUrl string
}

type LogoutDetail struct {
	UserJid  uint64
	ProxyUrl string
	Stream   protobuf.XmppStream_ConnectClient
}

func (m *Manager) syncAccountDetail(detail AccountDetail) {

	private := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(detail.PrivateMsgKey))
	public := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(detail.PublicMsgKey))
	keyPair := ecc.NewECKeyPair(public, private)
	publicKey1 := identity.NewKey(keyPair.PublicKey())

	identityKeyPair := identity.NewKeyPair(publicKey1, keyPair.PrivateKey())

	a := &Account{
		Serializer:        serialize.NewProtoBufSerializer(),
		RegistrationID:    RandomInt64(),
		PublicKey:         detail.PublicKey,
		PrivateKey:        detail.PrivateKey,
		IdentityKeyPair:   identityKeyPair,
		ResumptionSecret:  detail.ResumptionSecret,
		PrekeyBundles:     make(map[uint64]*PreKeyBundle),
		PrekeyChan:        make(map[uint64]chan struct{}),
		Cipher:            make(map[uint64]*session.Cipher),
		SessionStore:      make(map[uint64]*session.Builder),
		RemoteAddresses:   make(map[uint64]*protocol.SignalAddress),
		MediaConnInfoChan: make(map[uint64]chan struct{}),
		MediaConnInfo:     make(map[uint64]*MediaConnInfo),
	}

	a.SignalprotocolStore = NewSignalProtocolStore(a.Serializer, a.IdentityKeyPair, uint32(a.RegistrationID))

	m.accountsSync.Store(detail.UserJid, a)

	_ = m.insertDetail(detail)

}

// insertDetail 插入ETCD，key: /arthas/accountdetail/$userjid, value: protobuf序列化AccountDetail
func (m *Manager) insertDetail(detail AccountDetail) (err error) {
	protocDetail := protobuf.AccountDetail{
		UserJid:          detail.UserJid,
		PrivateKey:       detail.PrivateKey,
		PublicKey:        detail.PublicKey,
		PublicMsgKey:     detail.PublicMsgKey,
		PrivateMsgKey:    detail.PrivateMsgKey,
		Identify:         detail.Identify,
		ResumptionSecret: detail.ResumptionSecret,
		ClientPayload:    detail.ClientPayload,
	}
	protoData, err := proto.Marshal(&protocDetail)
	if err != nil {
		level.Error(m.logger).Log("msg", "Failed insert  to etcd.", "err", err)
		return
	}
	key := fmt.Sprintf("/arthas/accountdetail/%d", detail.UserJid)
	reqModel := etcd.TxnInsertReq{
		Key:     key,
		Value:   string(protoData),
		Options: nil,
	}
	m.txnInsertChan <- reqModel
	return
}
