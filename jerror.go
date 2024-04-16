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
	"runtime"
	"time"
)

var _ error = &JError{}

const (
	debug = true
)

// JErrorBase is used to construct specific errors.
type JErrorBase struct {
	message string
}

// New creates a new Error and fills the stack trace.
func (j *JErrorBase) New() *JError {
	return &JError{
		message: j.message,
		parent:  j,
		Frames:  fillFrames(3, 10),
	}
}

// Error implements error interface.
func (j *JErrorBase) Error() string {
	return j.message
}

// New creates a new Error with the given message.
func New(message string) *JErrorBase {
	err := &JErrorBase{
		message: message,
	}

	return err
}

// JError contains an error with a message and a unique identifier.
type JError struct {
	message string
	parent  error
	wrap    error
	Frames  []Frame
}

// Frame is a stack frame for the error.
type Frame struct {
	Function string
	File     string
	Line     int
}

// Args returns a version of the error with the parameters from the message
// substituted by its args.
func (j *JError) Args(args ...interface{}) *JError {
	j.message = fmt.Sprintf(j.message, args...)
	return j
}

// Wrap returns a version of the error wrapping another error.
func (j *JError) Wrap(err error) *JError {
	j.wrap = err
	return j
}

// Error implements error interface.
func (j *JError) Error() string {
	wmesg := ""
	if j.wrap != nil {
		wmesg = ": " + j.wrap.Error()
	}
	return j.message + wmesg
}

// Unwrap implements error interface.
func (j *JError) Unwrap() error {
	return j.wrap
}

// Is implements error interface.
func (j *JError) Is(err error) bool {
	if jerr, ok := err.(*JErrorBase); ok {
		return jerr == j.parent
	}

	return false
}

func fillFrames(skip, depth int) []Frame {
	if debug {
		start := time.Now()
		defer func() {
			fmt.Printf("fillFrames took %v\n", time.Since(start).String())
		}()
	}

	pc := make([]uintptr, depth)
	num := runtime.Callers(skip, pc)
	callerFrames := runtime.CallersFrames(pc[:num])

	var frames []Frame
	for {
		frame, more := callerFrames.Next()
		frames = append(frames, Frame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
	}

	return frames
}
