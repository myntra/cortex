package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/golang/glog"
)

// JoinRequest is the request to join a node
type JoinRequest struct {
	NodeID string `json:"nodeID"`
	Addr   string `json:"addr"`
}

// Validate validates the requet
func (j *JoinRequest) Validate() error {

	if j.NodeID == "" {
		return fmt.Errorf("nodeID is empty")
	}

	_, err := net.DialTimeout("tcp", j.Addr, time.Second*3)
	if err != nil {
		return fmt.Errorf("invalid addr %v", err)
	}

	return nil
}

// ErrStatus sends a http error status
func ErrStatus(w http.ResponseWriter, r *http.Request, message string, statusCode int, err error) {
	var content []byte
	var e error

	content, e = ioutil.ReadAll(r.Body)
	if e != nil {
		glog.Error("ioutil.ReadAll failed")
	}

	glog.Errorf("msg %v, r.Body %v, err: %v", message, string(content), err)

	http.Error(w, message, statusCode)
}
