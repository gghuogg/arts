package signal

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/util"
	"context"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/os/gctx"
	gopool "github.com/panjf2000/ants/v2"
	grpcpool "github.com/processout/grpc-go-pool"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
)

const (
	headerProxyUrl         = "x-proxy-url"
	headerClientPayload    = "x-client-payload"
	headerPrivateKey       = "x-private-key"
	headerResumptionSecret = "x-resumption-secret"
)

type Options struct {
	GrpcServer string
}

type Manager struct {
	logger                  log.Logger
	context                 context.Context
	grpcServer              string
	ListenAddress           string
	ServiceId               uint64
	mutex                   sync.Mutex
	accountDetailChan       chan AccountDetail
	preKeyBundleChan        chan *PreKeyBundle
	loginDetailChan         chan LoginDetail
	syncContactChan         chan SyncContact
	sendMessageChan         chan SendMessageDetail
	accountsSync            *sync.Map //后续写
	accounts                map[uint64]*Account
	requests                map[string]*Request
	blockedAccountSet       map[uint64]struct{}
	getChan                 chan etcd.GetReq
	txnInsertChan           chan etcd.TxnInsertReq
	txnDelayInsertChan      chan etcd.TxnDelayInsertReq //延时插入etcd通道
	putChan                 chan etcd.PutReq
	whatChan                chan etcd.WatchReq
	txnUpdateChan           chan etcd.TxnUpdateReq
	leaseChan               chan etcd.LeaseReq
	callbackChan            *callback.CallbackChan
	logoutDetailChan        chan LogoutDetail
	sendVcardMessageChan    chan SendVcardMessageDetail
	GetUserImageDetailChan  chan GetUserHeadImageDetail
	SendImageFileDetailChan chan SendImageFileDetail
	SendFileDetailChan      chan SendFileDetail
	SendVideoDetailChan     chan SendVideoFileDetail
	appVersion              AppVersion //whatsApp 版本号
	ns                      string
	env                     string
	vs                      string
	appName                 string

	goPool   *gopool.Pool
	grpcPool *grpcpool.Pool
}

type SyncContact struct {
	Account  uint64
	Contacts []uint64
}

var (
	loginSuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_success_total",
			Help: "Total number of successful user logins",
		},
		[]string{"username"},
	)
	loginFailureCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_failure_total",
			Help: "Total number of failed user logins",
		},
		[]string{"username", "reason"},
	)
	accountBannedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_account_banned_total",
			Help: "Total number of banned user logins",
		},
		[]string{"username", "reason"},
	)
	loginProxyCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_success_proxy",
			Help: "Total number of success Proxy user logins",
		},
		[]string{"username"},
	)

	ctxgc = gctx.New()
)

func init() {
	prometheus.MustRegister(loginSuccessCounter)
	prometheus.MustRegister(loginFailureCounter)
}

