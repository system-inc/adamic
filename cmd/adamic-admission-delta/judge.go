package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// programRunner runs a program in head on Node, the JavaScript backend and a native build, and judges what it saw
// against Node and the ruled lists. The delta's newly accepted programs and the agreement mode's accepted ones both
// run and are judged through it, so the two hold a program to one rule.
type programRunner struct {
	headTree, scratch, compiler string
	limit                       time.Duration
	// nativeFlags follow -o <out> on every native build: --sanitize for the agreement mode, none for the delta, whose
	// native build is the release one.
	nativeFlags []string
	// compileLimit bounds a native build, a compile, where limit bounds a run; a zero one is limit, as the delta has
	// always run its release builds. The agreement mode's sanitized builds under limit's 10 s were each killed while
	// the build cache was still making the runtime's archives, so on a box whose cache started cold none ever
	// finished and every native observation timed out (#x0zh2gq: 954 of 954 on Lunatic, Golem and Home alike).
	compileLimit time.Duration
	// nativeEnv is added to the native program's environment, and only its: the agreement mode's
	// ASAN_OPTIONS=detect_leaks=0, none for the delta.
	nativeEnv    []string
	negatives    negativeWitnessList
	runtimeStops runtimeNotYetStopList
	divergences  ruledBackendDivergenceList
	// jsonTypes is head's oracle/json_types.go built, the tool oracle/node.mjs runs to give a program that decodes or
	// encodes JSON its type descriptors, asked for on the first such program and kept. Without it node.mjs would go
	// run the tool in head's checkout, which the delta lays without cohere, so Node's run failed on the workspace
	// before the program ran and every JSON program disagreed (#x0zh2gq, 24 on the first whole-corpus run).
	jsonTypes     func() (string, error)
	jsonTypesOnce sync.Once
	jsonTypesPath string
	jsonTypesErr  error
	// counting guards the three counts below, since the agreement mode runs programs on several workers at once.
	counting sync.Mutex
	// timedOutRuns is how many runs, a program's or a repair's, hit a runtime timeout; retriedRuns how many of them
	// were retried alone; stillTimedOutRuns how many timed out again (#pajbaba).
	timedOutRuns, retriedRuns, stillTimedOutRuns int
}

// observe runs one program, a path in head or an absolute one, on Node and both backends, its outputs named by name
// in scratch.
func (runner *programRunner) observe(name, path string) (node, jsCompile, javascript, nativeCompile, native observation, err error) {
	source := path
	if !filepath.IsAbs(source) {
		source = filepath.Join(runner.headTree, path)
	}
	text, err := os.ReadFile(source)
	if err != nil {
		return
	}
	var nodeEnv []string
	if bytes.Contains(text, []byte("decodeJson")) || bytes.Contains(text, []byte("encodeJson")) {
		var tool string
		if tool, err = runner.jsonTypesTool(); err != nil {
			return
		}
		nodeEnv = []string{"ADAMIC_JSON_TYPES=" + tool}
	}
	node = executeWith(runner.headTree, runner.limit, nodeEnv, "node", "--disable-warning=ExperimentalWarning", filepath.Join(runner.headTree, "oracle/node.mjs"), source)
	jsCompile = execute(runner.headTree, runner.limit, runner.compiler, "js", path)
	jsPath := filepath.Join(runner.scratch, name+".mjs")
	javascript = observation{Exit: -1, Error: "JavaScript compilation failed"}
	if jsCompile.Exit == 0 && jsCompile.Error == "" {
		if err = os.WriteFile(jsPath, []byte(jsCompile.Stdout), 0600); err != nil {
			return
		}
		javascript = execute(runner.headTree, runner.limit, "node", "--disable-warning=ExperimentalWarning", filepath.Join(runner.headTree, "oracle/node.mjs"), jsPath)
	}
	jsCompile.Stdout = ""
	nativePath := filepath.Join(runner.scratch, name)
	nativeCompile = execute(runner.headTree, runner.nativeCompileLimit(), runner.compiler, append([]string{"build", path, "-o", nativePath}, runner.nativeFlags...)...)
	native = observation{Exit: -1, Error: "native compilation failed"}
	if nativeCompile.Exit == 0 && nativeCompile.Error == "" {
		native = executeWith(runner.headTree, runner.limit, runner.nativeEnv, nativePath)
	}
	return
}

