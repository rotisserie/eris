package eris

import (
	"fmt"
	"io"
)

// New creates a new root error.
func New(msg string) error {
	return &rootError{
		msg:   msg,
		stack: callers(),
	}
}

// Errorf create a new root error.
func Errorf(format string, args ...interface{}) error {
	return &rootError{
		msg:   fmt.Sprintf(format, args...),
		stack: callers(),
	}
}

// Wrapf adds additional context to all error types. For root error types, this method resets the stack
// trace, which solves the problem of incorrect stack traces when using global/sentinel error values. For
// wrapped error types, this method simply wraps the error with the new context. For all other error types,
// it returns a new error with a correct stack trace.
func Wrap(err error, msg string) error {
	if root, ok := err.(*rootError); ok {
		root.stack = callers()
		return &wrapError{
			msg:   msg,
			err:   root,
			frame: caller(),
		}
	}
	if wrap, ok := err.(*wrapError); ok {
		return &wrapError{
			msg:   msg,
			err:   wrap,
			frame: caller(),
		}
	}
	return &rootError{
		msg:   err.Error(),
		stack: callers(),
	}
}

// Wrapf adds additional context to all error types. For root error types, this method resets the stack
// trace, which solves the problem of incorrect stack traces when using global/sentinel error values. For
// wrapped error types, this method simply wraps the error with the new context. For all other error types,
// it returns a new error with a correct stack trace.
func Wrapf(err error, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	if root, ok := err.(*rootError); ok {
		root.stack = callers()
		return &wrapError{
			msg:   msg,
			err:   root,
			frame: caller(),
		}
	}
	if wrap, ok := err.(*wrapError); ok {
		return &wrapError{
			msg:   msg,
			err:   wrap,
			frame: caller(),
		}
	}
	return &rootError{
		msg:   err.Error(),
		stack: callers(),
	}
}

type rootError struct {
	msg   string
	stack *stack // todo: should this be a pointer?
}

func (e *rootError) Error() string {
	return fmt.Sprint(e)
}

// todo: this is pretty rough right now (use the DefaultFormatter to print)
func (e *rootError) Format(s fmt.State, verb rune) {
	io.WriteString(s, e.msg)
	for _, f := range *e.stack { // todo: check nil?
		frame := frame(f)
		io.WriteString(s, "\n\t")
		io.WriteString(s, fmt.Sprintf("%v | %v:%v", frame.file(), frame.name(), frame.line()))
	}
}

type wrapError struct {
	msg   string
	err   error
	frame *frame
}

func (e wrapError) Error() string {
	return fmt.Sprint(e)
}

func (e *wrapError) Format(s fmt.State, verb rune) {
	io.WriteString(s, e.msg)
	io.WriteString(s, " (")
	io.WriteString(s, fmt.Sprintf("%v | %v:%v", e.frame.file(), e.frame.name(), e.frame.line()))
	io.WriteString(s, ")")
}

func (e *wrapError) Unwrap() error {
	return e.err
}
