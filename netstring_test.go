package netstring_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/brianm/netstring"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestMarshal_HappyPath(t *testing.T) {
	s := "hello world!"
	out, err := netstring.Marshal(s)
	require.NoError(t, err)

	assert.Equal(t, "12:hello world!,", string(out))
}

func TestMarshal_HappyPathUTF8(t *testing.T) {
	s := "hello 🚀!"
	out, err := netstring.Marshal(s)
	require.NoError(t, err)

	assert.Equal(t, "11:hello 🚀!,", string(out))
}

func TestUnmarshal_TwoHappy(t *testing.T) {
	in := "2:hi,4:woof,"
	var ary []string
	err := netstring.Unmarshal([]byte(in), &ary)
	require.NoError(t, err)

	assert.Equal(t, 2, len(ary))
	assert.Equal(t, "hi", ary[0])
	assert.Equal(t, "woof", ary[1])
}

func TestUnmarshal_OneHappyUTF8(t *testing.T) {
	in := "11:hello 🚀!,"
	var ary []string
	err := netstring.Unmarshal([]byte(in), &ary)
	require.NoError(t, err)

	assert.Equal(t, 1, len(ary))
	assert.Equal(t, "hello 🚀!", ary[0])
}

func TestUnmarshal_BadNumber(t *testing.T) {
	in := "1j:hello 🚀!,"
	var ary []string
	err := netstring.Unmarshal([]byte(in), &ary)
	require.Error(t, err)
}

func TestUnmarshal_BadString(t *testing.T) {
	in := "11:hello 🚀!!"
	var ary []string
	err := netstring.Unmarshal([]byte(in), &ary)
	require.Error(t, err)
}

func TestUnmarshal_TooLong(t *testing.T) {
	in := "11:hello 🚀!,8"
	var ary []string
	err := netstring.Unmarshal([]byte(in), &ary)
	require.Error(t, err)
}

func TestUnmarshal_TooShort(t *testing.T) {
	in := "11:hello"
	var ary []string
	err := netstring.Unmarshal([]byte(in), &ary)
	require.Error(t, err)
}

func TestUnmarshal_AskInvalid(t *testing.T) {
	in := "1:a,"
	a := struct{}{}
	err := netstring.Unmarshal([]byte(in), &a)
	require.Error(t, err)
}

func TestUnmarshal_Fuzz(t *testing.T) {
	props := gopter.NewProperties(nil)
	props.Property(
		"unmarshalls things of arbitrary length",
		prop.ForAll(func(v string) bool {
			out, err := netstring.Marshal(v)
			if err != nil {
				return false
			}
			assert.Equal(t, fmt.Sprintf("%d:%s,", len([]byte(v)), v), string(out))

			var in []string
			err = netstring.Unmarshal(out, &in)
			if err != nil {
				return false
			}
			assert.Equal(t, v, in[0])
			return reflect.DeepEqual(v, in[0])
		}, gen.AnyString()))
	props.TestingRun(t)
}
