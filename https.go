package awsmocker

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
)

const (
	imdsHost4 = "169.254.169.254"
	imdsHost6 = "fd00:ec2::254"
)

var (
	awsDomainRegexp = regexp.MustCompile(`amazonaws\.com`)
	hasPort         = regexp.MustCompile(`:\d+$`)
	httpsRegexp     = regexp.MustCompile(`^https:\/\/`)

	globalTlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return globalCertStore.Fetch(chi.ServerName), nil
		},
	}
)

func (m *mocker) handleHttps(w http.ResponseWriter, r *http.Request) {
	hij, ok := w.(http.Hijacker)
	if !ok {
		panic("httpserver does not support hijacking")
	}

	proxyClient, _, e := hij.Hijack()
	if e != nil {
		panic("Cannot hijack connection " + e.Error())
	}

	hostname := r.URL.Hostname()

	if awsDomainRegexp.MatchString(hostname) {
		// PROXY HACK
		// httpError(proxyClient, nil)
		_, _ = proxyClient.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
		go m.handleAwsRequestHttps(proxyClient, r)
		return
	}

	// TODO: add additional url handlers if desired

	// No match, so pass through
	// m.Logf("AwsMocker HTTPS PASSTHROUGH! %s", hostname)
	// m.proxyPassthruHttps(proxyClient, r)
	httpErrorCode(proxyClient, 501)

}

func (m *mocker) handleAwsRequestHttps(proxyClient net.Conn, r *http.Request) {
	rawClientTls := tls.Server(proxyClient, globalTlsConfig)
	if err := rawClientTls.Handshake(); err != nil {
		m.Warnf("Cannot handshake client %v %v", r.Host, err)
		return
	}
	defer rawClientTls.Close()

	clientTlsReader := bufio.NewReader(rawClientTls)
	for !isEof(clientTlsReader) {
		req, err := http.ReadRequest(clientTlsReader)
		// var ctx = &ProxyCtx{Req: req, Session: atomic.AddInt64(&proxy.sess, 1), Proxy: proxy, UserData: ctx.UserData}
		if err != nil && !errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			m.Warnf("Cannot read TLS request from mitm'd client %v %v", r.Host, err)
			return
		}
		req.RemoteAddr = r.RemoteAddr

		if !httpsRegexp.MatchString(req.URL.String()) {
			req.URL, _ = url.Parse("https://" + r.Host + req.URL.String())
		}

		_, resp := m.handleRequest(req)

		defer resp.Body.Close()

		resp.Header.Set("Connection", "close")

		if err := resp.Write(rawClientTls); err != nil {
			m.Warnf("Failed to write response: %s", err)
		}
	}
}

/*
func (m *mocker) proxyPassthruHttps(proxyClient net.Conn, r *http.Request) {
	host := r.URL.Host
	if !hasPort.MatchString(host) {
		host += ":80"
	}
	targetSiteCon, err := net.DialTimeout("tcp", host, m.timeout)
	if err != nil {
		httpError(proxyClient, err)
		return
	}
	m.Logf("Accepting CONNECT to %s", host)
	_, _ = proxyClient.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))

	targetTCP, targetOK := targetSiteCon.(halfClosable)
	proxyClientTCP, clientOK := proxyClient.(halfClosable)
	if targetOK && clientOK {
		go copyAndClose(targetTCP, proxyClientTCP)
		go copyAndClose(proxyClientTCP, targetTCP)
	} else {
		go func() {
			var wg sync.WaitGroup
			wg.Add(2)
			go copyOrWarn(targetSiteCon, proxyClient, &wg)
			go copyOrWarn(proxyClient, targetSiteCon, &wg)
			wg.Wait()
			proxyClient.Close()
			targetSiteCon.Close()

		}()
	}
}
*/
