package awsmocker

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/google/uuid"
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
		resp.Header.Add("X-Amzn-Requestid", generateRequestId())
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
		fmt.Fprintln(DebugOutputWriter, "--- AWSMOCKER RESPONSE: -------------------------------")
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			DebugOutputWriter.Write(dump)
		} else {
			fmt.Fprintf(DebugOutputWriter, "FAILED TO DUMP RESPONSE!: %s", err)
		}
		fmt.Fprintln(DebugOutputWriter)
		fmt.Fprintln(DebugOutputWriter, "-------------------------------------------------------")
	}

	return resp
}

// generate a request using a real UUID. if that fails, who cares this is a test
func generateRequestId() string {
	id, err := uuid.NewRandom()
	if err != nil {
		return "1b206dd1-f9a8-11e5-becf-051c60f11c4a"
	}

	return id.String()
}
