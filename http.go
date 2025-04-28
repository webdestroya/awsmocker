package awsmocker

import (
	"io"
	"net/http"
)

func (m *mocker) handleHttp(w http.ResponseWriter, r *http.Request) {

	_, resp := m.handleRequest(r)

	origBody := resp.Body
	defer origBody.Close()

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	if err := resp.Body.Close(); err != nil {
		m.Warnf("Can't close response body %v", err)
	}
}

var handleNonProxyRequest = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "AWSMocker is meant to be used as a proxy server. Don't send requests directly to it.", http.StatusNotImplemented)
})
