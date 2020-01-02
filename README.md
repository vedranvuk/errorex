# ErrorEx

Package errorex provides additional error functionality. For now, implements a single type that helps with common error usage patterns introduced in `go1.13`, which it requires.

## Example

```
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
	return ErrMiddleSubVar.WithArgs("1337")
}

type Pkg struct {
	middle *Middle
}

func (p *Pkg) Bad() error {
	return ErrPkgSubVar.CauseArgs(p.middle.Bad(), "69")
}

type App struct {
	pkg *Pkg
}

func (a *App) Bad() error {
	return ErrAppSubVar.CauseArgs(a.pkg.Bad(), "42")
}

func TestMultiLevel(t *testing.T) {
	prog := &App{&Pkg{&Middle{}}}
	err := prog.Bad()
	if s := err.Error(); s != "command: method > detail '42' < package: method > detail: '69' < middleware: method > detail: '1337'" {
		t.Fatalf("Multilevel fail, want 'command: method > detail '42' < package: method > detail: '69' < middleware: method > detail: '1337'', got %s", s)
	}
}

```

## Status

Work in progress, subject to change.

## License

MIT, see the included LICENSE file.