func New(logger log.Logger, ctx context.Context, o *Options) *Manager {
	if logger == nil {
		logger = log.NewNopLogger()
	}

	goPool, _ := gopool.NewPool(250)

	factory := func() (*grpc.ClientConn, error) {
		return grpc.Dial(o.GrpcServer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	grpcPool, _ := grpcpool.New(factory, 50, 100, 20)
	m := &Manager{
		context:                 ctx,
		logger:                  logger,
		requests:                make(map[string]*Request),
		accounts:                make(map[uint64]*Account),
		grpcServer:              o.GrpcServer,
		accountDetailChan:       make(chan AccountDetail),
		loginDetailChan:         make(chan LoginDetail),
		syncContactChan:         make(chan SyncContact),
		sendMessageChan:         make(chan SendMessageDetail),
		accountsSync:            &sync.Map{},
		blockedAccountSet:       make(map[uint64]struct{}),
		getChan:                 make(chan etcd.GetReq),
		putChan:                 make(chan etcd.PutReq),
		whatChan:                make(chan etcd.WatchReq),
		txnInsertChan:           make(chan etcd.TxnInsertReq),
		txnDelayInsertChan:      make(chan etcd.TxnDelayInsertReq),
		txnUpdateChan:           make(chan etcd.TxnUpdateReq),
		logoutDetailChan:        make(chan LogoutDetail),
		sendVcardMessageChan:    make(chan SendVcardMessageDetail),
		GetUserImageDetailChan:  make(chan GetUserHeadImageDetail),
		SendImageFileDetailChan: make(chan SendImageFileDetail),
		SendFileDetailChan:      make(chan SendFileDetail),
		SendVideoDetailChan:     make(chan SendVideoFileDetail),
		goPool:                  goPool,
		grpcPool:                grpcPool,
	}
	return m
}

func (m *Manager) ApplyConfig(txnDelayInsertChan chan etcd.TxnDelayInsertReq, getChan chan etcd.GetReq, insertChan chan etcd.TxnInsertReq, putChan chan etcd.PutReq, updateChan chan etcd.TxnUpdateReq, leaseChan chan etcd.LeaseReq, callbackChan *callback.CallbackChan, address string, watchChan chan etcd.WatchReq, env string, space string, version string, appName string) error {
	m.txnDelayInsertChan = txnDelayInsertChan
	m.getChan = getChan
	m.txnInsertChan = insertChan
	m.putChan = putChan
	m.txnUpdateChan = updateChan
	m.leaseChan = leaseChan
	m.whatChan = watchChan
	m.callbackChan = callbackChan
	endpoints := util.CalculateListenedEndpoints(address)
	m.ListenAddress = endpoints[0].String()
	m.ns = space
	m.env = env
	m.vs = version
	m.appName = appName
	return nil
}

func (m *Manager) Run() error {
	level.Info(m.logger).Log("msg", "Running to signal manager.")

	// 监听版本号
	go m.MonitorAppVersion()

	for {
		select {
		case accountDetail := <-m.accountDetailChan:
			m.goPool.Submit(func() {
				m.syncAccountDetail(accountDetail)
			})

		case loginDetail := <-m.loginDetailChan:
			m.goPool.Submit(func() {
				m.loginSync(loginDetail)
			})

		case syncContact := <-m.syncContactChan:
			m.goPool.Submit(func() {
				m.syncContact(syncContact)
			})
		case sendMessage := <-m.sendMessageChan:
			value, _ := m.accountsSync.Load(sendMessage.Sender)
			account := value.(*Account)
			m.goPool.Submit(func() {
				account.SendTextMessage(sendMessage.Receiver, sendMessage.SendText)
			})
		case vcardMessage := <-m.sendVcardMessageChan:
			value, _ := m.accountsSync.Load(vcardMessage.Sender)
			account := value.(*Account)
			m.goPool.Submit(func() {
				account.SendVCardMessage(vcardMessage.Sender, vcardMessage)
			})
		case logoutDetail := <-m.logoutDetailChan:
			m.goPool.Submit(func() {
				m.logoutCompanionDevice(logoutDetail)
			})
		case userHeadImage := <-m.GetUserImageDetailChan:
			value, _ := m.accountsSync.Load(userHeadImage.Account)
			account := value.(*Account)
			m.goPool.Submit(func() {
				account.getUserHeadImage(userHeadImage.GetUserHeaderPhone)
			})
		case sendImageFileDetail := <-m.SendImageFileDetailChan:
			value, ok := m.accountsSync.Load(sendImageFileDetail.Sender)
			fmt.Println(ok)
			account := value.(*Account)
			m.goPool.Submit(func() {
				account.sendImageFile(sendImageFileDetail)

			})
		case sendFileDetail := <-m.SendFileDetailChan:
			value, _ := m.accountsSync.Load(sendFileDetail.Sender)
			account := value.(*Account)
			m.goPool.Submit(func() {
				account.sendFile(sendFileDetail)
			})
		case sendVideoDetail := <-m.SendVideoDetailChan:
			value, _ := m.accountsSync.Load(sendVideoDetail.Sender)
			account := value.(*Account)
			fmt.Println(account)
			m.goPool.Submit(func() {
				account.sendVideo(sendVideoDetail)
			})

		case <-m.context.Done():
			return nil
		}
	}
}
