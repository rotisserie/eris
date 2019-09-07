package eris

import (
	"reflect"
	"testing"
)

func TestNewDefaultFormatter(t *testing.T) {
	tests := []struct {
		name string
		want *format
	}{
		{
			"DefaultFormatter",
			&format{
				msg:       " (",
				op:        ":",
				ln:        ":",
				fpath:     ":",
				sep:       ")/n",
				withTrace: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultFormatter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultFormatter() = %v, want %v", got, tt.want)
			}
		})
	}
}
