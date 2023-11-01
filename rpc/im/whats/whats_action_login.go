package whats

import (
	"arthas/callback"
	"arthas/consts"
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/signal"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/util/bytehelper"
	"google.golang.org/protobuf/proto"
	"strconv"
)

func init() {
	im, err := rpc.GetIM(consts.IMWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_LOGIN, &whatsLoginAction{})
}

type whatsLoginAction struct{}

func (t *whatsLoginAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetOrdinaryAction().GetLoginDetail()
	var success int
	for userJid, detail := range details {
		get := etcd.GetReq{
			Key:     "/arthas/accountdetail/" + strconv.FormatUint(userJid, 10),
			Options: nil,
			ResChan: make(chan etcd.GetRes),
		}
		h.EtcdGetChan() <- get
		getResult := <-get.ResChan
		if getResult.Result.Kvs != nil {
			account := &signal.Account{}
			accountDetail := &protobuf.AccountDetail{}
			clientPayload := &protobuf.ClientPayload{}
			_ = proto.Unmarshal(getResult.Result.Kvs[0].Value, accountDetail)
			_ = proto.Unmarshal(accountDetail.ClientPayload, clientPayload)

			private := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(accountDetail.PrivateMsgKey))
			public := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(accountDetail.PublicMsgKey))
			keyPair := ecc.NewECKeyPair(public, private)
			publicKey1 := identity.NewKey(keyPair.PublicKey())

			identityKeyPair := identity.NewKeyPair(publicKey1, keyPair.PrivateKey())

			account = &signal.Account{
				Serializer:        serialize.NewProtoBufSerializer(),
				RegistrationID:    signal.RandomInt64(),
				PublicKey:         accountDetail.PublicKey,
				PrivateKey:        accountDetail.PrivateKey,
				IdentityKeyPair:   identityKeyPair,
				ClientPayLoad:     clientPayload,
				ResumptionSecret:  accountDetail.ResumptionSecret,
				PrekeyBundles:     make(map[uint64]*signal.PreKeyBundle),
				PrekeyChan:        make(map[uint64]chan struct{}),
				Cipher:            make(map[uint64]*session.Cipher),
				SessionStore:      make(map[uint64]*session.Builder),
				RemoteAddresses:   make(map[uint64]*protocol.SignalAddress),
				MediaConnInfoChan: make(map[uint64]chan struct{}),
				MediaConnInfo:     make(map[uint64]*signal.MediaConnInfo),
			}
			account.SignalprotocolStore = signal.NewSignalProtocolStore(account.Serializer, account.IdentityKeyPair, uint32(account.RegistrationID))
			h.GetAccountsSync().Store(userJid, account)
		}

		value, ok := h.GetAccountsSync().Load(userJid)
		if ok {
			account := value.(*signal.Account)
			if _, blocked := h.GetBlockedAccountSet()[userJid]; blocked {
				_ = level.Info(l).Log("msg", "Account is blocked, will not continue", "user", userJid)
				continue
			}
			if account.IsLogin {
				_ = level.Info(l).Log("msg", "Account is online,will be not continue login", "userJid", userJid)
				continue
			} else {
				success++
				level.Debug(l).Log("msg", "Account exist,continue login", "userJid", userJid)
				d := signal.LoginDetail{
					UserJid:  userJid,
					ProxyUrl: detail.ProxyUrl,
				}
				h.GetLoginDetailChan() <- d
			}
		} else {
			level.Debug(l).Log("msg", "Account not exist,will be not continue login", "userJid", userJid)
			loginCallbacks := make([]callback.LoginCallback, 0)
			item := callback.LoginCallback{
				UserJid:     userJid,
				ProxyUrl:    detail.ProxyUrl,
				LoginStatus: protobuf.AccountStatus_NOT_EXIST,
				Comment:     "Account not exist",
			}
			loginCallbacks = append(loginCallbacks, item)
			h.GetCallbackChan().LoginCallbackChan <- loginCallbacks
		}
	}

	if success == len(details) {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	} else if success > 0 {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_PARTIAL_SUCCESS}
	} else {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL}
	}
	return
}
