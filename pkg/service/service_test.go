package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/golang/glog"
	"github.com/myntra/aggo/pkg/event"
	"github.com/myntra/aggo/pkg/util"
	httpexpect "gopkg.in/gavv/httpexpect.v1"
)

type exampleData struct {
	Alpha string `json:"alpha"`
	Beta  int    `json:"beta"`
}

func tptr(t time.Time) *time.Time { return nil }

var testevent = event.Event{
	&cloudevent.CloudEvent{
		EventType:          "myntra.prod.icinga.check_disk",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		Source:             "/sink",
		EventID:            "42",
		EventTime:          tptr(time.Now()),
		SchemaURL:          "http://www.json.org",
		ContentType:        "application/json",
		Data:               &exampleData{Alpha: "julie", Beta: 42},
		Extensions:         map[string]string{"ext1": "value"},
	},
}

var testRule = event.Rule{
	ID:           "123",
	HookEndpoint: "http://localhost:3000/testrule",
	HookRetry:    2,
	EventTypes:   []string{"myntra.prod.icinga.check_disk", "myntra.prod.site247.cart_down"},
}

var scriptRequest = ScriptRequest{
	ID: "myscript",
	Data: []byte(`
	let result = 0;
	export default function() { result++; }`),
}

var scriptRequestUpdated = ScriptRequest{
	ID: "myscript",
	Data: []byte(`
	let result = 1;
	export default function() { result--; }`),
}

func startService(t *testing.T, cfg *util.Config, svc *Service) {

	go func() {
		if err := svc.HTTP().ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}
			t.Fatal(err)
		}
	}()

}

