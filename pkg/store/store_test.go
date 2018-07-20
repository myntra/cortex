package store

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/myntra/aggo/pkg/event"
)

func singleNode(t *testing.T, f func(store *defaultStore)) {

	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)
	// open store
	cfg := &Config{
		NodeID:                     "node0",
		BindAddr:                   "127.0.0.1:8878",
		Dir:                        tmpDir,
		DefaultWaitWindow:          2000, // 3 minutes
		DefaultMaxWaitWindow:       4000, // 6 minutes
		DefaultWaitWindowThreshold: 1500, // 2.5 minutes
	}

	store, err := newStore(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// run test
	time.Sleep(time.Second * 3)

	f(store)

	// close store
	if err = store.close(); err != nil {
		t.Fatal(err)
	}
}

func TestRuleSingleNode(t *testing.T) {
	singleNode(t, func(store *defaultStore) {

		testRule := event.Rule{
			ID:           "test-rule-id-1",
			HookEndpoint: "http://localhost:3000/testrule",
			HookRetry:    2,
			EventTypes:   []string{"myntra.prod.icinga.check_disk", "myntra.prod.site247.cart_down"},
		}

		err := store.AddRule(&testRule)
		if err != nil {
			t.Fatal(err)
		}

		rules := store.GetRules()
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

		err = store.RemoveRule(testRule.ID)
		if err != nil {
			t.Fatal(err)
		}

		rules = store.GetRules()
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
