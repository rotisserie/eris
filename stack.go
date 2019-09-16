package eris

import (
	"runtime"
	"strings"
)

// frame is a single program counter of a stack frame.
type frame uintptr

func caller() *frame {
	pc, _, _, _ := runtime.Caller(2)
	var f frame = frame(pc)
	return &f
}

func (f frame) pc() uintptr {
	return uintptr(f) - 1
}

func (f frame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

func (f frame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

func (f frame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return "unknown"
	}
	name := fn.Name()
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	return name
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
