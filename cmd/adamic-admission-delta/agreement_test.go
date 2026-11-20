package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// agreementCompiler is a fake head compiler for the agreement mode. Its lowering refuses a program holding the line
// refuse and accepts every other. It compiles a program to JavaScript as the program's own text, or, where the program
// holds js-print or js-stop lines, as those lines alone, made print and stop lines. It builds a program natively as a shell script of the
// program's native lines, and only with --sanitize after -o <out>: an unsanitized build fails, so a mode that forgot
// the sanitizers would disagree everywhere.
const agreementCompiler = `#!/bin/sh
case "$1" in
admission-lower)
	if grep -qx refuse "$2"; then echo "adamic: $2: stage 0 can't lower this yet" >&2; exit 1; fi
	exit 0;;
js)
	if grep -q -e '^js-print: ' -e '^js-stop: ' "$2"; then sed -n -e 's/^js-print: /print: /p' -e 's/^js-stop: /stop: /p' "$2"; else cat "$2"; fi;;
build)
	if [ "$5" != --sanitize ]; then echo "adamic: this native build is not sanitized" >&2; exit 1; fi
	{ echo '#!/bin/sh'; sed -n 's/^native: //p' "$2"; } > "$4.tmp" && chmod +x "$4.tmp" && mv "$4.tmp" "$4";;
esac
`

// agreementNode is the fixture's oracle/node.mjs: it prints a program's print lines, so Node and the JavaScript
// backend print what a program's print lines say, and at a stop line it writes the line's message to stderr and
// stops with exit 70, as a backend's checked stop does. It sets the exit code rather than exiting, so what it printed
// to a pipe is whole.
const agreementNode = "import {readFileSync} from 'node:fs';\nfor (const line of readFileSync(process.argv[2], 'utf8').split('\\n')) {\n  if (line.startsWith('print: ')) console.log(line.slice('print: '.length));\n  else if (line.startsWith('stop: ')) { process.stderr.write(line.slice('stop: '.length) + '\\n'); process.exitCode = 70; break; }\n}\n"

// The four kinds of program the agreement mode tells apart, as agreementCompiler and agreementNode run them.
const (
	agreeingSource    = "print: 42\nnative: echo 42\n"
	ruledSource       = "print: node\njs-print: backends\nnative: echo backends\n"
	disagreeingSource = "print: 42\nnative: echo 41\n"
	refusedSource     = "refuse\n"
)

type agreementSource struct {
	path, source string
}

// agreementFixture is commandFixture's repository with programs committed at head, in one corpus of a manifest that
// lists them in the order given, and agreementCompiler as head's compiler. The divergence list rules corpus/ruled.a:
// Node prints node and both backends print backends.
func agreementFixture(t *testing.T, programs []agreementSource) (string, string, string) {
	t.Helper()
	return agreementFixtureWith(t, programs, nil)
}

