// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorex

import "fmt"

// Wrap wraps an error with a message.
// If err is nil returns nil.
// If message is empty error is not wrapped.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	if message == "" {
		return err
	}
	return fmt.Errorf("%w: %s", err, message)
}

// WrapCause wraps err with a message and appends the cause.
// If err is empty returns nil.
// If cause is nil, returns err wrapped with message.
// If message is empty err is not wrapped.
func WrapCause(err, cause error, message string) error {
	if err == nil {
		return nil
	}
	if cause == nil {
		if message == "" {
			return err
		}
		return Wrap(err, message)
	}
	if message == "" {
		return fmt.Errorf("%w: %v", err, cause)
	}
	return fmt.Errorf("%w: %s: %v", err, message, cause)
}