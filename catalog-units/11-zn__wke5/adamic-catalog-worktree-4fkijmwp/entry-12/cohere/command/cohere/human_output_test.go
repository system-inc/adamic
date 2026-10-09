package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// The lines are Kirk's, from the visual spec of 2026-10-04, with the glyph slots and path column filled
// in as the spec describes them.
func TestChangedFileLinesAlignMixedGlyphs(t *testing.T) {
	t.Parallel()
	files := []changedFile{
		{Path: "modules/pensieve/Recall.ts", Formatted: true},
		{Path: "app/os/SessionRow.tsx", Fixed: true, Formatted: true,
			FixedBy: map[string]int{"prefer-const": 2, "prefer-nullish-coalescing": 1}},
		{Path: "modules/data/DataSync.ts", Fixed: true, FixedBy: map[string]int{"no-useless-default-assignment": 1}},
	}
	want := []string{
		"🪄💅 app/os/SessionRow.tsx      prefer-const ×2, prefer-nullish-coalescing",
		"🪄   modules/data/DataSync.ts   no-useless-default-assignment",
		"  💅 modules/pensieve/Recall.ts",
	}
	got := changedFileLines(files, plain)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("changed files:\n got\n%s\n want\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	// The rules are dim on a terminal, and nothing else on the line is styled.
	colored := changedFileLines(files[2:], textStyle{color: true})
	if want := "🪄   modules/data/DataSync.ts \x1b[2mno-useless-default-assignment\x1b[22m"; colored[0] != want {
		t.Errorf("colored changed file:\n got  %q\n want %q", colored[0], want)
	}
}

func TestChangedFileLinesAreCapped(t *testing.T) {
	t.Parallel()
	var files []changedFile
	for index := range changedFilesShown + 5 {
		files = append(files, changedFile{Path: fmt.Sprintf("file%02d.ts", index), Formatted: true})
	}
	got := changedFileLines(files, plain)
	if len(got) != changedFilesShown+1 {
		t.Fatalf("%d files listed %d lines, want %d and a line saying how many more", len(files), len(got), changedFilesShown)
	}
	if want := "… 5 more · --verbose for all"; got[len(got)-1] != want {
		t.Errorf("the last line is %q, want %q", got[len(got)-1], want)
	}
	if got := changedFileLines(files[:changedFilesShown], plain); len(got) != changedFilesShown {
		t.Errorf("exactly %d files listed %d lines, want them all and no more-line", changedFilesShown, len(got))
	}
}

func TestFindingLine(t *testing.T) {
	t.Parallel()
	finding := runFinding{
		Path: "app/os/Session.ts", Line: 12, Column: 5, Severity: "error",
		Rule: "nexus/consistency-no-abbreviated-identifier", MessageID: "abbreviated",
		Message: "`ctx` is an abbreviation.\nWrite `context`.",
	}
	want := "app/os/Session.ts:12:5 error nexus/consistency-no-abbreviated-identifier `ctx` is an abbreviation. Write `context`."
	if got := findingLine(finding, plain); got != want {
		t.Errorf("finding line:\n got  %s\n want %s", got, want)
	}

	want = "app/os/Session.ts:12:5 \x1b[31merror\x1b[39m \x1b[2mnexus/consistency-no-abbreviated-identifier\x1b[22m `ctx` is an abbreviation. Write `context`."
	if got := findingLine(finding, colored); got != want {
		t.Errorf("colored finding line:\n got  %q\n want %q", got, want)
	}

	finding.Severity = "warn"
	want = "app/os/Session.ts:12:5 \x1b[2mwarn\x1b[22m \x1b[2mnexus/consistency-no-abbreviated-identifier\x1b[22m `ctx` is an abbreviation. Write `context`."
	if got := findingLine(finding, colored); got != want {
		t.Errorf("colored warning line:\n got  %q\n want %q", got, want)
	}
}

// TestStyleForHonorsNoColor sets NO_COLOR empty, which the convention counts as set.
// Not parallel: it sets NO_COLOR with t.Setenv, which a parallel test may not
func TestStyleForHonorsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	if style := styleFor(nil); style.color {
		t.Error("NO_COLOR is set and the style still colors")
	}
}

// TestStyleForAFileIsPlain is output going to a file, as a log or a pipe would take it: no color.
// Not parallel: it sets and unsets NO_COLOR in the process environment, which a parallel test may not
func TestStyleForAFileIsPlain(t *testing.T) {
	// Set first so the test restores whatever the environment held, then unset.
	t.Setenv("NO_COLOR", "")
	os.Unsetenv("NO_COLOR")
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if style := styleFor(file); style.color {
		t.Error("output to a file is styled for a terminal")
	}
}
