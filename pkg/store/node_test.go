package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/golang/glog"

	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/js"
	"github.com/myntra/cortex/pkg/rules"
)

type exampleData struct {
	Alpha string `json:"alpha"`
	Beta  int    `json:"beta"`
}

func tptr(t time.Time) *time.Time { return nil }

var testevent = events.Event{
	EventType:          "myntra.prod.icinga.check_disk",
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
	ID:           "test-rule-id-1",
	HookEndpoint: "http://localhost:3000/testrule",
	HookRetry:    2,
	EventTypes:   []string{"myntra.prod.icinga.check_disk", "myntra.prod.site247.cart_down"},
	ScriptID:     "myscript",
}

var testRuleUpdated = rules.Rule{
	ID:           "test-rule-id-1",
	HookEndpoint: "http://localhost:3000/testrule",
	HookRetry:    2,
	EventTypes:   []string{"apple.prod.icinga.check_disk", "myntra.prod.site247.cart_down"},
	ScriptID:     "myscript",
}

func singleNode(t *testing.T, f func(node *Node)) {

	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	raftAddr := ":5878"
	httpAddr := ":5879"

	raftListener, err := net.Listen("tcp", raftAddr)
	if err != nil {
		t.Fatal(err)
	}

	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		t.Fatal(err)
	}

	// open store
	cfg := &config.Config{
		NodeID:               "node0",
		Dir:                  tmpDir,
		DefaultDwell:         4000, // 3 minutes
		DefaultMaxDwell:      8000, // 6 minutes
		DefaultDwellDeadline: 3800, // 2.5 minutes
		MaxHistory:           1000,
		FlushInterval:        1000,
		HTTPAddr:             httpAddr,
		RaftAddr:             raftAddr,
		HTTPListener:         httpListener,
		RaftListener:         raftListener,
	}

	node, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	glog.Infof("node started. 5s")
	// run test
	time.Sleep(time.Second * 5)
	glog.Infof("running test ")
	f(node)

	// close node
	err = node.Shutdown()
	if err != nil {
		t.Fatal(err)
	}

	err = httpListener.Close()

	glog.Info("done test ", err)
}

func TestRuleSingleNode(t *testing.T) {
	singleNode(t, func(node *Node) {

		err := node.AddRule(&testRule)
		if err != nil {
			t.Fatal(err)
		}

		rules := node.GetRules()
		found := false
		for _, rule := range rules {
			if rule.ID == testRule.ID {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("added rule  was not found")
		}

		err = node.UpdateRule(&testRuleUpdated)
		if err != nil {
			t.Fatal(err)
		}

		updatedRule := node.GetRule(testRule.ID)
		if updatedRule.EventTypes[0] != testRuleUpdated.EventTypes[0] {
			t.Fatal("rule was not updated")
		}

		err = node.RemoveRule(testRule.ID)
		if err != nil {
			t.Fatal(err)
		}

		rules = node.GetRules()
		found = false
		for _, rule := range rules {
			if rule.ID == testRule.ID {
				found = true
				break
			}
		}
		if found {
			t.Fatal("removed rule was found")
		}

	})
}

func TestScriptSingleNode(t *testing.T) {
	singleNode(t, func(node *Node) {
		script := []byte(`
			let result = 0;
			export default function() { result++; }`)

		// add script
		err := node.AddScript(&js.Script{ID: "myscript", Data: script})
		if err != nil {
			t.Fatal(err)
		}

		// get script

		respScript := node.GetScript("myscript")

		if !bytes.Equal(script, respScript.Data) {
			t.Fatal("unexpected get script response")
		}

		// remove script

		err = node.RemoveScript("myscript")
		if err != nil {
			t.Fatal(err)
		}

		// get script

		respScript = node.GetScript("myscript")

		if respScript != nil {
			t.Fatal("received removed script")
		}

	})
}

func TestOrphanEventSingleNode(t *testing.T) {
	singleNode(t, func(node *Node) {
		err := node.Stash(&testevent)
		if err != nil {
			t.Fatal(err)
		}

		var rb *events.Bucket
	loop:
		for {
			select {
			case rb = <-node.store.executionBucketQueue:
				fmt.Println("rb=>", rb)

			case <-time.After(time.Millisecond * time.Duration(node.store.opt.DefaultDwell+1000)):
				break loop
			}

		}

		if rb != nil {
			t.Fatal("orphan event was stashed")
		}
	})
}

func TestEventSingleNode(t *testing.T) {
	singleNode(t, func(node *Node) {

		err := node.AddRule(&testRule)
		if err != nil {
			t.Fatal(err)
		}

		err = node.Stash(&testevent)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond * time.Duration(node.store.opt.DefaultDwell+3000))
		records := node.GetRuleExectutions(testRule.ID)
		if len(records) == 0 {
			t.Fatal("no record of execution, event was not stashed")
		}
		if records[0].Bucket.Rule.ID != testRule.ID {
			t.Fatalf("unexpected rule id, event was not stashed %v %v", records[0].Bucket.Rule.ID, testRule.ID)
		}

		t.Logf("%+v\n", records[0])
	})
}

