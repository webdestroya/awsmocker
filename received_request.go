package awsmocker

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var credExtractRegexp = regexp.MustCompile(`Credential=(\S+)\b`)
var jsonRegexp = regexp.MustCompile(`json`)

type receivedRequest struct {
	request *http.Request

	action  string
	service string
	region  string

	// rawBody string

	assumeResponseType string

	invalid bool
}

func newReceivedRequest(req *http.Request) *receivedRequest {
	recvreq := &receivedRequest{
		request:            req,
		assumeResponseType: ContentTypeXML,
	}

	// bodyReader, err := req.GetBody()
	// if err != nil {
	// 	panic(err)
	// }

	// buf := new(bytes.Buffer)
	// if _, err := buf.ReadFrom(bodyReader); err != nil {
	// 	panic(err)
	// }
	// recvreq.rawBody = buf.String()

	_ = req.ParseForm()

	reqContentType := req.Header.Get("content-type")
	if jsonRegexp.MatchString(reqContentType) {
		recvreq.assumeResponseType = ContentTypeJSON
	}

	authHeader := req.Header.Get("authorization")
	if authHeader != "" {
		matches := credExtractRegexp.FindStringSubmatch(authHeader)
		if len(matches) > 1 {
			// 0       1         2      3    4
			// fake/20221030/us-east-1/ecs/aws4_request
			parts := strings.Split(matches[1], "/")
			recvreq.region = parts[2]
			recvreq.service = parts[3]
		}
	}

	amzTarget := req.Header.Get("x-amz-target")
	if amzTarget != "" {
		// X-Amz-Target: AmazonEC2ContainerServiceV20141113.ListClusters
		_, opname, ok := strings.Cut(amzTarget, ".")
		if ok {
			recvreq.action = opname
		}
	}

	if recvreq.action == "" && req.PostForm.Has("Action") {
		recvreq.action = req.PostForm.Get("Action")
	}

	if recvreq.action == "" {
		log.Println("WARN: Received a request with no action????")
		recvreq.invalid = true
		return recvreq
	}

	return recvreq
}

func (r *receivedRequest) DebugDump() {
	fmt.Println("RECEIVED REQUEST: ----------------------------")

	fmt.Printf("Operation: %s (service=%s @ %s)\n", r.action, r.service, r.region)

	fmt.Printf("%s %s\n", r.request.Method, r.request.RequestURI)
	fmt.Printf("Host: %s\n", r.request.Host)
	for k, vlist := range r.request.Header {
		for _, v := range vlist {
			fmt.Printf("%s: %s\n", k, v)
		}
	}

	fmt.Println()
	fmt.Println("PARAMS:")
	for k, vlist := range r.request.Form {
		for _, v := range vlist {
			fmt.Printf("  %s=%s\n", k, v)
		}
	}

	fmt.Println("----------------------------------------------")
}
