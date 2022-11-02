package awsmocker

import (
	"encoding/xml"
)

// aws/protocol/restjson/decoder_util.go

type errorResponse struct {
	XMLName xml.Name `xml:"ErrorResponse" json:"-"`

	Type    string `xml:"Error>Type" json:"__type,omitempty"`
	Code    string `xml:"Error>Code" json:"code"`
	Message string `xml:"Error>Message" json:"message"`

	RequestId string `xml:"RequestId" json:"-"`
}

func (e *errorResponse) getResponse(rr *receivedRequest) *httpResponse {
	switch rr.assumeResponseType {
	case ContentTypeJSON:
		return &httpResponse{
			contentType: ContentTypeJSON,
			Body:        encodeAsJson(e),
			StatusCode:  400,
		}
	case ContentTypeXML:
		return &httpResponse{
			contentType: ContentTypeXML,
			Body:        encodeAsXml(e),
			StatusCode:  400,
		}
	default:
		panic("Unknown Response Type???")
	}
}

func generateErrorStruct(code string, message string) *errorResponse {
	return &errorResponse{
		Type:      "Sender",
		Code:      code,
		Message:   message,
		RequestId: "7a62c49f-347e-4fc4-9331-6e8eEXAMPLE",
	}
}
