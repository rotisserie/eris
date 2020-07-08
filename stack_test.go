package eris_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rotisserie/eris"
)

const (
	file              = "eris/stack_test.go"
	readFunc          = "eris_test.ReadFile"
	parseFunc         = "eris_test.ParseFile"
	processFunc       = "eris_test.ProcessFile"
	globalTestFunc    = "eris_test.TestGlobalStack"
	localTestFunc     = "eris_test.TestLocalStack"
	extGlobalTestFunc = "eris_test.TestExtGlobalStack"
	extLocalTestFunc  = "eris_test.TestExtLocalStack"
)

var (
	errEOF = eris.New("unexpected EOF")
	errExt = errors.New("external error")
)

// example func that either returns a wrapped global or creates/wraps a new local error
func ReadFile(fname string, global bool, external bool) error {
	var err error
	if !external && !global { // local eris
		err = eris.New("unexpected EOF")
	} else if !external && global { // global eris
		err = errEOF
	} else if external && !global { // local external
		err = fmt.Errorf("external context: %w", errors.New("external error"))
	} else { // global external
		err = fmt.Errorf("external context: %w", errExt)
	}
	return eris.Wrapf(err, "error reading file '%v'", fname)
}

// example func that just catches and returns an error
func ParseFile(fname string, global bool, external bool) error {
	err := ReadFile(fname, global, external)
	if err != nil {
		return err
	}
	return nil
}

// example func that wraps an error with additional context
func ProcessFile(fname string, global bool, external bool) error {
	// parse the file
	err := ParseFile(fname, global, external)
	if err != nil {
		return eris.Wrapf(err, "error processing file '%v'", fname)
	}
	return nil
}

func TestGlobalStack(t *testing.T) {
	// expected results
	expectedChain := []eris.StackFrame{
		{Name: readFunc, File: file, Line: 41},
		{Name: processFunc, File: file, Line: 58},
	}
	expectedRoot := []eris.StackFrame{
		{Name: readFunc, File: file, Line: 41},
		{Name: parseFunc, File: file, Line: 46},
		{Name: processFunc, File: file, Line: 56},
		{Name: processFunc, File: file, Line: 58},
		{Name: globalTestFunc, File: file, Line: 77},
	}

	err := ProcessFile("example.json", true, false)
	uerr := eris.Unpack(err)
	validateWrapFrames(t, expectedChain, uerr)
	validateRootStack(t, expectedRoot, uerr)
}

func TestLocalStack(t *testing.T) {
	// expected results
	expectedChain := []eris.StackFrame{
		{Name: readFunc, File: file, Line: 41},
		{Name: processFunc, File: file, Line: 58},
	}
	expectedRoot := []eris.StackFrame{
		{Name: readFunc, File: file, Line: 33},
		{Name: readFunc, File: file, Line: 41},
		{Name: parseFunc, File: file, Line: 46},
		{Name: processFunc, File: file, Line: 56},
		{Name: processFunc, File: file, Line: 58},
		{Name: localTestFunc, File: file, Line: 98},
	}

	err := ProcessFile("example.json", false, false)
	uerr := eris.Unpack(err)
	validateWrapFrames(t, expectedChain, uerr)
	validateRootStack(t, expectedRoot, uerr)
}

func TestExtGlobalStack(t *testing.T) {
	// expected results
	expectedChain := []eris.StackFrame{
		{Name: processFunc, File: file, Line: 58},
	}
	expectedRoot := []eris.StackFrame{
		{Name: readFunc, File: file, Line: 41},
		{Name: parseFunc, File: file, Line: 46},
		{Name: processFunc, File: file, Line: 56},
		{Name: processFunc, File: file, Line: 58},
		{Name: extGlobalTestFunc, File: file, Line: 117},
	}

	err := ProcessFile("example.json", true, true)
	uerr := eris.Unpack(err)
	validateWrapFrames(t, expectedChain, uerr)
	validateRootStack(t, expectedRoot, uerr)
}

