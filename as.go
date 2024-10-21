package jerror

import "errors"

// As retrieves the first JError of the same type as tarjet. To use it create an
// error of the same type you want to get and pass a pointer to it. For example:
//
//	TestErr := jerror.New("test error")
//	target := TestErr.New()
//	ok := jerror.As(err, &target)
func As(err error, target **JError) bool {
	// do not modify an jerror without a parent or an empty one
	if target == nil || *target == nil || (*target).parent == nil {
		return false
	}

	for err != nil {
		if jerr, ok := err.(*JError); ok {
			if (*target).Is(jerr) {
				*target = jerr
				return true
			}
		}

		err = errors.Unwrap(err)
	}

	return false
}
