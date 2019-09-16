package eris

import (
// "testing"

// "github.com/stretchr/testify/assert"
)

// todo: add tests for nil errors (both with New/Errorf and Wrap/Wrapf)

// todo: need to fix the format method to make this test work again
// todo: add another section to this test using Errorf
// func TestErrorWrapping(t *testing.T) {
// 	tests := map[string]struct {
// 		wrapStrs []string // input error strings
// 		errStr   string   // expected output string
// 	}{
// 		"standard error wrapping": {
// 			wrapStrs: []string{"test error", "additional context", "even more context"},
// 			errStr:   "even more context: additional context: test error",
// 		},
// 	}
// 	for desc, tc := range tests {
// 		tc := tc
// 		t.Run(desc, func(t *testing.T) {
// 			t.Parallel()
// 			var err error
// 			for i, str := range tc.wrapStrs {
// 				if i == 0 {
// 					err = New(str)
// 				} else {
// 					err = Wrap(err, str)
// 				}
// 			}
// 			assert.Equalf(t, tc.errStr, err.Error(), "%v: expected { %v } got { %v }", desc, tc.errStr, err)
// 		})
// 	}
// }
