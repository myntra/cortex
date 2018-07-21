package service

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/golang/glog"

	"github.com/satori/go.uuid"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/myntra/aggo/pkg/event"
	"github.com/myntra/aggo/pkg/store"
	"github.com/myntra/aggo/pkg/util"
)

// Service encapsulates the http server and the raft store
type Service struct {
	srv  *http.Server
	node *store.Node
}

// HTTP returns a http server
func (s *Service) HTTP() *http.Server {
	return s.srv
}

// Shutdown everything
func (s *Service) Shutdown(ctx context.Context) error {
	s.srv.Shutdown(ctx)
	if err := s.node.Shutdown(); err != nil {
		return err
	}
	return nil

}

func (s *Service) leaderProxy(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		leaderAddr := s.node.LeaderAddr()
		if leaderAddr == "" {
			h.ServeHTTP(w, r)
		} else {
			glog.Infof("proxying request to leader at %v", leaderAddr)
			proxy := httputil.ReverseProxy{Director: func(r *http.Request) {
				r.URL.Scheme = "http"
				r.URL.Host = leaderAddr
				r.Host = leaderAddr
			}}

			proxy.ServeHTTP(w, r)

		}
	})
}

// eventHandler expects a event in request body and aggregates by type
func (s *Service) eventHandler(w http.ResponseWriter, r *http.Request) {

	event, err := event.FromRequest(r)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body, expected a cloudevents.io event", http.StatusNotAcceptable, err)
		return
	}

	err = s.node.Stash(event)
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

	defer r.Body.Close()

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

	err = s.node.AddRule(&rule)
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
	err := s.node.RemoveRule(ruleID)
	if err != nil {
		util.ErrStatus(w, r, "could not remove rule", http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Service) getRulesHandler(w http.ResponseWriter, r *http.Request) {
	rules := s.node.GetRules()

	b, err := json.Marshal(&rules)
	if err != nil {
		util.ErrStatus(w, r, "rules parsing failed", http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)

}

func (s *Service) leaveHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := s.node.Leave(id)
	if err != nil {
		util.ErrStatus(w, r, "could not leave node ", http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Service) joinHandler(w http.ResponseWriter, r *http.Request) {

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body, expected a valid joinRequest", http.StatusNotAcceptable, err)
		return
	}

	defer r.Body.Close()

	joinRequest := &util.JoinRequest{}
	err = json.Unmarshal(reqBody, joinRequest)
	if err != nil {
		util.ErrStatus(w, r, "joinRequest parsing failed", http.StatusNotAcceptable, err)
		return
	}

	err = joinRequest.Validate()
	if err != nil {
		util.ErrStatus(w, r, "joinRequest validation failed", http.StatusNotAcceptable, err)
		return
	}

	err = s.node.Join(joinRequest.NodeID, joinRequest.Addr)
	if err != nil {
		util.ErrStatus(w, r, "joinining failed", http.StatusNotAcceptable, err)
		return
	}

	w.WriteHeader(http.StatusOK)

}

// New returns the http service wrapper for the store.
func New(cfg *util.Config) (*Service, error) {

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

	router.Get("/rules", svc.getRulesHandler)
	router.Post("/rules", svc.leaderProxy(svc.addRuleHandler))
	router.Delete("/rules/{id}", svc.leaderProxy(svc.removeRuleHandler))

	router.Get("/leave/{id}", svc.leaveHandler)
	router.Post("/join", svc.joinHandler)

	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
		Addr:         cfg.GetHTTPAddr(),
	}

	svc.srv = srv

	return svc, nil
}
