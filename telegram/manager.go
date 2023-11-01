package telegram

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/telegram/kv"
	"arthas/util"
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gotd/td/telegram"
	gopool "github.com/panjf2000/ants/v2"
	"strconv"
	"sync"
	"time"
)

type Options struct {
	GrpcServer string
}

type Manager struct {
	logger        log.Logger
	context       context.Context
	grpcServer    string
	ListenAddress string

	accountDetailChan          chan AccountDetail
	loginDetailChan            chan LoginDetail
	syncContactChan            chan SyncContact
	sendMessageChan            chan SendMessageDetail
	sendPhotoChan              chan SendImageFileDetail
	sendFileChan               chan SendFileDetail
	sendVideoChan              chan SendVideoDetail
	createGroupChan            chan CreateGroupDetail
	addGroupMemberChan         chan AddGroupMemberDetail
	getGroupMembersChan        chan GetGroupMembers
	sendGroupMessageChan       chan SendGroupMessageDetail
	senContactCardChan         chan ContactCardDetail
	sendCodeChan               chan SendCodeDetail
	getContactListChan         chan ContactListDetail
	getDialogListChan          chan DialogListDetail
	getHistoryDetailChan       chan GetHistoryDetail
	tgLogoutChan               chan TgLogoutDetail
	getDownloadChan            chan DownChatFileDetail
	createChannelDetailChan    chan CreateChannelDetail
	inviteToChannelDetailChan  chan InviteToChannelDetail
	importTgSessionMsgChan     chan ImportTgSessionMsg
	JoinByLinkDetailChan       chan JoinByLinkDetail
	GetEmojiGroupsDetailChan   chan GetEmojiGroupsDetail
	MessageReactionDetailChan  chan MessageReactionDetail
	LeaveDetailChan            chan LeaveDetail
	GetChannelMemberDetailChan chan GetChannelMemberDetail
	GetOnlineAccountDetailChan chan GetOnlineAccountDetail

	loginIsSync        *sync.Map
	clientMap          *sync.Map
	mutex              sync.Mutex
	accountsSync       *sync.Map
	getChan            chan etcd.GetReq
	txnInsertChan      chan etcd.TxnInsertReq
	txnDelayInsertChan chan etcd.TxnDelayInsertReq //延时插入etcd通道
	putChan            chan etcd.PutReq
	whatChan           chan etcd.WatchReq
	txnUpdateChan      chan etcd.TxnUpdateReq
	leaseChan          chan etcd.LeaseReq
	kvd                kv.KV
	callbackChan       *callback.CallbackChan

	sessionChan chan etcd.SessionReq
	deleteChan  chan etcd.DeleteReq

	ns      string
	env     string
	vs      string
	appName string

	goPool *gopool.Pool
}

func New(logger log.Logger, ctx context.Context, o *Options) *Manager {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	goPool, _ := gopool.NewPool(100)
	m := &Manager{
		context:                    ctx,
		logger:                     logger,
		grpcServer:                 o.GrpcServer,
		accountDetailChan:          make(chan AccountDetail),
		loginDetailChan:            make(chan LoginDetail),
		syncContactChan:            make(chan SyncContact),
		sendMessageChan:            make(chan SendMessageDetail),
		sendPhotoChan:              make(chan SendImageFileDetail),
		sendFileChan:               make(chan SendFileDetail),
		sendVideoChan:              make(chan SendVideoDetail),
		createGroupChan:            make(chan CreateGroupDetail),
		addGroupMemberChan:         make(chan AddGroupMemberDetail),
		getGroupMembersChan:        make(chan GetGroupMembers),
		sendGroupMessageChan:       make(chan SendGroupMessageDetail),
		senContactCardChan:         make(chan ContactCardDetail),
		sendCodeChan:               make(chan SendCodeDetail),
		getContactListChan:         make(chan ContactListDetail),
		getDialogListChan:          make(chan DialogListDetail),
		getHistoryDetailChan:       make(chan GetHistoryDetail),
		tgLogoutChan:               make(chan TgLogoutDetail),
		getDownloadChan:            make(chan DownChatFileDetail),
		createChannelDetailChan:    make(chan CreateChannelDetail),
		inviteToChannelDetailChan:  make(chan InviteToChannelDetail),
		importTgSessionMsgChan:     make(chan ImportTgSessionMsg),
		JoinByLinkDetailChan:       make(chan JoinByLinkDetail),
		GetEmojiGroupsDetailChan:   make(chan GetEmojiGroupsDetail),
		MessageReactionDetailChan:  make(chan MessageReactionDetail),
		LeaveDetailChan:            make(chan LeaveDetail),
		GetChannelMemberDetailChan: make(chan GetChannelMemberDetail),
		GetOnlineAccountDetailChan: make(chan GetOnlineAccountDetail),
		accountsSync:               &sync.Map{},
		loginIsSync:                &sync.Map{},
		clientMap:                  &sync.Map{},
		goPool:                     goPool,
		sessionChan:                make(chan etcd.SessionReq),
		deleteChan:                 make(chan etcd.DeleteReq),
	}
	return m
}