// nativeCompileLimit is how long one native build may take.
func (runner *programRunner) nativeCompileLimit() time.Duration {
	if runner.compileLimit > 0 {
		return runner.compileLimit
	}
	return runner.limit
}

// warmNative builds one program of a single line natively, under limit, before any program of the run builds, so
// the build cache makes the runtime's archives once rather than under every program's own compile limit, where
// a cold cache's first build can't finish and caches nothing. An error names what the build said.
func (runner *programRunner) warmNative(limit time.Duration) error {
	source := filepath.Join(runner.scratch, "warm.a")
	if err := os.WriteFile(source, []byte("console.log('warm');\n"), 0600); err != nil {
		return err
	}
	built := execute(runner.headTree, limit, runner.compiler, append([]string{"build", source, "-o", filepath.Join(runner.scratch, "warm")}, runner.nativeFlags...)...)
	if built.Exit != 0 || built.Error != "" {
		return fmt.Errorf("the native runtime didn't build in %s: exit %d %s %s", limit, built.Exit, built.Error, strings.TrimSpace(built.Stderr))
	}
	return nil
}

// jsonTypesTool is the JSON types tool, made once for every worker that asks: node.mjs looks for decodeJson or
// encodeJson in a program's text, and so does observe, so a program without JSON never waits on the build.
func (runner *programRunner) jsonTypesTool() (string, error) {
	runner.jsonTypesOnce.Do(func() {
		if runner.jsonTypes == nil {
			runner.jsonTypesErr = fmt.Errorf("admission delta: a program decodes or encodes JSON and no JSON types tool was made")
			return
		}
		runner.jsonTypesPath, runner.jsonTypesErr = runner.jsonTypes()
	})
	return runner.jsonTypesPath, runner.jsonTypesErr
}

// observeAlone is observe with one retry: a run that hit a runtime timeout under the gate's load runs once more,
// alone, before its timeout counts, so a load timeout never reads as a red (#pajbaba). still names every command
// that timed out again; first names every one that timed out the first time.
func (runner *programRunner) observeAlone(name, path string) (node, jsCompile, javascript, nativeCompile, native observation, first, still []string, err error) {
	node, jsCompile, javascript, nativeCompile, native, err = runner.observe(name, path)
	if err != nil {
		return
	}
	if platformStopped(nativeCompile) {
		return
	}
	first = timedOut(node, jsCompile, javascript, nativeCompile, native)
	if len(first) == 0 {
		return
	}
	runner.counting.Lock()
	runner.timedOutRuns++
	runner.retriedRuns++
	runner.counting.Unlock()
	node, jsCompile, javascript, nativeCompile, native, err = runner.observe(name+"-alone", path)
	if err != nil {
		return
	}
	still = timedOut(node, jsCompile, javascript, nativeCompile, native)
	if len(still) > 0 {
		runner.counting.Lock()
		runner.stillTimedOutRuns++
		runner.counting.Unlock()
	}
	return
}

// runProgram runs record's program, retried alone after a runtime timeout, and records what each command did on it.
// It returns every command that timed out again.
func (runner *programRunner) runProgram(record *entry, name string) ([]string, error) {
	node, jsCompile, javascript, nativeCompile, native, first, still, err := runner.observeAlone(name, record.Path)
	if err != nil {
		return nil, err
	}
	record.TimedOut = append(record.TimedOut, first...)
	record.Retried = len(first) > 0
	record.Node = &node
	record.JavaScriptCompile = &jsCompile
	record.JavaScript = &javascript
	record.NativeCompile = &nativeCompile
	record.Native = &native
	return still, nil
}

