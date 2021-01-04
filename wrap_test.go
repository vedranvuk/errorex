// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorex

import (
	"errors"
	"testing"
)

var (
	ErrTest  = errors.New("test")
	ErrCause = errors.New("cause")
)

func TestUtilWrap(t *testing.T) {
	if Wrap(ErrTest, "").Error() != "test" {
		t.Fatal("TestUtilWrap failed")
	}
	if Wrap(ErrTest, "message").Error() != "test: message" {
		t.Fatal("TestUtilWrap failed")
	}
}

func TestUtilWrapCause(t *testing.T) {
	if WrapCause(ErrTest, nil, "").Error() != "test" {
		t.Fatal("TestUtilWrapCause failed")
	}
	if WrapCause(ErrTest, ErrCause, "").Error() != "test: cause" {
		t.Fatal("TestUtilWrapCause failed")
	}
	if WrapCause(ErrTest, nil, "message").Error() != "test: message" {
		t.Fatal("TestUtilWrapCause failed")
	}
	if WrapCause(ErrTest, ErrCause, "message").Error() != "test: cause: message" {
		t.Fatal("TestUtilWrapCause failed")
	}
}
