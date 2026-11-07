// Package static_single_assignment holds the test of cohere's static_single_assignment module ported to
// Adamic (the .ts files beside this one): reverse postorder, predecessors, evaluation order, Braun's
// construction, redundant phi elimination, the verifier and dominance, compiled natively by stage 0 and
// run on Node, running the same passes on the same functions as Go cohere's module and printing what
// they made, byte for byte as Go cohere does.
package static_single_assignment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/static_single_assignment/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../../.."

// portFiles are the port and its driver.
var portFiles = []string{
	"static_single_assignment.ts",
	"graph.ts",
	"phi.ts",
	"construct.ts",
	"eliminate.ts",
	"verify.ts",
	"main.ts",
}

// generatedSeed fixes the generated functions, so a failure names functions anyone can make again.
// COHERE_STATIC_SINGLE_ASSIGNMENT_SEED asks for others, COHERE_STATIC_SINGLE_ASSIGNMENT_GENERATED for
// more or fewer of them, and COHERE_STATIC_SINGLE_ASSIGNMENT_GRAPHS names a directory of exported
// functions to run beside testdata/react_graphs.txt.gz, the React Compiler fixtures cohere's
// high-level IR lowers (README.md says how both are made).
const generatedSeed = 20261007

// run is one execution's observable behavior.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

// The port makes of every function what Go cohere's module makes of it: the module's own ten tests'
// functions, thousands generated, and the hundreds React Compiler's fixtures lower to: natively, on Node and through
// the JavaScript backend, byte for byte; and the native port leaks nothing.
func TestThePortMakesWhatGoCohereMakes(test *testing.T) {
	test.Parallel()
	casesPath, goAnswers := askedCases(test)
	portSource := portDirectory(test, nil)
	program := lowered(test, filepath.Join(portSource, "main.ts"))

	test.Run("natively and on Node, as Go cohere", func(test *testing.T) {
		test.Parallel()
		nodeRun := onNode(test, filepath.Join(portSource, "main.ts"), casesPath)
		sanitized := filepath.Join(portSanitized(test, nil), "port")
		nativeRun := sanitizedRun(test, sanitized, casesPath)
		backendRun := onJavaScriptBackend(test, program, casesPath)
		for _, side := range []struct {
			name string
			run  run
		}{{"Node", nodeRun}, {"native", nativeRun}, {"the JavaScript backend", backendRun}} {
			if side.run.exitCode != 0 || len(side.run.stderr) > 0 {
				test.Fatalf("%s: exit %d, stderr %q", side.name, side.run.exitCode, side.run.stderr)
			}
			if difference := firstDifference(string(side.run.stdout), goAnswers); difference != "" {
				test.Errorf("%s and Go cohere differ: %s", side.name, difference)
			}
		}
		agreed := !test.Failed()
		if leaked := leaks(test, program, sanitized, casesPath); leaked != "" {
			test.Errorf("leaks:\n%s", leaked)
		}

		// Functions the passes make nothing of would agree without testing them: each thing the passes
		// make has to be among the answers, a phi of three operands, a placeholder, both kinds of
		// violation, a context store, a renamed returns place and parameter, and a function with
		// nothing to verify.
		var missing []string
		for _, each := range []string{
			`(?m)^phi \S+ \S+ \S+ \S+`,
			` placeholder `,
			`(?m)^violation MultipleDefinitions `,
			`(?m)^violation UseNotDominated `,
			` context uses=`,
			`(?m)^returns \d`,
			`(?m)^params \d`,
			`(?m)^stats 0 0 0$`,
			`(?m)^function \S+\.tsx?#\d+$`,
		} {
			if !regexp.MustCompile(each).MatchString(goAnswers) {
				missing = append(missing, each)
			}
		}
		if len(missing) > 0 {
			test.Errorf("the answers never show %q", missing)
		}
		if agreed {
			test.Logf(
				"%d functions (%d of them React's), %d blocks, %d phis, %d violations: every one the same from Go cohere, the port natively, on Node and through the JavaScript backend",
				strings.Count(goAnswers, "function "),
				len(regexp.MustCompile(`(?m)^function \S+#\d+$`).FindAllString(goAnswers, -1)),
				strings.Count(goAnswers, "\nblock "),
				strings.Count(goAnswers, "\nphi "),
				strings.Count(goAnswers, "\nviolation "),
			)
		}
	})

	// The comparison has to be able to fail. Each mutant changes one algorithm where only its answers can
	// show it, and the port must then disagree with Go cohere: natively, since that is the program this
	// test is about (the mutant's sanitized product, built ahead), and on Node, which shows the fault is the
	// port's and not stage 0's.
	for _, mutant := range mutants {
		test.Run("catches "+mutant.name, func(test *testing.T) {
			test.Parallel()
			mutated := portDirectory(test, &mutant)
			sanitized := filepath.Join(portSanitized(test, &mutant), "port")
			for _, side := range []struct {
				name string
				run  run
			}{
				{"natively", sanitizedRun(test, sanitized, casesPath)},
				{"on Node", onNode(test, filepath.Join(mutated, "main.ts"), casesPath)},
			} {
				if side.run.exitCode != 0 {
					test.Errorf(
						"%s the mutant exits %d (stderr %q); it must be caught by its answers, not by failing",
						side.name,
						side.run.exitCode,
						side.run.stderr,
					)
					continue
				}
				difference := firstDifference(string(side.run.stdout), goAnswers)
				if difference == "" {
					test.Errorf("%s the mutant agrees with Go cohere: the comparison cannot see it", side.name)
					continue
				}
				test.Logf("%s, caught: %s", side.name, difference)
			}
		})
	}
}

