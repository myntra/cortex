package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/satori/go.uuid"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/myntra/aggo/pkg/event"
	"github.com/myntra/aggo/pkg/util"
)

// Service encapsulates the http server and the raft store
type Service struct {
	srv    *http.Server
	store  Store
	config *Config
}

// Config is required for initializing the service
type Config struct {
	NodeID                     string `config:"id,required"`
	BindAddr                   string `config:"bind,required"`
	Dir                        string `config:"dir"`
	JoinAddr                   string `config:"join"`
	DefaultWaitWindow          uint64 `config:"wait_window"`
	DefaultWaitWindowThreshold uint64 `config:"wait_window_threshold"`
	DefaultMaxWaitWindow       uint64 `config:"max_wait_window"`
	Version                    string `config:"version"`
	Commit                     string `config:"commit"`
	Date                       string `config:"date"`
}

// HTTP returns a http server
func (s *Service) HTTP() *http.Server {
	return s.srv
}

// eventHandler expects a event in request body and aggregates by type
func (s *Service) eventHandler(w http.ResponseWriter, r *http.Request) {

	event, err := event.FromRequest(r)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body, expected a cloudevents.io event", http.StatusNotAcceptable, err)
		return
	}

	err = s.store.Stash(event)
	if err != nil {
		util.ErrStatus(w, r, "error stashing event", http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Service) addRuleHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body, expected a valid rule", http.StatusNotAcceptable, err)
		return
	}

	var rule event.Rule
	err = json.Unmarshal(reqBody, &rule)
	if err != nil {
		util.ErrStatus(w, r, "rule parsing failed", http.StatusNotAcceptable, err)
		return
	}

	if rule.ID == "" {
		uid, err := uuid.NewV4()
		if err != nil {
			util.ErrStatus(w, r, "id gen failed", http.StatusNotAcceptable, err)
			return
		}
		rule.ID = uid.String()
	}

	err = s.store.AddRule(&rule)
	if err != nil {
		util.ErrStatus(w, r, "adding rule failed", http.StatusNotAcceptable, err)
		return
	}

	b, err := json.Marshal(&rule)
	if err != nil {
		util.ErrStatus(w, r, "rules parsing failed", http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *Service) removeRuleHandler(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "id")
	err := s.store.RemoveRule(ruleID)
	if err != nil {
		util.ErrStatus(w, r, "could not remove rule", http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Service) getRulesHandler(w http.ResponseWriter, r *http.Request) {
	rules := s.store.GetRules()

	b, err := json.Marshal(&rules)
	if err != nil {
		util.ErrStatus(w, r, "rules parsing failed", http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)

}

func (s *Service) leaveHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	err := s.store.Leave(nodeID)
	if err != nil {
		util.ErrStatus(w, r, "could not leave cluster", http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Service) joinHandler(w http.ResponseWriter, r *http.Request) {
	addr := chi.URLParam(r, "addr")
	err := s.store.Join(addr)
	if err != nil {
		util.ErrStatus(w, r, "could not join cluster", http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// New returns the aggregate store service
func New(cfg *Config) (*Service, error) {

	if cfg.NodeID == "" {
		return nil, fmt.Errorf("no id provided")
	}

	store, err := newStore(cfg)

	if err != nil {
		return nil, err
	}

	svc := &Service{
		store:  store,
		config: cfg,
	}

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Post("/event", svc.eventHandler)
	router.Get("/rule/list", svc.getRulesHandler)
	router.Post("/rule", svc.addRuleHandler)
	router.Delete("/rule/:id", svc.removeRuleHandler)
	router.Get("/join/:addr", svc.joinHandler)
	router.Get("/leave/:id", svc.leaveHandler)

	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
	}

	svc.srv = srv

	return svc, nil
}
