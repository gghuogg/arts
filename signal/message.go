package signal

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/go-kit/log/level"
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/keys/prekey"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/state/record"
	"go.mau.fi/libsignal/util/bytehelper"
	"go.mau.fi/libsignal/util/optional"
	"google.golang.org/protobuf/proto"
	"strconv"
	"strings"
	"time"
)

type PreKeyBundle struct {
	RegistrationId        uint32
	OneTimePreKeyId       uint32
	DeviceId              int32
	OneTimePreKey         []byte
	SignedPreKeyId        uint32
	SignedPreKey          []byte
	SignedPreKeySignature *[64]byte
	IdentityKey           []byte
}

type MediaConnInfo struct {
	Result      string
	AuthKey     string
	LastMediaId string
	HostName    string
}

type SendMessageDetail struct {
	Sender   uint64
	Receiver uint64
	SendText []string
}

func (a *Account) SendImageMessage(contact uint64, image []string) {
	//a.signal.mutex.Lock()
	//
	//defer a.signal.mutex.Unlock()

	accountPhone := uint64FromPtr(a.ClientPayLoad.Username)
	recipientJid := fmt.Sprintf("%d@s.whatsapp.net", contact)
	if a.RemoteAddresses[contact] == nil {
		a.RemoteAddresses[contact] = protocol.NewSignalAddress(recipientJid, 0)
	}
	if a.SessionStore[contact] == nil {
		a.SessionStore[contact] = session.NewBuilderFromSignal(a.SignalprotocolStore, a.RemoteAddresses[contact], a.Serializer)
	}

	current := make([]byte, 0)

	exist := a.SignalprotocolStore.SessionStore.ContainsSession(a.RemoteAddresses[contact])
	if exist == true {
		current = a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact]).Serialize()
	} else {

		getReq := etcd.GetReq{
			Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
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

			a.PrekeyChan[contact] = make(chan struct{})
			select {
			case <-a.PrekeyChan[contact]:
				fmt.Println("接收prekey")

				preKeyID := optional.NewOptionalUint32(a.PrekeyBundles[contact].OneTimePreKeyId)
				preKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].OneTimePreKey))

				signedPreKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].SignedPreKey))
				identitykeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].IdentityKey))
				keys := identity.NewKey(identitykeyPublic)

				bundle := prekey.NewBundle(a.PrekeyBundles[contact].RegistrationId, 0, preKeyID,
					a.PrekeyBundles[contact].SignedPreKeyId, preKeyPublic, signedPreKeyPublic, *a.PrekeyBundles[contact].SignedPreKeySignature, keys)

				a.SessionStore[contact].ProcessBundle(bundle)

			case <-time.After(time.Second * 90):
				level.Info(a.logger).Log("msg", "Time up.exit", "user", a.ClientPayLoad.Username, "contact", contact)
				return
			}

			loadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact])

			insert := etcd.TxnInsertReq{
				Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
				Value:   string(loadedSession.Serialize()),
				Options: nil,
				ResChan: make(chan etcd.TxnInsertRes),
			}
			a.signal.txnInsertChan <- insert

			current = loadedSession.Serialize()

		} else {
			fmt.Println("读session数据")
			//从etcd获取session
			current = res.Result.Kvs[0].Value

		}
	}

	//反序列化session
	deserializedSession, _ := record.NewSessionFromBytes(current,
		a.SignalprotocolStore.SessionStore.serializer.Session,
		a.SignalprotocolStore.SessionStore.serializer.State)

	//加载session相关
	a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[contact], deserializedSession)
	a.SignalprotocolStore.IdentityKeyStore.SaveIdentity(a.RemoteAddresses[contact], deserializedSession.SessionState().RemoteIdentityKey())

	cipher := session.NewCipherFromSession(a.RemoteAddresses[contact], a.SignalprotocolStore.SessionStore,
		a.SignalprotocolStore.PreKeyStore, a.SignalprotocolStore.IdentityKeyStore, a.Serializer.PreKeySignalMessage, a.Serializer.SignalMessage)

	if a.Cipher[contact] == nil {
		a.Cipher[contact] = cipher
	}

	timeStamp := proto.Uint64(uint64(time.Now().Unix()))

	mediaKey, _ := hex.DecodeString("3b61de5dfe9ad2b4adc48a467c84c2e8767aac97f255e5a94915a80c9476660f")
	fileSha256, _ := hex.DecodeString("84c40e4272fae70e0244a0e27ebc791c7f5b0e89fd1ebd8b0d7ff4c365922137")
	fileEncSha256, _ := hex.DecodeString("76a4507c23ce14c032cb85a286908f69116bded49fe45b43127d7fe5fa431fb7")
	fileLength := uint64(38997)
	directPath := "/v/t62.7118-24/30927259_1125657852157689_1631449215857950242_n.enc?ccb=11-4&oh=01_AdT2tsq-7QNyALaahBKNsuO3Il5rK8h8OfF_YkWEXqbeXQ&oe=64A62411"
	url := "https://mmg.whatsapp.net/v/t62.7118-24/30927259_1125657852157689_1631449215857950242_n.enc?ccb=11-4&oh=01_AdT2tsq-7QNyALaahBKNsuO3Il5rK8h8OfF_YkWEXqbeXQ&oe=64A62411&mms3=true"

	version := proto.Int32(2)
	msg := &protobuf.Message{
		ImageMessage: &protobuf.Message_ImageMessage{
			Url:           &url,
			Mimetype:      proto.String(SIGNALMIMETYPE),
			MediaKey:      mediaKey,
			FileSha256:    fileSha256,
			FileLength:    &fileLength,
			FileEncSha256: fileEncSha256,
			DirectPath:    &directPath,
		},
		MessageContextInfo: &protobuf.MessageContextInfo{
			DeviceListMetadata: &protobuf.DeviceListMetadata{
				SenderTimestamp: timeStamp,
			},
			DeviceListMetadataVersion: version,
		},
	}
	//???
	msg.MessageContextInfo.DeviceListMetadata.SenderTimestamp = timeStamp

	msgbytes, _ := proto.Marshal(msg)
	ciphertext, _ := cipher.Encrypt(bytesWithPadding(msgbytes))
	//fmt.Printf("    Cipher:%v  ", base64.StdEncoding.EncodeToString(ciphertext.Serialize()))
	request := createMediaMessageRequest(recipientJid, SIGNALIMAGE, ciphertext)
	//fmt.Printf("   messageRequest:%v", request)
	a.send(request)

	lastLoadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact])

	put := etcd.PutReq{
		Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
		Value:   string(lastLoadedSession.Serialize()),
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	a.signal.putChan <- put
}

