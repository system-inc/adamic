package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runtimeNotYetFixture is a primitive receiver's stop, ruled on #zhgxzjg and lifted by Runtime's #2ear8zk.
func runtimeNotYetFixture(test *testing.T) ([]byte, runtimeNotYetStop, runtimeNotYetStopList) {
	test.Helper()
	source := []byte("console.log('before');\nconsole.log(typeof Reflect.get(Object('ENOSPC'), 'length', 'ENOSPC'));\n")
	stop := runtimeNotYetStop{
		Path:        "corpus/input.a",
		RulingTask:  "#zhgxzjg",
		LiftingTask: "#2ear8zk",
		Stderr:      "adamic: panic: a primitive receiver's properties aren't read yet (reading 'length')\n",
	}
	return source, stop, runtimeNotYetStopList{
		Comment: "Adding an entry is a ruling routed to @system_adamic",
		Stops:   map[string]runtimeNotYetStop{contentHash(source): stop},
	}
}
func writeRuntimeNotYetList(test *testing.T, path string, list runtimeNotYetStopList) {
	test.Helper()
	data, encodeError := json.Marshal(list)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	if writeError := os.WriteFile(path, data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
}

// The list as landed reads: each entry a ruled form, with its ruling and lifting tasks.
func TestRuntimeNotYetListReads(test *testing.T) {
	test.Parallel()
	list, readRuntimeNotYetStopsError := readRuntimeNotYetStops(
		filepath.Join("..", "..", "cloud/admission-corpus/runtime-notyet-stops.json"),
	)
	if readRuntimeNotYetStopsError != nil {
		test.Fatal(readRuntimeNotYetStopsError)
	}
	if len(list.Stops) == 0 {
		test.Fatal("the landed list holds no stops")
	}
}

// An entry is keyed by its exact bytes: a renamed copy keeps it, any edit falls out, and an edit at the
// listed path is still scheduled for a run.
func TestRuntimeNotYetContentHash(test *testing.T) {
	test.Parallel()
	source, _, list := runtimeNotYetFixture(test)
	if _, listed := list.Stops[contentHash(source)]; !listed {
		test.Fatal("listed bytes lost")
	}
	edit := append(append([]byte{}, source...), '\n')
	if _, listed := list.Stops[contentHash(edit)]; listed {
		test.Fatal("edit retained the ruling")
	}
	if !runtimeNotYetRelated("corpus/input.a", edit, list) || !runtimeNotYetRelated("elsewhere/input.a", source, list) {
		test.Fatal("edit or renamed copy not scheduled")
	}
	if runtimeNotYetRelated("elsewhere/input.a", edit, list) {
		test.Fatal("unrelated program scheduled")
	}
}

// Each entry names its ruling and the task that lifts it, and says one allowlisted form: the tool refuses
// the list otherwise, not only review, and its ruling is the one that made its form. Mutants:
// readRuntimeNotYetStops without the allowlist check accepts "off the allowlist", and without the
// ruling's comparison accepts "another ruling".
func TestRuntimeNotYetRulingMetadata(test *testing.T) {
	test.Parallel()
	source, stop, list := runtimeNotYetFixture(test)
	path := filepath.Join(test.TempDir(), "stops.json")
	writeRuntimeNotYetList(test, path, list)
	if _, readRuntimeNotYetStopsError := readRuntimeNotYetStops(path); readRuntimeNotYetStopsError != nil {
		test.Fatal("ruled stop refused:", readRuntimeNotYetStopsError)
	}
	for name, change := range map[string]func(*runtimeNotYetStop, *runtimeNotYetStopList){
		"no ruling task":            func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) { s.RulingTask = "" },
		"no lifting task":           func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) { s.LiftingTask = "" },
		"lifting task unreferenced": func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) { s.LiftingTask = "2ear8zk" },
		"lifted by its ruling": func(
			s *runtimeNotYetStop,
			_ *runtimeNotYetStopList,
		) {
			s.LiftingTask = s.RulingTask
		},
		"another ruling": func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) { s.RulingTask = "#abc1234" },
		"no path":        func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) { s.Path = "" },
		"path outside":   func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) { s.Path = "../input.a" },
		"off the allowlist": func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) {
			s.Stderr = "adamic: panic: a function's properties aren't read by a run-time key yet (reading 'length')\n"
		},
		"another guard": func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) {
			s.Stderr = "adamic: panic: index 3 is outside an array of length 1\n"
		},
		"no key": func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) {
			s.Stderr = "adamic: panic: a primitive receiver's properties aren't read yet\n"
		},
		"two lines": func(s *runtimeNotYetStop, _ *runtimeNotYetStopList) { s.Stderr += s.Stderr },
		"no newline": func(
			s *runtimeNotYetStop,
			_ *runtimeNotYetStopList,
		) {
			s.Stderr = strings.TrimSuffix(s.Stderr, "\n")
		},
		"header": func(_ *runtimeNotYetStop, l *runtimeNotYetStopList) { l.Comment = "" },
	} {
		test.Run(name, func(test *testing.T) {
			bad := stop
			changed := runtimeNotYetStopList{Comment: list.Comment}
			change(&bad, &changed)
			changed.Stops = map[string]runtimeNotYetStop{contentHash(source): bad}
			writeRuntimeNotYetList(test, path, changed)
			if _, readRuntimeNotYetStopsError := readRuntimeNotYetStops(path); readRuntimeNotYetStopsError == nil {
				test.Errorf("%s accepted", name)
			}
		})
	}
	writeRuntimeNotYetList(
		test,
		path,
		runtimeNotYetStopList{Comment: list.Comment, Stops: map[string]runtimeNotYetStop{"bad": stop}},
	)
	if _, readRuntimeNotYetStopsError := readRuntimeNotYetStops(path); readRuntimeNotYetStopsError == nil {
		test.Error("bad content hash accepted")
	}
}

