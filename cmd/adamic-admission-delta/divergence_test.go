package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// landedBackendDivergences is the oracle's own list, the file the command reads by default, as landed.
func landedBackendDivergences(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "internal/oracle/testdata/backend_divergences.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// requireMissing is what installed-layout's caught fixture prints, the family #3sh9398's programs are in: the source
// on Node, an ES module, has no require, and both backends catch installed-layout's error and print it. Their stdout
// differs from Node's, so the run is no agreement (the uncaught fixture's differs only on stderr, which ordinary
// comparison doesn't read). These are the program's runs, not a ruling: the ruling comes only from the landed row.
var requireMissing = struct {
	path          string
	node, backend observation
}{
	"internal/oracle/testdata/node_commonjs_require_missing.a",
	observation{Stdout: "ReferenceError none require is not defined\nReferenceError none require is not defined\nafter\n"},
	observation{Stdout: "Error MODULE_NOT_FOUND Cannot find module 'source-map-support'\nError MODULE_NOT_FOUND Cannot find module 'source-map-support'\nafter\n"},
}

// rewriteLandedBackendDivergences is the landed list with change made to the row at fixture.
func rewriteLandedBackendDivergences(t *testing.T, fixture string, change func(*ruledBackendDivergence)) []byte {
	t.Helper()
	var list ruledBackendDivergenceList
	if err := json.Unmarshal(landedBackendDivergences(t), &list); err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range list.Divergences {
		if list.Divergences[index].Fixture == fixture {
			change(&list.Divergences[index])
			changed = true
		}
	}
	if !changed {
		t.Fatalf("the landed list has no row at %s", fixture)
	}
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// backendDivergenceCommand runs the command on one newly admitted program at path, under the divergence list
// divergences, written where the command reads it by default. The source runs on Node as node, the JavaScript
// backend's output (marked BACKEND) as javascript, and the native build as native.
func backendDivergenceCommand(t *testing.T, path string, divergences []byte, node, javascript, native observation) (observation, report) {
	t.Helper()
	dir, _, base, head := commandFixture(t)
	outputs := t.TempDir()
	runs, err := json.Marshal(map[string]observation{"node": node, "javascript": javascript})
	if err != nil {
		t.Fatal(err)
	}
	runsPath := filepath.Join(outputs, "runs.json")
	runsName, err := json.Marshal(runsPath)
	if err != nil {
		t.Fatal(err)
	}
	for file, data := range map[string]string{"runs.json": string(runs), "native.stdout": native.Stdout, "native.stderr": native.Stderr} {
		if err := os.WriteFile(filepath.Join(outputs, file), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	nativeBinary := filepath.Join(outputs, "native")
	writeExecutable(t, nativeBinary, "#!/bin/sh\ncat '"+filepath.Join(outputs, "native.stdout")+"'\ncat '"+filepath.Join(outputs, "native.stderr")+"' >&2\nexit "+strconv.Itoa(native.Exit)+"\n")
	writeExecutable(t, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) echo '// BACKEND';;\nbuild) cp '"+nativeBinary+"' \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
	for file, data := range map[string]string{
		path:              "console.log('before');\n",
		"oracle/node.mjs": "import {readFileSync} from 'node:fs';const runs=JSON.parse(readFileSync(" + string(runsName) + ",'utf8'));const run=readFileSync(process.argv[2],'utf8').includes('BACKEND')?runs.javascript:runs.node;process.stdout.write(run.stdout);process.stderr.write(run.stderr);process.exitCode=run.exit;\n",
		"internal/oracle/testdata/backend_divergences.json": string(divergences),
	} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, file), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := git(dir, "add", path, "oracle/node.mjs", "internal/oracle/testdata/backend_divergences.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "ruled backend divergence input"); err != nil {
		t.Fatal(err)
	}
	sha, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--corpus", filepath.Dir(path), "--json")
	return o, runtimeNotYetReport(t, o)
}
func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// A program at a landed row's fixture, whose Node and native runs are exactly the row's and whose backends agree, is
// that ruled divergence: counted apart from agreement, witnesses and run-time stops, under ruled_backend_divergences
// with its path, and said as ruled backend divergences: N, and the verdict passes. Mutant: drop
// node_commonjs_require_missing.a's row from internal/oracle/testdata/backend_divergences.json, and the verdict
// reads fail, so this goes red.
func TestBackendDivergenceCommandAccepted(t *testing.T) {
	t.Parallel()
	o, r := backendDivergenceCommand(t, requireMissing.path, landedBackendDivergences(t), requireMissing.node, requireMissing.backend, requireMissing.backend)
	if o.Exit != 0 || r.Verdict != "pass" || r.RuledBackendDivergences != 1 || len(r.RuledBackendDivergencePaths) != 1 || r.RuledBackendDivergencePaths[0] != requireMissing.path || r.BackendDivergenceCount == 0 {
		t.Fatal("ruled backend divergence not counted", o, r)
	}
	p := r.Programs[0]
	if p.Class != "newly-accepted" || !p.RuledBackendDivergence || p.Agree == nil || *p.Agree || p.NegativeWitness || p.RuledRuntimeStop || r.NodeAgreements != 0 || r.RuledRuntimeStops != 0 {
		t.Fatal("ruled backend divergence not distinguished from agreement, a witness or a run-time stop", o, r)
	}
	if !strings.Contains(o.Stdout, `"ruled_backend_divergences":1,`) || !strings.Contains("\n"+o.Stderr, "\nruled backend divergences: 1\n") {
		t.Fatal("count not said", o)
	}
}

// Nothing else reads as a ruled divergence: the same runs at a path on no row, a row whose Node or native outcome
// differs from the run, and backends that disagree with each other each fail as any disagreement does. Mutants:
// acceptsRuledBackendDivergence matching the outcomes on any row's, without its path, accepts "a path on no row";
// without its Node comparison accepts "the row's Node outcome differs"; without its backends' agreement accepts
// "native and JavaScript disagree".
func TestBackendDivergenceCommandRefusals(t *testing.T) {
	t.Parallel()
	ruled := requireMissing
	for _, test := range []struct {
		name                     string
		path                     string
		divergences              []byte
		node, javascript, native observation
	}{
		{"a path on no row", "corpus/input.a", landedBackendDivergences(t), ruled.node, ruled.backend, ruled.backend},
		{"the row's Node outcome differs", ruled.path, rewriteLandedBackendDivergences(t, ruled.path, func(row *ruledBackendDivergence) {
			row.Node.Stdout = "before\nafter\n"
		}), ruled.node, ruled.backend, ruled.backend},
		{"the row's native outcome differs", ruled.path, rewriteLandedBackendDivergences(t, ruled.path, func(row *ruledBackendDivergence) {
			row.Native.Stdout = strings.ReplaceAll(row.Native.Stdout, "source-map-support", "typescript")
		}), ruled.node, ruled.backend, ruled.backend},
		{"native and JavaScript disagree", ruled.path, landedBackendDivergences(t), ruled.node, observation{Stdout: strings.TrimSuffix(ruled.backend.Stdout, "after\n")}, ruled.backend},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			o, r := backendDivergenceCommand(t, test.path, test.divergences, test.node, test.javascript, test.native)
			if o.Exit != 1 || r.Verdict != "fail" || r.RuledBackendDivergences != 0 || len(r.RuledBackendDivergencePaths) != 0 || r.Programs[0].RuledBackendDivergence {
				t.Fatal("not a ruled divergence, yet passed as one", o, r)
			}
			if !strings.Contains("\n"+o.Stderr, "\nruled backend divergences: 0\n") {
				t.Fatal("zero not said", o)
			}
		})
	}
}

