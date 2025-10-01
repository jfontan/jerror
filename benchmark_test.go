package jerror

import (
	"errors"
	"testing"
)

var benchError = New("benchmark")

func BenchmarkStack(b *testing.B) {
	disabledStack = false
	b.Run("WithStack", func(b *testing.B) {
		for b.Loop() {
			_ = deepError(1000)
		}
	})

	disabledStack = true
	b.Run("WithoutStack", func(b *testing.B) {
		for b.Loop() {
			_ = deepError(1000)
		}
	})

	disabledStack = false

	b.Run("Standard", func(b *testing.B) {
		for b.Loop() {
			_ = deepErrorStandard(1000)
		}
	})
}

func deepError(l int) *JError {
	if l <= 0 {
		return benchError.New()
	}

	return deepError(l - 1)
}

func deepErrorStandard(l int) error {
	if l <= 0 {
		return errors.New("standard")
	}

	return deepErrorStandard(l - 1)
}
