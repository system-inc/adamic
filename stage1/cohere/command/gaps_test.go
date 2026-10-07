package command

import (
	"bytes"
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostExitStatusGap(t *testing.T) {
	source := absolute(t, "gaps/exit_status.ts")
	node := onNode(t, source)
	if node.exitCode != 7 || len(node.stdout) > 0 || len(node.stderr) > 0 {
		t.Fatalf("Node process status: %+v", node)
	}
	program := lowered(t, source)
	nativeResult := nativelyRun(t, program)
	if nativeResult.exitCode != 70 || !strings.Contains(string(nativeResult.stderr), "Cannot set properties of undefined (setting 'exitCode')") {
		t.Fatalf("exit status gap changed: %+v", nativeResult)
	}
	t.Log("Node returns 7; unadapted native treats ambient process as undefined and panics. The unit's CLI adapter binds the typed exit result after cleanup.")
}
func TestNodeChildProcessModuleGap(t *testing.T) {
	source := absolute(t, "gaps/child.ts")
	node := onNode(t, source)
	if node.exitCode != 0 || string(node.stdout) != "0\n" || len(node.stderr) > 0 {
		t.Fatalf("Node child process: %+v", node)
	}
	program, err := load.Load([]string{source})
	if err == nil {
		_, err = lower.Lower(context.Background(), program)
	}
	if err == nil || !strings.Contains(err.Error(), "node:child_process") {
		t.Fatalf("child process gap changed: %v", err)
	}
	t.Logf("Node child succeeds; the general node:child_process module remains unsupported (runProcess is a separate native primitive): %v", err)
}
func TestExecutionGapNeverReturnsAGreenRun(t *testing.T) {
	directory := t.TempDir()
	write(t, filepath.Join(directory, "tsconfig.json"), `{"compilerOptions":{"strict":true},"include":["*.ts"]}`, 0644)
	write(t, filepath.Join(directory, "index.ts"), "export const value = 1;\n", 0644)
	goBinary := filepath.Join(t.TempDir(), "cohere")
	build := bounded(t, "go", "build", "-o", goBinary, "./command/cohere")
	build.Dir = absolute(t, filepath.Join(repository, "cohere"))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("Go build: %v %s", err, out)
	}
	arguments := []string{"--types", "--no-cache", "--directory", directory}
	want := execute(t, nil, goBinary, arguments...)
	if want.exitCode != 0 {
		t.Fatalf("Go control must succeed: %d %s %s", want.exitCode, want.stdout, want.stderr)
	}
	source := absolute(t, "main.ts")
	program := lowered(t, source)
	binary := filepath.Join(t.TempDir(), "port")
	// Build through the same host binding as the byte-for-byte CLI test.
	buildCLI(t, program, binary)
	got := execute(t, nil, binary, append([]string{goBinary}, arguments...)...)
	if got.exitCode != 1 || len(got.stdout) > 0 || !bytes.Equal(got.stderr, []byte("cohere: stage 1 command execution is not yet ported\n")) {
		t.Fatalf("execution gap no longer explicit: %+v", got)
	}
	t.Logf("Go exits 0 on the control project; port explicitly exits 1. Go stdout: %q; stderr: %q", want.stdout, want.stderr)
}
