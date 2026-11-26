package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A usage error is exit 2 with the usage line, and writes nothing. Mutant: parse accepting a third word that isn't
// --link, red at "root out extra".
func TestUsageIsExitTwo(t *testing.T) {
	t.Parallel()
	for _, arguments := range [][]string{
		nil,
		{"root.a"},
		{"root.a", "out", "extra"},
		{"--module-format", "commonjs", "root.a", "out"},
		{"--module-format", "esm"},
		{"--link", "root.a", "out"},
		{"-o", "out"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(arguments, &stdout, &stderr); code != 2 || stdout.Len() != 0 || strings.TrimSpace(stderr.String()) != usage {
			t.Errorf("%q: exit %d, stdout %q, stderr %q, want exit 2 with the usage alone", arguments, code, stdout.String(), stderr.String())
		}
	}
}

func TestParseReadsTheFormatAndLink(t *testing.T) {
	t.Parallel()
	asked, ok := parse([]string{"--module-format", "commonjs-bundle", "src/tsc/tsc.ts", "out", "--link"})
	if !ok || asked.root != "src/tsc/tsc.ts" || asked.output != "out" || !asked.link || asked.format != moduleFormats["commonjs-bundle"] {
		t.Fatalf("parse: %+v, %v", asked, ok)
	}
}

// writeUnder never writes outside the directory its caller named. Mutant: the Rel check dropped, red.
func TestWriteUnderStaysInItsDirectory(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if _, err := writeUnder(directory, "../escaped.txt", "x"); err == nil {
		t.Fatal("writeUnder wrote ../escaped.txt")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(directory), "escaped.txt")); !os.IsNotExist(err) {
		t.Fatalf("a file was written beside the directory: %v", err)
	}
}

// The report counts every trap by phase, kind and reason, and counts the traps lower couldn't place in a unit, which
// traps.md names as a measurement gap.
// Sites are written relative to the deepest directory holding the root and every site, so a program rooted at
// src/tsc/tsc.ts reads its src/compiler sites as compiler/...; a site with no path stays as it came, and so does every
// site when the only directory they share is the filesystem's root. A site with no ':' is a file with no position,
// the file alone a stop names, and an absolute one is placed too, its directory counted toward the common one.
// Mutants: relativeSites returning traps unchanged, red; the common-directory loop skipping a site with no ':', red
// at the file-only row, which reads ../compiler/types.ts.
func TestSitesAreRelativeToTheProgramsCommonDirectory(t *testing.T) {
	t.Parallel()
	tree := t.TempDir()
	placed, judgments := relativeSites(filepath.Join(tree, "src", "tsc"), []ir.Trap{
		{Site: filepath.Join(tree, "src", "compiler", "checker.ts") + ":3:5"},
		{Site: filepath.Join(tree, "src", "tsc", "tsc.ts") + ":1:1"},
		{Site: "a.ts:2:2"},
	}, []ir.TrapJudgmentAbandoned{{Pass: "checkSomething", Site: filepath.Join(tree, "src", "compiler", "utilities.ts") + ":7:1", Why: "Echo"}, {Pass: "checkAccessorSpreads", Site: "checker.ts", Why: "Gap"}})
	if placed[0].Site != "compiler/checker.ts:3:5" || placed[1].Site != "tsc/tsc.ts:1:1" || placed[2].Site != "a.ts:2:2" {
		t.Fatalf("relativeSites: %+v", placed)
	}
	if judgments[0].Site != "compiler/utilities.ts:7:1" || judgments[1].Site != "checker.ts" || judgments[0].Pass != "checkSomething" {
		t.Fatalf("relativeSites' abandoned judgments: %+v", judgments)
	}
	// A file with no position is placed as any site is, and its directory counts toward the common one.
	placed, judgments = relativeSites(filepath.Join(tree, "src", "tsc"), []ir.Trap{{Site: filepath.Join(tree, "src", "tsc", "tsc.ts") + ":1:1"}}, []ir.TrapJudgmentAbandoned{{Pass: "voidDiscards", Site: filepath.Join(tree, "src", "compiler", "types.ts"), Why: "Gap"}})
	if placed[0].Site != "tsc/tsc.ts:1:1" || judgments[0].Site != "compiler/types.ts" {
		t.Fatalf("relativeSites with a file-only site: %+v, %+v; want tsc/tsc.ts:1:1 and compiler/types.ts", placed, judgments)
	}
	apart, _ := relativeSites(filepath.Join(tree, "src"), []ir.Trap{{Site: filepath.Join(tree, "src", "a.ts") + ":1:1"}, {Site: "/b.ts:1:1"}}, nil)
	if apart[0].Site != filepath.Join(tree, "src", "a.ts")+":1:1" || apart[1].Site != "/b.ts:1:1" {
		t.Fatalf("sites sharing only the filesystem's root must stay as they came: %+v", apart)
	}
}

