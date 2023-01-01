package awsmocker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInferContentType(t *testing.T) {
	tables := []struct {
		exp string
		str string
	}{
		{ContentTypeXML, "<Test>Thing</Test>"},
		{ContentTypeXML, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\" ?>\n<Fake>Thing</Fake>"},
		{ContentTypeJSON, `{"test":"thing"}`},
		{ContentTypeJSON, `{}`},
		{ContentTypeText, `raw=thing`},
	}

	for _, table := range tables {
		require.Equal(t, table.exp, inferContentType(table.str))
	}
}
