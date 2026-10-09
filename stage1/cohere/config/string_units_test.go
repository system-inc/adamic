package config

import (
	"bytes"
	"context"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// Not parallel: the two-second native execution budget needs an uncontended case.
// Profile recipe: testdata/string_units_slowdown.md. The fixture itself measures nothing.
func TestHouseConfigStringUnits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CohereSettings.json")
	directory := filepath.Dir(portDirectory(t, "", "", ""))
	source := filepath.Join(directory, "house_config_string_units.a")
	fixture, err := os.ReadFile("testdata/house_config_string_units.a")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, fixture, 0644); err != nil {
		t.Fatal(err)
	}
	reference := onNode(t, source, path)
	if reference.exitCode != 0 || len(reference.stderr) != 0 {
		t.Fatalf("Node: exit %d, stderr %s", reference.exitCode, reference.stderr)
	}
	program := lowered(t, source)
	backend := onJavaScriptBackend(t, program, path)
	if backend.exitCode != 0 || len(backend.stderr) != 0 {
		t.Fatalf("JavaScript backend: exit %d, stderr %s", backend.exitCode, backend.stderr)
	}
	if !bytes.Equal(backend.stdout, reference.stdout) {
		t.Fatal("JavaScript backend stdout differs from Node")
	}
	t.Logf("JavaScript backend agrees with Node (%d output bytes)", len(reference.stdout))

	// The deadline covers execution, not checking, compilation or the reference runs.
	directory = t.TempDir()
	if destination := os.Getenv("ADAMIC_STRING_UNITS_ARTIFACTS"); destination != "" {
		directory = destination
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	generated := native.C(program)
	if err := os.WriteFile(filepath.Join(directory, "house.c"), []byte(generated), 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "house")
	if err := native.Build(generated, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, path)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		return syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	}
	command.WaitDelay = 100 * time.Millisecond
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	started := time.Now()
	err = command.Run()
	elapsed := time.Since(started)
	if ctx.Err() != nil {
		t.Fatalf("native house loader exceeded 2s (%s); see testdata/string_units_slowdown.md; stderr: %s", elapsed, stderr.Bytes())
	}
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("native: %v, stderr %s", err, stderr.Bytes())
	}
	if !bytes.Equal(stdout.Bytes(), reference.stdout) {
		t.Fatal("native stdout differs from Node")
	}
	t.Logf("native agrees with Node in %s", elapsed)
}
