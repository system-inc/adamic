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
func TestBuildSelectsOnlyShippedReleaseLTO(t *testing.T) {
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	log := filepath.Join(directory, "commands.jsonl")
	wrapper := fmt.Sprintf("#!/usr/bin/env python3\nimport json,os,sys\nargs=sys.argv[1:]\nwith open(%q,'a') as log:log.write(json.dumps(args)+'\\n')\nos.execv(%q,[%q,*args])\n", log, compiler, compiler)
	if err := os.WriteFile(filepath.Join(directory, "clang"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
	source := requestSource(t, "console.log('hello');\n")
	for _, row := range []struct {
		name      string
		arguments []string
		thin      bool
	}{
		{"shipped", nil, true}, {"counted", []string{"--count"}, false}, {"sanitized", []string{"--sanitize"}, false},
	} {
		if err := os.WriteFile(log, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if code := build(source, filepath.Join(directory, row.name), row.arguments); code != 0 {
			t.Fatalf("%s build exit %d", row.name, code)
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
			if slices.Contains(args, "--version") {
				continue
			}
			if slices.Contains(args, "-c") {
				compilations++
			} else {
				links++
			}
			if slices.Contains(args, "-flto=thin") != row.thin {
				t.Fatalf("%s command has wrong LTO policy: %q", row.name, args)
			}
			if !slices.Contains(args, "-ffp-contract=off") || !slices.Contains(args, "-fno-optimize-sibling-calls") {
				t.Fatalf("semantic options lost: %q", args)
			}
			if slices.Contains(args, "-c") && slices.Contains(args, "-fuse-ld=lld") {
				t.Fatal("linker option reached runtime compilation")
			}
		}
		if compilations == 0 || links != 1 {
			t.Fatalf("%s missing actual compiler audit: %d compile, %d link", row.name, compilations, links)
		}
	}
}
