package jerror

import "errors"

// First JError in a wrapped errors chain.
func First(err error) *JError {
	for {
		if err == nil {
			return nil
		}
		if jerr, ok := err.(*JError); ok {
			return jerr
		}

		err = errors.Unwrap(err)
	}
}

// Last JError in a wrapped errors chain.
func Last(err error) *JError {
	var jerr *JError
	for err != nil {
		if jerr2, ok := err.(*JError); ok {
			jerr = jerr2
		}

		err = errors.Unwrap(err)
	}

	return jerr
}

// Chain returns all JErrors in a wrapped error chain.
func Chain(err error) []*JError {
	var jerrs []*JError
	for err != nil {
		if jerr, ok := err.(*JError); ok {
			jerrs = append(jerrs, jerr)
		}

		err = errors.Unwrap(err)
	}

	return jerrs
}
