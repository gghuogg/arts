package signal

import (
	"arthas/protobuf"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/state/record"
	"go.mau.fi/libsignal/util/bytehelper"
	"google.golang.org/protobuf/proto"
	"strconv"
)

func (m *Manager) createAvailableMessage() *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder(SIGNALPRESENCE).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALAVAILABLE).Build()).
		Build()
}

func (m *Manager) createPingMessage(requestId string) *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder("iq").
		WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(requestId).Build()).
		WithAttribute(NewXmppAttributeBuilder("xmlns").WithStringValue("w:p").Build()).
		WithAttribute(NewXmppAttributeBuilder("type").WithStringValue("get").Build()).
		WithAttribute(NewXmppAttributeBuilder("to").WithWADomainJID("s.whatsapp.net").Build()).
		WithChild(NewXmppStanzaElementBuilder("ping").Build()).
		Build()
}

func (m *Manager) createAckReceiptRequest(recipientId string, messageId string, type_ string) *protobuf.XmppStanzaElement {
	if type_ == "" {
		msg := NewXmppStanzaElementBuilder("ack").
			WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(messageId).Build()).
			WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("receipt").Build()).
			WithAttribute(NewXmppAttributeBuilder("to").WithWAPhoneNumberUserJID(recipientId).Build()).
			Build()
		return msg
	}
	return NewXmppStanzaElementBuilder("ack").
		WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(messageId).Build()).
		WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("receipt").Build()).
		WithAttribute(NewXmppAttributeBuilder("to").WithWAPhoneNumberUserJID(recipientId).Build()).
		WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
		Build()
}

func (m *Manager) createAckNotificationPsaRequest(userid string, notificationId string, type_ string, participant string) *protobuf.XmppStanzaElement {
	if participant != "" {
		msg := NewXmppStanzaElementBuilder("ack").
			WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
			WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("notification").Build()).
			WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
			WithAttribute(NewXmppAttributeBuilder("to").WithWAPhoneNumberUserJID(userid).Build()).
			WithAttribute(NewXmppAttributeBuilder("participant").WithWAPhoneNumberUserJID(participant).Build()).
			Build()
		return msg
	}

	return NewXmppStanzaElementBuilder("ack").
		WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
		WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("notification").Build()).
		WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
		WithAttribute(NewXmppAttributeBuilder("to").WithWAPhoneNumberUserJID(userid).Build()).
		Build()

}

func (m *Manager) createAckNotificationTokenRequest(userid string, notificationId string, type_ string, participant string) *protobuf.XmppStanzaElement {
	if participant != "" {
		msg := NewXmppStanzaElementBuilder("ack").
			WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
			WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("notification").Build()).
			WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
			WithAttribute(NewXmppAttributeBuilder("to").WithWADomainJID(userid).Build()).
			WithAttribute(NewXmppAttributeBuilder("participant").WithWAPhoneNumberUserJID(participant).Build()).
			Build()
		return msg
	}

	return NewXmppStanzaElementBuilder("ack").
		WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
		WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("notification").Build()).
		WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
		WithAttribute(NewXmppAttributeBuilder("to").WithWADomainJID(userid).Build()).
		Build()
}

func (m *Manager) createAckNotificationBusinessRequest(userid string, notificationId string, type_ string, participant string) *protobuf.XmppStanzaElement {
	if participant != "" {
		msg := NewXmppStanzaElementBuilder("ack").
			WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
			WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("notification").Build()).
			WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
			WithAttribute(NewXmppAttributeBuilder("to").WithWADomainJID(userid).Build()).
			WithAttribute(NewXmppAttributeBuilder("participant").WithWAPhoneNumberUserJID(participant).Build()).
			Build()
		return msg
	}

	return NewXmppStanzaElementBuilder("ack").
		WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
		WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("notification").Build()).
		WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
		WithAttribute(NewXmppAttributeBuilder("to").WithWADomainJID(userid).Build()).
		Build()

}

