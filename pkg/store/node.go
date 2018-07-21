package store

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/hashicorp/raft"
	"github.com/myntra/aggo/pkg/event"
	"github.com/myntra/aggo/pkg/util"
)

// Node represents a raft node
type Node struct {
	mu       sync.RWMutex
	listener net.Listener
	store    *defaultStore
	quit     chan struct{}
}

// NewNode returns a new raft node
func NewNode(cfg *util.Config) (*Node, error) {
	if cfg.NodeID == "" {
		return nil, fmt.Errorf("no id provided")
	}

	store, err := newStore(cfg)

	if err != nil {
		return nil, err
	}

	// join a remote node
	if cfg.JoinAddr != "" {
		err := tcpRaftJoin(cfg.JoinAddr, cfg.NodeID, cfg.BindAddr)
		if err != nil {
			return nil, err
		}
	}

	listener, err := net.Listen("tcp", cfg.ListenAddr)
	glog.Infof("raft server listen in %s", cfg.ListenAddr)
	if err != nil {
		return nil, err
	}

	node := &Node{store: store, listener: listener, quit: make(chan struct{})}
	go node.run()
	return node, nil
}

func (n *Node) handleConn(conn net.Conn) {

	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		ln := scanner.Text()
		fs := strings.Fields(ln)

		switch fs[0] {
		case "join":
			if len(fs) != 3 {
				glog.Errorf("Invalid join command %v", fs)
				continue
			}
			nodeID := fs[1]
			addr := fs[2]
			glog.Info(nodeID, addr)
			n.store.acceptJoin(nodeID, addr)
		case "leave":
			if len(fs) != 2 {
				glog.Errorf("Invalid leave command %v", fs)
				continue
			}
			nodeID := fs[1]
			glog.Info(nodeID)
			n.store.acceptLeave(nodeID)

		case "addr":
			if len(fs) != 1 {
				glog.Errorf("Invalid addr command %v", fs)
				continue
			}

			_, err := conn.Write([]byte(n.store.opt.ListenAddr))
			if err != nil {
				glog.Errorf("could not write listenaddr %v", fs)
				continue
			}
		}

	}
}

// run accepts connections
func (n *Node) run() {

	// accept connections
	for {
		select {

		default:
			// accept new client connect and perform
			conn, err := n.listener.Accept()
			if err != nil {
				glog.Error(err.Error())
				select {
				case <-n.quit:
					return
				default:
					// thanks to martingx on reddit for noticing I am missing a default
					// case. without the default case the select will block.
				}

				glog.Errorf("raft tcp accept err %v", err)
				continue
			}

			go n.handleConn(conn)
		}
	}

}

// LeaderAddr returns the http addr of the leader of the cluster. If empty, the current node is the leader
func (n *Node) LeaderAddr() string {

	if n.store.raft.State() != raft.Leader {
		return ""
	}

	addr, err := getListenAddr(string(n.store.raft.Leader()))
	if err != nil {
		return ""
	}

	return addr
}

// Shutdown store
func (n *Node) Shutdown() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	glog.Info("quit accept conn loop")
	close(n.quit)
	n.listener.Close()
	glog.Info("accept conn loop shutdown")
	err := n.store.close()
	if err != nil {
		glog.Errorf("error shutting down node %v", err)
		return err
	}

	return nil
}

// AddRule adds a rule to the store
func (n *Node) AddRule(rule *event.Rule) error {
	return n.store.addRule(rule)
}

// Stash adds a event to the store
func (n *Node) Stash(event *event.Event) error {
	return n.store.stash(event)
}

// RemoveRule removes a rule from the store
func (n *Node) RemoveRule(ruleID string) error {
	return n.store.removeRule(ruleID)
}

// GetRules returns all the stored rules
func (n *Node) GetRules() []*event.Rule {
	return n.store.getRules()
}

func tcpRaftJoin(joinAddr, nodeID, bindAddr string) error {
	cmd := "join " + nodeID + " " + bindAddr
	return tcpcmd(joinAddr, []byte(cmd))
}

func tcpRaftLeave(leaveAddr, nodeID string) error {
	cmd := "leave " + nodeID
	return tcpcmd(leaveAddr, []byte(cmd))
}

func tcpcmd(remoteAddr string, cmd []byte) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		glog.Errorf("resolveTCPAddr failed: %v", err)
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		glog.Errorf("dial remote node failed: %v", err)
		return err
	}

	defer conn.Close()

	_, err = conn.Write(cmd)
	if err != nil {
		glog.Errorf("write cmd failed: %v", err)
		return err
	}

	return nil
}

func getListenAddr(remoteAddr string) (string, error) {
	var addr string

	tcpAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		glog.Errorf("resolveTCPAddr failed: %v", err)
		return "", err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		glog.Errorf("dial remote node failed: %v", err)
		return "", err
	}

	defer conn.Close()

	_, err = conn.Write([]byte("addr"))
	if err != nil {
		glog.Errorf("write cmd failed: %v", err)
		return "", err
	}

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		glog.Errorf("read from conn failed: %v", err)
		return "", err
	}

	addr = string(reply)

	return addr, nil
}
