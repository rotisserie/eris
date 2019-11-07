package eris

import (
	"reflect"
	"testing"
)

func TestNewDefaultFormatter(t *testing.T) {
	type args struct {
		withTrace bool
	}
	tests := []struct {
		name string
		args args
		want *defaultFormatter
	}{
		{
			"DefaultFormatter (Without Trace)",
			args{
				withTrace: false,
			},
			&defaultFormatter{
				fmt: format{
					msg:      "",
					traceFmt: nil,
					sep:      "",
				},
			},
		},
		{
			"DefaultFormatter (With Trace)",
			args{
				withTrace: true,
			},
			&defaultFormatter{
				fmt: format{
					msg: " ",
					traceFmt: &traceFormat{
						tBeg: "(",
						sep:  ": ",
						tEnd: ")\n",
					},
					sep: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultFormatter(tt.args.withTrace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultFormatter() = %v, want %v", got, tt.want)
			}
		})
	}
}
