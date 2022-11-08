package awsmocker_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
	"github.com/webdestroya/awsmocker/internal/testutil"
)

func TestEc2IMDS(t *testing.T) {
	awsmocker.Start(t, &awsmocker.MockerOptions{
		MockEc2Metadata: true,
	})

	client := imds.NewFromConfig(testutil.GetAwsConfig())

	ctx := context.TODO()

	t.Run("check region", func(t *testing.T) {
		region, err := client.GetRegion(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, awsmocker.DefaultRegion, region.Region)
	})

	t.Run("user data", func(t *testing.T) {
		resp, err := client.GetUserData(ctx, nil)
		require.NoError(t, err)

		data := testutil.ReaderToString(resp.Content)
		require.Equal(t, "# awsmocker", data)
	})

	t.Run("iam info", func(t *testing.T) {
		resp, err := client.GetIAMInfo(ctx, nil)
		require.NoError(t, err)

		require.Contains(t, resp.InstanceProfileArn, "awsmocker-instance-profile")
	})

	t.Run("iam creds", func(t *testing.T) {
		provider := ec2rolecreds.New()
		creds, err := provider.Retrieve(ctx)
		require.NoError(t, err)

		require.Equal(t, "FAKEKEY", creds.AccessKeyID)
		require.Equal(t, "fakeSecretKEY", creds.SecretAccessKey)
		require.Equal(t, "FAKETOKEN", creds.SessionToken)
	})

}
