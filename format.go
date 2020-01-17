package eris

import (
	"fmt"
)

// Format defines an error output format to be used with the default formatter.
type Format struct {
	WithTrace    bool   // Flag that enables stack trace output.
	MsgStackSep  string // Separator between error messages and stack frame data.
	PreStackSep  string // Separator at the beginning of each stack frame.
	StackElemSep string // Separator between elements of each stack frame.
	ErrorSep     string // Separator between each error in the chain.
}

// NewDefaultFormat conveniently returns a basic format for the default string formatter.
func NewDefaultFormat(withTrace bool) Format {
	stringFmt := Format{
		WithTrace: withTrace,
		ErrorSep:  ":",
	}
	if withTrace {
		stringFmt.MsgStackSep = "\n"
		stringFmt.PreStackSep = "\t"
		stringFmt.StackElemSep = ":"
		stringFmt.ErrorSep = "\n"
	}
	return stringFmt
}

// UnpackedError represents complete information about an error.
//
// This type can be used for custom error logging and parsing. Use `eris.Unpack` to build an UnpackedError
// from any error type. The ErrChain and ErrRoot fields correspond to `wrapError` and `rootError` types,
// respectively. If any other error type is unpacked, it will appear in the ExternalErr field.
type UnpackedError struct {
	ErrRoot     ErrRoot
	ErrChain    []ErrLink
	ExternalErr string
}

// Unpack returns UnpackedError type for a given golang error type.
func Unpack(err error) UnpackedError {
	upErr := UnpackedError{}
	switch err := err.(type) {
	case nil:
		return upErr
	case *rootError:
		upErr.unpackRootErr(err)
	case *wrapError:
		upErr.unpackWrapErr(err)
	default:
		upErr.ExternalErr = err.Error()
	}
	return upErr
}

// ToCustomString returns a custom formatted string for a given eris error.
//
// To declare custom format, the Format object has to be passed as an argument.
// An error without trace will be formatted as following:
//
//   <Root error msg>[Format.ErrorSep]<Wrap error msg>
//
// An error with trace will be formatted as following:
//
//   <Root error msg>[Format.MsgStackSep]
//   [Format.PreStackSep]<Method1>[Format.StackElemSep]<File1>[Format.StackElemSep]<Line1>[Format.ErrorSep]
//   [Format.PreStackSep]<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>[Format.ErrorSep]
//   <Wrap error msg>[Format.MsgStackSep]
//   [Format.PreStackSep]<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>[Format.ErrorSep]
func ToCustomString(err error, format Format) string {
	upErr := Unpack(err)
	if !format.WithTrace {
		format.ErrorSep = ": "
	}
	var str string
	if upErr.ErrRoot.Msg != "" || len(upErr.ErrRoot.Stack) > 0 {
		str += upErr.ErrRoot.formatStr(format)
		if format.WithTrace && len(upErr.ErrChain) > 0 {
			str += format.ErrorSep
		}
	}

	for _, eLink := range upErr.ErrChain {
		if !format.WithTrace {
			str += format.ErrorSep
		}
		str += eLink.formatStr(format)
		str += format.MsgStackSep
	}

	if upErr.ExternalErr != "" {
		str += fmt.Sprint(upErr.ExternalErr)
	}

	return str
}

// ToString returns a default formatted string for a given eris error.
//
// An error without trace will be formatted as following:
//
//   <Root error msg>: <Wrap error msg>
//
// An error with trace will be formatted as following:
//
//   <Root error msg>
//     <Method1>:<File1>:<Line1>
//     <Method2>:<File2>:<Line2>
//   <Wrap error msg>
//     <Method2>:<File2>:<Line2>
func ToString(err error, withTrace bool) string {
	return ToCustomString(err, NewDefaultFormat(withTrace))
}