// The keyless stops String and a template make of an unknown (#09dvwny) are the second shape: each of the
// three ruled whats, said whole, reads under its ruling, and nothing else of that shape does, a fourth
// what, a pattern's near miss or another ruling. Mutants: runtimeNotYetForm matching the shape's two ends
// around any what accepts "a fourth what"; without the keyless shape, every ruled what is refused.
func TestRuntimeNotYetConversionForms(test *testing.T) {
	test.Parallel()
	source, stop, list := runtimeNotYetFixture(test)
	path := filepath.Join(test.TempDir(), "stops.json")
	read := func(stderr, ruling string) error {
		converted := stop
		converted.Stderr, converted.RulingTask, converted.LiftingTask = stderr, ruling, "#0cmbkgt"
		writeRuntimeNotYetList(test, path, runtimeNotYetStopList{
			Comment: list.Comment,
			Stops:   map[string]runtimeNotYetStop{contentHash(source): converted},
		})
		_, readRuntimeNotYetStopsError := readRuntimeNotYetStops(path)
		return readRuntimeNotYetStopsError
	}
	for _, what := range []string{
		"an object whose own toString is a function this program can't name",
		"an array of only numbers, booleans, null or undefined",
		"a class object",
	} {
		if readError := read(
			"adamic: panic: converting "+what+" to a string is not yet lowered\n",
			"#09dvwny",
		); readError != nil {
			test.Errorf("ruled conversion %q refused: %v", what, readError)
		}
	}
	for name, stderr := range map[string]string{
		"a fourth what":       "adamic: panic: converting a function to a string is not yet lowered\n",
		"a what's near miss":  "adamic: panic: converting an array of only numbers to a string is not yet lowered\n",
		"no newline":          "adamic: panic: converting a class object to a string is not yet lowered",
		"two lines":           "adamic: panic: converting a class object to a string is not yet lowered\nadamic: panic: converting a class object to a string is not yet lowered\n",
		"another ending":      "adamic: panic: converting a class object to a string is not yet supported\n",
		"no panic prefix":     "converting a class object to a string is not yet lowered\n",
		"a key read after it": "adamic: panic: converting a class object to a string is not yet lowered (reading 'x')\n",
	} {
		if readError := read(stderr, "#09dvwny"); readError == nil {
			test.Errorf("%s accepted: %q", name, stderr)
		}
	}
	if readError := read(
		"adamic: panic: converting a class object to a string is not yet lowered\n",
		"#zhgxzjg",
	); readError == nil {
		test.Error("a conversion under another ruling accepted")
	}
}

