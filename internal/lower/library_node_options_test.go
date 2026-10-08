package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// Probe the predicate itself as well as the relation hook: a second guard at the
// hook must not conceal an unsound classification in the library's proof.
func TestNodeHostConsumesArgument(t *testing.T) {
	for _, probe := range []struct {
		name, source string
		index        int
		want         bool
	}{
		{"literal", `import {rmSync as remove} from 'node:fs'; remove('missing',{recursive:true});`, 1, true},
		{"exact const", `import {rmSync} from 'node:fs'; const options={recursive:true}; rmSync('missing',options);`, 1, true},
		{"const assertion", `import {statSync} from 'node:fs'; const options=(({throwIfNoEntry:false}) as const); statSync('missing',(options));`, 1, true},
		{"private module consumer", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; function status(path:string):void { statSync(path,options); }`, 1, true},
		{"captured escape", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; const capture=()=>options; function status(path:string):void { statSync(path,options); }`, 1, false},
		{"module escape after function", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; function status(path:string):void { statSync(path,options); } const alias=options;`, 1, false},
		{"local captured consumer", `import {statSync} from 'node:fs'; function outer():void { const options={throwIfNoEntry:false} as const; function status(path:string):void { statSync(path,options); } }`, 1, false},
		{"loop field write", `import {rmSync} from 'node:fs'; const options={recursive:true}; for (options.recursive of [false]) {} rmSync('missing',options);`, 1, false},
		{"let mutant", `import {rmSync} from 'node:fs'; let options={recursive:true}; rmSync('missing',options);`, 1, false},
		{"cast hidden field mutant", `import {rmSync} from 'node:fs'; const options=({recursive:false,force:'yes'} as {recursive:boolean}); rmSync('missing',options);`, 1, false},
		{"cast planted field mutant", `import {rmSync} from 'node:fs'; const options={recursive:false}; (options as {recursive:boolean;force:string}).force='yes'; rmSync('missing',options);`, 1, false},
		{"field write", `import {rmSync} from 'node:fs'; const options={recursive:true}; options.recursive=false; rmSync('missing',options);`, 1, false},
		{"later field write", `import {rmSync} from 'node:fs'; const options={recursive:true}; rmSync('missing',options); options.recursive=false;`, 1, false},
		{"capture", `import {rmSync} from 'node:fs'; const options={recursive:true}; const capture=()=>options; rmSync('missing',options);`, 1, false},
		{"scalar read", `import {rmSync} from 'node:fs'; const options={recursive:true}; const enabled=options.recursive; rmSync('missing',options);`, 1, true},
		{"const alias", `import {rmSync} from 'node:fs'; const original={recursive:true}; const options=original; rmSync('missing',options);`, 1, false},
		{"prior escape mutant", `import {rmSync} from 'node:fs'; function escape(x:{recursive:boolean}):void {} const options={recursive:true}; escape(options); rmSync('missing',options);`, 1, false},
		{"returned object mutant", `import {rmSync} from 'node:fs'; function make():{recursive:boolean} { return {recursive:true}; } const options=make(); rmSync('missing',options);`, 1, false},
		{"structural subtype", `import {rmSync} from 'node:fs'; const actual={recursive:true,force:'yes'}; const options:{recursive:boolean}=actual; rmSync('missing',options);`, 1, false},
		{"non-node", `function rmSync(path:string, options:{recursive?:boolean}):void {} rmSync('missing',{recursive:true});`, 1, false},
		{"forwarded object", `import {readdirSync} from 'node:fs'; readdirSync('.',{withFileTypes:true});`, 1, false},
		{"nested object", `import {rmSync} from 'node:fs'; const options={recursive:true,other:{value:true}}; rmSync('missing',options);`, 1, false},
		{"spread", `import {rmSync} from 'node:fs'; const options={recursive:true}; rmSync('missing',{...options});`, 1, false},
		{"path argument", `import {rmSync} from 'node:fs'; rmSync('missing',{recursive:true});`, 0, false},
		{"negative index", `import {rmSync} from 'node:fs'; rmSync('missing',{recursive:true});`, -1, false},
		{"missing argument", `import {rmSync} from 'node:fs'; rmSync('missing');`, 1, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(file, []byte(probe.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{file})
			if err != nil {
				t.Fatal(err)
			}
			checked, release := program.Checker(context.Background(), program.Files()[0])
			defer release()
			analysis := lowering{program: program, checker: checked}
			count := 0
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression && (analysis.optionalNodeHostCall(node) || probe.name == "non-node") {
					count++
					if got := analysis.nodeHostConsumesArgument(node, probe.index); got != probe.want {
						t.Errorf("consumes argument = %t, want %t", got, probe.want)
					}
				}
				node.ForEachChild(visit)
				return false
			}
			program.Files()[0].AsNode().ForEachChild(visit)
			if count != 1 {
				t.Fatalf("checked %d calls, want one", count)
			}
		})
	}
}

func TestNodeHostOptionsRefusalBoundaries(t *testing.T) {
	hidden, err := os.ReadFile("testdata/node_options_widening/hidden_force.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, source string }{
		{"let binding", `import {rmSync} from 'node:fs'; let options={recursive:true}; rmSync('missing',options);`},
		{"prior escape", `import {rmSync} from 'node:fs'; function escape(x:{recursive:boolean}):void {} const options={recursive:true}; escape(options); rmSync('missing',options);`},
		{"function return", `import {rmSync} from 'node:fs'; function make():{recursive:boolean} { return {recursive:true}; } const options=make(); rmSync('missing',options);`},
		{"hidden field cast", `import {rmSync} from 'node:fs'; const options=({recursive:false,force:'yes'} as {recursive:boolean}); rmSync('missing',options);`},
		{"hidden force", string(hidden)},
		{"ordinary call", `const actual={recursive:true,force:'yes'}; const options:{recursive:boolean}=actual; function consume(options:{readonly recursive:boolean; readonly force?:boolean}):void {} consume(options);`},
		{"forwarded host object", `import {readdirSync} from 'node:fs'; const options={withFileTypes:true as const}; readdirSync('.',options);`},
		{"stored wider view", `import {rmSync} from 'node:fs'; const options={recursive:true}; const stored:{readonly recursive:boolean;readonly force?:boolean}=options; rmSync('missing',options);`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "adamic/no-optional-widening") {
				t.Fatalf("want optional widening refusal, got %v", err)
			}
		})
	}
}

func TestNodeHostOptionsConsumptionLowers(t *testing.T) {
	for _, source := range []string{
		`import * as fs from 'node:fs'; const options=(({throwIfNoEntry:false}) as const); fs.statSync('x',options);`,
		`import * as fs from 'node:fs'; const options={throwIfNoEntry:false} as const; function status(path:string):import('node:fs').Stats|undefined { return fs.statSync(path,options); }`,
		`import {statSync as status} from 'node:fs'; const options={throwIfNoEntry:false}; const stat=status('.',options);`,
		`import * as fs from 'node:fs'; const options={recursive:true,force:true}; fs.rmSync('missing',options);`,
		`import {writeFileSync} from 'node:fs'; const options={flag:'w',mode:384,flush:false}; writeFileSync('missing','x',options);`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
