// Package eris provides a better way to handle, trace, and log errors in Go.
//
// Named after the Greek goddess of strife and discord, this package is designed to give you more
// control over error handling via error wrapping, stack tracing, and output formatting. eris was
// inspired by a simple question: what if you could fix a bug without wasting time replicating the
// issue or digging through the code?
//
// Many of the methods in this package will look familiar if you've used pkg/errors or xerrors, but
// eris employs some additional tricks during error wrapping and unwrapping that greatly improve the
// readability of the stack which should make debugging easier. This package also takes a unique
// approach to formatting errors that allows you to write custom formats that conform to your error
// or log aggregator of choice.
//
// Creating errors
//
// Creating errors is simple via eris.New.
//
//   var (
//     // global error values can be useful when wrapping errors or inspecting error types
//     ErrInternalServer = eris.New("error internal server")
//   )
//
//   func (req *Request) Validate() error {
//     if req.ID == "" {
//       // or return a new error at the source if you prefer
//       return eris.New("error bad request")
//     }
//     return nil
//   }
//
// Wrapping errors
//
// eris.Wrap adds context to an error while preserving the original error.
//
//   relPath, err := GetRelPath("/Users/roti/", resource.AbsPath)
//   if err != nil {
//     // wrap the error if you want to add more context
//     return nil, eris.Wrapf(err, "failed to get relative path for resource '%v'", resource.ID)
//   }
//
// Formatting and logging errors
//
// eris.ToString and eris.ToJSON should be used to log errors with the default format. The JSON
// method returns a map[string]interface{} type for compatibility with Go's encoding/json package
// and many common JSON loggers (e.g. logrus).
//
//   // format the error to JSON with the default format and stack traces enabled
//   formattedJSON := eris.ToJSON(err, true)
//   fmt.Println(json.Marshal(formattedJSON)) // marshal to JSON and print
//   logger.WithField("error", formattedJSON).Error() // or ideally, pass it directly to a logger
//
//   // format the error to a string and print it
//   formattedStr := eris.ToString(err, true)
//   fmt.Println(formattedStr)
//
// eris also enables control over the default format's separators and allows advanced users to write
// their own custom formats.
//
// Interpreting eris stack traces
//
// Errors created with this package contain stack traces that are managed automatically. They're
// currently mandatory when creating and wrapping errors but optional when printing or logging. The
// stack trace and all wrapped layers follow the same order as Go's `runtime` package, which means
// that the root cause of the error is shown first.
//
//   {
//     "root":{
//       "message":"error bad request", // root cause
//       "stack":[
//         "main.main:.../example.go:143", // original calling method
//         "main.ProcessResource:.../example.go:71",
//         "main.(*Request).Validate:.../example.go:29", // location of Wrap call
//         "main.(*Request).Validate:.../example.go:28" // location of the root
//       ]
//     },
//     "wrap":[
//       {
//         "message":"received a request with no ID", // additional context
//         "stack":"main.(*Request).Validate:.../example.go:29" // location of Wrap call
//       }
//     ]
//   }
//
// Inspecting errors
//
// The eris package provides a couple ways to inspect and compare error types. eris.Is returns true
// if a particular error appears anywhere in the error chain. Currently, it works simply by
// comparing error messages with each other. If an error contains a particular message (e.g. "error
// not found") anywhere in its chain, it's defined to be that error type.
//
//   ErrNotFound := eris.NewGlobal("error not found")
//   _, err := db.Get(id)
//   // check if the resource was not found
//   if eris.Is(err, ErrNotFound) {
//     // return the error with some useful context
//     return eris.Wrapf(err, "error getting resource '%v'", id)
//   }
//
// eris.Cause unwraps an error until it reaches the cause, which is defined as the first (i.e. root)
// error in the chain.
//
//   ErrNotFound := eris.NewGlobal("error not found")
//   _, err := db.Get(id)
//   // compare the cause to some sentinel value
//   if eris.Cause(err) == ErrNotFound {
//     // return the error with some useful context
//     return eris.Wrapf(err, "error getting resource '%v'", id)
//   }
//
package eris
