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

type withMessage struct {
	msg string
}

func (e withMessage) Error() string { return e.msg }
func (e withMessage) Is(target error) bool {
	if err, ok := target.(withMessage); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

type withLayer struct {
	err error
	msg string
}

func (e withLayer) Error() string { return e.msg + ": " + e.err.Error() }
func (e withLayer) Unwrap() error { return e.err }
func (e withLayer) Is(target error) bool {
	if err, ok := target.(withLayer); ok {
		return e.msg == err.msg
	}
	return e.msg == target.Error()
}

type withEmptyLayer struct {
	err error
}

func (e withEmptyLayer) Error() string { return e.err.Error() }
func (e withEmptyLayer) Unwrap() error { return e.err }

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
			output: "even more context: additional context: global error",
		},
		"formatted error wrapping with a global root cause": {
			cause:  formattedGlobalErr,
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: formatted global error",
		},
		"standard error wrapping with a local root cause": {
			cause:  eris.New("root error"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: root error",
		},
		"standard error wrapping with a local root cause (eris.Errorf)": {
			cause:  eris.Errorf("%v root error", "formatted"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: formatted root error",
		},
		"no error wrapping with a local root cause (eris.Errorf)": {
			cause:  eris.Errorf("%v root error", "formatted"),
			output: "formatted root error",
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

func TestExternalErrorWrapping(t *testing.T) {
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input for error wrapping
		output []string // expected output
	}{
		"no error wrapping with a third-party root cause (errors.New)": {
			cause: errors.New("external error"),
			output: []string{
				"external error",
			},
		},
		"standard error wrapping with a third-party root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause (errors.New and fmt.Errorf)": {
			cause: fmt.Errorf("additional context: %w", errors.New("external error")),
			input: []string{"even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause (multiple layers)": {
			cause: fmt.Errorf("even more context: %w", fmt.Errorf("additional context: %w", errors.New("external error"))),
			input: []string{"way too much context"},
			output: []string{
				"way too much context: even more context: additional context: external error",
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause that contains an empty layer": {
			cause: fmt.Errorf(": %w", errors.New("external error")),
			input: []string{"even more context"},
			output: []string{
				"even more context: : external error",
				": external error",
				"external error",
			},
		},
		"wrapping a wrapped third-party root cause that contains an empty layer without a delimiter": {
			cause: fmt.Errorf("%w", errors.New("external error")),
			input: []string{"even more context"},
			output: []string{
				"even more context: external error",
				"external error",
				"external error",
			},
		},
		"wrapping a pkg/errors style error (contains layers without messages)": {
			cause: &withLayer{ // var to mimic wrapping a pkg/errors style error
				msg: "additional context",
				err: &withEmptyLayer{
					err: &withMessage{
						msg: "external error",
					},
				},
			},
			input: []string{"even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
				"external error",
			},
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)

			// unwrap to make sure external errors are actually wrapped properly
			var inputErr []string
			for err != nil {
				inputErr = append(inputErr, err.Error())
				err = eris.Unwrap(err)
			}

			// compare each layer of the actual and expected output
			if len(inputErr) != len(tc.output) {
				t.Fatalf("%v: expected output to have '%v' layers but got '%v': { %#v } got { %#v }", desc, len(tc.output), len(inputErr), tc.output, inputErr)
			}
			for i := 0; i < len(inputErr); i++ {
				if inputErr[i] != tc.output[i] {
					t.Errorf("%v: expected { %#v } got { %#v }", desc, inputErr[i], tc.output[i])
				}
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
				"even more context: additional context: root error",
				"additional context: root error",
				"root error",
			},
		},
		"unwrapping error with external root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
				"external error",
			},
		},
		"unwrapping error with external root cause (custom type)": {
			cause: &withMessage{
				msg: "external error",
			},
			input: []string{"additional context", "even more context"},
			output: []string{
				"even more context: additional context: external error",
				"additional context: external error",
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
	externalErr := errors.New("external error")
	customErr := withLayer{
		msg: "additional context",
		err: withEmptyLayer{
			err: withMessage{
				msg: "external error",
			},
		},
	}

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
			cause:   externalErr,
			input:   []string{"additional context", "even more context"},
			compare: externalErr,
			output:  true,
		},
		"wrapped error from global root error": {
			cause:   globalErr,
			input:   []string{"additional context", "even more context"},
			compare: eris.Wrap(globalErr, "additional context"),
			output:  true,
		},
		"comparing against external error": {
			cause:   externalErr,
			input:   []string{"additional context", "even more context"},
			compare: externalErr,
			output:  true,
		},
		"comparing against custom error type": {
			cause:   customErr,
			input:   []string{"even more context"},
			compare: customErr,
			output:  true,
		},
		"comparing against custom error type (copied error)": {
			cause: customErr,
			input: []string{"even more context"},
			compare: &withMessage{
				msg: "external error",
			},
			output: true,
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
	extErr := errors.New("external error")
	customErr := withMessage{
		msg: "external error",
	}

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
		"external error": {
			cause:  extErr,
			input:  []string{"additional context", "even more context"},
			output: extErr,
		},
		"external error (custom type)": {
			cause:  customErr,
			input:  []string{"additional context", "even more context"},
			output: customErr,
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

func TestErrorAs(t *testing.T) {
	cause := withMessage{
		msg: "external error",
	}
	empty := withEmptyLayer{
		err: cause,
	}
	layer := withLayer{
		msg: "additional context",
		err: empty,
	}

	tests := map[string]struct {
		cause   error    // root error
		input   []string // input for error wrapping
		results []bool   // comparison results
		targets []error  // output results
	}{
		"external cause": {
			cause:   cause,
			input:   []string{"even more context"},
			results: []bool{true, false, false},
			targets: []error{cause, nil, nil},
		},
		"external error with empty layer": {
			cause:   empty,
			input:   []string{"even more context"},
			results: []bool{true, true, false},
			targets: []error{cause, empty, nil},
		},
		"external error with multiple layers": {
			cause:   layer,
			input:   []string{"even more context"},
			results: []bool{true, true, true},
			targets: []error{cause, empty, layer},
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)

			msgTarget := withMessage{}
			msgResult := errors.As(err, &msgTarget)
			if tc.results[0] != msgResult {
				t.Errorf("%v: expected errors.As('%v', &withMessage{}) to return {'%v', '%v'} but got {'%v', '%v'}",
					desc, err, tc.results[0], tc.targets[0], msgResult, msgTarget)
			} else if msgResult == true && tc.targets[0] != msgTarget {
				t.Errorf("%v: expected errors.As('%v', &withMessage{}) to return {'%v', '%v'} but got {'%v', '%v'}",
					desc, err, tc.results[0], tc.targets[0], msgResult, msgTarget)
			}

			emptyTarget := withEmptyLayer{}
			emptyResult := errors.As(err, &emptyTarget)
			if tc.results[1] != emptyResult {
				t.Errorf("%v: expected errors.As('%v', &withEmptyLayer{}) to return {'%v', '%v'} but got {'%v', '%v'}",
					desc, err, tc.results[1], tc.targets[1], emptyResult, emptyTarget)
			} else if emptyResult == true && tc.targets[1] != emptyTarget {
				t.Errorf("%v: expected errors.As('%v', &withEmptyLayer{}) to return {'%v', '%v'} but got {'%v', '%v'}",
					desc, err, tc.results[1], tc.targets[1], emptyResult, emptyTarget)
			}

			layerTarget := withLayer{}
			layerResult := errors.As(err, &layerTarget)
			if tc.results[2] != layerResult {
				t.Errorf("%v: expected errors.As('%v', &withLayer{}) to return {'%v', '%v'} but got {'%v', '%v'}",
					desc, err, tc.results[2], tc.targets[2], layerResult, layerTarget)
			} else if layerResult == true && tc.targets[2] != layerTarget {
				t.Errorf("%v: expected errors.As('%v', &withLayer{}) to return {'%v', '%v'} but got {'%v', '%v'}",
					desc, err, tc.results[2], tc.targets[2], layerResult, layerTarget)
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
			output: "even more context: additional context: root error",
		},
		"standard error wrapping with external root cause (errors.New)": {
			cause:  errors.New("external error"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: external error",
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

			_ = fmt.Sprintf("error formatting results (%v):\n", desc)
			_ = fmt.Sprintf("%v\n", err)
			_ = fmt.Sprintf("%+v", err)
		})
	}
}

func getFrames(pc []uintptr) []eris.StackFrame {
	var stackFrames []eris.StackFrame
	if len(pc) == 0 {
		return stackFrames
	}

	frames := runtime.CallersFrames(pc)
	for {
		frame, more := frames.Next()
		i := strings.LastIndex(frame.Function, "/")
		name := frame.Function[i+1:]
		stackFrames = append(stackFrames, eris.StackFrame{
			Name: name,
			File: frame.File,
			Line: frame.Line,
		})
		if !more {
			break
		}
	}

	return stackFrames
}

func getFrameFromLink(link eris.ErrLink) eris.Stack {
	var stackFrames []eris.StackFrame
	stackFrames = append(stackFrames, link.Frame)
	return eris.Stack(stackFrames)
}

func TestStackFrames(t *testing.T) {
	tests := map[string]struct {
		cause     error    // root error
		input     []string // input for error wrapping
		isWrapErr bool     // flag for wrap error
	}{
		"root error": {
			cause:     eris.New("root error"),
			isWrapErr: false,
		},
		"wrapped error": {
			cause:     eris.New("root error"),
			input:     []string{"additional context", "even more context"},
			isWrapErr: true,
		},
		"external error": {
			cause:     errors.New("external error"),
			isWrapErr: false,
		},
		"wrapped external error": {
			cause:     errors.New("external error"),
			input:     []string{"additional context", "even more context"},
			isWrapErr: true,
		},
		"global root error": {
			cause:     globalErr,
			isWrapErr: false,
		},
		"wrapped error from global root error": {
			cause:     globalErr,
			input:     []string{"additional context", "even more context"},
			isWrapErr: true,
		},
		"nil error": {
			cause:     nil,
			isWrapErr: false,
		},
	}

	for desc, tc := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tc.cause, tc.input)
			uErr := eris.Unpack(err)
			sFrames := eris.Stack(getFrames(eris.StackFrames(err)))
			if !tc.isWrapErr && !reflect.DeepEqual(uErr.ErrRoot.Stack, sFrames) {
				t.Errorf("%v: expected { %v } got { %v }", desc, uErr.ErrRoot.Stack, sFrames)
			}
			if tc.isWrapErr && !reflect.DeepEqual(getFrameFromLink(uErr.ErrChain[0]), sFrames) {
				t.Errorf("%v: expected { %v } got { %v }", desc, getFrameFromLink(uErr.ErrChain[0]), sFrames)
			}
		})
	}
}
