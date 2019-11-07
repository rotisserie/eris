// todo: add general eris intro here with explanations for all of the error types
package eris

import (
	"fmt"
)

// New creates a new root error with a static message.
func New(msg string) error {
	return &rootError{
		msg:   msg,
		stack: callers(),
	}
}

// Errorf creates a new root error with a formatted message.
func Errorf(format string, args ...interface{}) error {
	return &rootError{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

// Wrap adds additional context to all error types while maintaining the type of the original error.
//
// This method behaves differently for each error type. For root errors, the stack trace is reset to the current
// callers which ensures traces are correct when using global/sentinel error values. Wrapped error types are simply
// wrapped with the new context. For non-eris types (i.e. something other than root or wrap errors), a new root
// error is created for the original error and then it's wrapped with the additional context.
func Wrap(err error, msg string) error {
	return wrap(err, msg)
}

// Wrapf adds additional context to all error types while maintaining the type of the original error.
//
// This is a convenience method for wrapping errors with formatted messages and is otherwise the same as Wrap.
func Wrapf(err error, format string, args ...interface{}) error {
	return wrap(err, fmt.Sprintf(format, args...))
}

// todo: fix callers interface to allow passing the number to skip
//       this will allow callers to work from helper methods
//       pretty sure callers(4) is appropriate for this case and callers(3) is appropriate for New/Errorf
//       is this actually necessary?
// todo: test what happens to global error stack traces when this is called
//       does pointer cause global stack to be reset?
//       seems like it modifies the input error because of pointers, is this desirable?
func wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *rootError:
		e.stack = callers()
	case *wrapError:
	default:
		err = New(e.Error())
	}

	return &wrapError{
		msg:   msg,
		err:   err,
		frame: caller(),
	}
}

// todo: func Is(err, target error) bool

// todo: func Cause(err error) error

// todo: should this also have a single frame?
type rootError struct {
	msg   string
	stack *stack
}

func (e *rootError) Error() string {
	return fmt.Sprint(e)
}

func (e *rootError) Format(s fmt.State, verb rune) {
	var withTrace bool
	switch verb {
	case 'v':
		if s.Flag('+') {
			withTrace = true
		}
	}
	fmtr := NewDefaultFormatter(withTrace)
	Print(e, fmtr)
}

type wrapError struct {
	msg   string
	err   error
	frame *frame
}

func (e *wrapError) Error() string {
	return fmt.Sprint(e)
}

func (e *wrapError) Format(s fmt.State, verb rune) {
	var withTrace bool
	switch verb {
	case 'v':
		if s.Flag('+') {
			withTrace = true
		}
	}
	fmtr := NewDefaultFormatter(withTrace) // todo: formatting without trace is messed up
	Print(e, fmtr)
}

func (e *wrapError) Unwrap() error {
	return e.err
}
