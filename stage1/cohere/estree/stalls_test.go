package estree

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func refusedBeforeDeadline(t *testing.T, argv []string, diagnostic string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	output, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stdout = output
	cmd.Stderr = &stderr
	err = cmd.Run()
	output.Close()
	info, _ := os.Stat(output.Name())
	if ctx.Err() != nil || err == nil || info.Size() != 0 || !strings.Contains(stderr.String(), diagnostic) {
		t.Fatalf("%v: timeout=%v exit=%v stdout=%d stderr=%s", argv, ctx.Err(), err, info.Size(), &stderr)
	}
}
func TestBoundedPortParser(t *testing.T) {
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for _, text := range []string{"type X = {", "interface I {", "type X = { m(a: string): void;"} {
		path := filepath.Join(t.TempDir(), "input.ts")
		os.WriteFile(path, []byte(text), 0644)
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
			refusedBeforeDeadline(t, argv, "ESTree parser stopped advancing")
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
		for _, argv := range [][]string{{"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), main, path}, {binary, path}, {"node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), script, path}} {
			refusedBeforeDeadline(t, argv, "ESTree parser stopped advancing")
		}
	}
	t.Log("all 13 recorded stalls explicitly refuse before 2s in all three port builds")
	t.Log("three EOF recovery stalls explicitly refuse before 2s on Node, sanitized native and emitted JS")
}
func TestPortStallControl(t *testing.T) {
	main := mutantPort(t, "sourceParser.ts", "this.stalledScans > 32", "this.stalledScans > 320000000")
	binary, _ := build(t, main, true)
	path := filepath.Join(t.TempDir(), "input.ts")
	os.WriteFile(path, []byte("type X = {"), 0644)
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
