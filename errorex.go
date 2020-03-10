// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package errorex provides additional error functionality.
package errorex

import (
	"errors"
	"fmt"
)

// ErrorEx is an extended error type which provides utilities for
// error inheritance, causes and custom data payloads.
// ErrorEx is not safe for concurrent use.
type ErrorEx struct {
	// cause is the stored cause error.
	cause error
	// data is the stored custom data.
	data interface{}
	// err is optionally wrapped error.
	err error
	// txt is this error text/message/format string.
	txt string
	// fmt specifies if this error is a placeholder error whose
	// txt is used as a format string for derived errors.
	fmt bool
	// Extra are extra errors carried with this error.
	extra []error
}

// New returns a new ErrorEx and sets its' message.
func New(message string) *ErrorEx {
	return &ErrorEx{
		txt: message,
	}
}

// NewFormat returns a new ErrorEx and sets its text to a format string
// which will be used as a format string for errors deriving from it.
// Resulting error is used as a placeholder and will be skipped when
// printing but remains in the error chain and responds to Is() and As().
func NewFormat(format string) (err *ErrorEx) {
	err = New(format)
	err.fmt = true
	return
}

// Error implements the error interface. It uses a custom printing scheme:
// First error in the chain is separated with a ':' from derived error messages.
// Last error in the chain is separated from the error it derives from with a '>'.
// Multiple levels of derived errors are separated with a ';'.
// Cause errors format the same way and are appended to the error message after
// Extra errors carried by an error are separated by ;
// prefix '<'.
// Example:
//  mypackage: subsystem error; funcerror > detailederror; extra error < thirdpartypackage: subsystem error > detailederror
func (ee *ErrorEx) Error() (message string) {
	// Set base message.
	message = ee.txt
	if ee.fmt {
		message = ""
	}
	if ee.cause != nil {
		message = fmt.Sprintf("%s < %v", message, ee.cause)
	}
	// Build wrap stack.
	stack := []string{}
	for eex, ok := (ee.err).(*ErrorEx); ok; eex, ok = (eex.err).(*ErrorEx) {
		if cause := eex.Cause(); cause != nil {
			stack = append(stack, cause.Error())
		} else {
			if eex.fmt {
				continue
			}
		}
		stack = append(stack, eex.txt)
	}
	// Format stack.
	if len(stack) > 0 {
		if len(stack) == 1 {
			message = fmt.Sprintf("%s: %s", stack[0], message)
		} else {
			msg := fmt.Sprintf("%s:", stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			for len(stack) > 0 {
				if len(stack) == 1 {
					msg = fmt.Sprintf("%s %s", msg, stack[len(stack)-1])
				} else {
					msg = fmt.Sprintf("%s %s;", msg, stack[len(stack)-1])
				}
				stack = stack[:len(stack)-1]
			}
			message = fmt.Sprintf("%s > %s", msg, message)
		}

	}
	// Append extra.
	if len(ee.extra) > 0 {
		for _, ex := range ee.extra {
			message += fmt.Sprintf("; %s", ex.Error())
		}
	}
	return
}

// is is the implementation of Is.
func (ee *ErrorEx) is(target, cause error) (is bool) {
	is = ee == target
	if !is {
		is = errors.Is(ee.err, target)
	}
	if !is {
		is = errors.Is(cause, target)
	}
	return
}

// Is implements errors.Is().
func (ee *ErrorEx) Is(target error) bool {
	return ee.is(target, ee.cause)
}

// Unwrap implements error.Unwrap().
func (ee *ErrorEx) Unwrap() error {
	return ee.err
}

// Wrap wraps this error with a new error, sets new error message,
// then returns it.
func (ee *ErrorEx) Wrap(message string) *ErrorEx {
	return &ErrorEx{err: ee, txt: message}
}

// WrapFormat wraps this error with a new non-printable error whose
// message is a format string to derived errors.
// The resulting error txt is used as a format string for error text
// of derivation functions WrapArgs, WrapCauseArgs and WrapDataArgs.
//
// The resulting error is skipped when printing the error chain but
// remains in the error chan and responds to Is() and As() properly.
func (ee *ErrorEx) WrapFormat(format string) (err *ErrorEx) {
	err = ee.Wrap(format)
	err.fmt = true
	return
}

// WrapArgs derives a new error whose message will be formatted using
// specified args and this error message as a format string.
// WrapArgs should be used on errors which were constructed using
// NewFormat or WrapFormat and a format string.
func (ee *ErrorEx) WrapArgs(args ...interface{}) *ErrorEx {
	return ee.Wrap(fmt.Sprintf(ee.txt, args...))
}

// WrapCause returns a new derived ErrorEx that wraps a cause error.
// Calling errors.Is() on returned error returns true for target
// being the parent of both the returned error and the cause error
// that it wraps.
// Meaning:
//  ErrE := New("ErrA").Wrap("ErrB").WrapCause("ErrE", New("ErrC").Wrap("ErrD"))
//  errors.Is(ErrE, ErrA) // true
//  errors.Is(ErrE, ErrC) // true
//  fmt.Println(ErrF) // ErrA: ErrB > ErrC < ErrD: ErrE; ErrF
// Derived ErrorEx unwraps to this error.
// Wrapped cause error is retrievable with Cause().
func (ee *ErrorEx) WrapCause(message string, err error) *ErrorEx {
	return &ErrorEx{cause: err, err: ee, txt: message}
}

// WrapCauseArgs derives a new error which wraps a cause error and formats
// its error message from specified args and this error message as a format
// string. See WrapCause for more details.
func (ee *ErrorEx) WrapCauseArgs(err error, args ...interface{}) *ErrorEx {
	return &ErrorEx{cause: err, err: ee, txt: fmt.Sprintf(ee.txt, args...)}
}

// Cause returns the error that caused this error, which could be nil.
func (ee *ErrorEx) Cause() error {
	return ee.cause
}

// WrapData returns a new derived ErrorEx that wraps custom data.
func (ee *ErrorEx) WrapData(message string, data interface{}) *ErrorEx {
	return &ErrorEx{data: data, err: ee, txt: message}
}

// WrapDataArgs derives a new error which wraps custom data and formats
// its error message from specified args and this error message as a format
// string. See WrapData for more details.
func (ee *ErrorEx) WrapDataArgs(data interface{}, args ...interface{}) *ErrorEx {
	return &ErrorEx{data: data, err: ee, txt: fmt.Sprintf(ee.txt, args...)}
}

// Data returns custom data stored in this error, which could be nil.
func (ee *ErrorEx) Data() interface{} {
	return ee.data
}

// Extra appends an extra error to this error.
func (ee *ErrorEx) Extra(err error) {
	ee.extra = append(ee.extra, err)
}

// Extras returns extra errors.
func (ee *ErrorEx) Extras() []error {
	return ee.extra
}
