package lower

import (
	"context"
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const enumViewTypes = `enum AKind { A } enum BKind { B = 1 }
interface Wide { readonly kind: AKind; readonly child: { readonly ready: number }; readonly brand: number; readonly marker: number; readonly multiLine: number; }
interface Narrow { readonly kind: BKind; readonly child: { readonly ready: boolean }; readonly brand: undefined; readonly marker?: undefined; readonly multiLine?: boolean; }`

func TestEnumTagLazyPayloadAdmission(t *testing.T) {
	for _, probe := range []struct{ name, body string }{
		{"unread", `console.log('narrowed');`},
		{"undefined", `console.log(String(v.brand === undefined));`},
		{"optional undefined", `console.log(String(v.marker === undefined));`},
		{"structured", `const child = v.child; console.log(String(child.ready));`},
		{"optional boolean", `console.log(String(v.multiLine));`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, enumViewTypes+`function show(v: Wide | Narrow): void { if (v.kind === BKind.B) { `+probe.body+` } }`)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEnumTagMissingViewFamilyAtRead(t *testing.T) {
	source := `enum AKind { A } enum BKind { B = 1 } interface Wide { readonly kind: AKind; readonly opaque: string } interface Narrow { readonly kind: BKind; readonly opaque: never } function show(v: Wide | Narrow): void { if (v.kind === BKind.B) { console.log('narrowed'); READ } }`
	if _, err := lowerSource(t, strings.Replace(source, "READ", "", 1)); err != nil {
		t.Fatal(err)
	}
	_, err := lowerSource(t, strings.Replace(source, "READ", "const value = v.opaque;", 1))
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "field opaque") || !strings.Contains(refused.What, "never") {
		t.Fatalf("want named never refusal at read, got %v", err)
	}
}

func TestEnumTagComputedBindingReadIsNamed(t *testing.T) {
	source := enumViewTypes + `function show(v: Wide | Narrow): void { if (v.kind === BKind.B) { const { ['brand']: value } = v; console.log(String(value === undefined)); } }`
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "destructuring") {
		t.Fatalf("want named computed binding read refusal, got %v", err)
	}
}

func TestEnumTagCallableInventoryNamesDoNotPanic(t *testing.T) {
	for _, binding := range []string{`const { value } = { value: 1 };`, `const object = { ['other']: 1 };`} {
		source := binding + enumViewTypes + `interface CallableNarrow { readonly kind: BKind; readonly child: { readonly ready: boolean; readonly inspect: (n: number) => number } } function show(v: Wide | CallableNarrow): void { if (v.kind === BKind.B) { console.log('narrowed'); } }`
		path := filepath.Join(t.TempDir(), "main.a")
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		checker, release := program.Checker(context.Background(), program.Files()[0])
		var target *ast.Node
		for _, node := range program.Files()[0].Statements.Nodes {
			if node.Kind == ast.KindInterfaceDeclaration && node.Name().Text() == "CallableNarrow" {
				target = node
			}
		}
		if target == nil {
			t.Fatal("missing target interface")
		}
		audit := &lowering{checker: checker, program: program, result: &ir.Program{}}
		_, err = audit.viewSchema(target, checker.GetTypeAtLocation(target.Name()))
		release()
		if err != nil {
			t.Fatal(err)
		}

	}
}

func TestEnumTagUnsupportedArrayConsumerIsNamed(t *testing.T) {
	source := `enum AKind { A } enum BKind { B = 1 }
interface Wide { readonly kind: AKind; readonly values: readonly number[][] }
interface Narrow { readonly kind: BKind; readonly values: readonly boolean[][] }
function show(v: Wide | Narrow): void {
 if (v.kind === BKind.B) {
  console.log('narrowed');
  console.log(v.values.join(','));
 }
}`
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "[element]") || !strings.Contains(refused.What, "nested array .join") || !strings.Contains(refused.Where, ":7:") {
		t.Fatalf("want named array consumer refusal on read line 7, got %v", err)
	}
}