// agreementFixtureWith is agreementFixture with other files committed beside the programs, a sidecar or a list, that
// the manifest doesn't name.
func agreementFixtureWith(t *testing.T, programs []agreementSource, others map[string]string) (string, string, string) {
	t.Helper()
	dir, _, _, head := commandFixture(t)
	writeExecutable(t, head, agreementCompiler)
	divergences, err := json.Marshal(ruledBackendDivergenceList{Comment: "Adding a row is a ruling routed to @system_adamic", Divergences: []ruledBackendDivergence{{
		Fixture:    "corpus/ruled.a",
		Ruling:     "Adamic, on #tqm95v6",
		Reason:     "the backends are ruled to print what the source on Node can't",
		Node:       &backendOutcome{Stdout: "node\n"},
		Native:     backendOutcome{Stdout: "backends\n"},
		JavaScript: backendOutcome{Stdout: "backends\n"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{"oracle/node.mjs": agreementNode, "internal/oracle/testdata/backend_divergences.json": string(divergences)}
	for path, text := range others {
		files[path] = text
	}
	for _, p := range programs {
		files[p.path] = p.source
	}
	for path, text := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, path), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := git(dir, "add", path); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := git(dir, "commit", "-qm", "agreement programs"); err != nil {
		t.Fatal(err)
	}
	sha, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	listed := corpus{Name: "fixtures", Programs: []program{}}
	for _, p := range programs {
		blob, err := git(dir, "rev-parse", "HEAD:"+p.path)
		if err != nil {
			t.Fatal(err)
		}
		listed.Programs = append(listed.Programs, program{Path: p.path, Blob: blob})
	}
	data, err := json.Marshal(manifest{SHA: sha, Corpora: []corpus{listed}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	return dir, sha, head
}

// agreementCommand runs the agreement mode on the fixture's head and manifest, with extra arguments, and reads its
// JSON report.
func agreementCommand(t *testing.T, dir, sha, head string, extra ...string) (observation, agreementReport) {
	t.Helper()
	o := invokeCommand(t, dir, append([]string{"--agreement", "--head", sha, "--head-binary", head, "--manifest", filepath.Join(dir, "manifest.json"), "--json"}, extra...)...)
	var r agreementReport
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	return o, r
}

// stderrLine is the line of stderr that starts with prefix, or "".
func stderrLine(stderr, prefix string) string {
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	return ""
}

// Every program head accepts is run and judged, with no base: one agrees, one is a ruled backend divergence, one
// disagrees, and the one head refuses is counted and never run. The disagreement fails the run, named on its own line
// before the summary, which is the last line. Without it the same corpus passes. Mutant (agreement.mutants.md): the
// tally's comparison dropped, so every accepted program counts as agreeing, and the run reads agreeing 3 and exits 0.
func TestAgreementCountsEachJudgment(t *testing.T) {
	t.Parallel()
	dir, sha, head := agreementFixture(t, []agreementSource{{"corpus/agree.a", agreeingSource}, {"corpus/ruled.a", ruledSource}, {"corpus/disagree.a", disagreeingSource}, {"corpus/refused.a", refusedSource}})
	o, r := agreementCommand(t, dir, sha, head)
	if o.Exit != 1 || r.Verdict != "fail" || r.Compared != 3 || r.Agreeing != 1 || r.Ruled != 1 || r.Disagreeing != 1 || r.MemoryUnsafe != 0 || r.Refused != 1 {
		t.Fatalf("exit %d, %+v, want exit 1 with one each of agreeing, ruled, disagreeing and refused: %s", o.Exit, r, o.Stderr)
	}
	if !strings.HasSuffix(o.Stderr, "\nagreement: compared 3, agreeing 1, ruled 1, checked stops 0, pending 0, disagreeing 1, memory-unsafe 0, refused 1\n") {
		t.Fatalf("the summary isn't the last line of stderr: %s", o.Stderr)
	}
	if len(r.Disagreements) != 1 || r.Disagreements[0].Path != "corpus/disagree.a" || r.Disagreements[0].MemoryUnsafe || r.Disagreements[0].Native.Stdout != "41\n" || r.Disagreements[0].Node.Stdout != "42\n" {
		t.Fatalf("disagreements: %+v", r.Disagreements)
	}
	line := stderrLine(o.Stderr, "disagreeing corpus/disagree.a:")
	if !strings.HasPrefix(line, `disagreeing corpus/disagree.a: exit node 0, javascript 0, native 0; stdout node "42", javascript "42", native "41"`) || strings.HasSuffix(line, "memory-unsafe") {
		t.Fatalf("the disagreement's line: %q in %s", line, o.Stderr)
	}
	judgments := map[string]string{}
	for _, judged := range r.Programs {
		judgments[judged.Path] = judged.Judgment
	}
	if want := map[string]string{"corpus/agree.a": "agreeing", "corpus/ruled.a": "ruled", "corpus/disagree.a": "disagreeing", "corpus/refused.a": "refused"}; !reflect.DeepEqual(judgments, want) {
		t.Fatalf("judgments %v, want %v", judgments, want)
	}
	// Every program judged by what it did carries its outcome hash; the refused one, never run, has none.
	for _, judged := range r.Programs {
		if hashed := len(judged.OutcomeHash) == 64; hashed != (judged.Judgment != "refused") {
			t.Errorf("%s, %s, has outcome hash %q", judged.Path, judged.Judgment, judged.OutcomeHash)
		}
	}
	dir, sha, head = agreementFixture(t, []agreementSource{{"corpus/agree.a", agreeingSource}, {"corpus/ruled.a", ruledSource}, {"corpus/refused.a", refusedSource}})
	o, r = agreementCommand(t, dir, sha, head)
	if o.Exit != 0 || r.Verdict != "pass" || r.Disagreeing != 0 || !strings.HasSuffix(o.Stderr, "\nagreement: compared 2, agreeing 1, ruled 1, checked stops 0, pending 0, disagreeing 0, memory-unsafe 0, refused 1\n") {
		t.Fatalf("a corpus with no disagreement: exit %d, %+v, want exit 0: %s", o.Exit, r, o.Stderr)
	}
}

// A disagreement is memory-unsafe when the native run's stderr shows a sanitizer's report or the native process died
// on a signal: here AddressSanitizer's line beside an exit 1, and a SIGKILL. Each is counted, flagged in the report and
// said at the end of its line.
func TestAgreementMemoryUnsafe(t *testing.T) {
	t.Parallel()
	dir, sha, head := agreementFixture(t, []agreementSource{
		{"corpus/sanitizer.a", "print: 42\nnative: echo 42\nnative: echo '==1==ERROR: AddressSanitizer: heap-use-after-free on address 0x1' >&2\nnative: exit 1\n"},
		{"corpus/signal.a", "print: 42\nnative: echo 42\nnative: kill -KILL $$\n"},
	})
	o, r := agreementCommand(t, dir, sha, head)
	if o.Exit != 1 || r.Disagreeing != 2 || r.MemoryUnsafe != 2 || !strings.HasSuffix(o.Stderr, "\nagreement: compared 2, agreeing 0, ruled 0, checked stops 0, pending 0, disagreeing 2, memory-unsafe 2, refused 0\n") {
		t.Fatalf("exit %d, %+v, want two memory-unsafe disagreements: %s", o.Exit, r, o.Stderr)
	}
	for _, disagreement := range r.Disagreements {
		if !disagreement.MemoryUnsafe {
			t.Errorf("%s isn't flagged memory-unsafe: %+v", disagreement.Path, disagreement.Native)
		}
		if disagreement.Path == "corpus/signal.a" && disagreement.Native.Error != "crash" {
			t.Errorf("the signal's native run: %+v, want a crash", disagreement.Native)
		}
		if line := stderrLine(o.Stderr, "disagreeing "+disagreement.Path+":"); !strings.HasSuffix(line, "; memory-unsafe") {
			t.Errorf("%s's line doesn't say memory-unsafe: %q", disagreement.Path, line)
		}
	}
	if line := stderrLine(o.Stderr, "disagreeing corpus/sanitizer.a:"); !strings.Contains(line, `exit node 0, javascript 0, native 1; stderr node "", javascript "", native "==1==ERROR: AddressSanitizer: heap-use-after-free on address 0x1"`) {
		t.Errorf("the sanitizer's line: %q", line)
	}
}

// --shard k/n keeps the programs whose position in the manifest's order is k-1 modulo n, each path counted once at its
// first listing, so the shards are disjoint and together cover the corpus, and a shard alone is partial.
func TestAgreementShardsPartitionTheCorpus(t *testing.T) {
	t.Parallel()
	twice := []corpus{{Name: "first", Programs: []program{{Path: "a"}, {Path: "b"}}}, {Name: "second", Programs: []program{{Path: "b"}, {Path: "c"}}}}
	if kept := shardByIndex(twice, 2, 2); len(kept) != 2 || !reflect.DeepEqual(kept[0].Programs, []program{{Path: "b"}}) || len(kept[1].Programs) != 0 {
		t.Fatalf("a path listed twice: shard 2/2 kept %+v, want b once, from its first listing", kept)
	}
	var programs []agreementSource
	var paths []string
	for _, name := range []string{"a", "b", "c", "d", "e"} {
		programs = append(programs, agreementSource{"corpus/" + name + ".a", agreeingSource})
		paths = append(paths, "corpus/"+name+".a")
	}
	dir, sha, head := agreementFixture(t, programs)
	want := map[string][]string{"1/3": {"corpus/a.a", "corpus/d.a"}, "2/3": {"corpus/b.a", "corpus/e.a"}, "3/3": {"corpus/c.a"}}
	var covered []string
	for _, shard := range []string{"1/3", "2/3", "3/3"} {
		o, r := agreementCommand(t, dir, sha, head, "--shard", shard)
		var kept []string
		for _, judged := range r.Programs {
			kept = append(kept, judged.Path)
		}
		if o.Exit != 0 || r.Verdict != "partial" || r.CompleteCorpus || r.Shard != shard || !reflect.DeepEqual(kept, want[shard]) || r.Agreeing != len(want[shard]) {
			t.Fatalf("shard %s: exit %d, verdict %s, kept %v, want %v: %s", shard, o.Exit, r.Verdict, kept, want[shard], o.Stderr)
		}
		covered = append(covered, kept...)
	}
	sort.Strings(covered)
	if !reflect.DeepEqual(covered, paths) {
		t.Fatalf("the shards together hold %v, want each of %v once", covered, paths)
	}
}

// The mode judges head alone, so --base is refused and --head must be named, and a shard is k/n with 1 <= k <= n. Each
// is a usage error, said before anything runs, exiting 1 as every usage error of the command does.
func TestAgreementFlagErrors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"with --base", []string{"--agreement", "--base", "main", "--head", "HEAD", "--manifest", "manifest.json"}, "--agreement refuses --base"},
		{"without --head", []string{"--agreement", "--manifest", "manifest.json"}, "--agreement requires --head"},
		{"shard zero", []string{"--agreement", "--head", "HEAD", "--manifest", "manifest.json", "--shard", "0/2"}, `invalid shard "0/2"`},
		{"shard past its count", []string{"--agreement", "--head", "HEAD", "--manifest", "manifest.json", "--shard", "3/2"}, `invalid shard "3/2"`},
		{"shard without a count", []string{"--agreement", "--head", "HEAD", "--manifest", "manifest.json", "--shard", "2"}, "shard must be i/N"},
		{"shard not a number", []string{"--agreement", "--head", "HEAD", "--manifest", "manifest.json", "--shard", "a/2"}, `invalid shard "a/2"`},
	} {
		o := invokeCommand(t, dir, c.args...)
		if o.Exit != 1 || !strings.Contains(o.Stderr, c.want) {
			t.Errorf("%s: exit %d, stderr %q, want exit 1 saying %q", c.name, o.Exit, o.Stderr, c.want)
		}
	}
}

// The native program runs with ASAN_OPTIONS=detect_leaks=0, as the oracle's comparison run does, so a leak at exit,
// which a ruled stop's exit 70 always has, never reaches stderr. The fake native reports a leak and exits 23 unless its
// environment turns leak detection off, and the program must read as agreeing. Mutant (agreement.mutants.md): the
// agreement mode's nativeEnv dropped, and the program disagrees.
func TestAgreementNativeRunsWithoutLeakDetection(t *testing.T) {
	t.Parallel()
	dir, sha, head := agreementFixture(t, []agreementSource{{"corpus/leaks.a", "print: 42\nnative: echo 42\nnative: case \"$ASAN_OPTIONS\" in *detect_leaks=0*) ;; *) echo '==1==ERROR: LeakSanitizer: detected memory leaks' >&2; exit 23;; esac\n"}})
	o, r := agreementCommand(t, dir, sha, head)
	if o.Exit != 0 || r.Agreeing != 1 || r.Disagreeing != 0 || !strings.HasSuffix(o.Stderr, "\nagreement: compared 1, agreeing 1, ruled 0, checked stops 0, pending 0, disagreeing 0, memory-unsafe 0, refused 0\n") {
		t.Fatalf("exit %d, %+v, want the program agreeing with leak detection off: %s", o.Exit, r, o.Stderr)
	}
}

// Accepted programs run on --workers at once, and the output follows the manifest's order whatever order they finish
// in: the first program's native sleeps, so with four workers it finishes last, and the summary, the disagreement
// lines, the report's disagreements and its programs read as they do with one worker. Mutant (agreement.mutants.md):
// the disagreements collected in the worker as each program is judged, and they come out of order.
func TestAgreementWorkersKeepManifestOrder(t *testing.T) {
	t.Parallel()
	dir, sha, head := agreementFixture(t, []agreementSource{
		{"corpus/z.a", "print: 42\nnative: sleep 1\nnative: echo 41\n"},
		{"corpus/a.a", agreeingSource},
		{"corpus/m.a", disagreeingSource},
		{"corpus/b.a", "print: 42\nnative: echo 40\n"},
		{"corpus/r.a", refusedSource},
	})
	type outcome struct {
		summary       string
		lines         []string
		disagreements []string
		programs      []string
	}
	read := func(workers string) outcome {
		o, r := agreementCommand(t, dir, sha, head, "--workers", workers)
		if o.Exit != 1 {
			t.Fatalf("--workers %s: exit %d, want 1: %s", workers, o.Exit, o.Stderr)
		}
		var seen outcome
		stderr := strings.Split(strings.TrimSuffix(o.Stderr, "\n"), "\n")
		seen.summary = stderr[len(stderr)-1]
		for _, line := range stderr {
			if strings.HasPrefix(line, "disagreeing ") {
				seen.lines = append(seen.lines, line)
			}
		}
		for _, disagreement := range r.Disagreements {
			seen.disagreements = append(seen.disagreements, disagreement.Path)
		}
		for _, judged := range r.Programs {
			seen.programs = append(seen.programs, judged.Path)
		}
		return seen
	}
	one, four := read("1"), read("4")
	if !reflect.DeepEqual(one, four) {
		t.Fatalf("--workers 1 read %+v, --workers 4 read %+v", one, four)
	}
	if want := []string{"corpus/z.a", "corpus/m.a", "corpus/b.a"}; !reflect.DeepEqual(four.disagreements, want) || len(four.lines) != 3 || !strings.HasPrefix(four.lines[0], "disagreeing corpus/z.a:") {
		t.Fatalf("disagreements %v and lines %q, want %v in manifest order", four.disagreements, four.lines, want)
	}
	if four.summary != "agreement: compared 4, agreeing 1, ruled 0, checked stops 0, pending 0, disagreeing 3, memory-unsafe 0, refused 1" {
		t.Fatalf("summary %q", four.summary)
	}
}

// heldStopSource is a program both backends stop on alike, exit 70 with message on stderr after printing 1, where
// Node prints 1 and 2 and exits 0: a both-stop.
func heldStopSource(message string) string {
	return "print: 1\nprint: 2\njs-print: 1\njs-stop: " + message + "\nnative: echo 1\nnative: echo '" + message + "' >&2\nnative: exit 70\n"
}

// heldStopsFile is a held-stop list whose rows hold each path given at the stop message given.
func heldStopsFile(t *testing.T, stops map[string]string) string {
	t.Helper()
	list := heldStopList{Comment: "Adding a row is a ruling, and a row names the test that holds its stop", Stops: []heldStop{}}
	for path, message := range stops {
		list.Stops = append(list.Stops, heldStop{Path: path, Stderr: message + "\n", Test: "agreement_test.go: TestAgreementHeldStops"})
	}
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// A both-stop counts as a checked stop only at a path the held-stop list names, with the backends' stderr the row's
// byte for byte: held.a is one; changed.a stops with a message one byte off its row's, and unlisted.a has no row, so
// both disagree and fail the run. Without them the run passes, a checked stop never red. Mutant
// (agreement.mutants.md): the stderr comparison dropped, so changed.a's listed path alone makes it a checked stop.
func TestAgreementHeldStops(t *testing.T) {
	t.Parallel()
	const message = "adamic: panic: held for the test"
	list := heldStopsFile(t, map[string]string{"corpus/held.a": message, "corpus/changed.a": message})
	dir, sha, head := agreementFixtureWith(t, []agreementSource{
		{"corpus/held.a", heldStopSource(message)},
		{"corpus/changed.a", heldStopSource("adamic: panic: held for the tesd")},
		{"corpus/unlisted.a", heldStopSource(message)},
	}, map[string]string{"cloud/admission-corpus/held-stops.json": list})
	o, r := agreementCommand(t, dir, sha, head)
	if o.Exit != 1 || r.CheckedStops != 1 || r.Disagreeing != 2 || r.Compared != 3 || !strings.HasSuffix(o.Stderr, "\nagreement: compared 3, agreeing 0, ruled 0, checked stops 1, pending 0, disagreeing 2, memory-unsafe 0, refused 0\n") {
		t.Fatalf("exit %d, %+v, want one checked stop and two disagreements: %s", o.Exit, r, o.Stderr)
	}
	judgments := map[string]string{}
	for _, judged := range r.Programs {
		judgments[judged.Path] = judged.Judgment
	}
	if want := map[string]string{"corpus/held.a": "checked-stop", "corpus/changed.a": "disagreeing", "corpus/unlisted.a": "disagreeing"}; !reflect.DeepEqual(judgments, want) {
		t.Fatalf("judgments %v, want %v", judgments, want)
	}
	dir, sha, head = agreementFixtureWith(t, []agreementSource{{"corpus/held.a", heldStopSource(message)}}, map[string]string{"cloud/admission-corpus/held-stops.json": list})
	o, r = agreementCommand(t, dir, sha, head)
	if o.Exit != 0 || r.Verdict != "pass" || r.CheckedStops != 1 || !strings.HasSuffix(o.Stderr, "\nagreement: compared 1, agreeing 0, ruled 0, checked stops 1, pending 0, disagreeing 0, memory-unsafe 0, refused 0\n") {
		t.Fatalf("a corpus whose only stop is held: exit %d, %+v, want exit 0: %s", o.Exit, r, o.Stderr)
	}
}

// A program that disagrees and has a pending sidecar naming a task is pending, its task in its entry and on its own
// line, counted apart and never red. A sidecar naming a branch names no task to check, so its program disagrees. An
// agreeing program with a sidecar is agreeing. Mutant (agreement.mutants.md): the branch form excused, so branch.a is
// pending too and the run exits 0.
func TestAgreementPendingSidecars(t *testing.T) {
	t.Parallel()
	dir, sha, head := agreementFixtureWith(t, []agreementSource{
		{"corpus/task.a", disagreeingSource},
		{"corpus/branch.a", disagreeingSource},
		{"corpus/agree.a", agreeingSource},
	}, map[string]string{
		"corpus/task.pending":   "awaits #abc1234: reason\n",
		"corpus/branch.pending": "awaits some/branch: reason\n",
		"corpus/agree.pending":  "awaits #abc1234: reason\n",
	})
	o, r := agreementCommand(t, dir, sha, head)
	if o.Exit != 1 || r.Pending != 1 || r.Disagreeing != 1 || r.Agreeing != 1 || r.Compared != 3 || !strings.HasSuffix(o.Stderr, "\nagreement: compared 3, agreeing 1, ruled 0, checked stops 0, pending 1, disagreeing 1, memory-unsafe 0, refused 0\n") {
		t.Fatalf("exit %d, %+v, want one each of pending, disagreeing and agreeing: %s", o.Exit, r, o.Stderr)
	}
	type judgment struct{ judgment, task string }
	judgments := map[string]judgment{}
	for _, judged := range r.Programs {
		judgments[judged.Path] = judgment{judged.Judgment, judged.Task}
	}
	if want := map[string]judgment{"corpus/task.a": {"pending", "abc1234"}, "corpus/branch.a": {"disagreeing", ""}, "corpus/agree.a": {"agreeing", ""}}; !reflect.DeepEqual(judgments, want) {
		t.Fatalf("judgments %v, want %v", judgments, want)
	}
	if len(r.Disagreements) != 1 || r.Disagreements[0].Path != "corpus/branch.a" {
		t.Fatalf("disagreements: %+v, want branch.a alone", r.Disagreements)
	}
	if line := stderrLine(o.Stderr, "pending corpus/task.a:"); line != "pending corpus/task.a: awaits #abc1234" {
		t.Fatalf("the pending line: %q in %s", line, o.Stderr)
	}
}

// A sidecar never excuses a memory error: a program whose native run shows a sanitizer's report, filed by its sidecar
// under an open task, disagrees and is counted memory-unsafe, since that is how #sq84qqh's five release segfaults hid
// as library gaps.
func TestAgreementPendingNeverExcusesAMemoryError(t *testing.T) {
	t.Parallel()
	dir, sha, head := agreementFixtureWith(t, []agreementSource{
		{"corpus/unsafe.a", "print: 42\nnative: echo 42\nnative: echo '==1==ERROR: AddressSanitizer: heap-use-after-free on address 0x1' >&2\nnative: exit 1\n"},
	}, map[string]string{"corpus/unsafe.pending": "awaits #abc1234: reason\n"})
	o, r := agreementCommand(t, dir, sha, head)
	if o.Exit != 1 || r.Pending != 0 || r.Disagreeing != 1 || r.MemoryUnsafe != 1 || !strings.HasSuffix(o.Stderr, "\nagreement: compared 1, agreeing 0, ruled 0, checked stops 0, pending 0, disagreeing 1, memory-unsafe 1, refused 0\n") {
		t.Fatalf("exit %d, %+v, want one memory-unsafe disagreement, its sidecar not excusing it: %s", o.Exit, r, o.Stderr)
	}
}

// The outcome hash covers each of the three runs' exit, stdout and stderr: a change in any one backend's stderr alone
// changes it, a byte moved from one stream to the next changes it, and the same runs hash alike. In a run, an agreeing
// program whose native stderr changes, which agreement never reads, still shows the change in its hash.
func TestAgreementOutcomeHash(t *testing.T) {
	t.Parallel()
	node, javascript, native := observation{Stdout: "42\n"}, observation{Stdout: "42\n"}, observation{Stdout: "42\n"}
	same := outcomeHash(node, javascript, native)
	if len(same) != 64 || outcomeHash(node, javascript, native) != same {
		t.Fatalf("hash %q isn't one hex SHA-256 the same runs make alike", same)
	}
	seen := map[string]string{same: "the runs as they were"}
	changed := func(o observation) observation {
		o.Stderr = "warning\n"
		return o
	}
	for name, hash := range map[string]string{
		"node's stderr":                     outcomeHash(changed(node), javascript, native),
		"javascript's stderr":               outcomeHash(node, changed(javascript), native),
		"native's stderr":                   outcomeHash(node, javascript, changed(native)),
		"native's newline moved to stderr":  outcomeHash(node, javascript, observation{Stdout: "42", Stderr: "\n"}),
		"native's exit":                     outcomeHash(node, javascript, observation{Exit: 1, Stdout: "42\n"}),
		"javascript's stdout moved to node": outcomeHash(observation{Stdout: "42\n42\n"}, observation{}, native),
	} {
		if other, collides := seen[hash]; collides {
			t.Errorf("%s hashes as %s does", name, other)
		}
		seen[hash] = name
	}
	read := func(warning string) agreementProgram {
		dir, sha, head := agreementFixture(t, []agreementSource{{"corpus/warns.a", "print: 42\nnative: echo 42\nnative: echo " + warning + " >&2\n"}})
		o, r := agreementCommand(t, dir, sha, head)
		if o.Exit != 0 || len(r.Programs) != 1 || r.Programs[0].Judgment != "agreeing" {
			t.Fatalf("exit %d, %+v, want warns.a agreeing: %s", o.Exit, r.Programs, o.Stderr)
		}
		return r.Programs[0]
	}
	if warning, warnings := read("warning"), read("warnings"); warning.OutcomeHash == warnings.OutcomeHash || len(warning.OutcomeHash) != 64 {
		t.Fatalf("native's stderr changed and the hash didn't: %q, %q", warning.OutcomeHash, warnings.OutcomeHash)
	}
}
