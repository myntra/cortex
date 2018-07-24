package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/sethgrid/pester"
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

// RetryPost posts the value to a remote endpoint. also retries
func RetryPost(val interface{}, url string, retry int) {

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(val)
	if err != nil {
		glog.Errorf("post rule bucket failed. dropping it!! %v %v %v", err, b.String(), err)
		return
	}
	req, err := http.NewRequest("POST", url, b)
	if err != nil {
		glog.Errorf("post rule bucket failed. dropping it!! %v %v %v", err, b.String(), err)
		return
	}
	req.Header.Add("Content-type", "application/json")

	client := pester.New()
	client.MaxRetries = retry
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("post rule bucket failed. dropping it!! %v %v %v", err, b.String(), err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		glog.Errorf("post rule bucket failed. dropping it!! %v %v %v", err, b.String(), err)
		return //fmt.Errorf("invalid status code return from %v endpoint", url)
	}

	return

}

// PatternMatch checks if the input is a match to a field of the patterns array
func PatternMatch(in string, patterns []string) bool {
	for _, pattern := range patterns {
		if in == pattern {
			return true
		}
	}

	return false
}
