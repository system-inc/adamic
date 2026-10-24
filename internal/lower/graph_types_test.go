package lower

import (
	"context"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGraphLiteralMethodAllocation(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function make():void {let holder:{read():number}|undefined;const value={read():number{return holder===undefined?0:1;}};holder=value;console.log('made');}make();`)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("graph types %v", program.GraphTypes)
	found := false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if literal, ok := node.(ir.ObjectLiteral); ok {
				found = true
				if !program.IsGraph(literal.GraphTypes) {
					t.Errorf("cyclic method holder allocation is counted: types %v", literal.GraphTypes)
				}
			}
			return true
		})
	}
	if !found {
		t.Fatal("missing method holder allocation")
	}
}

func TestGraphSeedsOnlyTypedCycleRefusals(t *testing.T) {
	t.Parallel()
	other := errors.New("slot classification failed")
	var absent *Refused
	for _, err := range []error{nil, absent, other, fmt.Errorf("wrapped: %w", other), &NotYet{What: "unknown slot"}, &Refused{What: "not a cycle"}, &Refused{What: "a cycle reference counting can't free"}} {
		seed, got := graphSeed(err)
		if seed || got != err {
			t.Errorf("graphSeed(%v) = (%t, %v), want unchanged non-seed", err, seed, got)
		}
	}
	refusal := &Refused{cycle: true, What: "slot cycle"}
	for _, err := range []error{refusal, fmt.Errorf("wrapped: %w", refusal)} {
		if seed, got := graphSeed(err); !seed || got != nil {
			t.Errorf("cycle proof was not a seed: %t, %v", seed, got)
		}
	}
}

func TestGraphAdoptionBoundaryNamesOffendingNode(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, marker, what string }{
		{"checked view", interfaceCastSource + `console.log(identifier({kind:'identifier',pos:0,name:'ok'} as Identifier).name);`, "node as Identifier", "a checked-view adapter in a graph program"},
		{"legacy checked view", `type First={readonly kind:'a';readonly name:string};
type Second={readonly kind:'b';readonly name:string};
type Third={readonly kind:'c';readonly name:string};
function narrow(value:First|Second|Third):First|Second{return value as First|Second;}
console.log(narrow({kind:'a',name:'kept'}).name);`, "value as First|Second", "a checked-view adapter in a graph program"},
		{"dynamic literal", `const source: {visible:number;[key:string]:number|string} = {visible:1};
source['late']=2;
console.log(Object.keys(source).join(','));`, "{visible:1}", "a dynamic-shape object in a graph program"},
		{"dynamic property", `function code(value:unknown):string|undefined {
 return typeof value==='object' && value!==null && 'code' in value && typeof value.code==='string' ? value.code : undefined;
}
console.log(String(code({code:'ok'})));`, "value.code", "a dynamic-shape object in a graph program"},
		{"aliased enumeration", `const enumerate=Object.entries;
interface Visible {readonly visible:number}
function show(view:Visible):void {
 const entries:[string,number][]=enumerate(view);
 console.log(String(entries.length));
}
show({visible:1});`, "enumerate(view)", "a dynamic-shape object in a graph program"},
		{"checked enumeration", `interface Visible {readonly visible:number}
function show(view:Visible):void {
 const entries:[string,number][]=Object.entries(view);
 console.log(String(entries.length));
}
show({visible:1});`, "Object.entries(view)", "a dynamic-shape object in a graph program"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			main := filepath.Join(directory, "main.a")
			cause := filepath.Join(directory, "cause.a")
			source := "// The offending node is in an imported module.\n\n" + probe.source + "\nexport {};\n"
			if err := os.WriteFile(cause, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(main, []byte("import './cause.a';\n"+graphReviewCycle), 0600); err != nil {
				t.Fatal(err)
			}
			checked, err := load.Load([]string{main})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), checked)
			var refusal *Refused
			if !errors.As(err, &refusal) {
				t.Fatalf("want graph boundary refusal, got %v", err)
			}
			offset := strings.LastIndex(source, probe.marker)
			if offset < 0 {
				t.Fatal("missing source marker")
			}
			before := source[:offset]
			line := strings.Count(before, "\n") + 1
			column := offset - strings.LastIndex(before, "\n")
			where := fmt.Sprintf("%s:%d:%d", cause, line, column)
			if refusal.Where != where || refusal.What != probe.what || refusal.Fix != "keep checked views and dynamic shapes outside graph adoption (adamic/cycle-capable)" {
				t.Fatalf("got %#v, want %s with unchanged graph diagnostic", refusal, where)
			}
		})
	}
}
