package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRecordForms(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`interface R { [key:string]:number; named:number } const r:R={named:1}; console.log(String(r.named));`,
		`interface R { [key:string]: number } const r:R={}; r['x']=1; console.log(String(r['x']));`,
		`type R=Record<string,string>; const r:R={x:'one'}; console.log(Object.keys(r).join(','));`,
		`function show(r:Record<string,number>|undefined):void {console.log(String(JSON.stringify(r)));} show(undefined); show({x:1});`,
		`interface MapLike<T> { [key:string]:T } const r:MapLike<number>={}; console.log(String(delete r['absent']));`,
		`const r:Record<string,number>={}; for(const k in r) { console.log(k); }`,
		`interface Node {links:Record<string,Node>} const r:Record<string,Node>={}; r['fresh']={links:{}};`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestRecordRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		source, reason string
		notYet         bool
	}{
		{`import type {Weak} from 'adamic'; interface Item {name:string} const r:Record<string,Weak<Item>>={};`, "slot representation", true},
		{`function show(r:Record<string,number>|null):void {console.log(String(JSON.stringify(r)));}`, "value of type", true},
		{`function read(r:Record<string,string>|undefined,k:string):string {return r && r[k] || "fallback";}`, "logical record operand", true},
		{`interface R { readonly [key:string]:number } const r:R={};`, "signature", true},
		{`interface R { [key:number]:number } const r:R={};`, "signature", true},
		{`const r:Record<string,()=>string>={}; console.log(r.toString());`, "toString", false},
		{`const r:Record<string,number>={}; console.log(String(r['toString']));`, "toString", false},
		{`const r:Record<string,number>={}; console.log(String('constructor' in r));`, "constructor", false},
		{`interface Animal {name:string} interface Dog {name:'dog'} const dogs:Record<string,Dog>={}; const copy:Record<string,Animal>={...dogs};`, "invariant-mutable", false},
		{`interface Animal {name:string} interface Dog extends Animal {bark:string} const dogs:Record<string,Dog>={}; const animals:Record<string,Animal>=dogs;`, "invariant-mutable", false},
		{`function view(r:Record<string,number>|{x:number}):{x?:number} {return r;} const r:Record<string,number>={}; console.log(String(view(r).x));`, "seen as", true},
		{`const records:Record<string,number>={}; const view:{x?:number}=records;`, "seen as", true},
		{`const inner:Record<string,number>={}; const holder={inner}; const view:{inner:{x?:number}}=holder;`, "seen as", true},
		{`type R=Readonly<Record<string,number>>; const r:R={}; console.log(String(r['x']));`, "value of type", true},
		{`const fixed={x:1}; const r:Record<string,number>=fixed;`, "seen as", true},
		{`interface R {[key:string]:R} const r:R={}; r['self']=r;`, "cycle", false},
		{`interface Node {links:Record<string,Node>} const a:Node={links:{}}; a.links['self']=a;`, "cycle", false},
	} {
		t.Run(probe.source, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var ny *NotYet
			var refused *Refused
			if err == nil || !strings.Contains(err.Error(), probe.reason) || (probe.notYet && !errors.As(err, &ny)) || (!probe.notYet && !errors.As(err, &refused)) {
				t.Fatalf("got %v, want %q (NotYet=%t)", err, probe.reason, probe.notYet)
			}
		})
	}
}

func TestRecordPrototypeLiteralNames(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"constructor", "__defineGetter__", "__defineSetter__", "hasOwnProperty", "__lookupGetter__", "__lookupSetter__", "isPrototypeOf", "propertyIsEnumerable", "toString", "valueOf", "__proto__", "toLocaleString"} {
		for _, source := range []string{`const r:Record<string,number>={}; console.log(String(r['` + name + `']));`, `const r:Record<string,number>={}; console.log(String('` + name + `' in r));`} {
			t.Run(source, func(t *testing.T) {
				_, err := lowerSource(t, source)
				var refusal *Refused
				if !errors.As(err, &refusal) || !strings.Contains(err.Error(), name) {
					t.Fatalf("got %v, want literal member refusal naming %s", err, name)
				}
			})
		}
	}
}
