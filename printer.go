package eris

import (
	"encoding/json"
	"fmt"
)

// Format defines an error output format to be used with the default printer.
type Format struct {
	WithTrace bool   // Flag that enables stack trace output.
	Msg       string // Separator between error messages and stack frame data.
	TBeg      string // Separator at the beginning of each stack frame.
	TSep      string // Separator between elements of each stack frame.
	Sep       string // Separator between each error in the chain.
}

// NewDefaultFormat conveniently returns a basic format for the default string printer.
func NewDefaultFormat(withTrace bool) Format {
	stringFmt := Format{
		WithTrace: withTrace,
		Sep:       ": ",
	}
	if withTrace {
		stringFmt.Msg = "\n"
		stringFmt.TBeg = "\t"
		stringFmt.TSep = ":"
		stringFmt.Sep = "\n"
	}
	return stringFmt
}

// Printer defines a basic printer interface.
type Printer interface {
	// Sprint returns a formatted string for a given error.
	Sprint(err error) string
}

type defaultPrinter struct {
	format Format
}

// NewDefaultPrinter returns a basic printer that converts errors into strings.
func NewDefaultPrinter(format Format) Printer {
	return &defaultPrinter{
		format: format,
	}
}

// Sprint returns a default formatted string for a given error.
func (p *defaultPrinter) Sprint(err error) string {
	var str string
	switch err.(type) {
	case nil:
		return ""
	case *rootError:
		str = p.printRootError(err.(*rootError))
	case *wrapError:
		str = p.printWrapError(err.(*wrapError))
	default:
		str = fmt.Sprint(err) + p.format.Sep
	}
	return str
}

func (p *defaultPrinter) printRootError(err *rootError) string {
	str := err.msg
	str += p.format.Msg
	if p.format.WithTrace {
		stackArr := printStack(err.stack, p.format.TSep)
		for _, frame := range stackArr {
			str += p.format.TBeg
			str += frame
			str += p.format.Sep
		}
	}
	return str
}

func (p *defaultPrinter) printWrapError(err *wrapError) string {
	str := err.msg
	str += p.format.Msg
	if p.format.WithTrace {
		str += p.format.TBeg
		str += printFrame(err.frame, p.format.TSep)
	}
	str += p.format.Sep

	nextErr := err.Unwrap()
	switch nextErr.(type) {
	case nil:
		return ""
	case *rootError:
		str += p.printRootError(nextErr.(*rootError))
	case *wrapError:
		str += p.printWrapError(nextErr.(*wrapError))
	default:
		str += fmt.Sprint(nextErr) + p.format.Sep
	}

	return str
}

type jsonPrinter struct {
	format Format
}

// NewJSONPrinter returns a basic printer that converts errors into JSON formatted strings.
func NewJSONPrinter(format Format) Printer {
	return &jsonPrinter{
		format: format,
	}
}

// Sprint returns a JSON formatted string for a given error.
func (p *jsonPrinter) Sprint(err error) string {
	jsonMap := make(map[string]interface{})
	switch err.(type) {
	case nil:
		return "{}"
	case *rootError:
		jsonMap["error root"] = p.printRootError(err.(*rootError))
	case *wrapError:
		jsonMap = p.printWrapError(err.(*wrapError))
	default:
		jsonMap["external error"] = fmt.Sprint(err)
	}
	str, _ := json.Marshal(jsonMap)
	return string(str)
}

func (p *jsonPrinter) printRootError(err *rootError) map[string]interface{} {
	rootMap := make(map[string]interface{})
	rootMap["message"] = fmt.Sprint(err.msg)
	if p.format.WithTrace {
		rootMap["stack"] = printStack(err.stack, p.format.TSep)
	}
	return rootMap
}

func (p *jsonPrinter) printWrapError(err *wrapError) map[string]interface{} {
	jsonMap := make(map[string]interface{})

	nextErr := error(err)
	var wrapArr []map[string]interface{}
	for {
		if nextErr == nil {
			break
		} else if e, ok := nextErr.(*rootError); ok {
			jsonMap["error root"] = p.printRootError(e)
		} else if e, ok := nextErr.(*wrapError); ok {
			wrapMap := make(map[string]interface{})
			wrapMap["message"] = fmt.Sprint(e.msg)
			if p.format.WithTrace {
				wrapMap["stack"] = printFrame(e.frame, p.format.TSep)
			}
			wrapArr = append(wrapArr, wrapMap)
		} else {
			jsonMap["external error"] = fmt.Sprint(nextErr)
		}
		nextErr = Unwrap(nextErr)
	}
	jsonMap["error chain"] = wrapArr

	return jsonMap
}

func printFrame(f *frame, sep string) string {
	fData := f.get()
	return fmt.Sprintf("%v%v%v%v%v", fData.name, sep, fData.file, sep, fData.line)
}

func printStack(s *stack, sep string) []string {
	var str []string
	for _, f := range *s {
		frame := frame(f)
		str = append(str, printFrame(&frame, sep))
	}
	return str
}
