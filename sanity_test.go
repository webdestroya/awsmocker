package awsmocker_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go/document"
	"github.com/stretchr/testify/require"
)

func TestSanity(t *testing.T) {
	t.Run("NoSerde", func(t *testing.T) {
		require.False(t, document.IsNoSerde(uint(123)))
		require.False(t, document.IsNoSerde(uint64(123)))
		require.False(t, document.IsNoSerde(nil))

		require.True(t, document.IsNoSerde(ec2.DescribeSubnetsOutput{}))
		require.True(t, document.IsNoSerde(&ec2.DescribeSubnetsOutput{}))
	})

}

// func TestReflection(t *testing.T) {
// 	thing := &ec2.DescribeAccountAttributesOutput{}
// 	thing2 := ec2.DescribeAccountAttributesOutput{}
// 	var thing3 any = any(ec2.DescribeAccountAttributesOutput{})

// 	f1 := func(_ *awsmocker.ReceivedRequest) (*ec2.DescribeSubnetsOutput, error) {
// 		return nil, nil
// 	}

// 	f2 := func(_ *awsmocker.ReceivedRequest, _ *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
// 		return nil, nil
// 	}

// 	f3 := func(_ *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
// 		return nil, nil
// 	}
// 	_ = f2
// 	_ = f3
// 	_ = thing2

// 	t.Logf("Thing1:Kind: %s", reflect.TypeOf(thing).Kind().String())
// 	t.Logf("Thing2:Kind: %s", reflect.TypeOf(thing2).Kind().String())
// 	t.Logf("Thing3:Kind: %s", reflect.TypeOf(thing3).Kind().String())
// 	t.Logf("Thing3:Kind: %s", reflect.TypeOf(&thing3).Kind().String())
// 	t.Logf("F3:Kind: %s", reflect.TypeOf(f3).Kind().String())
// 	t.Logf("New: %s", reflect.TypeFor[*awsmocker.ReceivedRequest]().String())

// 	typ := reflect.TypeOf(thing)

// 	require.True(t, document.IsNoSerde(thing))

// 	t.Logf("Kind=%s", reflect.Indirect(reflect.ValueOf(thing)).Kind().String())
// 	t.Logf("TYPE=%s", typ.String())
// 	t.Logf("TYPE=%s", typ.Elem().String())
// 	t.Logf("PKG=%s", typ.Elem().PkgPath())

// 	f1type := reflect.TypeOf(f1)
// 	t.Logf("F1 TYPE=%s", f1type.String())
// 	t.Logf("F1 NumIn=%d", f1type.NumIn())
// 	t.Logf("F1 NumOut=%d", f1type.NumOut())
// 	t.Logf("F1 In.0=%s", f1type.In(0).String())
// 	t.Logf("F1 Out.0=%s [%s]", f1type.Out(0).String(), f1type.Out(0).Elem().PkgPath())

// }
