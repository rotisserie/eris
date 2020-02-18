package bench_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/rotisserie/eris"
)

var (
	global interface{}
	cases  = []struct {
		layers int
	}{
		{1},
		{10},
		{100},
		{1000},
	}
)

func wrapStdErrors(layers int) error {
	err := errors.New("std errors")
	for i := 0; i < layers; i++ {
		err = fmt.Errorf("wrap %v: %w", i, err)
	}
	return err
}

func wrapPkgErrors(layers int) error {
	err := pkgerrors.New("std errors")
	for i := 0; i < layers; i++ {
		err = pkgerrors.Wrapf(err, "wrap %v", i)
	}
	return err
}

func wrapEris(layers int) error {
	err := eris.New("std errors")
	for i := 0; i < layers; i++ {
		err = eris.Wrapf(err, "wrap %v", i)
	}
	return err
}

func BenchmarkWrap(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("std errors %v layers", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapStdErrors(tc.layers)
			}
			b.StopTimer()
			global = err
		})

		b.Run(fmt.Sprintf("pkg errors %v layers", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapPkgErrors(tc.layers)
			}
			b.StopTimer()
			global = err
		})

		b.Run(fmt.Sprintf("eris %v layers", tc.layers), func(b *testing.B) {
			var err error
			for n := 0; n < b.N; n++ {
				err = wrapEris(tc.layers)
			}
			b.StopTimer()
			global = err
		})
	}
}

func BenchmarkFormat(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("std errors %v layers", tc.layers), func(b *testing.B) {
			err := wrapStdErrors(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprint(err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("pkg errors %v layers", tc.layers), func(b *testing.B) {
			err := wrapPkgErrors(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprint(err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("eris %v layers", tc.layers), func(b *testing.B) {
			err := wrapEris(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprint(err)
			}
			b.StopTimer()
			global = str
		})
	}
}

func BenchmarkStack(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("pkg errors %v layers", tc.layers), func(b *testing.B) {
			err := wrapPkgErrors(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprintf("%+v", err)
			}
			b.StopTimer()
			global = str
		})

		b.Run(fmt.Sprintf("eris %v layers", tc.layers), func(b *testing.B) {
			err := wrapEris(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				str = fmt.Sprintf("%+v", err)
			}
			b.StopTimer()
			global = str
		})
	}
}

func BenchmarkJSON(b *testing.B) {
	for _, tc := range cases {
		b.Run(fmt.Sprintf("eris %v layers", tc.layers), func(b *testing.B) {
			err := wrapEris(tc.layers)
			b.ResetTimer()
			var str string
			for n := 0; n < b.N; n++ {
				jsonFmt := eris.NewDefaultJSONFormat(eris.FormatOptions{
					InvertOutput: true,
					WithTrace:    true,
					InvertTrace:  true,
				})
				jsonErr := eris.ToCustomJSON(err, jsonFmt)
				jsonStr, _ := json.Marshal(jsonErr)
				str = fmt.Sprint(jsonStr)
			}
			b.StopTimer()
			global = str
		})
	}
}
