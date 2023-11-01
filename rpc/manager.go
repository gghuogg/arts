package rpc

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/signal"
	"arthas/telegram"
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gopool "github.com/panjf2000/ants/v2"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Options struct {
	Context        context.Context
	ListenAddress  string
	MaxConnections int
	NameSpace      string
	Env            string
	Version        string
}

type Handler struct {
	logger  log.Logger
	context context.Context
	options *Options
	server  *grpc.Server

	//whatsapp
	accountDetailChan    chan signal.AccountDetail
	loginDetailChan      chan signal.LoginDetail
	logoutDetailChan     chan signal.LogoutDetail
	syncContactChan      chan signal.SyncContact
	sendMessageChan      chan signal.SendMessageDetail
	sendVCardMessageChan chan signal.SendVcardMessageDetail
	GetUserHeadImageChan chan signal.GetUserHeadImageDetail
	accounts             map[uint64]*signal.Account //readonly
	accountsSync         *sync.Map
	sendImageFileChan    chan signal.SendImageFileDetail
	sendFileChan         chan signal.SendFileDetail
	sendVideoChan        chan signal.SendVideoFileDetail

	//telegram
	tgAccountsSync          *sync.Map
	tgLoginCodeSync         sync.Map
	tgAccountDetailChan     chan telegram.AccountDetail
	tgLoginDetailChan       chan telegram.LoginDetail
	tgSyncContactChan       chan telegram.SyncContact
	tgSendMessageChan       chan telegram.SendMessageDetail
	tgSendPhotoChan         chan telegram.SendImageFileDetail
	tgSendVideoChan         chan telegram.SendVideoDetail
	tgSendFileChan          chan telegram.SendFileDetail
	tgCreateGroupChan       chan telegram.CreateGroupDetail
	tgGetGroupMembers       chan telegram.GetGroupMembers
	tgAddGroupMemberChan    chan telegram.AddGroupMemberDetail
	tgSendGroupMessageChan  chan telegram.SendGroupMessageDetail
	tgSendContactCardChan   chan telegram.ContactCardDetail
	tgSendCodeChan          chan telegram.SendCodeDetail
	tgGetContactListChan    chan telegram.ContactListDetail
	tgGetDialogListChan     chan telegram.DialogListDetail
	tgGetMsgHistoryChan     chan telegram.GetHistoryDetail
	tgLogoutChan            chan telegram.TgLogoutDetail
	tgDownloadFileChan      chan telegram.DownChatFileDetail
	tgCreateChannelChan     chan telegram.CreateChannelDetail
	tgInviteToChannelChan   chan telegram.InviteToChannelDetail
	tgImportSessionMsgChan  chan telegram.ImportTgSessionMsg
	tgJoinByLinkChan        chan telegram.JoinByLinkDetail
	tgGetEmojiGroupsChan    chan telegram.GetEmojiGroupsDetail
	tgMessagesReactionChan  chan telegram.MessageReactionDetail
	tgLeaveChan             chan telegram.LeaveDetail
	tgGetChannelMemberChan  chan telegram.GetChannelMemberDetail
	tgGetOnlineAccountsChan chan telegram.GetOnlineAccountDetail

	getChan            chan etcd.GetReq
	txnInsertChan      chan etcd.TxnInsertReq
	txnDelayInsertChan chan etcd.TxnDelayInsertReq //延时插入etcd通道
	putChan            chan etcd.PutReq
	txnUpdateChan      chan etcd.TxnUpdateReq
	leaseChan          chan etcd.LeaseReq
	watchChan          chan etcd.WatchReq
	sessionChan        chan etcd.SessionReq
	deleteChan         chan etcd.DeleteReq

	reloader          chan chan error
	mtx               sync.RWMutex
	ready             atomic.Uint32
	blockedAccountSet map[uint64]struct{}
	callbackType      string
	callbackChan      *callback.CallbackChan
	consumerChan      *callback.ConsumerChan

	NameSpace string
	Env       string
	Version   string

	ExitChan    chan struct{}
	ExitChanArt chan struct{}

	StreamMap          sync.Map
	serverMap          sync.Map
	finalResponse      chan *protobuf.ResponseMessage
	ticker             *time.Ticker
	cancelServerStream context.CancelFunc
}

func New(logger log.Logger, o *Options) *Handler {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	h := &Handler{
		logger:          logger,
		options:         o,
		reloader:        make(chan chan error),
		context:         o.Context,
		server:          grpc.NewServer(),
		ExitChan:        make(chan struct{}),
		ExitChanArt:     make(chan struct{}),
		StreamMap:       sync.Map{},
		serverMap:       sync.Map{},
		tgLoginCodeSync: sync.Map{},
		ticker:          time.NewTicker(30 * time.Second),
		finalResponse:   make(chan *protobuf.ResponseMessage),
	}
	h.SetReady(false)
	return h
}

