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
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/toolchain"
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

func TestInheritedLibraryReadsNeverLoadOwnFieldsObject(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "object")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsArray(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "array")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsTuple(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "tuple")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsNumber(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "number")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsString(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "string")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsBoolean(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "boolean")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsClass(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "class")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsMap(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "map")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsSet(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "set")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsFunction(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "function")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsError(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "error")
}

func TestInheritedLibraryReadsNeverLoadOwnFieldsUnion(test *testing.T) {
	test.Parallel()
	sweepInheritedLibraryReads(test, "union")
}

// Every receiver has exactly one wrapper whose name is the receiver's, and every wrapper names a
// receiver in the table. Read from the source, so it builds nothing and loads no program.
func TestInheritedLibraryReadReceiversEachHaveOneTest(test *testing.T) {
	test.Parallel()
	const prefix = "TestInheritedLibraryReadsNeverLoadOwnFields"
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "prototype_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	wrappers := make(map[string]int)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, prefix) {
			continue
		}
		// The wrapper's first statement must be test.Parallel() and its second the sweep of one named
		// receiver, so the name read here is the receiver that wrapper really runs.
		var receiver string
		if statements := function.Body.List; len(statements) == 2 && isParallelCall(statements[0]) {
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
			test.Errorf(
				"%s: want t.Parallel() then sweepInheritedLibraryReads(t, \"<receiver>\") and nothing else",
				function.Name.Name,
			)
			continue
		}
		if _, known := inheritedReadReceivers[receiver]; !known {
			test.Errorf(
				"%s sweeps receiver %q, which inheritedReadReceivers does not have",
				function.Name.Name,
				receiver,
			)
		}
		if want := prefix + strings.ToUpper(receiver[:1]) + receiver[1:]; function.Name.Name != want {
			test.Errorf("%s sweeps receiver %q; name it %s", function.Name.Name, receiver, want)
		}
		wrappers[receiver]++
	}
	for receiver := range inheritedReadReceivers {
		if wrappers[receiver] != 1 {
			test.Errorf("receiver %q has %d wrapper tests, want exactly 1", receiver, wrappers[receiver])
		}
	}
}

// isParallelCall is whether statement is t.Parallel().
func isParallelCall(statement ast.Stmt) bool {
	expression, ok := statement.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := expression.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Parallel" {
		return false
	}
	receiver, ok := selector.X.(*ast.Ident)
	return ok && receiver.Name == "t"
}

// Sweep the lowering entry points directly: the earlier unbound-method refusal must not mask a
// missing guard here. Each read has its own program so refusing the first cannot hide the rest.
func sweepInheritedLibraryReads(test *testing.T, receiver string) {
	test.Helper()
	source, known := inheritedReadReceivers[receiver]
	if !known {
		test.Fatalf("no receiver %q in inheritedReadReceivers", receiver)
	}
	members := nodePrototypeMembers(test)
	declared := map[string]bool{
		"constructor":          true,
		"toString":             true,
		"toLocaleString":       true,
		"valueOf":              true,
		"hasOwnProperty":       true,
		"propertyIsEnumerable": true,
		"isPrototypeOf":        true,
	}
	for _, member := range members {
		for _, bracket := range []bool{false, true} {
			test.Run(fmt.Sprintf("%s/bracket=%t", member, bracket), func(test *testing.T) {
				test.Parallel()
				read := "value." + member
				if bracket {
					read = fmt.Sprintf("value['%s']", member)
				}
				path := filepath.Join(test.TempDir(), "sweep.a")
				if writeError := os.WriteFile(path, []byte(source+"\n"+read+";\n"), 0644); writeError != nil {
					test.Fatal(writeError)
				}
				program, err := load.Load([]string{path})
				if err != nil {
					var check *load.CheckError
					if !declared[member] && errors.As(err, &check) {
						return
					} // Legacy members absent from es2024.
					test.Fatal(err)
				}
				entry := program.Files()[0]
				checker, release := program.Checker(context.Background(), entry)
				defer release()
				lowering := &lowering{
					program:       program,
					checker:       checker,
					result:        &ir.Program{},
					this:          -1,
					functionIndex: -1,
				}
				if declareModuleError := lowering.declareModule(entry.Statements.Nodes); declareModuleError != nil {
					test.Fatal(declareModuleError)
				}
				if _, statementsError := lowering.statements(
					entry.Statements.Nodes[:len(entry.Statements.Nodes)-1],
				); statementsError != nil {
					test.Fatal(statementsError)
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
					test.Fatalf(
						"inherited %s read became %T (%#v), error %v; want a lowering refusal before any own-field load",
						member,
						got,
						got,
						err,
					)
				}
			})
		}
	}
}

