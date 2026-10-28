package main

import (
	"context"
	"debug/elf"
	"debug/macho"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

func TestBuildDebugFlag(t *testing.T) {
	t.Parallel()
	path := requestSource(t, "function f(n: number): number { return n + 1; }\nconsole.log((f(3)).toString());\n")
	for _, arguments := range [][]string{{"--debug"}, {"--sanitize"}, nil} {
		calls := 0
		code := buildWithNative(path, filepath.Join(t.TempDir(), "program"), arguments, func(source, output string, options native.Options) error {
			calls++
			debug := slices.Contains(arguments, "--debug")
			if options.Debug != debug || strings.Contains(source, "#line") != debug {
				t.Fatalf("%v: options=%+v markers=%t", arguments, options, strings.Contains(source, "#line"))
			}
			wantDirectory := ""
			if debug {
				entry, err := filepath.Abs(path)
				if err != nil {
					t.Fatal(err)
				}
				wantDirectory = filepath.Dir(entry)
			}
			if options.DebugDirectory != wantDirectory {
				t.Fatalf("debug directory=%q, want %q", options.DebugDirectory, wantDirectory)
			}
			if debug && !slices.Contains(native.Flags(options), "-g") {
				t.Fatal("--debug lost -g")
			}
			return nil
		})
		if code != 0 || calls != 1 {
			t.Fatalf("%v: code=%d builds=%d", arguments, code, calls)
		}
	}
}

// The subprocess is the shipped command, without the test binary's runtime supplier.
func TestBuildDebugBinary(t *testing.T) {
	t.Parallel()
	path := requestSource(t, "function f(n: number): number { return n + 1; }\nconsole.log((f(3)).toString());\n")
	binary := filepath.Join(t.TempDir(), "program")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	command := toolchain.CommandContextIn(ctx, ".", "go", "run", ".", "build", path, "-o", binary, "--debug")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build --debug: %v: %s", err, output)
	}
	if object, err := elf.Open(binary); err == nil {
		defer object.Close()
		if _, err := object.DWARF(); err != nil {
			t.Fatal(err)
		}
	} else {
		object, err := macho.Open(binary)
		if err != nil {
			t.Fatal(err)
		}
		defer object.Close()
		if _, err := object.DWARF(); err != nil {
			t.Fatal(err)
		}
	}
	output, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil || string(output) != "4\n" {
		t.Fatalf("run: %v: %q", err, output)
	}
}