func (h *Handler) Listener() (net.Listener, error) {
	level.Info(h.logger).Log("msg", "Start listening for connections", "address", h.options.ListenAddress)

	listener, err := net.Listen("tcp", h.options.ListenAddress)
	if err != nil {
		return listener, err
	}
	listener = netutil.LimitListener(listener, h.options.MaxConnections)

	return listener, nil
}

func (h *Handler) SetReady(v bool) {
	if v {
		h.ready.Store(1)
		return
	}
	h.ready.Store(0)
}

func (h *Handler) ApplyConfigProxy(get chan etcd.GetReq, put chan etcd.PutReq, insert chan etcd.TxnInsertReq, update chan etcd.TxnUpdateReq, dinsert chan etcd.TxnDelayInsertReq, leaseChan chan etcd.LeaseReq, callbackChan *callback.CallbackChan, consumerChan *callback.ConsumerChan, watchChan chan etcd.WatchReq, sessionChan chan etcd.SessionReq, deleteChan chan etcd.DeleteReq) error {
	h.getChan = get
	h.putChan = put
	h.txnInsertChan = insert
	h.txnUpdateChan = update
	h.txnDelayInsertChan = dinsert
	h.leaseChan = leaseChan
	h.callbackChan = callbackChan
	h.consumerChan = consumerChan
	h.watchChan = watchChan
	h.NameSpace = h.options.NameSpace
	h.Version = h.options.Version
	h.Env = h.options.Env
	h.sessionChan = sessionChan
	h.deleteChan = deleteChan

	h.RegisterServiceToEtcd(h.options.Env, h.options.NameSpace, h.options.Version, "", h.options.ListenAddress, "arts_manager", "")
	h.watchProxy(h.options.Env, h.options.NameSpace)

	go h.telegramWatcher()

	return nil
}