func TestPrototypeMethodsAreRefusedWithReasons(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"const value = {}; value.isPrototypeOf({});",
		"const value = 1000; value.toLocaleString();",
		"const value = [1000]; value.toLocaleString();",
		"const value = { __proto__: { toString: (): string => 'custom' } }; value.toString();",
	} {
		_, lowerSourceError := lowerSource(test, source)
		var refused *Refused
		if !errors.As(lowerSourceError, &refused) || refused.Fix == "" {
			test.Fatalf("got %v, want a reasoned refusal", lowerSourceError)
		}
	}
	for _, member := range []string{
		"toString",
		"toLocaleString",
		"valueOf",
		"hasOwnProperty",
		"propertyIsEnumerable",
		"isPrototypeOf",
		"constructor",
	} {
		_, lowerSourceError := lowerSource(test, "const value = {}; const { "+member+": detached } = value;")
		var refused *Refused
		if !errors.As(lowerSourceError, &refused) {
			test.Errorf("destructured %s: %v", member, lowerSourceError)
		}
	}
}

// Null and undefined have no members to lower; the external checker refuses the access first.
func TestNullishPrototypeReadsAreRejectedByChecker(test *testing.T) {
	test.Parallel()
	for _, value := range []string{"null", "undefined"} {
		for _, member := range nodePrototypeMembers(test) {
			path := filepath.Join(test.TempDir(), "nullish.a")
			if writeError := os.WriteFile(
				path,
				[]byte("const value = "+value+"; value."+member+";"),
				0644,
			); writeError != nil {
				test.Fatal(writeError)
			}
			_, loadError := load.Load([]string{path})
			var check *load.CheckError
			if !errors.As(loadError, &check) {
				test.Errorf("%s.%s: %v", value, member, loadError)
			}
		}
	}
}

func TestPrototypeHazardsBehindObjectViewsAreNotYet(test *testing.T) {
	test.Parallel()
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
		_, lowerSourceError := lowerSource(test, source)
		var notYet *NotYet
		want := "through an object view"
		if strings.Contains(source, "declare field") {
			// Declared fields are refused before prototype analysis in integration 13.
			want = "a declare or abstract class field"
		}
		if !errors.As(lowerSourceError, &notYet) || !strings.Contains(notYet.What, want) {
			test.Errorf("%s: got %v, want an explicit NotYet containing %q", source, lowerSourceError, want)
		}
	}
}

// Node decides the sweep's membership, including the legacy accessors omitted by TypeScript's
// declarations. Those must be rejected by the checker before lowering ever sees them. Node is the
// toolchain's (#m27xj48); a sandbox without it fails instead of skipping, so the sweep never passes
// having swept nothing.
func nodePrototypeMembers(test *testing.T) []string {
	test.Helper()
	node, pathError := toolchain.Path("node")
	if pathError != nil {
		if toolchain.Sandboxed() {
			test.Fatal(pathError)
		}
		test.Skip("Node is needed to enumerate Object.prototype")
	}
	output, outputError := toolchain.Command(
		node,
		"-e",
		"console.log(JSON.stringify(Object.getOwnPropertyNames(Object.prototype)))",
	).Output()
	if outputError != nil {
		test.Fatal(outputError)
	}
	var names []string
	if decodeError := json.Unmarshal(output, &names); decodeError != nil {
		test.Fatal(decodeError)
	}
	return names
}

func TestUnrepresentedPrototypeCallsAreNotYet(test *testing.T) {
	test.Parallel()
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
		_, lowerSourceError := lowerSource(test, source)
		var notYet *NotYet
		if !errors.As(lowerSourceError, &notYet) {
			test.Errorf("%s: %v", source, lowerSourceError)
		}
	}
}

func TestIsPrototypeOfReadsExplainThePrototypeRefusal(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"const value = {}; const detached = value.isPrototypeOf;",
		"const value = {}; const detached = value['isPrototypeOf'];",
	} {
		_, lowerSourceError := lowerSource(test, source)
		var refused *Refused
		if !errors.As(lowerSourceError, &refused) || !strings.Contains(refused.Fix, "no observable prototype chain") {
			test.Errorf("got %v", lowerSourceError)
		}
	}
}
