package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/golang/glog"

	"github.com/myntra/cortex/pkg/config"
	"github.com/myntra/cortex/pkg/events"
	"github.com/myntra/cortex/pkg/js"
	"github.com/myntra/cortex/pkg/rules"
)

var testevent = events.Event{
	EventType:          "acme.prod.icinga.check_disk",
	EventTypeVersion:   "1.0",
	CloudEventsVersion: "0.1",
	Source:             "/sink",
	EventID:            "42",
	EventTime:          time.Now(),
	SchemaURL:          "http://www.json.org",
	ContentType:        "application/json",
	Data:               map[string]interface{}{"Alpha": "julie", "Beta": 42},
	Extensions:         map[string]string{"ext1": "value"},
}

var testRule = rules.Rule{
	ID:                "test-rule-id-1",
	HookEndpoint:      "http://localhost:3000/testrule",
	HookRetry:         2,
	EventTypePatterns: []string{"acme.prod.icinga.check_disk", "acme.prod.site247.cart_down"},
	ScriptID:          "myscript",
}

var testRuleUpdated = rules.Rule{
	ID:                "test-rule-id-1",
	HookEndpoint:      "http://localhost:3000/testrule",
	HookRetry:         2,
	EventTypePatterns: []string{"apple.prod.icinga.check_disk", "acme.prod.site247.cart_down"},
	ScriptID:          "myscript",
}

func newTestEvent(id, key string) events.Event {
	return events.Event{
		EventType:          key + "acme.prod.icinga.check_disk",
		EventTypeVersion:   "1.0",
		CloudEventsVersion: "0.1",
		Source:             "/sink",
		EventID:            id + "42",
		EventTime:          time.Now(),
		SchemaURL:          "http://www.json.org",
		ContentType:        "application/json",
		Data:               map[string]interface{}{id + "Alpha": "julie", "Beta": 42},
		Extensions:         map[string]string{"ext1": "value"},
	}
}

func newTestRule(key string) rules.Rule {
	return rules.Rule{
		ID:                key + "test-rule-id-1",
		HookEndpoint:      "http://localhost:3000/testrule",
		HookRetry:         2,
		EventTypePatterns: []string{key + "acme.prod.icinga.check_disk", key + "acme.prod.site247.cart_down"},
		ScriptID:          "myscript",
		Dwell:             30 * 1000,
		DwellDeadline:     20 * 1000,
		MaxDwell:          90 * 1000,
	}
}

func singleNode(t *testing.T, httpAddr, raftAddr string, f func(node *Node)) {

	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	raftListener, err := net.Listen("tcp", raftAddr)
	require.NoError(t, err)

	httpListener, err := net.Listen("tcp", httpAddr)
	require.NoError(t, err)

	// open store
	cfg := &config.Config{
		NodeID:               "node0",
		Dir:                  tmpDir,
		DefaultDwell:         4000,
		DefaultMaxDwell:      8000,
		DefaultDwellDeadline: 3800,
		MaxHistory:           1000,
		FlushInterval:        1000,
		SnapshotInterval:     30,
		HTTPAddr:             httpAddr,
		RaftAddr:             raftAddr,
		HTTPListener:         httpListener,
		RaftListener:         raftListener,
	}

	node, err := NewNode(cfg)
	require.NoError(t, err)

	err = node.Start()
	require.NoError(t, err)

	glog.Infof("node started. 5s")
	// run test
	time.Sleep(time.Second * 5)
	glog.Infof("running test ")
	f(node)

	// close node
	err = node.Shutdown()
	require.NoError(t, err)

	err = httpListener.Close()
	require.NoError(t, err)
	glog.Info("done test ")
}

