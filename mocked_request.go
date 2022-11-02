package awsmocker

import "net/url"

type MockedRequest struct {
	// Require that fields are matched exactly
	Strict  bool
	Service string
	Body    string
	Action  string
	Params  url.Values
	Path    string
}

func (mr *MockedRequest) prep() {

}

func (m *MockedRequest) matchRequest(rr *receivedRequest) bool {

	if !m.matchRequestLazy(rr) {
		return false
	}

	if m.Strict {
		return m.matchRequestStrict(rr)
	}

	return true
}

func (m *MockedRequest) matchRequestLazy(rr *receivedRequest) bool {
	if m.Service != "" && rr.service != m.Service {
		return false
	}

	if m.Action != "" && rr.action != m.Action {
		return false
	}

	if m.Path != "" && rr.request.RequestURI != m.Path {
		return false
	}

	return true
}

func (m *MockedRequest) matchRequestStrict(rr *receivedRequest) bool {
	// assume the lazy check has already run

	return rr.request.Form.Encode() == m.Params.Encode()
}
