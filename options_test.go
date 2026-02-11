package awsmocker

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/require"
)

func TestMockerOptions(t *testing.T) {
	t.Run("WithVerbosity", func(t *testing.T) {
		mo := newOptions()
		WithVerbosity(true)(mo)
		require.True(t, mo.Verbose)

		WithVerbosity(false)(mo)
		require.False(t, mo.Verbose)
	})

	t.Run("WithTimeout", func(t *testing.T) {
		mo := newOptions()
		WithTimeout(10 * time.Minute)(mo)
		require.Equal(t, 10*time.Minute, mo.Timeout)
	})

	t.Run("WithoutDefaultMocks", func(t *testing.T) {
		mo := newOptions()
		WithEC2Metadata()(mo)
		require.Contains(t, mo.Mocks, MockStsGetCallerIdentityValid)
		WithoutDefaultMocks()(mo)
		require.NotContains(t, mo.Mocks, MockStsGetCallerIdentityValid)
	})

	t.Run("WithAWSConfigOptions", func(t *testing.T) {
		m := Start(t,
			WithAWSConfigOptions(config.WithRegion("blah-yar")),
		)

		require.Equal(t, "blah-yar", m.Config().Region)

	})
}
