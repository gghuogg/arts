package main

import (
	"arthas/callback"
	"arthas/callback/consts"
	callbackFactory "arthas/callback/factory"
	"arthas/callback/grpc_stream"
	"arthas/callback/kafka"
	"arthas/conn"
	"arthas/etcd"
	alog "arthas/log"
	"arthas/prometheus"
	"arthas/proxy"
	"arthas/rpc"
	asignal "arthas/signal"
	"arthas/storage"
	"arthas/telegram"
	"arthas/version"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "arthas/rpc/im"
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
)

var (
	appName = "arthas"
)

type flagConfig struct {
	configFile string
	rpc        rpc.Options
	signal     asignal.Options
	etcd       etcd.Options
	telegram   telegram.Options
	callback   callback.Options
	logConfig  alog.Config
	prometheus prometheus.Options
	proxy      proxy.Options
	conn       conn.Options
	storage    storage.Options
}

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}
	cfg := flagConfig{}

	a := kingpin.New(filepath.Base(os.Args[0]), "The Arthas libsignal server").UsageWriter(os.Stdout).Terminate(nil)
	a.HelpFlag.Short('h')

	a.Flag("config.file", "Arthas configuration file path.").
		Default("arthas.yml").StringVar(&cfg.configFile)
	a.Flag("web.name-space", "namespace for service.").
		Default("ghxkk").StringVar(&cfg.rpc.NameSpace)
	a.Flag("web.env", "env for service.").
		Default("ghxkk").StringVar(&cfg.rpc.Env)
	a.Flag("web.version", "version for service.").
		Default("latest").StringVar(&cfg.rpc.Version)

	a.Flag("web.max-connections", "Maximum number of simultaneous connections.").
		Default("1024").IntVar(&cfg.rpc.MaxConnections)

	//10.8.5.21:2379
	a.Flag("etcd.endpoint", "Connect etcd server address.").
		Default("192.168.73.92:2379").StringsVar(&cfg.etcd.EndPoint)
	//Default("127.0.0.1:2379").StringsVar(&cfg.etcd.EndPoint)

	a.Flag("etcd.username", "Connect etcd server username.").
		Default("").StringVar(&cfg.etcd.Username)

	a.Flag("etcd.password", "Connect etcd server password.").
		Default("").StringVar(&cfg.etcd.Password)

	a.Flag("etcd.delay", "Insert etcd delay.").
		Default("10s").DurationVar(&cfg.etcd.Delay)

	a.Flag("proxy.deadline", "Proxy client deadline time").
		Default("30s").DurationVar(&cfg.proxy.Deadline)

	a.Flag("callback.type", "Callback type.").
		Default(kafka.Type).StringVar(&cfg.callback.Type)

	a.Flag("callback.server.address", "Callback server address.").
		Default("192.168.73.92:9092").StringsVar(&cfg.callback.Addrs)

	a.Flag("callback.server.username", "Callback server username.").
		Default().StringVar(&cfg.callback.Username)

	a.Flag("callback.server.password", "Callback server password.").
		Default().StringVar(&cfg.callback.Password)

	a.Flag("prometheus.Address", "Prometheus server Address.").
		Default("http://192.168.73.92:9090").StringVar(&cfg.prometheus.Address)

	a.Flag("prometheus.Username", "Prometheus server Username.").
		Default().StringVar(&cfg.prometheus.Username)

	a.Flag("prometheus.Password", "Prometheus server Password.").
		Default().StringVar(&cfg.prometheus.Password)

	a.Flag("prometheus.Port", "Prometheus server Port.").
		Default().StringVar(&cfg.prometheus.Port)

	a.Version(version.Info())

	//whatsapp
	wa := a.Command("whatsapp", "whatsapp server start")

	wa.Flag("signal.grpc-server", "Grpc server address to dial.").
		Default("207.148.68.250:3334").StringVar(&cfg.signal.GrpcServer) // 端口3333
	wa.Flag("web.listen-address", "Address to listen on for rpc.").
		Default("0.0.0.0:31599").StringVar(&cfg.rpc.ListenAddress)

	//telegram
	tg := a.Command("telegram", "telegram server start")
	tg.Flag("web.listen-address", "Address to listen on for rpc.").
		Default("0.0.0.0:31500").StringVar(&cfg.rpc.ListenAddress)

	//telegram
	//a.Flag("td.grpc-server", "Grpc server address to dial.").
	//	Default("207.148.68.250:3335").StringVar(&cfg.telegram.GrpcServer)
	var typeName string

	command, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Error parsing command line arguments: %w", err))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	switch command {
	case tg.FullCommand():
		typeName = tg.FullCommand()
	case wa.FullCommand():
		typeName = wa.FullCommand()
	}

	if cfg.callback.Type == grpc_stream.Type {
		cfg.callback.Addrs = []string{cfg.rpc.ListenAddress}
	}

	logger := alog.New(&cfg.logConfig)

	level.Info(logger).Log("msg", "Starting "+appName, "version", version.Info())

	var (
		ctxRpc, cancelRpc               = context.WithCancel(context.Background())
		ctxSignal, cancelSignal         = context.WithCancel(context.Background())
		ctxTelegram, cancelTelegram     = context.WithCancel(context.Background())
		ctxEtcd, cancelEtcd             = context.WithCancel(context.Background())
		ctxCallback, cancelCallback     = context.WithCancel(context.Background())
		ctxPrometheus, cancelPrometheus = context.WithCancel(context.Background())
		ctxConn, cancelConn             = context.WithCancel(context.Background())
		ctxProxy, cancelProxy           = context.WithCancel(context.Background())
		ctxStorage, cancelStorage       = context.WithCancel(context.Background())
	)

	cfg.rpc.Context = ctxRpc

	etcdManager := etcd.New(log.With(logger, "component", "etcd"), ctxEtcd, &cfg.etcd)
	signalManager := asignal.New(log.With(logger, "component", "signal"), ctxSignal, &cfg.signal)
	tgManager := telegram.New(log.With(logger, "component", "telegram"), ctxTelegram, &cfg.telegram)
	rpcHandler := rpc.New(log.With(logger, "component", "rpc"), &cfg.rpc)

	callbackManager := callbackFactory.New(log.With(logger, "component", "callback"), ctxCallback, &cfg.callback)
	prometheusManager := prometheus.New(log.With(logger, "component", "prometheus"), ctxPrometheus, &cfg.prometheus)
	proxyManager := proxy.New(log.With(logger, "component", "proxy"), ctxProxy, &cfg.proxy)
	connManager := conn.New(log.With(logger, "component", "conn"), ctxConn, &cfg.conn)
	storageManager := storage.New(log.With(logger, "component", "storage"), ctxStorage, &cfg.storage)

	reloaders := []reloader{
		{
			name: "etcd",
			reloader: func(config *flagConfig) error {
				return etcdManager.ApplyConfig()
			},
		},
		{
			name: "rpc",
			reloader: func(config *flagConfig) error {
				return rpcHandler.ApplyConfig(
					signalManager.Accounts(),
					signalManager.AccountsSync(),
					signalManager.BlockAccountSet(),
					signalManager.AccountDetailChan(),
					signalManager.LoginDetailChan(),
					signalManager.SyncContactChan(),
					signalManager.SendMessageChan(),
					callbackManager.GetManager().CallbackType(),
					callbackManager.GetManager().CallbackChan(),
					etcdManager.GetChan(),
					etcdManager.PutChan(),
					etcdManager.TxnUpdateChan(),
					etcdManager.TxnInsertChan(),
					etcdManager.TxnDelayInsertChan(),
					etcdManager.LeaseChan(),
					etcdManager.WatchChan(),
					signalManager.SendVCardMessageChan(),
					signalManager.LogoutDetailChan(),
					signalManager.SetUserHeadImageDetailChan(),
					tgManager.AccountsSync(),
					tgManager.AccountDetailChan(),
					tgManager.LoginDetailChan(),
					tgManager.SyncContactChan(),
					tgManager.SendMessageChan(),
					typeName,
					tgManager.CreateGroupChan(),
					tgManager.AddGroupMemberChan(),
					tgManager.GetGroupMembersChan(),
					callbackManager.GetManager().ConsumeChan(),
					tgManager.SendPhotoChan(),
					tgManager.SendFileChan(),
					tgManager.SendGroupMessageChan(),
					tgManager.SendContactCardChan(),
					signalManager.SendImageFileChan(),
					signalManager.SendFileChan(),
					signalManager.SendVideoChan(),
					tgManager.SendCodeChan(),
					tgManager.ContatcListChan(),
					tgManager.DialogListChan(),
					tgManager.GetMessageHistoryChan(),
					tgManager.TgLogoutChan(),
					tgManager.TgSendVideoChan(),
					tgManager.TgCreateChannelChan(),
					tgManager.TgGetDownloadFileChan(),
					tgManager.TgInviteToChannelChan(),
					tgManager.TgJoinByLinkChan(),
					tgManager.TgGetEmojiGroupChan(),
					tgManager.TgMessagesReactionChan(),
					tgManager.TgImportSessionChan(),
					tgManager.TgLeaveChan(),
					tgManager.TgGetChannelMemberChan(),
					tgManager.TgGetOnlineAccountsChan(),
					etcdManager.SessionChan(),
					etcdManager.DeleteChan(),
				)
			},
		},
		{
			name: "callback",
			reloader: func(config *flagConfig) error {
				return callbackManager.ApplyConfig()
			},
		},
		{
			name: "prometheus",
			reloader: func(config *flagConfig) error {
				return prometheusManager.ApplyConfig()
			},
		},
	}

	if typeName == "telegram" {
		tmp := reloader{
			name: "telegram",
			reloader: func(config *flagConfig) error {
				return tgManager.ApplyConfig(etcdManager.TxnDelayInsertChan(),
					etcdManager.GetChan(),
					etcdManager.TxnInsertChan(),
					etcdManager.PutChan(),
					etcdManager.DeleteChan(),
					etcdManager.TxnUpdateChan(),
					etcdManager.LeaseChan(),
					cfg.rpc.ListenAddress,
					etcdManager.WatchChan(),
					callbackManager.GetManager().CallbackChan(),
					cfg.rpc.Env,
					cfg.rpc.NameSpace,
					cfg.rpc.Version,
					typeName,
					etcdManager.SessionChan(),
					etcdManager.DeleteChan(),
				)
			},
		}
		reloaders = append(reloaders, tmp)
	} else if typeName == "whatsapp" {
		tmp := reloader{
			name: "signal",
			reloader: func(config *flagConfig) error {
				return signalManager.ApplyConfig(etcdManager.TxnDelayInsertChan(),
					etcdManager.GetChan(),
					etcdManager.TxnInsertChan(),
					etcdManager.PutChan(),
					etcdManager.TxnUpdateChan(),
					etcdManager.LeaseChan(),
					callbackManager.GetManager().CallbackChan(),
					cfg.rpc.ListenAddress,
					etcdManager.WatchChan(),
					cfg.rpc.Env,
					cfg.rpc.NameSpace,
					cfg.rpc.Version,
					typeName,
				)

			},
		}
		reloaders = append(reloaders, tmp)
	}

	listener, err := rpcHandler.Listener()
	if err != nil {
		level.Error(logger).Log("msg", "Unable to start web listener", "err", err)
		os.Exit(1)
	}

	type closeOnce struct {
		C     chan struct{}
		once  sync.Once
		Close func()
	}
	reloadReady := &closeOnce{
		C: make(chan struct{}),
	}
	reloadReady.Close = func() {
		reloadReady.once.Do(func() {
			close(reloadReady.C)
		})
	}

	var g run.Group
	{
		term := make(chan os.Signal, 1)
		signal.Notify(term, os.Interrupt, syscall.SIGTERM)
		cancel := make(chan struct{})
		g.Add(
			func() error {
				select {
				case <-term:
					level.Warn(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
					etcdManager.GetChan()
					artsServerDownCallback(cfg.rpc.Env, cfg.rpc.NameSpace, etcdManager.GetChan(), callbackManager.GetManager().CallbackChan().ImCallbackChan)
					reloadReady.Close()
					close(rpcHandler.ExitChanArt)
				case <-cancel:
					artsServerDownCallback(cfg.rpc.Env, cfg.rpc.NameSpace, etcdManager.GetChan(), callbackManager.GetManager().CallbackChan().ImCallbackChan)
					reloadReady.Close()
					close(rpcHandler.ExitChanArt)
				}
				return nil
			},
			func(err error) {
				close(cancel)
				rpcHandler.SetReady(false)
			},
		)
	}
	{
		g.Add(
			func() error {
				err := etcdManager.Run()
				level.Info(logger).Log("msg", "Etcd manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping etcd manager...")
				cancelEtcd()
			},
		)
	}
	{
		g.Add(
			func() error {
				err := rpcHandler.Run(ctxRpc, listener)
				level.Info(logger).Log("msg", "Rpc handler stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping rpc handler...")
				cancelRpc()
			},
		)
	}

	if typeName == "whatsapp" {
		{
			g.Add(
				func() error {
					err := signalManager.Run()
					level.Info(logger).Log("msg", "Signal manager stopped")
					return err
				},
				func(err error) {
					level.Info(logger).Log("msg", "Stopping signal manager...")
					cancelSignal()
				},
			)
		}
	} else if typeName == "telegram" {
		{
			g.Add(
				func() error {
					err := tgManager.Run()
					level.Info(logger).Log("msg", "telegram manager stopped")
					return err
				},
				func(err error) {
					level.Info(logger).Log("msg", "Stopping telegram manager...")
					cancelTelegram()
				},
			)
		}
	}

	{
		g.Add(
			func() error {
				err := prometheusManager.Run()
				level.Info(logger).Log("msg", "Prometheus manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping prometheus manager...")
				cancelPrometheus()
			},
		)
	}
	{
		g.Add(
			func() error {
				err := connManager.Run()
				level.Info(logger).Log("msg", "Conn manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping conn manager...")
				cancelConn()
			},
		)
	}
	{
		g.Add(
			func() error {
				err := storageManager.Run()
				level.Info(logger).Log("msg", "Storage manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping storage manager...")
				cancelStorage()
			},
		)
	}
	{
		g.Add(
			func() error {
				err := proxyManager.Run()
				level.Info(logger).Log("msg", "Proxy manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping proxy manager...")
				cancelProxy()
			},
		)
	}
	{
		g.Add(
			func() error {
				err := callbackManager.Run()
				level.Info(logger).Log("msg", "Callback manager stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping callback manager...")
				cancelCallback()
			},
		)
	}
	{
		cancel := make(chan struct{})
		g.Add(
			func() error {
				if err := reloadConfig(&cfg, reloaders...); err != nil {
					return nil
				}
				reloadReady.Close()
				level.Info(logger).Log("msg", "Arthas server is ready to receive tcp.")
				<-cancel
				return nil
			},
			func(err error) {
				close(cancel)
			},
		)
	}

	if err := g.Run(); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "See you next time!")

}

type reloader struct {
	name     string
	reloader func(config *flagConfig) error
}

func reloadConfig(config *flagConfig, rls ...reloader) error {
	for _, rl := range rls {
		_ = rl.reloader(config)
	}
	return nil
}

func artsServerDownCallback(env string, ns string, getChan chan etcd.GetReq, imCallbackChan chan callback.ImCallBackChan) {
	get := etcd.GetReq{
		Key:     "/service/" + env + "/" + ns + "/arthas/",
		Options: []clientv3.OpOption{clientv3.WithPrefix()},
		ResChan: make(chan etcd.GetRes),
	}
	getChan <- get
	result := <-get.ResChan
	var ip string
	if result.Result.Kvs != nil {
		ip = strings.TrimSpace(string(result.Result.Kvs[0].Value))
	}
	artsServerDown := callback.ArtsServerDown{
		Message: "arts server is down...",
		Ip:      ip,
	}
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicArtsServerDown,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(artsServerDown),
		},
	}
	imCallbackChan <- backChan
	time.Sleep(500 * time.Millisecond)

}
