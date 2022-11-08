package awsmocker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCACertPEM(t *testing.T) {
	require.ElementsMatch(t, caCert, CACertPEM())
}
