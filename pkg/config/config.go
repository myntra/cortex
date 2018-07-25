package config

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/golang/glog"
)

// Config is required for initializing the service
type Config struct {
	NodeID               string `config:"id"`
	RaftBindPort         int    `config:"bind"`
	Dir                  string `config:"dir"`
	JoinAddr             string `config:"join"`
	DefaultDwell         uint64 `config:"dwell"`
	DefaultDwellDeadline uint64 `config:"dwell_threshold"`
	DefaultMaxDwell      uint64 `config:"max_dwell"`
	MaxHistory           int    `config:"max_history"`
	Version              string `config:"version"`
	Commit               string `config:"commit"`
	Date                 string `config:"date"`
}

// Validate the config
func (c *Config) Validate() error {
	if !c.validateRaftBindAddr() {
		return fmt.Errorf("invalid bind port. must be a valid int value with the next port available e.g 8788 and 8799 must be available")
	}

	if !c.validateNodeID() {
		return fmt.Errorf("invalid id. must be valid node string e.g: node0")
	}
	return nil
}

func (c *Config) validateRaftBindAddr() bool {
	return checkAddrFree(c.GetBindAddr()) && checkAddrFree(c.GetHTTPAddr())
}

func (c *Config) validateNodeID() bool {
	return c.NodeID != ""
}

// GetBindAddr returns the raft bind address
func (c *Config) GetBindAddr() string {
	return getAddr(":" + strconv.Itoa(c.RaftBindPort+1))
}

// GetHTTPAddr returns the raft bind address
func (c *Config) GetHTTPAddr() string {
	return getAddr(":" + strconv.Itoa(c.RaftBindPort))
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
	conn, _ := net.DialTimeout("tcp", addr, time.Second)
	if conn != nil {
		conn.Close()
		return false
	}
	return true
}
