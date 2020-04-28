package eris_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/rotisserie/eris"
)

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
				ErrExternal: errors.New("external error"),
				ErrRoot: eris.ErrRoot{
					Msg: "additional context",
				},
				ErrChain: []eris.ErrLink{
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
				ErrExternal: errors.New("external error"),
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
			output: "even more context: additional context: root error",
		},
		"external wrapped error": {
			input:  eris.Wrap(errors.New("external error"), "additional context"),
			output: "additional context: external error",
		},
		"external error": {
			input:  errors.New("external error"),
			output: "external error",
		},
		"empty error": {
			input:  eris.New(""),
			output: "",
		},
		"empty wrapped external error": {
			input:  eris.Wrap(errors.New(""), "additional context"),
			output: "additional context: ",
		},
		"empty wrapped error": {
			input:  eris.Wrap(eris.New(""), "additional context"),
			output: "additional context: ",
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

func TestInvertedFormatStr(t *testing.T) {
	tests := map[string]struct {
		input  error
		output string
	}{
		"basic wrapped error": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error"), "additional context"), "even more context"),
			output: "root error: additional context: even more context",
		},
		"external wrapped error": {
			input:  eris.Wrap(errors.New("external error"), "additional context"),
			output: "external error: additional context",
		},
		"external error": {
			input:  errors.New("external error"),
			output: "external error",
		},
		"empty wrapped external error": {
			input:  eris.Wrap(errors.New(""), "additional context"),
			output: ": additional context",
		},
		"empty wrapped error": {
			input:  eris.Wrap(eris.New(""), "additional context"),
			output: ": additional context",
		},
	}
	for desc, tt := range tests {
		// without trace
		t.Run(desc, func(t *testing.T) {
			format := eris.NewDefaultStringFormat(eris.FormatOptions{
				InvertOutput: true,
				WithExternal: true,
			})
			if got := eris.ToCustomString(tt.input, format); !reflect.DeepEqual(got, tt.output) {
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
			output: `{"root":{"message":"root error"},"wrap":[{"message":"even more context"},{"message":"additional context"}]}`,
		},
		"external error": {
			input:  eris.Wrap(errors.New("external error"), "additional context"),
			output: `{"external":"external error","root":{"message":"additional context"}}`,
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

func TestInvertedFormatJSON(t *testing.T) {
	tests := map[string]struct {
		input  error
		output string
	}{
		"basic wrapped error": {
			input:  eris.Wrap(eris.Wrap(eris.New("root error"), "additional context"), "even more context"),
			output: `{"root":{"message":"root error"},"wrap":[{"message":"additional context"},{"message":"even more context"}]}`,
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			format := eris.NewDefaultJSONFormat(eris.FormatOptions{
				InvertOutput: true,
			})
			result, _ := json.Marshal(eris.ToCustomJSON(tt.input, format))
			if got := string(result); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ToJSON() = %v, want %v", got, tt.output)
			}
		})
	}
}

func TestFormatJSONWithStack(t *testing.T) {
	tests := map[string]struct {
		input      error
		rootOutput map[string]interface{}
		wrapOutput []map[string]interface{}
	}{
		"basic wrapped error": {
			input: eris.Wrap(eris.Wrap(eris.New("root error"), "additional context"), "even more context"),
			rootOutput: map[string]interface{}{
				"message": "root error",
			},
			wrapOutput: []map[string]interface{}{
				{"message": "even more context"},
				{"message": "additional context"},
			},
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			format := eris.NewDefaultJSONFormat(eris.FormatOptions{
				WithTrace:   true,
				InvertTrace: true,
			})
			errJSON := eris.ToCustomJSON(tt.input, format)

			// make sure messages are correct and stack elements exist (actual stack validation is in stack_test.go)
			if rootMap, ok := errJSON["root"].(map[string]interface{}); ok {
				if _, exists := rootMap["message"]; !exists {
					t.Fatalf("%v: expected a 'message' field in the output but didn't find one { %v }", desc, errJSON)
				}
				if rootMap["message"] != tt.rootOutput["message"] {
					t.Errorf("%v: expected { %v } got { %v }", desc, rootMap["message"], tt.rootOutput["message"])
				}
				if _, exists := rootMap["stack"]; !exists {
					t.Fatalf("%v: expected a 'stack' field in the output but didn't find one { %v }", desc, errJSON)
				}
			} else {
				t.Errorf("%v: expected root error is malformed { %v }", desc, errJSON)
			}

			// make sure messages are correct and stack elements exist (actual stack validation is in stack_test.go)
			if wrapMap, ok := errJSON["wrap"].([]map[string]interface{}); ok {
				if len(tt.wrapOutput) != len(wrapMap) {
					t.Fatalf("%v: expected number of wrap layers { %v } doesn't match actual { %v }", desc, len(tt.wrapOutput), len(wrapMap))
				}
				for i := 0; i < len(wrapMap); i++ {
					if _, exists := wrapMap[i]["message"]; !exists {
						t.Fatalf("%v: expected a 'message' field in the output but didn't find one { %v }", desc, errJSON)
					}
					if wrapMap[i]["message"] != tt.wrapOutput[i]["message"] {
						t.Errorf("%v: expected { %v } got { %v }", desc, wrapMap[i]["message"], tt.wrapOutput[i]["message"])
					}
					if _, exists := wrapMap[i]["stack"]; !exists {
						t.Fatalf("%v: expected a 'stack' field in the output but didn't find one { %v }", desc, errJSON)
					}
				}
			} else {
				t.Errorf("%v: expected wrap error is malformed { %v }", desc, errJSON)
			}
		})
	}
}
