package eris_test

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/rotisserie/eris"
)

var (
	ErrUnexpectedEOF          = eris.New("unexpected EOF")
	FormattedErrUnexpectedEOF = eris.Errorf("unexpected %v", "EOF")
)

// Demonstrates JSON formatting of wrapped errors that originate from external (non-eris) error
// types. You can try this example in the Go playground (https://play.golang.org/p/29yCByzK8wT).
func ExampleToJSON_external() {
	// example func that returns an IO error
	readFile := func(fname string) error {
		return io.ErrUnexpectedEOF
	}

	// unpack and print the error
	err := readFile("example.json")
	u, _ := json.Marshal(eris.ToJSON(err, false)) // false: omit stack trace
	fmt.Println(string(u))

	// example output:
	// {
	//   "external":"unexpected EOF"
	// }
}

// Hack to run examples that don't have a predictable output (i.e. all examples that involve
// printing stack traces).
func TestExampleToJSON_external(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleToJSON_external()
}

// Demonstrates JSON formatting of wrapped errors that originate from global root errors. You can
// try this example in the Go playground (https://play.golang.org/p/jkZHLfHsYHV).
func ExampleToJSON_global() {
	// example func that wraps a global error value
	readFile := func(fname string) error {
		return eris.Wrapf(ErrUnexpectedEOF, "error reading file '%v'", fname) // line 6
	}

	// example func that catches and returns an error without modification
	parseFile := func(fname string) error {
		// read the file
		err := readFile(fname) // line 12
		if err != nil {
			return err
		}
		return nil
	}

	// unpack and print the error via uerr.ToJSON(...)
	err := parseFile("example.json")                             // line 20
	u, _ := json.MarshalIndent(eris.ToJSON(err, true), "", "\t") // true: include stack trace
	fmt.Printf("%v\n", string(u))

	// example output:
	// {
	//   "root": {
	//     "message": "unexpected EOF",
	//     "stack": [
	//       "main.main:.../example/main.go:20",
	//       "main.parseFile:.../example/main.go:12",
	//       "main.readFile:.../example/main.go:6"
	//     ]
	//   },
	//   "wrap": [
	//     {
	//       "message": "error reading file 'example.json'",
	//       "stack": "main.readFile:.../example/main.go:6"
	//     }
	//   ]
	// }
}

func TestExampleToJSON_global(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleToJSON_global()
}

// Demonstrates JSON formatting of wrapped errors that originate from local root errors (created at
// the source of the error via eris.New). You can try this example in the Go playground
// (https://play.golang.org/p/66nsuoOgQWu).
func ExampleToJSON_local() {
	// example func that returns an eris error
	readFile := func(fname string) error {
		return eris.New("unexpected EOF") // line 3
	}

	// example func that catches an error and wraps it with additional context
	parseFile := func(fname string) error {
		// read the file
		err := readFile(fname) // line 9
		if err != nil {
			return eris.Wrapf(err, "error reading file '%v'", fname) // line 11
		}
		return nil
	}

	// example func that just catches and returns an error
	processFile := func(fname string) error {
		// parse the file
		err := parseFile(fname) // line 19
		if err != nil {
			return err
		}
		return nil
	}

	// another example func that catches and wraps an error
	printFile := func(fname string) error {
		// process the file
		err := processFile(fname) // line 29
		if err != nil {
			return eris.Wrapf(err, "error printing file '%v'", fname) // line 31
		}
		return nil
	}

	// unpack and print the raw error
	err := printFile("example.json") // line 37
	u, _ := json.MarshalIndent(eris.ToJSON(err, true), "", "\t")
	fmt.Printf("%v\n", string(u))

	// example output:
	// {
	//   "root": {
	//     "message": "unexpected EOF",
	//     "stack": [
	//       "main.main:.../example/main.go:37",
	//       "main.printFile:.../example/main.go:31",
	//       "main.printFile:.../example/main.go:29",
	//       "main.processFile:.../example/main.go:19",
	//       "main.parseFile:.../example/main.go:11",
	//       "main.parseFile:.../example/main.go:9",
	//       "main.readFile:.../example/main.go:3"
	//     ]
	//   },
	//   "wrap": [
	//     {
	//       "message": "error printing file 'example.json'",
	//       "stack": "main.printFile:.../example/main.go:31"
	//     },
	//     {
	//       "message": "error reading file 'example.json'",
	//       "stack": "main.parseFile: .../example/main.go: 11"
	//     }
	//   ]
	// }
}

