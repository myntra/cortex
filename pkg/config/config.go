package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
)

// Config is required for initializing the service
type Config struct {
	NodeID               string `config:"id"`
	RaftBindPort         int    `config:"bind"`
	Dir                  string `config:"dir"`
	JoinAddr             string `config:"join"`
	FlushInterval        uint64 `config:"flush_interval"`
	DefaultDwell         uint64 `config:"dwell"`
	DefaultDwellDeadline uint64 `config:"dwell_deadline"`
	DefaultMaxDwell      uint64 `config:"max_dwell"`
	MaxHistory           int    `config:"max_history"`
	Version              string `config:"version"`
	Commit               string `config:"commit"`
	Date                 string `config:"date"`

	bindAddr string
	httpAddr string
}

// Validate the config
func (c *Config) Validate() error {

	glog.Infof("Validating config %v \n", c)
	if !c.validateRaftBindAddr() {
		return fmt.Errorf("invalid bind port. must be a valid int value with the next port available e.g 8788 and 8799 must be available")
	}

	if !c.validateNodeID() {
		return fmt.Errorf("invalid id. must be valid node string e.g: node0")
	}

	err := c.validateDir()
	if err != nil {
		return err
	}

	if c.FlushInterval == 0 {
		return fmt.Errorf("flush_interval is not set")
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

func (c *Config) validateRaftBindAddr() bool {
	return checkAddrFree(c.GetBindAddr()) && checkAddrFree(c.GetHTTPAddr())
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

var onceGetBindAddr sync.Once

// GetBindAddr returns the raft bind address
func (c *Config) GetBindAddr() string {
	glog.Info("GetBindAddr")
	if c.bindAddr == "" {
		onceGetBindAddr.Do(func() {
			c.bindAddr = getAddr(":" + strconv.Itoa(c.RaftBindPort+1))
		})
	}

	return c.bindAddr
}

var onceGetHTTPAddr sync.Once

// GetHTTPAddr returns the raft bind address
func (c *Config) GetHTTPAddr() string {
	glog.Info("GetHTTPAddr")
	if c.httpAddr == "" {
		onceGetHTTPAddr.Do(func() {
			c.httpAddr = getAddr(":" + strconv.Itoa(c.RaftBindPort))
		})
	}
	return c.httpAddr
}

var mu sync.Mutex

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
