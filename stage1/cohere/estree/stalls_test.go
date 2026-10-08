package estree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"os"
	"os/exec"
	"path/filepath"
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
	if timedOut || err == nil || info.Size() != 0 || !strings.Contains(stderr.String(), diagnostic) {
		t.Fatalf("%v: timeout=%v exit=%v stdout=%d stderr=%s", argv, timedOut, err, info.Size(), &stderr)
	}
}
func TestBoundedPortParser(t *testing.T) {
	t.Parallel()
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for index, text := range []string{"type X = {", "interface I {", "type X = { m(a: string): void;"} {
		t.Run(fmt.Sprintf("EOF-%04d", index), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "input.ts")
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
			checkRefusalModes(t, main, binary, script, path, "ESTree parser")
		})
	}
	fixtures, err := filepath.Glob("validation/followup/stalls/*.input")
	if err != nil || len(fixtures) != 13 {
		t.Fatalf("stall fixtures: %d %v", len(fixtures), err)
	}
	for _, fixture := range fixtures {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
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
				checkPort(t, main, []string{path}, want, false, false)
			} else {
				checkRefusalModes(t, main, binary, script, path, "ESTree parser")
			}
		})
	}

	t.Log("all 13 recorded stalls terminate: Go-accepted inputs match and Go-refused inputs explicitly refuse in all three port builds")
	t.Log("three EOF recovery cases explicitly refuse with parser diagnostics before 2s of child CPU time on Node, sanitized native and emitted JS")
}
func TestPortStallControl(t *testing.T) {
	t.Parallel()
	main := mutantPort(t, "sourceStatements.ts", "if(this.parser.scanner.fullStart === start)", "if(false)")
	path := filepath.Join(t.TempDir(), "input.ts")
	if err := os.WriteFile(path, []byte("class C { ) }"), 0644); err != nil {
		t.Fatal(err)
	}
	commands := portCommands(t, main, []string{path}, true, false)
	for _, command := range commands {
		t.Run(command.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(command.argv[0], command.argv[1:]...)
			output, err := os.CreateTemp(t.TempDir(), "control-log")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			cmd.Stdout, cmd.Stderr = output, output
			err = childguard.Run(cmd, childguard.Options{FirstOutput: 500 * time.Millisecond, Stall: 500 * time.Millisecond, Ceiling: 500 * time.Millisecond})
			var guarded *childguard.Error
			if !errors.As(err, &guarded) {
				t.Fatalf("%s stall control must hit childguard deadline, got %v", command.name, err)
			}
			t.Logf("%s guard-disabled control caught by 500ms childguard deadline", command.name)
		})
	}
}
