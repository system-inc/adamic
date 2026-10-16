package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Not parallel: the command's clang recorder changes the process-wide PATH.
func TestBuildSelectsShippedReleaseLTO(t *testing.T) { auditReleaseBuild(t, "shipped", nil, true) }

// Not parallel: the command recorder changes PATH and ADAMIC_TOOLS.
func TestBuildCountedExcludesReleaseLTO(t *testing.T) {
	auditReleaseBuild(t, "counted", []string{"--count"}, false)
}

// Not parallel: the command recorder changes PATH and ADAMIC_TOOLS.
func TestBuildSanitizedExcludesReleaseLTO(t *testing.T) {
	auditReleaseBuild(t, "sanitized", []string{"--sanitize"}, false)
}

func auditReleaseBuild(t *testing.T, name string, arguments []string, thin bool) {
	t.Helper()
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	tools := t.TempDir()
	directory := filepath.Join(tools, "bin")
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(directory, "commands.jsonl")
	wrapper := fmt.Sprintf("#!/usr/bin/env python3\nimport json,os,sys\nargs=sys.argv[1:]\nwith open(%q,'a') as log:log.write(json.dumps(args)+'\\n')\nos.execv(%q,[%q,*args])\n", log, compiler, compiler)
	if err := os.WriteFile(filepath.Join(directory, "clang"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(compiler)
	if err != nil {
		t.Fatal(err)
	}
	archiver := filepath.Join(filepath.Dir(resolved), "llvm-ar")
	if _, err := os.Stat(archiver); err == nil {
		if err := os.Symlink(archiver, filepath.Join(directory, "llvm-ar")); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("ADAMIC_TOOLS", tools)
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	source := requestSource(t, "console.log('hello');\n")

	if err := os.WriteFile(log, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := build(source, filepath.Join(directory, name), arguments); code != 0 {
		t.Fatalf("%s build exit %d", name, code)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	compilations, links := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var args []string
		if err := json.Unmarshal([]byte(line), &args); err != nil {
			t.Fatal(err)
		}
		if slices.Contains(args, "--version") || slices.Contains(args, "-M") {
			continue
		}
		if slices.Contains(args, "-c") {
			compilations++
		} else {
			links++
		}
		if slices.Contains(args, "-flto=thin") != thin {
			t.Fatalf("%s command has wrong LTO policy: %q", name, args)
		}
		if !slices.Contains(args, "-ffp-contract=off") || !slices.Contains(args, "-fno-optimize-sibling-calls") {
			t.Fatalf("semantic options lost: %q", args)
		}
		if slices.Contains(args, "-c") && slices.Contains(args, "-fuse-ld=lld") {
			t.Fatal("linker option reached runtime compilation")
		}
	}
	if compilations == 0 || links != 1 {
		t.Fatalf("%s missing actual compiler audit: %d compile, %d link", name, compilations, links)
	}
}