func (m *Manager) createAckNotificationEncryptRequest(tag string, userid string, notificationId string, type_ string, participant string, recipient string) *protobuf.XmppStanzaElement {
	builder := NewXmppStanzaElementBuilder("ack").
		WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
		WithAttribute(NewXmppAttributeBuilder("class").WithStringValue(tag).Build()).
		WithAttribute(NewXmppAttributeBuilder("type").WithStringValue(type_).Build()).
		WithAttribute(NewXmppAttributeBuilder("to").WithWADomainJID(userid).Build())
	if participant != "" {
		builder = builder.
			WithAttribute(NewXmppAttributeBuilder("participant").WithWAPhoneNumberUserJID(participant).Build())
	}
	if recipient != "" {
		builder = builder.
			WithAttribute(NewXmppAttributeBuilder("recipient").WithWAPhoneNumberUserJID(recipient).Build())
	}
	return builder.Build()
}

func createPrekeyBundleXMPP(rid string, phoneNumJid string) *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(rid).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue(SIGNALENCRYPT).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALGET).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWADomainJID(SIGNALDOMAINWS).Build()).
		WithChild(
			NewXmppStanzaElementBuilder(SIGNALKEY).
				WithChild(
					NewXmppStanzaElementBuilder(SIGNALUSER).
						WithAttribute(NewXmppAttributeBuilder(SIGNALJID).WithWAPhoneNumberUserJID(phoneNumJid).Build()).
						Build()).
				Build()).
		Build()
}

func (m *Manager) SendVcardMsgRequst(recipientJid string, mediaType string, ciphertextMessage protocol.CiphertextMessage, reqId string) *protobuf.XmppStanzaElement {
	vvalue := "2"
	msg := "pkmsg"

	messageBuilder := NewXmppStanzaElementBuilder(SIGNALMESSAGE).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(reqId).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALTEXT).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWAPhoneNumberUserJID(recipientJid).Build())

	encBuilder := NewXmppStanzaElementBuilder(SIGNALENC).
		WithDataValue(ciphertextMessage.Serialize()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALV).WithStringValue(vvalue).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALMEDIATYPE).WithStringValue(mediaType).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(msg).Build())

	messageBuilder.WithChild(encBuilder.Build())

	return messageBuilder.Build()
}

func (m *Manager) CreateLoginClientPayload(username uint64) (*protobuf.ClientPayload, error) {
	sessionID := getRandomSessionId()
	passive := false
	appcached := false
	userAgentPlatform := protobuf.ClientPayload_UserAgent_SMB_IOS
	version := m.GetAppVersionFromEtcd()
	appVersion := &protobuf.ClientPayload_UserAgent_AppVersion{
		Primary:    proto.Uint32(*version.Primary),
		Secondary:  proto.Uint32(*version.Secondary),
		Tertiary:   proto.Uint32(*version.Tertiary),
		Quaternary: proto.Uint32(*version.Quaternary),
	}
	deviceConifg := GetRandomUserDeviceConfig()
	//deviceConifg := userDeviceCifg{
	//	OsVersion:     proto.String("16.1"),
	//	Device:        proto.String("iPhone 12"),
	//	OsBuildNumber: proto.String("20F66"),
	//	PhoneId:       proto.String("08150008-47a5-4576-b108-fd925e937941"),
	//}
	userAgent := &protobuf.ClientPayload_UserAgent{
		Platform:                       &userAgentPlatform,
		AppVersion:                     appVersion,
		Mcc:                            proto.String("000"),
		Mnc:                            proto.String("000"),
		OsVersion:                      deviceConifg.OsVersion,
		Manufacturer:                   proto.String("Apple"),
		Device:                         deviceConifg.Device,
		OsBuildNumber:                  deviceConifg.OsBuildNumber,
		PhoneId:                        deviceConifg.PhoneId,
		ReleaseChannel:                 protobuf.ClientPayload_UserAgent_RELEASE.Enum(),
		LocaleLanguageIso_639_1:        proto.String("en"),
		LocaleCountryIso_3166_1Alpha_2: proto.String("HK"),
	}
	clientPayload := &protobuf.ClientPayload{
		Username:      &username,
		Passive:       &passive,
		UserAgent:     userAgent,
		SessionId:     &sessionID,
		ShortConnect:  proto.Bool(true),
		ConnectType:   protobuf.ClientPayload_WIFI_UNKNOWN.Enum(),
		ConnectReason: protobuf.ClientPayload_USER_ACTIVATED.Enum(),
		DnsSource: &protobuf.ClientPayload_DNSSource{
			DnsMethod: protobuf.ClientPayload_DNSSource_SYSTEM.Enum(),
			AppCached: &appcached,
		},
		ConnectAttemptCount: proto.Uint32(0),
		Device:              proto.Uint32(0),
		Oc:                  proto.Bool(false),
	}

	return clientPayload, nil
}

