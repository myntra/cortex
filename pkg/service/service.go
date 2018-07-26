package service

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang/glog"

	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/events/sinks"
	"github.com/myntra/cortex/pkg/rules"
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

func (s *Service) addScriptHandler(w http.ResponseWriter, r *http.Request) {

	scriptData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	defer r.Body.Close()
	sr := &ScriptRequest{}
	err = json.Unmarshal(scriptData, sr)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	err = sr.Validate()
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	err = s.node.AddScript(sr.ID, sr.Data)
	if err != nil {
		util.ErrStatus(w, r, "error adding script", http.StatusNotAcceptable, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func (s *Service) updateScriptHandler(w http.ResponseWriter, r *http.Request) {

	scriptData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	defer r.Body.Close()
	sr := &ScriptRequest{}
	err = json.Unmarshal(scriptData, sr)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	err = sr.Validate()
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	err = s.node.UpdateScript(sr.ID, sr.Data)
	if err != nil {
		util.ErrStatus(w, r, "error adding script", http.StatusNotAcceptable, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func (s *Service) removeScriptHandler(w http.ResponseWriter, r *http.Request) {
	scriptID := chi.URLParam(r, "id")
	err := s.node.RemoveScript(scriptID)
	if err != nil {
		util.ErrStatus(w, r, "could not remove script", http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

func (s *Service) getScriptHandler(w http.ResponseWriter, r *http.Request) {
	scriptID := chi.URLParam(r, "id")
	scriptData := s.node.GetScript(scriptID)
	if len(scriptData) == 0 {
		util.ErrStatus(w, r, "no script data", http.StatusNotFound, fmt.Errorf("script data len 0"))
		return
	}

	sr := &ScriptRequest{
		ID:   scriptID,
		Data: scriptData,
	}

	b, err := json.Marshal(&sr)
	if err != nil {
		util.ErrStatus(w, r, "error writing script data ", http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}

func (s *Service) getScriptListHandler(w http.ResponseWriter, r *http.Request) {
	scriptIds := s.node.GetScripts()

	b, err := json.Marshal(&scriptIds)
	if err != nil {
		util.ErrStatus(w, r, "scripts list parsing failed", http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *Service) site247AlertHandler(w http.ResponseWriter, r *http.Request) {

	alertData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	defer r.Body.Close()
	alert := &sinks.Site247Alert{}
	err = json.Unmarshal(alertData, alert)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body", http.StatusNotAcceptable, err)
		return
	}

	event := sinks.EventFromSite247(*alert)

	err = s.node.Stash(event)
	if err != nil {
		util.ErrStatus(w, r, "error stashing event", http.StatusInternalServerError, err)
		return
	}

	b, err := json.Marshal(event)
	if err != nil {
		util.ErrStatus(w, r, "error writing event data", http.StatusNotAcceptable, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
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
