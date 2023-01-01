package awsmocker_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestReceivedRequest_DebugDump(t *testing.T) {
	origWriter := awsmocker.DebugOutputWriter
	origDebugMode := awsmocker.GlobalDebugMode

	t.Cleanup(func() {
		awsmocker.DebugOutputWriter = origWriter
		awsmocker.GlobalDebugMode = origDebugMode
	})

	buf := new(bytes.Buffer)
	awsmocker.GlobalDebugMode = true
	awsmocker.DebugOutputWriter = buf

	// tables := []struct{}{}

	// for _, table := range tables {

	// }

	info := awsmocker.Start(t, &awsmocker.MockerOptions{
		SkipDefaultMocks: true,
		ReturnAwsConfig:  true,
		// V
		Mocks: []*awsmocker.MockedEndpoint{
			awsmocker.Mock_Failure("ecs", "ListClusters"),
			awsmocker.Mock_Failure_WithCode(403, "ecs", "ListServices", "SomeCode", "SomeMessage"),
		},
	})

	ecsClient := ecs.NewFromConfig(*info.AwsConfig)

	_, err := ecsClient.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	require.Error(t, err)
	require.ErrorContains(t, err, "AccessDenied")

	debugStr := buf.String()
	buf.Reset()
	require.Contains(t, debugStr, "AWSMOCKER RECEIVED REQUEST:")
	require.Contains(t, debugStr, "AWSMOCKER RESPONSE:")
	require.Contains(t, debugStr, "POST")
	require.Contains(t, debugStr, "ecs.us-east-1.amazonaws.com")
	require.Contains(t, debugStr, "ListClusters")

}
