package awsmocker

import (
	"net/url"
	"regexp"

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
}

func (mr *MockedRequest) prep() {

}

func (m *MockedRequest) matchRequest(rr *ReceivedRequest) bool {

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

	return true
}

func (m *MockedRequest) matchRequestStrict(rr *ReceivedRequest) bool {
	// assume the lazy check has already run

	return maps.EqualFunc(rr.HttpRequest.Form, m.Params, func(v1, v2 []string) bool {
		return slices.Equal(v1, v2)
	})
}
