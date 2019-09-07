package eris

import (
	"fmt"

	"golang.org/x/xerrors"
)

func New(msg string) error {
	return &errorStr{
		msg:   msg,
		frame: xerrors.Caller(1),
	}
}

func Wrap(err error, msg string) error {
	return &errorStr{
		msg:   msg,
		err:   err,
		frame: xerrors.Caller(1),
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return &errorStr{
		msg:   msg,
		err:   err,
		frame: xerrors.Caller(1),
	}
}

type errorStr struct {
	msg   string
	err   error
	frame xerrors.Frame
}

func (e errorStr) Error() string {
	return fmt.Sprint(e)
}

func (e errorStr) Format(f fmt.State, c rune) {
	xerrors.FormatError(e, f, c)
}

func (e errorStr) FormatError(p xerrors.Printer) error {
	p.Print(e.msg)
	if p.Detail() {
		e.frame.Format(p)
	}
	return e.err
}

func (e errorStr) Unwrap() error {
	return e.err
}
