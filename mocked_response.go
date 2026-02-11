package awsmocker

import (
	"encoding/xml"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/document"
	"github.com/clbanning/mxj"
)

const (
	ContentTypeXML  = "text/xml"
	ContentTypeJSON = "application/x-amz-json-1.1"
	ContentTypeText = "text/plain"
)

var (
	byteArrayType = reflect.SliceOf(reflect.TypeOf((*byte)(nil)).Elem())
	rrType        = reflect.TypeFor[*ReceivedRequest]()
	errType       = reflect.TypeFor[error]()
)

// type used to generate
type directTypeFunc = func(*ReceivedRequest) (any, error)

type MockedResponse struct {
	// modify the status code. default is 200
	StatusCode int

	// force the content type. default will be determined by request content type
	ContentType string

	Encoding ResponseEncoding

	// a string, struct or map that will be encoded as the response
	//
	// Also accepts a function that is of the following signatures:
	// func(*ReceivedRequest) (string) = string payload (with 200 OK, inferred content type)
	// func(*ReceivedRequest) (string, int) = string payload, <int> status code (with inferred content type)
	// func(*ReceivedRequest) (string, int, string) = string payload, <int> status code, content type
	// func(*ReceivedRequest) (*service.ACTIONOutput, error) = return the result type directly, or error
	// func(*ReceivedRequest, *service.ACTIONInput) (*service.ACTIONOutput, error) = return the result type directly, or error
	// func(*service.ACTIONInput) (*service.ACTIONOutput, error) = return the result type directly, or error
	Body any

	// Do not wrap the xml response in ACTIONResponse>ACTIONResult
	DoNotWrap bool
	RootTag   string

	// If provided, then all other fields are ignored, and the user
	// is responsible for building an HTTP response themselves
	Handler MockedRequestHandler

	rawBody string

	action string
}

type wrapperStruct struct {
	XMLName   xml.Name `xml:"_ACTION_NAME_HERE_Response"`
	Result    any      `xml:"_ACTION_NAME_HERE_Result"`
	RequestId string   `xml:"ResponseMetadata>RequestId"`
}

func (m *MockedResponse) prep() {
	if m.StatusCode == 0 {
		m.StatusCode = http.StatusOK
	}
}

func (m *MockedResponse) getResponse(rr *ReceivedRequest) *httpResponse {

	if m.Handler != nil {
		// user wants to do it all themselves
		return &httpResponse{
			forcedHttpResponse: m.Handler(rr),
		}
	}

	if m.rawBody != "" && m.ContentType != "" {
		return &httpResponse{
			Body:        m.rawBody,
			StatusCode:  m.StatusCode,
			contentType: m.ContentType,
		}
	}

	if dir := m.processDirectRequest(rr); dir != nil {
		return dir
	}

	actionName := m.action
	if actionName == "" {
		actionName = rr.Action
	}

	rBody := reflect.Indirect(reflect.ValueOf(m.Body))
	bodyKind := rBody.Kind()

	switch bodyKind {
	case reflect.Func:

		switch rBody.Interface().(type) {
		case func(*ReceivedRequest) string:
		case func(*ReceivedRequest) (string, int):
		case func(*ReceivedRequest) (string, int, string):
			// valid function
		default:
			return generateErrorStruct(0, "InvalidBodyFunc", "the function you gave for the body has the wrong signature").getResponse(rr)
		}

		respRet := rBody.Call([]reflect.Value{reflect.ValueOf(rr)})
		respBody := respRet[0].String()
		respStatus := http.StatusOK
		var respContentType string

		if len(respRet) > 1 {
			respStatus = int(respRet[1].Int())
		}

		if len(respRet) > 2 {
			respContentType = respRet[2].String()
		} else {
			respContentType = inferContentType(respBody)
		}

		return &httpResponse{
			Body:        respBody,
			contentType: respContentType,
			StatusCode:  respStatus,
		}
	case reflect.String:

		m.rawBody = rBody.String()
		// m.ContentType = ContentTypeText
		if m.ContentType == "" && len(m.rawBody) > 1 {
			m.ContentType = inferContentType(m.rawBody)
		}
		return &httpResponse{
			Body:        m.rawBody,
			StatusCode:  m.StatusCode,
			contentType: m.ContentType,
		}

	case reflect.Map, reflect.Array, reflect.Slice, reflect.Struct:

		switch {
		case m.Encoding == ResponseEncodingJSON:
			fallthrough
		case m.Encoding == ResponseEncodingDefault && rr.AssumedResponseType == ContentTypeJSON:
			return &httpResponse{
				Body:        EncodeAsJson(m.Body),
				StatusCode:  m.StatusCode,
				contentType: ContentTypeJSON,
			}

		case m.Encoding == ResponseEncodingXML:
			fallthrough
		case m.Encoding == ResponseEncodingDefault && rr.AssumedResponseType == ContentTypeXML:

			if m.DoNotWrap {

				if m.RootTag == "" {
					m.RootTag = "" + actionName + "Response"
				}

				xmlout, err := mxj.AnyXmlIndent(m.Body, "", "  ", m.RootTag, "")
				if err != nil {
					return generateErrorStruct(0, "BadMockBody", "Could not serialize body to XML: %s", err).getResponse(rr)
				}

				return &httpResponse{
					bodyRaw:     xmlout,
					StatusCode:  m.StatusCode,
					contentType: ContentTypeXML,
				}
			} else if bodyKind == reflect.Struct {
				wrappedObj := wrapperStruct{
					Result:    m.Body,
					RequestId: "01234567-89ab-cdef-0123-456789abcdef",
				}

				xmlout, err := mxj.AnyXmlIndent(wrappedObj, "", "  ", ""+actionName+"Response")
				if err != nil {
					return generateErrorStruct(0, "BadMockBody", "Could not serialize body to XML: %s", err).getResponse(rr)
				}

				return &httpResponse{
					Body:        strings.ReplaceAll(string(xmlout), "_ACTION_NAME_HERE_", actionName),
					StatusCode:  m.StatusCode,
					contentType: ContentTypeXML,
				}
			}

			resultName := "" + actionName + "Result"
			wrappedMap := map[string]any{
				resultName: m.Body,
				"ResponseMetadata": map[string]string{
					"RequestId": "01234567-89ab-cdef-0123-456789abcdef",
				},
			}

			xmlout, err := mxj.AnyXmlIndent(wrappedMap, "", "  ", ""+actionName+"Response")
			if err != nil {
				return generateErrorStruct(0, "BadMockBody", "Could not serialize body to XML: %s", err).getResponse(rr)
			}

			return &httpResponse{
				bodyRaw:     xmlout,
				StatusCode:  m.StatusCode,
				contentType: ContentTypeXML,
			}

		case bodyKind == reflect.Slice && rBody.Type() == byteArrayType:

			cType := m.ContentType
			if cType == "" {
				cType = http.DetectContentType(m.Body.([]byte))
			}

			return &httpResponse{
				bodyRaw:     m.Body.([]byte),
				StatusCode:  m.StatusCode,
				contentType: cType,
			}
		}
	}
	return generateErrorStruct(0, "BadMockResponse", "Don't know how to encode a kind=%v using content type=%s", bodyKind, m.ContentType).getResponse(rr)
}

