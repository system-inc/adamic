package lower

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A Provided host binding routes its reads through its model path (host_bindings.go, #yewmv46). No
// row of today's table routes, so these pins lower against the performance row as its owners'
// clearance will make it: Provided on every target, through node:perf_hooks.
//
// Mutants, each one line, and the pin that goes red:
//   - (m1) library_node.go, nodeHostObject's arm for a Provided row's declared type:
//     `if model := l.hostBindingViews()[proven.Symbol()]; model != nil && model != proven {` becomes
//     `if model := l.hostBindingViews()[proven.Symbol()]; false && model != nil && model != proven {`.
//     Equivalent today, measured on Cloud at 52df3e2c: every road a later view could take is closed
//     ahead of the arm. The declaration check closes the interface's names and types, tsc refuses a
//     required member the interface lacks, and the hiding rule refuses an optional one (pin d2), so
//     no pin goes red. The arm is Language's must on #yewmv46, kept for the day one of those rules
//     changes.
//   - (m2) host_bindings.go, hostBindingModelRefusal's widened check:
//     `if found := l.widened(model, view, map[[2]*checker.Type]bool{}); found != nil {` becomes
//     `if found := l.widened(model, view, map[[2]*checker.Type]bool{}); false && found != nil {`.
//     Pins b2 and a2 lower: timeOrigin declared writable over the model's read only field. Pin b doesn't move,
//     since a member typed other than the model's isn't a widening; the assignability check holds it.
//   - (m4) library_node_process.go, nodeProcessPath routing by the registered type, not the symbol:
//     before `if row, _ := l.hostBindingOf(node); row != nil && hostBindingRouted(*row) {`, the line
//     `if l.hostBindingViews()[l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)).Symbol()] != nil { return "performance" }`.
//     Pin h07 goes red: the literal's now() becomes a host call.
//   - (m3) host_bindings.go, hostBindingModelRefusal's member assignability:
//     `if !l.checker.IsTypeAssignableTo(held, declared) {` becomes `if false {`. Pin b goes red: its
//     stop moves to the whole-type check's words, which name no member.

// routedPerformanceRows is today's table with performance's row Provided on every target, its model
// node:perf_hooks.
func routedPerformanceRows() ([]load.HostBinding, error) {
	rows, err := load.HostBindings()
	if err != nil {
		return nil, err
	}
	routed := slices.Clone(rows)
	for index := range routed {
		if routed[index].Name == "performance" {
			routed[index].Native = load.HostBindingColumn{Answer: load.HostBindingProvided}
			routed[index].WASM32 = load.HostBindingColumn{Answer: load.HostBindingProvided}
			routed[index].Model = load.HostBindingModel{Path: "node:perf_hooks"}
		}
	}
	return routed, nil
}

// routedPerformanceCore is a performanceCore.ts declaring performance on line 1, its row's declaration
// exactly, then an interface Performance holding members, then body.
func routedPerformanceCore(members, body string) string {
	return "declare const performance: Performance | undefined;\ninterface Performance {\n" + members + "}\n" + body
}

// routedPerformanceMembers is a flattened Performance for the probes below, every member registered for
// node:perf_hooks.Performance and typed so the model's fits it. It isn't tsc's text: tsc's Performance
// extends PerformanceTime, whose timeOrigin is writable until Correctness's adaptation makes it read
// only. Pin (a) reads tsc's text from tscPerformanceDeclarations instead.
const routedPerformanceMembers = "\treadonly timeOrigin: number;\n\tnow(): number;\n\tmark(name: string): void;\n\tmeasure(name: string, startMark?: string, endMark?: string): void;\n\tclearMarks(name?: string): void;\n\tclearMeasures(name?: string): void;\n"

// tscPerformanceDeclarations is lines 7 to 29 of tsc's adapted src/compiler/performanceCore.ts, byte for
// byte (CRLF line ends included): the file stage3/apply.py builds from TypeScript 6.0.3 at 050880ce at
// Adamic main 7ca58dd0, sha256 8b1993ab380a178b24fed0dcc9e7dc312172a3e3be642ee1b533161f792f0dae (Codex
// 01a1bba6). Adaptation 49 makes PerformanceHooks' optional fields `| undefined`, and 99 makes
// PerformanceTime's timeOrigin read only, which says what tsc already does (Node's timeOrigin is a
// getter with no setter, and tsc never writes it). The text lives in testdata so a stage3 check can
// hold it equal to the adapted source, and it can't drift from tsc's again (Runtime's must on #yewmv46:
// this pin once restated a flattened Performance tsc doesn't declare).
func tscPerformanceDeclarations(t *testing.T) string {
	t.Helper()
	text, err := os.ReadFile("testdata/host_binding/performance_core_adapted.txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

// tscPerformanceHooks is the adapted file's tryGetPerformanceHooks, lines 62 to 88, byte for byte. It holds tsc's three roads past the declaration (Language on #yewmv46): the binding into an
// object literal, destructured back out, then seen as PerformanceTime, a base interface no row
// registers, at `hooks.performanceTime = performance`. tscPerformanceGetter stands before it.
func tscPerformanceHooks(t *testing.T) string {
	t.Helper()
	text, err := os.ReadFile("testdata/host_binding/performance_core_hooks.txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

// tscPerformanceGetter is the adapted file's tryGetPerformance (lines 31 to 60), its declared return type
// as adaptation 99 writes it, with only its browser branch, the one that reads the row's binding: the
// Node branch reads node:perf_hooks' export, another road, which isn't this row's.
const tscPerformanceGetter = "function tryGetPerformance(): { shouldWriteNativeEvents: boolean; performance: Performance; } | undefined {\n    if (typeof performance === \"object\") {\n        return {\n            shouldWriteNativeEvents: true,\n            performance,\n        };\n    }\n\n    return undefined;\n}\n"

// declarationPosition is where a stop at performance's declaration points in source: its name, on the
// line that declares it.
func declarationPosition(t *testing.T, source string) string {
	t.Helper()
	for index, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, "declare const performance:") {
			return fmt.Sprintf("performanceCore.ts:%d:15", index+1)
		}
	}
	t.Fatal("no declaration of performance in the source")
	return ""
}

// routedPerformancePresent is a body that only asks whether performance is there, so a declaration's
// stop is the only thing that can keep it from lowering.
const routedPerformancePresent = "export function present(): boolean {\n\treturn typeof performance === \"object\";\n}\n"

// routedPerformanceMain imports node:perf_hooks, so @types/node's ambient module the row's model names is
// in the program as it is in tsc's, then one function from performanceCore.ts, and calls it on line 3.
func routedPerformanceMain(name, call string) string {
	return "import 'node:perf_hooks';\nimport { " + name + " } from './performanceCore.ts';\n" + call + "\n"
}

// These run one after another, never t.Parallel(): each swaps the package's hostBindingRows for the
// routed table and restores it with defer. A parallel test, or subtest, would read the swapped table
// from another test's lowering, or run after the restore. Go runs every top-level parallel test only
// once the sequential ones are done, so the parallel host binding tests never see the swap.
func TestHostBindingProvidedRoutesThroughItsModel(t *testing.T) {
	rows, err := routedPerformanceRows()
	if err != nil {
		t.Fatal(err)
	}
	defer func(table func() ([]load.HostBinding, error)) { hostBindingRows = table }(hostBindingRows)
	hostBindingRows = func() ([]load.HostBinding, error) { return rows, nil }

	// (a) performanceCore.ts's own declarations, as adapted, lower, and its now() is the host's:
	// performance.now() is the model's now, the call Node's own global's lowers to.
	t.Run("performanceCore.ts's shape", func(t *testing.T) {
		program, err := lowerModules(t, map[string]string{
			"performanceCore.ts": tscPerformanceDeclarations(t) + "\nexport function now(): number {\n\treturn typeof performance === \"object\" ? performance.now() : 0;\n}\n",
			"main.ts":            routedPerformanceMain("now", "console.log(typeof now());"),
		})
		if err != nil {
			t.Fatalf("want it to lower, got %v", err)
		}
		routed := false
		find := func(value any) bool {
			if call, ok := value.(ir.ProcessCall); ok && call.Operation == "now" {
				routed = true
			}
			return true
		}
		walk(program.Main, find)
		for _, function := range program.Functions {
			walk(function.Body, find)
		}
		if !routed {
			t.Error("performance.now() didn't lower to the host's now")
		}
	})

	// (a3) tsc's own body past the declaration, the literal, the destructuring and the PerformanceTime view
	// (Language's should on #yewmv46). Measured on Codex 01a1bbb6 at c5a21f3e, it doesn't lower yet: it
	// stops at tryGetPerformanceHooks' `typeof performance.now === "function"`, since performance there is
	// destructured from tryGetPerformance's literal, whose construction doesn't prove the method present.
	// That's the next wall on #m36zjrk's road, owned by lowering's presence rule, and this pins it by its
	// words and place, so a change to the routing that moves it shows here. Node prints number.
	t.Run("performanceCore.ts's hooks", func(t *testing.T) {
		source := tscPerformanceDeclarations(t) + "\n" + tscPerformanceGetter + "\n" + tscPerformanceHooks(t) + "\nexport function hookedNow(): string {\n\tconst hooks = tryGetPerformanceHooks();\n\treturn hooks?.performanceTime ? typeof hooks.performanceTime.now() : \"absent\";\n}\n"
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": source,
			"main.ts":            routedPerformanceMain("hookedNow", "console.log(hookedNow());"),
		})
		where := ""
		for index, line := range strings.Split(source, "\n") {
			if column := strings.Index(line, `typeof performance.now === "function"`); column >= 0 {
				where = fmt.Sprintf("performanceCore.ts:%d:%d", index+1, column+len("typeof ")+1)
			}
		}
		want := where + ": stage 0 can't lower a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it yet"
		var notYet *NotYet
		if where == "" || !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), want) {
			t.Errorf("got %v, want %s", err, want)
		}
	})

	// (a2) tsc's text before the adaptation, timeOrigin writable, stops at the declaration: the refusal
	// the adaptation exists to clear. Correctness's mutant (the writable field restored) lands here.
	t.Run("performanceCore.ts's shape before the adaptation", func(t *testing.T) {
		declarations := tscPerformanceDeclarations(t)
		unadapted := strings.Replace(declarations, "readonly timeOrigin: number;", "timeOrigin: number;", 1)
		if unadapted == declarations {
			t.Fatal("the adapted text holds no read only timeOrigin")
		}
		source := unadapted + "\n" + routedPerformancePresent
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": source,
			"main.ts":            routedPerformanceMain("present", "console.log(present() ? \"present\" : \"absent\");"),
		})
		want := declarationPosition(t, source) + ": stage 0 can't lower the host binding performance, declared as Performance, whose field timeOrigin is writable where its model export node:perf_hooks.performance's is read only, yet"
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), want) {
			t.Errorf("got %v, want %s", err, want)
		}
	})

	// Each declaration its model export doesn't fit stops at the declaration, named. The texts are
	// written from hostBindingModelRefusal's words and unmeasured until the box runs them.
	for _, probe := range []struct{ name, members, want string }{
		// (b) Compiler's: a member typed other than the model's.
		{"a member typed other than the model's", strings.Replace(routedPerformanceMembers, "now(): number", "now(): string", 1),
			"the host binding performance, declared as Performance, whose now is typed () => string where its model export node:perf_hooks.performance's is () => number,"},
		// (b2) a field writable where the model's is read only.
		{"a writable field over the model's read only one", strings.Replace(routedPerformanceMembers, "readonly timeOrigin", "timeOrigin", 1),
			"the host binding performance, declared as Performance, whose field timeOrigin is writable where its model export node:perf_hooks.performance's is read only,"},
		// (c) a member no registration names.
		{"a member no registration names", routedPerformanceMembers + "\teventLoopUtilization(): unknown;\n",
			"the host binding performance, declared as Performance, which reaches node:perf_hooks.Performance.eventLoopUtilization of its model export node:perf_hooks.performance, a member no host registration names,"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerModules(t, map[string]string{
				"performanceCore.ts": routedPerformanceCore(probe.members, routedPerformancePresent),
				"main.ts":            routedPerformanceMain("present", "console.log(present() ? \"present\" : \"absent\");"),
			})
			want := "performanceCore.ts:1:15: stage 0 can't lower " + probe.want + " yet"
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), want) {
				t.Errorf("got %v, want %s", err, want)
			}
		})
	}

	// (d) Language's: a later view of the program's interface that writes a slot the host owns. Expected,
	// unmeasured: "stage 0 can't lower writing performance host slots through a structural view yet",
	// nodeProcessStorageABI's, at the first module.
	// Without a Node module in the program, @types/node's ambient module isn't loaded, so the model export
	// can't be found and the declaration stops, failing closed (measured on Cloud at 45a05281, where every
	// pin stopped here before main imported node:perf_hooks).
	t.Run("no Node module in the program", func(t *testing.T) {
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, routedPerformancePresent),
			"main.ts":            "import { present } from './performanceCore.ts';\nconsole.log(present() ? \"present\" : \"absent\");\n",
		})
		var notYet *NotYet
		const want = "performanceCore.ts:1:15: stage 0 can't lower the host binding performance's model export node:perf_hooks.performance, which lowering can't find, yet"
		if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), want) {
			t.Fatalf("got %v, want %q", err, want)
		}
	})
	t.Run("a later view writing now", func(t *testing.T) {
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, "export function rewrite(): void {\n\tif (typeof performance === \"object\") {\n\t\tconst time: { now: () => number } = performance;\n\t\ttime.now = (): number => 0;\n\t\tconsole.log(time.now() > 0 ? \"later\" : \"origin\");\n\t}\n}\n"),
			"main.ts":            routedPerformanceMain("rewrite", "rewrite();"),
		})
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Errorf("got %v, want a not yet", err)
		}
	})

	// (h01) Language's two-hop write: the binding seen as a program interface no row registers, a
	// legal read-only hop, then that seen as a writable view and now written. nodeHostObject's arm never
	// sees Clock; what stops it is the whole-program storage check, and this pins that it does.
	t.Run("a two-hop view writing now", func(t *testing.T) {
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, "interface Clock {\n\tnow(): number;\n}\nexport function rewrite(): void {\n\tif (typeof performance === \"object\") {\n\t\tconst clock: Clock = performance;\n\t\tconst time: { now: () => number } = clock;\n\t\ttime.now = (): number => 0;\n\t\tconsole.log(time.now() > 0 ? \"later\" : \"origin\");\n\t}\n}\n"),
			"main.ts":            routedPerformanceMain("rewrite", "rewrite();"),
		})
		const want = "stage 0 can't lower writing performance host slots through a structural view yet"
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), want) {
			t.Errorf("got %v, want %q", err, want)
		}
	})

	// (h07) Compiler's routing condition: a program's own literal typed Performance is the program's,
	// never the host's. The registration feeds the view walk only; routing goes by the binding's symbol.
	// Node prints own. Mutant m4 routes by the registered type, and this lowers a host now call.
	t.Run("a program's own Performance literal", func(t *testing.T) {
		program, err := lowerModules(t, map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, "export function ownNow(): string {\n\tconst own: Performance = { timeOrigin: 0, now: (): number => 5, mark(name: string): void {}, measure(name: string, startMark?: string, endMark?: string): void {}, clearMarks(name?: string): void {}, clearMeasures(name?: string): void {} };\n\treturn own.now() === 5 ? \"own\" : \"host\";\n}\n"),
			"main.ts":            routedPerformanceMain("ownNow", "console.log(ownNow());"),
		})
		if err != nil {
			t.Fatalf("want it to lower, got %v", err)
		}
		host := false
		find := func(value any) bool {
			if call, ok := value.(ir.ProcessCall); ok && call.Operation == "now" {
				host = true
			}
			return true
		}
		walk(program.Main, find)
		for _, function := range program.Functions {
			walk(function.Body, find)
		}
		if host {
			t.Error("a program's own Performance literal reached the host")
		}
	})

	// The registration Language's must rests on: hostBindingViews keys the declared type's symbol, and
	// nodeHostObject answers that type as the model's host object. The two-hop and later-view pins stop
	// at rules ahead of it, so without this the arm could be dead by a missed lookup and no pin would say
	// so. It reports, too, whether a later read of a local typed Performance meets the same key, which
	// routing must never read (h07).
	t.Run("the registration holds the declared type", func(t *testing.T) {
		directory := t.TempDir()
		for name, source := range map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, "export function ownNow(): number {\n\tconst own: Performance = { timeOrigin: 0, now: (): number => 5, mark(name: string): void {}, measure(name: string, startMark?: string, endMark?: string): void {}, clearMarks(name?: string): void {}, clearMeasures(name?: string): void {} };\n\treturn own.now();\n}\n"),
			"main.ts":            routedPerformanceMain("ownNow", "console.log(`${ownNow()}`);"),
		} {
			if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		// Both are roots, since Program.Files holds only the roots it was given.
		program, err := load.Load([]string{filepath.Join(directory, "main.ts"), filepath.Join(directory, "performanceCore.ts")})
		if err != nil {
			t.Fatal(err)
		}
		var core *ast.SourceFile
		for _, file := range program.Files() {
			// The program names files by their canonical path, lower case on a case-insensitive disk.
			if strings.HasSuffix(strings.ToLower(file.FileName().AsString()), "/performancecore.ts") {
				core = file
			}
		}
		if core == nil {
			t.Fatal("no performanceCore.ts in the program")
		}
		checked, release := program.Checker(t.Context(), core)
		defer release()
		l := &lowering{program: program, checker: checked}
		var declared, own *ast.Node
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindVariableDeclaration && node.Name().Text() == "performance" {
				declared = node
			}
			if node.Kind == ast.KindPropertyAccessExpression && node.AsPropertyAccessExpression().Expression.Kind == ast.KindIdentifier && node.AsPropertyAccessExpression().Expression.Text() == "own" {
				own = node.AsPropertyAccessExpression().Expression
			}
			return node.ForEachChild(visit)
		}
		core.AsNode().ForEachChild(visit)
		if declared == nil || own == nil {
			t.Fatal("the declaration or the read of own is missing")
		}
		view := checked.GetNonNullableType(checked.GetTypeAtLocation(declared.Name()))
		if l.hostBindingViews()[view.Symbol()] == nil {
			t.Fatalf("hostBindingViews holds no entry for %s (%d entries)", checked.TypeToString(view), len(l.hostBindingViews()))
		}
		if _, host := l.nodeHostObject(view); !host {
			t.Fatalf("nodeHostObject doesn't answer %s as a host object", checked.TypeToString(view))
		}
		ownType := checked.GetNonNullableType(checked.GetTypeAtLocation(own))
		t.Logf("own reads as %s, same symbol as the declaration's: %v, registered: %v", checked.TypeToString(ownType), ownType.Symbol() == view.Symbol(), l.hostBindingViews()[ownType.Symbol()] != nil)
	})

	// (d2) a later view reaching a member the host object has and the program's interface doesn't
	// declare. Measured on Cloud at 52df3e2c, the hiding rule refuses it first (no-optional-widening):
	// an optional member the source lacks. A required one is a tsc error, and the declaration check closes
	// names and types, so no road reaches nodeHostObject's arm today, and m1 is an equivalent mutant (see
	// the top of this file). The arm stays as defense for the day a rule ahead of it changes.
	t.Run("a later view reaching an unregistered member", func(t *testing.T) {
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, "export function view(): void {\n\tif (typeof performance === \"object\") {\n\t\tconst time: { now(): number; eventLoopUtilization?: () => number } = performance;\n\t\tconsole.log(time.now() > 0 ? \"later\" : \"origin\");\n\t}\n}\n"),
			"main.ts":            routedPerformanceMain("view", "view();"),
		})
		const want = "performanceCore.ts:12:72: Adamic 0.1 refuses optional property eventLoopUtilization in { now(): number; eventLoopUtilization?: () => number; } absent from structural"
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), want) {
			t.Errorf("got %v, want the hiding rule's refusal %q", err, want)
		}
	})
}
