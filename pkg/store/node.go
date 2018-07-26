package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/myntra/cortex/pkg/executions"
	"github.com/myntra/cortex/pkg/js"

	"github.com/golang/glog"
	"github.com/hashicorp/raft"
	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/rules"
	"github.com/myntra/cortex/pkg/util"
)

// Node represents a raft node
type Node struct {
	mu    sync.RWMutex
	store *defaultStore
}

// NewNode returns a new raft node
func NewNode(cfg *config.Config) (*Node, error) {
	glog.Infof("NewNode %v\n", cfg)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	store, err := newStore(cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{store: store}

	return node, nil
}

// Start the node
func (n *Node) Start() error {
	return n.store.open()
}

// Shutdown the node
func (n *Node) Shutdown() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	err := n.store.close()
	if err != nil {
		glog.Errorf("error shutting down node %v", err)
		return err
	}
	glog.Info("node shut down")
	return nil
}

// LeaderAddr returns the http addr of the leader of the cluster. If empty, the current node is the leader
func (n *Node) LeaderAddr() string {

	if n.store.raft.State() == raft.Leader {
		return ""
	}

	raftAddress := string(n.store.raft.Leader())

	fields := strings.Split(raftAddress, ":")

	if len(fields) == 0 || len(fields) != 2 {
		glog.Errorf("invalid raftAddress %v", raftAddress)
		return ""
	}

	raftPortStr := fields[1]
	raftPort, err := strconv.Atoi(raftPortStr)
	if err != nil {
		glog.Errorf("invalid port %v %v", raftAddress, raftPortStr)
		return ""
	}

	tcpPort := raftPort + 1
	tcpURL := fields[0]
	if tcpURL == "" {
		tcpURL = "0.0.0.0"
	}

	glog.Info("LeaderAddr ", tcpURL, tcpPort)

	tcpAddr := fmt.Sprintf("%s:%d", tcpURL, tcpPort)

	return tcpAddr
}

// AddRule adds a rule to the store
func (n *Node) AddRule(rule *rules.Rule) error {
	if err := rule.Validate(); err != nil {
		return err
	}
	return n.store.addRule(rule)
}

// UpdateRule updates a rule to the store
func (n *Node) UpdateRule(rule *rules.Rule) error {
	if err := rule.Validate(); err != nil {
		return err
	}
	return n.store.updateRule(rule)
}

// Stash adds a event to the store
func (n *Node) Stash(event *events.Event) error {
	glog.Info("node stash event => ", event)
	return n.store.matchAndStash(event)
}

// RemoveRule removes a rule from the store
func (n *Node) RemoveRule(ruleID string) error {
	return n.store.removeRule(ruleID)
}

// GetRule returns all the stored rules
func (n *Node) GetRule(ruleID string) *rules.Rule {
	return n.store.getRule(ruleID)
}

// GetRuleExectutions returns the executions for a rule
func (n *Node) GetRuleExectutions(ruleID string) []*executions.Record {
	return n.store.getRecords(ruleID)
}

// GetRules returns all the stored rules
func (n *Node) GetRules() []*rules.Rule {
	return n.store.getRules()
}

// AddScript adds a script to the db
func (n *Node) AddScript(script *js.Script) error {
	return n.store.addScript(script)
}

// UpdateScript updates an already added script
func (n *Node) UpdateScript(script *js.Script) error {
	return n.store.updateScript(script)
}

// RemoveScript remove a script from the db
func (n *Node) RemoveScript(id string) error {
	return n.store.removeScript(id)
}

// GetScripts returns all  script ids
func (n *Node) GetScripts() []string {
	return n.store.getScripts()
}

// GetScript returns the script data
func (n *Node) GetScript(id string) *js.Script {
	return n.store.getScript(id)
}

// Join a remote node at the addr
func (n *Node) Join(nodeID, addr string) error {
	return n.store.acceptJoin(nodeID, addr)
}

// Leave a remote node
func (n *Node) Leave(nodeID string) error {
	return n.store.acceptLeave(nodeID)
}

// Snapshot takes a snapshot of the store
func (n *Node) Snapshot() error {
	return n.store.snapshot()
}

func httpRaftJoin(joinAddr, nodeID, bindAddr string) error {

	jr := &util.JoinRequest{
		NodeID: nodeID,
		Addr:   bindAddr,
	}

	err := jr.Validate()
	if err != nil {
		return err
	}

	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(jr)
	if err != nil {
		return err
	}

	glog.Infof("joinRequest Body %v", b.String())

	req, err := http.NewRequest("POST", "http://"+joinAddr+"/join", b)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("join failed, unexpected status code %v", resp.StatusCode)
	}

	return nil
}
