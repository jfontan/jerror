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
