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
		basicInput      eris.UnpackedError
		formattedInput  eris.UnpackedError
		basicOutput     string
		formattedOutput string
	}{
		"basic root error": {
			basicInput: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
				},
			},
			formattedInput: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
					Stack: []eris.StackFrame{
						{
							Name: "eris.TestFormatStr",
							File: "format_test.go",
							Line: 99,
						},
						{
							Name: "golang.Runtime",
							File: "runtime.go",
							Line: 100,
						},
					},
				},
			},
			basicOutput:     "root error",
			formattedOutput: "root error\n\teris.TestFormatStr: format_test.go: 99\n\tgolang.Runtime: runtime.go: 100",
		},
		"basic wrapped error": {
			basicInput: eris.UnpackedError{
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
			formattedInput: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
					Stack: []eris.StackFrame{
						{
							Name: "eris.TestFormatStr",
							File: "format_test.go",
							Line: 99,
						},
						{
							Name: "golang.Runtime",
							File: "runtime.go",
							Line: 100,
						},
					},
				},
				ErrChain: []eris.ErrLink{
					{
						Msg: "additional context",
						Frame: eris.StackFrame{
							Name: "eris.TestFormatStr",
							File: "format_test.go",
							Line: 300,
						},
					},
				},
			},
			basicOutput:     "root error: additional context: even more context",
			formattedOutput: "root error\n\teris.TestFormatStr: format_test.go: 99\n\tgolang.Runtime: runtime.go: 100\nadditional context\n\teris.TestFormatStr: format_test.go: 300",
		},
		"basic external error": {
			basicInput: eris.UnpackedError{
				ExternalErr: "external error",
			},
			formattedInput: eris.UnpackedError{},
			basicOutput:    "external error",
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			if got := tt.basicInput.ToString(eris.NewDefaultFormat(false)); !reflect.DeepEqual(got, tt.basicOutput) {
				t.Errorf("ToString() got\n'%v'\nwant\n'%v'", got, tt.basicOutput)
			}
		})
		t.Run(desc, func(t *testing.T) {
			if got := tt.formattedInput.ToString(eris.NewDefaultFormat(true)); !reflect.DeepEqual(got, tt.formattedOutput) {
				t.Errorf("ToString() got\n'%v'\nwant\n'%v'", got, tt.formattedOutput)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := map[string]struct {
		basicInput      eris.UnpackedError
		formattedInput  eris.UnpackedError
		basicOutput     string
		formattedOutput string
	}{
		"basic root error": {
			basicInput: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
				},
			},
			formattedInput: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
					Stack: []eris.StackFrame{
						{
							Name: "eris.TestFormatStr",
							File: "format_test.go",
							Line: 99,
						},
						{
							Name: "golang.Runtime",
							File: "runtime.go",
							Line: 100,
						},
					},
				},
			},
			basicOutput:     `{"root":{"message":"root error"}}`,
			formattedOutput: `{"root":{"message":"root error","stack":["eris.TestFormatStr: format_test.go: 99","golang.Runtime: runtime.go: 100"]}}`,
		},
		"basic wrapped error": {
			basicInput: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
				},
				ErrChain: []eris.ErrLink{
					{
						Msg: "even more context",
					},
					{
						Msg: "additional context",
					},
				},
			},
			formattedInput: eris.UnpackedError{
				ErrRoot: eris.ErrRoot{
					Msg: "root error",
					Stack: []eris.StackFrame{
						{
							Name: "eris.TestFormatStr",
							File: "format_test.go",
							Line: 99,
						},
						{
							Name: "golang.Runtime",
							File: "runtime.go",
							Line: 100,
						},
					},
				},
				ErrChain: []eris.ErrLink{
					{
						Msg: "additional context",
						Frame: eris.StackFrame{
							Name: "eris.TestFormatStr",
							File: "format_test.go",
							Line: 300,
						},
					},
				},
			},
			basicOutput:     `{"root":{"message":"root error"},"wrap":[{"message":"even more context"},{"message":"additional context"}]}`,
			formattedOutput: `{"root":{"message":"root error","stack":["eris.TestFormatStr: format_test.go: 99","golang.Runtime: runtime.go: 100"]},"wrap":[{"message":"additional context","stack":"eris.TestFormatStr: format_test.go: 300"}]}`,
		},
		"basic external error": {
			basicInput: eris.UnpackedError{
				ExternalErr: "external error",
			},
			formattedInput:  eris.UnpackedError{},
			basicOutput:     `{"external":"external error"}`,
			formattedOutput: `{}`,
		},
	}
	for desc, tt := range tests {
		t.Run(desc, func(t *testing.T) {
			result, _ := json.Marshal(tt.basicInput.ToJSON(eris.NewDefaultFormat(false)))
			if got := string(result); !reflect.DeepEqual(got, tt.basicOutput) {
				t.Errorf("ToJSON() = %v, want %v", got, tt.basicOutput)
			}
		})
		t.Run(desc, func(t *testing.T) {
			result, _ := json.Marshal(tt.formattedInput.ToJSON(eris.NewDefaultFormat(true)))
			if got := string(result); !reflect.DeepEqual(got, tt.formattedOutput) {
				t.Errorf("ToJSON() = %v, want %v", got, tt.formattedOutput)
			}
		})
	}
}