// ToCustomJSON returns a JSON formatted map for a given eris error.
//
// To declare custom format, the Format object has to be passed as an argument.
// An error without trace will be formatted as following:
//
//   {
//     "root": {
//       "message": "Root error msg",
//     },
//     "wrap": [
//       {
//         "message": "Wrap error msg'",
//       }
//     ]
//   }
//
// An error with trace will be formatted as following:
//
//   {
//     "root": {
//       "message": "Root error msg",
//       "stack": [
//         "<Method1>[Format.StackElemSep]<File1>[Format.StackElemSep]<Line1>",
//         "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>"
//       ]
//     }
//     "wrap": [
//       {
//         "message": "Wrap error msg",
//         "stack": "<Method2>[Format.StackElemSep]<File2>[Format.StackElemSep]<Line2>"
//       }
//     ]
//   }
func ToCustomJSON(err error, format Format) map[string]interface{} {
	upErr := Unpack(err)
	if !format.WithTrace {
		format.ErrorSep = ": "
	}
	jsonMap := make(map[string]interface{})
	if upErr.ErrRoot.Msg != "" || len(upErr.ErrRoot.Stack) > 0 {
		jsonMap["root"] = upErr.ErrRoot.formatJSON(format)
	}

	if len(upErr.ErrChain) > 0 {
		var wrapArr []map[string]interface{}
		for _, eLink := range upErr.ErrChain {
			wrapMap := eLink.formatJSON(format)
			wrapArr = append(wrapArr, wrapMap)
		}
		jsonMap["wrap"] = wrapArr
	}

	if upErr.ExternalErr != "" {
		jsonMap["external"] = fmt.Sprint(upErr.ExternalErr)
	}

	return jsonMap
}

// ToJSON returns a JSON formatted map for a given eris error.
//
// An error without trace will be formatted as following:
//
//   {
//     "root": [
//       {
//         "message": "Root error msg"
//       }
//     ],
//     "wrap": {
//       "message": "Wrap error msg"
//     }
//   }
//
// An error with trace will be formatted as following:
//
//   {
//     "root": [
//       {
//         "message": "Root error msg",
//         "stack": [
//           "<Method1>:<File1>:<Line1>",
//           "<Method2>:<File2>:<Line2>"
//         ]
//       }
//     ],
//     "wrap": {
//       "message": "Wrap error msg",
//       "stack": "<Method2>:<File2>:<Line2>"
//     }
//   }
func ToJSON(err error, withTrace bool) map[string]interface{} {
	return ToCustomJSON(err, NewDefaultFormat(withTrace))
}

func (upErr *UnpackedError) unpackRootErr(err *rootError) {
	upErr.ErrRoot.Msg = err.msg
	upErr.ErrRoot.Stack = err.stack.get()
}

func (upErr *UnpackedError) unpackWrapErr(err *wrapError) {
	// prepend links in stack trace order
	link := ErrLink{Msg: err.msg}
	wFrames := err.stack.get()
	if len(wFrames) > 0 {
		link.Frame = wFrames[0]
	}
	upErr.ErrChain = append([]ErrLink{link}, upErr.ErrChain...)

	nextErr := err.Unwrap()
	switch nextErr := nextErr.(type) {
	case *rootError:
		upErr.unpackRootErr(nextErr)
	case *wrapError:
		upErr.unpackWrapErr(nextErr)
	}

	// insert the wrap frame into the root stack
	upErr.ErrRoot.Stack.insertFrame(wFrames)
}

// ErrRoot represents an error stack and the accompanying message.
type ErrRoot struct {
	Msg   string
	Stack Stack
}

func (err *ErrRoot) formatStr(format Format) string {
	str := err.Msg
	str += format.MsgStackSep
	if format.WithTrace {
		stackArr := err.Stack.format(format.StackElemSep)
		for i, frame := range stackArr {
			str += format.PreStackSep
			str += frame
			if i < len(stackArr)-1 {
				str += format.ErrorSep
			}
		}
	}
	return str
}

func (err *ErrRoot) formatJSON(format Format) map[string]interface{} {
	rootMap := make(map[string]interface{})
	rootMap["message"] = fmt.Sprint(err.Msg)
	if format.WithTrace {
		rootMap["stack"] = err.Stack.format(format.StackElemSep)
	}
	return rootMap
}

// ErrLink represents a single error frame and the accompanying message.
type ErrLink struct {
	Msg   string
	Frame StackFrame
}

func (eLink *ErrLink) formatStr(format Format) string {
	str := eLink.Msg
	str += format.MsgStackSep
	if format.WithTrace {
		str += format.PreStackSep
		str += eLink.Frame.format(format.StackElemSep)
	}
	return str
}

func (eLink *ErrLink) formatJSON(format Format) map[string]interface{} {
	wrapMap := make(map[string]interface{})
	wrapMap["message"] = fmt.Sprint(eLink.Msg)
	if format.WithTrace {
		wrapMap["stack"] = eLink.Frame.format(format.StackElemSep)
	}
	return wrapMap
}
