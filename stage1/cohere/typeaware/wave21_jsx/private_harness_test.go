// Private harness for the owned JSX rules. Shared harness files are unchanged.
package wave21jsx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type result struct {
	stdout, stderr []byte
	elapsed        time.Duration
	err            error
}
type harness struct {
	t                     *testing.T
	repository, directory string
	next                  int
}

func (h *harness) run(name string, command *exec.Cmd) result {
	h.t.Helper()
	h.next++
	if command.Dir == "" {
		command.Dir = h.repository
	}
	stem := filepath.Join(h.directory, fmt.Sprintf("%03d-%s", h.next, name))
	out, err := os.Create(stem + ".stdout")
	if err != nil {
		h.t.Fatal(err)
	}
	defer out.Close()
	report, err := os.Create(stem + ".stderr")
	if err != nil {
		h.t.Fatal(err)
	}
	defer report.Close()
	command.Stdout = out
	command.Stderr = report
	started := time.Now()
	runError := command.Run()
	elapsed := time.Since(started)
	stdout, err := os.ReadFile(out.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	stderr, err := os.ReadFile(report.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	return result{stdout, stderr, elapsed, runError}
}
func (h *harness) must(name string, command *exec.Cmd) result {
	h.t.Helper()
	r := h.run(name, command)
	if r.err != nil {
		h.t.Fatalf("%s: %v\n%s\n%s", name, r.err, r.stdout, r.stderr)
	}
	return r
}
func (h *harness) write(name, text string) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		h.t.Fatal(err)
	}
	return path
}
func (h *harness) overlay(name, path, from, to string) string {
	h.t.Helper()
	original := filepath.Join(h.repository, path)
	data, err := os.ReadFile(original)
	if err != nil {
		h.t.Fatal(err)
	}
	if strings.Count(string(data), from) != 1 {
		h.t.Fatalf("nonunique mutant %s", name)
	}
	side := h.write(name+filepath.Ext(path), strings.Replace(string(data), from, to, 1))
	data, err = json.Marshal(map[string]any{"Replace": map[string]string{original: side}})
	if err != nil {
		h.t.Fatal(err)
	}
	return h.write(name+".json", string(data))
}
func (h *harness) archive(name, overlay string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name+".a")
	args := []string{"build", "-buildmode=c-archive", "-o", path}
	if overlay != "" {
		args = append(args, "-overlay", overlay)
	}
	args = append(args, "./bridge/tsgo/archive")
	cmd := exec.Command("go", args...)
	if sanitize {
		cmd.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	}
	h.must(name, cmd)
	return path
}
func (h *harness) build(stage0, name, entry, archive string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	args := []string{"build", entry, "-o", path, "--tsgo", archive}
	if sanitize {
		args = append(args, "--sanitize")
	}
	h.must(name, exec.Command(stage0, args...))
	return path
}
func firstDifference(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return min(len(a), len(b))
}
func (h *harness) compare(name, oracle, binary, config, manifest string) result {
	h.t.Helper()
	want := h.must(name+"-go", exec.Command(oracle, config, manifest))
	got := h.must(name+"-native", exec.Command(binary, config, manifest))
	if len(got.stderr) != 0 {
		h.t.Fatalf("sanitizer stderr: %s", got.stderr)
	}
	if !bytes.Equal(got.stdout, want.stdout) {
		i := firstDifference(got.stdout, want.stdout)
		h.t.Fatalf("%s mismatch byte %d: native %q Go %q", name, i, got.stdout[max(0, i-50):min(len(got.stdout), i+250)], want.stdout[max(0, i-50):min(len(want.stdout), i+250)])
	}
	h.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}
func summary(data []byte) string {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	return lines[len(lines)-1]
}