func (m *Manager) ApplyConfig(txnDelayInsertChan chan etcd.TxnDelayInsertReq, getChan chan etcd.GetReq, insertChan chan etcd.TxnInsertReq, putChan chan etcd.PutReq, delChan chan etcd.DeleteReq, updateChan chan etcd.TxnUpdateReq, leaseChan chan etcd.LeaseReq, address string, watchChan chan etcd.WatchReq, callbackChan *callback.CallbackChan, env string, space string, version string, appName string, sessionChan chan etcd.SessionReq, deleteChan chan etcd.DeleteReq) error {
	m.txnDelayInsertChan = txnDelayInsertChan
	m.getChan = getChan
	m.txnInsertChan = insertChan
	m.putChan = putChan
	m.txnUpdateChan = updateChan
	m.leaseChan = leaseChan
	m.whatChan = watchChan
	endpoints := util.CalculateListenedEndpoints(address)
	m.ListenAddress = endpoints[0].String()
	kvd, _ := kv.NewEtcd(kv.EtcdOptions{
		Ctx:     m.context,
		GetChan: getChan,
		PutChan: putChan,
		DelChan: delChan,
	})
	m.kvd = kvd
	m.callbackChan = callbackChan

	m.ns = space
	m.env = env
	m.vs = version
	m.appName = appName
	m.sessionChan = sessionChan
	m.deleteChan = deleteChan
	return nil
}

