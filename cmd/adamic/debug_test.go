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

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/native"
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

func debugDriverEnvironment(t *testing.T) []string {
	t.Helper()
	if os.Getenv("ADAMIC_BUILD_CACHE_DIR") == "" {
		t.Fatal("debug command tests require the unit's ADAMIC_BUILD_CACHE_DIR; refusing the home cache (#phzhe68)")
	}
	cache, err := location.BuildCache()
	if err != nil {
		t.Fatal(err)
	}
	output, err := location.Output()
	if err != nil {
		t.Fatal(err)
	}
	return append(os.Environ(), "ADAMIC_BUILD_DEBUG_DRIVER=1", "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_OUTPUT_DIR="+output)
}

// Re-exec the linked command, as the compiler-bug driver does. Its runtime
// supplier finds the archive built ahead; this unit never starts the Go compiler.
func TestBuildDebugBinary(t *testing.T) {
	t.Parallel()
	path := requestSource(t, "function f(n: number): number { return n + 1; }\nconsole.log((f(3)).toString());\n")
	binary := filepath.Join(t.TempDir(), "program")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestBuildDebugDriver$", "--", "build", path, "-o", binary, "--debug")
	command.Env = debugDriverEnvironment(t)
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

// Not parallel: the driver exits its own process through the real command.
func TestBuildDebugDriver(t *testing.T) {
	if os.Getenv("ADAMIC_BUILD_DEBUG_DRIVER") != "1" {
		return
	}
	debugDriverEnvironment(t)
	for index, argument := range os.Args {
		if argument == "--" {
			os.Exit(run(os.Args[index+1:]))
		}
	}
	t.Fatal("missing driver arguments")
}
