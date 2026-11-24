package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	heldStops                                                    heldStopList
	// openTasks is --open-tasks: the tasks open when the run was asked. A pending sidecar naming any other excuses
	// nothing, since a closed task's sidecar hid two live wrong answers for three weeks (#92cnfpf).
	openTasks map[string]bool
	// jsonTypesBinary is --json-types: head's JSON types tool, built from head when it's empty.
	jsonTypesBinary string
}

// agreementProgram is one program of the run and how it was judged: agreeing, ruled, checked-stop, pending,
// disagreeing or void, or refused, or its compile's class when head's compiler failed on it.
type agreementProgram struct {
	Path     string `json:"path"`
	Corpus   string `json:"corpus"`
	Judgment string `json:"judgment"`
	// OutcomeHash is outcomeHash of the program's three runs, on every program judged by what it did, so a diff of two
	// days' reports sees any change in output or stop message, even where the judgment stays. A refused or void
	// program has none.
	OutcomeHash string `json:"outcome_hash,omitempty"`
	// Task is the open task a pending program's sidecar names, without its #. The run can't ask whether it is still
	// open, so the reader of the report does.
	Task string `json:"task,omitempty"`
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
	// Compared is Agreeing, Ruled, CheckedStops, Pending and Disagreeing together: every accepted program the run
	// judged. A void program was run and proved nothing, so it is in none of them.
	Compared     int `json:"compared"`
	Agreeing     int `json:"agreeing"`
	Ruled        int `json:"ruled"`
	CheckedStops int `json:"checked_stops"`
	Pending      int `json:"pending"`
	Disagreeing  int `json:"disagreeing"`
	MemoryUnsafe int `json:"memory_unsafe"`
	Refused      int `json:"refused"`
	// TrappedPrograms counts the programs a run of which met a first-link trap: a program lowered in trap mode
	// (lower's Options.FirstLinkTraps), measurement and never a product. It's classed apart, never agreeing or a
	// checked stop, and in none of Compared, and any fails the run.
	TrappedPrograms int `json:"trapped_programs"`
	// TimedOutRuns, RetriedRuns and StillTimedOutRuns count runtime timeouts as the delta's report does (#pajbaba).
	TimedOutRuns      int                        `json:"timed_out_runs"`
	RetriedRuns       int                        `json:"retried_runs"`
	StillTimedOutRuns int                        `json:"still_timed_out_runs"`
	Disagreements     []agreementDisagreement    `json:"disagreements"`
	CompilerFailures  []agreementCompilerFailure `json:"compiler_failures"`
	StaleRuntimeStops []string                   `json:"stale_runtime_stops"`
	// ClosedSidecars is every program that would disagree whose pending sidecar names a task --open-tasks doesn't
	// hold, so it excused nothing and the program was judged as if it had none.
	ClosedSidecars []string           `json:"closed_sidecars"`
	Programs       []agreementProgram `json:"programs"`
	Verdict        string             `json:"verdict"`
	Void           bool               `json:"void"`
	VoidReason     string             `json:"void_reason,omitempty"`
}

// firstLinkTrapPrefix is what a first-link trap writes to stderr as it ends a program lowered in trap mode.
const firstLinkTrapPrefix = "adamic: first-link trap:"

// firstLinkTrapRun reports whether any of a program's runs met a first-link trap.
func firstLinkTrapRun(runs ...observation) bool {
	for _, run := range runs {
		if strings.Contains(run.Stderr, firstLinkTrapPrefix) {
			return true
		}
	}
	return false
}

// nativeWarmLimit bounds the one native build that makes the runtime's archives before the run's own builds: the
// archives of a sanitized build from a cold cache, which take a box minutes, never a program's 45 s.
const nativeWarmLimit = 15 * time.Minute

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
	if _, _, err := parseShard(shard); err != nil {
		return err
	}
	// A sidecar excuses a program only while its task is open, and the sealed run can't ask, so the asker says.
	if !given["open-tasks"] {
		return fmt.Errorf("--agreement requires --open-tasks: a pending sidecar excuses a program only while its task is open, and the run can't ask")
	}
	return nil
}