func createMessageRequest(recipientJid string, ciphertextMessage protocol.CiphertextMessage) *protobuf.XmppStanzaElement {
	vvalue := "2"
	msg := "pkmsg"
	return NewXmppStanzaElementBuilder(SIGNALMESSAGE).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getRequestId()).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALTEXT).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWAPhoneNumberUserJID(recipientJid).Build()).
		WithChild(
			NewXmppStanzaElementBuilder(SIGNALENC).
				WithDataValue(ciphertextMessage.Serialize()).
				WithAttribute(NewXmppAttributeBuilder(SIGNALV).WithStringValue(vvalue).Build()).
				WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(msg).Build()).
				Build()).
		Build()
}

func createMediaMessageRequest(recipientJid, mediaType string, ciphertextMessage protocol.CiphertextMessage) *protobuf.XmppStanzaElement {
	vvalue := "2"
	msg := "pkmsg"

	messageBuilder := NewXmppStanzaElementBuilder(SIGNALMESSAGE).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getRequestId()).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALMEDIA).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWAPhoneNumberUserJID(recipientJid).Build())

	encBuilder := NewXmppStanzaElementBuilder(SIGNALENC).
		WithDataValue(ciphertextMessage.Serialize()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALV).WithStringValue(vvalue).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALMEDIATYPE).WithStringValue(mediaType).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(msg).Build())

	messageBuilder.WithChild(encBuilder.Build())

	return messageBuilder.Build()
}

func createMediaConnRequest(lastMediaId int) *protobuf.XmppStanzaElement {
	xmlnsvalue := "w:m"
	iqBuilder := NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getNextRequestId(0)).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue(xmlnsvalue).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALSET).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).
			WithWADomainJID(SIGNALDOMAINWS).Build())

	mediaConnBuilder := NewXmppStanzaElementBuilder(SIGNALMEDIACONN)

	if lastMediaId != 0 {
		mediaConnBuilder.WithAttribute(NewXmppAttributeBuilder("last_id").
			WithStringValue(strconv.Itoa(lastMediaId)).Build())
	}

	iqBuilder.WithChild(mediaConnBuilder.Build())

	return iqBuilder.Build()
}

func (m *Manager) ProfilePicturePreviewRequest(userid string) *protobuf.XmppStanzaElement {
	picture := NewXmppStanzaElementBuilder(SIGNALPICTURE).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALPREVIEW).Build()).
		Build()

	return NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getNextRequestId(0)).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue("w:profile:picture").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALGET).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).
			WithWADomainJID(SIGNALDOMAINWS).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTARGET).
			WithWAPhoneNumberUserJID(userid).Build()).
		WithChild(picture).
		Build()
}

// 设置头像
func (m *Manager) SetProfilePictureRequest(image []byte) *protobuf.XmppStanzaElement {
	picture := NewXmppStanzaElementBuilder(SIGNALPICTURE).
		WithDataValue(image).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue("image").Build()).
		Build()
	return NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getNextRequestId(0)).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue("w:profile:picture").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALSET).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).
			WithWADomainJID(SIGNALDOMAINWS).Build()).
		WithChild(picture).
		Build()
}

// 发送文件
func (m *Manager) StatFileRequest(data []byte, timestamp string) *protobuf.XmppStanzaElement {
	file := NewXmppStanzaElementBuilder(SIGNALADD).
		WithDataValue(data).
		WithAttribute(NewXmppAttributeBuilder(SIGNALT).WithStringValue(timestamp).Build()).
		Build()

	return NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getNextRequestId(0)).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue("w:stats").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALSET).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).
			WithWADomainJID(SIGNALDOMAINWS).Build()).
		WithChild(file).
		Build()

}

