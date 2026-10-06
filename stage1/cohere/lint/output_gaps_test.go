package lint

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type outputObservation struct {
	stdout, stderr []byte
	code           int
}

// Files keep output intact even when the program exits nonzero or a sanitizer stops it.
func observeOutput(t *testing.T, name string, arguments ...string) outputObservation {
	t.Helper()
	directory := t.TempDir()
	stdout, err := os.Create(filepath.Join(directory, "stdout.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	stderr, err := os.Create(filepath.Join(directory, "stderr.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer stderr.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, arguments...)
	command.Stdout, command.Stderr = stdout, stderr
	err = command.Run()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	out, err := os.ReadFile(stdout.Name())
	if err != nil {
		t.Fatal(err)
	}
	diagnostic, err := os.ReadFile(stderr.Name())
	if err != nil {
		t.Fatal(err)
	}
	return outputObservation{out, diagnostic, code}
}

func outputNode(t *testing.T, path string) outputObservation {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return observeOutput(t, "node", "--disable-warning=ExperimentalWarning", runner, path)
}

func outputBinary(t *testing.T, path string) string {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "probe")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

// Not parallel: NO_COLOR presence, including an empty value, is the probe's input.
func TestOutputRuntimeGaps(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	for _, probe := range []struct {
		name, file, answer string
		code               int
	}{
		{"exit status", "6_output_exit_status.ts", "one finding\n", 1},
		{"environment", "7_output_environment.ts", "color disabled\n", 0},
		{"clock", "8_output_clock.ts", "clock available\n", 0},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join("gaps", probe.file))
			if err != nil {
				t.Fatal(err)
			}
			answer := outputNode(t, path)
			if string(answer.stdout) != probe.answer || answer.code != probe.code || len(answer.stderr) != 0 {
				t.Fatalf("Node: stdout %q, stderr %q, exit %d", answer.stdout, answer.stderr, answer.code)
			}
			t.Logf("Node: stdout %q, exit %d", answer.stdout, answer.code)
			_, err = load.Load([]string{path})
			switch probe.name {
			case "clock":
				var check *load.CheckError
				if !errors.As(err, &check) || !strings.Contains(err.Error(), "Cannot find name 'performance'") {
					t.Fatalf("clock gap changed: %v", err)
				}
				t.Logf("stage 0: %v", err)
			case "exit status", "environment":
				if err != nil {
					t.Fatal(err)
				}
				got := observeOutput(t, outputBinary(t, path))
				if got.code != answer.code || !bytes.Equal(got.stdout, answer.stdout) || !bytes.Equal(got.stderr, answer.stderr) {
					t.Fatalf("closed process gap disagrees: Node %+v, native %+v", answer, got)
				}
				t.Logf("closed: native stdout %q, exit %d", got.stdout, got.code)
			}
		})
	}
}

// Not parallel: the empty NO_COLOR value is inherited by Node and the sanitized native probes.
// These are proposed workarounds, not renderer mutants: each removes one unavailable input.
func TestOutputWorkaroundMutants(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	for _, change := range []struct{ name, file, from, to string }{
		{"exit discarded", "6_output_exit_status.ts", "process.exit(1);", ""},
		{"environment assumed absent", "7_output_environment.ts", "process.env.NO_COLOR === undefined", "true"},
		{"clock replaced with zero", "8_output_clock.ts", "performance.now()", "0"},
	} {
		t.Run(change.name, func(t *testing.T) {
			original, err := filepath.Abs(filepath.Join("gaps", change.file))
			if err != nil {
				t.Fatal(err)
			}
			want := outputNode(t, original)
			source, err := os.ReadFile(original)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(source), change.from) != 1 {
				t.Fatal("mutant anchor changed")
			}
			path := filepath.Join(t.TempDir(), "mutant.ts")
			if err := os.WriteFile(path, []byte(strings.Replace(string(source), change.from, change.to, 1)), 0644); err != nil {
				t.Fatal(err)
			}
			node := outputNode(t, path)
			binary := outputBinary(t, path)
			got := observeOutput(t, binary)
			if node.code != 0 || got.code != 0 || len(node.stderr) != 0 || len(got.stderr) != 0 || !bytes.Equal(node.stdout, got.stdout) {
				t.Fatalf("mutant must finish cleanly and agree on Node/native: Node %+v, native %+v", node, got)
			}
			if want.code == got.code && bytes.Equal(want.stdout, got.stdout) && bytes.Equal(want.stderr, got.stderr) {
				t.Fatal("workaround mutant survived the original Node observation")
			}
			t.Logf("caught: native stdout %q, exit %d; original Node stdout %q, exit %d", got.stdout, got.code, want.stdout, want.code)
		})
	}
}
