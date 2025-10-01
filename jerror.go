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
	"sync/atomic"
	"time"

	"golang.org/x/exp/maps"
)

var _ error = &JError{}

const (
	debug      = false
	stackDepth = 10
	stackSkip  = 4
)

type Values map[string]any

var (
	throttled bool
	budget    atomic.Int64
	ticker    *time.Ticker

	disabledStack bool
)

// New creates a new Error with the given message.
func New(message string) *JError {
	return NewWithValues(message, nil)
}

// New creates a new Error with the given message and a set of values that
// will be inherited by child errors.
func NewWithValues(message string, values Values) *JError {
	err := &JError{
		instance: false,
		message:  message,
		values:   values,

		// it's faster to copy a prebuilt struct
		pristine: &JError{
			instance: true,
			message:  message,
			values:   nil,
			frames:   nil,
		},
	}

	return err
}

// Throttle limits the maximum number of new errors with stack frames generated
// for each interval. Getting stack frames is expensive and can be a problem
// when a lot of errors are being generated.
func Throttle(interval time.Duration, max int) {
	if ticker != nil {
		ticker.Stop()
	}

	throttled = true
	ticker = time.NewTicker(interval)
	budget.Store(int64(max))

	go func() {
		for throttled {
			<-ticker.C
			budget.Store(int64(max))
		}
	}()
}

// Unthrottle disables frame generation throttling for new errors.
func Unthrottle() {
	ticker.Stop()
	throttled = false
}

// JError contains an error with a message and a unique identifier.
type JError struct {
	instance bool
	message  string
	parent   error
	wrap     error
	values   Values
	frames   []Frame

	pristine *JError
}

// Frame is a stack frame for the error.
type Frame struct {
	Function string
	File     string
	Line     int
}

// New creates a new Error instance and fills the stack trace.
func (j *JError) New() *JError {
	return j.newStack(stackSkip)
}

func (j *JError) newStack(stackSkip int) *JError {
	n := *j.pristine

	if len(j.values) > 0 {
		n.values = maps.Clone(j.values)
	}

	if !disabledStack && (!throttled || budget.Add(-1) >= 0) {
		n.frames = fillFrames(stackSkip, stackDepth)
	}

	return &n
}

func (j *JError) get() *JError {
	if j.instance {
		return j
	}

	return j.newStack(stackSkip + 1)
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
		return jerr == j.parent || jerr.parent == j.parent
	}

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

	last := Last(j)
	if last != nil && last != j {
		attrs = append(attrs, last.SlogAttributes("last_jerror", true))
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
