package checker

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"testing"
)

func TestReactTypeNames(t *testing.T) {
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "input.ts")
	for path, text := range map[string]string{config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`, file: `type Dispatch<A>=(value:A)=>void;interface RefObject<T>{current:T} declare const setState:Dispatch<number>;declare const reference:RefObject<number>;declare function useEffect(cb:()=>void):void;setState(1);reference.current;useEffect(()=>{});`} {
		if e := os.WriteFile(path, []byte(text), 0600); e != nil {
			t.Fatal(e)
		}
	}
	p, e := Open(config, []string{file})
	if e != nil {
		t.Fatal(e)
	}
	seen := map[string]bool{}
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n.Kind == ast.KindIdentifier {
			if want, ok := map[string][2]string{"setState": {"Dispatch", "\ufffdtype"}, "reference": {"", "RefObject"}, "useEffect": {"", "useEffect"}}[n.Text()]; ok {
				wire, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "react-type-names")
				if e != nil {
					t.Fatal(e)
				}
				f := decodedFields(t, wire)
				if len(f) != 5 || f[2] != "1" || f[3] != want[0] || f[4] != want[1] {
					t.Fatalf("raw type names for %s: %q", n.Text(), f)
				}
				seen[n.Text()] = true
				if _, e := p.Inspect(file, uint64(n.Pos()), uint64(n.End()), "Identifier", "react-type-names\ninvalid"); e == nil {
					t.Fatal("malformed suffix accepted")
				}
			}
		}
		n.ForEachChild(walk)
		return false
	}
	p.Compiler.GetSourceFile(file).AsNode().ForEachChild(walk)
	if len(seen) != 3 {
		t.Fatal("missing raw type-name controls", seen)
	}
}
