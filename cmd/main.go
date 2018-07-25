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
		RaftBindPort:         8878,
		Dir:                  "./data",
		JoinAddr:             "",
		DefaultDwell:         3 * 60 * 1000,   // 3 minutes
		DefaultMaxDwell:      6 * 60 * 1000,   // 6 minutes
		DefaultDwellDeadline: 2.5 * 60 * 1000, // 2.5 minutes
		MaxHistory:           1000,
	}
}

func main() {

	loader := confita.NewLoader(flags.NewBackend())
	lb := littleboss.New("cortex")
	lb.Command("lb", flag.String("lb", "start", "littleboss start command"))

	flag.Parse()

	err := loader.Load(context.Background(), cfg)
	if err != nil {
		fmt.Printf("%v\n", err)
		usage()
	}

	svc, err := service.New(cfg)
	if err != nil {
		glog.Fatal(err)
	}

	lb.Run(func(ctx context.Context) {
		run(ctx, svc)
	})

	glog.Info("cortex exited")
}

func run(ctx context.Context, svc *service.Service) {

	go func() {
		if err := svc.HTTP().ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			glog.Fatal(err)
		}
	}()

	<-ctx.Done()
	svc.Shutdown(ctx)

}