// mutant is one change to one port file. A condition a mutant turns off is turned off with a test no id
// passes (ids are never negative), not with `&& false`, which the checker narrows to unreachable code and
// refuses.
type mutant struct {
	name, file, from, to string
}

var mutants = []mutant{
	// reversePostorder: the real successors walked in their own order rather than reversed, so a
	// conditional's arms and a loop's body come out in another order.
	{
		name: "reverse postorder walking successors forwards",
		file: "graph.ts",
		from: "for(let index = real.length - 1; index >= 0; index--) {",
		to:   "for(let index = 0; index < real.length; index++) {",
	},
	// reversePostorder: a fallthrough nothing real reaches dropped rather than kept as a placeholder.
	{
		name: "a fallthrough nothing reaches dropped",
		file: "graph.ts",
		from: "else if(inRange(id, bound) && usedFallthroughs[id] === true) {",
		to:   "else if(inRange(id, bound) && usedFallthroughs[id] === true && id < 0) {",
	},
	// markPredecessors: a block's two edges to one successor make it a predecessor twice.
	{
		name: "a predecessor counted once per edge",
		file: "graph.ts",
		from: "if(edge.edge === 'Fallthrough' || seen.get(edge.successor) === true) {",
		to:   "if(edge.edge === 'Fallthrough') {",
	},
	// markEvaluationOrder: a terminal shares the next instruction's order rather than taking its own.
	{
		name: "a terminal taking no order of its own",
		file: "graph.ts",
		from: "        graph.setTerminalOrder(block, order);\n        order++;",
		to:   "        graph.setTerminalOrder(block, order);",
	},
	// construct: a context store versioned like any other store rather than reusing its definition.
	{
		name: "a context store versioned",
		file: "construct.ts",
		from: "if(contextStore && this.graph.contextual(this.fn, binding)) {",
		to:   "if(contextStore && this.graph.contextual(this.fn, binding) && binding < 0) {",
	},
	// construct: a block with one sealed predecessor gets a phi rather than its predecessor's value,
	// which elimination removes, after minting values the output counts.
	{
		name: "a single predecessor merged by a phi",
		file: "construct.ts",
		from: "if(predecessors.length === 1) {",
		to:   "if(predecessors.length === -1) {",
	},
	// construct: the returns place resolved at the first block, not at a block that returns.
	{
		name: "the returns place resolved at any block",
		file: "construct.ts",
		from: "if(!this.graph.endsInReturn(block)) {",
		to:   "if(!this.graph.endsInReturn(block) && false) {",
	},
	// eliminateRedundantPhis: a phi's own result counted as an operand, so a loop that never writes
	// keeps its header's phi.
	{
		name: "a phi's own result counted as an operand",
		file: "eliminate.ts",
		from: "if(operand === result) {",
		to:   "if(operand === result && false) {",
	},
	// withPhiOperand: an operand appended rather than kept in predecessor order.
	{
		name: "phi operands out of predecessor order",
		file: "phi.ts",
		from: "if(operand.predecessor >= predecessor) {",
		to:   "if(operand.predecessor === predecessor) {",
	},
	// verifySingleAssignment: a phi operand checked at the phi's block rather than its predecessor's
	// exit, which reports every correct loop phi.
	{
		name: "a phi operand checked at the phi's block",
		file: "verify.ts",
		from: "graph.identifierOf(operand.place),\n                    operand.predecessor,",
		to:   "graph.identifierOf(operand.place),\n                    blockId,",
	},
	// verifySingleAssignment: every place a context store defines counted, not only its own result.
	{
		name: "a context store's every definition counted",
		file: "verify.ts",
		from: "!graph.contextStoreDefines(fn, currentBlock, position, place)\n        ) {",
		to:   "!graph.contextStoreDefines(fn, currentBlock, position, place) &&\n            false\n        ) {",
	},
	// computeDominance: one round of Cooper-Harvey-Kennedy rather than its fixed point, so a block whose
	// dominator only settles once a loop's latch is processed keeps the first round's answer. Over the
	// real reverse postorder computeDominance walks, most graphs settle in one round; 17 of the
	// generated functions need a second, which is what catches this.
	{
		name: "dominance from one round, not a fixed point",
		file: "verify.ts",
		from: "immediate[index] = newImmediate;\n                changed = true;",
		to:   "immediate[index] = newImmediate;\n                changed = false;",
	},
}

