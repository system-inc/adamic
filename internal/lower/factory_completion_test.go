package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func TestParserFactoryCompletion(t *testing.T) {
	for _, test := range []struct {
		name, body      string
		proven, unknown bool
	}{
		{"before return", `const node = {kind:1,pos:-1,end:-1} as Target; node.text='ready'; return node;`, true, false},
		{"alias", `const node = {kind:1,pos:-1,end:-1} as Target; const alias=node; alias.text='ready'; return node;`, true, false},
		{"both branches", `const node = {kind:1,pos:-1,end:-1} as Target; if(flag){node.text='yes';}else{node.text='no';} return node;`, true, false},
		{"one branch", `const node = {kind:1,pos:-1,end:-1} as Target; if(flag){node.text='yes';} return node;`, false, false},
		{"before escape", `const node = {kind:1,pos:-1,end:-1} as Target; expose(node); node.text='late'; return node;`, false, true},
		{"wrong payload assertion", `const node = {kind:1,pos:-1,end:-1} as Target; node.text=42 as unknown as string; return node;`, false, false},
		{"early return", `const node = {kind:1,pos:-1,end:-1} as Target; if(flag){return node;} node.text='late'; return node;`, false, false},
		{"loop may not run", `const node = {kind:1,pos:-1,end:-1} as Target; while(flag){node.text='late';} return node;`, false, true},
		{"captured", `const node = {kind:1,pos:-1,end:-1} as Target; const callback=()=>node; node.text='late'; return node;`, false, true},
		{"short circuit", `const node = {kind:1,pos:-1,end:-1} as Target; flag && (node.text='late'); return node;`, false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "factory.a")
			source := `interface Target {kind:number;pos:number;end:number;text:string;} function expose(value:Target):void{} function factory(flag:boolean):Target {` + test.body + `}`
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			p, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			c, release := p.Checker(context.Background(), p.Files()[0])
			defer release()
			var factory *ast.Node
			p.Files()[0].AsNode().ForEachChild(func(n *ast.Node) bool {
				if n.Kind == ast.KindFunctionDeclaration && n.Name().Text() == "factory" {
					factory = n
				}
				return false
			})
			if factory == nil {
				t.Fatal("factory missing")
			}
			summary := AnalyzeFactory(factory, c.GetTypeAtLocation(factory.Type()), c)
			if summary.Proven("text") != test.proven || summary.Unknown != test.unknown {
				t.Fatalf("summary %+v, want text proven %v, unknown %v", summary, test.proven, test.unknown)
			}
			if !summary.Unknown && len(summary.Required) != 4 {
				t.Fatalf("required %+v", summary.Required)
			}
		})
	}
}

func TestParserFactoryReadBeforeCompletion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "read.a")
	source := `interface Target {text:string;} function factory():Target {const node={} as Target; const earlier=node.text; node.text='ready'; return node;}`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	c, release := p.Checker(context.Background(), p.Files()[0])
	defer release()
	var f *ast.Node
	p.Files()[0].AsNode().ForEachChild(func(n *ast.Node) bool {
		if n.Kind == ast.KindFunctionDeclaration {
			f = n
		}
		return false
	})
	summary := AnalyzeFactory(f, c.GetTypeAtLocation(f.Type()), c)
	if !summary.Proven("text") || len(summary.Reads) != 1 || !summary.Reads[0].Checked {
		t.Fatalf("later completion must not erase earlier read: %+v", summary)
	}
}