func TestRuleSingleNode(t *testing.T) {
	raftAddr := ":3878"
	httpAddr := ":3879"
	singleNode(t, httpAddr, raftAddr, func(node *Node) {

		err := node.AddRule(&testRule)
		require.NoError(t, err)

		rule := node.GetRule(testRule.ID)
		require.True(t, rule.ID == testRule.ID)

		err = node.UpdateRule(&testRuleUpdated)
		require.NoError(t, err)

		updatedRule := node.GetRule(testRule.ID)
		require.True(t, updatedRule.EventTypePatterns[0] == testRuleUpdated.EventTypePatterns[0])

		err = node.RemoveRule(testRule.ID)
		require.NoError(t, err)

		rule = node.GetRule(testRule.ID)
		require.Nil(t, rule)

	})
}

func TestScriptSingleNode(t *testing.T) {
	raftAddr := ":4878"
	httpAddr := ":4879"
	singleNode(t, httpAddr, raftAddr, func(node *Node) {
		script := []byte(`
			let result = 0;
			export default function() { result++; }`)

		// add script
		err := node.AddScript(&js.Script{ID: "myscript", Data: script})
		require.NoError(t, err)

		// get script

		respScript := node.GetScript("myscript")
		require.True(t, bytes.Equal(script, respScript.Data))

		// remove script

		err = node.RemoveScript("myscript")
		require.NoError(t, err)

		// get script
		respScript = node.GetScript("myscript")
		require.Nil(t, respScript)

	})
}

func TestOrphanEventSingleNode(t *testing.T) {
	raftAddr := ":5878"
	httpAddr := ":5879"
	singleNode(t, httpAddr, raftAddr, func(node *Node) {
		err := node.Stash(&testevent)
		require.NoError(t, err)

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

		require.Nil(t, rb)
	})
}

func TestEventSingleNode(t *testing.T) {
	raftAddr := ":6878"
	httpAddr := ":6879"
	singleNode(t, httpAddr, raftAddr, func(node *Node) {

		err := node.AddRule(&testRule)
		require.NoError(t, err)
		err = node.Stash(&testevent)
		require.NoError(t, err)

		time.Sleep(time.Millisecond * time.Duration(node.store.opt.DefaultDwell+3000))
		records := node.GetRuleExectutions(testRule.ID)
		require.False(t, len(records) == 0)
		require.True(t, records[0].Bucket.Rule.ID == testRule.ID)

		t.Logf("%+v\n", records[0])
	})
}

func TestMultipleEventSingleRule(t *testing.T) {
	raftAddr := ":7878"
	httpAddr := ":7879"
	singleNode(t, httpAddr, raftAddr, func(node *Node) {

		t.Run("Test stash multiple events before dwell time", func(t *testing.T) {
			key := "my"
			myTestRule := newTestRule("my")
			n := 5

			s := rand.NewSource(time.Now().UnixNano())
			r := rand.New(s)
			intervals := make(map[int]events.Event)

			for i := 0; i < n; i++ {
				interval := r.Intn(int(myTestRule.DwellDeadline - 100))
				intervals[interval] = newTestEvent(strconv.Itoa(i), key)
			}

			err := node.AddRule(&myTestRule)
			require.NoError(t, err)

			for interval, te := range intervals {
				go func(interval int, te events.Event) {
					time.Sleep(time.Millisecond * time.Duration(interval))
					err = node.Stash(&te)
					require.NoError(t, err)
				}(interval, te)
			}

			glog.Info("sleeping ...")
			time.Sleep(time.Millisecond * time.Duration(myTestRule.Dwell+10000))
			glog.Info("sleeping done")

			records := node.GetRuleExectutions(myTestRule.ID)
			require.True(t, len(records) == 1, fmt.Sprintf("len is %v", len(records)))
			require.True(t, len(records[0].Bucket.Events) == n, fmt.Sprintf("len is %v", len(records[0].Bucket.Events)))
		})

		t.Run("Test stash multiple events after dwell time", func(t *testing.T) {
			key := "hi"
			myTestRule := newTestRule("hi")
			n := 5

			s := rand.NewSource(time.Now().UnixNano())
			r := rand.New(s)
			intervals := make(map[int]events.Event)

			// before dwell deadline
			for i := 0; i < n; i++ {
				interval := int(myTestRule.DwellDeadline) + 1000*i
				intervals[interval] = newTestEvent(strconv.Itoa(i), key)
			}

			// after dwell deadline
			for i := 5; i < 10; i++ {
				interval := r.Intn(int(myTestRule.Dwell - myTestRule.DwellDeadline))
				intervals[interval] = newTestEvent(strconv.Itoa(i), key)
			}

			// 5 events will be deduped
			for i := 5; i < 10; i++ {
				interval := r.Intn(int(myTestRule.Dwell - myTestRule.DwellDeadline))
				intervals[interval] = newTestEvent(strconv.Itoa(i), key)
			}

			for k := range intervals {
				glog.Infof("intervals %v\n", k)
			}
			err := node.AddRule(&myTestRule)
			require.NoError(t, err)

			for interval, te := range intervals {
				go func(interval int, te events.Event) {
					time.Sleep(time.Millisecond * time.Duration(interval))
					glog.Info("send event ", time.Millisecond*time.Duration(interval))
					err = node.Stash(&te)
					require.NoError(t, err)
				}(interval, te)
			}

			glog.Info("sleeping ...")
			time.Sleep(time.Millisecond * time.Duration(myTestRule.MaxDwell+3000))
			glog.Info("sleeping done")

			records := node.GetRuleExectutions(myTestRule.ID)
			require.True(t, len(records) == 1, fmt.Sprintf("len is %v", len(records)))
			require.True(t, len(records[0].Bucket.Events) == 10, fmt.Sprintf("len is %v", len(records[0].Bucket.Events)))
		})

	})
}

