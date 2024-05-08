/*
Package jerror is a helper to create errors. It supports creating parametrized
messages to give more information on the error and easier way of wrapping.
These errors are compatible with standard errors.Is and errors.Unwrap.

Example:

	ErrCannotOpen := jerror.New("can not open file %s")

	fileName := "file.txt"
	err := os.Open(fileName)
	if err != nil {
		return ErrCannotOpen.New().Args(fileName).Wrap(err)
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
	debug      = true
	stackDepth = 10
	stackSkip  = 3
)

// New creates a new Error with the given message.
func New(message string) *JError {
	err := &JError{
		instance: false,
		message:  message,
	}

	return err
}

// JError contains an error with a message and a unique identifier.
type JError struct {
	instance bool
	message  string
	parent   error
	wrap     error
	Values   map[string]interface{}
	Frames   []Frame
}

// Frame is a stack frame for the error.
type Frame struct {
	Function string
	File     string
	Line     int
}

// New creates a new Error instance and fills the stack trace.
func (j *JError) New() *JError {
	return &JError{
		instance: true,
		message:  j.message,
		parent:   j,
		Frames:   fillFrames(stackSkip, stackDepth),
		Values:   make(map[string]interface{}),
	}
}

func (j *JError) get() *JError {
	if j.instance {
		return j
	}

	return j.New()
}

// Args returns a version of the error with the parameters from the message
// substituted by its args.
func (j *JError) Args(args ...interface{}) *JError {
	j = j.get()
	j.message = fmt.Sprintf(j.message, args...)
	return j
}

// Wrap returns a version of the error wrapping another error.
func (j *JError) Wrap(err error) *JError {
	j = j.get()
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
	if jerr, ok := err.(*JError); ok {
		return jerr == j.parent
	}

	return false
}

// GetString returns a string value from the error.
func (j *JError) GetString(key string) (string, bool) {
	val, ok := j.Values[key].(string)
	return val, ok
}

// GetInt returns an int value from the error.
func (j *JError) GetInt(key string) (int, bool) {
	val, ok := j.Values[key].(int)
	return val, ok
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
