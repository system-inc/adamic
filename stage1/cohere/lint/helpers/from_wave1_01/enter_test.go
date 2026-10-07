package enterhelper

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func run(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	log, err := os.CreateTemp(t.TempDir(), "command-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd.Stdout = log
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, stderr.String())
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func build(t *testing.T, entry string) (string, string) {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "enter")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(t.TempDir(), "enter.mjs")
	if err = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return binary, js
}

// Not parallel: consumer capture pins process environment and bounds sanitizer builds.
func TestEnterMatchesGo(t *testing.T) {
	root, err := filepath.Abs("../../../../..")
	if err != nil {
		t.Fatal(err)
	}
	capture := t.TempDir()
	script, _ := filepath.Abs("testdata/capture.py")
	run(t, root, "python3", script, capture)
	entry, _ := filepath.Abs("main.a")
	binary, js := build(t, entry)
	runner := filepath.Join(root, "oracle/node.mjs")
	sides := func(entry, binary, js, input string) [][]byte {
		return [][]byte{
			run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, input),
			run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, js, input),
			run(t, "", binary, input),
		}
	}
	consumers := []string{"array-callback-return", "consistent-return", "no-unreachable-loop", "react-hooks_rules-of-hooks", "controls"}
	var allInput, allWant []byte
	for _, consumer := range consumers {
		input := filepath.Join(capture, consumer+".input")
		want, err := os.ReadFile(filepath.Join(capture, consumer+".want"))
		if err != nil {
			t.Fatal(err)
		}
		rows, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		allInput = append(allInput, rows...)
		allWant = append(allWant, want...)
		for side, got := range sides(entry, binary, js, input) {
			if !bytes.Equal(got, want) {
				t.Fatalf("%s side %d disagrees with actual Go enter", consumer, side)
			}
		}
		t.Logf("%s: %d actual Go calls, %d bytes per backend", consumer, bytes.Count(rows, []byte{'\n'}), len(want))
	}
	input := filepath.Join(t.TempDir(), "all.input")
	if err := os.WriteFile(input, allInput, 0644); err != nil {
		t.Fatal(err)
	}
	helper, err := os.ReadFile("enter.a")
	if err != nil {
		t.Fatal(err)
	}
	driver, err := os.ReadFile("main.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct{ name, from, to string }{
		{"retain prior reachability", "node.reachable = node.incoming;", "node.reachable = node.reachable;"},
		{"retain prior current block", "state.current = node;", "state.current = state.current;"},
	} {
		directory := t.TempDir()
		mutant := strings.Replace(string(helper), change.from, change.to, 1)
		if mutant == string(helper) {
			t.Fatal("missing mutant anchor")
		}
		os.WriteFile(filepath.Join(directory, "enter.a"), []byte(mutant), 0644)
		entry := filepath.Join(directory, "main.a")
		os.WriteFile(entry, driver, 0644)
		binary, js := build(t, entry)
		for side, got := range sides(entry, binary, js, input) {
			if bytes.Equal(got, allWant) {
				t.Fatalf("%s survives side %d", change.name, side)
			}
			t.Logf("%s compiles and runs; comparison alone catches side %d", change.name, side)
		}
	}
}
