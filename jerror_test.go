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
