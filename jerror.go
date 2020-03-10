/*
Package jerror is a helper to create errors. It supports creating parametrized
messages to give more information on the error and easier way of wrapping.
These errors are compatible with standard errors.Is and errors.Unwrap.

Example:

	ErrCannotOpen := jerror.New("can not open file %s")

	fileName := "file.txt"
	err := os.Open(fileName)
	if err != nil {
		return ErrCannotOpen.Args(fileName).Wrap(err)
	}
*/
package jerror

import (
	"fmt"
	"sync"
)

var (
	m  sync.Mutex
	id int
)

var _ error = &Error{}

// Error contains an error with a message and a unique identifier.
type Error struct {
	message string
	id      int
	wrap    error
}

// New creates a new Error with the given message.
func New(message string) *Error {
	m.Lock()
	defer m.Unlock()

	err := &Error{
		message: message,
		id:      id,
	}

	id++

	return err
}

// Args returns a version of the error with the parameters from the message
// substituted by its args.
func (j *Error) Args(args ...interface{}) *Error {
	return &Error{
		message: fmt.Sprintf(j.message, args...),
		id:      j.id,
	}
}

// Wrap returns a version of the error wrapping another error.
func (j *Error) Wrap(err error) *Error {
	return &Error{
		message: j.message,
		id:      j.id,
		wrap:    err,
	}
}

// Error implements error interface.
func (j *Error) Error() string {
	wmesg := ""
	if j.wrap != nil {
		wmesg = ": " + j.wrap.Error()
	}
	return j.message + wmesg
}

// Unwrap implements error interface.
func (j *Error) Unwrap() error {
	return j.wrap
}

// Is implements error interface.
func (j *Error) Is(err error) bool {
	if jerr, ok := err.(*Error); ok {
		return jerr.id == j.id
	}

	return false
}