func stopService(t *testing.T, svc *Service) {
	// close svc
	err := svc.Shutdown(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func singleService(t *testing.T, f func(url string)) {

	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	cfg := &util.Config{
		NodeID:                     "service0",
		RaftBindPort:               6678,
		Dir:                        tmpDir,
		DefaultWaitWindow:          4000, // 3 minutes
		DefaultMaxWaitWindow:       8000, // 6 minutes
		DefaultWaitWindowThreshold: 3800, // 2.5 minutes
		DisablePostHook:            true,
	}

	svc, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	startService(t, cfg, svc)
	defer stopService(t, svc)

	url := "http://localhost" + cfg.GetHTTPAddr()
	// run test
	f(url)
}

func multiService(t *testing.T, f func(urls []string)) {

	var urls []string
	var services []*Service
	tmpDir1, _ := ioutil.TempDir("", "store_test1")
	defer os.RemoveAll(tmpDir1)

	// open store 1
	cfg1 := &util.Config{
		NodeID:                     "node0",
		RaftBindPort:               6778,
		Dir:                        tmpDir1,
		DefaultWaitWindow:          4000, // 3 minutes
		DefaultMaxWaitWindow:       8000, // 6 minutes
		DefaultWaitWindowThreshold: 3800, // 2.5 minutes
		DisablePostHook:            true,
	}

	svc1, err := New(cfg1)
	if err != nil {
		t.Fatal(err)
	}

	startService(t, cfg1, svc1)

	tmpDir2, _ := ioutil.TempDir("", "store_test2")
	defer os.RemoveAll(tmpDir2)

	// open store 2
	cfg2 := &util.Config{
		NodeID:                     "node1",
		RaftBindPort:               6878,
		JoinAddr:                   "0.0.0.0" + cfg1.GetHTTPAddr(),
		Dir:                        tmpDir2,
		DefaultWaitWindow:          4000, // 3 minutes
		DefaultMaxWaitWindow:       8000, // 6 minutes
		DefaultWaitWindowThreshold: 3800, // 2.5 minutes
		DisablePostHook:            true,
	}

	svc2, err := New(cfg2)
	if err != nil {
		t.Fatal(err)
	}

	startService(t, cfg2, svc2)

	tmpDir3, _ := ioutil.TempDir("", "store_test3")
	defer os.RemoveAll(tmpDir3)

	// open store 2
	cfg3 := &util.Config{
		NodeID:                     "node2",
		RaftBindPort:               6978,
		JoinAddr:                   "0.0.0.0" + cfg1.GetHTTPAddr(),
		Dir:                        tmpDir3,
		DefaultWaitWindow:          4000, // 3 minutes
		DefaultMaxWaitWindow:       8000, // 6 minutes
		DefaultWaitWindowThreshold: 3800, // 2.5 minutes
		DisablePostHook:            true,
	}

	svc3, err := New(cfg3)
	if err != nil {
		t.Fatal(err)
	}
	startService(t, cfg3, svc3)

	url1 := "http://localhost" + cfg1.GetHTTPAddr()
	url2 := "http://localhost" + cfg2.GetHTTPAddr()
	url3 := "http://localhost" + cfg3.GetHTTPAddr()
	urls = append(urls, url1, url2, url3)

	time.Sleep(time.Second * 5)

	services = append(services, svc1, svc2, svc3)
	for _, service := range services {
		defer stopService(t, service)
	}

	f(urls)
}

func fail(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func addRule(t *testing.T, url string) {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(&testRule)
	req, err := http.NewRequest("POST", url+"/rules", b)
	fail(t, err)
	glog.Infof("post rule at %v", req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	fail(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	fail(t, err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("unexpected status code", resp.StatusCode)
	}

	var r event.Rule
	err = json.Unmarshal(body, &r)
	fail(t, err)

	if r.ID != testRule.ID {
		t.Fatal("unexpected rule id")
	}
}

func getRulesVerify(t *testing.T, url string) bool {
	req, err := http.NewRequest("GET", url+"/rules", nil)
	fail(t, err)
	glog.Infof("get alls rule at %v", req.URL.String())
	client := &http.Client{}
	resp, err := client.Do(req)
	fail(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	fail(t, err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("unexpected status code", resp.StatusCode)
	}

	var rules []*event.Rule
	err = json.Unmarshal(body, &rules)
	fail(t, err)

	found := false
	for _, rule := range rules {
		glog.Infof("%+v", rule)
		if rule.ID == testRule.ID {
			found = true
			break
		}

	}
	return found
}

func removeRule(t *testing.T, url string) {

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(&testRule)
	req, err := http.NewRequest("DELETE", url+"/rules/"+testRule.ID, nil)
	fail(t, err)
	glog.Infof("delete rule at %v", req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	fail(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	fail(t, err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal("unexpected status code", resp.StatusCode, string(body))
	}

}

func addScript(t *testing.T, url string) {
	e := httpexpect.New(t, url)
	e.POST("/scripts").WithJSON(scriptRequest).Expect().Status(http.StatusOK)
	e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequest)
}

func removeScript(t *testing.T, url string) {

}

func updateScript(t *testing.T, url string) {
	e := httpexpect.New(t, url)
	e.PUT("/scripts").WithJSON(scriptRequestUpdated).Expect().Status(http.StatusOK)
	e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequestUpdated)
}

func TestRuleSingleService(t *testing.T) {
	singleService(t, func(url string) {

		// add rule
		addRule(t, url)

		// get rules and verify
		if !getRulesVerify(t, url) {
			t.Fatal("added rule was not found")
		}

		// remove rule
		removeRule(t, url)

		if getRulesVerify(t, url) {
			t.Fatal("removed rule was found")
		}

	})

}

func TestRuleMultiService(t *testing.T) {
	multiService(t, func(urls []string) {
		url1 := urls[0]
		url2 := urls[1]
		url3 := urls[2]

		// add rule to node0
		addRule(t, url1)

		time.Sleep(time.Second)

		// get rules and verify from node1, replication
		if !getRulesVerify(t, url2) {
			t.Fatal("added rule was not found")
		}

		// remove rule from node2, forwarding
		removeRule(t, url2)

		time.Sleep(time.Second)

		if getRulesVerify(t, url3) {
			t.Fatal("removed rule was found")
		}

	})
}

func TestScriptsSingleSerive(t *testing.T) {
	singleService(t, func(url string) {
		e := httpexpect.New(t, url)

		// add
		e.POST("/scripts").WithJSON(scriptRequest).Expect().Status(http.StatusOK)
		e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequest)

		// update
		e.PUT("/scripts").WithJSON(scriptRequestUpdated).Expect().Status(http.StatusOK)
		e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequestUpdated)

		//remove
		e.DELETE("/scripts/" + scriptRequest.ID).WithJSON(scriptRequestUpdated).Expect().Status(http.StatusOK)
		time.Sleep(time.Second)
		//e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Null()
	})
}
