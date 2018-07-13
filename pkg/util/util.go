package util

import (
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

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