// A row rules Node's side only when its node is set, and every run is held whole: a row with no Node outcome, or a
// Node run that errs, is never a ruled divergence.
func TestBackendDivergenceExactRun(t *testing.T) {
	t.Parallel()
	var list ruledBackendDivergenceList
	if err := json.Unmarshal(landedBackendDivergences(t), &list); err != nil {
		t.Fatal(err)
	}
	ruled := requireMissing
	if !acceptsRuledBackendDivergence(ruled.path, list, ruled.node, ruled.backend, ruled.backend) {
		t.Fatal("the ruled run refused")
	}
	timedOut := ruled.node
	timedOut.Error = "timeout"
	if acceptsRuledBackendDivergence(ruled.path, list, timedOut, ruled.backend, ruled.backend) {
		t.Error("a Node run that timed out accepted")
	}
	for _, row := range list.Divergences {
		if row.Node != nil {
			continue
		}
		backend := observation{Exit: row.Native.ExitCode, Stdout: row.Native.Stdout, Stderr: row.Native.Stderr}
		if acceptsRuledBackendDivergence(row.Fixture, list, observation{}, backend, backend) {
			t.Errorf("%s: a row that rules no Node outcome accepted", row.Fixture)
		}
	}
}

// The list as landed reads: every row names its fixture once, a ruling naming its task, and its reason.
func TestBackendDivergenceListReads(t *testing.T) {
	t.Parallel()
	list, err := readRuledBackendDivergences(filepath.Join("..", "..", "internal/oracle/testdata/backend_divergences.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Divergences) == 0 {
		t.Fatal("the landed list holds no rows")
	}
}

// Each row names its fixture's repository path once, a ruling naming its task, and its reason: the tool refuses
// the list otherwise. Mutant: readRuledBackendDivergences without its ruling check reads "no ruling".
func TestBackendDivergenceRulingMetadata(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "divergences.json")
	write := func(list ruledBackendDivergenceList) {
		t.Helper()
		if err := os.WriteFile(path, mustJSON(t, list), 0600); err != nil {
			t.Fatal(err)
		}
	}
	row := ruledBackendDivergence{
		Fixture: requireMissing.path, Name: "installed-layout", Ruling: "@system_adamic, task #redxes2, Nov 3", Reason: "The source on Node, an ES module, has no require.",
		Node:       &backendOutcome{Stdout: requireMissing.node.Stdout},
		Native:     backendOutcome{Stdout: requireMissing.backend.Stdout},
		JavaScript: backendOutcome{Stdout: requireMissing.backend.Stdout},
	}
	header := "Adding a row is a ruling routed to @system_adamic"
	write(ruledBackendDivergenceList{Comment: header, Divergences: []ruledBackendDivergence{row}})
	if _, err := readRuledBackendDivergences(path); err != nil {
		t.Fatal("ruled row refused:", err)
	}
	for name, change := range map[string]func(*ruledBackendDivergenceList){
		"no ruling":          func(l *ruledBackendDivergenceList) { l.Divergences[0].Ruling = "" },
		"a ruling of spaces": func(l *ruledBackendDivergenceList) { l.Divergences[0].Ruling = "   " },
		"no ruling task":     func(l *ruledBackendDivergenceList) { l.Divergences[0].Ruling = "@system_adamic, Nov 3" },
		"no reason":          func(l *ruledBackendDivergenceList) { l.Divergences[0].Reason = "" },
		"no fixture":         func(l *ruledBackendDivergenceList) { l.Divergences[0].Fixture = "" },
		"fixture outside":    func(l *ruledBackendDivergenceList) { l.Divergences[0].Fixture = "../input.a" },
		"fixture absolute":   func(l *ruledBackendDivergenceList) { l.Divergences[0].Fixture = "/input.a" },
		"fixture twice":      func(l *ruledBackendDivergenceList) { l.Divergences = append(l.Divergences, l.Divergences[0]) },
		"header":             func(l *ruledBackendDivergenceList) { l.Comment = "" },
		"no divergences":     func(l *ruledBackendDivergenceList) { l.Divergences = nil },
	} {
		changed := ruledBackendDivergenceList{Comment: header, Divergences: []ruledBackendDivergence{row}}
		change(&changed)
		write(changed)
		if _, err := readRuledBackendDivergences(path); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
