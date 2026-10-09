package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatfiles"
)

// treeSnapshot reads every file under a root, so a run can be held to having written nothing at all:
// no changed byte, and no new file either.
func treeSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkError error) error {
		// cohere's own cache for the project is outside `--no-fix`'s promise, which is about source: it is
		// cohere's to keep, and `--no-cache` is the flag that writes none of it. So is a nested
		// repository's, where the drift check keeps that repository's format record.
		if walkError == nil && entry.IsDir() && filepath.Base(filepath.Dir(path)) == ".cache" {
			if owner := filepath.Dir(filepath.Dir(path)); path == cacheDirectory(owner) && (owner == root || formatfiles.HasOwnRepository(owner)) {
				return filepath.SkipDir
			}
		}
		if walkError != nil || entry.IsDir() {
			return walkError
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snapshot[path] = string(contents)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func assertTreeUnchanged(t *testing.T, root string, before map[string]string, output string) {
	t.Helper()
	after := treeSnapshot(t, root)
	for path, contents := range before {
		if after[path] != contents {
			t.Errorf("--no-fix changed %s:\nbefore:\n%s\nafter:\n%s\noutput:\n%s", path, contents, after[path], output)
		}
	}
	for path := range after {
		if _, existed := before[path]; !existed {
			t.Errorf("--no-fix created %s\noutput:\n%s", path, output)
		}
	}
}

// TestNoFixReportsWhatFixAndFormatWouldChange is `--no-fix`'s promise, both halves: it reports every
// file `--fix` would rewrite, naming the formatter or the fixing rule, and it writes nothing.
//
// It used to skip the fix phase, so an unformatted file passed a clean `--no-fix` run with
// `lint: 0 findings`. Each step's control is the writing run: a file the check reports is one `--fix`
// then rewrites, and after that rewrite the check is silent, so the two cannot disagree in either
// direction.
func TestNoFixReportsWhatFixAndFormatWouldChange(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{"no-debugger":"error"}}`,
		"NexusCohereSettings.json": `{"format":{"ignore":[]}}`,
		"Clean.ts":                 "export const clean = 1;\n",
	})

	// Whatever the formatter wants of the fixture's own files, a writing run settles it. `--format-all`
	// alone, without `--format`, because naming the scope asks for formatting.
	if output, code := runCohere(t, binary, root, "--fix", "--format-all"); code != 0 {
		t.Fatalf("settling the fixture exited %d:\n%s", code, output)
	}

	// A formatted tree is silent, and the phase line says it looked.
	before := treeSnapshot(t, root)
	output, code := runCohere(t, binary, root, "--no-fix", "--format-all")
	if code != 0 {
		t.Fatalf("a formatted tree failed --no-fix with exit %d:\n%s", code, output)
	}
	if !strings.Contains(output, "fix checked in") || !strings.Contains(output, "0 files would change") {
		t.Errorf("the phase line does not say the fix phase was checked:\n%s", output)
	}
	if strings.Contains(output, "would-change") {
		t.Errorf("a formatted tree reported a file that would change:\n%s", output)
	}
	assertTreeUnchanged(t, root, before, output)

	// An unformatted file is a finding against that file, naming the formatter, and stays as it was.
	unformatted := "export   const  spaced = {a:1}\n"
	writeTree(t, root, map[string]string{"Unformatted.ts": unformatted})
	before = treeSnapshot(t, root)
	output, code = runCohere(t, binary, root, "--no-fix", "--format-all")
	if code == 0 {
		t.Fatalf("an unformatted file passed --no-fix:\n%s", output)
	}
	want := filepath.Join(root, "Unformatted.ts") + ":1:1 - --fix would rewrite this file: format [fix/would-change]"
	if !strings.Contains(output, want) {
		t.Errorf("want %q in:\n%s", want, output)
	}
	if !strings.Contains(output, "1 file would change") || !strings.Contains(output, "1 would be reformatted") {
		t.Errorf("the count did not reach the phase line and the summary:\n%s", output)
	}
	assertTreeUnchanged(t, root, before, output)

	// The control: the writing run rewrites exactly the file the check named, and the check is then
	// silent again.
	if output, code := runCohere(t, binary, root, "--fix", "--format-all"); code != 0 {
		t.Fatalf("the writing run exited %d:\n%s", code, output)
	}
	if readForTest(t, filepath.Join(root, "Unformatted.ts")) == unformatted {
		t.Fatal("--fix left the file --no-fix reported, so the two disagree")
	}
	if output, code := runCohere(t, binary, root, "--no-fix", "--format-all"); code != 0 {
		t.Fatalf("--no-fix still failed after --fix wrote what it reported, exit %d:\n%s", code, output)
	}

	// A fixable finding with formatting left out names its rule alone. A bare --no-fix formats too (#b1sjy7b),
	// so it names the formatter beside the rule, since the text the fix leaves is reformatted.
	writeTree(t, root, map[string]string{"Debugger.ts": "export function stop(): void {\n    debugger;\n}\n"})
	before = treeSnapshot(t, root)
	for _, testCase := range []struct {
		arguments []string
		changers  string
	}{
		{[]string{"--no-fix", "--no-format"}, "no-debugger"},
		{[]string{"--no-fix"}, "no-debugger, format"},
	} {
		output, code = runCohere(t, binary, root, testCase.arguments...)
		if code == 0 {
			t.Fatalf("a fixable finding passed %v:\n%s", testCase.arguments, output)
		}
		want = filepath.Join(root, "Debugger.ts") + ":1:1 - --fix would rewrite this file: " + testCase.changers + " [fix/would-change]"
		if !strings.Contains(output, want) {
			t.Errorf("%v: want %q in:\n%s", testCase.arguments, want, output)
		}
		assertTreeUnchanged(t, root, before, output)
	}
}