// A keyless stop that isn't a conversion is the third shape (#ey7aysy): Object.keys of a runtime host value,
// said whole, reads under #0sps03r, and nothing near it does. Mutants: runtimeNotYetForm without the
// keyless table refuses the ruled message; matching it by prefix accepts "a near miss".
func TestRuntimeNotYetKeylessForms(test *testing.T) {
	test.Parallel()
	source, stop, list := runtimeNotYetFixture(test)
	path := filepath.Join(test.TempDir(), "stops.json")
	read := func(stderr, ruling string) error {
		keyless := stop
		keyless.Stderr, keyless.RulingTask, keyless.LiftingTask = stderr, ruling, "#ey7aysy"
		writeRuntimeNotYetList(test, path, runtimeNotYetStopList{
			Comment: list.Comment,
			Stops:   map[string]runtimeNotYetStop{contentHash(source): keyless},
		})
		_, readRuntimeNotYetStopsError := readRuntimeNotYetStops(path)
		return readRuntimeNotYetStopsError
	}
	ruled := "adamic: panic: this runtime value's own keys aren't enumerated yet\n"
	if readError := read(ruled, "#0sps03r"); readError != nil {
		test.Errorf("the ruled keyless message refused: %v", readError)
	}
	for name, stderr := range map[string]string{
		"a message off the list": "adamic: panic: this runtime value's own values aren't enumerated yet\n",
		"a near miss":            "adamic: panic: this runtime value's own keys aren't enumerated yet, at all\n",
		"no newline":             "adamic: panic: this runtime value's own keys aren't enumerated yet",
		"two lines":              ruled + ruled,
		"no panic prefix":        "this runtime value's own keys aren't enumerated yet\n",
		"a key read after it":    "adamic: panic: this runtime value's own keys aren't enumerated yet (reading 'keys')\n",
		"as a conversion":        "adamic: panic: converting this runtime value's own keys aren't enumerated yet to a string is not yet lowered\n",
	} {
		if readError := read(stderr, "#0sps03r"); readError == nil {
			test.Errorf("%s accepted: %q", name, stderr)
		}
	}
	if readError := read(ruled, "#zhgxzjg"); readError == nil {
		test.Error("the keyless message under another ruling accepted")
	}
}

// internal/ir's two compiler-bug stops are never a NotYet, listed or allowlisted. Not parallel: it widens
// the allowlist for its run, before the parallel tests resume.
func TestRuntimeNotYetCompilerBugNeverQualifies(test *testing.T) {
	source, stop, list := runtimeNotYetFixture(test)
	path := filepath.Join(test.TempDir(), "stops.json")
	bugs := []string{
		"compiler bug: a run-time key read reached null or undefined",
		"compiler bug: a run-time key read reached a class with no row in its table",
	}
	for _, widened := range []bool{false, true} {
		if widened {
			for _, bug := range bugs {
				runtimeNotYetForms[bug] = "#zhgxzjg"
				defer delete(runtimeNotYetForms, bug)
			}
		}
		for _, bug := range bugs {
			bad := stop
			bad.Stderr = "adamic: panic: " + bug + " (reading 'length')\n"
			writeRuntimeNotYetList(test, path, runtimeNotYetStopList{
				Comment: list.Comment,
				Stops:   map[string]runtimeNotYetStop{contentHash(source): bad},
			})
			if _, readRuntimeNotYetStopsError := readRuntimeNotYetStops(path); readRuntimeNotYetStopsError == nil {
				test.Errorf("listed compiler bug accepted (allowlist widened: %v): %s", widened, bug)
			}
		}
	}
}

