package signal

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/state/record"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Account struct {
	logger              log.Logger
	signal              *Manager
	conn                *grpc.ClientConn
	timestamp           string
	id                  uint
	Stream              protobuf.XmppStream_ConnectClient
	reqeust             Request
	RegistrationID      int64
	PrivateKey          []byte
	PublicKey           []byte
	IdentityKeyPair     *identity.KeyPair
	ResumptionSecret    []byte
	ClientPayLoad       *protobuf.ClientPayload
	SignalprotocolStore *LSignalProtocolStore
	MediaConnInfo       map[uint64]*MediaConnInfo
	MediaConnInfoChan   map[uint64]chan struct{}
	SessionStore        map[uint64]*session.Builder
	RemoteAddresses     map[uint64]*protocol.SignalAddress
	Cipher              map[uint64]*session.Cipher
	Session             map[uint64]*record.Session
	PrekeyBundles       map[uint64]*PreKeyBundle
	PrekeyChan          map[uint64]chan struct{}
	Serializer          *serialize.Serializer
	UserJid             uint64
	IsLogin             bool
	proxyUrl            string
	callbackChan        *callback.CallbackChan
	uniqueID            string
	idCounter           uint32
}

type AccountLogin struct {
	isLogin    bool
	grpcServer string
	userJid    uint64
}

