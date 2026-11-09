package lower

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/load"
)

// A presence test of a method declared on a program's interface folds to true where hostOrigin proves
// the receiver is a registered host object by how it was made, and the model's registration names the
// member as a method (method_presence.go, #q9nfg88). Every road the walk doesn't prove keeps
// methodPresenceUnproven, in today's words, at the test. These pins lower against the performance row
// routed, as host_binding_route_test.go's do.
//
// Mutants, each one line in method_presence.go, and the rows that go red:
//   - (q1) the escape check dropped, in hostOriginBinding:
//     `if len(path) > 0 && !l.hostOriginContained(declaration, symbol) {` becomes
//     `if false && len(path) > 0 && !l.hostOriginContained(declaration, symbol) {`.
//     Compiler's write row, the alias row and the argument row fold.
//   - (q2) accessors allowed, in hostOriginProperty:
//     `case ast.KindSpreadAssignment, ast.KindGetAccessor, ast.KindSetAccessor:` becomes
//     `case ast.KindSpreadAssignment:`. Language's getter row folds.
//   - (q3) step 6 answering host: hostOrigin's last line, the `return nil` after its switch, becomes
//     `return l.hostBindingModel(*hostBindingNamed("performance"))`. Every negative row folds, and the
//     generator probe answers host. The timeOrigin probe holds, since the member check is its own.
//
// Two of Language's roads have no row, since tsc refuses them before Adamic sees them: a property written
// twice in one literal (`{ performance, performance: own }`, a duplicate tsc reports), and reading a
// property off a generator's result (`f().performance` on a Generator, which has none). The walk refuses
// both all the same (hostOriginProperty counts the writes, hostOriginFunction refuses a generator), and
// the generator probe asks the walk itself.

// hostOriginReturnType is tryGetPerformance's declared result as adaptation 99 writes it.
const hostOriginReturnType = "{ shouldWriteNativeEvents: boolean; performance: Performance; } | undefined"

// hostOriginOwn is a program object typed Performance, which no road makes the host's.
const hostOriginOwn = "const own: Performance = { timeOrigin: 0, now: (): number => 5, mark(name: string): void {}, measure(name: string, startMark?: string, endMark?: string): void {}, clearMarks(name?: string): void {}, clearMeasures(name?: string): void {} };\n"

// hostOriginGetter is tscPerformanceGetter with its literal's performance property written as value.
func hostOriginGetter(value string) string {
	return "function tryGetPerformance(): " + hostOriginReturnType + " {\n    if (typeof performance === \"object\") {\n        return {\n            shouldWriteNativeEvents: true,\n            performance: " + value + ",\n        };\n    }\n\n    return undefined;\n}\n"
}

// hostOriginTested is how every row ends: the presence test on what it destructured.
const hostOriginTested = "    return typeof performance.now === \"function\" ? \"yes\" : \"no\";\n}\n"

// hostOriginPresent is a present() that takes call's result apart as tryGetPerformanceHooks takes
// tryGetPerformance's, with between written after the test that it's there.
func hostOriginPresent(call, between string) string {
	return "export function present(): string {\n    const p = " + call + ";\n    if (!p) return \"absent\";\n" + between + "    const { performance } = p;\n" + hostOriginTested
}

// hostOriginCore is a performanceCore.ts: tsc's declarations, own, then body.
func hostOriginCore(t *testing.T, body string) string {
	t.Helper()
	return tscPerformanceDeclarations(t) + "\n" + hostOriginOwn + body
}

// hostOriginAt is where a stop at source's one `typeof performance.<member>` points: the member read,
// past typeof, as the a3 pin placed it.
func hostOriginAt(t *testing.T, source, member string) string {
	t.Helper()
	test := "typeof performance." + member + " === \"function\""
	for index, line := range strings.Split(source, "\n") {
		if column := strings.Index(line, test); column >= 0 {
			return fmt.Sprintf("performanceCore.ts:%d:%d", index+1, column+len("typeof ")+1)
		}
	}
	t.Fatalf("no %s in the source", test)
	return ""
}

// hostOriginLineOf is the line in source holding text, as a stop's position starts.
func hostOriginLineOf(t *testing.T, source, text string) string {
	t.Helper()
	for index, line := range strings.Split(source, "\n") {
		if strings.Contains(line, text) {
			return fmt.Sprintf("performanceCore.ts:%d:", index+1)
		}
	}
	t.Fatalf("no %s in the source", text)
	return ""
}

