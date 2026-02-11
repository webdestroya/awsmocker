package awsmocker_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/webdestroya/awsmocker"
)

func TestNewSimpleMockedEndpoint(t *testing.T) {
	me := awsmocker.NewSimpleMockedEndpoint(
		"sts",
		"GetCallerIdentity",
		"SimpleBody",
	)
	require.Equal(t, "sts", me.Request.Service)
	require.Equal(t, "GetCallerIdentity", me.Request.Action)
	require.Equal(t, "SimpleBody", me.Response.Body)
}
