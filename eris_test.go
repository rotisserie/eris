package eris_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/morningvera/eris"
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
		"no error wrapping with internal root cause (eris.Errorf)": {
			cause:  eris.Errorf("%v", "root error"),
			output: "root error",
		},
		"no error wrapping with external root cause (errors.New)": {
			cause:  errors.New("external error"),
			output: "external error",
		},
	}

	for desc, tc := range tests {
		err := setupTestCase(false, tc.cause, tc.input)
		if err != nil && tc.cause == nil {
			t.Errorf("%v: wrapping nil errors should return nil but got { %v }", desc, err)
		} else if err != nil && tc.output != err.Error() {
			t.Errorf("%v: expected { %v } got { %v }", desc, tc.output, err.Error())
		}

		// Default printing
		defaultPrinter := eris.NewDefaultPrinter(eris.NewDefaultFormat(true))
		fmt.Printf("\nDefault error output (%v):\n%v", desc, defaultPrinter.Sprint(err))

		// JSON printing
		jsonPrinter := eris.NewJSONPrinter(eris.NewDefaultFormat(true))
		fmt.Printf("\nJSON error output (%v):\n%v\n", desc, jsonPrinter.Sprint(err))
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
	}

	for desc, tc := range tests {
		err := setupTestCase(true, tc.cause, tc.input)
		for _, out := range tc.output {
			if err == nil {
				t.Errorf("%v: unwrapping error returned nil but expected { %v }", desc, out)
			} else if out != err.Error() {
				t.Errorf("%v: expected { %v } got { %v }", desc, out, err.Error())
			}
			err = eris.Unwrap(err)
		}
	}
}

// todo: err.Is()
func TestErrorIs(t *testing.T) {}

// todo: err.Cause()
func TestErrorCause(t *testing.T) {}

// todo: err.Error(), fmt.Sprint(err), fmt.Sprintf("%v", err), fmt.Sprintf("%+v", err), etc.
func TestErrorFormatting(t *testing.T) {}
