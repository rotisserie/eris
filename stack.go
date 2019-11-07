package eris

import (
	"runtime"
	"strings"
)

// frame is a single program counter of a stack frame.
type frame uintptr

type stackFrame struct {
	name string
	file string
	line int
}

func caller() *frame {
	pc, _, _, _ := runtime.Caller(2)
	var f frame = frame(pc)
	return &f
}

func (f frame) pc() uintptr {
	return uintptr(f) - 1
}

func (f frame) get() *stackFrame {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return &stackFrame{
			name: "unknown",
			file: "unknown",
		}
	}

	name := fn.Name()
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	file, line := fn.FileLine(f.pc())

	return &stackFrame{
		name: name,
		file: file,
		line: line,
	}
}

// stack is an array of program counters.
type stack []uintptr

func callers() *stack {
	const depth = 64
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return &st
}
