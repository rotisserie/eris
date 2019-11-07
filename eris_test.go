package eris_test

import (
	"fmt"
	"testing"

	"github.com/morningvera/eris"
	"github.com/stretchr/testify/assert"
)

// todo: need to fix error printing before this will work properly
func TestErrorWrapping(t *testing.T) {
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input messages for error wrapping
		output string   // expected output error string
	}{
		// "nil root error": {
		// 	cause: nil,
		// 	input: []string{"additional context"},
		// },
		"standard error wrapping": {
			cause:  eris.New("root error"),
			input:  []string{"additional context", "even more context"},
			output: "even more context: additional context: root error",
		},
		// "standard error wrapping using non-eris types": {
		// cause:
		// input:
		// output:
		// }
	}

	// todo: (maybe) create a generic func that takes in test cases and a closure
	for desc, tc := range tests {
		tc := tc
		t.Run(desc, func(t *testing.T) {
			t.Parallel()
			err := tc.cause
			for _, str := range tc.input {
				err = eris.Wrap(err, str)
			}
			fmt.Println("error:")
			fmt.Println(err)
			fmt.Println(err.Error())
			fmt.Println(fmt.Sprintf("%+v", err))
			if tc.cause == nil {
				assert.Nilf(t, err, "%v: wrapping nil errors should return nil but got { %v }", desc, err)
			} else {
				assert.Equalf(t, tc.output, err.Error(), "%v: expected { %v } got { %v }", desc, tc.output, err)
			}
		})
	}
}

// todo: test Is/Cause here
func TestErrorComparisons(t *testing.T) {
	tests := map[string]struct {
		cause  error    // root error
		input  []string // input messages for error wrapping
		output string   // expected output error string
	}{
		// "nil root error": {
		// 	cause: nil,
		// 	input: []string{"additional context"},
		// },
		// "standard error wrapping": {
		// 	cause:  eris.New("root error"),
		// 	input:  []string{"additional context", "even more context"},
		// 	output: "even more context: additional context: root error",
		// },
	}

	for desc, tc := range tests {
		tc := tc
		t.Run(desc, func(t *testing.T) {
			t.Parallel()
			err := tc.cause
			for _, str := range tc.input {
				err = eris.Wrap(err, str)
			}
			if tc.cause == nil {
				assert.Nilf(t, err, "%v: wrapping nil errors should return nil but got { %v }", desc, err)
			} else {
				assert.Equalf(t, tc.output, err.Error(), "%v: expected { %v } got { %v }", desc, tc.output, err)
			}
		})
	}
}