func TestNodeSnapshot(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	raftAddr := ":5878"
	httpAddr := ":5879"

	raftListener, err := net.Listen("tcp", raftAddr)
	if err != nil {
		t.Fatal(err)
	}

	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		t.Fatal(err)
	}

	// open store
	cfg := &config.Config{
		NodeID:               "node0",
		Dir:                  tmpDir,
		DefaultDwell:         4000, // 3 minutes
		DefaultMaxDwell:      8000, // 6 minutes
		DefaultDwellDeadline: 3800, // 2.5 minutes
		MaxHistory:           1000,
		FlushInterval:        1000,
		HTTPAddr:             httpAddr,
		RaftAddr:             raftAddr,
		HTTPListener:         httpListener,
		RaftListener:         raftListener,
	}

	node, err := NewNode(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	glog.Infof("node started. 5s")
	// run test
	time.Sleep(time.Second * 5)

	script := []byte(`
	let result = 0;
	export default function() { result++; }`)

	// add script
	err = node.AddScript(&js.Script{ID: "myscript", Data: script})
	if err != nil {
		t.Fatal(err)
	}

	err = node.AddRule(&testRule)
	if err != nil {
		t.Fatal(err)
	}

	rule := node.GetRule(testRule.ID)
	if testRule.ID != rule.ID {
		t.Fatal("rule not saved")
	}

	err = node.Stash(&testevent)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * time.Duration(node.store.opt.DefaultDwell+5000))

	glog.Infof("take a snapshot")

	err = node.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 2)
	// close node <===================
	err = node.Shutdown()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 2)

	raftListener, err = net.Listen("tcp", raftAddr)
	if err != nil {
		t.Fatal(err)
	}

	cfg.RaftListener = raftListener

	// start again ==================>
	err = node.Start()
	if err != nil {
		t.Fatal(err)
	}

	glog.Infof("node started. 5s")
	// run test
	time.Sleep(time.Second * 5)

	rule = node.GetRule(testRule.ID)
	if testRule.ID != rule.ID {
		t.Fatal("rule not saved")
	}

	respScript := node.GetScript("myscript")
	if respScript == nil {
		t.Fatal("script not found")
	}
	if !bytes.Equal(script, respScript.Data) {
		t.Fatal("unexpected get script response")
	}

	records := node.GetRuleExectutions(testRule.ID)
	if len(records) == 0 {
		t.Fatal("no record of execution, event was not stashed")
	}
	if records[0].Bucket.Rule.ID != testRule.ID {
		t.Fatalf("unexpected rule id, event was not stashed %v %v", records[0].Bucket.Rule.ID, testRule.ID)
	}

	// close node
	err = node.Shutdown()
	if err != nil {
		t.Fatal(err)
	}

	err = httpListener.Close()

	glog.Info("done test ", err)
}
