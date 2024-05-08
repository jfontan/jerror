# jerror

Package jerror is a helper to create errors. It supports creating parametrized
messages to give more information on the error and easier way of wrapping.
These errors are compatible with standard `errors.Is` and `errors.Unwrap`.

Example:

```go
    ErrCannotOpen := jerror.New("can not open file %s")

    fileName := "file.txt"
    err := os.Open(fileName)
    if err != nil {
        return ErrCannotOpen.New().Args(fileName).Wrap(err)
    }
```

It also generates a stack trace when the error is created, so you can see where the error was created.

```go
    ErrExample := jerror.New("error")
    err := ErrExample.New()
    spew.Dump(err.Frames())
    // ([]jerror.Frame) (len=3 cap=4) {
    //  (jerror.Frame) {
    //   Function: (string) (len=35) "github.com/jfontan/jerror.TestStack",
    //   File: (string) (len=44) "/home/jfontan/projects/jerror/jerror_test.go",
    //   Line: (int) 97
    //  },
    //  (jerror.Frame) {
    //   Function: (string) (len=15) "testing.tRunner",
    //   File: (string) (len=34) "/usr/lib/go/src/testing/testing.go",
    //   Line: (int) 1689
    //  },
    //  (jerror.Frame) {
    //   Function: (string) (len=14) "runtime.goexit",
    //   File: (string) (len=35) "/usr/lib/go/src/runtime/asm_amd64.s",
    //   Line: (int) 1695
    //  }
    // }
```