func (a *Account) Recv() {
	level.Info(a.logger).Log("msg", "Account begin recv", "user", a.ClientPayLoad.Username)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		in, err := a.Stream.Recv()
		if err != nil {
			level.Error(a.logger).Log("msg", "socket close", "err", err, "user", a.ClientPayLoad.Username)
			level.Info(a.logger).Log("msg", "Recv offline,account is logout.")
			a.IsLogin = false
			if err != io.EOF {
				u, proxyErr := url.Parse(a.proxyUrl)
				if proxyErr == nil {
					if port := u.Port(); port != "" {
						if strings.Contains(err.Error(), port) {
							a.loginCallback(protobuf.AccountStatus_PROXY_ERR, err.Error())
						} else {
							a.loginCallback(protobuf.AccountStatus_FAIL, err.Error())
						}
					}
				} else {
					a.loginCallback(protobuf.AccountStatus_PROXY_ERR, err.Error())
				}
			} else {
				a.signal.logoutCallback(a.UserJid, a.proxyUrl, ERROR_LOGOUT)
			}
			break
		}

		//jsonVar := protojson.Format(in)
		//g.DumpJson(jsonVar)
		//level.Info(a.logger).Log("msg", "Got message", "message", jsonVar, "user", a.ClientPayLoad.Username)
		level.Info(a.logger).Log("msg", "Got message", "message", in, "user", a.ClientPayLoad.Username)
		if *in.Name == SIGNALSUCCESS {
			a.timestamp = in.GetAttributeByKey(SIGNALT).GetValue().GetString_()
		}

		if a.IsLogin == false {
			if *in.Name == SIGNALIB {
				if len(in.Children) > 0 {
					if *in.Children[0].Name == SIGNALOFFLINE {
						level.Info(a.logger).Log("msg", "Recv offline,account is login.")
						a.IsLogin = true
						a.etcdAccountLoginState()
						a.signal.recordConnection(true)
						a.loginCallback(protobuf.AccountStatus_SUCCESS, a.proxyUrl)
						// 记录登录成功
						username := strconv.FormatUint(*a.ClientPayLoad.Username, 10)
						loginSuccessCounter.WithLabelValues(username).Inc()

						a.ping(a.timestamp)

						a.available()
						a.loginCallback(protobuf.AccountStatus_SUCCESS, a.proxyUrl)
						go a.keepAliveLoop(ctx)
					}
				}
			}
		}

		//MediaConnInfo := &MediaConnInfo{}
		switch *in.Name {
		case SIGNALRECEIPT:
			from := in.GetAttributeByKey(SIGNALFROM).Value.GetWAPhoneNumberUserJID()
			messageId := strings.ToUpper(hex.EncodeToString(in.GetAttributeByKey("id").Value.GetData()))
			if messageId == "" {
				messageId = in.GetAttributeByKey("id").Value.GetString_()
			}
			if in.GetChildrenByName("retry") != nil {
				msg := a.signal.createAckReceiptRequest(from, messageId, "retry")
				level.Info(a.logger).Log("msg", "Retry Session", "from", from, "messageID", messageId)
				a.send(msg)
			} else {
				if in.GetAttributeByKey("type") != nil {
					type_ := in.GetAttributeByKey("type").Value.GetString_()
					msg := a.signal.createAckReceiptRequest(from, messageId, type_)
					a.send(msg)
					if type_ == "read" {
						reqId := in.GetAttributeByKey("id").Value.GetString_()
						a.callbackChan.ReadMsgCallbackChan <- []callback.ReadMsgCallback{{ReqId: reqId}}
					}

				} else {
					msg := a.signal.createAckReceiptRequest(from, messageId, "")
					a.send(msg)
				}
			}
		case SIGNALNOTIFICATION:
			var id string
			if in.GetAttributeByKey(SIGNALID).Value.GetData() != nil {
				id = hex.EncodeToString(in.GetAttributeByKey(SIGNALID).Value.GetData())
			} else {
				s := in.GetAttributeByKey(SIGNALID).Value.GetString_()
				id = s
			}

			level.Info(a.logger).Log("msg", "Get notifacation id", "id", id, "user", a.ClientPayLoad.Username)
			if len(id) > 10 {
				id = id[0:9]
			}
			type_ := in.GetAttributeByKey(SIGNALTYPE).Value.GetString_()
			var participant string
			if in.GetAttributeByKey("participant").GetKey() == "" {
				participant = ""
			} else {
				participant = in.GetAttributeByKey("participant").Value.GetWAPhoneNumberUserJID()
			}
			var recipient string
			if in.GetAttributeByKey("recipient").GetKey() == "" {
				recipient = ""
			} else {
				recipient = in.GetAttributeByKey("recipient").Value.GetWAPhoneNumberUserJID()
			}
			if !a.IsLogin {
				a.IsLogin = true
				a.etcdAccountLoginState()
				// 记录登录成功
				username := strconv.FormatUint(*a.ClientPayLoad.Username, 10)
				loginSuccessCounter.WithLabelValues(username).Inc()

				a.loginCallback(protobuf.AccountStatus_SUCCESS, a.proxyUrl)
				go a.keepAliveLoop(ctx)
			}

			level.Info(a.logger).Log("msg", "Get notification", "type", type_)
			if type_ == "business" {
				userID := in.GetAttributeByKey(SIGNALFROM).Value.GetWADomainJID()
				msg := a.signal.createAckNotificationBusinessRequest(userID, id, type_, participant)
				a.send(msg)
			} else if type_ == "psa" || type_ == "devices" {
				userID := in.GetAttributeByKey(SIGNALFROM).Value.GetWAPhoneNumberUserJID()
				msg := a.signal.createAckNotificationPsaRequest(userID, id, type_, participant)
				a.send(msg)

			} else if type_ == "privacy_token" {
				userID := in.GetAttributeByKey(SIGNALFROM).Value.GetWAPhoneNumberUserJID()
				msg := a.signal.createAckNotificationTokenRequest(userID, id, type_, participant)
				a.send(msg)
			} else if type_ == "encrypt" {
				userID := in.GetAttributeByKey(SIGNALFROM).Value.GetWAPhoneNumberUserJID()
				msg := a.signal.createAckNotificationEncryptRequest(*in.Name, userID, id, type_, participant, recipient)
				a.send(msg)

			} else {
				userID := in.GetAttributeByKey(SIGNALFROM).Value.GetWAPhoneNumberUserJID()
				msg := a.signal.createAckNotificationEncryptRequest(*in.Name, userID, id, type_, participant, recipient)
				a.send(msg)
			}
		case SIGNALFAIlURE:
			// 登录失败 封号
			username := strconv.FormatUint(*a.ClientPayLoad.Username, 10)
			if reason := in.GetAttributeByKey("reason"); reason != nil {
				if status := reason.GetValue().GetString_(); status == "401" {
					level.Error(a.logger).Log("msg", "Authorization failed", "user", *a.ClientPayLoad.Username)
					a.signal.blockedAccountSet[*a.ClientPayLoad.Username] = struct{}{}
					//普罗米修斯监控数据
					loginFailureCounter.WithLabelValues(username, status).Inc()
					//消息回调
					a.loginCallback(protobuf.AccountStatus_PERMISSION, "")

					break
				} else if status == "403" {
					// 普罗米修斯封号记录
					accountBannedCounter.WithLabelValues(username, status).Inc()
					//消息回调
					a.loginCallback(protobuf.AccountStatus_SEAL, "")
				}

			} else {
				a.loginCallback(protobuf.AccountStatus_FAIL, err.Error())
			}
		case SIGNALIQ:
			if in.GetAttributeByKey("xmlns") != nil {
				if in.GetAttributeByKey("xmlns").Value.GetString_() == "urn:xmpp:ping" {
					t := in.GetAttributeByKey("t").Value.GetString_()
					a.ping(t)
				}
			}
			// 同步通信录
			if fromElem := in.GetAttributeByKey("from"); fromElem != nil {
				//var registeredNumbers []string // 存储已注册的电话号码
				if usyncChild := in.GetChildrenByName("usync"); usyncChild != nil {
					if listChild := usyncChild.GetChildrenByName("list"); listChild != nil {
						if userChild := listChild.GetChildrenByName("user"); listChild != nil {
							if contactChild := userChild.GetChildrenByName("contact"); contactChild != nil {
								contactAttr := contactChild.GetAttributes()[0]
								phone := contactChild.GetValue().GetString_()
								status := contactAttr.GetValue().GetString_()
								a.syncContactCallback(a.UserJid, phone, status)
								if contactAttr.GetValue().GetString_() == "in" {
									println("状态为in下的手机号:", phone)
									//a.sendInPhone(phone)
								} else {
									//a.sendOutPhone(phone)
									println("状态为out下的手机号:", phone)
								}
							}
						}
					}
				}
			}
			// 获取头像
			if pictureChild := in.GetChildrenByName("picture"); pictureChild != nil {
				picture := pictureChild.GetValue().GetData()
				a.getUserHeadImageCallback("", picture)
			}
			if in.GetChildrenByName(SIGNALLIST) != nil {

				user := in.GetChildrenByName(SIGNALLIST).GetChildrenByName(SIGNALUSER)

				//一次性预密钥用完，重新生成
				if user.GetChildrenByName(SIGNALKEY).GetName() == "" {
					a.registerEncryptionKeys()
				}

				a.createPreKeyBundle(in)
			}
		case SIGNALMESSAGE:
			senderId := in.GetAttributeByKey(SIGNALFROM).Value.GetWAPhoneNumberUserJID()
			msgType := in.GetAttributeByKey(SIGNALTYPE).Value.GetString_()

			msgId := ""
			if in.GetAttributeByKey(SIGNALID).Value.GetData() == nil {
				msgId = in.GetAttributeByKey(SIGNALID).Value.GetString_()
			} else {
				msgId = base64.StdEncoding.EncodeToString(in.GetAttributeByKey(SIGNALID).Value.GetData())
			}

			senderName := in.GetAttributeByKey("notify").Value.GetString_()
			//timeStamp := in.GetAttributeByKey(SIGNALT).Value.GetString_()
			encType := in.GetChildrenByName(SIGNALENC).GetAttributeByKey(SIGNALTYPE).Value.GetString_()
			cipherText := in.GetChildrenByName(SIGNALENC).Value.GetData()

			message := a.decryptMessage(senderId, encType, cipherText)
			request := createMessageDeliveredRequest(senderId, msgId)
			//receiptRequest := a.signal.createAckReceiptRequest(senderId, msgId, "")

			a.send(request)

			phoneStr := strings.Split(senderId, "@")[0]
			phone, _ := strconv.ParseUint(phoneStr, 10, 64)
			//发送消息后再更新一次session
			lastLoadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[phone])
			put := etcd.PutReq{
				Key:     strconv.FormatUint(phone, 10) + "Session",
				Value:   string(lastLoadedSession.Serialize()),
				Options: nil,
				ResChan: make(chan etcd.PutRes),
			}
			a.signal.putChan <- put

			switch msgType {
			case SIGNALTEXT:
				fmt.Printf("[text-message] %s %s: %s\n", senderName, senderId, message)
				a.receiveMsgCallback(senderId, msgId, message)
				break
			default:
				fmt.Printf("[message] %s %s %s\n", senderName, senderId, msgId)
			}

		case SIGNALACK:

		}

		//if preKeyBundle.RegistrationId != 0 {

		//	a.signal.EstablishRemoteSessionAndSendPlaintextMessage(a.stream, preKeyBundle, a.ClientPayLoad.Username)
		//}

		//data := &protobuf.AccountLogin{
		//	UserJid:    uint64FromPtr(a.ClientPayLoad.Username),
		//	IsLogin:    a.IsLogin,
		//	GrpcServer: a.signal.GrpcServer,
		//}
		//marshal, _ := proto.Marshal(data)
		//insert := etcd.TxnInsertReq{
		//	Key:     "/arthas/accountlogin/" + strconv.FormatUint(uint64FromPtr(a.ClientPayLoad.Username), 10),
		//	Value:   string(marshal),
		//	Options: nil,
		//	ResChan: make(chan etcd.TxnInsertRes),
		//}
		//a.signal.txnInsertChan <- insert

	}

}

