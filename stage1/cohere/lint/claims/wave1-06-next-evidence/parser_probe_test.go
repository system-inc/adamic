package evidence

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestJSXBoundary(t *testing.T) {
	program, err := load.Load([]string{"parser-probe.a"})
	if err != nil {
		t.Fatal(err)
	}
	lowered, diagnostics := lower.Lower(context.Background(), program)
	if diagnostics != nil {
		t.Fatal(diagnostics)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "parser")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	emitted := filepath.Join(directory, "parser.mjs")
	if err := os.WriteFile(emitted, []byte(javascript.JavaScript(lowered)), 0600); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../../../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("parser-probe.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"control.ts.txt", "font.tsx.txt"} {
		path, err := filepath.Abs(file)
		if err != nil {
			t.Fatal(err)
		}
		var referenceOut, referenceErr []byte
		referenceCode := -1
		commands := []*exec.Cmd{
			exec.Command("node", "--disable-warning=ExperimentalWarning", runner, source, path),
			exec.Command("node", "--disable-warning=ExperimentalWarning", runner, emitted, path),
			exec.Command(binary, path),
		}
		for index, command := range commands {
			out, err := os.CreateTemp(directory, "stdout-")
			if err != nil {
				t.Fatal(err)
			}
			stderr, err := os.CreateTemp(directory, "stderr-")
			if err != nil {
				t.Fatal(err)
			}
			command.Stdout, command.Stderr = out, stderr
			runErr := command.Run()
			out.Close()
			stderr.Close()
			code := 0
			if runErr != nil {
				exit, ok := runErr.(*exec.ExitError)
				if !ok {
					t.Fatal(runErr)
				}
				code = exit.ExitCode()
			}
			stdoutBytes, err := os.ReadFile(out.Name())
			if err != nil {
				t.Fatal(err)
			}
			stderrBytes, err := os.ReadFile(stderr.Name())
			if err != nil {
				t.Fatal(err)
			}
			if index == 0 {
				referenceOut, referenceErr, referenceCode = stdoutBytes, stderrBytes, code
			} else if !bytes.Equal(stdoutBytes, referenceOut) || !bytes.Equal(stderrBytes, referenceErr) || code != referenceCode {
				t.Fatalf("%s backend %d differs: exit=%d stdout=%q stderr=%q", file, index, code, stdoutBytes, stderrBytes)
			}
			if file == "control.ts.txt" {
				if code != 0 || string(stdoutBytes) != "SourceFile\n" || len(stderrBytes) != 0 {
					t.Fatalf("control failed: %d %q %q", code, stdoutBytes, stderrBytes)
				}
			} else if code != 70 || !strings.Contains(string(stderrBytes), "parser slice expected GreaterThanToken, got Identifier") {
				t.Fatalf("JSX boundary changed: %d %q", code, stderrBytes)
			}
			t.Logf("%s backend=%d exit=%d stdout=%q stderr=%q", file, index, code, stdoutBytes, stderrBytes)
		}
	}
}
