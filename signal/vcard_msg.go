package signal

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"fmt"
	"github.com/go-kit/log/level"
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/keys/prekey"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/state/record"
	"go.mau.fi/libsignal/util/bytehelper"
	"go.mau.fi/libsignal/util/optional"
	"google.golang.org/protobuf/proto"
	"strconv"
	"time"
)

type SendVcardMessageDetail struct {
	Sender       uint64
	Receiver     uint64
	VCardDetails []*VCardDetail
}

type VCardDetail struct {
	Fn  string
	Tel string
}

const VCARD_VERSION = "3.0"

func (a *Account) EzvcardWrite(card *VCardDetail) string {
	var msgVCard string

	msgVCard = fmt.Sprintf("BEGIN:VCARD\nVERSION:%s\nN:;%v;;;\nFN:%v\nTEL;type=CELL;waid=%v:+%v\nEND:VCARD",
		VCARD_VERSION, card.Fn, card.Fn, card.Tel, card.Tel)

	fmt.Println(msgVCard)
	return msgVCard
}

func (a *Account) SendVCardMessage(username uint64, detail SendVcardMessageDetail) {
	a.signal.mutex.Lock()

	defer a.signal.mutex.Unlock()

	receiver := detail.Receiver

	recipientJid := fmt.Sprintf("%d@s.whatsapp.net", receiver)

	accountPhone := username

	if a.RemoteAddresses[receiver] == nil {
		a.RemoteAddresses[receiver] = protocol.NewSignalAddress(recipientJid, 0)
	}
	if a.SessionStore[receiver] == nil {
		a.SessionStore[receiver] = session.NewBuilderFromSignal(a.SignalprotocolStore, a.RemoteAddresses[receiver], a.Serializer)
	}

	current := make([]byte, 0)

	getReq := etcd.GetReq{
		Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(receiver, 10),
		Options: nil,
		ResChan: make(chan etcd.GetRes),
	}
	a.signal.getChan <- getReq
	res := <-getReq.ResChan

	if res.Result.Kvs == nil {
		fmt.Println("初始化")
		//初始化
		requestId := fmt.Sprintf("%s-%d", a.timestamp, a.id)
		a.id++
		a.getPrekeyBundleWithPhoneNumberJid(requestId, recipientJid)

		a.PrekeyChan[receiver] = make(chan struct{})
		select {
		case <-a.PrekeyChan[receiver]:
			fmt.Println("接收prekey")

			preKeyID := optional.NewOptionalUint32(a.PrekeyBundles[receiver].OneTimePreKeyId)
			preKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[receiver].OneTimePreKey))

			signedPreKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[receiver].SignedPreKey))
			identitykeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[receiver].IdentityKey))
			keys := identity.NewKey(identitykeyPublic)

			bundle := prekey.NewBundle(a.PrekeyBundles[receiver].RegistrationId, 0, preKeyID,
				a.PrekeyBundles[receiver].SignedPreKeyId, preKeyPublic, signedPreKeyPublic, *a.PrekeyBundles[receiver].SignedPreKeySignature, keys)

			a.SessionStore[receiver].ProcessBundle(bundle)

		case <-time.After(time.Second * 90):
			level.Info(a.logger).Log("msg", "Time up.exit", "user", a.ClientPayLoad.Username, "contact", receiver)
			return
		}

		loadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[receiver])

		insert := etcd.TxnInsertReq{
			Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(receiver, 10),
			Value:   string(loadedSession.Serialize()),
			Options: nil,
			ResChan: make(chan etcd.TxnInsertRes),
		}
		a.signal.txnInsertChan <- insert

		current = loadedSession.Serialize()

	} else {
		fmt.Println(receiver)
		fmt.Println("读session数据")
		//从etcd获取session
		current = res.Result.Kvs[0].Value

	}

	//反序列化session
	deserializedSession, _ := record.NewSessionFromBytes(current,
		a.SignalprotocolStore.SessionStore.serializer.Session,
		a.SignalprotocolStore.SessionStore.serializer.State)

	//加载session相关
	deserializedSession.SessionState().SetLocalRegistrationID(uint32(a.RegistrationID))
	a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[receiver], deserializedSession)
	a.SignalprotocolStore.IdentityKeyStore.SaveIdentity(a.RemoteAddresses[receiver], deserializedSession.SessionState().RemoteIdentityKey())

	cipher := session.NewCipherFromSession(a.RemoteAddresses[receiver], a.SignalprotocolStore.SessionStore,
		a.SignalprotocolStore.PreKeyStore, a.SignalprotocolStore.IdentityKeyStore, a.Serializer.PreKeySignalMessage, a.Serializer.SignalMessage)

	textMsgCallbacks := make([]callback.TextMsgCallback, 0)

	for _, card := range detail.VCardDetails {

		respMsg := a.EzvcardWrite(card)

		timeStamp := proto.Uint64(uint64(time.Now().Unix()))
		version := proto.Int32(2)

		msg := &protobuf.Message{
			ContactMessage: &protobuf.Message_ContactMessage{
				DisplayName: &card.Fn,
				Vcard:       &respMsg,
			},
			MessageContextInfo: &protobuf.MessageContextInfo{
				DeviceListMetadata: &protobuf.DeviceListMetadata{
					SenderTimestamp: timeStamp,
				},
				DeviceListMetadataVersion: version,
			},
		}
		msg.MessageContextInfo.DeviceListMetadata.SenderTimestamp = timeStamp

		msgbytes, _ := proto.Marshal(msg)
		ciphertext, _ := cipher.Encrypt(bytesWithPadding(msgbytes))
		reqId := getRequestId()

		composing := createComposing(recipientJid)
		paused := createPaused(recipientJid)

		request := a.signal.SendVcardMsgRequst(recipientJid, "vcard", ciphertext, reqId)
		//jsonData, _ := json.Marshal(request)
		//fmt.Println("json:" + string(jsonData))
		//fmt.Println(request)
		a.send(composing)
		a.send(paused)
		a.send(request)

		//发送消息回调
		msgCallback := callback.TextMsgCallback{
			Sender:    a.UserJid,
			Receiver:  username,
			SendText:  respMsg,
			MsgType:   2,
			SendTime:  time.Now(),
			ReqId:     request.GetAttributeByKey(SIGNALID).GetValue().GetString_(),
			Read:      1,
			Initiator: a.UserJid,
		}
		textMsgCallbacks = append(textMsgCallbacks, msgCallback)

		if a.Cipher[receiver] == nil {
			a.Cipher[receiver] = cipher
		}
	}

	if a.Cipher[receiver] == nil {
		a.Cipher[receiver] = cipher
	}
	//发送消息后再更新一次session
	lastLoadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[receiver])

	//a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[contact], lastLoadedSession)

	put := etcd.TxnUpdateReq{
		Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(receiver, 10),
		Value:   string(lastLoadedSession.Serialize()),
		Options: nil,
		ResChan: make(chan etcd.TxnUpdateRes),
	}
	a.signal.txnUpdateChan <- put
	a.sendTextMsgCallback(textMsgCallbacks)
}

func (a *Account) vcardCallback(sender uint64, revice uint64, reqId string, respMsg string) {
	vcardCallbacks := make([]callback.VcardMsgCallback, 0)
	vcardCallbacks = append(vcardCallbacks, callback.VcardMsgCallback{
		Sender:   sender,            //发送人
		Receiver: revice,            //接收人
		SendMsg:  respMsg,           //名片内容
		SendTime: time.Now().Unix(), //发送时间
		MsgId:    respMsg,           //消息id
	})
	a.callbackChan.VcardMsgCallbackChan <- vcardCallbacks
}
