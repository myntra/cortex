package matcher

import (
	"fmt"
	"regexp"
	"strings"
)

var metricLineRE = regexp.MustCompile(`^(\*\.|[^.]+\.|\.)*(\*|[^.]+)$`)

// Matcher matches a rule.EventTypePatterns patterns with eventTypePatterns
type Matcher struct {
	regex *regexp.Regexp
}

// New accepts a rulePattern
func New(rulePattern string) (*Matcher, error) {
	regex, err := getRegexp(rulePattern)
	if err != nil {
		return nil, err
	}

	m := &Matcher{
		regex: regex,
	}

	return m, nil
}

// NewCompile accepts a regex string
func NewCompile(regexStr string) *Matcher {
	return &Matcher{
		regex: regexp.MustCompile(regexStr),
	}
}

// GetRegexString returns the compiled regex string
func (m *Matcher) GetRegexString() string {
	return m.regex.String()
}

// HasMatches checks if eventType has matches with the supplied regex
func (m *Matcher) HasMatches(eventType string) bool {
	matches := m.regex.FindStringSubmatchIndex(eventType)
	if len(matches) > 0 {
		return true
	}
	return false
}

// getRegexp returns a *regexp.Regexp for the pattern
// reference: https://github.com/prometheus/graphite_exporter/blob/master/mapper.go#L65
func getRegexp(rulePattern string) (*regexp.Regexp, error) {
	var regex *regexp.Regexp

	if !metricLineRE.MatchString(rulePattern) {
		return nil, fmt.Errorf("unexpected pattern %v. must match %v", rulePattern, metricLineRE.String())
	}

	rulePatternRe := strings.Replace(rulePattern, ".", "\\.", -1)
	rulePatternRe = strings.Replace(rulePatternRe, "*", "([^*]+)", -1)
	regex = regexp.MustCompile("^" + rulePatternRe + "$")
	return regex, nil
}
