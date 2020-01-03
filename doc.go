// Package eris provides a better way to handle, trace, and log errors in Go.
//
// Types of errors
//
// This package is concerned with only three different types of errors: root
// errors, wrap errors, and external errors. Root and wrap errors are defined
// types in this package and all other error types are external or third-party
// errors.
//
// Root errors are created via eris.New and eris.Errorf. Generally, it's a
// good idea to maintain a set of root errors that are then wrapped with
// additional context whenever an error of that type occurs. Wrap errors
// represent a stack of errors that have been wrapped with additional context.
// Unwrapping these errors via eris.Unwrap will return the next error in the
// stack until a root error is reached. eris.Cause will also retrieve the root
// error.
//
// When external error types are wrapped with additional context, a root error
// is first created from the original error. This creates a stack trace for the
// error and allows it to function with the rest of the `eris` package.
//
// Wrapping errors with additional context
//
// eris.Wrap adds context to an error while preserving the type of the
// original error. This method behaves differently for each error type. For
// root errors, the stack trace is reset to the current callers which ensures
// traces are correct when using global/sentinel error values. Wrapped error
// types are simply wrapped with the new context. For external types (i.e.
// something other than root or wrap errors), a new root error is created for
// the original error and then it's wrapped with the additional context.
//
//    _, err := db.Get(id)
//    if err != nil {
//      // return the error with some useful context
//      return eris.Wrapf(err, "error getting resource '%v'", id)
//    }
//
// Inspecting error types
//
// The eris package provides a few ways to inspect and compare error types.
// eris.Is returns true if a particular error appears anywhere in the error
// chain, and eris.Cause returns the root cause of the error. Currently,
// eris.Is works simply by comparing error messages with each other. If an
// error contains a particular error message anywhere in its chain (e.g. "not
// found"), it's defined to be that error type (i.e. eris.Is will return
// true).
//
//    NotFound := eris.New("not found")
//    _, err := db.Get(id)
//    // check if the resource was not found
//    if eris.Is(err, NotFound) {
//      // return the error with some useful context
//      return eris.Wrapf(err, "error getting resource '%v'", id)
//    }
//
//    NotFound := eris.New("not found")
//    _, err := db.Get(id)
//    // compare the cause to some sentinel value
//    if eris.Cause(err) == NotFound {
//      // return the error with some useful context
//      return eris.Wrapf(err, "error getting resource '%v'", id)
//    }
//
// Stack traces
//
// Errors created with this package contain stack traces that are managed
// automatically even when wrapping global errors or errors from other
// libraries. Stack traces are currently mandatory when creating and wrapping
// errors but optional when printing or logging errors. Printing an error with
// or without the stack trace is simple:
//
//    _, err := db.Get(id)
//    if err != nil {
//      return eris.Wrapf(err, "error getting resource '%v'", id)
//    }
//    fmt.Printf("%v", err) // print without the stack trace
//    fmt.Printf("%+v", err) // print with the stack trace
//
// For an error that has been wrapped once, the output will look something
// like this:
//
//    # output without the stack trace
//    error getting resource 'example-id': not found
//
//    # output with the stack trace
//    error getting resource 'example-id'
//      api.GetResource: /path/to/file/api.go: 30
//    not found
//      api.GetResource: /path/to/file/api.go: 30
//      db.Get: /path/to/file/db.go: 99
//      runtime.goexit: /path/to/go/src/libexec/src/runtime/asm_amd64.s: 1337
//
// The first layer of the full error output shows a message ("error getting
// resource 'example-id'") and a single stack frame. The next layer shows the
// root error ("not found") and the full stack trace.
//
// Logging errors with more control
//
// While eris supports logging errors with Go's fmt package, it's often
// advantageous to use the provided string and JSON formatters instead. These
// methods provide much more control over the error output and should work
// seamlessly with whatever logging package you choose.
//
//    var fields log.Fields
//    unpackedErr := eris.Unpack(err)
//    fields["method"] = "api.GetResource"
//    fields["error"] = unpackedErr.ToJSON(eris.NewDefaultFormat(true))
//    logger.WithFields(fields).Errorf("method completed with error (%v)", err)
//
// When using a JSON logger, the output should look something like this:
//
//    {
//      "method":"api.GetResource",
//      "error":{
//        "error chain":[
//          {
//            "message":"error getting resource 'example-id'",
//            "stack":"api.GetResource: /path/to/file/api.go: 30"
//          }
//        ],
//        "error root":{
//          "message":"not found",
//          "stack":[
//            "api.GetResource: /path/to/file/api.go: 30",
//            "db.Get: /path/to/file/db.go: 99",
//            "runtime.goexit: /path/to/go/src/runtime/asm_amd64.s: 1337"
//          ]
//        }
//      }
//    }
//
package eris