func TestNodeSnapshot(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)

	raftAddr := ":8878"
	httpAddr := ":8879"

	raftListener, err := net.Listen("tcp", raftAddr)
	require.NoError(t, err)
	httpListener, err := net.Listen("tcp", httpAddr)
	require.NoError(t, err)

	// open store
	cfg := &config.Config{
		NodeID:               "node0",
		Dir:                  tmpDir,
		DefaultDwell:         4000,
		DefaultMaxDwell:      8000,
		DefaultDwellDeadline: 3800,
		MaxHistory:           1000,
		FlushInterval:        1000,
		SnapshotInterval:     30,
		HTTPAddr:             httpAddr,
		RaftAddr:             raftAddr,
		HTTPListener:         httpListener,
		RaftListener:         raftListener,
	}

	node, err := NewNode(cfg)
	require.NoError(t, err)

	err = node.Start()
	require.NoError(t, err)

	glog.Infof("node started. 5s")
	// run test
	time.Sleep(time.Second * 5)

	script := []byte(`
	let result = 0;
	export default function() { result++; }`)

	// add script
	err = node.AddScript(&js.Script{ID: "myscript", Data: script})
	require.NoError(t, err)
	err = node.AddRule(&testRule)
	require.NoError(t, err)

	rule := node.GetRule(testRule.ID)
	require.True(t, testRule.ID == rule.ID)

	err = node.Stash(&testevent)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * time.Duration(node.store.opt.DefaultDwell+5000))

	glog.Infof("take a snapshot")

	err = node.Snapshot()
	require.NoError(t, err)

	time.Sleep(time.Second * 2)
	// close node <===================
	err = node.Shutdown()
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	raftListener, err = net.Listen("tcp", raftAddr)
	require.NoError(t, err)

	cfg.RaftListener = raftListener

	// start again ==================>
	err = node.Start()
	require.NoError(t, err)

	glog.Infof("node started. 5s")
	// run test
	time.Sleep(time.Second * 5)

	rule = node.GetRule(testRule.ID)
	require.True(t, testRule.ID == rule.ID)

	respScript := node.GetScript("myscript")
	require.NotNil(t, respScript)
	require.True(t, bytes.Equal(script, respScript.Data))

	records := node.GetRuleExectutions(testRule.ID)
	require.False(t, len(records) == 0)
	require.True(t, records[0].Bucket.Rule.ID == testRule.ID)

	// close node
	err = node.Shutdown()
	require.NoError(t, err)

	err = httpListener.Close()
	require.NoError(t, err)
}
