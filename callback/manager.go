package callback

import (
	"arthas/protobuf"
	"context"
	"github.com/go-kit/log"
	"github.com/gogf/gf/v2/os/gtime"
	"time"
)

type (
	IManager struct {
		logger       log.Logger
		context      context.Context
		callbackChan *CallbackChan
		consumerChan *ConsumerChan
		options      *Options
		callbackType string
		sendFunc     func(topic string, message interface{}) (partition int32, offset int64, err error)
	}
	Options struct {
		Type     string
		Addrs    []string
		Username string
		Password string
		OpMap    map[string]interface{}
		GradeId  string
		Group    string
	}
	ConsumerChan struct {
		SendRequestMsgToProxyChan chan SendRequestMsgToProxy
	}
	CallbackChan struct {
		LogoutCallbackChan           chan []LogoutCallback
		LoginCallbackChan            chan []LoginCallback
		TextMsgCallbackChan          chan []TextMsgCallback
		ReadMsgCallbackChan          chan []ReadMsgCallback
		SyncContactCallbackChan      chan []SyncContactCallback
		VcardMsgCallbackChan         chan []VcardMsgCallback
		PushRequestMsgCallbackChan   chan []PushRequestMsgCallbackChan
		GetUserHeadImageCallbackChan chan []GetUserHeadImageCallback
		ImCallbackChan               chan ImCallBackChan

		TgLoginCallbackChan chan []TgLoginCallbackChan
	}
	LoginCallback struct {
		UserJid     uint64                 `json:"userJid"`
		LoginStatus protobuf.AccountStatus `json:"loginStatus"`
		ProxyUrl    string                 `json:"proxyUrl"`
		Comment     string                 `json:"comment"`
	}
	SyncContactCallback struct {
		AccountDb uint64 // 发送人账号
		Status    string // 同步状态 in/out
		Synchro   string // 同步的联系人
	}
	LogoutCallback struct {
		UserJid    uint64 `json:"userJid"`
		Proxy      string `json:"proxy"`
		LogoutType int    `json:"logoutType"`
	}
	TextMsgCallback struct {
		Sender    uint64    `json:"sender"`    //发送人
		Receiver  uint64    `json:"receiver"`  //接收人
		SendText  string    `json:"sendText"`  //消息内容
		SendTime  time.Time `json:"sendTime"`  //发送时间
		ReqId     string    `json:"reqId"`     //请求ID
		Read      int       `json:"read"`      //是否已读
		MsgType   int       `json:"msgType"`   //消息类型
		Initiator uint64    `json:"initiator"` //发起人
	}
	TgMsgCallback struct {
		Initiator     uint64    `json:"initiator"     description:"聊天发起人"`
		Sender        uint64    `json:"sender"        description:"发送人"`
		Receiver      string    `json:"receiver"      description:"接收人"`
		ReqId         string    `json:"reqId"         description:"请求id"`
		SendMsg       []byte    `json:"sendMsg"       description:"发送消息原文(加密)"`
		TranslatedMsg []byte    `json:"translatedMsg" description:"发送消息译文(加密)"`
		MsgType       int       `json:"msgType"       description:"消息类型"`
		SendTime      time.Time `json:"sendTime"      description:"发送时间"`
		Read          int       `json:"read"          description:"是否已读"`
		Comment       string    `json:"comment"       description:"备注"`
		SendStatus    int       `json:"sendStatus"    description:"发送状态"`
		FileName      string    `json:"fileName"      description:"文件名称"`
		FileSize      int64     `json:"fileSize"` //文件大小
		FileType      string    `json:"fileType"` //文件类型
		Md5           string    `json:"md5"`
		Out           int       `json:"out"           description:"自己发出"`
		AccountType   int       `json:"accountType"   description:"账号类型1:id,2:phone,3:username"`
	}

	TgCardMsgCallback struct {
		Initiator     uint64    `json:"initiator"     description:"聊天发起人"`
		Sender        uint64    `json:"sender"        description:"发送人"`
		Receiver      string    `json:"receiver"      description:"接收人"`
		ReqId         string    `json:"reqId"         description:"请求id"`
		CardFirstName string    `json:"cardFirstName" description:"卡片联系人的头称"`
		CardLastName  string    `json:"cardLastName"  description:"卡片联系人的尾称"`
		CardPhone     string    `json:"cardPhone"     description:"卡片联系人的手机号"`
		MsgType       int       `json:"msgType"       description:"消息类型"`
		SendTime      time.Time `json:"sendTime"      description:"发送时间"`
		SendMsg       []byte    `json:"sendMsg"       description:"发送内容"`
		Read          int       `json:"read"          description:"是否已读"`
		Comment       string    `json:"comment"       description:"备注"`
		SendStatus    int       `json:"sendStatus"    description:"发送状态"`
		Out           int       `json:"out"           description:"自己发出"`
	}
	TgUserLoginCallback struct {
		IsSuccess     int    `json:"isSuccess"     description:"是否登录成功"`
		TgId          int64  `json:"tgId"          description:"tg id"`
		Username      string `json:"username"      description:"账号号码"`
		FirstName     string `json:"firstName"     description:"First Name"`
		LastName      string `json:"lastName"      description:"Last Name"`
		Phone         string `json:"phone"         description:"手机号"`
		Photo         string `json:"photo"         description:"账号头像"`
		AccountStatus int    `json:"accountStatus" description:"账号状态"`
		IsOnline      int    `json:"isOnline"      description:"是否在线"`
		ProxyAddress  string `json:"proxyAddress"  description:"代理地址"`
		Comment       string `json:"comment"       description:"备注"`
		Session       []byte `json:"session"       description:"session"`
	}

	TgUserLogoutCallback struct {
		IsSuccess     int    `json:"isSuccess"     description:"是否登录成功"`
		TgId          int64  `json:"tgId"          description:"tg id"`
		Username      string `json:"username"      description:"账号号码"`
		FirstName     string `json:"firstName"     description:"First Name"`
		LastName      string `json:"lastName"      description:"Last Name"`
		Phone         string `json:"phone"         description:"手机号"`
		Photo         string `json:"photo"         description:"账号头像"`
		AccountStatus int    `json:"accountStatus" description:"账号状态"`
		IsOnline      int    `json:"isOnline"      description:"是否在线"`
		ProxyAddress  string `json:"proxyAddress"  description:"代理地址"`
		Comment       string `json:"comment"       description:"备注"`
		Session       []byte `json:"session"       description:"session"`
	}
	TgContactsCallback struct {
		TgId      int64  `json:"tgId"      description:"tg id"`
		Username  string `json:"username"  description:"username"`
		FirstName string `json:"firstName" description:"First Name"`
		LastName  string `json:"lastName"  description:"Last Name"`
		Phone     string `json:"phone"     description:"phone"`
		Photo     string `json:"photo"     description:"photo"`
		Type      int    `json:"type"      description:"type"`
		Comment   string `json:"comment"   description:"comment"`
	}
	TgDialogsCallback struct {
		Comment string `json:"comment"   description:"comment"`
	}
	TgReadMsgCallback struct {
		Sender   uint64 `json:"sender"   description:"sender"`
		Receiver uint64 `json:"receiver" description:"receiver"`
		ReqId    string `json:"reqId"    description:"reqId"`
		Comment  string `json:"comment"   description:"comment"`
	}

	TgGetMsgHistoryCallBack struct {
		Msg       string
		MsgId     int
		MsgFromId int64
		Out       bool
		PeerId    int64
		Media     any
		Data      int
	}
	GroupMsgCallback struct {
		Sender    uint64    `json:"sender"`    //发送人
		Title     string    `json:"title"`     //接收人
		SendText  string    `json:"sendText"`  //消息内容
		SendTime  time.Time `json:"sendTime"`  //发送时间
		ReqId     string    `json:"reqId"`     //请求ID
		Read      int       `json:"read"`      //是否已读
		MsgType   int       `json:"msgType"`   //消息类型
		Initiator uint64    `json:"initiator"` //发起人
	}
	ArtsServerDown struct {
		Ip      string `json:"ip"`
		Message string `json:"message"`
	}

	CreateGroupCallback struct {
		ChatId int64
		Title  string
	}

	AddGroupMemberCallback struct {
		MemberId  int64
		UserName  string
		FirstName string
		LastName  string
		Phone     string
	}

	ReceiverCallback struct {
		Msg       string
		MsgId     int
		MsgFromId int64
		Out       bool
		PeerId    int64
		Date      *gtime.Time
	}

	ReadMsgCallback struct {
		ReqId string `json:"reqId"` //请求ID
	}
	VcardMsgCallback struct {
		Sender   uint64 //发送人
		Receiver uint64 //接收人
		SendMsg  string //名片内容
		SendTime int64  //发送时间
		MsgId    string //消息id
	}
	ImCallBackChan struct {
		Topic    string     `json:"topic"` // topic
		Callback ImCallback `json:"callback"`
	}
	ImCallback struct {
		Type int    //类型
		Data []byte // data
	}

	PushRequestMsgCallbackChan struct {
		RequestMessage []byte
		Action         string
		UniqueId       string
		Type           string
	}
	GetUserHeadImageCallback struct {
		Account string
		Picture []byte
	}
	TgLoginCallbackChan struct {
		LoginType int
		Type      string
	}

	SendRequestMsgToProxy struct {
		RequestMsg *protobuf.RequestMessage
	}

	TgCreateChannel struct {
		ChannelId       int64  `json:"channelId"`
		AccessHash      int64  `json:"accessHash"`
		ChannelTitle    string `json:"channelTitle"`
		ChannelUserName string `json:"channelUserName"`
		Link            string `json:"link"`
		ErrInfo         string `json:"errInfo"`
	}

	TgGetEmojiGroups struct {
		Title       string   `json:"title"`
		IconEmojiID int64    `json:"iconEmojiID"`
		Emoticons   []string `json:"emoticons"`
		ErrInfo     string   `json:"errInfo"`
	}

	TgInviteChannel           struct{}
	TgJoinPublicChannelByLink struct {
		TgId       int64  `json:"tgId"`
		AccessHash int64  `json:"accessHash"`
		Title      string `json:"title"`
		UserName   string `json:"userName"`
	}
	TgAddGroupMember   struct{}
	TgCreateGroup      struct{}
	TgMessagesReaction struct{}
	TgLeave            struct{}
	TgGetChannelMember struct {
		TgId      int64  `json:"tgId"          description:"tg id"`
		Username  string `json:"username"      description:"账号号码"`
		FirstName string `json:"firstName"     description:"First Name"`
		LastName  string `json:"lastName"      description:"Last Name"`
		Phone     string `json:"phone"         description:"手机号"`
		Comment   string `json:"comment"       description:"备注"`
	}
	TgGetGroupMember struct {
		TgId      int64  `json:"tgId"          description:"tg id"`
		Username  string `json:"username"      description:"账号号码"`
		FirstName string `json:"firstName"     description:"First Name"`
		LastName  string `json:"lastName"      description:"Last Name"`
		Phone     string `json:"phone"         description:"手机号"`
		Comment   string `json:"comment"       description:"备注"`
	}
	GetServerOnlineAccount struct {
		TgId      int64  `json:"tgId"          description:"tg id"`
		Username  string `json:"username"      description:"账号号码"`
		FirstName string `json:"firstName"     description:"First Name"`
		LastName  string `json:"lastName"      description:"Last Name"`
		Phone     string `json:"phone"         description:"手机号"`
	}
)

