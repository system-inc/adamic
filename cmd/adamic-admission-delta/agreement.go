package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The agreement mode runs every program head accepts, with no base. The delta runs only what head newly accepts, so a
// wrong answer already on main, a native segfault where Node prints a value, is invisible to every head's gate; this
// mode is where it shows.

// agreementRun is what the agreement mode reads of the command's flags and the ruled lists run has read.
type agreementRun struct {
	root, head, headBinary, headLowerBinary, manifestPath, shard string
	asJSON                                                       bool
	limit, compileLimit                                          time.Duration
	workers                                                      int
	negatives                                                    negativeWitnessList
	runtimeStops                                                 runtimeNotYetStopList
	divergences                                                  ruledBackendDivergenceList
}

// agreementProgram is one program of the run and how it was judged: agreeing, ruled, disagreeing or void, or refused,
// or its compile's class when head's compiler failed on it.
type agreementProgram struct {
	Path     string `json:"path"`
	Corpus   string `json:"corpus"`
	Judgment string `json:"judgment"`
}

// agreementDisagreement is a program that disagrees with Node, with what each of the three runs did and what each
// backend's compile did.
type agreementDisagreement struct {
	Path              string      `json:"path"`
	Corpus            string      `json:"corpus"`
	Node              observation `json:"node"`
	JavaScript        observation `json:"javascript"`
	Native            observation `json:"native"`
	JavaScriptCompile observation `json:"javascript_compile"`
	NativeCompile     observation `json:"native_compile"`
	MemoryUnsafe      bool        `json:"memory_unsafe"`
}

// agreementCompilerFailure is a program head's compiler crashed on, failed on with an unrecognised diagnostic, or
// timed out on, so it was never run.
type agreementCompilerFailure struct {
	Path    string      `json:"path"`
	Class   string      `json:"class"`
	Compile observation `json:"compile"`
}

type agreementReport struct {
	Head           string `json:"head"`
	ManifestBlob   string `json:"manifest_blob"`
	Shard          string `json:"shard"`
	CompleteCorpus bool   `json:"complete_corpus"`
	// Compared is Agreeing, Ruled and Disagreeing together: every accepted program the run judged. A void program was
	// run and proved nothing, so it is in none of them.
	Compared     int `json:"compared"`
	Agreeing     int `json:"agreeing"`
	Ruled        int `json:"ruled"`
	Disagreeing  int `json:"disagreeing"`
	MemoryUnsafe int `json:"memory_unsafe"`
	Refused      int `json:"refused"`
	// TimedOutRuns, RetriedRuns and StillTimedOutRuns count runtime timeouts as the delta's report does (#pajbaba).
	TimedOutRuns      int                        `json:"timed_out_runs"`
	RetriedRuns       int                        `json:"retried_runs"`
	StillTimedOutRuns int                        `json:"still_timed_out_runs"`
	Disagreements     []agreementDisagreement    `json:"disagreements"`
	CompilerFailures  []agreementCompilerFailure `json:"compiler_failures"`
	StaleRuntimeStops []string                   `json:"stale_runtime_stops"`
	Programs          []agreementProgram         `json:"programs"`
	Verdict           string                     `json:"verdict"`
	Void              bool                       `json:"void"`
	VoidReason        string                     `json:"void_reason,omitempty"`
}

// agreementIgnores is every flag of the delta's that the agreement mode never reads, in the order it names them. Each
// is refused rather than passed over, so a run can't look as if a base, a filter or a budget shaped it.
var agreementIgnores = []string{"base", "base-binary", "base-lower-binary", "corpus", "corpus-filter", "budget", "seed", "manifest-generator", "manifest-generator-revision"}

// agreementUsage says what is wrong with an agreement run's flags, before anything runs: it judges head alone, so it
// takes no base and needs head named, and it reads the manifest's corpus, sharded k/n.
func agreementUsage(given map[string]bool, manifestPath, shard string) error {
	if given["base"] {
		return fmt.Errorf("--agreement refuses --base: it judges head alone, against Node")
	}
	for _, name := range agreementIgnores {
		if given[name] {
			return fmt.Errorf("--agreement doesn't read --%s", name)
		}
	}
	if !given["head"] {
		return fmt.Errorf("--agreement requires --head")
	}
	if manifestPath == "" {
		return fmt.Errorf("--agreement requires --manifest")
	}
	_, _, err := parseShard(shard)
	return err
}

