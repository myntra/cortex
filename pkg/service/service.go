package service

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang/glog"

	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/store"
)

// Service encapsulates the http server and the raft store
type Service struct {
	srv              *http.Server
	node             *store.Node
	listener         net.Listener
	snapshotInterval int
	httpAddr         string
}

// Shutdown the service
func (s *Service) Shutdown(ctx context.Context) error {
	s.srv.Shutdown(ctx)
	if err := s.node.Shutdown(); err != nil {
		return err
	}
	return nil
}

// Start the service
func (s *Service) Start() error {

	// start the raft node
	if err := s.node.Start(); err != nil {
		return err
	}

	// start the http service
	go func() {
		if err := s.srv.Serve(s.listener); err != nil {
			glog.Infof("server closed %v", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Minute * time.Duration(s.snapshotInterval))
		for {
			select {
			case <-ticker.C:
				glog.Infof("take snapshot => %v", s.node.Snapshot())
			}
		}
	}()

	glog.Infof("======> join addr %v%v\n", getOutboundIP(), s.httpAddr)
	glog.Infof("======> open ui http://%v%v/ui or http://localhost%v/ui\n", getOutboundIP(), s.httpAddr, s.httpAddr)

	return nil
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// fileServer starts the file server and return the file
func fileServer(r chi.Router, path string) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(rice.MustFindBox("build").HTTPBox()))

	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

// New returns the http service wrapper for the store.
func New(cfg *config.Config) (*Service, error) {

	node, err := store.NewNode(cfg)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		node:             node,
		snapshotInterval: cfg.SnapshotInterval,
		httpAddr:         cfg.HTTPAddr,
	}

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			glog.Infof("Received request on %s %s", r.URL.String(), r.RequestURI)
			next.ServeHTTP(w, r)
		})
	})

	if cfg.EnableFileServer {
		fileServer(router, "/ui")
	}

	router.Post("/event", svc.leaderProxy(svc.eventHandler))
	router.Post("/event/sink/site247", svc.leaderProxy(svc.site247AlertHandler))
	router.Post("/event/sink/icinga", svc.leaderProxy(svc.icingaAlertHandler))

	router.Get("/rules", svc.getRulesHandler)
	router.Get("/rules/{id}", svc.getRuleHandler)
	router.Get("/rules/{id}/executions", svc.getRulesExecutions)
	router.Post("/rules", svc.leaderProxy(svc.addRuleHandler))
	router.Put("/rules", svc.leaderProxy(svc.updateRuleHandler))
	router.Delete("/rules/{id}", svc.leaderProxy(svc.removeRuleHandler))

	router.Get("/scripts", svc.getScriptListHandler)
	router.Get("/scripts/{id}", svc.getScriptHandler)
	router.Post("/scripts", svc.leaderProxy(svc.addScriptHandler))
	router.Put("/scripts", svc.leaderProxy(svc.updateScriptHandler))
	router.Delete("/scripts/{id}", svc.leaderProxy(svc.removeScriptHandler))

	router.Get("/leave/{id}", svc.leaveHandler)
	router.Post("/join", svc.joinHandler)

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/"+r.URL.String(), 302)
	})

	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
	}

	svc.srv = srv
	svc.listener = cfg.HTTPListener

	return svc, nil
}
