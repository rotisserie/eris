# eris ![Logo][eris-logo]

[![GoDoc][doc-img]][doc] [![Build][ci-img]][ci] [![Coverage Status][cov-img]][cov] [![GoReport][report-img]][report] [![Discord][chat-img]][chat] [![Mentioned in Awesome Go][awesome-img]][awesome]

Package `eris` provides a better way to handle, trace, and log errors in Go.

`go get github.com/rotisserie/eris`

<!-- toc -->

- [Why you'll want to switch to eris](#why-youll-want-to-switch-to-eris)
- [Using eris](#using-eris)
  * [Creating errors](#creating-errors)
  * [Wrapping errors](#wrapping-errors)
  * [Formatting and logging errors](#formatting-and-logging-errors)
  * [Interpreting eris stack traces](#interpreting-eris-stack-traces)
  * [Inspecting errors](#inspecting-errors)
  * [Formatting with custom separators](#formatting-with-custom-separators)
  * [Writing a custom output format](#writing-a-custom-output-format)
  * [Sending error traces to Sentry](#sending-error-traces-to-sentry)
- [Comparison to other packages (e.g. pkg/errors)](#comparison-to-other-packages-eg-pkgerrors)
  * [Error formatting and stack traces](#error-formatting-and-stack-traces)
- [Migrating to eris](#migrating-to-eris)
- [Contributing](#contributing)

<!-- tocstop -->

## Why you'll want to switch to eris

Named after the Greek goddess of strife and discord, this package is designed to give you more control over error handling via error wrapping, stack tracing, and output formatting. `eris` was inspired by a simple question: what if you could fix a bug without wasting time replicating the issue or digging through the code?

`eris` is intended to help developers diagnose issues faster. The [example](https://github.com/rotisserie/examples/blob/master/eris/logging/example.go) that generated the output below simulates a realistic error handling scenario and demonstrates how to wrap and log errors with minimal effort. This specific error occurred because a user tried to access a file that can't be located, and the output shows a clear path from the source to the top of the call stack.

```json
{
  "error":{
    "root":{
      "message":"error internal server",
      "stack":[
        "main.GetRelPath:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:61",
        "main.ProcessResource:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:82",
        "main.ProcessResource:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:85",
        "main.main:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:143"
      ]
    },
    "wrap":[
      {
        "message":"Rel: can't make ./some/malformed/absolute/path/data.json relative to /Users/roti/",
        "stack":"main.GetRelPath:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:61"
      },
      {
        "message":"failed to get relative path for resource 'res2'",
        "stack":"main.ProcessResource:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:85"
      }
    ]
  },
  "level":"error",
  "method":"ProcessResource",
  "msg":"method completed with error",
  "time":"2020-01-16T11:20:01-05:00"
}
```

Many of the methods in this package will look familiar if you've used [pkg/errors](https://github.com/pkg/errors) or [xerrors](https://github.com/golang/xerrors), but `eris` employs some additional tricks during error wrapping and unwrapping that greatly improve the readability of the stack which should make debugging easier. This package also takes a unique approach to formatting errors that allows you to write custom formats that conform to your error or log aggregator of choice. You can find more information on the differences between `eris` and `pkg/errors` [here](#comparison-to-other-packages-eg-pkgerrors).

## Using eris

### Creating errors

Creating errors is simple via [`eris.New`](https://godoc.org/github.com/rotisserie/eris#New) and [`eris.NewGlobal`](https://godoc.org/github.com/rotisserie/eris#NewGlobal).

```golang
var (
  // global error values can be useful when wrapping errors or inspecting error types
  ErrInternalServer = eris.NewGlobal("error internal server")
)

func (req *Request) Validate() error {
  if req.ID == "" {
    // or return a new error at the source if you prefer
    return eris.New("error bad request")
  }
  return nil
}
```

### Wrapping errors

[`eris.Wrap`](https://godoc.org/github.com/rotisserie/eris#Wrap) adds context to an error while preserving the original error.

```golang
relPath, err := GetRelPath("/Users/roti/", resource.AbsPath)
if err != nil {
  // wrap the error if you want to add more context
  return nil, eris.Wrapf(err, "failed to get relative path for resource '%v'", resource.ID)
}
```

### Formatting and logging errors

[`eris.ToString`](https://godoc.org/github.com/rotisserie/eris#ToString) and [`eris.ToJSON`](https://godoc.org/github.com/rotisserie/eris#ToJSON) should be used to log errors with the default format (shown above). The JSON method returns a `map[string]interface{}` type for compatibility with Go's `encoding/json` package and many common JSON loggers (e.g. [logrus](https://github.com/sirupsen/logrus)).

```golang
// format the error to JSON with the default format and stack traces enabled
formattedJSON := eris.ToJSON(err, true)
fmt.Println(json.Marshal(formattedJSON)) // marshal to JSON and print
logger.WithField("error", formattedJSON).Error() // or ideally, pass it directly to a logger

// format the error to a string and print it
formattedStr := eris.ToString(err, true)
fmt.Println(formattedStr)
```

`eris` also enables control over the [default format's separators](#formatting-with-custom-separators) and allows advanced users to write their own [custom output format](#writing-a-custom-output-format).

### Interpreting eris stack traces

Errors created with this package contain stack traces that are managed automatically. They're currently mandatory when creating and wrapping errors but optional when printing or logging. The stack trace and all wrapped layers follow the same order as Go's `runtime` package, which means that the root cause of the error is shown first.

```golang
{
  "root":{
    "message":"error bad request", // root cause
    "stack":[
      "main.(*Request).Validate:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:28", // location of the root
      "main.(*Request).Validate:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:29", // location of Wrap call
      "main.ProcessResource:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:71",
      "main.main:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:143"
    ]
  },
  "wrap":[
    {
      "message":"received a request with no ID", // additional context
      "stack":"main.(*Request).Validate:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:29" // location of Wrap call
    }
  ]
}
```

### Inspecting errors

The `eris` package provides a couple ways to inspect and compare error types. [`eris.Is`](https://godoc.org/github.com/rotisserie/eris#Is) returns true if a particular error appears anywhere in the error chain. Currently, it works simply by comparing error messages with each other. If an error contains a particular message (e.g. `"error not found"`) anywhere in its chain, it's defined to be that error type.

```golang
ErrNotFound := eris.NewGlobal("error not found")
_, err := db.Get(id)
// check if the resource was not found
if eris.Is(err, ErrNotFound) {
  // return the error with some useful context
  return eris.Wrapf(err, "error getting resource '%v'", id)
}
```

[`eris.Cause`](https://godoc.org/github.com/rotisserie/eris#Cause) unwraps an error until it reaches the cause, which is defined as the first (i.e. root) error in the chain.

```golang
ErrNotFound := eris.NewGlobal("error not found")
_, err := db.Get(id)
// compare the cause to some sentinel value
if eris.Cause(err) == ErrNotFound {
  // return the error with some useful context
  return eris.Wrapf(err, "error getting resource '%v'", id)
}
```

### Formatting with custom separators

For users who need more control over the error output, `eris` allows for some control over the separators between each piece of the output via the [`eris.Format`](https://godoc.org/github.com/rotisserie/eris#Format) type. Currently, the default order of the error and stack trace output is rigid. If this isn't flexible enough for your needs, see the [custom output format](#writing-a-custom-output-format) section below. To format errors with custom separators, you can define and pass a format object to [`eris.ToCustomString`](https://godoc.org/github.com/rotisserie/eris#ToCustomString) or [`eris.ToCustomJSON`](https://godoc.org/github.com/rotisserie/eris#ToCustomJSON).

```golang
// format the error to a string with custom separators
formattedStr := eris.ToCustomString(err, Format{
  WithTrace: true,     // flag that enables stack trace output
  MsgStackSep: "\n",   // separator between error messages and stack frame data
  PreStackSep: "\t",   // separator at the beginning of each stack frame
  StackElemSep: " | ", // separator between elements of each stack frame
  ErrorSep: "\n",      // separator between each error in the chain
})
fmt.Println(formattedStr)

// example output:
// unexpected EOF
//   main.readFile | .../example/main.go | 6
//   main.parseFile | .../example/main.go | 12
//   main.main | .../example/main.go | 20
// error reading file 'example.json'
//   main.readFile | .../example/main.go | 6
```

### Writing a custom output format

`eris` also allows advanced users to construct custom error strings or objects in case the default error doesn't fit their requirements. The [`UnpackedError`](https://godoc.org/github.com/rotisserie/eris#UnpackedError) object provides a convenient and developer friendly way to store and access existing error traces. The `ErrRoot` and `ErrChain` fields correspond to the root error and wrap error chain, respectively. If any other error type is unpacked, it will appear in the `ExternalErr` field. You can access all of the information contained in an error via [`eris.Unpack`](https://godoc.org/github.com/rotisserie/eris#Unpack).

```golang
// get the unpacked error object
uErr := eris.Unpack(err)
// send only the root error message to a logging server instead of the complete error trace
sentry.CaptureMessage(uErr.ErrRoot.Msg)
```

### Sending error traces to Sentry

`eris` supports sending your error traces to [Sentry](https://sentry.io/) using the Sentry Go [client SDK](https://github.com/getsentry/sentry-go). You can run the example that generated the following output on Sentry UI using the command `go run eris/sentry/example.go -dsn=<DSN>` in our [examples](https://github.com/rotisserie/examples) repository.

```
*eris.wrapError: test: wrap 1: wrap 2: wrap 3
  File "main.go", line 19, in Example
    return eris.New("test")
  File "main.go", line 23, in WrapExample
    err := Example()
  File "main.go", line 25, in WrapExample
    return eris.Wrap(err, "wrap 1")
  File "main.go", line 31, in WrapSecondExample
    err := WrapExample()
  File "main.go", line 33, in WrapSecondExample
    return eris.Wrap(err, "wrap 2")
  File "main.go", line 44, in main
    err := WrapSecondExample()
  File "main.go", line 45, in main
    err = eris.Wrap(err, "wrap 3")
```

## Comparison to other packages (e.g. pkg/errors)

### Error formatting and stack traces

Readability is a major design requirement for `eris`. In addition to the JSON output shown above, `eris` also supports formatting errors to a simple string.

```
error not found
  main.GetResource:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:52
  main.ProcessResource:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:76
  main.main:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:143
failed to get resource 'res1'
  main.GetResource:/Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:52
```

The `eris` error stack is designed to be easier to interpret than other error handling packages, and it achieves this by omitting extraneous information and avoiding unnecessary repetition. The stack trace above omits calls from Go's `runtime` package and includes just a single frame for wrapped layers which are inserted into the root error stack trace in the correct order. `eris` also correctly handles and updates stack traces for global error values.

The output of `pkg/errors` for the same error is shown below. In this case, the root error stack trace is incorrect because it was declared as a global value, and it includes several extraneous lines from the `runtime` package. The output is also much more difficult to read and does not allow for custom formatting.

```
error not found
main.init
  /Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:18
runtime.doInit
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/proc.go:5222
runtime.main
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/proc.go:190
runtime.goexit
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/asm_amd64.s:1357
failed to get resource 'res1'
main.GetResource
  /Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:52
main.ProcessResource
  /Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:76
main.main
  /Users/roti/go/src/github.com/rotisserie/examples/eris/logging/example.go:143
runtime.main
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/proc.go:203
runtime.goexit
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/asm_amd64.s:1357
```

## Migrating to eris

Migrating to `eris` should be a very simple process. If it doesn't offer something that you currently use from existing error packages, feel free to submit an issue to us. If you don't want to refactor all of your error handling yet, `eris` should work relatively seamlessly with your existing error types. Please submit an issue if this isn't the case for some reason.

Many of your dependencies will likely still use [pkg/errors](https://github.com/pkg/errors) for error handling. Currently, when external error types are wrapped with additional context, the original error is flattened (via `err.Error()`) and used to create a root error. This adds a stack trace for the error and allows it to function more seamlessly with the rest of the `eris` package. However, we're looking into potentially integrating with other error packages to unwrap and format external errors.

## Contributing

If you'd like to contribute to `eris`, we'd love your input! Please submit an issue first so we can discuss your proposal. We're also available to discuss potential issues and features on our [Discord channel](https://discord.gg/gMfXeXR).

-------------------------------------------------------------------------------

Released under the [MIT License].

[MIT License]: LICENSE.txt
[eris-logo]: https://cdn.emojidex.com/emoji/hdpi/minecraft_golden_apple.png?1511637499
[doc-img]: https://img.shields.io/badge/godoc-eris-blue
[doc]: https://godoc.org/github.com/rotisserie/eris
[ci-img]: https://github.com/rotisserie/eris/workflows/eris/badge.svg
[ci]: https://github.com/rotisserie/eris/actions
[cov-img]: https://codecov.io/gh/rotisserie/eris/branch/master/graph/badge.svg
[cov]: https://codecov.io/gh/rotisserie/eris
[report-img]: https://goreportcard.com/badge/github.com/rotisserie/eris
[report]: https://goreportcard.com/report/github.com/rotisserie/eris
[chat-img]: https://img.shields.io/discord/659952923073183749?color=738adb&label=discord&logo=discord
[chat]: https://discord.gg/gMfXeXR
[awesome-img]: https://awesome.re/mentioned-badge.svg
[awesome]: https://github.com/avelino/awesome-go#error-handling
