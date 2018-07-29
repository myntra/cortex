package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

// Config is required for initializing the service
type Config struct {
	NodeID               string `config:"id"`
	Dir                  string `config:"dir"`
	JoinAddr             string `config:"join"`
	FlushInterval        uint64 `config:"flush_interval"`
	SnapshotInterval     int    `config:"snapshot_interval"`
	DefaultDwell         uint64 `config:"dwell"`
	DefaultDwellDeadline uint64 `config:"dwell_deadline"`
	DefaultMaxDwell      uint64 `config:"max_dwell"`
	MaxHistory           int    `config:"max_history"`
	Version              string `config:"version"`
	Commit               string `config:"commit"`
	Date                 string `config:"date"`
	EnableFileServer     bool

	RaftAddr     string
	HTTPAddr     string
	RaftListener net.Listener
	HTTPListener net.Listener
}

// Validate the config
func (c *Config) Validate() error {

	glog.Infof("Validating config %v \n", c)

	if c.RaftAddr == "" {
		return fmt.Errorf("missing raft address. eg: -raft :8080")
	}

	if c.HTTPAddr == "" {
		return fmt.Errorf("missing http address. eg: -http :8081")
	}

	rf := strings.SplitAfter(c.RaftAddr, ":")
	if len(rf) != 2 || rf[0] != ":" {
		return fmt.Errorf("invalid raft address. eg: -raft :8080")
	}

	hf := strings.SplitAfter(c.HTTPAddr, ":")
	if len(hf) != 2 || hf[0] != ":" {
		return fmt.Errorf("invalid http address. eg: -http :8081")
	}

	raftPort, err := strconv.Atoi(rf[1])
	if err != nil {
		return fmt.Errorf("invalid raft address. eg: -raft :8080")
	}

	httpPort, err := strconv.Atoi(hf[1])
	if err != nil {
		return fmt.Errorf("invalid http address. eg: -http :8081")
	}

	if httpPort-raftPort != 1 {
		return fmt.Errorf("invalid raft http address. eg: -raft : 8080" +
			"eg: -http :8081. the http port should be the next port relative to the raft port")
	}

	if c.RaftListener == nil {
		return fmt.Errorf("raft listener is nil")
	}

	if c.HTTPListener == nil {
		return fmt.Errorf("http listener is nil")
	}

	err = c.validateDir()
	if err != nil {
		return err
	}

	if c.FlushInterval == 0 {
		return fmt.Errorf("flush_interval is not set")
	}

	if c.SnapshotInterval == 0 {
		return fmt.Errorf("snapshot_interval in minutes is not set")
	}

	if c.DefaultDwell == 0 {
		return fmt.Errorf("dwell is not set")
	}

	if c.DefaultDwellDeadline == 0 {
		return fmt.Errorf("dwell_deadline is not set")
	}

	if c.DefaultMaxDwell == 0 {
		return fmt.Errorf("max_dwell is not set")
	}

	return nil

}

func (c *Config) validateNodeID() bool {
	return c.NodeID != ""
}

func (c *Config) validateDir() error {
	if c.Dir == "" {
		return fmt.Errorf("raft dir is not set")
	}

	if _, err := os.Stat(c.Dir); os.IsNotExist(err) {
		err := os.Mkdir(c.Dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("raft dir err %v", err)
		}
	}

	return nil
}

func getAddr(addr string) string {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(fmt.Errorf("resolveTCPAddr failed: %v", err))

	}
	glog.Infof("getAddr: %v", tcpAddr.String())
	return tcpAddr.String()
}

func checkAddrFree(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		glog.Errorf("err %v\n", err)
	}
	if conn != nil {
		conn.Close()
		glog.Errorf("addr %v is is not available ", addr)
		return false
	}
	return true
}
