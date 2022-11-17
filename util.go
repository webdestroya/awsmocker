package awsmocker

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// var awsDomainRegexp = regexp.MustCompile(`(amazonaws\.com|\.aws)$`)

func encodeAsXml(obj any) string {
	out, err := xml.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(out)
}

func EncodeAsJson(obj any) string {
	out, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func inferContentType(value string) string {
	switch {
	case strings.HasPrefix(value, "<"):
		return ContentTypeXML
	case strings.HasPrefix(value, "{"):
		return ContentTypeJSON
	default:
		return ContentTypeText
	}
}

func isEof(r *bufio.Reader) bool {
	_, err := r.Peek(1)
	return errors.Is(err, io.EOF)
}

/*
// whether this is an AWS hostname that should be handled
func isAwsHostname(hostname string) bool {
	if strings.HasSuffix(hostname, "amazonaws.com") {
		return true
	}

	if strings.HasSuffix(hostname, "aws") {
		return true
	}

	if strings.HasSuffix(hostname, "amazonaws.com.cn") {
		return true
	}

	return false
}
*/
