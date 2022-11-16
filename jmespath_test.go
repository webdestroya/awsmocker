package awsmocker

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJmesPathMatching(t *testing.T) {
	var j = []byte(`{"foo": {"bar": {"baz": [0, 1, 2, 3, 4, 4.5, true, false, null, "hello"]}}}`)
	var d interface{}
	err := json.Unmarshal(j, &d)
	require.NoError(t, err)

	t.Run("test generic expressions", func(t *testing.T) {
		tables := []struct {
			expr string
			val  interface{}
		}{
			{"foo.bar.baz[99]", nil},
			{"foo.bar.baz[2]", int(2)},
			{"foo.bar.baz[2]", int8(2)},
			{"foo.bar.baz[2]", int16(2)},
			{"foo.bar.baz[2]", int32(2)},
			{"foo.bar.baz[2]", int64(2)},

			{"foo.bar.baz[2]", uint(2)},
			{"foo.bar.baz[2]", uint8(2)},
			{"foo.bar.baz[2]", uint16(2)},
			{"foo.bar.baz[2]", uint32(2)},
			{"foo.bar.baz[2]", uint64(2)},

			{"foo.bar.baz[2]", float32(2.0)},
			{"foo.bar.baz[2]", float64(2.0)},
			{"foo.bar.baz[2]", func(v any) bool { return reflect.DeepEqual(v, float64(2.0)) }},
		}

		for _, table := range tables {
			require.Truef(t, JMESMatch(d, table.expr, table.val), "expected true for %T", table.val)
		}
	})

	t.Run("bad expected values", func(t *testing.T) {
		tables := []struct {
			expr string
			val  interface{}
		}{
			{"some goofy expression", []int{2}},

			{"foo.bar.baz[2]", []int{2}},
			{"foo.bar.baz[2]", []int8{2}},
			{"foo.bar.baz[2]", []int16{2}},
			{"foo.bar.baz[2]", []int32{2}},
			{"foo.bar.baz[2]", []int64{2}},

			{"foo.bar.baz[2]", []uint{2}},
			{"foo.bar.baz[2]", []uint8{2}},
			{"foo.bar.baz[2]", []uint16{2}},
			{"foo.bar.baz[2]", []uint32{2}},
			{"foo.bar.baz[2]", []uint64{2}},

			{"foo.bar.baz[2]", []float32{2.0}},
			{"foo.bar.baz[2]", []float64{2.0}},
			{"foo.bar.baz[2]", func(v float64) bool { return reflect.DeepEqual(v, float64(2.0)) }},
		}

		for _, table := range tables {
			require.Panicsf(t, func() { JMESMatch(d, table.expr, table.val) }, "expected panic for %T", table.val)
		}
	})

}
