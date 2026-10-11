package lower

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// The receivers the inherited-read sweep reads Object.prototype's members through, keyed by name.
// Each receiver is its own top-level test below, because Loom runs a top-level test as one atom and
// all twelve together cost about 22 s of one core; TestInheritedLibraryReadReceiversEachHaveOneTest
// holds the wrappers and this table to each other so a receiver cannot be added and never run.
var inheritedReadReceivers = map[string]string{
	"object":   "const value = { field: 1 };",
	"array":    "const value: number[] = [1];",
	"tuple":    "const value = [1, 'a'] as const;",
	"number":   "const value = 42;",
	"string":   "const value = 'a';",
	"boolean":  "const value = true;",
	"class":    "class Box { readonly field = 1; } const value = new Box();",
	"map":      "const value = new Map<string, number>();",
	"set":      "const value = new Set<number>();",
	"function": "const value = (): number => 1;",
	"error":    "const value = new Error('message');",
	"union":    "const value: number | string = true ? 42 : 'a';",
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsObject(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "object")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsArray(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "array")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsTuple(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "tuple")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsNumber(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "number")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsString(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "string")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsBoolean(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "boolean")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsClass(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "class")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsMap(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "map")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsSet(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "set")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsFunction(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "function")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsError(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "error")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsUnion(t *testing.T) {
	t.Parallel()
	sweepInheritedLibraryReads(t, "union")
}

// Every receiver has exactly one wrapper whose name is the receiver's, and every wrapper names a
// receiver in the table. Read from the source, so it builds nothing and loads no program.
func TestInheritedLibraryReadReceiversEachHaveOneTest(t *testing.T) {
	t.Parallel()
	const prefix = "TestInheritedLibraryReadsNeverLoadOwnFields"
	file, err := parser.ParseFile(token.NewFileSet(), "prototype_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	wrappers := make(map[string]int)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, prefix) {
			continue
		}
		// The wrapper's second statement must be the sweep of one named receiver, so the name read
		// here is the receiver that wrapper really runs.
		var receiver string
		if statements := function.Body.List; len(statements) == 2 {
			if expression, ok := statements[1].(*ast.ExprStmt); ok {
				if call, ok := expression.X.(*ast.CallExpr); ok && len(call.Args) == 2 {
					if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "sweepInheritedLibraryReads" {
						if literal, ok := call.Args[1].(*ast.BasicLit); ok && literal.Kind == token.STRING {
							receiver, _ = strconv.Unquote(literal.Value)
						}
					}
				}
			}
		}
		if receiver == "" {
			t.Errorf("%s: want t.Parallel() then sweepInheritedLibraryReads(t, \"<receiver>\") and nothing else", function.Name.Name)
			continue
		}
		if _, known := inheritedReadReceivers[receiver]; !known {
			t.Errorf("%s sweeps receiver %q, which inheritedReadReceivers does not have", function.Name.Name, receiver)
		}
		if want := prefix + strings.ToUpper(receiver[:1]) + receiver[1:]; function.Name.Name != want {
			t.Errorf("%s sweeps receiver %q; name it %s", function.Name.Name, receiver, want)
		}
		wrappers[receiver]++
	}
	for receiver := range inheritedReadReceivers {
		if wrappers[receiver] != 1 {
			t.Errorf("receiver %q has %d wrapper tests, want exactly 1", receiver, wrappers[receiver])
		}
	}
}

// Sweep the lowering entry points directly: the earlier unbound-method refusal must not mask a
// missing guard here. Each read has its own program so refusing the first cannot hide the rest.
func sweepInheritedLibraryReads(t *testing.T, receiver string) {
	t.Helper()
	source, known := inheritedReadReceivers[receiver]
	if !known {
		t.Fatalf("no receiver %q in inheritedReadReceivers", receiver)
	}
	members := nodePrototypeMembers(t)
	declared := map[string]bool{"constructor": true, "toString": true, "toLocaleString": true, "valueOf": true, "hasOwnProperty": true, "propertyIsEnumerable": true, "isPrototypeOf": true}
	for _, member := range members {
		for _, bracket := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/bracket=%t", member, bracket), func(t *testing.T) {
				t.Parallel()
				read := "value." + member
				if bracket {
					read = fmt.Sprintf("value['%s']", member)
				}
				path := filepath.Join(t.TempDir(), "sweep.a")
				if err := os.WriteFile(path, []byte(source+"\n"+read+";\n"), 0644); err != nil {
					t.Fatal(err)
				}
				program, err := load.Load([]string{path})
				if err != nil {
					var check *load.CheckError
					if !declared[member] && errors.As(err, &check) {
						return
					} // Legacy members absent from es2024.
					t.Fatal(err)
				}
				entry := program.Files()[0]
				checker, release := program.Checker(context.Background(), entry)
				defer release()
				lowering := &lowering{program: program, checker: checker, result: &ir.Program{}, this: -1, functionIndex: -1}
				if err := lowering.declareModule(entry.Statements.Nodes); err != nil {
					t.Fatal(err)
				}
				if _, err := lowering.statements(entry.Statements.Nodes[:len(entry.Statements.Nodes)-1]); err != nil {
					t.Fatal(err)
				}
				expression := entry.Statements.Nodes[len(entry.Statements.Nodes)-1].AsExpressionStatement().Expression
				var got ir.Expression
				if bracket {
					got, err = lowering.elementAccess(expression)
				} else {
					got, err = lowering.property(expression)
				}
				var refused *Refused
				if !errors.As(err, &refused) || got != nil {
					t.Fatalf("inherited %s read became %T (%#v), error %v; want a lowering refusal before any own-field load", member, got, got, err)
				}
			})
		}
	}
}

