package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"crawshaw.io/littleboss"
	"github.com/golang/glog"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/flags"

	"github.com/myntra/aggo/pkg/store"
)

var (
	//raft
	bind                       string
	join                       string
	dir                        string
	id                         string
	defaultWaitWindow          uint64
	defaultWaitWindowThreshold uint64
	defaultMaxWaitWindow       uint64

	// build
	version = "dev"
	commit  = "none"
	date    = "unknown"

	config *store.Config
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: \n\n ./aggo -stderrthreshold=INFO -log_dir=$(pwd) -id=node1 lb=start & \n "+
		"./aggo-new-version -stderrthreshold=INFO  -log_dir=$(pwd) -id=node1 -lb=reload \n ./aggo -lb=stop \n \n")
	flag.PrintDefaults()

	os.Exit(2)
}

func init() {
	flag.Usage = usage
	config = &store.Config{
		NodeID:                     "",
		BindAddr:                   ":8878",
		Dir:                        "./data",
		JoinAddr:                   "",
		DefaultWaitWindow:          3 * 60 * 1000,   // 3 minutes
		DefaultMaxWaitWindow:       6 * 60 * 1000,   // 6 minutes
		DefaultWaitWindowThreshold: 2.5 * 60 * 1000, // 2.5 minutes
	}
}

func main() {

	loader := confita.NewLoader(flags.NewBackend())
	lb := littleboss.New("aggo")
	lb.Command("lb", flag.String("lb", "start", "littleboss start command"))

	flagHTTP := lb.Listener("http", "tcp", ":8877", "littleboss listener address")
	flag.Parse()

	err := loader.Load(context.Background(), config)
	if err != nil {
		fmt.Printf("%v\n", err)
		usage()
	}

	svc, err := store.New(config)
	if err != nil {
		glog.Fatal(err)
	}

	lb.Run(func(ctx context.Context) {
		run(ctx, svc.HTTP(), flagHTTP.Listener())
	})

	glog.Info("aggo exited")
}

func run(ctx context.Context, srv *http.Server, ln net.Listener) {

	go func() {
		if err := srv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			glog.Fatal(err)
		}
	}()

	<-ctx.Done()
	srv.Shutdown(ctx)

}