func (a *Account) getMediaConnRequest() {

	msg := createMediaConnRequest(0)
	a.send(msg)
}

func (a *Account) createMediaConnInfo(in *protobuf.XmppStanzaElement, mediaConnInfo *MediaConnInfo, userJid uint64) {
	if in.GetAttributeByKey(SIGNALTYPE).Value.GetString_() != "result" {
		return
	}
	mediaConnInfo.Result = in.GetChildrenByName(SIGNALMEDIACONN).GetName()
	mediaConnInfo.AuthKey = in.GetChildrenByName(SIGNALMEDIACONN).GetAttributeByKey("auth").Value.GetString_()
	mediaConnInfo.LastMediaId = in.GetChildrenByName(SIGNALMEDIACONN).GetAttributeByKey(SIGNALID).Value.GetString_()
	mediaConnInfo.HostName = in.GetChildrenByName(SIGNALMEDIACONN).GetChildrenByName("host").GetAttributeByKey("hostname").Value.GetString_()

	a.MediaConnInfo[userJid] = mediaConnInfo
	c := struct{}{}
	a.MediaConnInfoChan[userJid] <- c
}

// 发送消息
func (a *Account) SendTextMessage(contact uint64, texts []string) {
	a.signal.mutex.Lock()

	defer a.signal.mutex.Unlock()

	accountPhone := uint64FromPtr(a.ClientPayLoad.Username)
	recipientJid := fmt.Sprintf("%d@s.whatsapp.net", contact)
	if a.RemoteAddresses[contact] == nil {
		a.RemoteAddresses[contact] = protocol.NewSignalAddress(recipientJid, 0)
	}
	if a.SessionStore[contact] == nil {
		a.SessionStore[contact] = session.NewBuilderFromSignal(a.SignalprotocolStore, a.RemoteAddresses[contact], a.Serializer)
	}

	current := make([]byte, 0)

	//exist := a.SignalprotocolStore.SessionStore.ContainsSession(a.RemoteAddresses[contact])
	//if exist == true {
	//	fmt.Println("内存存在session")
	//	current = a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact]).Serialize()
	//} else {

	getReq := etcd.GetReq{
		Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
		Options: nil,
		ResChan: make(chan etcd.GetRes),
	}
	a.signal.getChan <- getReq
	res := <-getReq.ResChan

	if res.Result.Kvs == nil {
		//初始化
		requestId := fmt.Sprintf("%s-%d", a.timestamp, a.id)
		a.id++
		a.getPrekeyBundleWithPhoneNumberJid(requestId, recipientJid)

		a.PrekeyChan[contact] = make(chan struct{})
		select {
		case <-a.PrekeyChan[contact]:

			preKeyID := optional.NewOptionalUint32(a.PrekeyBundles[contact].OneTimePreKeyId)
			preKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].OneTimePreKey))

			signedPreKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].SignedPreKey))
			identitykeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].IdentityKey))
			keys := identity.NewKey(identitykeyPublic)

			bundle := prekey.NewBundle(a.PrekeyBundles[contact].RegistrationId, 0, preKeyID,
				a.PrekeyBundles[contact].SignedPreKeyId, preKeyPublic, signedPreKeyPublic, *a.PrekeyBundles[contact].SignedPreKeySignature, keys)

			a.SessionStore[contact].ProcessBundle(bundle)

		case <-time.After(time.Second * 90):
			level.Info(a.logger).Log("msg", "Time up.exit", "user", a.ClientPayLoad.Username, "contact", contact)
			return
		}

		loadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact])

		insert := etcd.TxnInsertReq{
			Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
			Value:   string(loadedSession.Serialize()),
			Options: nil,
			ResChan: make(chan etcd.TxnInsertRes),
		}
		a.signal.txnInsertChan <- insert

		current = loadedSession.Serialize()

	} else {
		//从etcd获取session
		current = res.Result.Kvs[0].Value

	}
	//}

	//反序列化session
	deserializedSession, _ := record.NewSessionFromBytes(current,
		a.SignalprotocolStore.SessionStore.serializer.Session,
		a.SignalprotocolStore.SessionStore.serializer.State)

	//加载session相关
	deserializedSession.SessionState().SetLocalRegistrationID(uint32(a.RegistrationID))
	a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[contact], deserializedSession)
	a.SignalprotocolStore.IdentityKeyStore.SaveIdentity(a.RemoteAddresses[contact], deserializedSession.SessionState().RemoteIdentityKey())

	cipher := session.NewCipherFromSession(a.RemoteAddresses[contact], a.SignalprotocolStore.SessionStore,
		a.SignalprotocolStore.PreKeyStore, a.SignalprotocolStore.IdentityKeyStore, a.Serializer.PreKeySignalMessage, a.Serializer.SignalMessage)
	textMsgCallbacks := make([]callback.TextMsgCallback, 0)
	for _, text := range texts {
		content := proto.String(text)
		now := proto.Uint64(uint64(time.Now().Unix()))

		version := proto.Int32(2)
		//创建明文消息
		msg := &protobuf.Message{
			Conversation: content,
			MessageContextInfo: &protobuf.MessageContextInfo{
				DeviceListMetadata: &protobuf.DeviceListMetadata{
					SenderTimestamp: now,
				},
				DeviceListMetadataVersion: version,
			},
		}

		msgByte, _ := proto.Marshal(msg)

		ciphertext, _ := cipher.Encrypt(bytesWithPadding(msgByte))

		request := createMessageRequest(recipientJid, ciphertext)
		composing := createComposing(recipientJid)
		paused := createPaused(recipientJid)

		a.send(composing)
		a.send(paused)

		a.send(request)
		msgCallback := callback.TextMsgCallback{
			Sender:    a.UserJid,
			Receiver:  contact,
			SendText:  text,
			SendTime:  time.Now(),
			MsgType:   1,
			ReqId:     request.GetAttributeByKey(SIGNALID).GetValue().GetString_(),
			Read:      1,
			Initiator: a.UserJid,
		}
		textMsgCallbacks = append(textMsgCallbacks, msgCallback)
	}

	if a.Cipher[contact] == nil {
		a.Cipher[contact] = cipher
	}

	//发送消息后再更新一次session
	lastLoadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact])

	//a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[contact], lastLoadedSession)

	put := etcd.TxnUpdateReq{
		Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
		Value:   string(lastLoadedSession.Serialize()),
		Options: nil,
		ResChan: make(chan etcd.TxnUpdateRes),
	}
	a.signal.txnUpdateChan <- put
	//发送消息回调
	if len(textMsgCallbacks) > 0 {
		a.sendTextMsgCallback(textMsgCallbacks)
	}

}

