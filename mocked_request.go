package awsmocker

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Describes a request that should be matched
type MockedRequest struct {
	// Require that fields are matched exactly
	//
	// Nonstrict (default) means that Params listed are matched against
	// the request to ensure the ones specified match
	//
	// Strict mode requires that the request contain ONLY the params listed
	// any extra parameters will cause the request to fail to match
	Strict bool

	// The hostname only. Does not include the port
	Hostname string

	// The AWS service shortcode
	Service string

	// The AWS API Action being performed
	Action string

	// Body to match against
	Body string

	// Match against specific parameters in the request.
	// This is only used for XML/Form requests (not the newer JSON ones)
	Params url.Values

	// Match a specific HTTP method
	Method string

	// Match the URL path
	Path string

	// Match the URL path, using a regex
	PathRegex *regexp.Regexp

	// Is this an instance metadata request?
	// setting this to true will match against both the IPv4 and IPv6 hostnames
	IsEc2IMDS bool

	// Matches a JSON request body by resolving the jmespath expression as keys
	// and comparing the values returned against the value provided in the map
	JMESPathMatches map[string]any

	// Write a custom matcher function that will be used to match a request.
	// this runs after checking the other fields, so you can use those as filters.
	Matcher func(*ReceivedRequest) bool

	// Stop matching this request after it has been matched X times
	//
	// 0 (default) means it will live forever
	MaxMatchCount int

	// number of times this request has matched
	matchCount int64
	mu         sync.Mutex
}

func (mr *MockedRequest) prep() {

}

func (m *MockedRequest) incMatchCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matchCount += 1
}

// Returns a string to help identify this MockedRequest
func (m *MockedRequest) Inspect() string {
	parts := make([]string, 0, 10)

	if m.Strict {
		parts = append(parts, "STRICT")
	}

	if m.Service != "" {
		parts = append(parts, fmt.Sprintf("Service=%s", m.Service))
	}

	if m.Action != "" {
		parts = append(parts, fmt.Sprintf("Action=%s", m.Action))
	}

	if m.IsEc2IMDS {
		parts = append(parts, fmt.Sprintf("imds=%t", m.IsEc2IMDS))
	}

	if m.Hostname != "" {
		parts = append(parts, fmt.Sprintf("Hostname=%s", m.Hostname))
	}

	if m.Path != "" {
		parts = append(parts, fmt.Sprintf("Path=%s", m.Path))
	}

	if m.Method != "" {
		parts = append(parts, fmt.Sprintf("Method=%s", m.Method))
	}

	if m.PathRegex != nil {
		parts = append(parts, fmt.Sprintf("PathRegex=%s", m.PathRegex.String()))
	}

	if len(m.Params) > 0 {
		parts = append(parts, fmt.Sprintf("Params=%s", m.Params.Encode()))
	}

	if m.Body != "" {
		parts = append(parts, fmt.Sprintf("Body=%s", m.Body))
	}

	return "MReq<" + strings.Join(parts, " ") + ">"
}

func (m *MockedRequest) matchRequest(rr *ReceivedRequest) bool {

	if m.MaxMatchCount > 0 && m.matchCount >= int64(m.MaxMatchCount) {
		return false
	}

	if !m.matchRequestLazy(rr) {
		return false
	}

	if m.Strict {
		return m.matchRequestStrict(rr)
	}

	return true
}

func (m *MockedRequest) matchRequestLazy(rr *ReceivedRequest) bool {

	if m.Hostname != "" && rr.Hostname != m.Hostname {
		return false
	}

	if m.Service != "" && rr.Service != m.Service {
		return false
	}

	if m.Action != "" && rr.Action != m.Action {
		return false
	}

	if m.Path != "" && rr.Path != m.Path {
		return false
	}

	if m.Method != "" && rr.HttpRequest.Method != m.Method {
		return false
	}

	if m.IsEc2IMDS && !(rr.Hostname == imdsHost4 || rr.Hostname == imdsHost6) {
		return false
	}

	if m.PathRegex != nil && !m.PathRegex.MatchString(rr.Path) {
		return false
	}

	if m.JMESPathMatches != nil && len(m.JMESPathMatches) > 0 {
		if ret := m.matchJmespath(rr); !ret {
			return false
		}
	}

	if m.Matcher != nil && !m.Matcher(rr) {
		return false
	}

	return true
}

func (m *MockedRequest) matchJmespath(rr *ReceivedRequest) bool {
	// just bail out if there is nothing to match
	if m.JMESPathMatches == nil || len(m.JMESPathMatches) == 0 {
		return true
	}

	// you provided Jmes matchers, but this isnt a JSON payload, so it will never match
	if rr.JsonPayload == nil {
		return false
	}

	for k, v := range m.JMESPathMatches {
		if !JMESMatch(rr.JsonPayload, k, v) {
			// if any are false, then bail out checking
			return false
		}
	}

	return true
}

func (m *MockedRequest) matchRequestStrict(rr *ReceivedRequest) bool {
	// assume the lazy check has already run

	return maps.EqualFunc(rr.HttpRequest.Form, m.Params, func(v1, v2 []string) bool {
		return slices.Equal(v1, v2)
	})
}
