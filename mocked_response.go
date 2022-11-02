package awsmocker

import (
	"encoding/xml"
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
	Body interface{}

	// Do not wrap the xml response in ACTIONResponse>ACTIONResult
	DoNotWrap bool

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

func (m *MockedResponse) getResponse(rr *receivedRequest) *httpResponse {

	if m.rawBody != "" && m.ContentType != "" {
		return &httpResponse{
			Body:        m.rawBody,
			StatusCode:  m.StatusCode,
			contentType: m.ContentType,
		}
	}

	rBody := reflect.Indirect(reflect.ValueOf(m.Body))
	bodyKind := rBody.Kind()

	switch bodyKind {
	case reflect.String:

		m.rawBody = rBody.String()
		m.ContentType = ContentTypeText
		if m.ContentType == "" && len(m.rawBody) > 1 {
			switch {
			case m.rawBody[0:1] == "<":
				m.ContentType = ContentTypeXML
			case m.rawBody[0:1] == "{":
				m.ContentType = ContentTypeJSON
			}
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
		case m.Encoding == ResponseEncodingDefault && rr.assumeResponseType == ContentTypeJSON:
			return &httpResponse{
				Body:        encodeAsJson(m.Body),
				StatusCode:  m.StatusCode,
				contentType: ContentTypeJSON,
			}

		case m.Encoding == ResponseEncodingXML:
			fallthrough
		case m.Encoding == ResponseEncodingDefault && rr.assumeResponseType == ContentTypeXML:

			if m.DoNotWrap {

				xmlout, err := mxj.AnyXmlIndent(m.Body, "", "  ")
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

				xmlout, err := mxj.AnyXmlIndent(wrappedObj, "", "  ", ""+m.action+"Response")
				if err != nil {
					panic(err)
				}

				return &httpResponse{
					Body:        strings.ReplaceAll(string(xmlout), "_ACTION_NAME_HERE_", m.action),
					StatusCode:  m.StatusCode,
					contentType: ContentTypeXML,
				}
			} else {
				resultName := "" + m.action + "Result"
				wrappedMap := map[string]interface{}{
					resultName: m.Body,
					"ResponseMetadata": map[string]string{
						"RequestId": "01234567-89ab-cdef-0123-456789abcdef",
					},
				}

				xmlout, err := mxj.AnyXmlIndent(wrappedMap, "", "  ", ""+m.action+"Response")
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
	panic("Unknown type provided for response Body. Make a string/struct/map/slice")
}
