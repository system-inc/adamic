package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNamedRecordReadTypes(t *testing.T) {
	t.Parallel()
	p, err := lowerSource(t, `interface Options { target?:number; [key:string]:number|string|undefined } function named(o:Options):number|undefined{return o.target;} function indexed(o:Options,k:string):number|string|undefined{return o[k];}`)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Functions {
		if f.Name == "named" {
			r := f.Body[len(f.Body)-1].(ir.Return)
			if r.Value.Type() != ir.MaybeNumber {
				t.Fatalf("named read used index type %v", r.Value.Type())
			}
		}
		if f.Name == "indexed" {
			r := f.Body[len(f.Body)-1].(ir.Return)
			if r.Value.Type() != ir.Union {
				t.Fatalf("indexed read lost index type %v", r.Value.Type())
			}
		}
	}
}

// A slot whose union holds null boxes it, as tsc's CompilerOptions does. A declared read still unboxes through the
// checked tag, since a write through the index signature can put any value of the union under a declared key.
func TestNullableRecordSlots(t *testing.T) {
	t.Parallel()
	p, err := lowerSource(t, `interface R {named?:number;[key:string]:number|string|boolean|null|undefined} function named(r:R):number|undefined{return r.named;} function indexed(r:R,k:string):number|string|boolean|null|undefined{return r[k];}`)
	if err != nil {
		t.Fatal(err)
	}
	checked := false
	for _, f := range p.Functions {
		switch f.Name {
		case "named":
			if r := f.Body[len(f.Body)-1].(ir.Return); r.Value.Type() != ir.MaybeNumber {
				t.Fatalf("named read used the slot type %v", r.Value.Type())
			}
		case "indexed":
			if r := f.Body[len(f.Body)-1].(ir.Return); r.Value.Type() != ir.Union {
				t.Fatalf("indexed read lost the slot type %v", r.Value.Type())
			}
		case "record_declared_member":
			checked = true
		}
	}
	if !checked {
		t.Fatal("the declared read skipped the checked unbox")
	}
}

// typeof says object for a dictionary and a fixed object alike, so a local narrowed to one of them can't be checked
// against the other after a call writes it again. Without the dictionary among the kinds that stop, native read the
// record's storage as a Box and printed 4.6e-310 where Node prints 5.
func TestNarrowedUnionMemberBesideRecord(t *testing.T) {
	t.Parallel()
	for _, declared := range []string{"Box | Record<string, number>", "Box | Record<string, number> | null"} {
		t.Run(declared, func(t *testing.T) {
			_, err := lowerSource(t, `class Box { v = 1; }
let x: `+declared+` = new Box();
function swap(): void { const d: Record<string, number> = {}; d["v"] = 5; x = d; }
if (x instanceof Box) { swap(); console.log(String(x.v)); }`)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(err.Error(), "object tag cannot be checked with typeof") {
				t.Fatalf("got %v, want the object-tag NotYet", err)
			}
		})
	}
}

func TestNamedRecordRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`interface R {named?:number[];[key:string]:number[]|string|undefined} function read(r:R):number|undefined {return r.named?.length;}`, "dictionary member narrowed to an object kind"},
		{`interface R {named?:number;[key:string]:number|string|undefined} function write(r:R,k:string):void {r[k]="wrong";}`, "named property named"},
		{`interface R {named?:number;[key:string]:number|string|undefined} const r:R={}; const view:Record<string,number|string|undefined>=r;`, "seen as"},
		{`interface R {readonly named:number;[key:string]:number} const r:R={named:1}; const view:Record<string,number>=r;`, "seen as"},
		{`interface R {named:number;[key:string]:number} const fixed={named:1}; const r:R=fixed;`, "seen as"},
		{`interface R {named?:number;readonly [key:string]:number|undefined} const r:R={};`, "signature"},
		{`interface R {named:number;[key:number]:number} const r:R={named:1};`, "signature"},
	} {
		t.Run(probe.source, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want NotYet %s", err, probe.reason)
			}
		})
	}
}