func (m *IManager) CallbackChan() *CallbackChan {
	return m.callbackChan
}

func (m *IManager) ConsumeChan() *ConsumerChan {
	return m.consumerChan
}

func (m *IManager) CallbackType() string {
	return m.callbackType
}

func (m *IManager) SetSendFunc(sendFunc func(topic string, message interface{}) (partition int32, offset int64, err error)) {
	m.sendFunc = sendFunc
}

type ICallback interface {
	Run() error
	ApplyConfig() (err error)
	GetManager() *IManager
}

func DefaultManager(logger log.Logger, ctx context.Context, o *Options) *IManager {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	iManager := &IManager{
		logger:       logger,
		context:      ctx,
		options:      o,
		callbackType: o.Type,
		callbackChan: &CallbackChan{
			LoginCallbackChan:            make(chan []LoginCallback),
			LogoutCallbackChan:           make(chan []LogoutCallback),
			TextMsgCallbackChan:          make(chan []TextMsgCallback),
			ReadMsgCallbackChan:          make(chan []ReadMsgCallback),
			SyncContactCallbackChan:      make(chan []SyncContactCallback),
			PushRequestMsgCallbackChan:   make(chan []PushRequestMsgCallbackChan),
			GetUserHeadImageCallbackChan: make(chan []GetUserHeadImageCallback),
			ImCallbackChan:               make(chan ImCallBackChan),
		},
		consumerChan: &ConsumerChan{
			SendRequestMsgToProxyChan: make(chan SendRequestMsgToProxy, 10000),
		},
	}
	return iManager
}

