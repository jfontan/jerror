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

You can set and retrieve values to a base error:

```go
jerr := jerror.New("jerror example")
err := jerr.Set("key", "value")
val, _ := err.GetString("key")
fmt.Println(val)
// value
```

It can also generate `slog` compatible attributes to log the error and of there is older errors it also gives information about the oldest error:

```go
jerr2 := jerror.New("jerror 2").Set("key", "value 2")
jerr1 := jerror.New("jerror 1").Set("key", "value 1").Wrap(jerr2)

var buf bytes.Buffer
log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{}))
log.Error("message", jerr1.SlogAttributes("data", true))

var j any
_ = json.Unmarshal(buf.Bytes(), &j)
i, _ := json.MarshalIndent(j, "", "  ")
fmt.Println(string(i))
// {
//   "data": {
//     "error": "jerror 1: jerror 2",
//     "last_jerror": {
//       "error": "jerror 2",
//       "stack": {
//         "0": "github.com/jfontan/jerror/examples.TestSlog github.com/jfontan/jerror/examples/examples_test.go:14",
//         "1": "testing.tRunner testing/testing.go:1689",
//         "2": "runtime.goexit runtime/asm_amd64.s:1695"
//       },
//       "values": {
//         "key": "value 2"
//       }
//     },
//     "stack": {
//       "0": "github.com/jfontan/jerror/examples.TestSlog github.com/jfontan/jerror/examples/examples_test.go:15",
//       "1": "testing.tRunner testing/testing.go:1689",
//       "2": "runtime.goexit runtime/asm_amd64.s:1695"
//     },
//     "values": {
//       "key": "value 1"
//     }
//   },
//   "level": "ERROR",
//   "msg": "message",
//   "time": "2024-05-30T16:26:33.292105723+02:00"
// }
```