func TestReportCountsAndNamesTheGap(t *testing.T) {
	t.Parallel()
	traps := []ir.Trap{
		{Kind: "Refused", Reason: "sort without a comparator", Site: "a.a:3:5", Unit: "a.a second", Phase: "Body"},
		{Kind: "Refused", Reason: "sort without a comparator", Site: "a.a:9:5", Unit: "a.a third", Phase: "Body"},
		{Kind: "NotYet", Reason: "a pass's stop", Site: "a.a:12:1", Unit: "", Phase: "checkSlotWrites"},
		{Kind: "Panic", Reason: "index out of range", Site: "a.a:20:1", Unit: "a.a fourth", Phase: "Body", Suspect: true},
		{Kind: "NotYet", Reason: "an echo", Site: "a.a:21:1", Unit: "a.a fourth", Phase: "findCycles", Echo: true},
	}
	counted := summarize(traps, []string{"normalizeAsync"}, nil)
	if counted.Total != 5 || counted.Suspect != 1 || counted.Echo != 1 || len(counted.RewritesSkipped) != 1 || counted.ByPhase["Body"] != 3 || counted.ByPhase["checkSlotWrites"] != 1 || counted.ByKind["Refused"] != 2 ||
		counted.ByReason["Refused: sort without a comparator"] != 2 || counted.Unplaced != 1 {
		t.Fatalf("summarize: %+v", counted)
	}
	if counted.Findings != findingsRule {
		t.Fatalf("traps.json must open with the findings rule: %q", counted.Findings)
	}
	text := markdown(counted)
	if !strings.HasPrefix(text, "Findings: "+findingsRule+".") {
		t.Errorf("traps.md must open with the findings rule:\n%s", text)
	}
	for _, want := range []string{"5 traps, 1 of them unplaced", ", 1 of them suspect", ", 1 of them echoes", "Rewrites skipped past a stop, in order: normalizeAsync.", "## Body, 3", "| 2 | Refused | sort without a comparator | a.a:3:5 |", "## checkSlotWrites, 1"} {
		if !strings.Contains(text, want) {
			t.Errorf("traps.md lacks %q:\n%s", want, text)
		}
	}
	if empty := summarize(nil, nil, nil); empty.Traps == nil || empty.RewritesSkipped == nil || empty.JudgmentsAbandoned == nil || empty.Total != 0 {
		t.Fatalf("no traps should write an empty list, not null: %+v", empty)
	}
}

// An abandoned judgment whose stop names no site, an Error's, reads "with an error that names no site" in the wall
// line and in traps.md, never "at" and a blank; one with a site reads at it. Mutant: abandonedAt's empty case
// dropped, red at both.
func TestAnAbandonedJudgmentWithNoSiteSaysSo(t *testing.T) {
	t.Parallel()
	unsited := ir.TrapJudgmentAbandoned{Pass: "checkSomething", Why: "Error"}
	sited := ir.TrapJudgmentAbandoned{Pass: "voidDiscards", Site: "voids.a", Why: "Gap"}
	if line := incomplete(unsited); line != "adamic-firstlink: the program's judgment is incomplete, so nothing was emitted: checkSomething gave up with an error that names no site (Error), and every unit past it went unjudged" {
		t.Errorf("the wall line: %q", line)
	}
	if line := incomplete(sited); !strings.Contains(line, ": voidDiscards gave up at voids.a (Gap), and") {
		t.Errorf("the wall line of a sited judgment: %q", line)
	}
	text := markdown(summarize(nil, nil, []ir.TrapJudgmentAbandoned{unsited, sited}))
	if !strings.Contains(text, "Judgments abandoned past a stop, in order: checkSomething with an error that names no site (Error); voidDiscards at voids.a (Gap).") {
		t.Errorf("traps.md doesn't say the judgment names no site:\n%s", text)
	}
}

// twoFunctions prints from its first function and calls a second one that sorts without a comparator, which Adamic
// refuses (docs/rulings.md, row 496), so in trap mode the second lowers as a trap.
const twoFunctions = `function first(): void {
    console.log("first");
}
function second(): number {
    const values: number[] = [3, 1, 2];
    values.sort();
    return values.length;
}
first();
console.log(String(second()));
`

