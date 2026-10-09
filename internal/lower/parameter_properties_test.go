package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestParameterPropertiesSoundness(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"readonly view", "class C { constructor(readonly value: number) {} } const view: {value:number} = new C(1); view.value = 2;", "readonly"},
		{"mutable override", "class A {constructor(public value: number|string){}} class B extends A { constructor(public override value: number){ super(value); }}", "invariant-mutable"},
		{"readonly override", "class A {constructor(readonly value: number){}} class B extends A { constructor(public override value: number){ super(value); }}", "readonly inherited"},
		{"inherited initializer", "function invoke(read:()=>number):number{return read();} class A {constructor(public value:number){}} class B extends A {readonly before=invoke(()=>this.value); constructor(public override value:number){super(1);}} const c=new B(2);", "initialized"},
		{"early default", "class C {constructor(readonly x:number, readonly y:number=this.x){}} const c=new C(1);", "initialized"},
		{"field initializer", "function invoke(read:()=>number):number{return read();} class C { readonly before = invoke(() => this.value); constructor(readonly value:number){} } const c = new C(1);", "initialized"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want refusal %s", err, probe.reason)
			}
		})
	}
}

func TestParameterPropertyCheckerContracts(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"class C {constructor(readonly value:number){}} const c=new C(1); c.value=2;",
		"class C {constructor(private readonly value:number){}} const c=new C(1); console.log(`${c.value}`);",
		"class C {constructor(protected value:number){}} const c=new C(1); console.log(`${c.value}`);",
	} {
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := load.Load([]string{path})
		var rejected *load.CheckError
		if !errors.As(err, &rejected) {
			t.Fatalf("got %v, want checker rejection", err)
		}
	}
	lowersAndAgreesWithNode(t, "class C {constructor(public value:number){}} const c=new C(1); console.log(`${c.value}`); c.value=2; console.log(`${c.value}`);")
}

func TestParameterPropertyCallbackReceiver(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "class C {constructor(readonly run:(this:C)=>number){} read():number{return this.run();}} function callback(this:C):number{return 1;} const c=new C(callback); console.log(`${c.read()}`);")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "this parameter used as a value") {
		t.Fatalf("got %v, want NotYet dynamic receiver", err)
	}
}
