package decisions

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func refused(t *testing.T, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	file, e := os.CreateTemp(t.TempDir(), "stdout-")
	if e != nil {
		t.Fatal(e)
	}
	defer file.Close()
	cmd.Stdout = file
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	e = cmd.Run()
	var code int
	if x, ok := e.(*exec.ExitError); ok {
		code = x.ExitCode()
	} else if e != nil {
		t.Fatal(e)
	}
	data, _ := os.ReadFile(file.Name())
	if code != 70 || len(data) != 0 || !strings.Contains(stderr.String(), "parser slice ") {
		t.Fatalf("gap observation changed: code=%d stdout=%q stderr=%q", code, data, stderr.String())
	}
	return strings.TrimSpace(stderr.String())
}

// Not parallel: the three parser graphs compile sequentially to bound compiler memory.
func TestSharedNexusParserGaps(t *testing.T) {
	root, _ := filepath.Abs("../../../../..")
	owned, _ := filepath.Abs(".")
	cohere := filepath.Join(root, "cohere")
	dir := t.TempDir()
	virtual := filepath.Join(cohere, "adamic_wave12_nexus_gaps.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(owned, "gaps/oracle.go.txt")}})
	op := filepath.Join(dir, "overlay.json")
	os.WriteFile(op, overlay, 0644)
	oracle := filepath.Join(dir, "oracle")
	run(t, cohere, "go", "build", "-overlay="+op, "-o", oracle, virtual)
	want := run(t, "", oracle, filepath.Join(owned, "gaps/cases.json"))
	t.Logf("actual Go reports on all three blocked fixtures: %s", want)
	for _, name := range []string{"internal-recovery", "project-recovery", "outside-top-level-await"} {
		entry := filepath.Join(owned, "gaps", name+".a")
		program, e := load.Load([]string{entry})
		if e != nil {
			t.Fatal(e)
		}
		lowered, e := lower.Lower(context.Background(), program)
		if e != nil {
			t.Fatal(e)
		}
		binary := filepath.Join(t.TempDir(), "probe")
		if e = native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); e != nil {
			t.Fatal(e)
		}
		module := filepath.Join(t.TempDir(), "probe.mjs")
		os.WriteFile(module, []byte(javascript.JavaScript(lowered)), 0644)
		runner := filepath.Join(root, "oracle/node.mjs")
		var firstFailure string
		for _, side := range []struct {
			name, command string
			args          []string
		}{{"source Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, entry}}, {"emitted JavaScript", "node", []string{"--disable-warning=ExperimentalWarning", runner, module}}, {"sanitized native", binary, nil}} {
			failure := refused(t, side.command, side.args...)
			if firstFailure == "" {
				firstFailure = failure
			} else if failure != firstFailure {
				t.Fatalf("parser refusal differs: %s vs %s", firstFailure, failure)
			}
			t.Logf("%s %s exits 70: %s", name, side.name, failure)
		}
	}
}
