package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A rule's crash is cohere's bug, and the run that meets one fails, names it in every view, and finishes
// everything else (#v1ah2qq). TanStack/query's run met six react-hooks crashes and ended with a footer
// marker as the only trace in the default view, and with no findings beside them it would have ended green.
//
// The rule that crashes is planted: planted_panic_rule.go, compiled into this test's binary alone by the
// cohere_planted_panic tag. Two projects. In the first a real rule, no-empty, runs beside the planted one,
// and must still report on both files, the crashed file included. In the second the planted rule runs
// alone, so nothing but the crash can fail the run.
func TestARuleCrashFailsTheRunAndNamesItselfAsCohere(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(t.TempDir(), "cohere")
	build := exec.Command("go", "build", "-buildvcs=false", "-tags", "cohere_planted_panic", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("cannot build cohere with the planted rule: %v\n%s", err, output)
	}

	// The settings name the planted rule on top of the house's, so a body other rules have nothing to say
	// about is the only way to leave the crash as the run's one problem.
	project := func(rules string, body string) string {
		t.Helper()
		root := t.TempDir()
		writeTree(t, root, map[string]string{
			"tsconfig.json":       fixScopeTsconfig,
			"CohereSettings.json": `{"rules":{` + rules + `}}`,
			"crashes.ts":          "// cohere-test: panic here\n" + body,
			"survives.ts":         body,
		})
		return root
	}
	run := func(root string, arguments ...string) (string, int) {
		t.Helper()
		command := exec.Command(binary, append([]string{"--no-format", "--no-cache"}, arguments...)...)
		command.Dir = root
		output, err := command.CombinedOutput()
		var exitError *exec.ExitError
		switch {
		case err == nil:
			return string(output), 0
		case errors.As(err, &exitError):
			return string(output), childExitCode(t, exitError)
		}
		t.Fatalf("running cohere: %v\n%s", err, output)
		return "", 0
	}
	const planted = "cohere-test/panics-on-marked-file"

	beside := project(`"`+planted+`":"error","no-empty":"error"`, "export function check(flag: boolean): void {\n\tif (flag) {}\n}\n")
	output, code := run(beside)
	if code != 1 {
		t.Errorf("exit %d with a crash beside findings, want 1:\n%s", code, output)
	}
	// The crashed file's empty block is a line lower, under the marker comment.
	for _, location := range []string{"crashes.ts:3:12", "survives.ts:2:12"} {
		if !strings.Contains(output, location+" error no-empty") {
			t.Errorf("the rule beside the crash did not report at %s, so the crash cost more than its own verdict:\n%s", location, output)
		}
	}

	alone := project(`"`+planted+`":"error"`, "export function check(flag: boolean): boolean {\n\treturn !flag;\n}\n")
	output, code = run(alone)
	if code != 1 {
		t.Errorf("exit %d from a run whose only problem is a crash, want 1, since a missing verdict is not a pass:\n%s", code, output)
	}
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	crashLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, " crash "+planted+" ") {
			crashLines = append(crashLines, line)
		}
	}
	if len(crashLines) != 1 {
		t.Fatalf("the default view printed %d crash lines for one crash, want 1:\n%s", len(crashLines), output)
	}
	for _, want := range []string{filepath.Join(alone, "crashes.ts"), "This is a bug in cohere, not in your code", reportBugsAt,
		"Unhandled case in Node.Text"} {
		if !strings.Contains(crashLines[0], want) {
			t.Errorf("the crash line lacks %q:\n%s", want, crashLines[0])
		}
	}
	if footer := lines[len(lines)-1]; !strings.HasPrefix(footer, "✗ ☠️ ") || !strings.Contains(footer, "1 rule crash") {
		t.Errorf("the footer of a run that lost a verdict to a crash is %q, want a failing verdict that names the crash", footer)
	}

	output, code = run(alone, "--json")
	if code != 1 {
		t.Errorf("exit %d under --json, want 1", code)
	}
	var crashes, summaries int
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		var record struct {
			Kind    string `json:"kind"`
			Path    string `json:"path"`
			Rule    string `json:"rule"`
			Cause   string `json:"cause"`
			Verdict string `json:"verdict"`
			Gaps    struct {
				RuleCrashes int `json:"ruleCrashes"`
			} `json:"gaps"`
		}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("a --json line is not JSON: %v\n%s", err, line)
		}
		switch record.Kind {
		case "crash":
			crashes++
			if record.Rule != planted || filepath.Base(record.Path) != "crashes.ts" || !strings.Contains(record.Cause, "Unhandled case") {
				t.Errorf("the crash record names %s on %s with cause %q", record.Rule, record.Path, record.Cause)
			}
		case "summary":
			summaries++
			if record.Verdict != "fail" || record.Gaps.RuleCrashes != 1 {
				t.Errorf("the summary says %s with %d rule crashes, want fail and 1", record.Verdict, record.Gaps.RuleCrashes)
			}
		}
	}
	if crashes != 1 || summaries != 1 {
		t.Errorf("--json printed %d crash records and %d summaries, want 1 of each:\n%s", crashes, summaries, output)
	}
}