// sendTextMsgCallback 发送消息回调
func (a *Account) sendTextMsgCallback(textMsgCallbacks []callback.TextMsgCallback) {
	a.callbackChan.TextMsgCallbackChan <- textMsgCallbacks
}

func (a *Account) getPrekeyBundleWithPhoneNumberJid(rid string, phoneNumerJid string) {
	msg := createPrekeyBundleXMPP(rid, phoneNumerJid)
	a.send(msg)
}

func (a *Account) createPreKeyBundle(in *protobuf.XmppStanzaElement) {
	preKeyBundle := &PreKeyBundle{}

	if in.GetAttributeByKey(SIGNALTYPE).Value.GetString_() != "result" {
		return
	}

	user := in.GetChildrenByName(SIGNALLIST).GetChildrenByName(SIGNALUSER)

	signedPreKeyData := user.GetChildrenByName(SIGNALSKEY)
	oneTimePreKeyData := user.GetChildrenByName(SIGNALKEY)

	WAPhoneNumberUserJID := user.GetAttributeByKey(SIGNALJID).Value.GetWAPhoneNumberUserJID()
	phoneStr := strings.Split(WAPhoneNumberUserJID, "@")[0]
	phone, _ := strconv.ParseUint(phoneStr, 10, 64)

	if user.GetChildrenByName(SIGNALREGISTERATION).Value.GetData() != nil {
		preKeyBundle.RegistrationId = binary.BigEndian.Uint32(user.GetChildrenByName(SIGNALREGISTERATION).Value.GetData())
	} else {
		preKeyBundle.RegistrationId = binary.BigEndian.Uint32([]byte(user.GetChildrenByName(SIGNALREGISTERATION).Value.GetString_()))
	}

	if user.GetChildrenByName(SIGNALIDENTITY).Value.GetData() != nil {
		preKeyBundle.IdentityKey = user.GetChildrenByName(SIGNALIDENTITY).Value.GetData()
	} else {
		preKeyBundle.IdentityKey = []byte(user.GetChildrenByName(SIGNALIDENTITY).Value.GetString_())
	}

	if signedPreKeyData.GetChildrenByName(SIGNALID).Value.GetData() != nil {
		preKeyBundle.SignedPreKeyId = binary.BigEndian.Uint32(append([]byte{0}, signedPreKeyData.GetChildrenByName(SIGNALID).Value.GetData()...))
	} else {
		preKeyBundle.SignedPreKeyId = binary.BigEndian.Uint32(append([]byte{0}, []byte(signedPreKeyData.GetChildrenByName(SIGNALID).Value.GetString_())...))
	}

	if signedPreKeyData.GetChildrenByName(SIGNALSIGNATURE).Value.GetData() != nil {
		preKeyBundle.SignedPreKeySignature = (*[64]byte)(signedPreKeyData.GetChildrenByName(SIGNALSIGNATURE).Value.GetData())
	} else {
		preKeyBundle.SignedPreKeySignature = (*[64]byte)([]byte(signedPreKeyData.GetChildrenByName(SIGNALSIGNATURE).Value.GetString_()))
	}

	if signedPreKeyData.GetChildrenByName(SIGNALVALUE).Value.GetData() != nil {
		preKeyBundle.SignedPreKey = signedPreKeyData.GetChildrenByName(SIGNALVALUE).Value.GetData()
	} else {
		preKeyBundle.SignedPreKey = []byte(signedPreKeyData.GetChildrenByName(SIGNALVALUE).Value.GetString_())
	}

	if oneTimePreKeyData.GetChildrenByName(SIGNALID).Value.GetData() != nil {
		preKeyBundle.OneTimePreKeyId = binary.BigEndian.Uint32(append([]byte{0}, oneTimePreKeyData.GetChildrenByName(SIGNALID).Value.GetData()...))
	} else {
		preKeyBundle.OneTimePreKeyId = binary.BigEndian.Uint32(append([]byte{0}, []byte(oneTimePreKeyData.GetChildrenByName(SIGNALID).Value.GetString_())...))
	}

	if oneTimePreKeyData.GetChildrenByName(SIGNALVALUE).Value.GetData() != nil {
		preKeyBundle.OneTimePreKey = oneTimePreKeyData.GetChildrenByName(SIGNALVALUE).Value.GetData()
	} else {
		preKeyBundle.OneTimePreKey = []byte(oneTimePreKeyData.GetChildrenByName(SIGNALVALUE).Value.GetString_())
	}

	a.PrekeyBundles[phone] = preKeyBundle
	c := struct{}{}
	a.PrekeyChan[phone] <- c

}

