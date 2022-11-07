package awsmocker

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/clbanning/mxj"
)

const (
	ContentTypeXML  = "text/xml"
	ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain"
)

type MockedResponse struct {
	// modify the status code. default is 200
	StatusCode int

	// force the content type. default will be determined by request content type
	ContentType string

	Encoding ResponseEncoding

	// a string, struct or map that will be encoded as the response
	//
	// Also accepts a function that is of the following signatures:
	// func(ReceivedRequest) (string) = string payload (with 200 OK, inferred content type)
	// func(ReceivedRequest) (string, int) = string payload, <int> status code (with inferred content type)
	// func(ReceivedRequest) (string, int, string) = string payload, <int> status code, content type
	Body interface{}

	// Do not wrap the xml response in ACTIONResponse>ACTIONResult
	DoNotWrap bool
	RootTag   string

	rawBody string

	action string
}

type wrapperStruct struct {
	XMLName   xml.Name    `xml:"_ACTION_NAME_HERE_Response"`
	Result    interface{} `xml:"_ACTION_NAME_HERE_Result"`
	RequestId string      `xml:"ResponseMetadata>RequestId"`
}

func (m *MockedResponse) prep() {
	if m.StatusCode == 0 {
		m.StatusCode = http.StatusOK
	}
}

func (m *MockedResponse) getResponse(rr *ReceivedRequest) *httpResponse {

	if m.rawBody != "" && m.ContentType != "" {
		return &httpResponse{
			Body:        m.rawBody,
			StatusCode:  m.StatusCode,
			contentType: m.ContentType,
		}
	}

	actionName := m.action
	if actionName == "" {
		actionName = rr.Action
	}

	rBody := reflect.Indirect(reflect.ValueOf(m.Body))
	bodyKind := rBody.Kind()

	switch bodyKind {
	case reflect.Func:
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
		m.ContentType = ContentTypeText
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
					panic(err)
				}

				return &httpResponse{
					Body:        string(xmlout),
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
					panic(err)
				}

				return &httpResponse{
					Body:        strings.ReplaceAll(string(xmlout), "_ACTION_NAME_HERE_", actionName),
					StatusCode:  m.StatusCode,
					contentType: ContentTypeXML,
				}
			} else {
				resultName := "" + actionName + "Result"
				wrappedMap := map[string]interface{}{
					resultName: m.Body,
					"ResponseMetadata": map[string]string{
						"RequestId": "01234567-89ab-cdef-0123-456789abcdef",
					},
				}

				xmlout, err := mxj.AnyXmlIndent(wrappedMap, "", "  ", ""+actionName+"Response")
				if err != nil {
					panic(err)
				}

				return &httpResponse{
					Body:        string(xmlout),
					StatusCode:  m.StatusCode,
					contentType: ContentTypeXML,
				}
			}
		}
	}
	panic(fmt.Errorf("Unknown type provided for response Body. Make a string/struct/map/slice was(%t) || %v", m.Body, EncodeAsJson(m.Body)))
}
