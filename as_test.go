package jerror

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAs(t *testing.T) {
	require := require.New(t)

	ErrAs1 := New("1")
	ErrAs2 := New("2")

	err1 := ErrAs1.New().Set("key", 1)
	err2 := ErrAs2.New().Set("key", 2).Wrap(err1)

	err := ErrAs1.New().Set("test", "test")
	ok := As(err2, &err)
	require.True(ok)
	val, ok := err.GetInt("key")
	require.True(ok)
	require.Equal(1, val)

	err = ErrAs2.New()
	ok = As(err2, &err)
	require.True(ok)
	val, ok = err.GetInt("key")
	require.True(ok)
	require.Equal(2, val)

	// should not modify an error without a parent
	ok = As(err2, &ErrAs1)
	require.False(ok)

	// should not modify a nil pointer
	ok = As(err2, nil)
	require.False(ok)

	// should not modify a pointer to a nil JError
	var nilJerror *JError
	ok = As(err2, &nilJerror)
	require.False(ok)
}