func (m *Manager) Run() error {
	level.Info(m.logger).Log("msg", "Running to telegram manager.")
	for {
		select {
		case accountDetail := <-m.accountDetailChan:
			m.goPool.Submit(func() {
				m.syncAccountDetail(accountDetail)
			})
		case loginDetail := <-m.loginDetailChan:
			_, ok := m.clientMap.Load(loginDetail.PhoneNumber)
			if !ok {
				// 没登陆则登录
				m.goPool.Submit(func() {
					m.loginWithCode(loginDetail)
				})
			} else {
				tgUserLoginCallback := make([]callback.TgUserLoginCallback, 0)
				msgCallback := callback.TgUserLoginCallback{
					IsSuccess:     1,
					AccountStatus: 0,
					IsOnline:      1,
					Phone:         loginDetail.PhoneNumber,
					Comment:       "Already logged in ",
				}
				tgUserLoginCallback = append(tgUserLoginCallback, msgCallback)
				res := LoginDetailRes{
					LoginStatus: protobuf.AccountStatus_SUCCESS,
					Comment:     "Already logged in ",
				}
				loginDetail.ResChan <- res
			}
		case logoutDetail := <-m.tgLogoutChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(logoutDetail.UserJid, 10))
			if ok {
				m.goPool.Submit(func() {
					logouErr := m.TgLogout(context.Background(), value.(*telegram.Client), logoutDetail)
					if logouErr != nil {
						tgLogoutCallback := callback.TgUserLogoutCallback{
							IsSuccess:     2,
							Phone:         strconv.FormatUint(logoutDetail.UserJid, 10),
							AccountStatus: 2,
							ProxyAddress:  logoutDetail.ProxyUrl,
							Comment:       logouErr.Error(),
						}
						m.logoutMsgCallback(tgLogoutCallback)
						logoutDetail.ResChan <- LogoutRes{LogoutStatus: protobuf.AccountStatus_LOGOUT_FAIL, Comment: strconv.FormatUint(logoutDetail.UserJid, 10) + " user logout fail ..."}
					}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", logoutDetail.UserJid)
				logoutDetail.ResChan <- LogoutRes{LogoutStatus: protobuf.AccountStatus_FAIL}
			}
		case syncContactDetail := <-m.syncContactChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(syncContactDetail.Account, 10))
			if ok {
				m.goPool.Submit(func() {
					err := m.AddContact(context.Background(), value.(*telegram.Client), syncContactDetail)
					if err != nil {
						syncContactDetail.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						syncContactDetail.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
					}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", syncContactDetail.Account)
			}

		case sendMessageDetail := <-m.sendMessageChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(sendMessageDetail.Sender, 10))
			if ok {
				m.goPool.Submit(func() {
					sendMsgErr := m.SendMessage(context.Background(), value.(*telegram.Client), sendMessageDetail)
					if sendMsgErr != nil {
						tgTextMsgCallbacks := make([]callback.TgMsgCallback, 0)
						tgTextMsgCallbacks = append(tgTextMsgCallbacks, callback.TgMsgCallback{
							Sender:      sendMessageDetail.Sender,
							Receiver:    sendMessageDetail.Receiver,
							SendStatus:  2,
							SendTime:    time.Now(),
							Initiator:   sendMessageDetail.Sender,
							Comment:     sendMsgErr.Error(),
							Read:        2,
							Out:         1,
							AccountType: 2,
						})
						if len(tgTextMsgCallbacks) > 0 {
							m.sendTgMsgCallback(tgTextMsgCallbacks)
						}
					}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", sendMessageDetail.Sender)
			}
		case sendPhotoDetail := <-m.sendPhotoChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(sendPhotoDetail.Sender, 10))
			if ok {
				m.goPool.Submit(func() {
					sendMsgErr := m.SendPhoto(m.context, value.(*telegram.Client), sendPhotoDetail)
					if sendMsgErr != nil {
						tgFileMsgErrCallbacks := make([]callback.TgMsgCallback, 0)
						tgFileMsgErrCallbacks = append(tgFileMsgErrCallbacks, callback.TgMsgCallback{
							Sender:      sendPhotoDetail.Sender,
							Receiver:    sendPhotoDetail.Receiver,
							SendStatus:  2,
							SendTime:    time.Now(),
							Initiator:   sendPhotoDetail.Sender,
							ReqId:       strconv.Itoa(sendMsgErr.ReqId),
							SendMsg:     sendMsgErr.SendMsg,
							Comment:     sendMsgErr.Err.Error(),
							Read:        2,
							AccountType: 2,
						})
						if len(tgFileMsgErrCallbacks) > 0 {
							m.sendTgMsgCallback(tgFileMsgErrCallbacks)
						}
						//sendPhotoDetail.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: sendMsgErr.Err.Error()}
					}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", sendPhotoDetail.Sender)
			}
		case sendFileDetail := <-m.sendFileChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(sendFileDetail.Sender, 10))
			if ok {
				m.goPool.Submit(func() {
					sendMsgErr := m.SendFile(m.context, value.(*telegram.Client), sendFileDetail)
					if sendMsgErr != nil {
						tgFileMsgErrCallbacks := make([]callback.TgMsgCallback, 0)
						tgFileMsgErrCallbacks = append(tgFileMsgErrCallbacks, callback.TgMsgCallback{
							Sender:      sendFileDetail.Sender,
							Receiver:    sendFileDetail.Receiver,
							SendStatus:  2,
							SendTime:    time.Now(),
							Initiator:   sendFileDetail.Sender,
							ReqId:       strconv.Itoa(sendMsgErr.ReqId),
							SendMsg:     sendMsgErr.SendMsg,
							Comment:     sendMsgErr.Err.Error(),
							Read:        2,
							AccountType: 2,
						})
						if len(tgFileMsgErrCallbacks) > 0 {
							m.sendTgMsgCallback(tgFileMsgErrCallbacks)
						}
						//sendFileDetail.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: sendMsgErr.Err.Error()}
					}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", sendFileDetail.Sender)
			}
		case sendVideoDetail := <-m.sendVideoChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(sendVideoDetail.Sender, 10))
			if ok {
				m.goPool.Submit(func() {
					sendMsgErr := m.SendVideo(m.context, value.(*telegram.Client), sendVideoDetail)
					if sendMsgErr != nil {
						tgFileMsgErrCallbacks := make([]callback.TgMsgCallback, 0)
						tgFileMsgErrCallbacks = append(tgFileMsgErrCallbacks, callback.TgMsgCallback{
							Sender:      sendVideoDetail.Sender,
							Receiver:    sendVideoDetail.Receiver,
							SendStatus:  2,
							SendTime:    time.Now(),
							Initiator:   sendVideoDetail.Sender,
							Comment:     sendMsgErr.Error(),
							Read:        2,
							AccountType: 2,
						})
						if len(tgFileMsgErrCallbacks) > 0 {
							m.sendTgMsgCallback(tgFileMsgErrCallbacks)
						}
						//sendVideoDetail.ResChan <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: sendMsgErr.Error()}
					}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", sendVideoDetail.Sender)
			}

		case createGroupDetail := <-m.createGroupChan:

			value, ok := m.clientMap.Load(strconv.FormatUint(createGroupDetail.Initiator, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &createGroupDetail
			if ok {

				m.goPool.Submit(func() {
					result, err := m.CreateGroup(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}

				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", createGroupDetail.Initiator)
			}

		case addGroupMemberDetail := <-m.addGroupMemberChan:

			value, ok := m.clientMap.Load(strconv.FormatUint(addGroupMemberDetail.Initiator, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &addGroupMemberDetail
			if ok {

				m.goPool.Submit(func() {
					result, err := m.AddGroupMember(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}

				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", addGroupMemberDetail.Initiator)
			}
		case getGroupMembersChan := <-m.getGroupMembersChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(getGroupMembersChan.Initiator, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &getGroupMembersChan
			if ok {
				m.goPool.Submit(func() {
					result, err := m.GetGroupMembers(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}

				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", getGroupMembersChan.Initiator)
			}
		case sendGroupMessageDetail := <-m.sendGroupMessageChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(sendGroupMessageDetail.Sender, 10))
			if ok {
				m.goPool.Submit(func() {
					m.SendMessageInGroup(m.context, value.(*telegram.Client), sendGroupMessageDetail)
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", sendGroupMessageDetail.Sender)
			}

		case sendContactCardDetail := <-m.senContactCardChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(sendContactCardDetail.Sender, 10))
			if ok {
				m.goPool.Submit(func() {
					sendMsgErr := m.SendContactCard(m.context, value.(*telegram.Client), sendContactCardDetail)
					if sendMsgErr != nil {
						tgCardMsgCallbacks := make([]callback.TgCardMsgCallback, 0)
						tgCardMsgCallbacks = append(tgCardMsgCallbacks, callback.TgCardMsgCallback{
							Sender:     sendContactCardDetail.Sender,
							Receiver:   sendContactCardDetail.Receiver,
							SendStatus: 2,
							SendTime:   time.Now(),
							Initiator:  sendContactCardDetail.Sender,
							ReqId:      strconv.Itoa(sendMsgErr.ReqId),
							SendMsg:    sendMsgErr.SendMsg,
							Comment:    sendMsgErr.Err.Error(),
							Read:       2,
						})
						if len(tgCardMsgCallbacks) > 0 {
							m.sendTgCardMsgCallback(tgCardMsgCallbacks)
						}
					}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", sendContactCardDetail.Sender)
			}
		case contactListDetail := <-m.getContactListChan:

			value, ok := m.clientMap.Load(strconv.FormatUint(contactListDetail.Account, 10))
			if ok {
				m.goPool.Submit(func() {
					m.GetContactList(m.context, value.(*telegram.Client), contactListDetail)
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", contactListDetail.Account)
			}
		case dialogListDetail := <-m.getDialogListChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(dialogListDetail.Account, 10))
			if ok {
				m.goPool.Submit(func() {
					m.GetDialogList(m.context, value.(*telegram.Client), dialogListDetail)
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", dialogListDetail.Account)
			}
		case getHistoryDetail := <-m.getHistoryDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(getHistoryDetail.Self, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &getHistoryDetail
			if ok {

				m.goPool.Submit(func() {
					result, err := m.GetHistory(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}

				})

			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", getHistoryDetail.Self)
			}
		case getDownLoadFile := <-m.getDownloadChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(getDownLoadFile.Account, 10))
			if ok {
				thatClient := value.(*telegram.Client)
				_ = grpool.AddWithRecover(m.context, func(ctx context.Context) {
					_ = m.getDownloadFile(m.context, thatClient, getDownLoadFile)
				}, func(ctx context.Context, err error) {
					getDownLoadFile.ResChan <- DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}}
				})
			} else {
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", getDownLoadFile.Account)
				getDownLoadFile.ResChan <- DownChatFileRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "send protobuf msg is nil or having err "}}
			}
		case createChannelDetail := <-m.createChannelDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(createChannelDetail.Creator, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &createChannelDetail
			if ok {

				m.goPool.Submit(func() {
					result, err := m.CreateChannel(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}
				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", createChannelDetail.Creator)
			}
		case inviteToChannelDetail := <-m.inviteToChannelDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(inviteToChannelDetail.Invite, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &inviteToChannelDetail
			if ok {

				m.goPool.Submit(func() {
					result, err := m.InviteToChannel(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}
				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", inviteToChannelDetail.Invite)
			}
		case tgImportSessionMsg := <-m.importTgSessionMsgChan:

			_ = grpool.AddWithRecover(m.context, func(ctx context.Context) {
				_ = m.ImportTgSessionToEtcd(m.context, tgImportSessionMsg)
			}, func(ctx context.Context, err error) {
			})

		case joinByLinkDetail := <-m.JoinByLinkDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(joinByLinkDetail.Sender, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &joinByLinkDetail
			if ok {

				m.goPool.Submit(func() {
					result, err := m.JoinPublicChannelByLink(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}
				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", joinByLinkDetail.Sender)
			}
		case getEmojiGroupsDetail := <-m.GetEmojiGroupsDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(getEmojiGroupsDetail.Sender, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &getEmojiGroupsDetail
			if ok {

				m.goPool.Submit(func() {
					result, err := m.GetEmojiGroups(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}
				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", getEmojiGroupsDetail.Sender)
			}
		case messageReactionDetail := <-m.MessageReactionDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(messageReactionDetail.Sender, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &messageReactionDetail
			if ok {
				m.goPool.Submit(func() {
					result, err := m.MessageReaction(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}
				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", messageReactionDetail.Sender)
			}
		case leaveDetail := <-m.LeaveDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(leaveDetail.Sender, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &leaveDetail
			if ok {
				m.goPool.Submit(func() {
					result, err := m.Leave(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}
				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", leaveDetail.Sender)
			}
		case getChannelMemberDetail := <-m.GetChannelMemberDetailChan:
			value, ok := m.clientMap.Load(strconv.FormatUint(getChannelMemberDetail.Sender, 10))
			thatClient := value.(*telegram.Client)
			thatReq := &getChannelMemberDetail
			if ok {
				m.goPool.Submit(func() {
					result, err := m.GetChannelMember(m.context, thatClient, thatReq)
					if err != nil {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
					} else {
						thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
					}
				})
			} else {
				thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "no login"}
				level.Info(m.logger).Log("msg", "User not logged in", "phone:", getChannelMemberDetail.Sender)
			}
		case getOnlineAccountDetail := <-m.GetOnlineAccountDetailChan:
			thatReq := &getOnlineAccountDetail
			m.goPool.Submit(func() {
				result, err := m.GetOnlineAccounts(m.context, thatReq)
				if err != nil {
					thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: err.Error()}
				} else {
					thatReq.Res <- &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, Data: gjson.MustEncode(result)}
				}
			})
		case <-m.context.Done():
			return nil
		}
	}
}
