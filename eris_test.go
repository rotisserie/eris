package eris_test

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/rotisserie/eris"
)

var (
	globalErr          = eris.New("global error")
	formattedGlobalErr = eris.Errorf("%v global error", "formatted")
)

func setupTestCase(wrapf bool, cause error, input []string) error {
	err := cause
	for _, str := range input {
		if wrapf {
			err = eris.Wrapf(err, "%v", str)
		} else {
			err = eris.Wrap(err, str)
		}
	}
	return err
}

func TestErrorWrapping(t *testing.T) {
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output string   // expected output
	}{
		"nil root error": {
			cause: nil,
			input: []string{"additional context"},
		},
		"standard error wrapping with a global root cause": {
			cause:  globalErr,
			input:  []string{"additional context", "even more context"},
			output: "global error: additional context: even more context",
		},
		"formatted error wrapping with a global root cause": {
			cause:  formattedGlobalErr,
			input:  []string{"additional context", "even more context"},
			output: "formatted global error: additional context: even more context",
		},
		"standard error wrapping with a local root cause": {
			cause:  eris.New("root error"),
			input:  []string{"additional context", "even more context"},
			output: "root error: additional context: even more context",
		},
		"standard error wrapping with a local root cause (eris.Errorf)": {
			cause:  eris.Errorf("%v root error", "formatted"),
			input:  []string{"additional context", "even more context"},
			output: "formatted root error: additional context: even more context",
		},
		"standard error wrapping with a third-party root cause (errors.New)": {
			cause:  errors.New("external error"),
			input:  []string{"additional context", "even more context"},
			output: "external error: additional context: even more context",
		},
		"no error wrapping with a local root cause (eris.Errorf)": { // todo: also test globals with Errorf (wrapping included)
			cause:  eris.Errorf("%v root error", "formatted"),
			output: "formatted root error",
		},
		"no error wrapping with a third-party root cause (errors.New)": {
			cause:  errors.New("external error"),
			output: "external error",
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			if err != nil && tc.cause == nil {
				t.Errorf("%v: wrapping nil errors should return nil but got { %v }", desc, err)
			} else if err != nil && tc.output != err.Error() {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, err)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output []string // expected output
	}{
		"unwrapping error with internal root cause (eris.New)": {
			cause: eris.New("root error"),
			input: []string{"additional context", "even more context"},
			output: []string{
				"root error: additional context: even more context",
				"root error: additional context",
				"root error",
			},
		},
		"unwrapping error with external root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			output: []string{
				"external error: additional context: even more context",
				"external error: additional context",
				"external error",
			},
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(true, tc.cause, tc.input)
			for _, out := range tc.output {
				if err == nil {
					t.Errorf("%v: unwrapping error returned nil but expected { %v }", desc, out)
				} else if out != err.Error() {
					t.Errorf("%v: expected { %v } got { %v }", desc, out, err)
				}
				err = eris.Unwrap(err)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	tests := map[string]struct {
		cause   error    // root error
		input   []string // input for error wrapping
		compare error    // errors for comparison
		output  bool     // expected comparison result
	}{
		"root error (internal)": {
			cause:   eris.New("root error"),
			input:   []string{"additional context", "even more context"},
			compare: eris.New("root error"),
			output:  true,
		},
		"error not in chain": {
			cause:   eris.New("root error"),
			compare: eris.New("other error"),
			output:  false,
		},
		"middle of chain (internal)": {
			cause:   eris.New("root error"),
			input:   []string{"additional context", "even more context"},
			compare: eris.New("additional context"),
			output:  true,
		},
		"another in middle of chain (internal)": {
			cause:   eris.New("root error"),
			input:   []string{"additional context", "even more context"},
			compare: eris.New("even more context"),
			output:  true,
		},
		"root error (external)": {
			cause:   errors.New("external error"),
			input:   []string{"additional context", "even more context"},
			compare: eris.New("external error"),
			output:  true,
		},
		"wrapped error from global root error": {
			cause:   globalErr,
			input:   []string{"additional context", "even more context"},
			compare: eris.Wrap(globalErr, "additional context"),
			output:  true,
		},
		"comparing against external error": {
			cause:   errors.New("external error"),
			input:   []string{"additional context", "even more context"},
			compare: errors.New("external error"),
			output:  true,
		},
		"comparing against nil error": {
			cause:   eris.New("root error"),
			compare: nil,
			output:  false,
		},
		"comparing error against itself": {
			cause:   globalErr,
			compare: globalErr,
			output:  true,
		},
		"comparing two nil errors": {
			cause:   nil,
			compare: nil,
			output:  true,
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			if tc.output && !eris.Is(err, tc.compare) {
				t.Errorf("%v: expected eris.Is('%v', '%v') to return true but got false", desc, err, tc.compare)
			} else if !tc.output && eris.Is(err, tc.compare) {
				t.Errorf("%v: expected eris.Is('%v', '%v') to return false but got true", desc, err, tc.compare)
			}
		})
	}
}