func TestExtLocalStack(t *testing.T) {
	// expected results
	expectedChain := []eris.StackFrame{
		{Name: processFunc, File: file, Line: 58},
	}
	expectedRoot := []eris.StackFrame{
		{Name: readFunc, File: file, Line: 41},
		{Name: parseFunc, File: file, Line: 46},
		{Name: processFunc, File: file, Line: 56},
		{Name: processFunc, File: file, Line: 58},
		{Name: extLocalTestFunc, File: file, Line: 136},
	}

	err := ProcessFile("example.json", false, true)
	uerr := eris.Unpack(err)
	validateWrapFrames(t, expectedChain, uerr)
	validateRootStack(t, expectedRoot, uerr)
}

func validateWrapFrames(t *testing.T, expectedChain []eris.StackFrame, uerr eris.UnpackedError) {
	// verify the expected and actual error chain have the same length
	if len(expectedChain) != len(uerr.ErrChain) {
		t.Fatalf("%v: expected number of wrapped frames { %v } got { %v }", localTestFunc, len(expectedChain), len(uerr.ErrChain))
	}

	// verify the wrapped frames match expected values
	for i := 0; i < len(expectedChain); i++ {
		if expectedChain[i].Name != uerr.ErrChain[i].Frame.Name {
			t.Errorf("%v: expected wrap func name { %v } got { %v }", localTestFunc, expectedChain[i].Name, uerr.ErrChain[i].Frame.Name)
		}
		if !strings.Contains(uerr.ErrChain[i].Frame.File, expectedChain[i].File) {
			t.Errorf("%v: expected wrap file name to contain { %v } got { %v }", localTestFunc, uerr.ErrChain[i].Frame.File, expectedChain[i].File)
		}
		if expectedChain[i].Line != uerr.ErrChain[i].Frame.Line {
			t.Errorf("%v: expected wrap line number { %v } got { %v }", localTestFunc, expectedChain[i].Line, uerr.ErrChain[i].Frame.Line)
		}
	}
}

func validateRootStack(t *testing.T, expectedRoot []eris.StackFrame, uerr eris.UnpackedError) {
	// verify the expected and actual stack have the same length
	if len(expectedRoot) != len(uerr.ErrRoot.Stack) {
		t.Fatalf("%v: expected number of root error frames { %v } got { %v }", localTestFunc, len(expectedRoot), len(uerr.ErrRoot.Stack))
	}

	// verify the root error stack frames match expected values
	for i := 0; i < len(expectedRoot); i++ {
		if expectedRoot[i].Name != uerr.ErrRoot.Stack[i].Name {
			t.Errorf("%v: expected root func name { %v } got { %v }", localTestFunc, expectedRoot[i].Name, uerr.ErrRoot.Stack[i].Name)
		}
		if !strings.Contains(uerr.ErrRoot.Stack[i].File, expectedRoot[i].File) {
			t.Errorf("%v: expected root file name to contain { %v } got { %v }", localTestFunc, uerr.ErrRoot.Stack[i].File, expectedRoot[i].File)
		}
		if expectedRoot[i].Line != uerr.ErrRoot.Stack[i].Line {
			t.Errorf("%v: expected root line number { %v } got { %v }", localTestFunc, expectedRoot[i].Line, uerr.ErrRoot.Stack[i].Line)
		}
	}
}

func TestGoRoutines(t *testing.T) {
	expectedChain := []eris.StackFrame{
		{Name: "eris_test.TestGoRoutines.func1", File: file, Line: 192},
	}
	expectedRoot := []eris.StackFrame{
		{Name: "eris_test.dummyStack", File: file, Line: 204},
	}

	go func() {
		err := dummyStack()
		err = eris.Wrap(err, "error reading file")

		// verify the stack frames match expected values
		uerr := eris.Unpack(err)
		validateWrapFrames(t, expectedChain, uerr)
		validateRootStack(t, expectedRoot, uerr)
	}()

	time.Sleep(250 * time.Millisecond)
}

func dummyStack() error {
	return eris.New("unexpected EOF")
}
