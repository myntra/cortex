package service

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/executions"

	"bytes"
	"fmt"

	"github.com/imdario/mergo"
	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events/sinks"
	"github.com/myntra/cortex/pkg/rules"
	"gopkg.in/gavv/httpexpect.v1"
)

type exampleData struct {
	Alpha string `json:"alpha"`
	Beta  int    `json:"beta"`
}

var testalertsite247 = &sinks.Site247Alert{
	MonitorName:          "brand_test",
	MonitorGroupName:     "search",
	SearchPollFrequency:  1,
	MonitorID:            2136797812307,
	FailedLocations:      "Delhi,Bangalore",
	MonitorURL:           "https://localhost:4000/search?query=brand_test",
	IncidentTimeISO:      "2018-07-24T18:43:08+0530",
	MonitorType:          "URL",
	Status:               "DOWN",
	Timezone:             "Asia/Calcutta",
	IncidentTime:         "July 24, 2018 6:43 PM IST",
	IncidentReason:       "Host Unavailable",
	OutageTimeUnixFormat: 1532437988741,
	RCALink:              "https://www.rcalinkdummy.com/somelink",
}

var testevent = &events.Event{
	EventType:          "acme.prod.icinga.check_disk",
	EventTypeVersion:   "1.0",
	CloudEventsVersion: "0.1",
	Source:             "/sink",
	EventID:            "42",
	EventTime:          time.Now(),
	SchemaURL:          "http://www.json.org",
	ContentType:        "application/json",
	Data:               &exampleData{Alpha: "julie", Beta: 42},
	Extensions:         map[string]string{"ext1": "value"},
}

var testRule = rules.Rule{
	ID:            "123",
	ScriptID:      "myscript",
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

var testBucketScript = ScriptRequest{
	ID: "myscript",
	Data: []byte(`
	let result = null;
	export default function(data) { 
		let event = data.events[0];
		if (!event){
			result = {error:"event is undefined"}
			return
		}
		event.data.alpha = event.data.alpha + "test";
		event.data.beta = event.data.beta + 42;
		result = event.data;
	}`),
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
		FlushInterval:        1000,
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
		FlushInterval:        1000,
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
		FlushInterval:        1000,
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
		FlushInterval:        1000,
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
	var cloneTestRule rules.Rule
	if err := mergo.Merge(&cloneTestRule, testRule); err != nil {
		t.Fatal(err)
	}
	cloneTestRule.EventTypes = testRuleUpdated.EventTypes
	e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(cloneTestRule)

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
		var cloneTestRule rules.Rule
		if err := mergo.Merge(&cloneTestRule, testRule); err != nil {
			t.Fatal(err)
		}
		cloneTestRule.EventTypes = testRuleUpdated.EventTypes
		// verifiy from node 1
		e = httpexpect.New(t, urls[0])
		e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(cloneTestRule)

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

	//fmt.Printf("testRuleUpdated %+v", testRuleUpdated)

}

func TestSingleEventSingleService(t *testing.T) {
	singleService(t, func(url string) {
		e := httpexpect.New(t, url)

		//post script
		e.POST("/scripts").WithJSON(testBucketScript).Expect().Status(http.StatusOK)
		e.GET("/scripts/" + testBucketScript.ID).Expect().JSON().Equal(testBucketScript)

		// post rule
		e.POST("/rules").WithJSON(testRule).Expect().Status(http.StatusOK)
		e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(testRule)

		// post event
		e.POST("/event").WithJSON(testevent).Expect().Status(http.StatusOK)

		// wait for rule execution

		time.Sleep(time.Millisecond * time.Duration(testRule.Dwell+3000))

		// fetch rule executions
		resp, err := http.Get(url + "/rules/" + testRule.ID + "/executions")
		if err != nil {
			t.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		defer resp.Body.Close()

		var ruleExecutions []*executions.Record
		err = json.Unmarshal(body, &ruleExecutions)
		if err != nil {
			t.Fatal(err)
		}

		if len(ruleExecutions) == 0 {
			glog.Info("no executions found")
			t.Fatal()
		}

		if testRule.ID != ruleExecutions[0].Bucket.Rule.ID {
			t.Fatal("unexpected rule id")
		}

		if testevent.EventID != ruleExecutions[0].Bucket.Events[0].EventID {
			t.Fatal("unexpected event id")
		}

		scriptResult, ok := ruleExecutions[0].ScriptResult.(map[string]interface{})
		glog.Infof("%+v %v %v\n", ruleExecutions[0], scriptResult, ok)
		if !ok {
			t.Fatal("unexpected script result")
		}

		if !strings.Contains(scriptResult["alpha"].(string), "julietest") {
			t.Fatal("unexpected script result ", scriptResult["alpha"])
		}

	})
}

func TestSingleEventMultipleService(t *testing.T) {
	multiService(t, func(urls []string) {

		// post rule to node 2
		e := httpexpect.New(t, urls[1])
		e.POST("/rules").WithJSON(testRule).Expect().Status(http.StatusOK)
		// verify rule from node 1
		e = httpexpect.New(t, urls[0])
		e.GET("/rules/" + testRule.ID).Expect().JSON().Equal(testRule)

		// post event to node 3
		e = httpexpect.New(t, urls[2])
		e.POST("/event").WithJSON(testevent).Expect().Status(http.StatusOK)

		// wait for rule execution

		time.Sleep(time.Millisecond * time.Duration(testRule.Dwell+3000))

		// fetch rule executions from node 1
		resp, err := http.Get(urls[0] + "/rules/" + testRule.ID + "/executions")
		if err != nil {
			t.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		defer resp.Body.Close()

		var ruleExecutions []*executions.Record
		err = json.Unmarshal(body, &ruleExecutions)
		if err != nil {
			t.Fatal(err)
		}

		if len(ruleExecutions) == 0 {
			glog.Info("no executions found")
			t.Fatal()
		}

		if testRule.ID != ruleExecutions[0].Bucket.Rule.ID {
			t.Fatal("unexpected rule id")
		}

		if testevent.EventID != ruleExecutions[0].Bucket.Events[0].EventID {
			t.Fatal("unexpected event id")
		}

	})
}

func TestSite247Handler(t *testing.T) {
	singleService(t, func(url string) {
		e := httpexpect.New(t, url)

		// post event
		e.POST("/event/sink/site247").WithJSON(testalertsite247).Expect().Status(http.StatusOK)

		// fetch rule executions
		s, err := json.Marshal(testalertsite247)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.Post(url+"/event/sink/site247", "application/json", bytes.NewReader(s))
		if err != nil {
			t.Fatal(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		defer resp.Body.Close()

		var eventBody events.Event
		err = json.Unmarshal(body, &eventBody)
		if err != nil {
			t.Fatal(err)
		}

		byteData, err := json.Marshal(eventBody.Data)
		if err != nil {
			t.Fatal(err)
		}

		var eventData sinks.Site247Alert
		err = json.Unmarshal(byteData, &eventData)
		if err != nil {
			t.Fatal(err)
		}

		if eventData != *testalertsite247 {
			t.Fatal("unexpected eventbody data")
		}

		if eventBody.EventType != fmt.Sprintf("site247.%s.%s", testalertsite247.MonitorGroupName, testalertsite247.MonitorName) {
			t.Fatal("unexpected eventtype data")
		}

	})
}
