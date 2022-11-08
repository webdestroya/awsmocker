package awsmocker

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type httpResponse struct {
	Header     http.Header
	StatusCode int
	Body       string

	bodyRaw []byte

	contentType string

	extraHeaders map[string]string

	forcedHttpResponse *http.Response

	// isError bool
}

/*
func (hr *httpResponse) notifyIfError(m *mocker) *httpResponse {
	if hr.isError {
		m.t.Errorf("AWSMocker errored during the test")
	}
	return hr
}
*/

func (hr *httpResponse) toHttpResponse(req *http.Request) *http.Response {

	if hr.forcedHttpResponse != nil {
		return hr.forcedHttpResponse
	}

	resp := &http.Response{
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Request:          req,
		TransferEncoding: req.TransferEncoding,
		Header:           hr.Header,
		StatusCode:       hr.StatusCode,
	}

	if resp.ProtoMajor == 0 {
		resp.ProtoMajor = 1
	}

	if resp.Header == nil {
		resp.Header = make(http.Header)
	}

	resp.Header.Add("Content-Type", hr.contentType)
	resp.Header.Add("Server", "AWSMocker")

	if hr.extraHeaders != nil && len(hr.extraHeaders) > 0 {
		for k, v := range hr.extraHeaders {
			resp.Header.Set(k, v)
		}
	}

	if x := resp.Header.Get("Date"); x == "" {
		resp.Header.Add("Date", time.Now().Format(http.TimeFormat))
	}

	if x := resp.Header.Get("X-Amzn-Requestid"); x == "" {
		resp.Header.Add("X-Amzn-Requestid", "1b206dd1-f9a8-11e5-becf-051c60f11c4a")
	}

	resp.Status = http.StatusText(resp.StatusCode)

	var buf *bytes.Buffer
	if len(hr.bodyRaw) > 0 {
		buf = bytes.NewBuffer(hr.bodyRaw)
	} else {
		buf = bytes.NewBufferString(hr.Body)
	}

	resp.ContentLength = int64(buf.Len())
	resp.Body = io.NopCloser(buf)

	if GlobalDebugMode {
		fmt.Println("MOCK RESPONSE: -----------------------------")
		fmt.Printf("HTTP/%d.%d %d %s\n", resp.ProtoMajor, resp.ProtoMinor, hr.StatusCode, resp.Status)
		for k, vlist := range resp.Header {
			for _, v := range vlist {
				fmt.Printf("%s: %s\n", k, v)
			}
		}
		fmt.Printf("Content-Length: %d\n", resp.ContentLength)
		fmt.Println()
		fmt.Println(hr.Body)
		fmt.Println("--------------------------------------------")
	}

	return resp
}
