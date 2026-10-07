// Package mutation_aliasing holds the test of cohere's mutation_aliasing module ported to Adamic (the
// .ts files beside this one): the effect vocabulary, the alias graph, the mutate worklist and the mutable
// ranges, compiled natively by stage 0 and run on Node, running the same passes on the same functions as
// Go cohere's module and printing what they made, byte for byte as Go cohere does.
package mutation_aliasing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../../.."

// portFiles are the port and its driver, and the single assignment port they import, by their paths from
// here.
var portFiles = []string{
	"mutation_aliasing.ts", "ranges.ts", "main.ts",
	"../static_single_assignment/static_single_assignment.ts", "../static_single_assignment/graph.ts",
	"../static_single_assignment/phi.ts", "../static_single_assignment/construct.ts",
	"../static_single_assignment/eliminate.ts", "../static_single_assignment/verify.ts",
}

// generatedSeed fixes the generated functions, so a failure names functions anyone can make again.
// COHERE_MUTATION_ALIASING_SEED asks for others, COHERE_MUTATION_ALIASING_GENERATED for more or fewer of
// them, and COHERE_MUTATION_ALIASING_GRAPHS names a directory of exported functions to run beside
// testdata/react_ranges.txt.gz, the React Compiler fixtures cohere's high-level IR lowers, with their
// effects (README.md says how both are made).
const generatedSeed = 20261007

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port makes of every function what Go cohere's module makes of it: the module's own tests' functions
// and probes, thousands generated, and the ones React Compiler's fixtures lower to, with their effects:
// natively, on Node and through the JavaScript backend, byte for byte; and the native port leaks nothing.
func TestThePortMakesWhatGoCohereMakes(t *testing.T) {
	t.Parallel()
	casesPath, goAnswers := askedCases(t)
	portSource := portDirectory(t, nil)
	program := lowered(t, filepath.Join(portSource, "main.ts"))

	t.Run("natively and on Node, as Go cohere", func(t *testing.T) {
		t.Parallel()
		nodeRun := onNode(t, filepath.Join(portSource, "main.ts"), casesPath)
		nativeRun, sanitized := natively(t, program, casesPath)
		backendRun := onJavaScriptBackend(t, program, casesPath)
		for _, side := range []struct {
			name string
			run  run
		}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
			if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
				t.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
			}
			if difference := firstDifference(string(side.run.stdout), goAnswers); difference != "" {
				t.Errorf("%s and Go cohere differ: %s", side.name, difference)
			}
		}
		agreed := !t.Failed()
		if leaked := leaks(t, program, sanitized, casesPath); leaked != "" {
			t.Errorf("leaks:\n%s", leaked)
		}

		// Functions the passes make nothing of would agree without testing them: each thing the passes
		// make has to be among the answers. A phi node, every kind of edge and back edge, a mutation of
		// each kind, every value kind, a range the widening opened and one only the definition half did, an
		// instruction the function doesn't hold, a block finalize never numbered, and React's functions
		// from both of cohere's pipelines.
		var missing []string
		for _, each := range []string{`(?m)^node \d+ phi `, `:maybe-alias`, `:capture`, `:alias`, ` created-from=\d`, ` captures=\d`,
			` maybe-aliases=\d`, `(?m)^mutation \d+ \d+ \d+ 1 1$`, `(?m)^mutation \d+ \d+ \d+ 1 2$`, `(?m)^mutation \d+ \d+ \d+ 0 1$`,
			`(?m)^mutation \d+ \d+ \d+ 0 2$`, ` kind=frozen `, ` kind=maybe-frozen `, ` kind=primitive `, ` kind=global `,
			` range=\d+,\d+ widened=\d+,\d+ `, ` range=\d+,\d+ widened=- `, `instructions=[^\n]*-`, ` first=0 terminal=0 `,
			`(?m)^function \S+\.tsx?#\d+$`, `(?m)^function \S+~memo#\d+$`, `(?m)^function parameters-defined-on-entry-with$`} {
			if !regexp.MustCompile(each).MatchString(goAnswers) {
				missing = append(missing, each)
			}
		}
		if len(missing) > 0 {
			t.Errorf("the answers never show %q", missing)
		}
		// Two of the Go module's own claims, held over every function: its ranges are valid, and a second
		// run gives an equal table.
		if invalid := regexp.MustCompile(`(?m)^invalid \d.*$`).FindString(goAnswers); invalid != "" {
			t.Errorf("Go cohere's ranges break upstream's invariant: %s", invalid)
		}
		if strings.Contains(goAnswers, "\nidempotent 0\n") {
			t.Error("Go cohere's ranges differ on a second run over some function")
		}
		if agreed {
			t.Logf("%d functions (%d of them React's), %d alias graph nodes, %d mutations, %d values: every one the same from Go cohere, the port natively, on Node and through the JavaScript backend",
				strings.Count(goAnswers, "function "), len(regexp.MustCompile(`(?m)^function \S+#\d+$`).FindAllString(goAnswers, -1)),
				strings.Count(goAnswers, "\nnode "), strings.Count(goAnswers, "\nmutation "), strings.Count(goAnswers, "\nvalue "))
		}
	})

	// The comparison has to be able to fail. Each mutant changes one algorithm where only its answers can
	// show it, and the port must then disagree with Go cohere: natively, since that is the program this
	// test holds, and on Node, which shows the fault is the port's and not stage 0's.
	for _, mutant := range mutants {
		t.Run("catches "+mutant.name, func(t *testing.T) {
			t.Parallel()
			mutated := portDirectory(t, &mutant)
			mutatedProgram := lowered(t, filepath.Join(mutated, "main.ts"))
			for _, side := range []struct {
				name string
				run  run
			}{{"natively", nativelyRun(t, mutatedProgram, casesPath)}, {"on Node", onNode(t, filepath.Join(mutated, "main.ts"), casesPath)}} {
				if side.run.exitCode != 0 {
					t.Errorf("%s the mutant exits %d (stderr %q); it must be caught by its answers, not by failing", side.name, side.run.exitCode, side.run.stderr)
					continue
				}
				difference := firstDifference(string(side.run.stdout), goAnswers)
				if difference == "" {
					t.Errorf("%s the mutant agrees with Go cohere: the comparison cannot see it", side.name)
					continue
				}
				t.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}
}

// mutant is one change to one port file. A condition a mutant turns off is turned off with a test the
// answers decide (an index or order is never negative), not with `&& false`, which the checker narrows to
// unreachable code and refuses.
type mutant struct {
	name, file, from, to string
}

// The mutation kind a walk carries has no mutant: within this pass it decides only whether a node is
// visited again, and a node is widened and its edges followed the same way at either kind, so no range
// depends on it. A mutation's kind is in the answers, as collected.
var mutants = []mutant{
	// isMutation: Freeze counted as a mutation.
	{name: "Freeze counted as a mutation", file: "mutation_aliasing.ts",
		from: "        case 'MutateTransitiveConditionally':\n            return true;",
		to:   "        case 'MutateTransitiveConditionally':\n        case 'Freeze':\n            return true;"},
	// MutableRange.contains: closed at its end, not half-open.
	{name: "a range closed at its end", file: "ranges.ts",
		from: "return order >= this.start && order < this.end;",
		to:   "return order >= this.start && order <= this.end;"},
	// MutableRange.isValid: an empty range taken for a valid one.
	{name: "an empty range valid", file: "ranges.ts",
		from: "return this.end > this.start;",
		to:   "return this.end >= this.start;"},
	// insertBackEdge: the last index for a repeated edge kept, not the first.
	{name: "a repeated back edge keeping its last index", file: "ranges.ts",
		from: "if(edges.has(from)) {",
		to:   "if(edges.has(from) && index < 0) {"},
	// buildAliasingGraph: a Create consumes an index, shifting every later edge's time.
	{name: "a Create consuming an index", file: "ranges.ts",
		from: "state.immutable.delete(into);\n                        }\n                        break;",
		to:   "state.immutable.delete(into);\n                        }\n                        index++;\n                        break;"},
	// buildAliasingGraph: an Alias from or into an immutable value kept as an edge.
	{name: "Alias's refinement dropped", file: "ranges.ts",
		from: "if(!isImmutableEnd(state.immutable.get(from)) && !isImmutableEnd(state.immutable.get(into))) {\n                            state.assign(index, from, into);",
		to:   "if(index >= 0) {\n                            state.assign(index, from, into);"},
	// buildAliasingGraph: a MaybeAlias from a Frozen value kept as an edge.
	{name: "a MaybeAlias from a frozen value kept", file: "ranges.ts",
		from: "if(state.immutable.get(from) !== 'Frozen') {",
		to:   "if(index >= 0) {"},
	// buildAliasingGraph: an Assign leaving its target an identity of its own, so freezing one name no
	// longer freezes the other.
	{name: "an Assign sharing no identity", file: "ranges.ts",
		from: "state.shareIdentity(from, into);\n",
		to:   ""},
	// AliasingState.freeze: a value frozen alone, not with every name its identity has.
	{name: "a freeze skipping the value's identity", file: "ranges.ts",
		from: "if(identity === undefined) {\n            this.markImmutable(id, 'Frozen');",
		to:   "if(identity === undefined || id >= 0) {\n            this.markImmutable(id, 'Frozen');"},
	// derivePhiImmutable: a phi of frozen and mutable operands taken as Frozen, not MaybeFrozen.
	{name: "a mixed phi taken as frozen", file: "ranges.ts",
		from: "kind = 'MaybeFrozen';",
		to:   "kind = 'Frozen';"},
	// buildAliasingGraph: a conditional mutation of a value that isn't mutable kept.
	{name: "a conditional mutation of an immutable value kept", file: "ranges.ts",
		from: "case 'MutateConditionally':\n                        if(!state.notMutable(into)) {",
		to:   "case 'MutateConditionally':\n                        if(!state.notMutable(into) || index >= 0) {"},
	// buildAliasingGraph: a back edge's phi operand applied at the index of its predecessor's end, not the
	// one it was given at its phi.
	{name: "a deferred phi operand at a fresh index", file: "ranges.ts",
		from: "state.assign(pending.index, pending.from, pending.into);",
		to:   "state.assign(index, pending.from, pending.into);"},
	// mutate: a forward edge followed whatever its time.
	{name: "a forward edge followed from the future", file: "ranges.ts",
		from: "if(edge.index >= index) {",
		to:   "if(edge.index >= index && index < 0) {"},
	// mutate: captures followed by a mutation that isn't transitive.
	{name: "captures followed without a transitive mutation", file: "ranges.ts",
		from: "if(entry.transitive) {",
		to:   "if(entry.transitive || index >= 0) {"},
	// mutate: a created-from edge keeping the mutation's own transitivity rather than forcing it.
	{name: "created-from not forcing transitive", file: "ranges.ts",
		from: "queue.push({ place: alias, transitive: true, direction: 'Backwards', kind: entry.kind });",
		to:   "queue.push({ place: alias, transitive: entry.transitive, direction: 'Backwards', kind: entry.kind });"},
	// mutate: a phi reached travelling forwards propagating back through its aliases.
	{name: "a phi passing a forward mutation back", file: "ranges.ts",
		from: "if(entry.direction === 'Backwards' || node.value !== 'Phi') {",
		to:   "if(entry.direction === 'Backwards' || node.value !== 'Phi' || index >= 0) {"},
	// widen: a mutation at an unnumbered instruction widened.
	{name: "an unnumbered mutation widened", file: "ranges.ts",
		from: "if(mutation.end === 0 || mutation.end === 1) {",
		to:   "if(mutation.end === 0) {"},
	// inferMutableRanges: the operand loop, the third, left out.
	{name: "the operand loop left out", file: "ranges.ts",
		from: "walk.lvalues = false;\n            graph.eachInstructionPlace(fn, block, index, visit);",
		to:   "walk.lvalues = false;"},
	// openLValue: an lvalue's start opened again where an operand already opened it.
	{name: "an lvalue reopening its start", file: "ranges.ts",
		from: "if(start === 0) {",
		to:   "if(start === 0 || this.order >= 0) {"},
	// openLValue: an inverted loop-carried range kept, as upstream ships it.
	{name: "an inverted range kept", file: "ranges.ts",
		from: "if(range.isSet() && range.end <= range.start) {",
		to:   "if(range.isSet() && range.end <= range.start && id < 0) {"},
	// inferMutableRanges: a store into a captured binding widening nothing.
	{name: "a context store widening nothing", file: "ranges.ts",
		from: "if(existing.end <= order) {",
		to:   "if(existing.end <= order && order < 0) {"},
	// phiOpenedRange: a phi opened at its block's first instruction, not one before.
	{name: "a phi opened at its first instruction", file: "ranges.ts",
		from: "return new MutableRange(firstOrder - 1, existing.end);",
		to:   "return new MutableRange(firstOrder, existing.end);"},
	// inferMutableRanges: a parameter nothing widened given two instructions on entry, not one.
	{name: "a parameter on entry two instructions wide", file: "ranges.ts",
		from: "existing.end === 0 ? 2 : existing.end",
		to:   "existing.end === 0 ? 3 : existing.end"},
	// blockFirstOrder: an empty block's first order taken as zero, not its terminal's.
	{name: "an empty block's first order zero", file: "ranges.ts",
		from: "    return graph.terminalOrder(block);\n}",
		to:   "    return 0;\n}"},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(t *testing.T) (string, string) {
	t.Helper()
	seed := int64(generatedSeed)
	var err error
	if value := os.Getenv("COHERE_MUTATION_ALIASING_SEED"); value != "" {
		if seed, err = strconv.ParseInt(value, 10, 64); err != nil {
			t.Fatal(err)
		}
	}
	generated := 2000
	if value := os.Getenv("COHERE_MUTATION_ALIASING_GENERATED"); value != "" {
		if generated, err = strconv.Atoi(value); err != nil {
			t.Fatal(err)
		}
	}
	react, err := filepath.Abs(filepath.Join("testdata", "react_ranges.txt.gz"))
	if err != nil {
		t.Fatal(err)
	}
	graphs := []string{react}
	if value := os.Getenv("COHERE_MUTATION_ALIASING_GRAPHS"); value != "" {
		directory, err := filepath.Abs(value)
		if err != nil {
			t.Fatal(err)
		}
		graphs = append(graphs, directory)
	}
	directory := t.TempDir()
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	cohereSide(t, map[string]any{"seed": seed, "generated": generated, "graphs": graphs, "cases": casesPath, "answers": answersPath})
	answers, err := os.ReadFile(answersPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("seed %d, %d generated functions, graphs from %q", seed, generated, graphs)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's mutation_aliasing module, by overlay, in
// the package of the module's own tests, with a request. The module is one of its own, with no go.work
// above it that names it, so the side runs with GOWORK off.
func cohereSide(t *testing.T, request map[string]any) {
	t.Helper()
	directory := t.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(requestPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	module, err := filepath.Abs(filepath.Join(repository, "cohere", "mutation_aliasing"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	replace := map[string]string{filepath.Join(module, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPortCases$", ".")
	command.Dir = module
	command.Env = append(os.Environ(), "GOWORK=off", "ADAMIC_PORT_REQUEST="+requestPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("cohere's side: %v\n%s", err, output)
	}
}

// portDirectory copies the port, and the single assignment port its driver imports, into a directory of
// their own, side by side as they are here, with a mutant applied when there is one. It returns the
// copy's mutation_aliasing directory.
func portDirectory(t *testing.T, applied *mutant) string {
	t.Helper()
	directory := t.TempDir()
	applies := 0
	for _, name := range portFiles {
		contents, err := os.ReadFile(filepath.FromSlash(name))
		if err != nil {
			t.Fatal(err)
		}
		source := string(contents)
		if applied != nil && applied.file == name {
			if strings.Count(source, applied.from) != 1 {
				t.Fatalf("the mutant %q must change exactly one place in %s", applied.name, name)
			}
			source = strings.Replace(source, applied.from, applied.to, 1)
			applies++
		}
		copied := filepath.Join(directory, "mutation_aliasing", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(copied), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(copied, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if applied != nil && applies != 1 {
		t.Fatalf("the mutant %q names %s, which isn't a port file", applied.name, applied.file)
	}
	return filepath.Join(directory, "mutation_aliasing")
}

// firstDifference says where two outputs first differ, line by line, or "" when they don't.
func firstDifference(got string, want string) string {
	if got == want {
		return ""
	}
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want, "\n")
	for index := 0; index < len(gotLines) || index < len(wantLines); index++ {
		var gotLine, wantLine string
		if index < len(gotLines) {
			gotLine = gotLines[index]
		}
		if index < len(wantLines) {
			wantLine = wantLines[index]
		}
		if gotLine != wantLine {
			return fmt.Sprintf("line %d: %q, Go cohere %q", index+1, gotLine, wantLine)
		}
	}
	return "the same lines, not the same bytes"
}

// lowered checks and lowers a program, failing the test with stage 0's refusal if it can't.
func lowered(t *testing.T, path string) *ir.Program {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	result, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	return result
}

// bounded is a command that can't outlive its test: it has a deadline, it runs in a process group of
// its own, and when the deadline passes or the test ends, the whole group is killed.
func bounded(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	t.Cleanup(cancel)
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 5 * time.Second
	return command
}

func execute(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := bounded(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(t *testing.T, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return execute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return onNode(t, path, arguments...)
}

// nativelyRun is natively's run alone.
func nativelyRun(t *testing.T, program *ir.Program, arguments ...string) run {
	t.Helper()
	result, _ := natively(t, program, arguments...)
	return result
}

// natively builds the lowered port under the address and undefined-behavior sanitizers and runs it,
// returning the binary too, for the leak check. Leak detection is off here, as in the oracle; leaks is
// its own run.
func natively(t *testing.T, program *ir.Program, arguments ...string) (run, string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "port")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(t, environment, binary, arguments...), binary
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(t.TempDir(), "port")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		report := execute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	t.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}
