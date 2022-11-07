package awsmocker

import (
	"net/http"
)

func (m *mocker) handleHttp(w http.ResponseWriter, r *http.Request) {
	// hostname := r.URL.Hostname()

	// if hostname == imdsHost4 || hostname == imdsHost6 {

	// 	m.handleAwsRequestHttp(w, r)
	// 	return
	// }

	// // m.proxyPassthruHttp(nil, nil)

	// w.WriteHeader(501)
	_, resp := m.handleRequest(r)

	resp.Write(w)
}

func (m *mocker) handleAwsRequestHttp(w http.ResponseWriter, r *http.Request) {
	// path := r.URL.Path

	// path = "/" + strings.TrimPrefix(path, "/")

	_, resp := m.handleRequest(r)

	resp.Write(w)
}
