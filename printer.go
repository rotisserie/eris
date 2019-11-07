package eris

import (
	"fmt"
)

type Printer interface {
	Print(e error)
}

func Print(e error, fmtr formatter) {
	format := fmtr.GetFormat()
	switch e.(type) {
	case *wrapError:
		format.printWrapError(e.(*wrapError))
	case *rootError:
		format.printRootError(e.(*rootError))
	default:
		format.printError(e)
	}
}

func (format *format) printWrapError(err *wrapError) {
	fmt.Print(err.msg)
	fmt.Print(format.msg)
	fmt.Print(err.frame.formatFrame(format))
	fmt.Print(format.sep)

	nextErr := err.Unwrap()
	if nextErr == nil {
		return
	} else {
		switch nextErr.(type) {
		case *wrapError:
			format.printWrapError(nextErr.(*wrapError))
		case *rootError:
			format.printRootError(nextErr.(*rootError))
		default:
			format.printError(nextErr)
		}
	}
}

func (format *format) printRootError(err *rootError) {
	fmt.Print(err.msg)
	fmt.Print(format.msg)

	// todo: maybe move the rest to a stack.format method
	if format.traceFmt == nil {
		fmt.Print(format.sep)
	}

	for _, f := range *err.stack {
		frame := frame(f)
		s := frame.formatFrame(format)
		fmt.Print(s)
	}
	fmt.Print(format.sep)
}

func (format *format) printError(err error) {
	fmt.Print(err)
}

// todo: should this be moved to stack.go?
func (f frame) formatFrame(format *format) string {
	var s string
	if format.traceFmt == nil {
		s = fmt.Sprintf("%v", format.sep)
	} else {
		fdata := f.get()
		traceSep := format.traceFmt.sep
		s = fmt.Sprintf("%v%v%v%v%v%v%v", format.traceFmt.tBeg, fdata.name, traceSep,
			fdata.file, traceSep, fdata.line, format.traceFmt.tEnd)
	}
	return s
}
