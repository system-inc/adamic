package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"strings"
	"testing"
)

func TestFactoryCompletionBeforeEscape(t *testing.T) {
	for _, probe := range []struct {
		name, body string
		proven     bool
	}{
		{"write", `node.value=7;return node;`, true},
		{"both branches", `if(flag){node.value=7;}else{node.value=8;}return node;`, true},
		{"one branch", `if(flag){node.value=7;}return node;`, false},
		{"alias", `const alias=node;alias.value=7;return node;`, true},
		{"passed", `take(node);node.value=7;return node;`, false},
		{"stored", `holder.node=node;node.value=7;return node;`, false},
		{"zero iterations", `while(flag){node.value=7;flag=false;}return node;`, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := `interface Node{kind:string;value:number;}interface Holder{node:Node|undefined;}const holder:Holder={node:undefined};function take(node:Node):void{}function create(flag:boolean):Node{const node={kind:'Node'} as Node;` + probe.body + `}const node=create(true);console.log(String(node.value));`
			program, err := lowerSource(t, source)
			if err != nil {
				t.Fatal(err)
			}
			proofs := FactoryCompletion(program)
			found := false
			for index, function := range program.Functions {
				if function.Name == "create" {
					found = true
					if proofs[index]["value"] != probe.proven {
						t.Fatalf("value proven=%t, want %t", proofs[index]["value"], probe.proven)
					}
				}
			}
			if !found {
				t.Fatal("missing factory")
			}
			checks := 0
			walk(program.Main, func(value any) bool {
				if property, ok := value.(ir.Property); ok && property.Name == "value" && property.Readiness != "" {
					checks++
				}
				return true
			})
			want := 1
			if probe.proven {
				want = 0
			}
			if checks != want {
				t.Fatalf("%d deferred checks, want %d", checks, want)
			}
		})
	}
}

func TestStagedFactoryRefusals(t *testing.T) {
	for _, probe := range []struct{ name, source, reason string }{
		{"incompatible initializer", `interface N {kind:number; value:number} function create(kind:number|string):N {const node={kind} as N;node.value=1;return node;}`, "incompatible with field kind"},
		{"optional field", `interface N {kind:string; value:number; extra?:number} function create():N {const node={kind:'N'} as N;node.value=1;return node;}`, "optional staged factory field extra"},
		{"presence observer", `interface N {kind:string; value:number} function create():N {const node={kind:'N'} as N;console.log(String(node.hasOwnProperty('value')));node.value=1;return node;}`, "observing staged factory field presence"},
		{"computed presence observer", `interface N {kind:string; value:number} function create():N {const node={kind:'N'} as N;console.log(String(node['hasOwnProperty']('value')));node.value=1;return node;}`, "observing staged factory field presence"},
		{"existing object", `interface N {kind:string; value:number} function create(base:{kind:string}):N {const node=base as N;node.value=1;return node;}`, "adamic/no-unchecked-cast"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %s, got %v", probe.reason, err)
			}
		})
	}
}

func TestParserFactoryCertificates(t *testing.T) {
	for _, probe := range []struct {
		file, function string
		fields         []string
		deferred       string
	}{
		{"01-binary-complete", "createBinaryExpression", []string{"kind", "left", "operatorToken", "right"}, ""},
		{"02-variable-complete", "createVariableDeclaration", []string{"kind", "name", "initializer"}, ""},
		{"03-numeric-conditional", "createNumericLiteral", []string{"kind", "text", "transformFlags"}, ""},
		{"deferred", "createSourceFile", []string{"kind"}, "bindDiagnostics"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			source, err := os.ReadFile("../oracle/testdata/parser_factory_" + probe.file + ".a")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSource(t, string(source))
			if err != nil {
				t.Fatal(err)
			}
			summaries := FactoryCompletion(program)
			found := false
			for i, function := range program.Functions {
				if function.Name != probe.function {
					continue
				}
				found = true
				for _, field := range probe.fields {
					if !summaries[i][field] {
						t.Errorf("%s not proven", field)
					}
				}
				if probe.deferred != "" && summaries[i][probe.deferred] {
					t.Error("deferred field incorrectly proven")
				}
			}
			if !found {
				t.Fatal("factory not found")
			}
		})
	}
}