func (m *IManager) GetLogger() log.Logger {
	return m.logger
}

func (m *IManager) GetCtx() context.Context {
	return m.context
}

func (m *IManager) GetOptions() *Options {
	return m.options
}

func (m *IManager) Run() error {
	for {
		if m.callbackType == "stream" {
			continue
		}
		select {
		case callbackMsg := <-m.CallbackChan().LoginCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("login", callbackMsg)
			}()
		case callbackMsg := <-m.CallbackChan().LogoutCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("logout", callbackMsg)
			}()
		case callbackMsg := <-m.CallbackChan().TextMsgCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("textMsg", callbackMsg)
			}()
		case callbackMsg := <-m.CallbackChan().ReadMsgCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("read", callbackMsg)
			}()
		case callbackMsg := <-m.CallbackChan().SyncContactCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("syncContact", callbackMsg)
			}()
		case callbackMsg := <-m.CallbackChan().VcardMsgCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("vcardMsg", callbackMsg)
			}()
		case pushChanMsg := <-m.CallbackChan().PushRequestMsgCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("RequestMessage", pushChanMsg)
			}()
		case callbackMsg := <-m.CallbackChan().ImCallbackChan:
			go func() {
				_, _, _ = m.sendFunc(callbackMsg.Topic, callbackMsg.Callback)
			}()
		case getHeadImage := <-m.CallbackChan().GetUserHeadImageCallbackChan:
			go func() {
				_, _, _ = m.sendFunc("GetUserImage", getHeadImage)
			}()

		case <-m.context.Done():
			return nil
		}
	}
}
