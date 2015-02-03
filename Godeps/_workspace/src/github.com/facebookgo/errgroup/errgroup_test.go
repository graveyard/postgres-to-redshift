package errgroup_test

import (
	"errors"
	"testing"

	"github.com/facebookgo/ensure"
	"github.com/facebookgo/errgroup"
)

func TestNada(t *testing.T) {
	t.Parallel()
	var g errgroup.Group
	ensure.Nil(t, g.Wait())
}

func TestOneError(t *testing.T) {
	t.Parallel()
	e := errors.New("")
	var g errgroup.Group
	g.Error(e)
	ensure.True(t, g.Wait() == e)
}

func TestTwoErrors(t *testing.T) {
	t.Parallel()
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	var g errgroup.Group
	g.Error(e1)
	g.Error(e2)
	ensure.DeepEqual(t, g.Wait().Error(), "multiple errors: e1 | e2")
}

func TestInvalidNilError(t *testing.T) {
	defer ensure.PanicDeepEqual(t, "error must not be nil")
	(&errgroup.Group{}).Error(nil)
}

func TestInvalidZeroLengthMultiError(t *testing.T) {
	defer ensure.PanicDeepEqual(t, "MultiError with no errors")
	(errgroup.MultiError{}).Error()
}

func TestInvalidOneLengthMultiError(t *testing.T) {
	defer ensure.PanicDeepEqual(t, "MultiError with only 1 error")
	(errgroup.MultiError{errors.New("")}).Error()
}

func TestAddDone(t *testing.T) {
	t.Parallel()
	var g errgroup.Group
	l := 10
	g.Add(l)
	for i := 0; i < l; i++ {
		go g.Done()
	}
	ensure.Nil(t, g.Wait())
}
