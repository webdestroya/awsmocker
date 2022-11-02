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

	contentType string

	extraHeaders map[string]string
}

func (hr *httpResponse) toHttpResponse(req *http.Request) *http.Response {
	resp := &http.Response{
		Request:          req,
		TransferEncoding: req.TransferEncoding,
		Header:           hr.Header,
		StatusCode:       hr.StatusCode,
	}

	if resp.Header == nil {
		resp.Header = make(http.Header)
	}

	resp.Header.Add("Content-Type", hr.contentType)

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
	buf := bytes.NewBufferString(hr.Body)
	resp.ContentLength = int64(buf.Len())
	resp.Body = io.NopCloser(buf)

	if GlobalDebugMode {
		fmt.Println("MOCK RESPONSE: -----------------------------")
		fmt.Printf("HTTP/1.1 %d %s\n", hr.StatusCode, resp.Status)
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
