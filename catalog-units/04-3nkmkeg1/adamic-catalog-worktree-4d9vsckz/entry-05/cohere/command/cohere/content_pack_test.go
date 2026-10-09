package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// An installed declaration replaced the way a package manager replaces one, a new file renamed into place
// with the same size and the old modification time, reaches the next run through the content pack: a
// warm run reports what a cold one does (#j9d5nm6).
//
// `--lint` keeps the run cache out of it, since that run is never recorded or replayed, so the pack is
// the only thing between the disk and the compiler here. The rule is type-aware, so the declaration's
// bytes decide the finding.
func TestTheContentPackServesAReplacedDeclarationFresh(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("the content pack serves nothing on this platform")
	}
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// The same size both ways, so only the bytes differ.
	const withoutFinding = "export type Event = number;\n"
	const withFinding = "export type Event = 'abcd';\n"
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write("CohereSettings.json", `{"rules":{"@typescript-eslint/no-redundant-type-constituents":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	write("source/Use.ts", "import type { Event } from 'events-package';\nexport type EitherType = string | Event;\n")
	write("node_modules/events-package/package.json", `{"name":"events-package","types":"index.d.ts"}`)
	declaration := filepath.Join(root, "node_modules", "events-package", "index.d.ts")
	write("node_modules/events-package/index.d.ts", withoutFinding)

	finding := regexp.MustCompile(`(?m)^\S+:\d+:\d+ - .*$`)
	findings := func(output string) []string { return finding.FindAllString(output, -1) }
	run := func(arguments ...string) string {
		t.Helper()
		output, _ := runCohere(t, binary, root, append([]string{"--no-fix", "--lint"}, arguments...)...)
		return output
	}

	run()
	if _, err := os.Stat(filepath.Join(root, ".cache", "cohere", "contents.index")); err != nil {
		t.Fatalf("the first run wrote no content pack, so the runs below prove nothing: %v", err)
	}
	for _, step := range []struct {
		replacement string
		want        int
	}{{withFinding, 1}, {withoutFinding, 0}} {
		information, err := os.Stat(declaration)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
		staged := declaration + ".staged"
		if err := os.WriteFile(staged, []byte(step.replacement), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(staged, information.ModTime(), information.ModTime()); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(staged, declaration); err != nil {
			t.Fatal(err)
		}
		warm, cold := run(), run("--no-cache")
		if strings.Join(findings(warm), "\n") != strings.Join(findings(cold), "\n") || len(findings(cold)) != step.want {
			t.Fatalf("after the declaration became %q, warm reports %d findings and cold %d, want %d\n--- warm\n%s\n--- cold\n%s",
				strings.TrimSpace(step.replacement), len(findings(warm)), len(findings(cold)), step.want, warm, cold)
		}
	}
}

// A graph rebuilt after the fix phase reads the bytes the fixer wrote, not the ones the run cache's check saw
// (#kdee854). A missed run's pack validates against the check's stats rather than statting again, and the rebuild
// reads through that same pack, so a rewritten file whose pack entry and check stat both predate the rewrite was
// served to the rebuilt graph at its old bytes: its types and lint then reported text that no longer existed.
//
// Here the rewritten file is one nobody touched. Use.ts asserts a value from Value.ts `as string`, which is no
// finding while the value is a string or a number; Value.ts changes to a string or undefined, so Use.ts, unchanged
// since the last run, now has a non-null assertion spelled long, and the fixer rewrites it to `value!`.
func TestARebuildAfterAFixReadsTheFixedBytes(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("the content pack serves nothing on this platform")
	}
	binary := buildCohere(t)
	root := t.TempDir()
	write := func(name string, contents string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`)
	write("CohereSettings.json", `{"rules":{"@typescript-eslint/non-nullable-type-assertion-style":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	write("source/Value.ts", "export const value: string | number = Math.random() > 0.5 ? 'text' : 1;\n")
	write("source/Use.ts", "import { value } from './Value';\nexport const used = value as string;\n")
	runCohere(t, binary, root)
	runCohere(t, binary, root)
	if _, err := os.Stat(filepath.Join(root, ".cache", "cohere", "contents.index")); err != nil {
		t.Fatalf("the runs wrote no content pack, so the run below proves nothing: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	write("source/Value.ts", "export const value: string | undefined = Math.random() > 0.5 ? 'text' : undefined;\n")
	warm, _ := runCohere(t, binary, root)
	if !strings.Contains(warm, "graph rebuilt") {
		t.Fatalf("the fixer rewrote nothing, so the rebuild this test is about never ran:\n%s", warm)
	}
	fixed, err := os.ReadFile(filepath.Join(root, "source", "Use.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fixed), "value!") {
		t.Fatalf("the fixer did not rewrite Use.ts as this test expects:\n%s", fixed)
	}
	finding := regexp.MustCompile(`(?m)^\S+:\d+:\d+ - .*$`)
	cold, _ := runCohere(t, binary, root, "--no-fix", "--no-cache")
	if got, want := finding.FindAllString(warm, -1), finding.FindAllString(cold, -1); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("the run that fixed Use.ts reported other findings after its rebuild than a run with no cache over the fixed tree\n--- warm\n%s\n--- cold\n%s",
			warm, cold)
	}
}
