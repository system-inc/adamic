package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNamedRecordReadTypes(t *testing.T) {
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

func TestNamedRecordRefusals(t *testing.T) {
	for _, probe := range []struct{ source, reason string }{
		{`interface R {named?:number[];[key:string]:number[]|string|undefined} function read(r:R):number|undefined {return r.named?.length;}`, "dictionary member narrowed to an object kind"},
		{`interface R {named?:number;[key:string]:number|string|boolean|null|undefined} const r:R={};`, "slot representation"},
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
