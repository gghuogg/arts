package main

import (
	"arthas/callback"
	"arthas/callback/consumer"
	callbackFactory "arthas/callback/factory"
	"arthas/callback/grpc_stream"
	"arthas/callback/kafka"
	"arthas/etcd"
	alog "arthas/log"
	"arthas/rpc"
	_ "arthas/rpc/pr"
	"arthas/version"
	"context"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

var (
	appName = "server_manager"
)

type flagConfig struct {
	configFile string
	rpc        rpc.Options
	etcd       etcd.Options
	logConfig  alog.Config
	callback   callback.Options
	consumer   callback.Options
}

func main() {

	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}

	cfg := flagConfig{}

	a := kingpin.New(filepath.Base(os.Args[0]), "The Arthas Manager server").UsageWriter(os.Stdout)

	a.HelpFlag.Short('h')

	a.Flag("config.file", "Arthas configuration file path.").
		Default("arthas.yml").StringVar(&cfg.configFile)

	a.Flag("web.listen-address", "Address to listen on for rpc.").
		Default("0.0.0.0:50052").StringVar(&cfg.rpc.ListenAddress)
	a.Flag("web.name-space", "namespace for service.").
		Default("kk").StringVar(&cfg.rpc.NameSpace)
	a.Flag("web.env", "env for service.").
		Default("kk").StringVar(&cfg.rpc.Env)
	a.Flag("web.version", "version for service.").
		Default("latest").StringVar(&cfg.rpc.Version)
	a.Flag("web.max-connections", "Maximum number of simultaneous connections.").
		Default("512").IntVar(&cfg.rpc.MaxConnections)
	a.Flag("etcd.endpoint", "Connect etcd server address.").
		Default("10.8.5.21:2379").StringsVar(&cfg.etcd.EndPoint)
	//Default("127.0.0.1:2379").StringsVar(&cfg.etcd.EndPoint)

	a.Flag("etcd.username", "Connect etcd server username.").
		Default("").StringVar(&cfg.etcd.Username)

	a.Flag("etcd.password", "Connect etcd server password.").
		Default("").StringVar(&cfg.etcd.Password)

	a.Flag("etcd.delay", "Insert etcd delay.").
		Default("10s").DurationVar(&cfg.etcd.Delay)

	a.Flag("callback.type", "Callback type.").
		Default(kafka.Type).StringVar(&cfg.callback.Type)

	a.Flag("callback.server.address", "Callback server address.").
		Default("101.35.51.132:30092").StringsVar(&cfg.callback.Addrs)

	a.Flag("callback.server.username", "Callback server username.").
		Default().StringVar(&cfg.callback.Username)

	a.Flag("callback.server.password", "Callback server password.").
		Default().StringVar(&cfg.callback.Password)

	a.Flag("consumer.server.gradeId", "Consumer server gradeId.").
		Default("grade").StringVar(&cfg.consumer.GradeId)
	a.Flag("consumer.type", "Consumer type.").
		Default(consumer.Type).StringVar(&cfg.consumer.Type)
	a.Flag("consumer.server.address", "Consumer server address.").
		Default("101.35.51.132:30092").StringsVar(&cfg.consumer.Addrs)
	a.Flag("consumer.server.username", "Consumer server username.").
		Default().StringVar(&cfg.consumer.Username)
	a.Flag("consumer.server.password", "Consumer server password.").
		Default().StringVar(&cfg.callback.Password)

	a.Version(version.Info())

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Error parsing command line arguments: %w", err))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	if cfg.callback.Type == grpc_stream.Type {
		cfg.callback.Addrs = []string{cfg.rpc.ListenAddress}
	}

	logger := alog.New(&cfg.logConfig)

	level.Info(logger).Log("msg", "Starting "+appName, "version", version.Info())

	var (
		ctxRpc, cancelRpc           = context.WithCancel(context.Background())
		ctxEtcd, cancelEtcd         = context.WithCancel(context.Background())
		ctxCallback, cancelCallback = context.WithCancel(context.Background())
		ctxConsumer, _              = context.WithCancel(context.Background())
	)

	cfg.rpc.Context = ctxRpc

	rpcHandler := rpc.New(log.With(logger, "component", "rpc"), &cfg.rpc)
	etcdManager := etcd.New(log.With(logger, "component", "etcd"), ctxEtcd, &cfg.etcd)
	callbackManager := callbackFactory.New(log.With(logger, "component", "callback"), ctxCallback, &cfg.callback)
	consumerManager := callbackFactory.New(log.With(logger, "component", "consumer"), ctxConsumer, &cfg.consumer)
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
				return rpcHandler.ApplyConfigProxy(
					etcdManager.GetChan(),
					etcdManager.PutChan(),
					etcdManager.TxnInsertChan(),
					etcdManager.TxnUpdateChan(),
					etcdManager.TxnDelayInsertChan(),
					etcdManager.LeaseChan(),
					callbackManager.GetManager().CallbackChan(),
					consumerManager.GetManager().ConsumeChan(),
					etcdManager.WatchChan(),
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
			name: "consumer",
			reloader: func(config *flagConfig) error {
				return consumerManager.ApplyConfig()
			},
		},
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
					close(rpcHandler.ExitChan)
					reloadReady.Close()
				case <-cancel:
					close(rpcHandler.ExitChan)
					reloadReady.Close()
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
		g.Add(
			func() error {
				err := rpcHandler.RunProxy(ctxRpc, listener)
				level.Info(logger).Log("msg", "Rpc handler stopped")
				return err
			},
			func(err error) {
				level.Info(logger).Log("msg", "Stopping rpc handler...")
				cancelRpc()
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
				level.Info(logger).Log("msg", "Arthas proxy is ready to receive tcp.")

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
