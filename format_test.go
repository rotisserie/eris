package eris_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/rotisserie/eris"
)

func errChainsEqual(a []eris.ErrLink, b []eris.ErrLink) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Msg != b[i].Msg {
			return false
		}
	}

	return true
}

func TestUnpack(t *testing.T) {
	tests := map[string]struct {
		cause  error
		input  []string
		output eris.UnpackedError
	}{
		"nil error": {
			cause:  nil,
			input:  nil,
			output: eris.UnpackedError{},
		},
		"nil root error": {
			cause:  nil,
			input:  []string{"additional context"},
			output: eris.UnpackedError{},
		},
		"standard error wrapping with internal root cause (eris.New)": {
			cause: eris.New("root error"),
			input: []string{"additional context", "even more context"},
			output: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
				},
				ErrChain: []eris.ErrLink{
					{
						Msg: "additional context",
					},
					{
						Msg: "even more context",
					},
				},
			},
		},
		"standard error wrapping with external root cause (errors.New)": {
			cause: errors.New("external error"),
			input: []string{"additional context", "even more context"},
			output: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "external error",
				},
				ErrChain: []eris.ErrLink{
					{
						Msg: "additional context",
					},
					{
						Msg: "even more context",
					},
				},
			},
		},
		"no error wrapping with internal root cause (eris.Errorf)": {
			cause: eris.Errorf("%v", "root error"),
			output: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
				},
			},
		},
		"no error wrapping with external root cause (errors.New)": {
			cause: errors.New("external error"),
			output: eris.UnpackedError{
				ExternalErr: "external error",
			},
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			err := setupTestCase(false, tt.cause, tt.input)
			if got := eris.Unpack(err); got.ErrChain != nil && tt.output.ErrChain != nil && !errChainsEqual(got.ErrChain, tt.output.ErrChain) {
				t.Errorf("Unpack() ErrorChain = %v, want %v", got.ErrChain, tt.output.ErrChain)
			}
			if got := eris.Unpack(err); !reflect.DeepEqual(got.ErrRoot.Msg, tt.output.ErrRoot.Msg) {
				t.Errorf("Unpack() ErrorRoot = %v, want %v", got.ErrRoot.Msg, tt.output.ErrRoot.Msg)
			}
		})
	}
}

func TestFormatStr(t *testing.T) {
	tests := map[string]struct {
		input  error
		output string
	}{
		"basic root error": {
			input:  eris.New("root error"),
			output: "root error",
		},
		"basic wrapped error": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error"), "additional context"), "even more context"),
			output: "root error: additional context: even more context",
		},
	}
	for desc, tt := range tests {
		// without trace
		t.Run(desc, func(t *testing.T) {
			if got := eris.ToString(tt.input, false); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToString() got\n'%v'\nwant\n'%v'", got, tt.output)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := map[string]struct {
		input  error
		output string
	}{
		"basic root error": {
			input:  eris.New("root error"),
			output: `{"root":{"message":"root error"}}`,
		},
		"basic wrapped error": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error"), "additional context"), "even more context"),
			output: `{"root":{"message":"root error"},"wrap":[{"message":"additional context"},{"message":"even more context"}]}`,
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			result, _ := json.Marshal(eris.ToJSON(tt.input, false))
			if got := string(result); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToJSON() = %v, want %v", got, tt.output)
			}
		})
	}
}
