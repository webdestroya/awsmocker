package awsmocker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var (
	credExtractRegexp = regexp.MustCompile(`Credential=(\S+)\b`)
	jsonRegexp        = regexp.MustCompile(`json`)
)

type ReceivedRequest struct {
	HttpRequest *http.Request

	Action  string
	Service string
	Region  string

	Hostname string
	Path     string

	// The expected response type based upon the request. JSON requests answered with JSON,
	// form param posts respond with XML
	AssumedResponseType string

	// This will only be populated if the request was NOT a form
	RawBody []byte

	// If the request was a JSON request, then this will be the parsed JSON
	JsonPayload any

	// TBA: maybe in the future we'll add invalid request flagging, for now allow all types
	// invalid bool
}

func (rr *ReceivedRequest) Inspect() string {

	if rr.Action != "" && rr.Service != "" {
		return rr.Service + ":" + rr.Action
	}

	return fmt.Sprintf("%s %s/%s", rr.HttpRequest.Method, rr.Hostname, rr.Path)
}

func newReceivedRequest(req *http.Request) *ReceivedRequest {
	recvreq := &ReceivedRequest{
		HttpRequest:         req,
		AssumedResponseType: ContentTypeText,
		Hostname:            req.URL.Hostname(),
		Path:                req.URL.Path,
	}

	_ = req.ParseForm()

	bodyBytes, err := io.ReadAll(req.Body)
	if err == nil {
		recvreq.RawBody = bodyBytes
	}

	reqContentType := req.Header.Get("content-type")
	if jsonRegexp.MatchString(reqContentType) {
		recvreq.AssumedResponseType = ContentTypeJSON

		if len(bodyBytes) > 0 {
			var jsonData any
			if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
				recvreq.JsonPayload = jsonData
			}
		}

	}

	authHeader := req.Header.Get("authorization")
	if authHeader != "" {
		matches := credExtractRegexp.FindStringSubmatch(authHeader)
		if len(matches) > 1 {
			// 0       1         2      3    4
			// fake/20221030/us-east-1/ecs/aws4_request
			parts := strings.Split(matches[1], "/")
			recvreq.Region = parts[2]
			recvreq.Service = parts[3]
		}
	}

	amzTarget := req.Header.Get("x-amz-target")
	if amzTarget != "" {
		// X-Amz-Target: AmazonEC2ContainerServiceV20141113.ListClusters
		_, opname, ok := strings.Cut(amzTarget, ".")
		if ok {
			recvreq.Action = opname
		}
	}

	if recvreq.Action == "" && req.PostForm.Has("Action") {
		recvreq.Action = req.PostForm.Get("Action")
	}

	if recvreq.AssumedResponseType == ContentTypeText && recvreq.Action != "" && recvreq.Service != "" && reqContentType == "application/x-www-form-urlencoded" {
		recvreq.AssumedResponseType = ContentTypeXML
	}

	// if recvreq.Action == "" {
	// 	log.Println("WARN: Received a request with no action????")
	// 	recvreq.invalid = true
	// 	return recvreq
	// }

	return recvreq
}

func (r *ReceivedRequest) DebugDump() {
	// var buf *bytes.Buffer
	buf := new(bytes.Buffer)

	fmt.Fprintln(buf, "--- AWSMOCKER RECEIVED REQUEST: -----------------------")

	if r.Action != "" || r.Service != "" {
		fmt.Fprintf(buf, "Operation: %s (service=%s @ %s)\n", r.Action, r.Service, r.Region)
	}

	fmt.Fprintf(buf, "%s %s\n", r.HttpRequest.Method, r.HttpRequest.RequestURI)
	fmt.Fprintf(buf, "Host: %s\n", r.HttpRequest.Host)
	for k, vlist := range r.HttpRequest.Header {
		for _, v := range vlist {
			fmt.Fprintf(buf, "%s: %s\n", k, v)
		}
	}

	fmt.Fprintln(buf)

	if len(r.RawBody) > 0 {
		fmt.Fprintln(buf, "BODY:")
		fmt.Fprintln(buf, string(r.RawBody))
	} else if r.HttpRequest.Form != nil && len(r.HttpRequest.Form) > 0 {
		fmt.Fprintln(buf, "PARAMS:")
		for k, vlist := range r.HttpRequest.Form {
			for _, v := range vlist {
				fmt.Fprintf(buf, "  %s=%s\n", k, v)
			}
		}
	}

	fmt.Fprintln(buf, "-------------------------------------------------------")
	fmt.Fprintln(buf)

	_, _ = buf.WriteTo(DebugOutputWriter)
}