// askedCases has cohere's side write every case, and Go cohere's answer to each, returning the cases
// file's path and the answers.
func askedCases(test *testing.T) (string, string) {
	test.Helper()
	seed := int64(generatedSeed)
	var parseError error
	if value := os.Getenv("COHERE_STATIC_SINGLE_ASSIGNMENT_SEED"); value != "" {
		if seed, parseError = strconv.ParseInt(value, 10, 64); parseError != nil {
			test.Fatal(parseError)
		}
	}
	generated := 2000
	if value := os.Getenv("COHERE_STATIC_SINGLE_ASSIGNMENT_GENERATED"); value != "" {
		if generated, parseError = strconv.Atoi(value); parseError != nil {
			test.Fatal(parseError)
		}
	}
	react, absolutePathError := filepath.Abs(filepath.Join("testdata", "react_graphs.txt.gz"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	graphs := []string{react}
	if value := os.Getenv("COHERE_STATIC_SINGLE_ASSIGNMENT_GRAPHS"); value != "" {
		directory, absolutePathError := filepath.Abs(value)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		graphs = append(graphs, directory)
	}
	directory := test.TempDir()
	casesPath := filepath.Join(directory, "cases.txt")
	answersPath := filepath.Join(directory, "answers.txt")
	cohereSide(test, map[string]any{
		"seed":      seed,
		"generated": generated,
		"graphs":    graphs,
		"cases":     casesPath,
		"answers":   answersPath,
	})
	answers, readError := os.ReadFile(answersPath)
	if readError != nil {
		test.Fatal(readError)
	}
	test.Logf("seed %d, %d generated functions, graphs from %q", seed, generated, graphs)
	return casesPath, string(answers)
}

// cohereSide runs testdata/cohere_side_test.go inside cohere's static_single_assignment module, by
// overlay, with a request. The module is one of its own, with no go.work above it that names it, so the
// side runs with GOWORK off.
func cohereSide(test *testing.T, request map[string]any) {
	test.Helper()
	directory := test.TempDir()
	requestPath := filepath.Join(directory, "request.json")
	encoded, encodeError := json.Marshal(request)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(requestPath, encoded, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	binary, module := cohereSideOracle(test)
	command := bounded(test, binary, "-test.run=^TestAdamicPortCases$")
	command.Dir = module
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	if output, combinedOutputError := childguard.CombinedOutput(
		command,
		childguard.Options{},
	); combinedOutputError != nil {
		test.Fatalf("cohere's side: %v\n%s", combinedOutputError, output)
	}
}

// portDirectory copies the port into a directory of its own, with a mutant applied when there is one.
func portDirectory(test *testing.T, applied *mutant) string {
	test.Helper()
	directory := test.TempDir()
	for _, name := range portFiles {
		contents, readError := os.ReadFile(name)
		if readError != nil {
			test.Fatal(readError)
		}
		source := string(contents)
		if applied != nil && applied.file == name {
			if strings.Count(source, applied.from) != 1 {
				test.Fatalf("the mutant %q must change exactly one place in %s", applied.name, name)
			}
			source = strings.Replace(source, applied.from, applied.to, 1)
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return directory
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
func lowered(test *testing.T, path string) *ir.Program {
	test.Helper()
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatalf("Load: %v", loadError)
	}
	result, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		test.Fatalf("Lower: %v", lowerError)
	}
	return result
}

// bounded constructs an unstarted command; its callers use childguard to bound silence.
func bounded(test *testing.T, name string, arguments ...string) *exec.Cmd {
	test.Helper()
	return exec.Command(name, arguments...)
}

func execute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	command := bounded(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := childguard.Run(command, childguard.Options{})
	var exitError *exec.ExitError
	if runError != nil && !errors.As(runError, &exitError) {
		test.Fatalf("running %s: %v", name, runError)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node, through the oracle's runner, with its arguments.
func onNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return execute(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

// onJavaScriptBackend runs the lowered port through the JavaScript backend, on Node.
func onJavaScriptBackend(test *testing.T, program *ir.Program, arguments ...string) run {
	test.Helper()
	path := filepath.Join(test.TempDir(), "program.mjs")
	if writeError := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	return onNode(test, path, arguments...)
}

// leaks returns a report of everything the finished port never let go of, or "": macOS's leaks tool on
// an unsanitized build, or LeakSanitizer on Linux running the sanitized binary again, as the oracle
// checks every fixture.
func leaks(test *testing.T, program *ir.Program, sanitized string, arguments ...string) string {
	test.Helper()
	switch runtime.GOOS {
	case "darwin":
		binary := filepath.Join(test.TempDir(), "port")
		if buildError := native.Build(native.C(program), binary, native.Options{}); buildError != nil {
			test.Fatal(buildError)
		}
		report := execute(test, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode == 0 {
			return ""
		}
		return string(report.stdout)
	case "linux":
		report := execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode == 0 {
			return ""
		}
		return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
	}
	test.Fatalf("no leak check for %s", runtime.GOOS)
	return ""
}

// cohereSideOracle is the test binary go test would build with testdata/cohere_side_test.go laid over cohere's code, a
// product (buildcache.GoTest) built ahead, and the directory go test would run it in.
func cohereSideOracle(test testing.TB) (string, string) {
	test.Helper()
	return oracles.StaticSingleAssignment(test)
}

func TestProduct_StaticSingleAssignmentFindsGoOracle(test *testing.T) {
	test.Parallel()
	cohereSideOracle(test)
}