// judgeProgram holds record's run, as runProgram recorded it, to Node: it agrees, or it is a listed negative
// witness, a listed run-time stop or a ruled backend divergence, each exactly as ruled, or it disagrees. A
// disagreement whose run, or whose repair's, still timed out after its retry is void, its reason in VoidReason, and
// never charged as a red. stale says a listed run-time stop now runs to the end and agrees, so its entry leaves the
// list. index names the repair's files in scratch.
func (runner *programRunner) judgeProgram(record *entry, index int, still []string) (stale bool, err error) {
	source := filepath.Join(runner.headTree, record.Path)
	data, err := os.ReadFile(source)
	if err != nil {
		return false, err
	}
	// repairStill is every repair command that timed out again: the witness is then unjudged, so the program is void,
	// never charged as a red.
	var repairStill []string
	witness, listed := runner.negatives.Witnesses[contentHash(data)]
	agree := outputsAgree(*record.Node, *record.JavaScript, *record.Native)
	if listed {
		if e := validateWitnessSource(data, witness); e != nil {
			return false, fmt.Errorf("%s: %w", record.Path, e)
		}
		if witness.Form == castFailedForm && filepath.Ext(record.Path) != ".ts" {
			return false, fmt.Errorf("%s: a cast_failed witness is .ts alone; in .a its cast is refused", record.Path)
		}
		record.NegativeWitness = !agree && acceptsNegativeWitness(witness, runner.headTree, *record.Node, *record.JavaScript, *record.Native)
		if record.NegativeWitness && repairRuns(witness.Form) {
			// The repair is the stop's mutant: storing a value of the declared type, it agrees with Node. It keeps the
			// witness's extension, so it compiles in the witness's own mode: a .ts witness's repair as .ts.
			repairPath := filepath.Join(runner.scratch, fmt.Sprintf("repair-%d%s", index, filepath.Ext(source)))
			if e := os.WriteFile(repairPath, repairedSource(data, witness), 0600); e != nil {
				return false, e
			}
			repairNode, _, repairJavaScript, _, repairNative, repairFirst, repairTimedOut, e := runner.observeAlone(fmt.Sprintf("repair-%d", index), repairPath)
			if e != nil {
				return false, e
			}
			for _, command := range repairFirst {
				record.TimedOut = append(record.TimedOut, "repair "+command)
			}
			record.Retried = record.Retried || len(repairFirst) > 0
			repairStill = repairTimedOut
			repairAgree := outputsAgree(repairNode, repairJavaScript, repairNative)
			record.RepairAgree = &repairAgree
			record.NegativeWitness = repairAgree
		}
	}
	if stop, ruled := runner.runtimeStops.Stops[contentHash(data)]; ruled {
		// A listed program that now runs to the end and agrees with Node is stale: it fails, so its entry leaves the
		// list. Otherwise it passes only as the very stop it was ruled.
		if agree {
			stale = true
		} else if acceptsRuntimeNotYet(stop, *record.Node, *record.JavaScript, *record.Native) {
			record.RuledRuntimeStop = true
		}
	}
	// A program at a ruled divergence's fixture passes as that divergence only when its run is exactly the row's
	// (#tqm95v6); anything else disagreeing still fails.
	if !agree && !record.NegativeWitness && !record.RuledRuntimeStop && acceptsRuledBackendDivergence(record.Path, runner.headTree, runner.divergences, *record.Node, *record.JavaScript, *record.Native) {
		record.RuledBackendDivergence = true
	}
	record.Agree = &agree
	if record.disagrees() {
		var reasons []string
		reasons = append(reasons, still...)
		for _, command := range repairStill {
			reasons = append(reasons, "repair "+command)
		}
		if len(reasons) > 0 {
			record.VoidReason = strings.Join(reasons, ", ") + " timeout"
		}
	}
	return stale, nil
}

// ruled says a run that disagrees with Node passes as a ruled exception: a negative witness, a run-time stop or a
// backend divergence.
func (record entry) ruled() bool {
	return record.NegativeWitness || record.RuledRuntimeStop || record.RuledBackendDivergence
}

// disagrees says a judged run neither agrees with Node nor passes as a ruled exception. Whether it is charged or void
// is VoidReason's to say.
func (record entry) disagrees() bool {
	return record.Agree != nil && !*record.Agree && !record.ruled()
}
