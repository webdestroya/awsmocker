package awsmocker

import (
	"fmt"
	"reflect"

	"github.com/jmespath/go-jmespath"
)

// this will normalize a value to allow it to be compared more easily against another one
// used for JMES path, because json numbers are float64 so you cant compare to other value types
func jmesValueNormalize(value any) any {
	if value == nil {
		return nil
	}

	// switch v := value.(type) {
	// case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	// 	return fmt.Sprintf("%d", v)
	// case float32:
	// 	return strconv.FormatFloat(float64(v), 'f', -1, 64)
	// case float64:
	// 	return strconv.FormatFloat(float64(v), 'f', -1, 64)
	// case string, bool:
	// 	return v
	// default:
	// 	return v
	// }
	switch v := value.(type) {
	case string, bool, float64:
		return v
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case func(any) bool:
		return v
	default:
		panic("jmes expressions should evaluate to a string/bool/number/nil. Advanced checks should use a function")
	}

}

// Performs a JMES Expression match on the object.
// The expected value should be a scalar or a function that takes a single any
// and returns a boolean
// the value returned from the jmes expression should equal the expected value
// all numerical values will be casted to a float64 (as that is what json numbers are treated as)
func JMESMatch(obj any, expression string, expected any) bool {

	resp, err := jmespath.Search(expression, obj)
	if err != nil {
		panic(fmt.Errorf("Failed to parse expression: '%s': %w", expression, err))
	}

	exp := jmesValueNormalize(expected)

	funcCheck, ok := exp.(func(any) bool)
	if ok {
		return funcCheck(resp)
	}

	return reflect.DeepEqual(resp, exp)
}
