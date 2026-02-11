package awsmocker

type MockedEndpoint struct {
	Request  *MockedRequest
	Response *MockedResponse
}

func (m *MockedEndpoint) prep() {
	m.Response.action = m.Request.Action

	m.Response.prep()
	m.Request.prep()

}

func (m *MockedEndpoint) matchRequest(rr *ReceivedRequest) bool {
	return m.Request.matchRequest(rr)
}

func (m *MockedEndpoint) getResponse(rr *ReceivedRequest) *httpResponse {
	return m.Response.getResponse(rr)
}

// Generates a simple [MockedEndpoint] for the Service:Action
func NewSimpleMockedEndpoint(service, action string, responseObj any) *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Service: service,
			Action:  action,
		},
		Response: &MockedResponse{
			Body: responseObj,
		},
	}
}
