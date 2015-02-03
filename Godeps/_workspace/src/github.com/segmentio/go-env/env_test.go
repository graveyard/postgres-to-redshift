package env

import "github.com/bmizerany/assert"
import "testing"
import "os"

func TestGet(t *testing.T) {
	{
		s, err := Get("FOO")
		assert.Equal(t, "", s)
		assert.Equal(t, "environment variable FOO missing", err.Error())
	}

	{
		os.Setenv("FOO", "bar")
		s, err := Get("FOO")
		assert.Equal(t, "bar", s)
		assert.Equal(t, nil, err)
	}
}

func TestMustGet(t *testing.T) {
	defer func() {
		v := recover().(error)
		assert.Equal(t, v.Error(), "environment variable BAZ missing")
	}()

	{
		os.Setenv("BAR", "baz")
		s := MustGet("BAR")
		assert.Equal(t, "baz", s)
	}

	MustGet("BAZ")
}
