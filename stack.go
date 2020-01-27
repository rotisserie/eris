package eris

import (
	"fmt"
	"runtime"
	"strings"
)

// Stack is an array of stack frames stored in a human readable format.
type Stack []StackFrame

// insertPC inserts a wrap error program counter (pc) into the correct place of the root error stack trace.
// TODO: this function can be optimized
func (rootPCs *stack) insertPC(wrapPCs stack) {
	if rootPCs == nil || len(wrapPCs) == 0 {
		return
	} else if len(wrapPCs) == 1 {
		// append the pc to the end if there's only one
		*rootPCs = append(*rootPCs, wrapPCs[0])
		return
	}
	for at, f := range *rootPCs {
		if f == wrapPCs[0] {
			// break if the stack already contains the pc
			break
		} else if f == wrapPCs[1] {
			// insert the first pc into the stack if the second pc is found
			*rootPCs = insert(*rootPCs, wrapPCs[0], at)
			break
		}
	}
}

// format returns an array of formatted stack frames.
func (s Stack) format(sep string) []string {
	var str []string
	for _, f := range s {
		str = append(str, f.format(sep))
	}
	return str
}

// StackFrame stores a frame's runtime information in a human readable format.
type StackFrame struct {
	Name string
	File string
	Line int
}

// format returns a formatted stack frame.
func (f *StackFrame) format(sep string) string {
	return fmt.Sprintf("%v%v%v%v%v", f.Name, sep, f.File, sep, f.Line)
}

// caller returns a single stack frame. the argument skip is the number of stack frames
// to ascend, with 0 identifying the caller of Caller.
func caller(skip int) *frame {
	pc, _, _, _ := runtime.Caller(skip)
	var f frame = frame(pc)
	return &f
}

// frame is a single program counter of a stack frame.
type frame uintptr

// get returns a human readable stack frame.
func (f frame) get() StackFrame {
	frame := StackFrame{
		Name: "unknown",
		File: "unknown",
	}

	pc := uintptr(f) - 1
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		name := fn.Name()
		i := strings.LastIndex(name, "/")
		name = name[i+1:]
		file, line := fn.FileLine(pc)

		frame = StackFrame{
			Name: name,
			File: file,
			Line: line,
		}
	}

	return frame
}

// callers returns a stack trace. the argument skip is the number of stack frames to skip
// before recording in pc, with 0 identifying the frame for Callers itself and 1 identifying
// the caller of Callers.
func callers(skip int) *stack {
	const depth = 64
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	var st stack = pcs[0 : n-2]
	return &st
}

// stack is an array of program counters.
type stack []uintptr

// get returns a human readable stack trace.
func (s *stack) get() []StackFrame {
	var sFrames []StackFrame
	for _, f := range *s {
		frame := frame(f)
		sFrame := frame.get()
		sFrames = append(sFrames, sFrame)
	}
	return sFrames
}

// isGlobal determines if the stack trace represents a global error
func (s *stack) isGlobal() bool {
	frames := s.get()
	for _, f := range frames {
		if strings.ToLower(f.Name) == "runtime.doinit" {
			return true
		}
	}
	return false
}

func insert(s stack, u uintptr, at int) stack {
	// this inserts the pc by breaking the stack into two slices (s[:at] and s[at:])
	return append(s[:at], append([]uintptr{u}, s[at:]...)...)
}