// runAgreement builds head's compiler, classifies every program of the manifest's shard by head alone, and runs every
// one head accepts through the delta's own runner, natively under the sanitizers, judging each as the delta judges a
// newly accepted program. A disagreement, a compiler failure or a stale run-time stop exits 1, as the delta's red
// does; a run that proved nothing for a timeout or a missing clang exits 3, as the delta's void does.
func runAgreement(agreement agreementRun) error {
	result := agreementReport{Shard: agreement.shard, Verdict: "pass", Disagreements: []agreementDisagreement{}, CompilerFailures: []agreementCompilerFailure{}, StaleRuntimeStops: []string{}, Programs: []agreementProgram{}}
	var err error
	result.Head, err = git(agreement.root, "rev-parse", agreement.head+"^{commit}")
	if err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "admission-agreement-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	headTree := filepath.Join(scratch, "head")
	if _, err = git(agreement.root, "worktree", "add", "--detach", headTree, result.Head); err != nil {
		return err
	}
	defer git(agreement.root, "worktree", "remove", "--force", headTree)
	if err = prepareCheckout(agreement.root, headTree); err != nil {
		return err
	}
	// Head's compiler and its lowering companion, as the delta takes head's: supplied, or built from head.
	var compiler string
	if agreement.headBinary != "" {
		compiler, err = filepath.Abs(agreement.headBinary)
	} else {
		compiler, err = cachedCompiler(agreement.root, scratch, result.Head, "head-build")
	}
	if err != nil {
		return err
	}
	lower := compiler
	if agreement.headLowerBinary != "" {
		lower, err = filepath.Abs(agreement.headLowerBinary)
	} else if agreement.headBinary != "" {
		lower, err = cachedCompiler(agreement.root, scratch, result.Head, "head-lower-build")
	}
	if err != nil {
		return err
	}
	data, err := os.ReadFile(agreement.manifestPath)
	if err != nil {
		return err
	}
	var m manifest
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}
	if m.SHA != result.Head {
		return fmt.Errorf("manifest sha %s differs from head %s", m.SHA, result.Head)
	}
	result.ManifestBlob, err = git(agreement.root, "hash-object", agreement.manifestPath)
	if err != nil {
		return err
	}
	shardIndex, shardCount, err := parseShard(agreement.shard)
	if err != nil {
		return err
	}
	result.CompleteCorpus = shardCount == 1
	programs, _, err := verifyPrograms(agreement.root, headTree, result.Head, shardByIndex(m.Corpora, shardIndex, shardCount), agreement.negatives, agreement.runtimeStops)
	if err != nil {
		return err
	}
	// Head lowers each program once: with one compiler given as both, a program's class is accepted-by-both,
	// refused-by-both, or a compiler failure.
	classifyPrograms(headTree, lower, lower, programs, agreement.workers, agreement.compileLimit)
	var voids []string
	// The native build resolves clang as toolchain.Path does; without it every native observation would be a
	// compilation failure head didn't cause, so nothing runs and the run is void, as in the delta.
	clang := true
	for _, record := range programs {
		if record.Class == "accepted-by-both" {
			if _, e := toolchain.Path("clang"); e != nil {
				voids = append(voids, "native toolchain missing: "+e.Error())
				clang = false
			}
			break
		}
	}
	runner := &programRunner{headTree: headTree, scratch: scratch, compiler: compiler, limit: agreement.limit, nativeFlags: []string{"--sanitize"}, nativeEnv: []string{"ASAN_OPTIONS=detect_leaks=0"}, negatives: agreement.negatives, runtimeStops: agreement.runtimeStops, divergences: agreement.divergences}
	// Accepted programs run on agreement.workers at once, a corpus being thousands of sanitized builds. Each worker
	// writes only its own program's record, stale and failed slot, and scratch files named by the program's index; the
	// counting and the lines wait for the pool, in manifest order, so the output is the same whatever order the
	// workers finish in. A timed-out run's retry here runs beside the other workers' runs, never wholly alone.
	stale := make([]bool, len(programs))
	failed := make([]error, len(programs))
	jobs := make(chan int)
	var group sync.WaitGroup
	for worker := 0; worker < agreement.workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for index := range jobs {
				record := &programs[index]
				record.Sampled = true
				still, err := runner.runProgram(record, fmt.Sprintf("program-%d", index))
				if err != nil {
					failed[index] = err
					continue
				}
				stale[index], failed[index] = runner.judgeProgram(record, index, still)
			}
		}()
	}
	for index := range programs {
		if clang && programs[index].Class == "accepted-by-both" {
			jobs <- index
		}
	}
	close(jobs)
	group.Wait()
	for index := range programs {
		if failed[index] != nil {
			return failed[index]
		}
	}
	var lines []string
	for index := range programs {
		record := &programs[index]
		judged := agreementProgram{Path: record.Path, Corpus: record.Corpus, Judgment: record.Class}
		switch {
		case record.Class == "refused-by-both":
			result.Refused++
			judged.Judgment = "refused"
		case record.Class == "compiler-timeout":
			voids = append(voids, record.Path+": compile timeout")
			judged.Judgment = "void"
		case record.Class != "accepted-by-both":
			result.CompilerFailures = append(result.CompilerFailures, agreementCompilerFailure{Path: record.Path, Class: record.Class, Compile: record.Head})
			lines = append(lines, fmt.Sprintf("compiler failure %s: %s, exit %s", record.Path, record.Class, exitText(record.Head)))
		case !clang:
			judged.Judgment = "void"
		default:
			if stale[index] {
				result.StaleRuntimeStops = append(result.StaleRuntimeStops, record.Path)
				lines = append(lines, fmt.Sprintf("stale run-time stop %s: it agrees with Node, so its entry leaves the list", record.Path))
			}
			switch {
			case *record.Agree:
				result.Agreeing++
				judged.Judgment = "agreeing"
			case record.ruled():
				result.Ruled++
				judged.Judgment = "ruled"
			case record.VoidReason != "":
				voids = append(voids, record.Path+": "+record.VoidReason)
				judged.Judgment = "void"
			default:
				result.Disagreeing++
				judged.Judgment = "disagreeing"
				unsafe := memoryUnsafe(*record.Native)
				if unsafe {
					result.MemoryUnsafe++
				}
				result.Disagreements = append(result.Disagreements, agreementDisagreement{Path: record.Path, Corpus: record.Corpus, Node: *record.Node, JavaScript: *record.JavaScript, Native: *record.Native, JavaScriptCompile: *record.JavaScriptCompile, NativeCompile: *record.NativeCompile, MemoryUnsafe: unsafe})
				lines = append(lines, disagreementLine(*record, unsafe))
			}
		}
		result.Programs = append(result.Programs, judged)
	}
	result.Compared = result.Agreeing + result.Ruled + result.Disagreeing
	result.TimedOutRuns, result.RetriedRuns, result.StillTimedOutRuns = runner.timedOutRuns, runner.retriedRuns, runner.stillTimedOutRuns
	red := result.Disagreeing > 0 || len(result.CompilerFailures) > 0 || len(result.StaleRuntimeStops) > 0
	result.VoidReason = strings.Join(voids, "; ")
	switch {
	case red:
		result.Verdict = "fail"
	case len(voids) > 0:
		result.Verdict, result.Void = "void", true
	case !result.CompleteCorpus:
		// A shard alone never stands in for the whole corpus, as the delta's partial run doesn't.
		result.Verdict = "partial"
	}
	fmt.Fprintf(os.Stderr, "runtime timeouts: %d runs timed out, %d retried alone, %d timed out again\n", result.TimedOutRuns, result.RetriedRuns, result.StillTimedOutRuns)
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
	if agreement.asJSON {
		if err = json.NewEncoder(os.Stdout).Encode(result); err != nil {
			return err
		}
	} else {
		fmt.Printf("verdict: %s\n", result.Verdict)
		for _, judged := range result.Programs {
			fmt.Printf("%s %s\n", judged.Judgment, judged.Path)
		}
	}
	summary := fmt.Sprintf("agreement: compared %d, agreeing %d, ruled %d, disagreeing %d, memory-unsafe %d, refused %d", result.Compared, result.Agreeing, result.Ruled, result.Disagreeing, result.MemoryUnsafe, result.Refused)
	// exitCode prints a returned error last on stderr, so a red or a void run returns the summary as its error's last
	// line, and the summary stays the last line whatever the exit.
	if red {
		return errors.New(summary)
	}
	if result.Void {
		return fmt.Errorf("%w: %s\n%s", errVoid, result.VoidReason, summary)
	}
	fmt.Fprintln(os.Stderr, summary)
	return nil
}

