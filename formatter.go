package eris

// format holds formatting definitions for error message, trace and their separator
type format struct {
	msg      string
	traceFmt *traceFormat
	sep      string
}

// traceFormat holds formatting definitions for the trace object
type traceFormat struct {
	tBeg string
	sep  string
	tEnd string
}

// defaultFormatter represents the default format for the errors in eris
type defaultFormatter struct {
	fmt format
}

// jsonFormatter represents the json format for the errors in eris
type jsonFormatter struct {
	fmt format
}

// NewDefaultFormatter returns a new defaultFormatter with or without trace
func NewDefaultFormatter(withTrace bool) *defaultFormatter {
	defaultFmtr := defaultFormatter{
		fmt: format{
			msg:      " ",
			traceFmt: nil,
			sep:      "\n",
		},
	}
	if withTrace {
		defaultFmtr.fmt.traceFmt = &traceFormat{
			tBeg: "(",
			sep:  ": ",
			tEnd: ")",
		}
	}
	return &defaultFmtr
}