func TestExampleToJSON_local(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleToJSON_local()
}

// Demonstrates string formatting of wrapped errors that originate from external (non-eris) error
// types. You can try this example in the Go playground (https://play.golang.org/p/OKbU3gzIZvZ).
func ExampleToString_external() {
	// example func that returns an IO error
	readFile := func(fname string) error {
		return io.ErrUnexpectedEOF
	}

	// unpack and print the error
	err := readFile("example.json")
	fmt.Println(eris.ToString(err, false)) // false: omit stack trace

	// example output:
	// unexpected EOF
}

func TestExampleToString_external(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleToString_external()
}

// Demonstrates string formatting of wrapped errors that originate from global root errors. You can
// try this example in the Go playground (https://play.golang.org/p/8YgyDwk9xBJ).
func ExampleToString_global() {
	// example func that wraps a global error value
	readFile := func(fname string) error {
		return eris.Wrapf(FormattedErrUnexpectedEOF, "error reading file '%v'", fname) // line 6
	}

	// example func that catches and returns an error without modification
	parseFile := func(fname string) error {
		// read the file
		err := readFile(fname) // line 12
		if err != nil {
			return err
		}
		return nil
	}

	// example func that just catches and returns an error
	processFile := func(fname string) error {
		// parse the file
		err := parseFile(fname) // line 22
		if err != nil {
			return eris.Wrapf(err, "error processing file '%v'", fname) // line 24
		}
		return nil
	}

	// call processFile and catch the error
	err := processFile("example.json") // line 30

	// print the error via fmt.Printf
	fmt.Printf("%v\n", err) // %v: omit stack trace

	// example output:
	// unexpected EOF: error reading file 'example.json'

	// unpack and print the error via uerr.ToString(...)
	fmt.Printf("%v\n", eris.ToString(err, true)) // true: include stack trace

	// example output:
	// error reading file 'example.json'
	//   main.readFile:.../example/main.go:6
	// unexpected EOF
	//   main.main:.../example/main.go:30
	//   main.processFile:.../example/main.go:24
	//   main.processFile:.../example/main.go:22
	//   main.parseFile:.../example/main.go:12
	//   main.readFile:.../example/main.go:6
}

func TestExampleToString_global(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleToString_global()
}

// Demonstrates string formatting of wrapped errors that originate from local root errors (created
// at the source of the error via eris.New).  You can try this example in the Go playground
// (https://play.golang.org/p/d49gTNx3OtA).
func ExampleToString_local() {
	// example func that returns an eris error
	readFile := func(fname string) error {
		return eris.New("unexpected EOF") // line 3
	}

	// example func that catches an error and wraps it with additional context
	parseFile := func(fname string) error {
		// read the file
		err := readFile(fname) // line 9
		if err != nil {
			return eris.Wrapf(err, "error reading file '%v'", fname) // line 11
		}
		return nil
	}

	// call parseFile and catch the error
	err := parseFile("example.json") // line 17

	// print the error via fmt.Printf
	fmt.Printf("%v\n", err) // %v: omit stack trace

	// example output:
	// unexpected EOF: error reading file 'example.json'

	// unpack and print the error via uerr.ToString(...)
	fmt.Println(eris.ToString(err, true)) // true: include stack trace

	// example output:
	// error reading file 'example.json'
	//   main.parseFile:.../example/main.go:11
	// unexpected EOF
	//   main.main:.../example/main.go:17
	//   main.parseFile:.../example/main.go:11
	//   main.parseFile:.../example/main.go:9
	//   main.readFile:.../example/main.go:3
}

func TestExampleToString_local(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleToString_local()
}