func TestPrototypeMethodsAreRefusedWithReasons(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const value = {}; value.isPrototypeOf({});",
		"const value = 1000; value.toLocaleString();",
		"const value = [1000]; value.toLocaleString();",
		"const value = { __proto__: { toString: (): string => 'custom' } }; value.toString();",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || refused.Fix == "" {
			t.Fatalf("got %v, want a reasoned refusal", err)
		}
	}
	for _, member := range []string{"toString", "toLocaleString", "valueOf", "hasOwnProperty", "propertyIsEnumerable", "isPrototypeOf", "constructor"} {
		_, err := lowerSource(t, "const value = {}; const { "+member+": detached } = value;")
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Errorf("destructured %s: %v", member, err)
		}
	}
}

// Null and undefined have no members to lower; the external checker refuses the access first.
func TestNullishPrototypeReadsAreRejectedByChecker(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"null", "undefined"} {
		for _, member := range nodePrototypeMembers(t) {
			path := filepath.Join(t.TempDir(), "nullish.a")
			if err := os.WriteFile(path, []byte("const value = "+value+"; value."+member+";"), 0644); err != nil {
				t.Fatal(err)
			}
			_, err := load.Load([]string{path})
			var check *load.CheckError
			if !errors.As(err, &check) {
				t.Errorf("%s.%s: %v", value, member, err)
			}
		}
	}
}

func TestPrototypeHazardsBehindObjectViewsAreNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const value: {} = /a/; console.log(value.toString());",
		"const value: {} = new RegExp('a'); console.log(value.toString());",
		"const value: {} = RegExp('a'); console.log(value.toLocaleString());",
		"function show(value: {}): string { return value.toString(); } const value = { toString: (): string => 'custom' }; console.log(show(value));",
		"function show(value: {}): string { return value.toLocaleString(); } const value = { toString: (): string => 'custom' }; console.log(show(value));",
		"function show(value: {}): string { return value.toString(); } console.log(show(new Error('message')));",
		"function show(value: {}): string { return value.toString(); } console.log(show(42));",
		"function show(value: {}): string { return value.toString(); } console.log(show([1]));",
		"class Box { declare field?: number; } const value = new Box(); value.hasOwnProperty('field');",
		"class Box { #secret = 1; read(): number { return this.#secret; } } const value = new Box(); console.log(`${value.hasOwnProperty('#secret')}`);",
		"const value = { 'a\\0b': 1 }; value.hasOwnProperty('a\\0b');",
		"function absent(): { readonly field: number; readonly label: string } | undefined { return undefined; } const value = { ...absent(), label: 'x' }; console.log(`${value.hasOwnProperty('field')}`);",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		want := "through an object view"
		if strings.Contains(source, "declare field") {
			// Declared fields are refused before prototype analysis in integration 13.
			want = "a declare or abstract class field"
		}
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, want) {
			t.Errorf("%s: got %v, want an explicit NotYet containing %q", source, err, want)
		}
	}
}

// Node decides the sweep's membership, including the legacy accessors omitted by TypeScript's
// declarations. Those must be rejected by the checker before lowering ever sees them.
func nodePrototypeMembers(t *testing.T) []string {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node is needed to enumerate Object.prototype")
	}
	output, err := exec.Command(node, "-e", "console.log(JSON.stringify(Object.getOwnPropertyNames(Object.prototype)))").Output()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(output, &names); err != nil {
		t.Fatal(err)
	}
	return names
}

func TestUnrepresentedPrototypeCallsAreNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const value = [1]; const kept = value.valueOf();",
		"const value = new Map<string, number>(); const kept = value.valueOf();",
		"const value = (): number => 1; const kept = value.valueOf();",
		"const value = (): number => 1; value.toString();",
		"const value = [[1]]; value.toString();",
		"function test(value: () => number): boolean { return value.hasOwnProperty('prototype'); } console.log(`${test(() => 1)}`);",
		"const value = [1, 'x'] as const; value.propertyIsEnumerable('0');",
		"const value = new Error('message'); value.toString();",
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Errorf("%s: %v", source, err)
		}
	}
}

func TestIsPrototypeOfReadsExplainThePrototypeRefusal(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"const value = {}; const detached = value.isPrototypeOf;",
		"const value = {}; const detached = value['isPrototypeOf'];",
	} {
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.Fix, "no observable prototype chain") {
			t.Errorf("got %v", err)
		}
	}
}
