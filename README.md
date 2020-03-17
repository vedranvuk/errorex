# ErrorEx

Package errorex implements the error interface with a custom error type that supports pre-formatting, inheritance, causes, custom data and extra error payloads.

It requires`go1.13`.

## Example

Errors can be derived and derived errors will respond properly to errors.Is().

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

```

ErrorEx can also carry an error that caused _this_ error (retrievable by ErrorEx.Cause(), and custom data.

```
var (
	ErrBase       = errorex.New("mypackage")
	ErrUnmarshal  = ErrBase.WrapFormat("marshal error: '%s'")
	ErrInvalidPos = ErrBase.WrapFormat("invalid position: '%d:%d'")
)

func unmarshal(name string) error {

	data := ""
	if err := json.Unmarshal([]byte{}, data); err != nil {
		return ErrUnmarshal.WrapCauseArgs(err, name)
	}
	return nil
}

type ErrorData struct {
	Line   int
	Column int
}

func gotopos() error {
	return ErrInvalidPos.WrapDataArgs(&ErrorData{32, 64}, 32, 64)
}

func main() {
	err := unmarshal("MyValue")
	fmt.Println(err)
	// Outputs:
	// mypackage: marshal error: 'MyValue' < unexpected end of JSON input

	err = gotopos()
	fmt.Println(err)
	// Output:
	// mypackage: invalid position: '32:64'

	if eex, ok := (err).(*errorex.ErrorEx); ok {
		if data, ok := (eex.Datas()).(*ErrorData); ok {
			fmt.Println(data)
		}
	}
	// Output:
	// &{32 64}
}

```

If you have, say, some sort of a sink function that manages multiple objects that return errors you can use Extra() to store those errors and Extras() to retrieve them later.

```
err := New("base error")
err.Extra(New("Extra 1"))
err.Extra(New("Extra 2"))
err.Extra(New("Extra 3"))

// range over extra errors.
for _, e := range err.Extras() {
	_ = e
}

fmt.Println(err)
// Output: base error + Extra 1 + Extra 2 + Extra 3
```

## Status

Work in progress, subject to change.

## License

MIT, see the included LICENSE file.