// 生成消息key
func (a *Account) registerEncryptionKeys() {
	serialize1 := serialize.NewProtoBufSerializer()
	pair1, _ := ecc.GenerateKeyPair()

	local := a.IdentityKeyPair

	signature := ecc.CalculateSignature(local.PrivateKey(), pair1.PublicKey().Serialize())

	signedPreKeyId := getRandomSignedPreKeyId()

	signedPreKeyTimestamp := time.Now().Unix()
	signedPreKey := record.NewSignedPreKey(signedPreKeyId, signedPreKeyTimestamp, pair1, signature, serialize1.SignedPreKeyRecord)

	preKeyRecords := make([]*record.PreKey, 1000)

	pair2, _ := ecc.GenerateKeyPair()
	//RegistrationID, _ := strconv.ParseInt("270ffeee", 16, 32)
	for i := 0; i < len(preKeyRecords); i++ {
		preKeyRecords[i] = record.NewPreKey(uint32(i+1), pair2, serialize1.PreKeyRecord)
	}

	req := SetEncryptionKeysRequest(a.RegistrationID, local.PublicKey().PublicKey(), signedPreKey, preKeyRecords)
	//for _, v := range preKeyRecords {
	//	fmt.Printf("序号：%v,数据：%v", v.ID(), hex.EncodeToString(bytehelper.ArrayToSlice(v.KeyPair().PrivateKey().Serialize())))
	//}
	a.send(req)
	//fmt.Printf("测试： %v  结束", req)

	//存储key

	//return preKeyRecords

}