func (m *Manager) SyncContactRequest(userJID string, phoneNumbers []uint64, syncID string) *protobuf.XmppStanzaElement {
	query := NewXmppStanzaElementBuilder(SIGNALQUERY).
		WithChild(NewXmppStanzaElementBuilder(SIGNALBUSINESS).
			WithChild(NewXmppStanzaElementBuilder(SIGNALVERIFIEDNAME).Build()).
			WithChild(NewXmppStanzaElementBuilder(SIGNALPROFILE).
				WithAttribute(NewXmppAttributeBuilder(SIGNALV).WithStringValue("372").Build()).
				Build()).
			Build()).
		WithChild(NewXmppStanzaElementBuilder(SIGNALCONTACT).Build()).
		WithChild(NewXmppStanzaElementBuilder(SIGNALDEVICES).
			WithAttribute(NewXmppAttributeBuilder(SIGNALVERSION).WithStringValue("2").Build()).
			Build()).
		WithChild(NewXmppStanzaElementBuilder(SIGNALDISAPPEARINGMODE).Build()).
		WithChild(NewXmppStanzaElementBuilder(SIGNALLID).Build()).
		Build()

	list := NewXmppStanzaElementBuilder(SIGNALLIST)
	for _, phoneNumber := range phoneNumbers {
		list = list.WithChild(NewXmppStanzaElementBuilder(SIGNALUSER).
			WithChild(NewXmppStanzaElementBuilder(SIGNALCONTACT).
				WithStringValue(strconv.FormatUint(phoneNumber, 10)).Build()).
			Build())
	}

	sideList := NewXmppStanzaElementBuilder(SIGNALSIDELIST).Build()

	return NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getNextRequestId(0)).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue("usync").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALGET).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).
			WithWAPhoneNumberUserJID(userJID).Build()).
		WithChild(NewXmppStanzaElementBuilder(SIGNALUSYNC).
			WithAttribute(NewXmppAttributeBuilder(SIGNALMODE).WithStringValue("delta").Build()).
			WithAttribute(NewXmppAttributeBuilder(SIGNALCONTEXT).WithStringValue("background").Build()).
			WithAttribute(NewXmppAttributeBuilder(SIGNALLAST).WithStringValue("true").Build()).
			WithAttribute(NewXmppAttributeBuilder(SIGNALSID).WithStringValue(syncID).Build()).
			WithAttribute(NewXmppAttributeBuilder(SIGNALINDEX).WithStringValue("0").Build()).
			WithChild(query).
			WithChild(list.Build()).
			WithChild(sideList).
			Build()).
		Build()
}

func uint32ToString(value uint32) string {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, value)

	// 我们需要前三个字节，因为1058550是一个24位的数，只需要3个字节就能表示
	bytes = bytes[1:4]

	encoded := base64.StdEncoding.EncodeToString(bytes)
	return encoded
}

func SetEncryptionKeysRequest(registrationId int64, identityKey ecc.ECPublicKeyable, signedPreKey *record.SignedPreKey, preKeyRecords []*record.PreKey) *protobuf.XmppStanzaElement {

	list := NewXmppStanzaElementBuilder(SIGNALLIST)
	for _, preKey := range preKeyRecords {
		keyIdv := uint32ToString(preKey.ID().Value)
		list.WithChild(
			NewXmppStanzaElementBuilder(SIGNALKEY).
				WithChild(
					NewXmppStanzaElementBuilder(SIGNALID).
						WithDataValue([]byte(keyIdv)).
						Build()).
				WithChild(
					NewXmppStanzaElementBuilder(SIGNALVALUE).
						WithDataValue([]byte(base64.StdEncoding.EncodeToString(bytehelper.ArrayToSlice(preKey.KeyPair().PublicKey().PublicKey())))).Build()).
				Build())
	}

	registrationv := base64.StdEncoding.EncodeToString(uint32ToBytes(uint32(registrationId)))

	b, _ := hex.DecodeString("05")
	typevalue := base64.StdEncoding.EncodeToString(b)

	identityv := base64.StdEncoding.EncodeToString(bytehelper.ArrayToSlice(identityKey.PublicKey()))

	skeyIdv := uint32ToString(signedPreKey.ID())

	return NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getNextRequestId(0)).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue(SIGNALENCRYPT).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue(SIGNALSET).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWADomainJID(SIGNALDOMAINWS).Build()).
		WithChild(
			NewXmppStanzaElementBuilder(SIGNALREGISTERATION).
				WithDataValue([]byte(registrationv)).Build()).
		WithChild(
			NewXmppStanzaElementBuilder(SIGNALTYPE).
				WithDataValue([]byte(typevalue)).Build()).
		WithChild(
			NewXmppStanzaElementBuilder(SIGNALIDENTITY).
				WithDataValue([]byte(identityv)).Build()).
		WithChild(
			NewXmppStanzaElementBuilder(SIGNALSKEY).
				WithChild(
					NewXmppStanzaElementBuilder(SIGNALID).
						WithDataValue([]byte(skeyIdv)).Build()).
				WithChild(
					NewXmppStanzaElementBuilder(SIGNALVALUE).
						WithDataValue([]byte(base64.StdEncoding.EncodeToString(bytehelper.ArrayToSlice(signedPreKey.KeyPair().PublicKey().PublicKey())))).Build()).
				WithChild(
					NewXmppStanzaElementBuilder(SIGNALSIGNATURE).
						WithDataValue([]byte(base64.StdEncoding.EncodeToString(bytehelper.ArrayToSlice64(signedPreKey.Signature())))).Build()).Build()).
		WithChild(list.Build()).
		Build()
}