// hostOriginWalk loads performanceCore.ts beside a main.ts that imports node:perf_hooks, both as roots,
// and gives a lowering over performanceCore.ts's checker, as the registration pin in
// host_binding_route_test.go does, so a probe can ask the walk itself about a road tsc lets through but
// no lowering reaches.
func hostOriginWalk(t *testing.T, core string) (*lowering, *ast.SourceFile) {
	t.Helper()
	directory := t.TempDir()
	for name, source := range map[string]string{"performanceCore.ts": core, "main.ts": routedPerformanceMain("present", "console.log(present());")} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts"), filepath.Join(directory, "performanceCore.ts")})
	if err != nil {
		t.Fatal(err)
	}
	var file *ast.SourceFile
	for _, candidate := range program.Files() {
		// The program names files by their canonical path, lower case on a case-insensitive disk.
		if strings.HasSuffix(strings.ToLower(candidate.FileName().AsString()), "/performancecore.ts") {
			file = candidate
		}
	}
	if file == nil {
		t.Fatal("no performanceCore.ts in the program")
	}
	checked, release := program.Checker(t.Context(), file)
	t.Cleanup(release)
	return &lowering{program: program, checker: checked}, file
}

// hostOriginNode is the first node in file of kind whose source text is text.
func hostOriginNode(t *testing.T, file *ast.SourceFile, kind ast.Kind, text string) *ast.Node {
	t.Helper()
	var found *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found == nil && node.Kind == kind && file.Text()[scanner.GetTokenPosOfNode(node, file, false):node.End()] == text {
			found = node
		}
		return found != nil || node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if found == nil {
		t.Fatalf("no %s in performanceCore.ts", text)
	}
	return found
}

