package matcher

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var matcherTests = []struct {
	pattern   string // rule pattern
	eventType string // event.EventType
	expected  bool   // expected result
}{
	{"acme*", "acme", false},
	{"acme*", "acme.prod", true},
	{"acme.prod*", "acme.prod.search", true},
	{"acme.prod*.checkout", "acme.prod.search", false},
	{"acme.prod*.*", "acme.prod.search", false},
	{"acme.prod*.*", "acme.prod-1.search", true},
	{"acme.prod.*.*.*", "acme.prod.search.node1.check_disk", true},
	{"acme.prod.*.*.check_disk", "acme.prod.search.node1.check_disk", true},
	{"acme.prod.*.*.check_loadavg", "acme.prod.search.node1.check_disk", false},
	{"*.prod.*.*.check_loadavg", "acme.prod.search.node1.check_loadavg", true},
	{"acme.prod.*", "acme.prod.search.node1.check_disk", true},
	{"acme.prod.search.node*.check_disk", "acme.prod.search.node1.check_disk", true},
	{"acme.prod.search.node*.*", "acme.prod.search.node1.check_disk", true},
	{"acme.prod.search.dc1-node*.*", "acme.prod.search.node1.check_disk", false},
}

func TestMatchers(t *testing.T) {
	for _, tc := range matcherTests {
		t.Run(fmt.Sprintf("Test if(%v==%v)", tc.eventType, tc.pattern), func(t *testing.T) {
			m, err := New(tc.pattern)
			require.NoError(t, err)
			hasMatch := m.HasMatches(tc.eventType)
			require.Equal(t, tc.expected, hasMatch)
		})
	}
}
