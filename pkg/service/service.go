package service

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang/glog"

	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/store"
)

// Service encapsulates the http server and the raft store
type Service struct {
	srv      *http.Server
	node     *store.Node
	listener net.Listener
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
	if err := s.srv.Serve(s.listener); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		for {
			select {
			case <-ticker.C:
				glog.Infof("take snapshot => %v", s.node.Snapshot())
			}
		}
	}()

	return nil
}

// FileServer starts the file server and return the file
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	fmt.Println("Starting the file server at", root)

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
		node: node,
	}

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	workDir, err := os.Getwd()
	if err != nil {
		glog.Fatal("Error in fetching the current working directory")
	}
	filesDir := filepath.Join(workDir, "build")
	FileServer(router, "/ui", http.Dir(filesDir))

	router.Post("/event", svc.leaderProxy(svc.eventHandler))
	router.Post("/event/sink/site247", svc.site247AlertHandler)

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
