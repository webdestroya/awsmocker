package awsmocker_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestAwsConfigBuilder(t *testing.T) {
	info := awsmocker.Start(t)
	stsClient := sts.NewFromConfig(info.Config())

	resp, err := stsClient.GetCallerIdentity(context.TODO(), nil)
	require.NoError(t, err)
	require.EqualValuesf(t, awsmocker.DefaultAccountId, *resp.Account, "account id mismatch")
}