func (a *Account) decryptMessage(senderId, encType string, cipherText []byte) string {

	msg := &protobuf.Message{}
	cipher := &session.Cipher{}

	phoneStr := strings.Split(senderId, "@")[0]
	phone, _ := strconv.ParseUint(phoneStr, 10, 64)

	if a.RemoteAddresses[phone] == nil {
		a.RemoteAddresses[phone] = protocol.NewSignalAddress(senderId, 0)
	}

	if a.Cipher[phone] != nil {
		cipher = a.Cipher[phone]
		level.Info(a.logger).Log("Cipher", cipher)
		fmt.Println(cipher)
	} else {

		//getReq := etcd.GetReq{
		//	Key:     strconv.FormatUint(phone, 10) + "Session",
		//	Options: nil,
		//	ResChan: make(chan etcd.GetRes),
		//}
		//a.signal.getChan <- getReq
		//res := <-getReq.ResChan
		//
		//if res.Result.Kvs != nil {
		//	fmt.Println("读session数据")
		//	//反序列化session
		//	deserializedSession, _ := record.NewSessionFromBytes(res.Result.Kvs[0].Value,
		//		a.SignalprotocolStore.SessionStore.Serializer.Session,
		//		a.SignalprotocolStore.SessionStore.Serializer.State)
		//
		//	//加载session相关
		//	a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[phone], deserializedSession)
		//	a.SignalprotocolStore.IdentityKeyStore.SaveIdentity(a.RemoteAddresses[phone], deserializedSession.SessionState().RemoteIdentityKey())
		//
		//	Cipher = Session.NewCipherFromSession(a.RemoteAddresses[phone], a.SignalprotocolStore.SessionStore,
		//		a.SignalprotocolStore.PreKeyStore, a.SignalprotocolStore.IdentityKeyStore, a.Serializer.PreKeySignalMessage, a.Serializer.SignalMessage)
		//}

	}

	switch encType {
	case "msg":

		bytes, _ := protocol.NewSignalMessageFromBytes(cipherText, a.Serializer.SignalMessage)

		decrypt, err := cipher.Decrypt(bytes)
		if err != nil {
			fmt.Println("err:", err)
		}
		proto.Unmarshal(decrypt, msg)

	case "pkmsg":
		bytes, _ := protocol.NewPreKeySignalMessageFromBytes(cipherText, a.Serializer.PreKeySignalMessage, a.Serializer.SignalMessage)
		decrypt, _ := cipher.DecryptMessage(bytes)

		proto.Unmarshal(decrypt, msg)

	case "skmsg":
	default:
		level.Info(a.logger).Log("msg", "Get msg type", "encType", "")
	}

	//a.Cipher[phone] = Cipher

	//发送消息后再更新一次session
	//lastLoadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[phone])
	//a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[phone], lastLoadedSession)
	//put := etcd.TxnUpdateReq{
	//	Key:     strconv.FormatUint(phone, 10) + "Session",
	//	Value:   string(lastLoadedSession.Serialize()),
	//	Options: nil,
	//	ResChan: make(chan etcd.TxnUpdateRes),
	//}
	//a.signal.txnUpdateChan <- put

	return stringFromPtr(msg.Conversation)
}
