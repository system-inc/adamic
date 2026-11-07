package main

import (
	"context"
	"debug/elf"
	"debug/macho"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/native"
)

func TestBuildDebugFlag(test *testing.T) {
	test.Parallel()
	path := requestSource(test, "function f(n: number): number { return n + 1; }\nconsole.log((f(3)).toString());\n")
	for _, arguments := range [][]string{{"--debug"}, {"--sanitize"}, nil} {
		calls := 0
		code := buildWithNative(path, filepath.Join(
			test.TempDir(),
			"program",
		), arguments, func(source, output string, options native.Options) error {
			calls++
			debug := slices.Contains(arguments, "--debug")
			if options.Debug != debug || strings.Contains(source, "#line") != debug {
				test.Fatalf("%v: options=%+v markers=%t", arguments, options, strings.Contains(source, "#line"))
			}
			wantDirectory := ""
			if debug {
				entry, absolutePathError := filepath.Abs(path)
				if absolutePathError != nil {
					test.Fatal(absolutePathError)
				}
				wantDirectory = filepath.Dir(entry)
			}
			if options.DebugDirectory != wantDirectory {
				test.Fatalf("debug directory=%q, want %q", options.DebugDirectory, wantDirectory)
			}
			if debug && !slices.Contains(native.Flags(options), "-g") {
				test.Fatal("--debug lost -g")
			}
			return nil
		})
		if code != 0 || calls != 1 {
			test.Fatalf("%v: code=%d builds=%d", arguments, code, calls)
		}
	}
}

func debugDriverEnvironment(test *testing.T) []string {
	test.Helper()
	if os.Getenv("ADAMIC_BUILD_CACHE_DIR") == "" {
		test.Fatal("debug command tests require the unit's ADAMIC_BUILD_CACHE_DIR; refusing the home cache (#phzhe68)")
	}
	cache, buildCacheError := location.BuildCache()
	if buildCacheError != nil {
		test.Fatal(buildCacheError)
	}
	output, outputError := location.Output()
	if outputError != nil {
		test.Fatal(outputError)
	}
	return append(
		os.Environ(),
		"ADAMIC_BUILD_DEBUG_DRIVER=1",
		"ADAMIC_BUILD_CACHE_DIR="+cache,
		"ADAMIC_BUILD_OUTPUT_DIR="+output,
	)
}

// Re-exec the linked command, as the compiler-bug driver does. Its runtime
// supplier finds the archive built ahead; this unit never starts the Go compiler.
func TestBuildDebugBinary(test *testing.T) {
	test.Parallel()
	path := requestSource(test, "function f(n: number): number { return n + 1; }\nconsole.log((f(3)).toString());\n")
	binary := filepath.Join(test.TempDir(), "program")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(
		ctx,
		os.Args[0],
		"-test.run=^TestBuildDebugDriver$",
		"--",
		"build",
		path,
		"-o",
		binary,
		"--debug",
	)
	command.Env = debugDriverEnvironment(test)
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("build --debug: %v: %s", combinedOutputError, output)
	}
	if object, openError := elf.Open(binary); openError == nil {
		defer object.Close()
		if _, dwarfError := object.DWARF(); dwarfError != nil {
			test.Fatal(dwarfError)
		}
	} else {
		object, openError := macho.Open(binary)
		if openError != nil {
			test.Fatal(openError)
		}
		defer object.Close()
		if _, dwarfError := object.DWARF(); dwarfError != nil {
			test.Fatal(dwarfError)
		}
	}
	output, combinedOutputError := exec.CommandContext(ctx, binary).CombinedOutput()
	if combinedOutputError != nil || string(output) != "4\n" {
		test.Fatalf("run: %v: %q", combinedOutputError, output)
	}
}

// Not parallel: the driver exits its own process through the real command.
func TestBuildDebugDriver(test *testing.T) {
	if os.Getenv("ADAMIC_BUILD_DEBUG_DRIVER") != "1" {
		return
	}
	debugDriverEnvironment(test)
	for index, argument := range os.Args {
		if argument == "--" {
			os.Exit(run(os.Args[index+1:]))
		}
	}
	test.Fatal("missing driver arguments")
}

// The child removes its own cwd so filepath.Abs sees a machine failure, without
// changing the working directory of parallel tests in this process.
func TestBuildDebugDirectoryFailure(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TEST_DEBUG_DIRECTORY_FAILURE") == "1" {
		directory, temporaryDirectoryError := os.MkdirTemp("", "adamic-debug-cwd-")
		if temporaryDirectoryError != nil {
			test.Fatal(temporaryDirectoryError)
		}
		if chdirError := os.Chdir(directory); chdirError != nil {
			test.Fatal(chdirError)
		}
		if removeError := os.Remove(directory); removeError != nil {
			test.Fatal(removeError)
		}
		os.Exit(buildWithNative(
			"fixture.a",
			"program",
			[]string{"--debug"},
			func(string, string, native.Options) error { panic("compile reached after cwd failure") },
		))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestBuildDebugDirectoryFailure$")
	command.Env = append(command.Environ(), "ADAMIC_TEST_DEBUG_DIRECTORY_FAILURE=1")
	output, combinedOutputError := command.CombinedOutput()
	if command.ProcessState == nil ||
		command.ProcessState.ExitCode() != 3 ||
		!strings.Contains(string(output), "getwd") {
		test.Fatalf("unreadable cwd: error=%v output=%s, want exit 3 and getwd error", combinedOutputError, output)
	}
	test.Log("unreadable cwd reports getwd failure and exits 3")
}
