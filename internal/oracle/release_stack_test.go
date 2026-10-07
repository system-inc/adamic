package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// Not parallel: the mutant intercepts clang through this test process's PATH.
// Only the generated-program compilation/link invocation loses the option;
// every runtime clang -c retains it, and neither warnings nor sanitizers catch it.
func TestReleaseRecursionKeepsFrames(t *testing.T) {
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		t.Fatal(err)
	}
	archiver := filepath.Join(filepath.Dir(resolved), "llvm-ar")
	_, err = os.Stat(archiver)
	if err != nil {
		archiver, err = exec.LookPath("ar")
	}
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	log := filepath.Join(directory, "commands.jsonl")
	wrapper := `#!/usr/bin/env python3
import json, os, sys
args=sys.argv[1:]
with open(LOG, 'a') as log:
 log.write(json.dumps(args)+'\n')
if '-c' not in args and '-flto=thin' in args:
 args=[a for a in args if a!='-fno-optimize-sibling-calls']
os.execv(COMPILER, [COMPILER, *args])
`
	encode := func(s string) string { b, _ := json.Marshal(s); return string(b) }
	wrapper = strings.ReplaceAll(strings.ReplaceAll(wrapper, "LOG", encode(log)), "COMPILER", encode(compiler))
	if err := os.WriteFile(filepath.Join(directory, "clang"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(archiver, filepath.Join(directory, "llvm-ar")); err != nil {
		t.Fatal(err)
	}
	originalPath := os.Getenv("PATH")
	for _, fixture := range []string{"stack_forever.a", "stack_tail_call.a"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 70 || !bytes.Equal(want.stderr, []byte("adamic: panic: RangeError: Maximum call stack size exceeded\n")) {
				t.Fatalf("Node did not give a stack panic: %+v", want)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			good := filepath.Join(t.TempDir(), "good")
			t.Setenv("PATH", originalPath)
			if err := native.Build(source, good, native.Options{Release: true}); err != nil {
				t.Fatal(err)
			}
			if got := execute(t, good); disagreement(want, got) != "" {
				t.Fatalf("shipping recursion differs: %+v vs %+v", want, got)
			}
			mutant := filepath.Join(t.TempDir(), "mutant")
			t.Setenv("PATH", directory+string(os.PathListSeparator)+originalPath)
			if err := native.Build(source, mutant, native.Options{Release: true}); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, mutant)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err = command.Run()
			if ctx.Err() != context.DeadlineExceeded {
				got := run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
				if disagreement(want, got) == "" {
					t.Fatal("link tail-call-option mutant survived")
				}
				t.Logf("link-only omission caught by Node disagreement: exit %d stdout %q stderr %q (%v)", got.exitCode, got.stdout, got.stderr, err)
			} else {
				t.Log("link-only omission caught: mutant loops past deadline where Node and shipped release panic")
			}
		})
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	runtimeCalls, linkCalls := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var args []string
		if err := json.Unmarshal([]byte(line), &args); err != nil {
			t.Fatal(err)
		}
		flags := strings.Join(args, " ")
		if strings.Contains(flags, " -c ") {
			runtimeCalls++
			if !strings.Contains(flags, "-fno-optimize-sibling-calls") {
				t.Fatal("mutant also changed runtime compilation")
			}
		}
		if strings.Contains(flags, "main.c") && !strings.Contains(flags, " -c ") {
			linkCalls++
			if !strings.Contains(flags, "-fno-optimize-sibling-calls") {
				t.Fatal("production link already omitted semantic option")
			}
		}
	}
	if runtimeCalls == 0 || linkCalls != 2 {
		t.Fatalf("missing audit: runtime=%d link=%d", runtimeCalls, linkCalls)
	}
}
