package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/golang/glog"
	"github.com/imdario/mergo"

	"github.com/satori/go.uuid"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/rules"
	"github.com/myntra/cortex/pkg/store"
	"github.com/myntra/cortex/pkg/util"
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

	event, err := events.FromRequest(r)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body, expected a cloudevents.io event", http.StatusNotAcceptable, err)
		return
	}

	err = s.node.Stash(event)
	if err != nil {
		util.ErrStatus(w, r, "error stashing event", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *Service) addRuleHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body, expected a valid rule", http.StatusNotAcceptable, err)
		return
	}

	defer r.Body.Close()

	var rule rules.Rule
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (s *Service) updateRuleHandler(w http.ResponseWriter, r *http.Request) {
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		util.ErrStatus(w, r, "invalid request body, expected a valid rule", http.StatusNotAcceptable, err)
		return
	}

	defer r.Body.Close()

	var rule rules.Rule
	err = json.Unmarshal(reqBody, &rule)
	if err != nil {
		util.ErrStatus(w, r, "rule parsing failed", http.StatusNotAcceptable, err)
		return
	}

	existingRule := s.node.GetRule(rule.ID)
	if existingRule == nil {
		util.ErrStatus(w, r, "update rule failed, rule not found", http.StatusNotFound, fmt.Errorf("rule is nil"))
	}

	if err := mergo.Merge(&rule, existingRule); err != nil {
		util.ErrStatus(w, r, "updating rule failed", http.StatusInternalServerError, err)
		return
	}

	err = s.node.UpdateRule(&rule)
	if err != nil {
		util.ErrStatus(w, r, "updating rule failed", http.StatusNotAcceptable, err)
		return
	}

	b, err := json.Marshal(&rule)
	if err != nil {
		util.ErrStatus(w, r, "updating rule failed. rules parsing failed", http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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

func (s *Service) getRuleHandler(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "id")

	rule := s.node.GetRule(ruleID)
	if rule == nil {
		util.ErrStatus(w, r, "rule not found", http.StatusNotFound, fmt.Errorf("rule is nil"))
	}

	b, err := json.Marshal(rule)
	if err != nil {
		util.ErrStatus(w, r, "rules parsing failed", http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}

func (s *Service) getRulesHandler(w http.ResponseWriter, r *http.Request) {
	rules := s.node.GetRules()

	b, err := json.Marshal(&rules)
	if err != nil {
		util.ErrStatus(w, r, "rules parsing failed", http.StatusNotFound, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

// ScriptRequest is the container for add/update script
type ScriptRequest struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}

// Validate validates the scriptrequst
func (s *ScriptRequest) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("no id provided")
	}

	if len(s.Data) == 0 {
		return fmt.Errorf("script data len 0")
	}
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

	router.Get("/rules", svc.getRulesHandler)
	router.Get("/rules/{id}", svc.getRuleHandler)
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
		Addr:         cfg.GetHTTPAddr(),
	}

	svc.srv = srv

	return svc, nil
}
