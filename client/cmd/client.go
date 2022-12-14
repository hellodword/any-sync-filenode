package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/account"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/api"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/badgerprovider"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/clientspace"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/clientspace/clientcache"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/document"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/filestorage"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/client/storage"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/app"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/app/logger"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/commonfile/rpcstore"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/commonspace"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/config"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/metric"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/net/dialer"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/net/pool"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/net/rpc/server"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/net/secure"
	"github.com/anytypeio/go-anytype-infrastructure-experiments/common/nodeconf"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var log = logger.NewNamed("main")

var (
	flagConfigFile = flag.String("c", "etc/client.yml", "path to config file")
	// we can't use "v" here because of glog init (through badger) setting flag.Bool with "v"
	flagVersion = flag.Bool("ver", false, "show version and exit")
	flagHelp    = flag.Bool("h", false, "show help and exit")
)

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println(app.VersionDescription())
		return
	}
	if *flagHelp {
		flag.PrintDefaults()
		return
	}

	if debug, ok := os.LookupEnv("ANYPROF"); ok && debug != "" {
		go func() {
			http.ListenAndServe(debug, nil)
		}()
	}

	// create app
	ctx := context.Background()
	a := new(app.App)

	// open config file
	conf, err := config.NewFromFile(*flagConfigFile)
	if err != nil {
		log.Fatal("can't open config file", zap.Error(err))
	}

	// bootstrap components
	a.Register(conf)
	Bootstrap(a)

	// start app
	if err := a.Start(ctx); err != nil {
		log.Fatal("can't start app", zap.Error(err))
	}
	log.Info("app started", zap.String("version", a.Version()))

	// wait exit signal
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-exit
	log.Info("received exit signal, stop app...", zap.String("signal", fmt.Sprint(sig)))

	// close app
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if err := a.Close(ctx); err != nil {
		log.Fatal("close error", zap.Error(err))
	} else {
		log.Info("goodbye!")
	}
	time.Sleep(time.Second / 3)
}

func Bootstrap(a *app.App) {
	a.Register(account.New()).
		Register(nodeconf.New()).
		Register(metric.New()).
		Register(badgerprovider.New()).
		Register(storage.New()).
		Register(clientcache.New(200)).
		Register(secure.New()).
		Register(dialer.New()).
		Register(pool.New()).
		Register(commonspace.New()).
		Register(clientspace.New()).
		Register(server.New()).
		Register(document.New()).
		Register(rpcstore.New()).
		Register(filestorage.New()).
		Register(api.New())
}