// sanitizerNames are the sanitizers whose report in a native run's stderr makes a disagreement memory-unsafe. Each
// names itself in its report: AddressSanitizer's ERROR line, UndefinedBehaviorSanitizer's SUMMARY. LeakSanitizer is not
// among them: the native run has detect_leaks=0, as the oracle's comparison run has (internal/oracle's
// nativelyUncached), since a program that stops with exit 70 still holds what it held, and a leak report would then
// break every ruled stop's exact stderr.
var sanitizerNames = []string{"AddressSanitizer", "UndefinedBehaviorSanitizer"}

// memoryUnsafe says a native run shows a sanitizer's report, or died on a signal, which execute records as a crash
// (a negative exit with no other error). A timeout is never one: execute records it as a timeout.
func memoryUnsafe(native observation) bool {
	if native.Error == "crash" {
		return true
	}
	for _, name := range sanitizerNames {
		if strings.Contains(native.Stderr, name) {
			return true
		}
	}
	return false
}

// disagreementLine names a disagreeing program, then Node's, JavaScript's and native's exit codes, then, for stdout
// and for stderr where the three differ, the first line at which they do, as each has it, and memory-unsafe when it
// is.
func disagreementLine(record entry, unsafe bool) string {
	line := fmt.Sprintf("disagreeing %s: exit node %s, javascript %s, native %s", record.Path, exitText(*record.Node), exitText(*record.JavaScript), exitText(*record.Native))
	for _, stream := range []struct {
		name                     string
		node, javascript, native string
	}{
		{"stdout", record.Node.Stdout, record.JavaScript.Stdout, record.Native.Stdout},
		{"stderr", record.Node.Stderr, record.JavaScript.Stderr, record.Native.Stderr},
	} {
		if shown := firstDifferentLine(stream.node, stream.javascript, stream.native); shown != nil {
			line += fmt.Sprintf("; %s node %s, javascript %s, native %s", stream.name, shown[0], shown[1], shown[2])
		}
	}
	if unsafe {
		line += "; memory-unsafe"
	}
	return line
}

// exitText is a run's exit code, and its execution error when it had one, so a timeout or a failed build reads as
// one rather than as a bare -1.
func exitText(o observation) string {
	if o.Error != "" {
		return strconv.Itoa(o.Exit) + " (" + o.Error + ")"
	}
	return strconv.Itoa(o.Exit)
}

// firstDifferentLine is the first line at which outputs aren't all alike, as each has it, quoted, with "end" for an
// output that ended before it; nil when they are the same bytes. An output with a trailing newline holds one more,
// empty, line than one without, so the two differ there.
func firstDifferentLine(outputs ...string) []string {
	split := make([][]string, len(outputs))
	longest := 0
	for index, output := range outputs {
		split[index] = strings.Split(output, "\n")
		longest = max(longest, len(split[index]))
	}
	for at := 0; at < longest; at++ {
		shown := make([]string, len(outputs))
		for index := range outputs {
			shown[index] = "end"
			if at < len(split[index]) {
				shown[index] = strconv.Quote(split[index][at])
			}
		}
		for _, other := range shown[1:] {
			if other != shown[0] {
				return shown
			}
		}
	}
	return nil
}
