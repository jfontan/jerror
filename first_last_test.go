package jerror

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirstLast(t *testing.T) {
	jerr2 := New("Jerror 2").Set("key", 2).Wrap(fmt.Errorf("err 3"))
	jerr1 := New("jerror 1").Set("key", 1).Wrap(fmt.Errorf("err 2: %w", jerr2))

	err := fmt.Errorf("err 1: %w", jerr1)

	first := First(err)
	require.Error(t, first)
	val, ok := first.GetInt("key")
	require.True(t, ok)
	require.Equal(t, val, 1)

	last := Last(err)
	require.Error(t, last)
	val, ok = last.GetInt("key")
	require.True(t, ok)
	require.Equal(t, val, 2)

	err = fmt.Errorf("err1: %w",
		fmt.Errorf("err2: %w",
			fmt.Errorf("err3"),
		),
	)

	none := First(err)
	require.Nil(t, none)

	none = Last(err)
	require.Nil(t, none)
}