// A run is the listed stop only with both backends' exit 70 and whole message, after exactly Node's
// first bytes, where Node goes on to the end. Mutant: acceptsRuntimeNotYet without its prefix check
// accepts "stdout one byte off Node's".
func TestRuntimeNotYetExactStop(test *testing.T) {
	test.Parallel()
	_, stop, _ := runtimeNotYetFixture(test)
	node := observation{Exit: 0, Stdout: "before\nnumber\n"}
	stopped := observation{Exit: 70, Stdout: "before\n", Stderr: stop.Stderr}
	if !acceptsRuntimeNotYet(stop, node, stopped, stopped) {
		test.Fatal("ruled stop rejected")
	}
	for _, testCase := range []struct {
		name                     string
		node, javascript, native observation
	}{
		{"native exits other than 70", node, stopped, observation{Exit: 1, Stdout: "before\n", Stderr: stop.Stderr}},
		{"native exits 0", node, stopped, observation{Exit: 0, Stdout: "before\n", Stderr: stop.Stderr}},
		{
			"javascript exits other than 70",
			node,
			observation{Exit: 1, Stdout: "before\n", Stderr: stop.Stderr},
			stopped,
		},
		{
			"javascript times out",
			node,
			observation{Exit: 70, Stdout: "before\n", Stderr: stop.Stderr, Error: "timeout"},
			stopped,
		},
		{"another allowlisted form", node, stopped, observation{
			Exit:   70,
			Stdout: "before\n",
			Stderr: "adamic: panic: a method isn't read by a run-time key yet (reading 'length')\n",
		}},
		{"another key", node, observation{
			Exit:   70,
			Stdout: "before\n",
			Stderr: "adamic: panic: a primitive receiver's properties aren't read yet (reading 'size')\n",
		}, stopped},
		{
			"stdout one byte off Node's",
			node,
			observation{Exit: 70, Stdout: "beforf\n", Stderr: stop.Stderr},
			observation{Exit: 70, Stdout: "beforf\n", Stderr: stop.Stderr},
		},
		{"backends' stdout differ", node, observation{Exit: 70, Stdout: "", Stderr: stop.Stderr}, stopped},
		{"Node stops there", observation{Exit: 70, Stdout: "before\n", Stderr: stop.Stderr}, stopped, stopped},
		{"Node exits nonzero", observation{Exit: 1, Stdout: "before\nnumber\n"}, stopped, stopped},
		{"Node prints nothing more", observation{Exit: 0, Stdout: "before\n"}, stopped, stopped},
		{"Node times out", observation{Exit: 0, Stdout: "before\nnumber\n", Error: "timeout"}, stopped, stopped},
	} {
		test.Run(testCase.name, func(test *testing.T) {
			if acceptsRuntimeNotYet(stop, testCase.node, testCase.javascript, testCase.native) {
				test.Errorf("%s accepted", testCase.name)
			}
		})
	}
}