// The door lowers in trap mode and writes its report, C and JavaScript under the directory it was given, and nothing
// else, exiting 0 with one trap: the second function's, in phase Body at its sort. With --link the binary runs the
// first function and stops at the trap with exit 75 and the trap's prefix, never reading as a program that ran.
// It runs main's own backends, whose guard refuses a trap-mode program unless their options set FirstLinkTraps
// (#w2mmzwt), so it pins that the door sets it on all three roads. Mutants: lowering without FirstLinkTraps, red here
// (exit 1 at lower, the refusal); the C's EmitOptions without it, red here (native's guard panics at the emit);
// the JavaScript twin's Options without it, red here (javascript's guard panics). The --link build's Options carry
// it too, but no flag or key reads it there (internal/native_runtime/options.go), so dropping it there is no mutant.
func TestTheDoorReportsATrapAndTheBinaryStopsAtIt(t *testing.T) {
	t.Parallel()
	work := t.TempDir()
	root := filepath.Join(work, "two.a")
	if err := os.WriteFile(root, []byte(twoFunctions), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(work, "out")
	var stdout, stderr bytes.Buffer
	if code := run([]string{root, output, "--link"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d, stderr %s", code, stderr.String())
	}
	printed := strings.Fields(stdout.String())
	want := []string{"traps.json", "traps.md", "c/main.c", "js/main.js", "firstlink"}
	if len(printed) != len(want) {
		t.Fatalf("printed %q, want %d paths", printed, len(want))
	}
	for index, name := range want {
		if printed[index] != filepath.Join(output, name) {
			t.Errorf("path %d: %s, want %s", index, printed[index], filepath.Join(output, name))
		}
	}
	var found []string
	if err := filepath.Walk(work, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			found = append(found, path)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(found)
	for _, path := range found {
		if path != root && !strings.HasPrefix(path, output+string(filepath.Separator)) {
			t.Errorf("wrote %s, outside the output directory", path)
		}
	}
	var counted report
	data, err := os.ReadFile(filepath.Join(output, "traps.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &counted); err != nil {
		t.Fatal(err)
	}
	if counted.Total != 1 || counted.Traps[0].Phase != "Body" || counted.Traps[0].Kind != "Refused" || !strings.HasPrefix(counted.Traps[0].Site, "two.a:6:") {
		t.Fatalf("traps.json: %s", data)
	}
	binary := exec.Command(filepath.Join(output, "firstlink"))
	var runOut, runErr bytes.Buffer
	binary.Stdout, binary.Stderr = &runOut, &runErr
	err = binary.Run()
	exit, isExit := err.(*exec.ExitError)
	if !isExit || exit.ExitCode() != 75 || runOut.String() != "first\n" || !strings.HasPrefix(runErr.String(), "adamic: first-link trap: Refused: ") {
		t.Fatalf("binary: %v, stdout %q, stderr %q, want exit 75 after first, at the trap", err, runOut.String(), runErr.String())
	}
}

// voidsThatDisagree makes a void call through an interface whose class methods of the name disagree on their result,
// an object's and a number's (internal/oracle's void_discards_refused/methods.a, its template literals taken out),
// which voidDiscards, a judging pass, stops at with a NotYet whose Where names only the file, as checkAccessorSpreads'
// did on tsc before it named its spread (#nxb3xh1). In trap mode no unit holds that stop, so the pass gives up there:
// a Gap, and an abandoned judgment.
const voidsThatDisagree = `interface Box {
    readonly label: string;
}
interface Action {
    run(): void;
}
class Maker implements Action {
    name: string;
    constructor(name: string) {
        this.name = name;
    }
    run(): Box {
        return { label: this.name };
    }
}
class Counter implements Action {
    count = 0;
    run(): number {
        this.count++;
        return this.count;
    }
}
const actions: Action[] = [new Maker("a"), new Counter()];
for (const action of actions) {
    action.run();
}
console.log(String(actions.length));
`

// A lowering in which a judging pass gave up left units nobody judged, so the door writes traps.json and traps.md,
// each naming the pass, its site and why, and exits 1 naming the first as the wall, with no C, no JavaScript and no
// binary written. Mutant: the door's check of TrapJudgmentsAbandoned dropped, so it emits and exits 0, red here.
func TestTheDoorWontEmitPastAnAbandonedJudgment(t *testing.T) {
	t.Parallel()
	work := t.TempDir()
	root := filepath.Join(work, "voids.a")
	if err := os.WriteFile(root, []byte(voidsThatDisagree), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(work, "out")
	var stdout, stderr bytes.Buffer
	code := run([]string{root, output, "--link"}, &stdout, &stderr)
	const wall = "voidDiscards gave up at voids.a (Gap)"
	if code != 1 || !strings.Contains(stderr.String(), "adamic-firstlink: the program's judgment is incomplete, so nothing was emitted: "+wall) || strings.Contains(stderr.String(), "compiler bug") {
		t.Fatalf("exit %d, stderr %q; want exit 1 naming %q as the wall", code, stderr.String(), wall)
	}
	if printed := strings.Fields(stdout.String()); len(printed) != 2 || printed[0] != filepath.Join(output, "traps.json") || printed[1] != filepath.Join(output, "traps.md") {
		t.Fatalf("printed %q, want traps.json and traps.md alone", printed)
	}
	for _, name := range []string{"c/main.c", "js/main.js", "firstlink"} {
		if _, err := os.Stat(filepath.Join(output, name)); !os.IsNotExist(err) {
			t.Errorf("%s was written past an abandoned judgment: %v", name, err)
		}
	}
	var counted report
	data, err := os.ReadFile(filepath.Join(output, "traps.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &counted); err != nil {
		t.Fatal(err)
	}
	want := ir.TrapJudgmentAbandoned{Pass: "voidDiscards", Site: "voids.a", Why: "Gap"}
	if len(counted.JudgmentsAbandoned) == 0 || counted.JudgmentsAbandoned[0] != want {
		t.Fatalf("traps.json's judgments_abandoned: %s; want %+v first", data, want)
	}
	text, err := os.ReadFile(filepath.Join(output, "traps.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(text), "Judgments abandoned past a stop, in order: voidDiscards at voids.a (Gap)") {
		t.Fatalf("traps.md doesn't name the abandoned judgment:\n%s", text)
	}
}
