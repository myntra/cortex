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
	"github.com/myntra/aggo/pkg/aggregate"
)

var (
	//raft
	bind string
	join string
	dir  string
	id   string

	// build
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: \n\n ./aggo -stderrthreshold=INFO -log_dir=$(pwd) -id=node1 lb=start & \n "+
		"./aggo-new-version -stderrthreshold=INFO  -log_dir=$(pwd) -id=node1 -lb=reload \n ./aggo -lb=stop \n \n")
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
	flag.StringVar(&id, "id", "", "raft node id")
	flag.StringVar(&bind, "bind", ":8878", "raft bind addr")
	flag.StringVar(&dir, "dir", "data", "raft data directory")
	flag.StringVar(&join, "join", "", "raft join by cluster addr")

}

func main() {

	lb := littleboss.New("aggo")
	lb.Command("lb", flag.String("lb", "start", "littleboss start command"))

	flagHTTP := lb.Listener("http", "tcp", ":8877", "littleboss listener address")
	flag.Parse()

	if len(os.Args) < 2 {
		usage()
	}

	svc, err := aggregate.New(id, bind, dir, join)
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
