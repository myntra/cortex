package main

import "testing"

var ruleTests = []struct {
	in  string
	out bool
}{
	{"metrics.myntra.com", true},
	{"metrics.deasas.com", true},
	{"metrics.myntra.assass", true},
	{"metrics1.myntra.com", false},
	{"metrics.myntra12212121.com", true},
	{"metrics.myntra.com12112", true},
	{"metrics.myntra12212121.asaas.com", true},
	{"metrics.myntra.com.assaasa", false},
	{"metrics.myntra.com.a", false},
	{"metrics.qwqq.com", true},
	{"metrics.mynt11ra.com", true},
}

var ruleType = "metrics.myntra.com"

func Test_regexMatcher1(t *testing.T) {
	for _, tt := range ruleTests {
		got := regexMatcher(ruleType, tt.in)
		if got != tt.out {
			t.Errorf("regexMatcher(%s): expected %v, actual %v", tt.in, tt.out, got)
		}
	}
}
