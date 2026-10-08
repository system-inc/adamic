package estree

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func refusedBeforeDeadline(t *testing.T, argv []string, diagnostic string) {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	err, timedOut := runWithCPUBudget(t, argv, output, &stderr, 2*time.Second)
	output.Close()
	info, _ := os.Stat(output.Name())
	matchesDiagnostic := strings.Contains(stderr.String(), diagnostic)
	if diagnostic == "ESTree parser" {
		// Parser diagnostics use the Go ESTree boundary spelling and byte position.
		matchesDiagnostic = matchesDiagnostic || regexp.MustCompile(`(?m)parse error at [0-9]+: [^\r\n]+`).MatchString(stderr.String())
	}
	if timedOut || err == nil || info.Size() != 0 || !matchesDiagnostic {
		t.Fatalf("%v: timeout=%v exit=%v stdout=%d stderr=%s", argv, timedOut, err, info.Size(), &stderr)
	}
}
func TestBoundedPortParser(t *testing.T) {
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for _, text := range []string{"type X = {", "interface I {", "type X = { m(a: string): void;"} {
		path := filepath.Join(t.TempDir(), "input.ts")
		os.WriteFile(path, []byte(text), 0644)
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
			refusedBeforeDeadline(t, argv, "ESTree parser")
		}
	}
	fixtures, err := filepath.Glob("validation/followup/stalls/*.input")
	if err != nil || len(fixtures) != 13 {
		t.Fatalf("stall fixtures: %d %v", len(fixtures), err)
	}
	for _, fixture := range fixtures {
		body, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), strings.TrimSuffix(filepath.Base(fixture), ".input"))
		os.WriteFile(path, body, 0644)
		list := filepath.Join(t.TempDir(), "manifest")
		os.WriteFile(list, []byte(path+"\n"), 0644)
		var record struct{ Status string }
		if err := json.Unmarshal(execute(t, "", goOracle(t), "--audit", list, t.TempDir()), &record); err != nil {
			t.Fatal(err)
		}
		if record.Status == "ok" {
			want := execute(t, "", goOracle(t), path)
			for name, got := range map[string][]byte{"Node": onNode(t, main, path), "native": execute(t, "", binary, path), "emitted": onNode(t, script, path)} {
				if diff := firstDifference(want, got); diff != "" {
					t.Fatal(name + ": " + diff)
				}
			}
		} else {
			for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
				refusedBeforeDeadline(t, argv, "ESTree parser")
			}
		}
	}
	t.Log("all 13 recorded stalls terminate: Go-accepted inputs match and Go-refused inputs explicitly refuse in all three port builds")
	t.Log("three EOF recovery cases explicitly refuse with parser diagnostics before 2s of child CPU time on Node, sanitized native and emitted JS")
}
func TestPortStallControl(t *testing.T) {
	main := mutantPort(t, "sourceStatements.ts", "if(this.parser.scanner.fullStart === start)", "if(false)")
	binary, _ := build(t, main, true)
	path := filepath.Join(t.TempDir(), "input.ts")
	os.WriteFile(path, []byte("class C { ) }"), 0644)
	for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}} {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
		output, _ := os.CreateTemp(t.TempDir(), "control-log")
		cmd.Stdout = output
		cmd.Stderr = output
		err := cmd.Run()
		output.Close()
		timedOut := ctx.Err() == context.DeadlineExceeded
		cancel()
		if !timedOut {
			t.Fatalf("stall control must hit deadline, got %v", err)
		}
		t.Logf("%s guard-disabled control caught by 500ms deadline", argv[0])
	}
}