func createMessageDeliveredRequest(senderId, messageId string) *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder(SIGNALRECEIPT).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(messageId).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue("delivery").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWAPhoneNumberUserJID(senderId).Build()).
		Build()
}

func createMessageReadRequest(senderId, messageId string) *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder(SIGNALRECEIPT).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(messageId).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue("read").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithStringValue(senderId).Build()).
		Build()
}

func createMessageRetryRequest(senderId, messageId string, registrationId int64, timestamp string) *protobuf.XmppStanzaElement {
	registrationIdv := base64.StdEncoding.EncodeToString(uint32ToBytes(uint32(registrationId)))
	return NewXmppStanzaElementBuilder(SIGNALRECEIPT).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(messageId).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue("retry").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWAPhoneNumberUserJID(senderId).Build()).
		WithChild(NewXmppStanzaElementBuilder("retry").
			WithAttribute(NewXmppAttributeBuilder(SIGNALT).WithStringValue(timestamp).Build()).
			WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(messageId).Build()).
			WithAttribute(NewXmppAttributeBuilder("count").WithStringValue("1").Build()).
			WithAttribute(NewXmppAttributeBuilder("v").WithStringValue("1").Build()).Build()).
		WithChild(NewXmppStanzaElementBuilder(SIGNALREGISTERATION).WithDataValue([]byte(registrationIdv)).
			Build()).
		Build()
}

func (m *Manager) RemoveAllCompanionDeviceRequest(reason string) *protobuf.XmppStanzaElement {
	device := NewXmppStanzaElementBuilder(SIGNALREMOVEDEVICE).
		WithAttribute(NewXmppAttributeBuilder(SIGNALREASON).WithStringValue(reason).Build()).
		WithAttribute(NewXmppAttributeBuilder("all").WithStringValue("true").Build()).Build()

	return NewXmppStanzaElementBuilder(SIGNALIQ).
		WithAttribute(NewXmppAttributeBuilder(SIGNALID).WithStringValue(getNextRequestId(0)).Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALXMLNS).WithStringValue("md").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTYPE).WithStringValue("set").Build()).
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).
			WithWADomainJID(SIGNALDOMAINWS).Build()).
		WithChild(device).
		Build()
}

func (m *Manager) createErrorAck(userid string, notificationId string) *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder("ack").
		WithAttribute(NewXmppAttributeBuilder("id").WithStringValue(notificationId).Build()).
		WithAttribute(NewXmppAttributeBuilder("class").WithStringValue("notification").Build()).
		WithAttribute(NewXmppAttributeBuilder("type").WithStringValue("error").Build()).
		WithAttribute(NewXmppAttributeBuilder("to").WithWAPhoneNumberUserJID(userid).Build()).
		Build()
}

func createComposing(contact string) *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder("chatstate").
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWAPhoneNumberUserJID(contact).Build()).
		WithChild(NewXmppStanzaElementBuilder("composing").Build()).
		Build()
}

func createPaused(contact string) *protobuf.XmppStanzaElement {
	return NewXmppStanzaElementBuilder("chatstate").
		WithAttribute(NewXmppAttributeBuilder(SIGNALTO).WithWAPhoneNumberUserJID(contact).Build()).
		WithChild(NewXmppStanzaElementBuilder("paused").Build()).
		Build()
}
