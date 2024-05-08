package jerror

import (
	"errors"
	"fmt"
	"strings"
	"testing"

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
	require := require.New(t)

	err := New("stack error").New()
	require.Len(err.Frames, 3)

	last := err.Frames[0]
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
	err.Set("key", "value")
	require.EqualError(err, "values error")
	_, ok := orig.Get("key")
	require.False(ok)

	val, ok := err.Get("key")
	require.True(ok)
	require.Equal("value", val)
}