// runAgreement builds head's compiler, classifies every program of the manifest's shard by head alone, and runs every
// one head accepts through the delta's own runner, natively under the sanitizers, judging each as the delta judges a
// newly accepted program, then counting apart a listed held stop and a program whose pending sidecar names a task.
// A disagreement, a compiler failure or a stale run-time stop exits 1, as the delta's red
// does; a run that proved nothing for a timeout or a missing clang exits 3, as the delta's void does.
func runAgreement(agreement agreementRun) error {
	result := agreementReport{Shard: agreement.shard, Verdict: "pass", Disagreements: []agreementDisagreement{}, CompilerFailures: []agreementCompilerFailure{}, StaleRuntimeStops: []string{}, ClosedSidecars: []string{}, Programs: []agreementProgram{}}
	// progress says on stderr where a run is, with its seconds so far, as it goes: a whole corpus takes a box many
	// minutes, and a run its job's ceiling stops would otherwise end with nothing said about where its time went.
	started := time.Now()
	progress := func(format string, arguments ...any) {
		fmt.Fprintf(os.Stderr, "agreement: %.0f s: %s\n", time.Since(started).Seconds(), fmt.Sprintf(format, arguments...))
	}
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
	progress("head's compiler is built; classifying %d programs on %d workers", len(programs), agreement.workers)
	classifyPrograms(headTree, lower, lower, programs, agreement.workers, agreement.compileLimit)
	accepted := 0
	for _, record := range programs {
		if record.Class == "accepted-by-both" {
			accepted++
		}
	}
	progress("classified: %d accepted of %d", accepted, len(programs))
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
	runner := &programRunner{headTree: headTree, scratch: scratch, compiler: compiler, limit: agreement.limit, nativeFlags: []string{"--sanitize"}, nativeEnv: []string{"ASAN_OPTIONS=detect_leaks=0"}, negatives: agreement.negatives, runtimeStops: agreement.runtimeStops, divergences: agreement.divergences, jsonTypes: jsonTypesTool(agreement.root, scratch, result.Head, agreement.jsonTypesBinary), compileLimit: agreement.compileLimit}
	// One native build first, with room to make the runtime's archives, so no program's own build has to.
	if clang && accepted > 0 {
		progress("building the native runtime")
		if err := runner.warmNative(nativeWarmLimit); err != nil {
			voids = append(voids, err.Error())
			clang = false
		}
	}
	// Accepted programs run on agreement.workers at once, a corpus being thousands of sanitized builds. Each worker
	// writes only its own program's record, stale and failed slot, and scratch files named by the program's index; the
	// counting and the lines wait for the pool, in manifest order, so the output is the same whatever order the
	// workers finish in. A timed-out run's retry here runs beside the other workers' runs, never wholly alone.
	stale := make([]bool, len(programs))
	failed := make([]error, len(programs))
	var finished atomic.Int64
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
				if done := finished.Add(1); done%250 == 0 {
					progress("%d of %d accepted programs run", done, accepted)
				}
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
	progress("every accepted program run: %d", finished.Load())
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
			if firstLinkTrapRun(*record.Node, *record.JavaScript, *record.Native) {
				result.TrappedPrograms++
				judged.Judgment = "first-link trap"
				lines = append(lines, fmt.Sprintf("first-link trap %s: a run wrote %q, so the program was lowered in trap mode and proves nothing", record.Path, firstLinkTrapPrefix))
				break
			}
			if stale[index] {
				result.StaleRuntimeStops = append(result.StaleRuntimeStops, record.Path)
				lines = append(lines, fmt.Sprintf("stale run-time stop %s: it agrees with Node, so its entry leaves the list", record.Path))
			}
			judged.OutcomeHash = outcomeHash(headTree, *record.Node, *record.JavaScript, *record.Native)
			// A sidecar is read only for a program that would otherwise disagree: an agreeing program stays agreeing
			// whether or not one sits beside it.
			if record.disagrees() {
				judged.Task, err = pendingTask(headTree, record.Path)
				if err != nil {
					return err
				}
				// A sidecar naming a task that isn't open excuses nothing: the program is judged as if it had none.
				if judged.Task != "" && !agreement.openTasks[judged.Task] {
					result.ClosedSidecars = append(result.ClosedSidecars, record.Path)
					lines = append(lines, fmt.Sprintf("closed sidecar %s: names #%s, which isn't open, so it excuses nothing", record.Path, judged.Task))
					judged.Task = ""
				}
			}
			switch {
			case *record.Agree:
				result.Agreeing++
				judged.Judgment = "agreeing"
			case record.ruled():
				result.Ruled++
				judged.Judgment = "ruled"
			case agreement.heldStops.holds(record.Path, headTree, *record.Node, *record.JavaScript, *record.Native):
				result.CheckedStops++
				judged.Judgment = "checked-stop"
			case judged.Task != "" && !memoryUnsafe(*record.Native):
				// The sidecar excuses whatever the program does while its task is open, a timeout included, so a
				// pending program that hangs never voids the run. A sanitizer's report or a signal is never excused:
				// a sidecar filing a memory error as a library gap is how #sq84qqh's five release segfaults hid.
				result.Pending++
				judged.Judgment = "pending"
				lines = append(lines, fmt.Sprintf("pending %s: awaits #%s", record.Path, judged.Task))
			case record.VoidReason != "":
				voids = append(voids, record.Path+": "+record.VoidReason)
				judged.Judgment = "void"
				judged.OutcomeHash = ""
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
	result.Compared = result.Agreeing + result.Ruled + result.CheckedStops + result.Pending + result.Disagreeing
	result.TimedOutRuns, result.RetriedRuns, result.StillTimedOutRuns = runner.timedOutRuns, runner.retriedRuns, runner.stillTimedOutRuns
	red := result.Disagreeing > 0 || len(result.CompilerFailures) > 0 || len(result.StaleRuntimeStops) > 0 || result.TrappedPrograms > 0
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
	// The delta retries a timed-out run alone; here the retry runs beside the other workers' runs, so the line says so
	// and never reads as the delta's.
	fmt.Fprintf(os.Stderr, "runtime timeouts: %d runs timed out, %d retried once beside the other workers, %d timed out again\n", result.TimedOutRuns, result.RetriedRuns, result.StillTimedOutRuns)
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
	summary := fmt.Sprintf("agreement: compared %d, agreeing %d, ruled %d, checked stops %d, pending %d, disagreeing %d, memory-unsafe %d, refused %d", result.Compared, result.Agreeing, result.Ruled, result.CheckedStops, result.Pending, result.Disagreeing, result.MemoryUnsafe, result.Refused)
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

// pendingSidecar is a pending sidecar's message, the form internal/oracle's review test holds it to (reviewPending in
// internal/oracle/review_test.go): awaits #<task>: <reason>, or awaits <branch>: <reason>.
var pendingSidecar = regexp.MustCompile(`^awaits (#[0-9a-z]{6,9}|[A-Za-z0-9_./-]+): [^\r\n]+$`)

// pendingTask is the task a program's pending sidecar names, without its #, or "" when it has no sidecar, or one that
// names a branch or isn't of the form. The sidecar is the program's name without its extension, plus .pending, beside
// it in head. Only a task excuses: the run is sealed and can't ask whether it is open, so the report names it and its
// reader checks. A branch names no task to check, so its program disagrees.
func pendingTask(headTree, path string) (string, error) {
	data, err := os.ReadFile(filepath.Join(headTree, strings.TrimSuffix(path, filepath.Ext(path))+".pending"))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	match := pendingSidecar.FindStringSubmatch(strings.TrimSuffix(string(data), "\n"))
	if match == nil || !strings.HasPrefix(match[1], "#") {
		return "", nil
	}
	return strings.TrimPrefix(match[1], "#"), nil
}

// openTaskID is a task's id as --open-tasks writes it: the form a pending sidecar names, without its #.
var openTaskID = regexp.MustCompile(`^[0-9a-z]{6,9}$`)

// readOpenTasks reads --open-tasks: one task id a line, without its #, blank lines skipped. An empty file is a run at
// which no task is open, so no sidecar excuses anything. A line that isn't an id is refused, so a list written in the
// wrong form, which would otherwise read as every task closed or open, stops the run before it starts.
func readOpenTasks(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading --open-tasks: %w", err)
	}
	open := map[string]bool{}
	for number, line := range strings.Split(string(data), "\n") {
		id := strings.TrimSpace(line)
		if id == "" {
			continue
		}
		if !openTaskID.MatchString(id) {
			return nil, fmt.Errorf("--open-tasks %s:%d: %q is no task id (one a line, without its #)", path, number+1, id)
		}
		open[id] = true
	}
	return open, nil
}

// outcomeHash is hex SHA-256 over Node's, then JavaScript's, then native's exit, stdout and stderr, in that order.
// Each field is written as its length, a colon and its bytes, so no output can be read as the end of one field and the
// start of the next, and a byte moved from one stream to another changes the hash. The checkout the run judged, root,
// is written {root} in stdout and stderr first, as a held stop's row writes it: a program whose stop names its own
// file names it inside the run's random scratch directory, so without it that program's hash would change every day
// and bury the changes the hash is there to show.
func outcomeHash(root string, node, javascript, native observation) string {
	// An empty root would write {root} between every byte, so only a named checkout is written.
	rooted := func(output string) string {
		if root == "" {
			return output
		}
		return strings.ReplaceAll(output, root, "{root}")
	}
	digest := sha256.New()
	for _, run := range []observation{node, javascript, native} {
		for _, field := range []string{strconv.Itoa(run.Exit), rooted(run.Stdout), rooted(run.Stderr)} {
			fmt.Fprintf(digest, "%d:%s", len(field), field)
		}
	}
	return hex.EncodeToString(digest.Sum(nil))
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