// runtimeNotYetCommand runs the command on the fixture with a compiler whose backends both stop as the
// entry says after printing before, or, when stale, run to the end as Node does. Node's oracle runs the
// source to the end, and the JavaScript backend's output, marked STOP, stops as native does. rewrite, when
// given, rewrites the list before it's committed.
func runtimeNotYetCommand(
	test *testing.T,
	stale, alreadyAccepted bool,
	rewrite func(*runtimeNotYetStopList),
	args ...string,
) observation {
	test.Helper()
	dir, _, base, head := commandFixture(test)
	source, stop, list := runtimeNotYetFixture(test)
	message, encodeError := json.Marshal(strings.TrimSuffix(stop.Stderr, "\n"))
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	native := filepath.Join(test.TempDir(), "native")
	javascript := "// STOP"
	if stale {
		writeExecutable(test, native, "#!/bin/sh\necho before\necho number\n")
		javascript = "// GO"
	} else {
		writeExecutable(
			test,
			native,
			"#!/bin/sh\necho before\ncat >&2 <<'END'\n"+strings.TrimSuffix(stop.Stderr, "\n")+"\nEND\nexit 70\n",
		)
	}
	writeExecutable(test, head, "#!/bin/sh\ncase \"$1\" in\nadmission-lower) exit 0;;\njs) echo '"+
		javascript+
		"';;\nbuild) cp '"+
		native+
		"' \"$4.tmp\" && chmod +x \"$4.tmp\" && mv \"$4.tmp\" \"$4\";;\nesac\n")
	for path, data := range map[string]string{
		"corpus/input.a": string(source),
		"oracle/node.mjs": "import {readFileSync} from 'node:fs';const s=readFileSync(process.argv[2],'utf8');console.log('before');if(s.includes('STOP')){console.error(" +
			string(message) +
			");process.exit(70);}console.log('number');\n",
	} {
		if writeError := os.WriteFile(filepath.Join(dir, path), []byte(data), 0755); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if rewrite != nil {
		rewrite(&list)
	}
	writeRuntimeNotYetList(test, filepath.Join(dir, "cloud/admission-corpus/runtime-notyet-stops.json"), list)
	if _, gitError := git(
		dir,
		"add",
		"corpus/input.a",
		"oracle/node.mjs",
		"cloud/admission-corpus/runtime-notyet-stops.json",
	); gitError != nil {
		test.Fatal(gitError)
	}
	if _, gitError := git(dir, "commit", "-qm", "run-time NotYet input"); gitError != nil {
		test.Fatal(gitError)
	}
	sha, gitError := git(dir, "rev-parse", "HEAD")
	if gitError != nil {
		test.Fatal(gitError)
	}
	if alreadyAccepted {
		base = head
	}
	return invokeCommand(test, dir, append(
		[]string{"--base", sha, "--head", sha, "--base-binary", base, "--head-binary", head, "--corpus", "corpus"},
		args...,
	)...)
}
func runtimeNotYetReport(test *testing.T, o observation) report {
	test.Helper()
	var r report
	if decodeError := json.Unmarshal([]byte(o.Stdout), &r); decodeError != nil {
		test.Fatal(decodeError, o.Stderr, o.Stdout)
	}
	return r
}

// A listed stop is counted as its own kind, apart from agreeing and from negative witnesses, under
// ruled_runtime_stops with its path, and in the line ruled run-time stops: N.
func TestRuntimeNotYetCommandAccepted(test *testing.T) {
	test.Parallel()
	o := runtimeNotYetCommand(test, false, false, nil, "--json")
	r := runtimeNotYetReport(test, o)
	if o.Exit != 0 ||
		r.Verdict != "pass" ||
		r.RuledRuntimeStops != 1 ||
		len(r.RuledRuntimeStopPaths) != 1 ||
		r.RuledRuntimeStopPaths[0] != "corpus/input.a" ||
		r.RuntimeNotYetCount != 1 {
		test.Fatal("ruled run-time stop not counted", o, r)
	}
	p := r.Programs[0]
	if p.Class != "newly-accepted" ||
		!p.RuledRuntimeStop ||
		p.Agree == nil ||
		*p.Agree ||
		p.NegativeWitness ||
		r.NodeAgreements != 0 ||
		r.AcceptedWitnesses != 0 ||
		len(r.StaleRuntimeStops) != 0 {
		test.Fatal("ruled run-time stop not distinguished from agreement or witness", o, r)
	}
	if !strings.Contains(o.Stdout, `"ruled_runtime_stops":1,`) ||
		!strings.Contains("\n"+o.Stderr, "\nruled run-time stops: 1\n") {
		test.Fatal("count not said", o)
	}
}

// The human run says the same line.
func TestRuntimeNotYetCommandHumanLine(test *testing.T) {
	test.Parallel()
	o := runtimeNotYetCommand(test, false, false, nil)
	if o.Exit != 0 ||
		!strings.Contains(o.Stdout, "verdict: pass\n") ||
		!strings.Contains("\n"+o.Stderr, "\nruled run-time stops: 1\n") {
		test.Fatal("human run did not say its ruled run-time stops", o)
	}
}

// A listed program both compilers already admit still runs and validates, so the list can't rot between
// changes.
func TestRuntimeNotYetCommandAlreadyAccepted(test *testing.T) {
	test.Parallel()
	o := runtimeNotYetCommand(test, false, true, nil, "--json")
	r := runtimeNotYetReport(test, o)
	if o.Exit != 0 ||
		r.Verdict != "pass" ||
		r.Programs[0].Class != "accepted-by-both" ||
		!r.Programs[0].Sampled ||
		!r.Programs[0].NotYetRelated ||
		r.RuledRuntimeStops != 1 {
		test.Fatal("already-admitted listed stop not validated", o, r)
	}
}

// A listed program that now runs to the end on both backends and agrees with Node fails, named as stale,
// so its entry leaves the list. Mutant: the run without its stale verdict passes it as an agreement.
func TestRuntimeNotYetCommandStale(test *testing.T) {
	test.Parallel()
	for _, alreadyAccepted := range []bool{false, true} {
		o := runtimeNotYetCommand(test, true, alreadyAccepted, nil, "--json")
		r := runtimeNotYetReport(test, o)
		if o.Exit != 1 ||
			r.Verdict != "fail" ||
			len(r.StaleRuntimeStops) != 1 ||
			r.StaleRuntimeStops[0] != "corpus/input.a" ||
			r.RuledRuntimeStops != 0 ||
			r.Programs[0].RuledRuntimeStop ||
			r.Programs[0].Agree == nil ||
			!*r.Programs[0].Agree {
			test.Fatal("stale listed stop passed", alreadyAccepted, o, r)
		}
		if !strings.Contains("\n"+o.Stderr, "\nruled run-time stops: 0\n") {
			test.Fatal("zero not said", o)
		}
	}
}

// orphanRuntimeNotYet lists the stop under bytes no program holds, at path: an edit of the listed program
// when path holds other bytes, a program not landed yet when nothing is there.
func orphanRuntimeNotYet(path string) func(*runtimeNotYetStopList) {
	return func(list *runtimeNotYetStopList) {
		orphaned := map[string]runtimeNotYetStop{}
		for _, stop := range list.Stops {
			stop.Path = path
			orphaned[contentHash([]byte("an earlier program\n"))] = stop
		}
		list.Stops = orphaned
	}
}

// A listed program edited in place leaves its entry matching nothing: a run holding every corpus fails it
// as stale, though the edit itself agrees with Node. Mutant: the run without its edited-path check passes.
func TestRuntimeNotYetCommandEdited(test *testing.T) {
	test.Parallel()
	o := runtimeNotYetCommand(test, true, false, orphanRuntimeNotYet("corpus/input.a"), "--json")
	r := runtimeNotYetReport(test, o)
	if o.Exit != 1 ||
		r.Verdict != "fail" ||
		len(r.StaleRuntimeStops) != 1 ||
		r.StaleRuntimeStops[0] != "corpus/input.a" ||
		len(r.PendingRuntimeStops) != 0 ||
		r.Programs[0].Agree == nil ||
		!*r.Programs[0].Agree {
		test.Fatal("edited listed program left its entry unnoticed", o, r)
	}
}

// A listed program not landed at head passes, said as pending, so the list can land before its programs.
func TestRuntimeNotYetCommandPending(test *testing.T) {
	test.Parallel()
	o := runtimeNotYetCommand(test, true, false, orphanRuntimeNotYet("corpus/landing.a"), "--json")
	r := runtimeNotYetReport(test, o)
	if o.Exit != 0 ||
		r.Verdict != "pass" ||
		len(r.PendingRuntimeStops) != 1 ||
		r.PendingRuntimeStops[0] != "corpus/landing.a" ||
		len(r.StaleRuntimeStops) != 0 ||
		r.RuntimeNotYetCount != 1 {
		test.Fatal("unlanded listed program not said as pending", o, r)
	}
}

// A sharded run can't know what another shard held, so it judges no entry nothing held.
func TestRuntimeNotYetCommandShardDoesNotJudge(test *testing.T) {
	test.Parallel()
	for _, path := range []string{"corpus/input.a", "corpus/landing.a"} {
		o := runtimeNotYetCommand(test, true, false, orphanRuntimeNotYet(path), "--json", "--shard", "1/2")
		r := runtimeNotYetReport(test, o)
		if o.Exit != 0 ||
			r.Verdict != "partial" ||
			r.CompleteCorpus ||
			len(r.StaleRuntimeStops) != 0 ||
			len(r.PendingRuntimeStops) != 0 {
			test.Fatal("sharded run judged an entry nothing held", path, o, r)
		}
	}
}

// One program can't be both a type lie and a run-time NotYet.
func TestRuntimeNotYetNeverAWitness(test *testing.T) {
	test.Parallel()
	dir, sha, base, head := commandFixture(test)
	source, stop, list := runtimeNotYetFixture(test)
	_, witness, negatives := negativeFixture(test)
	negatives.Witnesses = map[string]negativeWitness{contentHash(source): witness}
	writeNegativeList(test, filepath.Join(dir, "cloud/admission-corpus/negative-witnesses.json"), negatives)
	list.Stops[contentHash(source)] = stop
	writeRuntimeNotYetList(test, filepath.Join(dir, "cloud/admission-corpus/runtime-notyet-stops.json"), list)
	o := invokeCommand(
		test,
		dir,
		"--base",
		sha,
		"--head",
		sha,
		"--base-binary",
		base,
		"--head-binary",
		head,
		"--corpus",
		"corpus",
		"--json",
	)
	if o.Exit != 1 || !strings.Contains(o.Stderr, "listed both as a negative witness and as a run-time NotYet") {
		test.Fatal("program listed in both lists", o)
	}
}
