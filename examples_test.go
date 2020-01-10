package eris_test

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/rotisserie/eris"
)

// Demonstrates JSON formatting of wrapped errors that originate from
// external (non-eris) error types.
func ExampleUnpackedError_ToJSON_external() {
	// example func that returns an IO error
	readFile := func(fname string) error {
		return io.ErrUnexpectedEOF
	}

	// unpack and print the error
	err := readFile("example.json")
	uerr := eris.Unpack(err)
	format := eris.NewDefaultFormat(false) // false: omit stack trace
	u, _ := json.Marshal(uerr.ToJSON(format))
	fmt.Println(string(u))
	// Output:
	// {"external":"unexpected EOF"}
}

// Demonstrates JSON formatting of wrapped errors that originate from
// global root errors (created via eris.NewGlobal).
func ExampleUnpackedError_ToJSON_global() {
	// declare a "global" error type
	ErrUnexpectedEOF := eris.NewGlobal("unexpected EOF")

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
	err := parseFile("example.json") // line 20
	uerr := eris.Unpack(err)
	format := eris.NewDefaultFormat(true) // true: include stack trace
	u, _ := json.MarshalIndent(uerr.ToJSON(format), "", "\t")
	fmt.Printf("%v\n", string(u))

	// Output:
	// {
	// 	"root": {
	// 		"message": "unexpected EOF",
	// 		"stack": [
	// 			"main.readFile: .../example/main.go: 6",
	// 			"main.parseFile: .../example/main.go: 12",
	// 			"main.main: .../example/main.go: 20",
	// 		]
	// 	},
	// 	"wrap": [
	// 		{
	// 			"message": "error reading file 'example.json'",
	// 			"stack": "main.readFile: .../example/main.go: 6"
	// 		}
	// 	]
	// }
}

// Hack to run examples that don't have a predictable output (i.e. all
// examples that involve printing stack traces).
func TestExampleUnpackedError_ToJSON_global(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleUnpackedError_ToJSON_global()
}

// Demonstrates JSON formatting of wrapped errors that originate from
// local root errors (created at the source of the error via eris.New).
func ExampleUnpackedError_ToJSON_local() {
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
	uerr := eris.Unpack(err)
	format := eris.NewDefaultFormat(true) // true: include stack trace
	u, _ := json.MarshalIndent(uerr.ToJSON(format), "", "\t")
	fmt.Printf("%v\n", string(u))

	// Output:
	// 	{
	// 	"root": {
	// 		"message": "unexpected EOF",
	// 		"stack": [
	// 			"main.readFile: .../example/main.go: 3",
	// 			"main.parseFile: .../example/main.go: 9",
	// 			"main.parseFile: .../example/main.go: 11",
	// 			"main.processFile: .../example/main.go: 19",
	// 			"main.printFile: .../example/main.go: 29",
	// 			"main.printFile: .../example/main.go: 31",
	// 			"main.main: .../example/main.go: 37",
	// 		]
	// 	},
	// 	"wrap": [
	// 		{
	// 			"message": "error reading file 'example.json'",
	// 			"stack": "main.parseFile: .../example/main.go: 11"
	// 		},
	// 		{
	// 			"message": "error printing file 'example.json'",
	// 			"stack": "main.printFile: .../example/main.go: 31"
	// 		}
	// 	]
	// }
}

func TestExampleUnpackedError_ToJSON_local(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleUnpackedError_ToJSON_local()
}

// Demonstrates string formatting of wrapped errors that originate from
// external (non-eris) error types.
func ExampleUnpackedError_ToString_external() {
	// example func that returns an IO error
	readFile := func(fname string) error {
		return io.ErrUnexpectedEOF
	}

	// unpack and print the error
	err := readFile("example.json")
	uerr := eris.Unpack(err)
	format := eris.NewDefaultFormat(false) // false: omit stack trace
	fmt.Println(uerr.ToString(format))
	// Output:
	// unexpected EOF
}

// Demonstrates string formatting of wrapped errors that originate from
// global root errors (created via eris.NewGlobal).
func ExampleUnpackedError_ToString_global() {
	// declare a "global" error type
	ErrUnexpectedEOF := eris.NewGlobal("unexpected EOF")

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

	// call parseFile and catch the error
	err := parseFile("example.json") // line 20

	// print the error via fmt.Printf
	fmt.Printf("%v\n", err) // %v: omit stack trace

	// Output:
	// unexpected EOF: error reading file 'example.json'

	// unpack and print the error via uerr.ToString(...)
	uerr := eris.Unpack(err)
	format := eris.NewDefaultFormat(true) // true: include stack trace
	fmt.Printf("%v\n", uerr.ToString(format))

	// Output:
	// unexpected EOF
	// 	main.readFile: .../example/main.go: 6
	// 	main.parseFile: .../example/main.go: 12
	// 	main.main: .../example/main.go: 20
	// error reading file 'example.json'
	// 	main.readFile: .../example/main.go: 6
}

func TestExampleUnpackedError_ToString_global(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleUnpackedError_ToString_global()
}

// Demonstrates string formatting of wrapped errors that originate from
// local root errors (created at the source of the error via eris.New).
func ExampleUnpackedError_ToString_local() {
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

	// Output:
	// unexpected EOF: error reading file 'example.json'

	// unpack and print the error via uerr.ToString(...)
	uerr := eris.Unpack(err)
	format := eris.NewDefaultFormat(true) // true: include stack trace
	fmt.Println(uerr.ToString(format))

	// Output:
	// unexpected EOF
	// 	main.readFile: .../example/main.go: 3
	// 	main.parseFile: .../example/main.go: 9
	// 	main.parseFile: .../example/main.go: 11
	// 	main.main: .../example/main.go: 17
	// error reading file 'example.json'
	// 	main.parseFile: .../example/main.go: 11
}

func TestExampleUnpackedError_ToString_local(t *testing.T) {
	if !testing.Verbose() {
		return
	}
	ExampleUnpackedError_ToString_local()
}
