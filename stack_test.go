package eris

import (
	// "errors"
	"fmt"
	"testing"
	// "runtime"
)

var (
	BadRequest = New("bad request")
)

// todo: improve test before submitting PR
func TestErrorStack(t *testing.T) {
	// Testing stack with New
	err := ErrorUsingNew()
	fmt.Println("error output (new):\n", err)

	// Testing stack with globals
	err = ErrorUsingGlobal()
	fmt.Println("error output (global):\n", err)
	fmt.Println("error output after wrapping:\n", BadRequest)
}

func RootErrorNew() error {
	return New("test error")
}

func SecondErrorNew() error {
	return RootErrorNew()
}

func ErrorUsingNew() error {
	return SecondErrorNew()
}

// todo: test cases
//       global var (root error)
//       using WithStack multiple times
//       wrapped error type (how should this look?)
//       root error (resetting stack on existing root error)

func RootErrorGlobal() error {
	return Wrap(BadRequest, "testing")
}

func SecondErrorGlobal() error {
	return RootErrorGlobal()
}

func ErrorUsingGlobal() error {
	return SecondErrorGlobal()
}

// this stuff is just demonstrating the runtime library capabilities
// remove when no longer needed

// func TestStack(t *testing.T) {
// 	pc, file, line, ok := runtime.Caller(1)
// 	fmt.Println("")
// 	fmt.Println(pc)
// 	fmt.Println(file)
// 	fmt.Println(line)
// 	fmt.Println(ok)

// 	const depth = 32
// 	var pcs [depth]uintptr
// 	n := runtime.Callers(1, pcs[:])

// 	fmt.Println("")
// 	fmt.Println(n)
// 	fmt.Println(pcs[0:n])
// 	fmt.Println(pcs)

// 	for _, pc := range pcs {
// 		fn := runtime.FuncForPC(pc)
// 		if fn == nil {
// 			continue
// 		}
// 		file, line := fn.FileLine(pc)
// 		name := fn.Name()

// 		fmt.Println("")
// 		fmt.Println(file)
// 		fmt.Println(line)
// 		fmt.Println(name)
// 	}

// 	err := TopMethod()
// 	fmt.Println(err)

// 	err = AltTopMethod()
// 	fmt.Println(err)
// }

// func DeepestMethod() error {
// 	const depth = 32 // todo: make this a sufficiently large number
// 	var pcs [depth]uintptr
// 	n := runtime.Callers(1, pcs[:])

// 	fmt.Println("")
// 	fmt.Println(n)
// 	fmt.Println(pcs[0:n])
// 	fmt.Println(pcs)

// 	for _, pc := range pcs {
// 		fn := runtime.FuncForPC(pc)
// 		if fn == nil {
// 			continue
// 		}
// 		file, line := fn.FileLine(pc)
// 		name := fn.Name()

// 		fmt.Println("")
// 		fmt.Println(file)
// 		fmt.Println(line)
// 		fmt.Println(name)
// 	}

// 	return New("test error")
// }

// func NextMethod() error {
// 	return Wrap(DeepestMethod(), "additional context")
// }

// func TopMethod() error {
// 	return Wrap(NextMethod(), "even more context")
// }

// func AltNextMethod() error {
// 	return Wrap(DeepestMethod(), "additional context")
// }

// func AltTopMethod() error {
// 	return Wrap(AltNextMethod(), "even more context")
// }

// func TestPkgErrors(t *testing.T) {

// }
