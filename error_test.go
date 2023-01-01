package awsmocker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorResponse_getResponse(t *testing.T) {
	t.Run("XML", func(t *testing.T) {
		er := generateErrorStruct(418, "GenericErrorCode", "Some %s message %d", "THING", 123)

		hr := er.getResponse(&ReceivedRequest{AssumedResponseType: ContentTypeXML})
		require.Equal(t, 418, hr.StatusCode)
		require.Equal(t, ContentTypeXML, hr.contentType)
		require.Contains(t, hr.Body, "<Type>Sender</Type>")
	})
}
