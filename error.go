package awsmocker

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

// aws/protocol/restjson/decoder_util.go

type errorResponse struct {
	XMLName xml.Name `xml:"ErrorResponse" json:"-"`

	Type    string `xml:"Error>Type" json:"-"`
	Code    string `xml:"Error>Code" json:"code"`
	Message string `xml:"Error>Message" json:"message"`

	RequestId string `xml:"RequestId" json:"-"`

	statusCode int `xml:"-" json:"-"`
}

func (e *errorResponse) getResponse(rr *ReceivedRequest) *httpResponse {

	if e.statusCode == 0 {
		e.statusCode = http.StatusNotImplemented
	}

	switch rr.AssumedResponseType {
	case ContentTypeJSON:
		return &httpResponse{
			contentType: ContentTypeJSON,
			Body:        EncodeAsJson(e),
			StatusCode:  e.statusCode, // 501
		}
	case ContentTypeXML:
		return &httpResponse{
			contentType: ContentTypeXML,
			Body:        encodeAsXml(e),
			StatusCode:  e.statusCode, // 501
		}
	default:
		return &httpResponse{
			contentType: ContentTypeText,
			Body:        fmt.Sprintf("ERROR! %s: %s", e.Code, e.Message),
			StatusCode:  e.statusCode, // 501
		}
	}
}

func generateErrorStruct(code string, message string, args ...any) *errorResponse {
	return &errorResponse{
		Type:      "Sender",
		Code:      code,
		Message:   fmt.Sprintf(message, args...),
		RequestId: "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE",
	}
}
