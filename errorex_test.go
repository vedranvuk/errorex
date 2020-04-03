// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorex

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	if New("test error").Error() != "test error" {
		t.Fatal()
	}
	if New("%s error").Error() != "%s error" {
		t.Fatal()
	}
	if s := New("%s error").WrapArgs("test").Error(); s != "%s error: test" {
		t.Fatal(s)
	}
}

func TestNewFormat(t *testing.T) {
	if NewFormat("%s error").WrapArgs("test").Error() != "test error" {
		t.Fatal()
	}
}

func TestWrap(t *testing.T) {
	if New("base").Wrap("sub1").Error() != "base: sub1" {
		t.Fatal()
	}
	if New("base").Wrap("sub1").Wrap("sub2").Error() != "base: sub1 > sub2" {
		t.Fatal()
	}
	if New("base").Wrap("sub1").Wrap("sub2").Wrap("sub3").Error() != "base: sub1; sub2 > sub3" {
		t.Fatal()
	}
	if New("base").Wrap("sub1").Wrap("sub2").Wrap("sub3").Wrap("sub4").Error() != "base: sub1; sub2; sub3 > sub4" {
		t.Fatal()
	}
}

func TestWrapFormat(t *testing.T) {
	if New("base").WrapFormat("sub%s").WrapArgs("1").Error() != "base: sub1" {
		t.Fatal()
	}
	if New("base").WrapFormat("sub%s").WrapArgs("1").WrapFormat("sub%s").WrapArgs("2").Error() != "base: sub1 > sub2" {
		t.Fatal()
	}
	if New("base").WrapFormat("sub%s").WrapArgs("1").WrapFormat("sub%s").WrapArgs("2").WrapFormat("sub%s").WrapArgs("3").Error() != "base: sub1; sub2 > sub3" {
		t.Fatal()
	}
	if New("base").WrapFormat("sub%s").WrapArgs("1").WrapFormat("sub%s").WrapArgs("2").WrapFormat("sub%s").WrapArgs("3").WrapFormat("sub%s").WrapArgs("4").Error() != "base: sub1; sub2; sub3 > sub4" {
		t.Fatal()
	}
}

func TestCause(t *testing.T) {
	if New("base").Wrap("sub1").WrapCause("fail", New("cause")).Error() != "base: sub1 > fail < cause" {
		t.Fatal()
	}
	if New("base").Wrap("sub1").WrapCause("fail", New("cause").WrapCause("deep", New("cause"))).Error() != "base: sub1 > fail < cause: deep < cause" {
		t.Fatal()
	}
	if New("base").WrapFormat("%s").WrapCauseArgs(New("cause"), "error").Error() != "base: error < cause" {
		t.Fatal()
	}

	base := New("base").WrapCause("base error", New("basecause"))
	cause := New("cause").WrapCause("cause error", New("causecause"))
	if s := base.WrapCause("error", cause).Error(); s != "base: base error < basecause > error < cause: cause error < causecause" {
		t.Fatal(s)
	}
}

func TestData(t *testing.T) {

	data := "test"

	if New("base").WrapData("error", data).Data().(string) != data {
		t.Fatal()
	}

	if New("base").WrapFormat("%s").WrapDataArgs(data, "error").Data().(string) != data {
		t.Fatal()
	}

	if New("base").WrapDataFormat("%s", data).WrapArgs("test").Data().(string) != data {
		t.Fatal()
	}
}

func TestExtra(t *testing.T) {

	extra1 := New("extra1")
	extra2 := New("extra2")
	extra3 := New("extra3")
	err := New("base").Extra(extra1).Extra(extra2).Extra(extra3)

	if err.Error() != "base + extra1 + extra2 + extra3" {
		t.Fatal()
	}

	extras := err.Extras()
	if len(extras) != 3 {
		t.Fatal()
	}
	if extras[0] != extra1 {
		t.Fatal()
	}
	if extras[1] != extra2 {
		t.Fatal()
	}
	if extras[2] != extra3 {
		t.Fatal()
	}
}

func TestUnwrap(t *testing.T) {
	base := New("base")
	wrap := base.Wrap("wrap")
	if wrap.Unwrap() != base {
		t.Fatal()
	}
}

func TestIs(t *testing.T) {
	base := New("base")
	wrap1 := base.Wrap("wrap1")
	wrap2 := wrap1.Wrap("wrap2")
	basecause := New("basecause")
	cause := basecause.Wrap("cause")
	err := wrap2.WrapCause("error", cause)
	if !errors.Is(err, basecause) {
		t.Fatal()
	}
	if !errors.Is(err, base) {
		t.Fatal()
	}
}
