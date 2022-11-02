package awsmocker

import (
	"encoding/json"
	"encoding/xml"
)

func encodeAsXml(obj interface{}) string {
	out, err := xml.MarshalIndent(obj, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(out)
}

func encodeAsJson(obj interface{}) string {
	out, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(out)
}