func (a *Account) etcdAccountLoginState() {
	a.ServiceWithUserJid()
	a.UserJidProxyIp()

	req := etcd.LeaseReq{ResChan: make(chan etcd.LeaseRes)}
	a.signal.leaseChan <- req
	res := <-req.ResChan
	lease := res.Result

	leaseGrant, err := lease.Grant(a.signal.context, 30)
	if err != nil {
		level.Error(a.logger).Log("msg", "leaseGrant err", "err", err)
	}

	etcdvalue := &protobuf.AccountLogin{
		IsLogin:    a.IsLogin,
		GrpcServer: a.signal.ListenAddress,
		UserJid:    a.UserJid,
	}
	bvalue, _ := proto.Marshal(etcdvalue)
	put := etcd.PutReq{
		Key:     "/services/" + "whatsapp" + "/" + strconv.FormatUint(a.UserJid, 10),
		Value:   string(bvalue),
		Options: []clientv3.OpOption{clientv3.WithLease(leaseGrant.ID)},
		ResChan: make(chan etcd.PutRes),
	}
	a.signal.putChan <- put

	keepRespChan, err := lease.KeepAlive(a.signal.context, leaseGrant.ID)
	if err != nil {
		level.Error(a.logger).Log("msg", "lease keepAlive err", "err", err)
	}

	go func() {
		for {
			select {
			case resp := <-keepRespChan:
				if resp == nil {
					level.Info(a.logger).Log("msg", "lease keepAlive fail")
					return
				}
			}
		}
	}()
}