func (h *Handler) ApplyConfig(signalAccount map[uint64]*signal.Account, signalAccountSync *sync.Map, signalBlockAccountSet map[uint64]struct{}, signalAccountDetail chan signal.AccountDetail, signalLoginDetail chan signal.LoginDetail, signalSyncContact chan signal.SyncContact, messageChan chan signal.SendMessageDetail, callbackType string, callbackChan *callback.CallbackChan, getChan chan etcd.GetReq, putChan chan etcd.PutReq, updateChan chan etcd.TxnUpdateReq, insertChan chan etcd.TxnInsertReq, delayInsertChan chan etcd.TxnDelayInsertReq, leaseChan chan etcd.LeaseReq, watchChan chan etcd.WatchReq, sendVCardMessageChan chan signal.SendVcardMessageDetail, signalLogoutDetail chan signal.LogoutDetail, getUserHeadImageDetail chan signal.GetUserHeadImageDetail, tgAccountsSync *sync.Map, tgAccountDetailChan chan telegram.AccountDetail, tgLoginDetailChan chan telegram.LoginDetail, tgContactChan chan telegram.SyncContact, tgSendMessageChan chan telegram.SendMessageDetail, appName string, createGroupChan chan telegram.CreateGroupDetail, addGroupMemberChan chan telegram.AddGroupMemberDetail, getGroupMembersChan chan telegram.GetGroupMembers, consumerChan *callback.ConsumerChan, tgSendPhotoChan chan telegram.SendImageFileDetail, tgSendFileChan chan telegram.SendFileDetail, sendGroupMessageChan chan telegram.SendGroupMessageDetail, sendContactCardChan chan telegram.ContactCardDetail, sendImageFileChan chan signal.SendImageFileDetail, sendFileChan chan signal.SendFileDetail, sendVideoChan chan signal.SendVideoFileDetail, tgSendCodeChan chan telegram.SendCodeDetail, tgGetContactListChan chan telegram.ContactListDetail, tgGetDialogListChan chan telegram.DialogListDetail, tgMsgHistoryChan chan telegram.GetHistoryDetail, tgLogoutChan chan telegram.TgLogoutDetail, tgSendVideoChan chan telegram.SendVideoDetail, channelChan chan telegram.CreateChannelDetail, fileChan chan telegram.DownChatFileDetail, inviteToChannelChan chan telegram.InviteToChannelDetail, joinLinkChan chan telegram.JoinByLinkDetail, getEmojiGroupChan chan telegram.GetEmojiGroupsDetail, messagesReactionChan chan telegram.MessageReactionDetail, tgImportSessionChan chan telegram.ImportTgSessionMsg, leaveChan chan telegram.LeaveDetail, getChannelMemberChan chan telegram.GetChannelMemberDetail, getOnlineAccountsChan chan telegram.GetOnlineAccountDetail, sessionChan chan etcd.SessionReq, deleteChan chan etcd.DeleteReq) error {

	h.accounts = signalAccount
	h.accountDetailChan = signalAccountDetail
	h.loginDetailChan = signalLoginDetail
	h.syncContactChan = signalSyncContact
	h.accountsSync = signalAccountSync
	h.blockedAccountSet = signalBlockAccountSet
	h.sendMessageChan = messageChan
	h.callbackType = callbackType
	h.callbackChan = callbackChan
	h.consumerChan = consumerChan
	h.getChan = getChan
	h.putChan = putChan
	h.txnInsertChan = insertChan
	h.txnUpdateChan = updateChan
	h.txnDelayInsertChan = delayInsertChan
	h.leaseChan = leaseChan
	h.watchChan = watchChan
	h.logoutDetailChan = signalLogoutDetail
	h.GetUserHeadImageChan = getUserHeadImageDetail
	h.sendVCardMessageChan = sendVCardMessageChan
	h.sendImageFileChan = sendImageFileChan
	h.sendFileChan = sendFileChan
	h.sendVideoChan = sendVideoChan
	h.sessionChan = sessionChan
	h.deleteChan = deleteChan

	//telegram
	h.tgAccountsSync = tgAccountsSync
	h.tgAccountDetailChan = tgAccountDetailChan
	h.tgLoginDetailChan = tgLoginDetailChan
	h.tgSyncContactChan = tgContactChan
	h.tgSendMessageChan = tgSendMessageChan
	h.tgSendPhotoChan = tgSendPhotoChan
	h.tgSendVideoChan = tgSendVideoChan
	h.tgSendFileChan = tgSendFileChan
	h.tgCreateGroupChan = createGroupChan
	h.tgGetGroupMembers = getGroupMembersChan
	h.tgAddGroupMemberChan = addGroupMemberChan
	h.tgSendGroupMessageChan = sendGroupMessageChan
	h.tgSendCodeChan = tgSendCodeChan
	h.tgSendContactCardChan = sendContactCardChan
	h.tgGetContactListChan = tgGetContactListChan
	h.tgGetDialogListChan = tgGetDialogListChan
	h.tgGetMsgHistoryChan = tgMsgHistoryChan
	h.tgLogoutChan = tgLogoutChan
	h.tgCreateChannelChan = channelChan
	h.tgDownloadFileChan = fileChan
	h.tgInviteToChannelChan = inviteToChannelChan
	h.tgImportSessionMsgChan = tgImportSessionChan
	h.tgJoinByLinkChan = joinLinkChan
	h.tgGetEmojiGroupsChan = getEmojiGroupChan
	h.tgMessagesReactionChan = messagesReactionChan
	h.tgLeaveChan = leaveChan
	h.tgGetChannelMemberChan = getChannelMemberChan
	h.tgGetOnlineAccountsChan = getOnlineAccountsChan

	h.record(h.options.Env, h.options.NameSpace, h.options.Version, h.options.ListenAddress, appName)
	h.RegisterServiceToEtcd(h.options.Env, h.options.NameSpace, h.options.Version, "arthas", h.options.ListenAddress, "", appName)
	h.watchService(h.options.Env, h.options.NameSpace)

	return nil
}

func (h *Handler) RunProxy(ctx context.Context, listener net.Listener) error {
	level.Info(h.logger).Log("msg", "Running to proxy grpc manager.")
	if listener == nil {
		var err error
		listener, err = h.Listener()
		if err != nil {
			return err
		}
	}

	goPool, _ := gopool.NewPool(250)
	srvProxy := ProxyServer{logger: h.logger, Handler: h, gp: goPool}

	protobuf.RegisterArthasServer(h.server, &srvProxy)

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.server.Serve(listener)
	}()

	for {
		select {
		case e := <-errCh:
			return e
		//case <-time.After(time.Second * 10):
		//	fmt.Println(h.accounts)
		case <-ctx.Done():
			return nil
		}
	}
}

func (h *Handler) Run(ctx context.Context, listener net.Listener) error {
	level.Info(h.logger).Log("msg", "Running to grpc manager.")
	if listener == nil {
		var err error
		listener, err = h.Listener()
		if err != nil {
			return err
		}
	}

	srv := server{logger: h.logger, Handler: h, ticker: time.NewTicker(30 * time.Second)}
	protobuf.RegisterArthasStreamServer(h.server, srv)

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.server.Serve(listener)
	}()

	for {
		select {
		case e := <-errCh:
			return e

		case <-ctx.Done():
			return nil
		}
	}
}
