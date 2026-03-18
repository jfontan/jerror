package jerror

import (
	"errors"
	"fmt"
	"slices"
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

func TestChain(t *testing.T) {
	jerr1 := New("jerr 1")
	jerr2 := New("jerr 2")
	serr1 := errors.New("nerr 1")
	serr2 := errors.New("nerr 2")

	testChain := []struct {
		err    error
		values map[string]string
	}{
		{
			err: jerr1,
			values: map[string]string{
				"name": "jerr1-1",
				"pos":  "0",
			},
		},
		{
			err: serr1,
		},
		{
			err: jerr2,
			values: map[string]string{
				"name": "jerr2-1",
				"pos":  "1",
			},
		},
		{
			err: jerr1,
			values: map[string]string{
				"name": "jerr1-2",
				"pos":  "2",
			},
		},
		{
			err: jerr2,
			values: map[string]string{
				"name": "jerr2-2",
				"pos":  "3",
			},
		},
		{
			err: serr1,
		},
		{
			err: serr2,
		},
	}

	// create chain
	var wrapChain []error
	for _, test := range testChain {
		err := test.err
		if jerr, ok := err.(*JError); ok {
			nerr := jerr.New()
			for k, v := range test.values {
				nerr.Set(k, v)
			}
			err = nerr
		}

		wrapChain = append(wrapChain, err)
	}

	slices.Reverse(wrapChain)

	err := wrapChain[0]
	for _, e := range wrapChain[1:] {
		if jerr, ok := e.(*JError); ok {
			err = jerr.Wrap(err)
		} else {
			err = fmt.Errorf("%s: %w", e.Error(), err)
		}
	}

	var pos int
	for _, err := range Chain(err) {
		// advance until the first jerror if needed
		for {
			require.Less(t, pos, len(testChain))
			if Is(testChain[pos].err) {
				break
			}
			pos++
		}

		test := testChain[pos]
		jerr, ok := test.err.(*JError)
		require.True(t, ok)

		// check that is the same error
		require.True(t, jerr.Is(err))

		// check values
		for k, v := range test.values {
			val, ok := err.GetString(k)
			require.True(t, ok)
			require.Equal(t, v, val)
		}

		pos++
	}
}