func TestErrorCause(t *testing.T) {
	globalErr := eris.New("global error")

	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output error    // expected output
	}{
		"internal root error": {
			cause:  globalErr,
			input:  []string{"additional context", "even more context"},
			output: globalErr,
		},
		"nil error": {
			cause:  nil,
			output: nil,
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			cause := eris.Cause(err)
			if tc.output != eris.Cause(err) {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, cause)
			}
		})
	}
}

func TestErrorFormatting(t *testing.T) {
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output string   // expected output
	}{
		"standard error wrapping with internal root cause (eris.New)": {
			cause:  eris.New("root error"),
			input:  []string{"additional context", "even more context"},
			output: "root error: additional context: even more context",
		},
		"standard error wrapping with external root cause (errors.New)": {
			cause:  errors.New("external error"),
			input:  []string{"additional context", "even more context"},
			output: "external error: additional context: even more context",
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			if err != nil && tc.cause == nil {
				t.Errorf("%v: wrapping nil errors should return nil but got { %v }", desc, err)
			} else if err != nil && tc.output != err.Error() {
				t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, err)
			}

			// todo: automate stack trace verification
			_ = fmt.Sprintf("error formatting results (%v):\n", desc)
			_ = fmt.Sprintf("%v\n", err)
			_ = fmt.Sprintf("%+v", err)
		})
	}
}

func getFrames(frames []uintptr) []eris.StackFrame {
	var sFrames []eris.StackFrame
	for _, u := range frames {
		pc := u - 1
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			frame := eris.StackFrame{
				Name: "unknown",
				File: "unknown",
			}
			sFrames = append(sFrames, frame)
		}

		name := fn.Name()
		i := strings.LastIndex(name, "/")
		name = name[i+1:]
		file, line := fn.FileLine(pc)

		frame := eris.StackFrame{
			Name: name,
			File: file,
			Line: line,
		}
		sFrames = append(sFrames, frame)
	}
	return sFrames
}

func TestStackFrames(t *testing.T) {
	tests := map[string]struct {
		cause error    // root error
		input []string // input for error wrapping
	}{
		"root error": {
			cause: eris.New("root error"),
		},
		"wrapped error": {
			cause: eris.New("root error"),
			input: []string{"additional context", "even more context"},
		},
		"external error": {
			cause: errors.New("external error"),
		},
		"wrapped external error": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
		},
		"global root error": {
			cause: globalErr,
		},
		"wrapped error from global root error": {
			cause: globalErr,
			input: []string{"additional context", "even more context"},
		},
		"nil error": {
			cause: nil,
		},
	}
	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			uErr := eris.Unpack(err)
			sFrames := eris.Stack(getFrames(eris.StackFrames(err)))
			if !reflect.DeepEqual(uErr.ErrRoot.Stack, sFrames) {
				t.Errorf("Stackframes() returned { %v }, was expecting { %v }", sFrames, uErr.ErrRoot.Stack)
			}
		})
	}
}
