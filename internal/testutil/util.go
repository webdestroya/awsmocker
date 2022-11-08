package testutil

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func GetAwsConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		// add creds just in case something happens
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("XXfakekey", "XXfakesecret", "xxtoken")),
		config.WithDefaultRegion("us-east-1"),
		config.WithRetryer(func() aws.Retryer {
			return aws.NopRetryer{}
		}),
	)
	if err != nil {
		panic(err)
	}
	return cfg
}

func PrintHttpResponse(t *testing.T, resp *http.Response) {
	if !testing.Verbose() {
		return
	}

	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)
	err := resp.Write(writer)
	if err != nil {
		t.Errorf("error outputting http response: %s", err)
	}
	writer.Flush()
	t.Log("RESPONSE:\n>>>>>>>>>>>>>>>>>>>>>>>>\n" + buff.String() + "\n<<<<<<<<<<<<<<<<<<<<<<<<")
}

func ReaderToString(rdr io.ReadCloser) string {
	defer rdr.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rdr)

	return buf.String()
}
