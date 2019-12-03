# eris 😈

Package eris provides a better way to handle errors in Go. This package is inspired by a few existing packages: [xerrors](https://github.com/golang/xerrors), [pkg/errors](https://github.com/pkg/errors), and [Go 1.13 errors](https://golang.org/pkg/errors/).

`go get github.com/morningvera/eris`

Basic error wrapping was added in Go 1.13, but it omitted things like `Wrap` methods and built-in stack tracing. Other error packages already provide these features but in a slightly inflexible way. This package is intended to make error wrapping and stack tracing easier while also giving users more control over the output format.

Check out the [package docs](https://godoc.org/github.com/morningvera/eris) for more detailed information.

## Error types

`eris` is concerned with only three different types of errors: root errors, wrap errors, and external errors. Root and wrap errors are defined types in this package and all other error types are external or third-party errors.

Root errors are created via `eris.New` and `eris.Errorf` and are defined as the root cause of an error. Generally, it's a good idea to maintain a set of root errors that are then wrapped with additional context whenever an error of that type occurs.

Wrap errors represent a stack of errors that have been wrapped with additional context. Unwrapping these errors via `eris.Unwrap` will return the next error in the stack until a root error is reached. `eris.Cause` will also retrieve the root error.

When external error types are wrapped with additional context, a root error is first created from the original error. This creates a stack trace for the error and allows it to function with the rest of the `eris` package.

## Wrapping errors with additional context

`eris.Wrap` adds context to an error while preserving the type of the original error. This method behaves differently for each error type. For root errors, the stack trace is reset to the current callers which ensures traces are correct when using global/sentinel error values. Wrapped error types are simply wrapped with the new context. For external types (i.e. something other than root or wrap errors), a new root error is created for the original error and then it's wrapped with the additional context.

```golang
_, err := db.Get(id)
if err != nil {
  return eris.Wrapf(err, "error getting resource '%v'", id)
}
```

## Checking the type of an error

The `eris` package provides a couple ways to inspect and compare error types. `eris.Is` returns true if a particular error appears anywhere in the error chain and `eris.Cause` returns the root cause of the error.

```golang
_, err := db.Get(id)
if eris.Is(err, NotFound) {
  return eris.Wrapf(err, "resource '%v' not found", id)
}
```

```golang
_, err := db.Get(id)
if eris.Cause(err) == NotFound {
  return eris.Wrapf(err, "resource '%v' not found", id)
}
```

## Contributing

If you'd like to contribute to `eris`, we'd love your input! Please submit an issue first so we can discuss your proposal.
