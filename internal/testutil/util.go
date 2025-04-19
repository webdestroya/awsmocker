package testutil

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// Deprecated: DONT USE THIS
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

func ReaderToString(rdr io.ReadCloser) string {
	defer rdr.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rdr)

	return buf.String()
}
