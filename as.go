package jerror

import (
	"errors"

	"github.com/davecgh/go-spew/spew"
)

func As(err error, target interface{}) bool {
	_, ok := target.(**JError)
	if !ok {
		println("not jerror")
		return false
	}

	// perr := target.(error)
	jerr, ok := target.(**JError)
	if !ok {
		println("not jerror")
		return false
	}

	if !(*jerr).instance {
		println("not instance")
		return false
	}

	// cur := jerr
	cur := err
	for {
		// spew.Dump(cur)
		if nerr, ok := cur.(interface{ As(interface{}) bool }); ok {
			// if (*jerr).As(cur) {
			if (nerr).As(&cur) {
				spew.Dump(cur)
				return true
			}
		}

		next := errors.Unwrap(cur)
		if next == nil {
			break
		}
		cur = next
	}

	return false
}
