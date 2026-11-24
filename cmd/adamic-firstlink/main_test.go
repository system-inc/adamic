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
func TestReportCountsAndNamesTheGap(t *testing.T) {
	t.Parallel()
	traps := []ir.Trap{
		{Kind: "Refused", Reason: "sort without a comparator", Site: "a.a:3:5", Unit: "a.a second", Phase: "Body"},
		{Kind: "Refused", Reason: "sort without a comparator", Site: "a.a:9:5", Unit: "a.a third", Phase: "Body"},
		{Kind: "NotYet", Reason: "a pass's stop", Site: "a.a:12:1", Unit: "", Phase: "checkSlotWrites"},
		{Kind: "Panic", Reason: "index out of range", Site: "a.a:20:1", Unit: "a.a fourth", Phase: "Body", Suspect: true},
	}
	counted := summarize(traps)
	if counted.Total != 4 || counted.Suspect != 1 || counted.ByPhase["Body"] != 3 || counted.ByPhase["checkSlotWrites"] != 1 || counted.ByKind["Refused"] != 2 ||
		counted.ByReason["Refused: sort without a comparator"] != 2 || counted.Unplaced != 1 {
		t.Fatalf("summarize: %+v", counted)
	}
	text := markdown(counted)
	for _, want := range []string{"4 traps, 1 of them unplaced", ", 1 of them suspect", "## Body, 3", "| 2 | Refused | sort without a comparator | a.a:3:5 |", "## checkSlotWrites, 1"} {
		if !strings.Contains(text, want) {
			t.Errorf("traps.md lacks %q:\n%s", want, text)
		}
	}
	if empty := summarize(nil); empty.Traps == nil || empty.Total != 0 {
		t.Fatalf("no traps should write an empty list, not null: %+v", empty)
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
    return values[0];
}
first();
console.log(String(second()));
`

// The door lowers in trap mode and writes its report, C and JavaScript under the directory it was given, and nothing
// else, exiting 0 with one trap: the second function's, in phase Body at its sort. With --link the binary runs the
// first function and stops at the trap with exit 75 and the trap's prefix, never reading as a program that ran.
// Mutant: run lowering without FirstLinkTraps, red here (exit 1 at lower, the refusal).
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
	if counted.Total != 1 || counted.Traps[0].Phase != "Body" || counted.Traps[0].Kind != "Refused" || !strings.Contains(counted.Traps[0].Site, "two.a:6:") {
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
