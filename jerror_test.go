package jerror

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	require := require.New(t)

	err := New("simple string")
	require.EqualError(err, "simple string")
}

func TestArgs(t *testing.T) {
	require := require.New(t)

	jerr := New("args: %s, %v")
	require.EqualError(jerr, "args: %s, %v")

	err := jerr.New().Args("one", 2)
	require.EqualError(err, "args: one, 2")
}

func TestWrap(t *testing.T) {
	require := require.New(t)

	err := fmt.Errorf("standard error")
	jerr := New("jerror error")
	jerr2 := New("another error")

	require.True(errors.Is(jerr, jerr))
	require.False(errors.Is(jerr, err))

	jerrw := jerr.New().Wrap(err)
	require.True(errors.Is(jerrw, err))
	require.EqualError(jerrw, "jerror error: standard error")

	errw := fmt.Errorf("error: %w", jerrw)
	require.True(errors.Is(errw, jerr))
	require.True(errors.Is(errw, err))

	jerrw2 := jerr2.New().Wrap(errw)
	require.True(errors.Is(jerrw2, jerr))
	require.True(errors.Is(jerrw2, err))
	require.True(errors.Is(jerrw2, jerr2))
}

var (
	ErrTest   = New("test error")
	ErrNormal = fmt.Errorf("normal error")
)

type ErrEmbed struct {
	*JError
	Code int
}

func ErrEmbedWrap(code int, err error) *ErrEmbed {
	return &ErrEmbed{
		JError: ErrTest.New().Wrap(err),
		Code:   code,
	}
}

func ErrEmbedNew(code int) *ErrEmbed {
	return &ErrEmbed{
		JError: ErrTest.New(),
		Code:   code,
	}
}

func TestEmbed(t *testing.T) {
	require.True(t, errors.Is(ErrEmbedNew(42), ErrTest))

	jerr := ErrEmbedWrap(42, ErrNormal)
	oerr := fmt.Errorf("outer error: %w", jerr)

	require.True(t, errors.Is(oerr, ErrNormal))
	require.True(t, errors.Is(oerr, ErrTest))

	var perr *ErrEmbed
	ok := errors.As(oerr, &perr)
	require.True(t, ok)
	require.Error(t, perr)
	require.Equal(t, 42, perr.Code)
}

func TestStack(t *testing.T) {
	// created with New
	err := New("stack error").New()
	testStack(t, err)

	// created with Set
	err = New("stack error").Set("key", "value")
	testStack(t, err)

	// created with Wrap
	err = New("stack error").Wrap(os.ErrClosed)
	testStack(t, err)

	// created with Args
	err = New("stack error %d").Args(1)
	testStack(t, err)
}

func testStack(t *testing.T, err *JError) {
	t.Helper()
	require := require.New(t)

	require.Len(err.frames, 3)

	frames := err.Frames()
	last := frames[0]
	parts := strings.Split(last.Function, "/")
	require.Equal("jerror.TestStack", parts[len(parts)-1])

	parts = strings.Split(last.File, "/")
	require.Equal("jerror_test.go", parts[len(parts)-1])

	require.True(last.Line > 0)
}

func TestValues(t *testing.T) {
	require := require.New(t)

	orig := New("values error")
	err := orig.New()
	_ = err.Set("key", "value")
	_ = err.Set("key2", 42)
	require.EqualError(err, "values error")

	// make sure the original error is not modified
	_, ok := orig.Get("key")
	require.False(ok)

	val, ok := err.Get("key")
	require.True(ok)
	require.Equal("value", val)

	val, ok = err.Get("key2")
	require.True(ok)
	require.Equal(42, val)

	val, ok = err.GetInt("key2")
	require.True(ok)
	require.Equal(42, val)

	// return not ok if the type is not correct
	val, ok = err.GetInt("key")
	require.False(ok)
	require.Equal(0, val)

	logAttrs := err.SlogAttributes("test", true)
	require.Equal("test", logAttrs.Key)

	group := logAttrs.Value.Group()

	require.Equal("error", group[0].Key)
	require.Equal("values error", group[0].Value.String())

	require.Equal("stack", group[1].Key)
	require.Len(group[1].Value.Group(), 3)

	values := group[2].Value.Group()
	require.Equal("values", group[2].Key)
	require.Len(values, 2)

	require.Equal("key", values[0].Key)
	require.Equal("value", values[0].Value.String())
	require.Equal("key2", values[1].Key)
	require.Equal("42", values[1].Value.String())
}

func TestSlogOldestError(t *testing.T) {
	jerr2 := New("jerr2").Set("jerr", 2)
	jerr1 := New("jerr1").Set("jerr", 1).Wrap(jerr2)

	attrs := jerr1.SlogAttributes("test", true)
	grp := attrs.Value.Group()

	var found_first, found_last bool
	for _, a := range grp {
		switch a.Key {
		case "values":
			values := a.Value.Group()
			require.Len(t, values, 1)
			require.Equal(t, "1", values[0].Value.String())
			found_first = true

		case "last_jerror":
			ngrp := a.Value.Group()
			for _, a := range ngrp {
				if a.Key == "values" {
					values := a.Value.Group()
					require.Len(t, values, 1)
					require.Equal(t, "2", values[0].Value.String())
					found_last = true
				}
			}
		}
	}

	require.True(t, found_first)
	require.True(t, found_last)
}

func TestNewWithValues(t *testing.T) {
	base := NewWithValues("test", Values{
		"foo": "bar",
		"baz": "qux",
	})

	child := base.New().Set("foo", "baz")

	v, ok := base.GetString("foo")
	require.True(t, ok)
	require.Equal(t, "bar", v)

	v, ok = child.GetString("foo")
	require.True(t, ok)
	require.Equal(t, "baz", v)
}

func TestThrottle(t *testing.T) {
	t.Cleanup(Unthrottle)

	e := New("msg")

	Throttle(time.Millisecond, 5)

	start := time.Now()
	var frames int
	var noFrames int
	for time.Since(start) < 4*time.Millisecond {
		err := e.New()
		if err.Frames() != nil {
			frames++
		} else {
			noFrames++
		}
	}

	require.Equal(t, 5*4, frames)
	require.NotZero(t, noFrames)

	Unthrottle()

	frames = 0
	noFrames = 0
	start = time.Now()
	for time.Since(start) < 4*time.Millisecond {
		err := e.New()
		if err.Frames() != nil {
			frames++
		} else {
			noFrames++
		}
	}

	require.NotZero(t, frames)
	require.Zero(t, noFrames)
}
