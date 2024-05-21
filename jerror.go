/*
Package jerror is a helper to create errors. It supports creating parametrized
messages to give more information on the error and easier way of wrapping.
These errors are compatible with standard errors.Is and errors.Unwrap.

More information at http://github.com/jfontan/jerror

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
	"log/slog"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/exp/maps"
)

var _ error = &JError{}

const (
	debug      = false
	stackDepth = 10
	stackSkip  = 3
)

// New creates a new Error with the given message.
func New(message string) *JError {
	err := &JError{
		instance: false,
		message:  message,
	}
	err.parent = err

	return err
}

// JError contains an error with a message and a unique identifier.
type JError struct {
	instance bool
	message  string
	parent   error
	wrap     error
	values   map[string]interface{}
	frames   []Frame
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
		frames:   fillFrames(stackSkip, stackDepth),
		values:   make(map[string]interface{}),
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
	println("Is", err)
	if jerr, ok := err.(*JError); ok {
		if jerr.parent == j.parent {
			// spew.Dump(j)
			// spew.Dump(jerr)
			return true
		}
	}

	return false
}

// As implements error interface.
func (j *JError) As(target interface{}) bool {
	// println("As", target.(*JError).Error())
	// val := reflect.ValueOf(target)
	// typ := val.Type()
	// targetType := typ.Elem()

	// if targetType.Kind() == reflect.Interface && targetType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
	spew.Dump(target)
	if jerr, ok := target.(**JError); ok {
		println("correct")
		// spew.Dump(j)
		// spew.Dump(target)
		// err := (target.(*JError))
		if j.Is(*jerr) {
			// err = j
			*j = **jerr
			spew.Dump(j)
			println("IS!!!")
		}
		return true
	}

	// if jerr, ok := target.(**JError); ok && j.Is(*jerr) {
	// 	println("correct")
	// 	// spew.Dump(j)
	// 	// spew.Dump(jerr)
	// 	// spew.Dump(target)
	// 	// jerr = j

	// 	target = &j
	// 	spew.Dump(j)
	// 	spew.Dump(target)
	// 	return true
	// }

	return false
}

// Set sets a value in the error.
func (j *JError) Set(key string, value interface{}) *JError {
	j = j.get()
	j.values[key] = value
	return j
}

// Get returns a value from the error.
func (j *JError) Get(key string) (interface{}, bool) {
	val, ok := j.values[key]
	return val, ok
}

// GetString returns a string value from the error.
func (j *JError) GetString(key string) (string, bool) {
	val, ok := j.values[key].(string)
	return val, ok
}

// GetInt returns an int value from the error.
func (j *JError) GetInt(key string) (int, bool) {
	val, ok := j.values[key].(int)
	return val, ok
}

// Frames returns the stack frames of the error.
func (j *JError) Frames() []Frame {
	return j.frames
}

func (j *JError) SlogAttributes(group string, error bool) slog.Attr {
	var attrs []any

	if error {
		attrs = append(attrs, slog.String("error", j.Error()))
	}

	if len(j.frames) > 0 {
		var lines []any
		for i, frame := range j.frames {
			lines = append(lines, slog.String(strconv.Itoa(i), fmt.Sprintf(
				"%s %s:%d",
				frame.Function,
				frame.File,
				frame.Line,
			)))
		}
		attrs = append(attrs, slog.Group("stack", lines...))
	}

	// TODO: do not convert values to string
	if len(j.values) > 0 {
		var values []any

		keys := maps.Keys(j.values)
		sort.Strings(keys)

		for _, key := range keys {
			value := j.values[key]
			values = append(values, slog.String(key, fmt.Sprintf("%v", value)))
		}
		attrs = append(attrs, slog.Group("values", values...))
	}

	return slog.Group(group, attrs...)
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
