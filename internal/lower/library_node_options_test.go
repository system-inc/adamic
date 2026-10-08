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
		{"multiple module const reused by stat helper", `import * as fs from 'node:fs'; const options={throwIfNoEntry:false} as const; function stat(path:string){return fs.statSync(path,options);} fs.statSync('x',options);`, 1, true},
		{"const assertion", `import {statSync} from 'node:fs'; const options=(({throwIfNoEntry:false}) as const); statSync('missing',(options));`, 1, true},
		{"private module consumer", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; function status(path:string):void { statSync(path,options); }`, 1, true},
		{"captured escape", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; const capture=()=>options; function status(path:string):void { statSync(path,options); }`, 1, false},
		{"module escape after function", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; function status(path:string):void { statSync(path,options); } const alias=options;`, 1, false},
		{"local captured consumer", `import {statSync} from 'node:fs'; function outer():void { const options={throwIfNoEntry:false} as const; function status(path:string):void { statSync(path,options); } }`, 1, true},
		{"multiple consumers", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; function first(path:string):void { statSync(path,options); } function second(path:string):void { statSync(path,options); }`, 1, true},
		{"multiple refused consumers", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; statSync('.',options); function status(path:string):void { statSync(path,options); console.log(options); }`, 1, false},
		{"exported binding", `import {rmSync} from 'node:fs'; export const options={recursive:true}; rmSync('missing',options);`, 1, false},
		{"exported alias", `import {rmSync} from 'node:fs'; const options={recursive:true}; export {options}; rmSync('missing',options);`, 1, false},
		{"captured extra reference mutant", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; function status(path:string):void { console.log(options); statSync(path,options); }`, 1, false},
		{"later escape", `import {rmSync} from 'node:fs'; const options={recursive:true}; rmSync('missing',options); console.log(options);`, 1, false},
		{"shorthand storage", `import {rmSync} from 'node:fs'; const options={recursive:true}; const stored={options}; rmSync('missing',options);`, 1, false},
		{"array storage", `import {rmSync} from 'node:fs'; const options={recursive:true}; const stored=[options]; rmSync('missing',options);`, 1, false},
		{"spread storage", `import {rmSync} from 'node:fs'; const options={recursive:true}; const stored={...options}; rmSync('missing',options);`, 1, false},
		{"loop field write", `import {rmSync} from 'node:fs'; const options={recursive:true}; for (options.recursive of [false]) {} rmSync('missing',options);`, 1, false},
		{"let mutant", `import {rmSync} from 'node:fs'; let options={recursive:true}; rmSync('missing',options);`, 1, false},
		{"cast hidden field mutant", `import {rmSync} from 'node:fs'; const options=({recursive:false,force:'yes'} as {recursive:boolean}); rmSync('missing',options);`, 1, false},
		{"cast planted field mutant", `import {rmSync} from 'node:fs'; const options={recursive:false}; (options as {recursive:boolean;force:string}).force='yes'; rmSync('missing',options);`, 1, false},
		{"field write", `import {rmSync} from 'node:fs'; const options={recursive:true}; options.recursive=false; rmSync('missing',options);`, 1, false},
		{"later field write", `import {rmSync} from 'node:fs'; const options={recursive:true}; rmSync('missing',options); options.recursive=false;`, 1, false},
		{"capture", `import {rmSync} from 'node:fs'; const options={recursive:true}; const capture=()=>options; rmSync('missing',options);`, 1, false},
		{"scalar read", `import {rmSync} from 'node:fs'; const options={recursive:true}; const enabled=options.recursive; rmSync('missing',options);`, 1, false},
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
			wantCalls := 1
			if strings.HasPrefix(probe.name, "multiple ") {
				wantCalls = 2
			}
			if count != wantCalls {
				t.Fatalf("checked %d calls, want %d", count, wantCalls)
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
		{"captured extra reference", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; function status(path:string):void { console.log(options); statSync(path,options); }`},
		{"later escape refuses all consumers", `import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; statSync('.',options); function status(path:string):void { statSync(path,options); console.log(options); }`},
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
		`import {statSync} from 'node:fs'; const options={throwIfNoEntry:false} as const; statSync('.',options); statSync('missing',options);`,
		`import {statSync,Stats} from 'node:fs'; const options={throwIfNoEntry:false} as const; function first(path:string):Stats|undefined { return statSync(path,options); } function second(path:string):Stats|undefined { return statSync(path,options); }`,
		`import {statSync,Stats} from 'node:fs'; function outer():void { const options={throwIfNoEntry:false} as const; const status=(path:string):Stats|undefined => statSync(path,options); status('.'); } outer();`,
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

// Importing a consuming closure preserves a private literal; exporting the
// literal itself exposes it to references in another source file and refuses.
func TestNodeHostOptionsReferencesAcrossFiles(t *testing.T) {
	for _, escape := range []bool{false, true} {
		directory := t.TempDir()
		files := []string{filepath.Join(directory, "options.a"), filepath.Join(directory, "consumer.a"), filepath.Join(directory, "other.a")}
		other := "import {status} from './options.a'; status('.');"
		optionsSource := "import {statSync,Stats} from 'node:fs'; const options={throwIfNoEntry:false} as const; export function status(path:string):Stats|undefined { return statSync(path,options); }"
		if escape {
			other = "import {options} from './options.a'; console.log(options);"
			optionsSource += " export {options};"
		}
		for index, source := range []string{
			optionsSource,
			"import {status} from './options.a'; status('.');",
			other,
		} {
			if err := os.WriteFile(files[index], []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
		}
		program, err := load.Load(files)
		if err != nil {
			t.Fatal(err)
		}
		checked, release := program.Checker(context.Background(), program.Files()[0])
		analysis := lowering{program: program, checker: checked}
		count := 0
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression && analysis.optionalNodeHostCall(node) {
				count++
				if got := analysis.nodeHostConsumesArgument(node, 1); got != !escape {
					t.Errorf("cross-file escape=%t: consumes argument=%t, want %t", escape, got, !escape)
				}
			}
			node.ForEachChild(visit)
			return false
		}
		program.Files()[0].AsNode().ForEachChild(visit)
		release()
		if count != 1 {
			t.Fatalf("checked %d calls, want one", count)
		}
	}
}