func (m *MockedResponse) processDirectRequest(rr *ReceivedRequest) *httpResponse {

	if m.Body == nil {
		return nil
	}

	body := m.Body
	var err error

	if rr.HttpRequest == nil || len(rr.HttpRequest.Header) == 0 {
		return nil
	}
	reqId, perr := strconv.ParseUint(rr.HttpRequest.Header.Get(mwHeaderRequestId), 10, 64)
	if perr != nil {
		return generateErrorStruct(0, "BadMockBody", "Failed to get direct mocker: %s", perr.Error()).getResponse(rr)

	}

	mkr := rr.mocker
	if mkr == nil {
		return generateErrorStruct(0, "BadMockBody", "Failed to get direct mocker").getResponse(rr)
	}

	entry, ok := mkr.requestLog.Load(reqId)
	if !ok {
		return generateErrorStruct(0, "BadMockBody", "Failed to find mock in DB??").getResponse(rr)
	}

	rec := entry.(mwDBEntry)

	if !document.IsNoSerde(body) {

		// check if fancy func
		if fn, ok := body.(directTypeFunc); ok {
			body, err = fn(rr)
		} else {

			body, err = processDirectRequestFunc(rec, rr, reflect.Indirect(reflect.ValueOf(body)))

		}

		if body != nil && !document.IsNoSerde(body) {
			return nil
		}
	}

	if body == nil && err == nil {
		return nil
	}

	if reflect.TypeOf(body).Kind() == reflect.Struct {
		val := reflect.ValueOf(body)
		vp := reflect.New(val.Type())
		vp.Elem().Set(val)
		body = vp.Interface()
	}

	rec.Error = err
	rec.Response = body

	mkr.requestLog.Store(reqId, rec)

	return &httpResponse{
		StatusCode:  http.StatusOK,
		Body:        "",
		contentType: ContentTypeJSON,
		extraHeaders: map[string]string{
			mwHeaderUseDB: "true",
		},
	}
}

func processDirectRequestFunc(entry mwDBEntry, rr *ReceivedRequest, fnv reflect.Value) (any, error) {
	typ := fnv.Type()

	if typ.Kind() != reflect.Func {
		return nil, nil
	}

	params := entry.Parameters
	paramT := reflect.TypeOf(params)

	inputs := make([]reflect.Value, 0, 2)

	if typ.NumIn() == 1 {

		in1 := typ.In(0)
		if in1 == rrType {
			inputs = append(inputs, reflect.ValueOf(rr))
		} else if in1 == paramT {
			inputs = append(inputs, reflect.ValueOf(params))
		} else {
			return nil, nil
		}

	} else if typ.NumIn() == 2 {
		if in1 := typ.In(0); in1 == rrType {
			inputs = append(inputs, reflect.ValueOf(rr))
		} else {
			return nil, nil
		}

		if in2 := typ.In(1); in2 == paramT {
			inputs = append(inputs, reflect.ValueOf(params))
		} else {
			return nil, nil
		}
	} else {
		// invalid signature
		return nil, nil
	}

	if typ.NumOut() != 2 {
		return nil, nil
	}

	if out2 := typ.Out(1); out2 != errType {
		// 2nd return must be error
		return nil, nil
	}

	outputs := fnv.Call(inputs)

	ret := outputs[0].Interface()

	if typ.NumOut() == 2 {
		if err := outputs[1].Interface(); err != nil {
			return ret, err.(error)
		}
	}

	return ret, nil
}
