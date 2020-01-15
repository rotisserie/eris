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
  * [Writing your own custom format](#writing-your-own-custom-format)
- [Comparison to other packages (e.g. pkg/errors)](#comparison-to-other-packages-eg-pkgerrors)
  * [Error formatting and stack traces](#error-formatting-and-stack-traces)
- [Migrating to eris](#migrating-to-eris)
- [Contributing](#contributing)

<!-- tocstop -->

## Why you'll want to switch to eris

Named after the Greek goddess of strife and discord, this package is designed to give you more control over error handling via error wrapping, stack tracing, and output formatting. `eris` was inspired by a simple question: what if you could fix a bug without wasting time replicating the issue or digging through the code?

`eris` is intended to help developers diagnose issues faster. The [example](example_logger_test.go) that generated the output below simulates a realistic error handling scenario and demonstrates how to wrap and log errors with minimal effort. This specific error occurred because a user tried to access a file that can't be located, and the output shows a clear path from the source to the top of the call stack.

```json
{
  "error":{
    "root":{
      "message":"error internal server",
      "stack":[
        "eris_test.GetRelPath:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:58",
        "eris_test.ProcessResource:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:79",
        "eris_test.ProcessResource:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:82",
        "eris_test.Example_logger:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:140",
      ]
    },
    "wrap":[
      {
        "message":"Rel: can't make ./some/malformed/absolute/path/data.json relative to /Users/roti/",
        "stack":"eris_test.GetRelPath:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:58"
      },
      {
        "message":"failed to get relative path for resource 'res2'",
        "stack":"eris_test.ProcessResource:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:82"
      }
    ]
  },
  "level":"error",
  "method":"ProcessResource",
  "msg":"method completed with error",
  "time":"2020-01-12T13:50:00-05:00"
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

`eris` also enables control over the [default format's separators](#formatting-with-custom-separators) and allows advanced users to write their own [custom formats](#writing-your-own-custom-format).

### Interpreting eris stack traces

Errors created with this package contain stack traces that are managed automatically. They're currently mandatory when creating and wrapping errors but optional when printing or logging. The stack trace and all wrapped layers follow the same order as Go's `runtime` package, which means that the root cause of the error is shown first.

```golang
{
  "root":{
    "message":"error bad request", // root cause
    "stack":[
      "eris_test.(*Request).Validate:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:25", // location of the root
      "eris_test.(*Request).Validate:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:26", // location of Wrap call
      "eris_test.ProcessResource:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:68",
      "eris_test.Example_logger:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:140",
    ]
  },
  "wrap":[
    {
      "message":"received a request with no ID", // additional context
      "stack":"eris_test.(*Request).Validate:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:26" // location of Wrap call
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

The default format in `eris` is returned by the method [`NewDefaultFormat()`](https://godoc.org/github.com/rotisserie/eris#NewDefaultFormat). Below you can see what a default formatted error in `eris` might look like.

Errors printed without trace using `fmt.Printf("%v\n", err)`

```
even more context: additional context: root error
```

Errors printed with trace using `fmt.Printf("%+v\n", err)`

```
even more context
        eris_test.setupTestCase: ../eris/eris_test.go: 17
additional context
        eris_test.setupTestCase: ../eris/eris_test.go: 17
root error
        eris_test.setupTestCase: ../eris/eris_test.go: 17
        eris_test.TestErrorFormatting: ../eris/eris_test.go: 226
        testing.tRunner: ../go1.11.4/src/testing/testing.go: 827
        runtime.goexit: ../go1.11.4/src/runtime/asm_amd64.s: 1333
```

'eris' also provides developers a way to define a custom format to print the errors. The [`Format`](https://godoc.org/github.com/rotisserie/eris#Format) object defines separators for various components of the error/trace and can be passed to utility methods for printing string and JSON formats.

### Writing your own custom format

The [`UnpackedError`](https://godoc.org/github.com/rotisserie/eris#UnpackedError) object provides a convenient and developer friendly way to store and access existing error traces. The `ErrChain` and `ErrRoot` fields correspond to `wrapError` and `rootError` types, respectively. If any other error type is unpacked, it will appear in the ExternalErr field.

The [`Unpack()`](https://godoc.org/github.com/rotisserie/eris#Unpack) method returns the corresponding `UnpackedError` object for a given error. This object can also be converted to string and JSON for logging and printing error traces. This can be done by using the methods [`ToString()`](https://godoc.org/github.com/rotisserie/eris#UnpackedError.ToString) and [`ToJSON()`](https://godoc.org/github.com/rotisserie/eris#UnpackedError.ToJSON). Note the `ToJSON()` method returns a `map[string]interface{}` type which can be marshalled to JSON using the `encoding/json` package.

## Comparison to other packages (e.g. pkg/errors)

### Error formatting and stack traces

Readability is a major design requirement for `eris`. In addition to the JSON output shown above, `eris` also supports formatting errors to a simple string.

```
error not found
  eris_test.GetResource:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:53
  eris_test.ProcessResource:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:77
  eris_test.Example_logger:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:144
  eris_test.TestExample_logger:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:162
failed to get resource 'res1'
  eris_test.GetResource:/Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:53
```

The `eris` error stack is designed to be easier to interpret than other error handling packages, and it achieves this by omitting extraneous information and avoiding unnecessary repetition. The stack trace above omits calls from Go's `runtime` package and includes just a single frame for wrapped layers which are inserted into the root error stack trace in the correct order. `eris` also correctly handles and updates stack traces for global error values.

The output of `pkg/errors` for the same error is shown below. In this case, the root error stack trace is incorrect because it was declared as a global value, and it includes several extraneous lines from the `runtime` package. The output is also much more difficult to read and does not allow for custom formatting.

```
error not found
github.com/rotisserie/eris_test.init
  /Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:19
runtime.doInit
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/proc.go:5222
runtime.doInit
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/proc.go:5217
runtime.main
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/proc.go:190
runtime.goexit
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/asm_amd64.s:1357
failed to get resource 'res1'
github.com/rotisserie/eris_test.GetResource
  /Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:53
github.com/rotisserie/eris_test.ProcessResource
  /Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:77
github.com/rotisserie/eris_test.Example_logger
  /Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:144
github.com/rotisserie/eris_test.TestExample_logger
  /Users/roti/go/src/github.com/rotisserie/eris/example_logger_test.go:162
testing.tRunner
  /usr/local/Cellar/go/1.13.6/libexec/src/testing/testing.go:909
runtime.goexit
  /usr/local/Cellar/go/1.13.6/libexec/src/runtime/asm_amd64.s:1357
```

## Migrating to eris

Migrating to `eris` should be a very simple process. If it doesn't offer something that you currently use from existing error packages, feel free to submit an issue to us. If you don't want to refactor all of your error handling yet, `eris` should work relatively seamlessly with your existing error types. Please submit an issue if this isn't the case for some reason.

Many of your dependencies will likely still use [pkg/errors](https://github.com/pkg/errors) for error handling. Currently, when external error types are wrapped with additional context, the original error is flattened (via `err.Error()`) and used to create a root error. This adds a stack trace for the error and allows it to function more seamlessly with the rest of the `eris` package. However, we're looking into potentially integrating with other error packages to unwrap and format external errors.

## Contributing

If you'd like to contribute to `eris`, we'd love your input! Please submit an issue first so we can discuss your proposal. We're also available to discuss potential issues and features on our [Discord channel](https://discordapp.com/channels/659952923073183749/659952923073183756).

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
[chat-img]: https://img.shields.io/discord/659952923073183749
[chat]: https://discordapp.com/channels/659952923073183749/659952923073183756
[awesome-img]: https://awesome.re/mentioned-badge.svg
[awesome]: https://github.com/avelino/awesome-go#error-handling
