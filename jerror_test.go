package jerror

import (
	"errors"
	"fmt"
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

	err := jerr.Args("one", 2)
	require.EqualError(err, "args: one, 2")
}

func TestWrap(t *testing.T) {
	require := require.New(t)

	err := fmt.Errorf("standard error")
	jerr := New("jerror error")
	jerr2 := New("another error")

	require.True(errors.Is(jerr, jerr))
	require.False(errors.Is(jerr, err))

	jerrw := jerr.Wrap(err)
	require.True(errors.Is(jerrw, err))
	require.EqualError(jerrw, "jerror error: standard error")

	errw := fmt.Errorf("error: %w", jerrw)
	require.True(errors.Is(errw, jerr))
	require.True(errors.Is(errw, err))

	jerrw2 := jerr2.Wrap(errw)
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
		JError: ErrTest.Wrap(err),
		Code:   code,
	}
}

func ErrEmbedNew(code int) *ErrEmbed {
	return &ErrEmbed{
		JError: ErrTest.Args(),
		Code:   code,
	}
}

func TestEmbed(t *testing.T) {
	require.True(t, errors.Is(ErrEmbedNew(42), ErrTest))

	jerr := ErrEmbedWrap(42, ErrNormal)
	oerr := fmt.Errorf("outer error: %w", jerr)

	require.True(t, errors.Is(oerr, ErrNormal))
	require.True(t, errors.Is(oerr, ErrTest))
}
