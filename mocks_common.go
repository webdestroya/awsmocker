package awsmocker

import "net/http"

// Returns an error response for a given service/action call
func Mock_Failure(service, action string) *MockedEndpoint {
	return Mock_Failure_WithCode(0, service, action, "AccessDenied", "This mock was requested to fail")
}

func Mock_Failure_WithCode(statusCode int, service, action, errorCode, errorMessage string) *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Service: service,
			Action:  action,
		},
		Response: MockResponse_Error(statusCode, errorCode, errorMessage),
	}
}

// Returns an error response with a custom code and message
func MockResponse_Error(statusCode int, errorCode, errorMessage string) *MockedResponse {
	errObj := generateErrorStruct(errorCode, errorMessage)
	if statusCode == 0 {
		statusCode = 400
	}
	return &MockedResponse{
		Handler: func(rr *ReceivedRequest) *http.Response {
			resp := errObj.getResponse(rr).toHttpResponse(rr.HttpRequest)
			resp.StatusCode = statusCode
			return resp
		},
	}
}
