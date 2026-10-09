package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// A type the project imports from a nested repository changes, and the next warm run reports what a cold
// run reports: both caches notice an edit under a nested repository, in both directions.
//
// A nested repository is read and never written by the project's run, and its files are walked by its own
// run, so it is exactly where an input could fall out of what the caches hash: a run-cache input list or
// a fingerprint closure that skipped the nested repository's files would replay the importer's old verdict
// after the type it reads changed (#xhwjdyx). The rule here is type-aware and shape-keyed, so the edit
// reaches the importer's verdict only through the imported file's shape.
//
// The report that prompted this was a lint-only run, which reads no cache, catching BaseWorker.ts between
// two writes of one edit. It was the truth about that tree, reproduced cold. This holds the line it raised.
func TestANestedRepositoryEditReachesItsImportersWarm(t *testing.T) {
	t.Parallel()
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
	git := func(directory string, arguments ...string) {
		t.Helper()
		command := exec.Command("git", append([]string{"-c", "user.email=fixture@example.com", "-c", "user.name=fixture"}, arguments...)...)
		command.Dir = directory
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}

	// EventType is a literal in the finding state, which a string union makes redundant, and a number otherwise.
	const withFinding = "export type EventType = 'alarm';\n"
	const withoutFinding = "export type EventType = number;\n"
	write("tsconfig.json", `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source","libraries"]}`)
	write("CohereSettings.json", `{"rules":{"@typescript-eslint/no-redundant-type-constituents":"error"}}`)
	write("package.json", `{"name":"fixture","private":true,"type":"module"}`)
	write(".gitignore", ".cache/\nnode_modules/\n")
	write("source/Use.ts", "import type { EventType } from '../libraries/base/Event';\nexport type EitherType = string | EventType;\n")
	write("libraries/base/Event.ts", withoutFinding)
	write("libraries/base/package.json", `{"name":"base","private":true,"type":"module"}`)
	nested := filepath.Join(root, "libraries", "base")
	git(nested, "init", "--quiet")
	git(nested, "add", "-A")
	git(nested, "commit", "--quiet", "-m", "base")
	git(root, "init", "--quiet")
	git(root, "add", "-A")
	git(root, "commit", "--quiet", "-m", "project")

	finding := regexp.MustCompile(`(?m)^\S+:\d+:\d+ - .*$`)
	findings := func(output string) string { return strings.Join(finding.FindAllString(output, -1), "\n") }
	// Lint findings across a nested repository are the subject, so formatting is left out of every run: the
	// fixture's settings are written unformatted, and their would-change findings are not what it counts.
	run := func(arguments ...string) string {
		t.Helper()
		output, _ := runCohere(t, binary, root, append(arguments, "--no-format")...)
		return output
	}

	for _, step := range []struct {
		name   string
		before string
		after  string
		want   int
	}{
		{"a finding appears", withoutFinding, withFinding, 1},
		{"a finding goes away", withFinding, withoutFinding, 0},
	} {
		write("libraries/base/Event.ts", step.before)
		run("--no-fix")
		if replayed := run("--no-fix"); !strings.HasPrefix(replayed, "cached: ") {
			t.Fatalf("%s: an unchanged tree did not replay, so the edit below proves nothing:\n%s", step.name, replayed)
		}

		write("libraries/base/Event.ts", step.after)
		warm := run("--no-fix")
		cold := run("--no-fix", "--no-cache")
		if strings.HasPrefix(warm, "cached: ") {
			t.Fatalf("%s: a run after the nested repository's Event.ts changed replayed the run from before:\n%s", step.name, warm)
		}
		if findings(warm) != findings(cold) {
			t.Fatalf("%s: warm and cold disagree after the nested repository's Event.ts changed\n--- warm\n%s\n--- cold\n%s",
				step.name, warm, cold)
		}
		if got := len(finding.FindAllString(cold, -1)); got != step.want {
			t.Fatalf("%s: cold reports %d findings, want %d, so the fixture does not test what it says:\n%s", step.name, got, step.want, cold)
		}
	}
}
