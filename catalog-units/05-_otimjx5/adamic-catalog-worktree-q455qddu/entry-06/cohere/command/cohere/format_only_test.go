package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/edit"
)

// TestTheFormatOnlyGateExitsOnFormattingAlone is the commit gate's command, through the binary: `--no-fix
// --format-only` exits 0 over a tree whose only findings are lint's, and nonzero, naming the file, when
// formatting would change one in the repository or in a nested repository it reads, or when the formatter
// could not read one there.
func TestTheFormatOnlyGateExitsOnFormattingAlone(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	debugged := "export function value(): number {\n  debugger;\n  return 1;\n}\n"
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{"no-debugger":"error"}}`,
		"NexusCohereSettings.json": `{"format":{"ignore":["tsconfig.json"]}}`,
		".gitignore":               ".cache/\n",
		// Formatted, and a lint finding with a fix: the gate's verdict must not hear it.
		"Producer.ts":         debugged,
		".gitmodules":         "[submodule \"library\"]\n\tpath = library\n\turl = ../library\n",
		"library/.git":        "gitdir: ../.git/modules/library\n",
		"library/Inner.ts":    "export const inner = 1;\n",
		"library/Debugged.ts": debugged,
	})
	// The settings files are formatted by a writing run first, which proposes no fixes either.
	if output, code := runCohere(t, binary, root, "--format-only"); code != 0 {
		t.Fatalf("the writing format-only run exited %d:\n%s", code, output)
	}
	if got := readForTest(t, filepath.Join(root, "Producer.ts")); got != debugged {
		t.Fatalf("--format-only applied a lint fix:\n%s", got)
	}
	// A file formatting does change is re-linted on its next pass in a run that fixes. Here nothing may
	// propose: it is formatted, and its finding is left for the run that fixes.
	writeTree(t, root, map[string]string{"Reformatted.ts": "export function other(): number {\n  debugger;\n  return   2;\n}\n"})
	if output, code := runCohere(t, binary, root, "--format-only", "Reformatted.ts"); code != 0 {
		t.Fatalf("the writing format-only run exited %d:\n%s", code, output)
	}
	if got := readForTest(t, filepath.Join(root, "Reformatted.ts")); got != "export function other(): number {\n  debugger;\n  return 2;\n}\n" {
		t.Fatalf("--format-only did not format the file, or fixed it once formatting changed it:\n%s", got)
	}

	gate := func(label string, wantExit int, wantLines ...string) {
		t.Helper()
		for _, arguments := range [][]string{{"--no-fix", "--format-only"}, {"--no-fix", "--format-only", "--format-all"}} {
			output, code := runCohere(t, binary, root, arguments...)
			if code != wantExit {
				t.Fatalf("%s, %v: exit %d, want %d:\n%s", label, arguments, code, wantExit, output)
			}
			for _, line := range wantLines {
				if !strings.Contains(output, line) {
					t.Fatalf("%s, %v: missing %q:\n%s", label, arguments, line, output)
				}
			}
			if !strings.Contains(output, "fix checked in") || !strings.Contains(output, "(formatting only, no fixes proposed)") ||
				!strings.Contains(output, "types skipped (not requested) · lint skipped (not requested)") {
				t.Fatalf("%s, %v: the phase line does not say what ran:\n%s", label, arguments, output)
			}
		}
	}

	// Nothing to format, lint findings present here and in the library: 0. A whole run says otherwise.
	gate("lint findings and nothing to format", 0, "0 files would change")
	if _, code := runCohere(t, binary, root, "--no-fix"); code == 0 {
		t.Fatalf("the lint finding was not a finding to a whole run, so this test proves nothing")
	}

	inner := filepath.Join(root, "library", "Inner.ts")
	writeTree(t, root, map[string]string{"library/Inner.ts": "export const inner   =   1;\n"})
	gate("an unformatted file in a nested repository", 1, inner+":1:1 - the formatter would rewrite this file in nested repository library")
	writeTree(t, root, map[string]string{"library/Inner.ts": "export const inner = 1;\n"})

	ugly := filepath.Join(root, "Ugly.ts")
	writeTree(t, root, map[string]string{"Ugly.ts": "export const ugly   =   1;\n"})
	gate("an unformatted file in the repository", 1, ugly+":1:1 - --fix would rewrite this file: format [fix/would-change]")
	removeForTest(t, ugly)

	broken := filepath.Join(root, "broken.json")
	writeTree(t, root, map[string]string{"broken.json": "{ \"a\": \n"})
	gate("a file the formatter cannot read in the repository", 1,
		broken+":1:1 - the formatter could not check this file: "+unparseableSkipReason+" [format/unchecked]")
	removeForTest(t, broken)

	nestedBroken := filepath.Join(root, "library", "broken.json")
	writeTree(t, root, map[string]string{"library/broken.json": "{ \"a\": \n"})
	gate("a file the formatter cannot read in a nested repository", 1,
		nestedBroken+":1:1 - the formatter could not check this file: in nested repository library, "+unparseableSkipReason+" [format/unchecked]")
	removeForTest(t, nestedBroken)

	gate("neither", 0, "0 files would change")

	for _, other := range []string{"--fix", "--types", "--lint", "--unused"} {
		output, code := runCohere(t, binary, root, "--no-fix", "--format-only", other)
		if code == 0 || !strings.Contains(output, "--format-only and "+other+" contradict each other") {
			t.Fatalf("--format-only with %s was not refused by name, exit %d:\n%s", other, code, output)
		}
	}
}

// TestOnlyAFileTheFormatterCouldNotReadIsUnchecked: a skip that is about the file (it does not parse) or a
// formatter that broke on it is a file nobody checked; a skip about the run (outside the named paths, a
// type the formatter does not handle) is a file that was never the run's to check.
func TestOnlyAFileTheFormatterCouldNotReadIsUnchecked(t *testing.T) {
	t.Parallel()
	summary := edit.Summary{
		NotTransformed: []edit.NotTransformedFile{
			{FileName: "/a/Broken.json", Reason: unparseableSkipReason},
			{FileName: "/a/Crashed.ts", Reason: "the whole-text transform failed (boom)", Failed: true},
			{FileName: "/a/Elsewhere.ts", Reason: "outside the format scope, 1 named path"},
			{FileName: "/a/notes.txt", Reason: ".txt is not a file type the formatter handles"},
		},
		FilesFailed:  []string{"/a/Unreadable.ts"},
		FilesRefused: []edit.RefusedFile{{FileName: "/a/Reprinted.ts", Rules: []string{"format"}}},
	}
	got := []string{}
	for _, file := range formatOnlyUnchecked(summary, formatScope{}) {
		got = append(got, filepath.Base(file.FileName))
	}
	if strings.Join(got, ", ") != "Broken.json, Crashed.ts, Unreadable.ts, Reprinted.ts" {
		t.Fatalf("unchecked %v", got)
	}

	failed := formatOnlyUnchecked(edit.Summary{}, formatScope{failure: os.ErrPermission})
	if len(failed) != 1 || failed[0].FileName != "" || !strings.Contains(failed[0].Reason, "the format walk failed") {
		t.Fatalf("a failed walk was not a finding: %+v", failed)
	}
}

func removeForTest(t *testing.T, fileName string) {
	t.Helper()
	if err := os.Remove(fileName); err != nil {
		t.Fatal(err)
	}
}

// TestFormatOnlyBuildsNoGraph: `--format-only` reads the disk, never the program (#m0dktbn). Its tsconfig
// here cannot build a program at all (kept out of the format scope, so only a build would read it), so a
// whole run fails building the graph, and a format-only run must not notice: it checks, formats and says
// it built no graph.
func TestFormatOnlyBuildsNoGraph(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            "{ \"compilerOptions\": \n",
		"CohereSettings.json":      "{ \"extends\": \"./NexusCohereSettings.json\" }\n",
		"NexusCohereSettings.json": "{ \"format\": { \"ignore\": [\"tsconfig.json\"] } }\n",
		".gitignore":               ".cache/\n",
		"Tidy.ts":                  "export const tidy = 1;\n",
	})
	if output, code := runCohere(t, binary, root, "--no-fix"); code == 0 || !strings.Contains(output, "building the type graph") {
		t.Fatalf("the tsconfig built a program, so this test proves nothing, exit %d:\n%s", code, output)
	}

	output, code := runCohere(t, binary, root, "--no-fix", "--format-only")
	if code != 0 || !strings.Contains(output, "graph not built: --format-only reads the disk, not the program") ||
		strings.Contains(output, "graph built in") || !strings.Contains(output, "no graph built (formatting reads none)") {
		t.Fatalf("--format-only built the graph or did not say it did not, exit %d:\n%s", code, output)
	}

	writeTree(t, root, map[string]string{"Ugly.ts": "export const ugly   =   1;\n"})
	output, code = runCohere(t, binary, root, "--no-fix", "--format-only")
	if code == 0 || !strings.Contains(output, filepath.Join(root, "Ugly.ts")+":1:1 - --fix would rewrite this file: format") {
		t.Fatalf("with no graph, an unformatted file was not reported, exit %d:\n%s", code, output)
	}
	if output, code = runCohere(t, binary, root, "--format-only", "Ugly.ts"); code != 0 || readForTest(t, filepath.Join(root, "Ugly.ts")) != "export const ugly = 1;\n" {
		t.Fatalf("with no graph, the writing run did not format the named file, exit %d:\n%s", code, output)
	}
}
