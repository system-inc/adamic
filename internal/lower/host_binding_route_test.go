package lower

import (
	"errors"
	"slices"
	"strings"
	"testing"

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
//     Pin d2 lowers: the later view naming eventLoopUtilization, which the program's interface doesn't
//     declare, is never walked. Pin d stays not yet either way, since its write stops at
//     nodeProcessStorageABI, which reads the IR rather than the view, so d2 is the pin that holds the arm.
//   - (m2) host_bindings.go, hostBindingModelRefusal's widened check:
//     `if found := l.widened(model, view, map[[2]*checker.Type]bool{}); found != nil {` becomes
//     `if found := l.widened(model, view, map[[2]*checker.Type]bool{}); false && found != nil {`.
//     Pin b2 lowers: timeOrigin declared writable over the model's read only field. Pin b doesn't move,
//     since a member typed other than the model's isn't a widening; the assignability check holds it.
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

// routedPerformanceMembers is performanceCore.ts's own Performance, every member registered for
// node:perf_hooks.Performance and typed so the model's fits it.
const routedPerformanceMembers = "\treadonly timeOrigin: number;\n\tnow(): number;\n\tmark(name: string): void;\n\tmeasure(name: string, startMark?: string, endMark?: string): void;\n\tclearMarks(name?: string): void;\n\tclearMeasures(name?: string): void;\n"

// routedPerformancePresent is a body that only asks whether performance is there, so a declaration's
// stop is the only thing that can keep it from lowering.
const routedPerformancePresent = "export function present(): boolean {\n\treturn typeof performance === \"object\";\n}\n"

// routedPerformanceMain imports one function from performanceCore.ts and calls it as main.ts's line 2.
func routedPerformanceMain(name, call string) string {
	return "import { " + name + " } from './performanceCore.ts';\n" + call + "\n"
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

	// (a) performanceCore.ts's own shape lowers, and its now() is the host's: performance.now() is the
	// model's now, the call Node's own global's lowers to.
	t.Run("performanceCore.ts's shape", func(t *testing.T) {
		program, err := lowerModules(t, map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, "export function now(): number {\n\treturn typeof performance === \"object\" ? performance.now() : 0;\n}\n"),
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

	// (d2) a later view reaching a member the host object has and the program's interface doesn't
	// declare: walked only because nodeHostObject knows the interface as the model's host object.
	t.Run("a later view reaching an unregistered member", func(t *testing.T) {
		_, err := lowerModules(t, map[string]string{
			"performanceCore.ts": routedPerformanceCore(routedPerformanceMembers, "export function view(): void {\n\tif (typeof performance === \"object\") {\n\t\tconst time: { now(): number; eventLoopUtilization?: () => number } = performance;\n\t\tconsole.log(time.now() > 0 ? \"later\" : \"origin\");\n\t}\n}\n"),
			"main.ts":            routedPerformanceMain("view", "view();"),
		})
		const want = "which reaches Performance.eventLoopUtilization, a member no host registration names"
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), want) {
			t.Errorf("got %v, want a not yet that %s", err, want)
		}
	})
}