func (a *Account) ServiceWithUserJid() {
	putReq := etcd.PutReq{
		Key:     "/service/" + a.signal.ListenAddress + "/" + strconv.FormatUint(a.UserJid, 10),
		Value:   "",
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	a.signal.putChan <- putReq
}

func (a *Account) UserJidProxyIp() {
	putReq := etcd.PutReq{
		Key:     "/service/proxyaddr/" + strconv.FormatUint(a.UserJid, 10),
		Value:   a.proxyUrl,
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	a.signal.putChan <- putReq
}

func (a *Account) loginCallback(loginStatus protobuf.AccountStatus, comment string) {
	loginCallbacks := make([]callback.LoginCallback, 0)

	item := callback.LoginCallback{
		UserJid:     a.UserJid,
		ProxyUrl:    a.proxyUrl,
		LoginStatus: loginStatus,
		Comment:     comment,
	}
	loginCallbacks = append(loginCallbacks, item)

	a.callbackChan.LoginCallbackChan <- loginCallbacks
}

func (a *Account) syncContactCallback(accountDb uint64, synchro string, status string) {
	syncContactCallbacks := make([]callback.SyncContactCallback, 0)
	syncContactCallbacks = append(syncContactCallbacks, callback.SyncContactCallback{
		AccountDb: accountDb,
		Status:    status,
		Synchro:   synchro,
	})
	a.callbackChan.SyncContactCallbackChan <- syncContactCallbacks
}

// receiveMsgCallback 接收消息回调 senderId
func (a *Account) receiveMsgCallback(senderId string, msgId string, message string) {
	level.Info(a.logger).Log("msg", "receiveMsgCallback")
	sender := strings.Replace(senderId, "@s.whatsapp.net", "", 1)
	parseUint, err := strconv.ParseUint(sender, 10, 64)
	if err != nil {
		level.Error(a.logger).Log("msg", "Failed to convert sender", "err", err)
	}
	textMsgCallbacks := make([]callback.TextMsgCallback, 0)
	textMsgCallbacks = append(textMsgCallbacks, callback.TextMsgCallback{
		Sender:    parseUint,
		Receiver:  a.UserJid,
		SendText:  message,
		SendTime:  time.Now(),
		ReqId:     msgId,
		MsgType:   1,
		Read:      1,
		Initiator: a.UserJid,
	})
	a.callbackChan.TextMsgCallbackChan <- textMsgCallbacks
}

// 获取头像回调
func (a *Account) getUserHeadImageCallback(account string, picture []byte) {
	userImageCallbacks := make([]callback.GetUserHeadImageCallback, 0)

	userImage := callback.GetUserHeadImageCallback{
		Account: account,
		Picture: picture,
	}
	userImageCallbacks = append(userImageCallbacks, userImage)
	a.callbackChan.GetUserHeadImageCallbackChan <- userImageCallbacks
}

func (a *Account) available() {
	availableMsg := a.signal.createAvailableMessage()
	a.send(availableMsg)
}

func (a *Account) ping(t string) {
	requestId := fmt.Sprintf("%s-%d", t, a.id)
	a.id++
	pingMsg := a.signal.createPingMessage(requestId)
	a.send(pingMsg)
}

func (a *Account) send(msg *protobuf.XmppStanzaElement) {
	level.Info(a.logger).Log("msg", "Send message", "msg", msg, "user", a.ClientPayLoad.Username)
	a.Stream.Send(msg)
}

func (m *Manager) recordConnection(re bool) {
	if re == true {
		key := fmt.Sprintf("/service/%v/%v/%v/%v/%v/%v", m.env, m.ns, "count", m.appName, m.vs, m.ListenAddress)
		//记录连接数
		putValue := &protobuf.ServiceInfo{}
		getRes := &protobuf.ServiceInfo{}
		getReq := etcd.GetReq{
			Key:     key,
			Options: nil,
			ResChan: make(chan etcd.GetRes),
		}
		m.getChan <- getReq
		getResult := <-getReq.ResChan
		if getResult.Result.Kvs != nil {
			proto.Unmarshal(getResult.Result.Kvs[0].Value, getRes)
			tmp := &protobuf.ServiceInfo{
				IP:          m.ListenAddress,
				Connections: getRes.Connections + 1,
			}
			putValue = tmp
		}

		putvalueb, _ := proto.Marshal(putValue)

		putReq := etcd.PutReq{
			Key:     key,
			Value:   string(putvalueb),
			Options: nil,
			ResChan: make(chan etcd.PutRes),
		}
		m.putChan <- putReq
	} else {
		key := fmt.Sprintf("/service/%v/%v/%v/%v/%v/%v", m.env, m.ns, "count", m.appName, m.vs, m.ListenAddress)
		//记录连接数
		putValue := &protobuf.ServiceInfo{}
		getRes := &protobuf.ServiceInfo{}
		getReq := etcd.GetReq{
			Key:     key,
			Options: nil,
			ResChan: make(chan etcd.GetRes),
		}
		m.getChan <- getReq
		getResult := <-getReq.ResChan
		if getResult.Result.Kvs != nil {
			proto.Unmarshal(getResult.Result.Kvs[0].Value, getRes)
			if getRes.Connections > 0 {
				tmp := &protobuf.ServiceInfo{
					IP:          m.ListenAddress,
					Connections: getRes.Connections - 1,
				}
				putValue = tmp
			}

		}

		putvalueb, _ := proto.Marshal(putValue)

		putReq := etcd.PutReq{
			Key:     key,
			Value:   string(putvalueb),
			Options: nil,
			ResChan: make(chan etcd.PutRes),
		}
		m.putChan <- putReq
	}

}
