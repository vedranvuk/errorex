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
// error inheritance, causes, custom data payloads and extra errors.
// ErrorEx is not safe for concurrent use.
type ErrorEx struct {
	// txt is this error text/message/format string.
	txt string
	// err is optionally wrapped error.
	err error
	// fmt specifies if this error is a placeholder error whose
	// txt is used as a format string for derived errors.
	fmt bool
	// cause is the stored cause error.
	cause error
	// data is the stored custom data.
	data interface{}
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

// extrastring returns preformated extra error messages as a string.
func (ee *ErrorEx) extrastring() (message string) {
	if len(ee.extra) > 0 {
		for _, ex := range ee.extra {
			message += fmt.Sprintf(" + %s", ex.Error())
		}
	}
	return
}

// Error implements the error interface.
//
// It uses a custom printing scheme:
//
// First error in the chain is always separated with a ':' from derived error
// messages.
// Wrapped errors are separated with a ';' if there are more than 3 wrap levels
// and the error is between 3rd and last level.
// Last error in the wrap stack is always separated with a '>' unless it
// directly wraps the base error in which case it is separated by ':'.
//
// Example:
//  New("base").Wrap("sub1").Error()
//  Output: base: sub1
//
// Example:
//  New("base").Wrap("sub1").Wrap("sub2").Error()
//  Output: base: sub1 > sub2
//
// Example:
//  New("base").Wrap("sub1").Wrap("sub2").Wrap("sub3").Error()
//  Output: base: sub1; sub2 > sub3
//
// Example:
//  New("base").Wrap("sub1").Wrap("sub2").Wrap("sub3").Wrap("sub4").Error()
//  Output: base: sub1; sub2; sub3 > sub4
//
// Cause errors format the same way and are appended to final error after a '<'
// prefix.
//
// Example:
//  New("base").WrapCause("error", New("cause"))
//  Output: base: error < cause
//
// Extra errors carried by an error are appended and separated by ' + '
//
// Example:
//  New("base").Wrap("sub").Extra(New("extra"))
//  Output: base: sub + extra
//
// Errors created with NewFormat and WrapFormat are format placeholder errors
// and are not printed when printing the wrap chain.
//
// Errors with an empty message are skipped when printing, regardless if they
// carry causes or extra errors.
func (ee *ErrorEx) Error() (message string) {

	// Set base message.
	if ee.txt == "" {
		if ee.err != nil {
			message = ee.err.Error()
		}
	}
	if !ee.fmt {
		message = ee.txt
	}
	if ee.cause != nil {
		message = fmt.Sprintf("%s < %v", message, ee.cause)
	}

	// Build wrap stack.
	stack := []string{}
	for eex, ok := (ee.err).(*ErrorEx); ok; eex, ok = (eex.err).(*ErrorEx) {
		if eex.fmt || len(eex.txt) == 0 {
			continue
		}
		stack = append(stack, eex.txt+eex.extrastring())
		if cause := eex.Cause(); cause != nil {
			stack[len(stack)-1] += fmt.Sprintf(" < %s", cause.Error())
		}
	}

	// Format stack.
	if len(stack) > 0 {
		if len(stack) == 1 {
			if message == "" {
				message = stack[0]
			} else {
				message = fmt.Sprintf("%s: %s", stack[0], message)
			}
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
	message += ee.extrastring()

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
// Is returns true if either this or the cause error are siblings of target.
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
// message is a format string to errors further derived from it.
//
// Resulting error can be formatted to a derived error with WrapArgs,
// WrapCauseArgs and WrapDataArgs.
//
// The resulting error is skipped when printing the error chain but
// remains in the error chain and responds to Is() and As() properly.
func (ee *ErrorEx) WrapFormat(format string) (err *ErrorEx) {
	err = ee.Wrap(format)
	err.fmt = true
	return
}

// autoformat returns a formatted error message using this error message
// as a format string and specified args if this error is a format error.
// Otherwise, returns args as a single string separated by single space.
func (ee *ErrorEx) autoformat(args ...interface{}) string {
	if ee.fmt {
		return fmt.Sprintf(ee.txt, args...)
	}
	return fmt.Sprint(args...)
}

// WrapArgs derives a new error whose message will be formatted using
// specified args and this error message as a format string.
// WrapArgs should be used on errors which were constructed using
// NewFormat or WrapFormat using a format string as error message.
func (ee *ErrorEx) WrapArgs(args ...interface{}) *ErrorEx {
	return ee.Wrap(ee.autoformat(args...))
}

// WrapCause returns a new derived ErrorEx that wraps a cause error.
// Calling errors.Is() on returned error returns true for target
// being the parent of either the returned error and the cause error
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
	return &ErrorEx{cause: err, err: ee, txt: ee.autoformat(args...)}
}

// Cause returns the error that caused this error, which could be nil.
func (ee *ErrorEx) Cause() error {
	return ee.cause
}

// WrapData returns a new derived ErrorEx that wraps custom data.
func (ee *ErrorEx) WrapData(message string, data interface{}) *ErrorEx {
	return &ErrorEx{data: data, err: ee, txt: message}
}

// WrapDataFormat wraps an error like WrapFormat but attatches data to it.
func (ee *ErrorEx) WrapDataFormat(format string, data interface{}) *ErrorEx {
	err := ee.WrapFormat(format)
	err.data = data
	return err
}

// WrapDataArgs derives a new error which wraps custom data and formats
// its error message from specified args and this error message as a format
// string. See WrapData for more details.
func (ee *ErrorEx) WrapDataArgs(data interface{}, args ...interface{}) *ErrorEx {
	return &ErrorEx{data: data, err: ee, txt: ee.autoformat(args...)}
}

// Data returns this error data, which could be nil.
func (ee *ErrorEx) Data() (data interface{}) {
	return ee.data
}

// AnyData returns first set data down the complete error wrap chain starting from
// this error. Errors not of ErrorEx type are skipped. If no set data is found
// result will be nil.
func (ee *ErrorEx) AnyData() (data interface{}) {
	for err := error(ee); ; {
		if err == nil {
			break
		}
		if eex, ok := err.(*ErrorEx); ok {
			data = eex.Data()
			if data != nil {
				break
			}
		}
		err = errors.Unwrap(err)
	}
	return
}

// Extra appends an extra error to this error and returns self.
func (ee *ErrorEx) Extra(err error) *ErrorEx {
	ee.extra = append(ee.extra, err)
	return ee
}

// Extras returns extra errors.
func (ee *ErrorEx) Extras() []error {
	return ee.extra
}
