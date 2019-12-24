# eris 😈

todo: add code coverage/report, CI links, etc.

Package eris provides a better way to handle errors in Go. This package is inspired by a few existing packages: [xerrors](https://github.com/golang/xerrors), [pkg/errors](https://github.com/pkg/errors), and [Go 1.13 errors](https://golang.org/pkg/errors/).

`go get github.com/morningvera/eris`

Check out the [package docs](https://godoc.org/github.com/morningvera/eris) for more detailed information or connect with us on our [Slack channel](https://rotisserieworkspace.slack.com/archives/CS13EC3T6) if you want to discuss anything in depth.

## How is eris different?

Named after the Greek goddess of strife and discord, this package is intended to give you more control over error handling via error wrapping, stack tracing, and output formatting. Basic error wrapping was added in Go 1.13, but it omitted user-friendly `Wrap` methods and built-in stack tracing. Other error packages provide some of the features found in `eris` but without flexible control over error output formatting. This package provides default string and JSON formatters with options to control things like separators and stack trace output. However, it also provides an option to write custom formatters via `eris.Unpack`.

Error wrapping behaves somewhat differently than existing packages. It relies on root errors that contain a full stack trace and wrap errors that contain a single stack frame. When errors from other packages are wrapped, a root error is automatically created before wrapping it with the new context. This allows `eris` to work with other error packages transparently and elimates the need to manage stack traces manually. Unlike other packages, `eris` also works well with global error types by automatically updating stack traces during error wrapping.

## Types of errors

`eris` is concerned with only three different types of errors: root errors, wrap errors, and external errors. Root and wrap errors are defined types in this package and all other error types are external or third-party errors.

Root errors are created via `eris.New` and `eris.Errorf`. Generally, it's a good idea to maintain a set of root errors that are then wrapped with additional context whenever an error of that type occurs. Wrap errors represent a stack of errors that have been wrapped with additional context. Unwrapping these errors via `eris.Unwrap` will return the next error in the stack until a root error is reached. `eris.Cause` will also retrieve the root error.

When external error types are wrapped with additional context, a root error is first created from the original error. This creates a stack trace for the error and allows it to function with the rest of the `eris` package.

## Wrapping errors with additional context

`eris.Wrap` adds context to an error while preserving the type of the original error. This method behaves differently for each error type. For root errors, the stack trace is reset to the current callers which ensures traces are correct when using global/sentinel error values. Wrapped error types are simply wrapped with the new context. For external types (i.e. something other than root or wrap errors), a new root error is created for the original error and then it's wrapped with the additional context.

```golang
_, err := db.Get(id)
if err != nil {
  return eris.Wrapf(err, "error getting resource '%v'", id)
}
```

## Inspecting error types

The `eris` package provides a few ways to inspect and compare error types. `eris.Is` returns true if a particular error appears anywhere in the error chain, and `eris.Cause` returns the root cause of the error. Currently, `eris.Is` works simply by comparing error messages with each other. If an error contains a particular error message anywhere in its chain (e.g. "not found"), it's defined to be that error type (i.e. `eris.Is` will return `true`).

```golang
NotFound := eris.New("not found")
_, err := db.Get(id)
if eris.Is(err, NotFound) || eris.Cause(err) == NotFound {
  return eris.Wrapf(err, "resource '%v' not found", id)
}
```

## Migrating to eris

Migrating to `eris` should be a very simple process. If it doesn't offer something that you currently use from existing error packages, feel free to submit an issue to us. If you don't want to refactor all of your error handling yet, `eris` should work relatively seamlessly with your existing error types. Please submit an issue if this isn't the case for some reason.

## Contributing

If you'd like to contribute to `eris`, we'd love your input! Please submit an issue first so we can discuss your proposal. We're also available to discuss potential issues and features on our [Slack channel](https://rotisserieworkspace.slack.com/archives/CS13EC3T6).
