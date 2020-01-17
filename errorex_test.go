// Copyright 2019 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorex

import (
	"errors"
	"testing"
)

func TestWrap(t *testing.T) {

	var (
		ErrPackageA    = New("PackageA")
		ErrPackageA1   = ErrPackageA.Wrap("Error 1")
		ErrPackageA2   = ErrPackageA.Wrap("Error 2")
		ErrPackageA21  = ErrPackageA2.Wrap("Warning 1")
		ErrPackageA211 = ErrPackageA21.Wrap("Info 1")
		ErrPackageAn   = ErrPackageA2.WrapFormat("Warning %d")

		ErrPackageB = NewFormat("Error %d")
	)

	if truth := errors.Is(ErrPackageA, ErrPackageA); !truth {
		t.Fatalf("Is [-1] failed, want 'true', got '%t'", truth)
	}

	if truth := errors.Is(ErrPackageAn, ErrPackageA2); !truth {
		t.Fatalf("Is [0] failed, want 'true', got '%t'", truth)
	}

	if truth := errors.Is(ErrPackageAn, ErrPackageA1); truth {
		t.Fatalf("Is [1] failed, want 'false', got '%t'", truth)
	}

	if truth := errors.Is(ErrPackageAn, ErrPackageA); !truth {
		t.Fatalf("Is [2] failed, want 'true', got '%t'", truth)
	}

	if truth := errors.Is(ErrPackageA21, ErrPackageA2); !truth {
		t.Fatalf("Is [3] failed, want 'true', got '%t'", truth)
	}

	if truth := errors.Is(ErrPackageA21, ErrPackageA); !truth {
		t.Fatalf("Is [4] failed, want 'true', got '%t'", truth)
	}

	if msg := ErrPackageA21.Error(); msg != "PackageA: Error 2 > Warning 1" {
		t.Fatalf("Error failed, want 'PackageA: Error 2 > Warning 1', got '%s'", msg)
	}

	ErrPackageAx := ErrPackageAn.WrapArgs(42)
	if msg := ErrPackageAx.Error(); msg != "PackageA: Error 2 > Warning 42" {
		t.Fatalf("Format() error, want 'PackageA: Error 2 > Warning 42', got '%s'", msg)
	}

	if msg := ErrPackageA211.Error(); msg != "PackageA: Error 2; Warning 1 > Info 1" {
		t.Fatalf("Error failed, want 'PackageA: Error 2; Warning 1 > Info 1', got '%s'", msg)
	}

	if truth := errors.Is(ErrPackageAx, ErrPackageAn); !truth {
		t.Fatalf("Is [5] failed, want 'true', got '%t'", truth)
	}

	if truth := errors.Is(ErrPackageAx, ErrPackageA); !truth {
		t.Fatalf("Is [6] failed, want 'true', got '%t'", truth)
	}

	ErrPackageBx := ErrPackageB.WrapArgs(69)
	if msg := ErrPackageBx.Error(); msg != "Error 69" {
		t.Fatalf("WrapArgs() failed, want 'Error 69', got '%s'", msg)
	}

}

func TestIs(t *testing.T) {

	var (
		ErrBaseA    = New("baseA")
		ErrDerivedA = ErrBaseA.Wrap("derivedA")

		ErrBaseB    = New("baseB")
		ErrDerivedB = ErrBaseB.Wrap("derivedB")
	)

	funcB := func() error {
		return ErrDerivedB
	}

	funcA := func() error {
		err := funcB()
		return ErrDerivedA.WrapCause("new", err)
	}

	err := funcA()

	if truth := errors.Is(err, ErrBaseA); !truth {
		t.Fatalf("Is(baseA) failed")
	}

	if truth := errors.Is(err, ErrBaseB); !truth {
		t.Fatalf("Is(baseB) failed")
	}

	if err.Error() != "baseA: derivedA > new < baseB: derivedB" {
		t.Fatal("fail")
	}
}

func TestCause(t *testing.T) {

	ErrA := New("ErrA")
	ErrB := ErrA.Wrap("ErrB")

	ErrC := New("ErrC")
	ErrD := ErrC.WrapFormat("Err%s")

	ErrE := ErrB.WrapCause("ErrE", ErrD.WrapArgs("X"))

	if s := ErrE.Error(); s != "ErrA: ErrB > ErrE < ErrC: ErrX" {
		t.Fatalf("TestCause failed, want 'ErrA: ErrB > ErrE < ErrC: ErrX', got '%s'", s)
	}
}

var (
	ErrApp       = New("command")
	ErrAppSub    = ErrApp.Wrap("method")
	ErrAppSubVar = ErrAppSub.WrapFormat("detail '%s'")

	ErrPkg       = New("package")
	ErrPkgSub    = ErrPkg.Wrap("method")
	ErrPkgSubVar = ErrPkgSub.WrapFormat("detail: '%s'")

	ErrMiddle       = New("middleware")
	ErrMiddleSub    = ErrMiddle.Wrap("method")
	ErrMiddleSubVar = ErrMiddleSub.WrapFormat("detail: '%s'")
)

type Middle struct{}

func (m *Middle) Bad() error {
	return ErrMiddleSubVar.WrapArgs("1337")
}

type Pkg struct {
	middle *Middle
}

func (p *Pkg) Bad() error {
	return ErrPkgSubVar.WrapCauseArgs(p.middle.Bad(), "69")
}

type App struct {
	pkg *Pkg
}

func (a *App) Bad() error {
	return ErrAppSubVar.WrapCauseArgs(a.pkg.Bad(), "42")
}

func TestMultiLevel(t *testing.T) {
	prog := &App{&Pkg{&Middle{}}}
	err := prog.Bad()
	if s := err.Error(); s != "command: method > detail '42' < package: method > detail: '69' < middleware: method > detail: '1337'" {
		t.Fatalf("Multilevel fail, want 'command: method > detail '42' < package: method > detail: '69' < middleware: method > detail: '1337'', got %s", s)
	}
}

func TestData(t *testing.T) {

	type Data struct {
		Line   int
		Column int
	}

	base := New("base").WrapFormat("error at '%d:%d'")

	err := base.WrapDataArgs(&Data{32, 64}, 32, 64)
	if err.Error() != "base: error at '32:64'" {
		t.Fatal("Data failed")
	}

	data, ok := (err.Data()).(*Data)
	if !ok {
		t.Fatal("Data failed")
	}
	if data.Line != 32 || data.Column != 64 {
		t.Fatal("Data failed")
	}

}
