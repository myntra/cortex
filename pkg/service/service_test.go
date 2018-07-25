package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fnproject/cloudevent"
	"github.com/imdario/mergo"
	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/rules"
	httpexpect "gopkg.in/gavv/httpexpect.v1"
)

type exampleData struct {
	Alpha string `json:"alpha"`
	Beta  int    `json:"beta"`
}

func tptr(t time.Time) *time.Time { return nil }

var testevent = events.Event{
	CloudEvent: &cloudevent.CloudEvent{
		EventType:          "acme.prod.icinga.check_disk",
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

var testRule = rules.Rule{
	ID:            "123",
	HookEndpoint:  "http://localhost:3000/testrule",
	HookRetry:     2,
	EventTypes:    []string{"acme.prod.icinga.check_disk", "acme.prod.site247.cart_down"},
	Dwell:         1000,
	DwellDeadline: 800,
	MaxDwell:      2000,
}

var testRuleUpdated = rules.Rule{
	ID:         "123",
	EventTypes: []string{"apple.prod.icinga.check_disk", "acme.prod.site247.cart_down"},
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

func startService(t *testing.T, cfg *config.Config, svc *Service) {

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

	cfg := &config.Config{
		NodeID:               "service0",
		RaftBindPort:         6678,
		Dir:                  tmpDir,
		DefaultDwell:         4000, // 3 minutes
		DefaultMaxDwell:      8000, // 6 minutes
		DefaultDwellDeadline: 3800, // 2.5 minutes
		MaxHistory:           1000,
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
	cfg1 := &config.Config{
		NodeID:               "node0",
		RaftBindPort:         6778,
		Dir:                  tmpDir1,
		DefaultDwell:         4000, // 3 minutes
		DefaultMaxDwell:      8000, // 6 minutes
		DefaultDwellDeadline: 3800, // 2.5 minutes
		MaxHistory:           1000,
	}

	svc1, err := New(cfg1)
	if err != nil {
		t.Fatal(err)
	}

	startService(t, cfg1, svc1)

	tmpDir2, _ := ioutil.TempDir("", "store_test2")
	defer os.RemoveAll(tmpDir2)

	// open store 2
	cfg2 := &config.Config{
		NodeID:               "node1",
		RaftBindPort:         6878,
		JoinAddr:             "0.0.0.0" + cfg1.GetHTTPAddr(),
		Dir:                  tmpDir2,
		DefaultDwell:         4000, // 3 minutes
		DefaultMaxDwell:      8000, // 6 minutes
		DefaultDwellDeadline: 3800, // 2.5 minutes
		MaxHistory:           1000,
	}

	svc2, err := New(cfg2)
	if err != nil {
		t.Fatal(err)
	}

	startService(t, cfg2, svc2)

	tmpDir3, _ := ioutil.TempDir("", "store_test3")
	defer os.RemoveAll(tmpDir3)

	// open store 2
	cfg3 := &config.Config{
		NodeID:               "node2",
		RaftBindPort:         6978,
		JoinAddr:             "0.0.0.0" + cfg1.GetHTTPAddr(),
		Dir:                  tmpDir3,
		DefaultDwell:         4000, // 3 minutes
		DefaultMaxDwell:      8000, // 6 minutes
		DefaultDwellDeadline: 3800, // 2.5 minutes
		MaxHistory:           1000,
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

func ruletest(t *testing.T, url string) {
	// add rule
	e := httpexpect.New(t, url)

	// add
	e.POST("/rules").WithJSON(testRule).Expect().Status(http.StatusOK)
	e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(testRule)

	// update
	e.PUT("/rules").WithJSON(testRuleUpdated).Expect().Status(http.StatusOK)
	testRule.EventTypes = testRuleUpdated.EventTypes
	e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(testRule)

	//remove
	e.DELETE("/rules/" + testRule.ID).Expect().Status(http.StatusOK)
	e.GET("/rules/" + testRule.ID).Expect().Status(http.StatusNotFound)
}

func TestRuleSingleService(t *testing.T) {
	singleService(t, func(url string) {
		ruletest(t, url)
	})

}

func TestRuleMultiService(t *testing.T) {
	multiService(t, func(urls []string) {

		// add to node 1
		e := httpexpect.New(t, urls[0])
		e.POST("/rules").WithJSON(testRule).Expect().Status(http.StatusOK)
		time.Sleep(time.Second)
		// verify from node 2
		e = httpexpect.New(t, urls[1])
		e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(testRule)

		// update on node3
		e = httpexpect.New(t, urls[2])
		e.PUT("/rules").WithJSON(testRuleUpdated).Expect().Status(http.StatusOK)
		testRule.EventTypes = testRuleUpdated.EventTypes
		// verifiy from node 1
		e = httpexpect.New(t, urls[0])
		e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(testRule)

		// delete from node3
		e = httpexpect.New(t, urls[2])
		e.DELETE("/rules/" + testRule.ID).Expect().Status(http.StatusOK)
		time.Sleep(time.Second)
		// verify from node 3
		e = httpexpect.New(t, urls[2])
		e.GET("/rules/" + testRule.ID).Expect().Status(http.StatusNotFound)

	})
}

func scriptstest(t *testing.T, url string) {
	e := httpexpect.New(t, url)
	// add
	e.POST("/scripts").WithJSON(scriptRequest).Expect().Status(http.StatusOK)
	e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequest)

	// update
	e.PUT("/scripts").WithJSON(scriptRequestUpdated).Expect().Status(http.StatusOK)
	e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequestUpdated)

	//remove
	e.DELETE("/scripts/" + scriptRequest.ID).Expect().Status(http.StatusOK)
	e.GET("/scripts/" + scriptRequest.ID).Expect().Status(http.StatusNotFound)
}

func TestScriptsSingleSerive(t *testing.T) {
	singleService(t, func(url string) {
		scriptstest(t, url)
	})
}

func TestScriptsMultiService(t *testing.T) {
	multiService(t, func(urls []string) {

		// add scripts to node 1
		e := httpexpect.New(t, urls[0])
		e.POST("/scripts").WithJSON(scriptRequest).Expect().Status(http.StatusOK)
		time.Sleep(time.Second)

		// verify on node 2
		e = httpexpect.New(t, urls[1])
		e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequest)

		// update scripts on node 2
		e = httpexpect.New(t, urls[1])
		e.POST("/scripts").WithJSON(scriptRequest).Expect().Status(http.StatusOK)
		time.Sleep(time.Second)

		// verify on node 3
		e = httpexpect.New(t, urls[2])
		e.GET("/scripts/" + scriptRequest.ID).Expect().JSON().Equal(scriptRequest)

		// delete on node 3
		e = httpexpect.New(t, urls[2])
		e.DELETE("/scripts/" + scriptRequest.ID).Expect().Status(http.StatusOK)

		// verify on node 1
		e = httpexpect.New(t, urls[0])
		e.GET("/scripts/" + scriptRequest.ID).Expect().Status(http.StatusNotFound)

	})
}

func TestMergeRule(t *testing.T) {
	if err := mergo.Merge(&testRuleUpdated, testRule); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("testRuleUpdated %+v", testRuleUpdated)

}