// These run one after another, never t.Parallel(), for host_binding_route_test.go's reason: each lowers
// against the package's hostBindingRows swapped for the routed table and restored with defer.
// Not parallel: it swaps hostBindingRows, which every lowering in the package reads.
func TestMethodPresenceThroughHostOrigin(t *testing.T) {
	rows, err := routedPerformanceRows()
	if err != nil {
		t.Fatal(err)
	}
	defer func(table func() ([]load.HostBinding, error)) { hostBindingRows = table }(hostBindingRows)
	hostBindingRows = func() ([]load.HostBinding, error) { return rows, nil }

	// (a3, moved here from host_binding_route_test.go, where it pinned this task's wall) tsc's own body:
	// tryGetPerformance's literal holding the routed binding, destructured back out of its const, then
	// tested for now, mark, measure, clearMarks and clearMeasures. Every one of those tests gets past the
	// presence rule, and it lowers whole (measured on Codex 01a1bcd8 at 40af54a1). Node prints number.
	t.Run("performanceCore.ts's hooks", func(t *testing.T) {
		source := tscPerformanceDeclarations(t) + "\n" + tscPerformanceGetter + "\n" + tscPerformanceHooks(t) + "\nexport function hookedNow(): string {\n\tconst hooks = tryGetPerformanceHooks();\n\treturn hooks?.performanceTime ? typeof hooks.performanceTime.now() : \"absent\";\n}\n"
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": source,
			"main.ts":            routedPerformanceMain("hookedNow", "console.log(hookedNow());"),
		})
		if err == nil {
			return
		}
		for _, member := range []string{"now", "mark", "measure", "clearMarks", "clearMeasures"} {
			if stop := hostOriginAt(t, source, member) + ": stage 0 can't lower " + methodPresenceUnproven + " yet"; strings.Contains(err.Error(), stop) {
				t.Errorf("the presence test of %s still stops: %v", member, err)
			}
		}
		t.Errorf("want it to lower, got %v", err)
	})

	// The Node branch alone, adaptation 99's road with nothing between: performance here is
	// node:perf_hooks' own Performance, whose now the existing observation admits
	// (nodeProcessMethodObservation), so this holds the road beside the walk.
	t.Run("the Node branch alone", func(t *testing.T) {
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": "import * as perfHooks from \"node:perf_hooks\";\nconst { performance } = perfHooks;\nexport function present(): string {\n\treturn typeof performance.now === \"function\" ? \"yes\" : \"no\";\n}\n",
			"main.ts":            routedPerformanceMain("present", "console.log(present());"),
		})
		if err != nil {
			t.Errorf("want it to lower, got %v", err)
		}
	})

	// The Node branch as tsc's getter writes it: node:perf_hooks' export destructured from the namespace
	// import into tsc's literal, then out again as tsc's Performance. This is the walk's second road,
	// under its third, fourth and fifth, and it lowers whole (measured on Codex 01a1bcd8 at 40af54a1).
	t.Run("the Node branch through tsc's getter", func(t *testing.T) {
		source := "import * as perfHooks from \"node:perf_hooks\";\n" + tscPerformanceDeclarations(t) + "\nfunction tryGetPerformance(): " + hostOriginReturnType + " {\n    const { performance } = perfHooks;\n    return { shouldWriteNativeEvents: false, performance };\n}\n" + hostOriginPresent("tryGetPerformance()", "")
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": source,
			"main.ts":            routedPerformanceMain("present", "console.log(present());"),
		})
		if err != nil {
			t.Errorf("want it to lower, got %v", err)
		}
	})

	// Every road the walk doesn't prove keeps today's stop, at the test, in today's words. Each row's
	// program is written so that the presence test is its only wall, and each stops there (measured on
	// Codex 01a1bcd8 at 40af54a1).
	for _, probe := range []struct{ name, body string }{
		// Lower's: a program object typed Performance in the literal.
		{"a program's own Performance", hostOriginGetter("own") + hostOriginPresent("tryGetPerformance()", "")},
		// Lower's: a parameter, whatever it's handed.
		{"a parameter", "function tryGetPerformance(given: Performance): " + hostOriginReturnType + " {\n    return { shouldWriteNativeEvents: true, performance: given };\n}\n" + hostOriginPresent("tryGetPerformance(own)", "")},
		// Lower's: a let, which a write can change.
		{"a let reassigned", hostOriginGetter("performance") + "export function present(): string {\n    let p = tryGetPerformance();\n    p = { shouldWriteNativeEvents: false, performance: own };\n    const { performance } = p;\n" + hostOriginTested},
		// Compiler's must: a write to the literal's property between the call and the destructuring.
		{"a write between the call and the destructuring", hostOriginGetter("performance") + hostOriginPresent("tryGetPerformance()", "    p.performance = own;\n")},
		// Compiler's: the same write through an alias.
		{"a write through an alias", hostOriginGetter("performance") + hostOriginPresent("tryGetPerformance()", "    const q = p;\n    q.performance = own;\n")},
		// Compiler's: the result handed to a function, which writes it.
		{"the result passed to a function", hostOriginGetter("performance") + "function replace(hooks: { performance: Performance }): void {\n    hooks.performance = own;\n}\n" + hostOriginPresent("tryGetPerformance()", "    replace(p);\n")},
		// Language's must: a getter on another property, which the destructuring runs first and which
		// writes performance through this. Node prints "no" for it (Language, run on Node v24.14.1).
		{"a getter on another property", "function tryGetPerformance(): { readonly first: number; performance: Performance; } | undefined {\n    if (typeof performance === \"object\") {\n        return {\n            get first(): number {\n                this.performance = own;\n                return 1;\n            },\n            performance,\n        };\n    }\n\n    return undefined;\n}\nexport function present(): string {\n    const p = tryGetPerformance();\n    if (!p) return \"absent\";\n    const { first, performance } = p;\n    if (first < 0) return \"negative\";\n" + hostOriginTested},
		// Language's: a return only inside a nested arrow, which is the arrow's, never the function's.
		{"a return only in a nested arrow", "function tryGetPerformance(): " + hostOriginReturnType + " {\n    const make = (): " + hostOriginReturnType + " => {\n        if (typeof performance === \"object\") {\n            return { shouldWriteNativeEvents: true, performance };\n        }\n        return undefined;\n    };\n    make();\n    return undefined;\n}\n" + hostOriginPresent("tryGetPerformance()", "")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := hostOriginCore(t, probe.body)
			_, err := lowerModules(t, map[string]string{
				"performanceCore.ts": source,
				"main.ts":            routedPerformanceMain("present", "console.log(present());"),
			})
			want := hostOriginAt(t, source, "now") + ": stage 0 can't lower " + methodPresenceUnproven + " yet"
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), want) {
				t.Errorf("got %v, want %s", err, want)
			}
		})
	}

	// Two of Language's roads never reach the walk: an earlier stop closes each, measured on Codex
	// 01a1bcd8 at 40af54a1. hostOriginBinding's default check and hostOriginProperty's spread check
	// stay as defense for when those stops lift, and these rows go red then, to be moved into the table
	// above.
	for _, probe := range []struct{ name, body, want string }{
		// A default, which hands over own wherever the property is undefined: a destructured name with
		// a default isn't lowered yet.
		{"a destructuring with a default", hostOriginGetter("performance") + "export function present(): string {\n    const p = tryGetPerformance();\n    if (!p) return \"absent\";\n    const { performance = own } = p;\n" + hostOriginTested, "performanceCore.ts:39:13: stage 0 can't lower a destructured name that isn't plain yet"},
		// A spread after the host field, which overrides it: a spread anywhere but first is refused
		// (adamic/single-spread), so the override can't be written. Its performance is optional so tsc
		// doesn't refuse the field before it as always overwritten.
		{"a spread after the host field", "const override: { performance?: Performance } = { performance: own };\nfunction tryGetPerformance(): " + hostOriginReturnType + " {\n    if (typeof performance === \"object\") {\n        return { shouldWriteNativeEvents: true, performance, ...override };\n    }\n\n    return undefined;\n}\n" + hostOriginPresent("tryGetPerformance()", ""), "performanceCore.ts:29:62: Adamic 0.1 refuses a spread after the first field; spread once, first: { ...source, field: value } (adamic/single-spread)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := hostOriginCore(t, probe.body)
			_, err := lowerModules(t, map[string]string{
				"performanceCore.ts": source,
				"main.ts":            routedPerformanceMain("present", "console.log(present());"),
			})
			if err == nil || !strings.Contains(err.Error(), probe.want) {
				t.Errorf("got %v, want %s", err, probe.want)
			}
		})
	}

	// Language's condition 5: the fold rests on nothing taking now off the host object, and each road
	// that could stops at the write, never at a folded test. A write through the destructured local is a
	// method read as a value (methodValueRefusal); Object.defineProperty is refused, or the host object
	// handed to its type parameter stops first. Either way the stop is on the write's line, in these
	// words (measured on Codex 01a1bcd8 at 40af54a1).
	for _, probe := range []struct{ name, write, words string }{
		{"a write to now through the local", "    performance.now = (): number => 0;\n", "Adamic 0.1 refuses a method read as a value (now would lose its object, and this with it)"},
		{"Object.defineProperty of now through the local", "    Object.defineProperty(performance, \"now\", { value: undefined });\n", "Adamic 0.1 refuses Object.defineProperty"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source := hostOriginCore(t, hostOriginGetter("performance")+"export function present(): string {\n    const p = tryGetPerformance();\n    if (!p) return \"absent\";\n    const { performance } = p;\n"+probe.write+hostOriginTested)
			_, err := lowerModules(t, map[string]string{
				"performanceCore.ts": source,
				"main.ts":            routedPerformanceMain("present", "console.log(present());"),
			})
			if want := hostOriginLineOf(t, source, strings.TrimSpace(probe.write)); err == nil || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), probe.words) {
				t.Errorf("got %v, want a stop on the write's line, %s, saying %s", err, want, probe.words)
			}
		})
	}

	// Language's condition 4: the walk proves the destructured local is the host object, and the fold
	// still asks the model whether the member is a method. timeOrigin is registered but a field, whose
	// typeof Node answers "number", so it never folds to "function". A field read is no presence test
	// (alwaysPresentMethod), so methodPresence doesn't meet it today; this holds the check that would
	// answer if it did.
	t.Run("timeOrigin is not a method", func(t *testing.T) {
		l, file := hostOriginWalk(t, hostOriginCore(t, hostOriginGetter("performance")+"export function present(): string {\n    const p = tryGetPerformance();\n    if (!p) return \"absent\";\n    const { performance } = p;\n    return typeof performance.timeOrigin === \"function\" || typeof performance.now === \"function\" ? \"yes\" : \"no\";\n}\n"))
		timeOrigin := hostOriginNode(t, file, ast.KindPropertyAccessExpression, "performance.timeOrigin")
		if l.hostOrigin(timeOrigin.AsPropertyAccessExpression().Expression, nil, map[*ast.Node]bool{}) == nil {
			t.Fatal("the walk doesn't prove the destructured local the host object, so this row proves nothing")
		}
		if l.hostOriginMethod(timeOrigin) {
			t.Error("timeOrigin, a field, answers as a method")
		}
		if !l.hostOriginMethod(hostOriginNode(t, file, ast.KindPropertyAccessExpression, "performance.now")) {
			t.Error("now doesn't answer as a method")
		}
	})

	// Language's generator: its result is an iterator, never its return value, so a call to one is
	// unknown. tsc refuses reading performance off a Generator, so this asks the walk itself, beside the
	// same literal from a plain function, which it follows.
	t.Run("a generator's result", func(t *testing.T) {
		l, file := hostOriginWalk(t, hostOriginCore(t, tscPerformanceGetter+"function* tryGetGenerated(): Generator<never, "+hostOriginReturnType+", unknown> {\n    if (typeof performance === \"object\") {\n        return { shouldWriteNativeEvents: true, performance };\n    }\n\n    return undefined;\n}\nexport function present(): string {\n    const p = tryGetPerformance();\n    const g = tryGetGenerated();\n    return !p ? \"absent\" : !g ? \"never\" : \"present\";\n}\n"))
		read := []string{"performance"}
		if l.hostOrigin(hostOriginNode(t, file, ast.KindPrefixUnaryExpression, "!p").AsPrefixUnaryExpression().Operand, read, map[*ast.Node]bool{}) == nil {
			t.Fatal("the plain function's literal isn't followed, so this row proves nothing")
		}
		if l.hostOrigin(hostOriginNode(t, file, ast.KindPrefixUnaryExpression, "!g").AsPrefixUnaryExpression().Operand, read, map[*ast.Node]bool{}) != nil {
			t.Error("a generator's result answers as its return value's")
		}
	})
}
