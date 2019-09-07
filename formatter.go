package eris

type format struct {
	msg       string
	op        string
	ln        string
	fpath     string
	sep       string
	withTrace bool
}

type defaultFormatter struct {
	f format
}

type jsonFormatter struct {
	f format
}

func NewDefaultFormatter() *format {
	d := defaultFormatter{}
	d.f = format{
		msg:       " (",
		op:        ":",
		ln:        ":",
		fpath:     ":",
		sep:       ")/n",
		withTrace: true,
	}
	return &d.f
}
