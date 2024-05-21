package jerror

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
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
	err.Set("key", "value")
	err.Set("key2", 42)
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

func TestAs(t *testing.T) {
	require := require.New(t)
	spew.Config.DisableMethods = true

	ErrAs1 := New("1")
	ErrAs2 := New("2")

	err1 := ErrAs1.New().Set("key", 1)
	err2 := ErrAs2.New().Set("key", 2).Wrap(err1)

	err := ErrAs1.New().Set("test", "test")
	// ok := errors.As(err2, &err)
	ok := As(err2, &err)
	spew.Dump(err)
	require.True(ok)
	println(err.Error())
	val, ok := err.GetInt("key")
	require.True(ok)
	require.Equal(1, val)

	err = ErrAs2.New()
	// ok = errors.As(err2, &err)
	ok = As(err2, &err)
	require.True(ok)
	val, ok = err.GetInt("key")
	require.True(ok)
	require.Equal(2, val)
}
