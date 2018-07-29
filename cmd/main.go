package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"crawshaw.io/littleboss"
	"github.com/golang/glog"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/flags"

	"github.com/GeertJohan/go.rice"
	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/service"
)

var (
	//raft
	bind                 string
	join                 string
	dir                  string
	id                   string
	defaultDwell         uint64
	defaultDwellDeadline uint64
	defaultMaxDwell      uint64

	// build
	version = "dev"
	commit  = "none"
	date    = "unknown"

	cfg *config.Config
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: \n\n ./cortex -stderrthreshold=INFO -log_dir=$(pwd) -id=node1 lb=start & \n "+
		"./cortex-new-version -stderrthreshold=INFO  -log_dir=$(pwd) -id=node1 -lb=reload \n ./cortex -lb=stop \n \n")
	flag.PrintDefaults()

	os.Exit(2)
}

func init() {
	flag.Usage = usage
	cfg = &config.Config{
		NodeID:               "",
		Dir:                  "./data",
		JoinAddr:             "",
		DefaultDwell:         3 * 60 * 1000,   // 3 minutes
		DefaultMaxDwell:      6 * 60 * 1000,   // 6 minutes
		DefaultDwellDeadline: 2.5 * 60 * 1000, // 2.5 minutes
		MaxHistory:           1000,
		FlushInterval:        1000,
		SnapshotInterval:     30,
	}
}

func main() {

	lb := littleboss.New("cortex")
	lb.Command("service", flag.String("service", "start", "littleboss start command"))
	flagRaft := lb.Listener("raft", "tcp", ":4444", "-raft :4444")
	flagHTTP := lb.Listener("http", "tcp", ":4445", "-http :4445")

	box := rice.MustFindBox("build")

	glog.Infof("Boxing the build folder - %s", box.Name())

	loader := confita.NewLoader(flags.NewBackend())
	err := loader.Load(context.Background(), cfg)
	if err != nil {
		glog.Infof("%v\n", err)
		usage()
	}

	glog.Infof("raft addr %v, http addr %v\n", flagRaft.String(), flagHTTP.String())

	lb.Run(func(ctx context.Context) {
		run(context.Background(), flagRaft, flagHTTP)
	})

	glog.Info("cortex exited")
}

func run(ctx context.Context, flagRaft, flagHTTP *littleboss.ListenerFlag) {

	cfg.HTTPAddr = flagHTTP.String()
	cfg.RaftAddr = flagRaft.String()
	cfg.HTTPListener = flagHTTP.Listener()
	cfg.RaftListener = flagRaft.Listener()
	cfg.EnableFileServer = true

	svc, err := service.New(cfg)
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}

	go func() {
		if err := svc.Start(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			glog.Fatal(err)
		}
	}()

	<-ctx.Done()
	svc.Shutdown(ctx)

}
