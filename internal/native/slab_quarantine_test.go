package native

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A freed slab slot that the next value of its class takes again is unpoisoned by take, so a stale
// pointer into it reads the new value as a live one and ASan says nothing. The quarantine
// (runtime/slab_quarantine.c) holds freed slots poisoned before they can be taken again. The
// harness frees a string, makes another of the same class, and reads the freed one's length. With
// the quarantine compiled out the new string takes the old slot and the stale read goes unreported;
// with it in, the new string gets another slot and the stale read is a use after poison.
func TestSlabQuarantineCatchesAStaleReadOfAReusedSlot(test *testing.T) {
	test.Parallel()
	const harness = `#include "adamic.h"
#include <stdint.h>
#include <stdio.h>

int main(void) {
	adamic_string *old = adamic_string_allocate(40);
	uintptr_t address = (uintptr_t)old;
	adamic_release(old);
	adamic_string *taken = adamic_string_allocate(40);
	puts((uintptr_t)taken == address ? "slot reused" : "slot quarantined");
	fflush(stdout);
	printf("stale length %zu\n", old->length);
	adamic_release(taken);
	return 0;
}
`
	// The slab lane's build: sanitized, with the size classes kept on.
	slabs := append(Flags(Options{Sanitize: true}), "-DADAMIC_SLABS")
	started := time.Now()
	compiler, library := runtimeWithFlags(test, slabs)
	test.Logf("shared slab runtime archive: %s", time.Since(started))
	for _, build := range []struct {
		name  string
		flags []string
	}{
		{"without quarantine", append(append([]string{}, slabs...), "-DADAMIC_SLAB_QUARANTINE_OFF")},
		{"with quarantine", slabs},
	} {
		started := time.Now()
		archive := library
		if build.name == "without quarantine" {
			archive = slabQuarantineOffLibrary(test, compiler, library, build.flags)
		}
		binary := buildHarnessWithLibrary(
			test,
			harness,
			append(append([]string{}, build.flags...), "-I", filepath.Dir(library)),
			compiler,
			archive,
		)
		test.Logf("%s build: %s", build.name, time.Since(started))
		started = time.Now()
		command := exec.Command(binary)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		test.Logf("%s process: %s", build.name, time.Since(started))
		report := firstSanitizerLine(stderr.String())
		if build.name == "without quarantine" {
			if err != nil || stdout.String() != "slot reused\nstale length 40\n" || report != "" {
				test.Fatalf(
					"%s: want the slot reused and the stale read unreported, got %v\nstdout %q\nstderr %.600s",
					build.name,
					err,
					stdout.String(),
					stderr.String(),
				)
			}
		} else if err == nil ||
			stdout.String() != "slot quarantined\n" ||
			!strings.Contains(report, "ERROR: AddressSanitizer: use-after-poison") {
			test.Fatalf(
				"%s: want another slot and the stale read reported as a use after poison, got %v\nstdout %q\nstderr %.600s",
				build.name,
				err,
				stdout.String(),
				stderr.String(),
			)
		}
		test.Logf("%s: exit %v, stdout %q, first sanitizer line %q", build.name, err, stdout.String(), report)
	}
}

// buildWithFlags builds a harness against the embedded runtime compiled with exactly flags, which
// lets a test set a define Options has no field for.
func buildWithFlags(test *testing.T, source string, flags []string) string {
	test.Helper()
	compiler, library := runtimeWithFlags(test, flags)
	return buildHarnessWithLibrary(test, source, flags, compiler, library)
}

func runtimeWithFlags(test *testing.T, flags []string) (string, string) {
	test.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		test.Fatal(err)
	}
	compiler, err := exec.LookPath(compilerName(Options{}))
	if err != nil {
		test.Fatal(err)
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		test.Fatalf("clang --version: %v\n%s", err, version)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		test.Fatal(err)
	}
	library, err := cachedRuntime(files, flags, compiler, string(version), filepath.Join(cache, "adamic", "runtime"))
	if err != nil {
		test.Fatal(err)
	}
	return compiler, library
}

// The off define affects only the units including slab_quarantine.h. Reuse every
// other object from the exact same sanitized slab archive, including its headers.
func slabQuarantineOffLibrary(test *testing.T, compiler, original string, flags []string) string {
	test.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		test.Fatal(err)
	}
	root := filepath.Dir(original)
	directory := test.TempDir()
	var objects []string
	for _, file := range files {
		if !strings.HasSuffix(file.name, ".c") {
			continue
		}
		object := filepath.Join(root, strings.TrimSuffix(file.name, ".c")+".o")
		if strings.Contains(string(file.contents), `#include "slab_quarantine.h"`) {
			object = filepath.Join(directory, filepath.Base(object))
			arguments := append(
				append([]string{}, flags...),
				"-I",
				root,
				"-c",
				filepath.Join(root, file.name),
				"-o",
				object,
			)
			if output, err := productCompilerCommand(
				test,
				context.Background(),
				compiler,
				arguments...,
			).CombinedOutput(); err != nil {
				test.Fatalf("quarantine mutant must compile: %v\n%s", err, output)
			}
		}
		objects = append(objects, object)
	}
	library := filepath.Join(directory, "runtime.a")
	if output, err := exec.Command(
		archiverName(compiler),
		append([]string{"rcs", library}, objects...)...,
	).CombinedOutput(); err != nil {
		test.Fatalf("quarantine mutant archive: %v\n%s", err, output)
	}
	return library
}

func buildHarnessWithLibrary(test *testing.T, source string, flags []string, compiler, library string) string {
	test.Helper()
	directory := test.TempDir()
	main, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "harness")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		test.Fatal(err)
	}
	arguments := append(append([]string{}, flags...), "-I", filepath.Dir(library), "-o", binary, main)
	arguments = append(append(arguments, RuntimeLinkFlags(library)...), "-lm")
	if output, err := productCompilerCommand(
		test,
		context.Background(),
		compiler,
		arguments...,
	).CombinedOutput(); err != nil {
		test.Fatalf("clang: %v\n%s", err, output)
	}
	return binary
}

func firstSanitizerLine(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		if strings.Contains(line, "ERROR: AddressSanitizer") || strings.Contains(line, "runtime error:") {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// Pool threads free slab slots too, so the quarantine keeps a queue per thread. Under the thread
// sanitizer with ADAMIC_SLABS the quarantine runs (without poisoning, which TSan has no shadow for),
// and the parallel harnesses free on several threads at once, locally and to other threads'
// chunks: memory.c on four threads of its own, map.c through the pool. Neither may report a race.
// The mutant makes the queue one shared, unlocked ring, and TSan must report it.
func TestQuarantineIsPerThreadUnderTSan(test *testing.T) {
	test.Parallel()
	if !parallelTSan(test) {
		test.Skip("TSan is unavailable on this platform; see capability probe")
	}
	options := Options{ThreadSanitize: true, slabs: true}
	test.Run("memory", func(test *testing.T) {
		test.Parallel()
		binary := parallelHarness(test, "memory.c", options)
		if stdout, _ := parallelRun(test, binary, "4"); stdout != "memory clean\n" {
			test.Fatalf("got %q", stdout)
		}
	})
	test.Run("map", func(test *testing.T) {
		test.Parallel()
		binary := parallelHarness(test, "map.c", options)
		one, _ := parallelRun(test, binary, "1", "strings")
		many, _ := parallelRun(test, binary, "4", "strings")
		if one != many {
			test.Fatalf("one worker %q, four workers %q", one, many)
		}
	})
	test.Run("shared_ring_mutant", func(test *testing.T) {
		test.Parallel()
		directory := test.TempDir()
		files, err := fs.ReadDir(runtime, "runtime")
		if err != nil {
			test.Fatal(err)
		}
		for _, file := range files {
			contents, err := fs.ReadFile(runtime, "runtime/"+file.Name())
			if err != nil {
				test.Fatal(err)
			}
			if file.Name() == "slab_quarantine.c" {
				shared := strings.ReplaceAll(string(contents), "static _Thread_local", "static")
				if strings.Count(string(contents), "static _Thread_local") != 2 ||
					strings.Contains(shared, "_Thread_local") {
					test.Fatal("mutant seam changed: want exactly the ring and its index thread-local")
				}
				contents = []byte(shared)
			}
			if err := os.WriteFile(filepath.Join(directory, file.Name()), contents, 0o600); err != nil {
				test.Fatal(err)
			}
		}
		library, err := RuntimeLibrary(directory, options)
		if err != nil {
			test.Fatal(err)
		}
		source, err := os.ReadFile(filepath.Join("testdata", "parallel", "memory.c"))
		if err != nil {
			test.Fatal(err)
		}
		main, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "mutant")
		if err := os.WriteFile(main, []byte(parallelSource(string(source))), 0o600); err != nil {
			test.Fatal(err)
		}
		arguments := append(Flags(options), "-I", filepath.Dir(library), main, "-o", binary)
		arguments = append(append(arguments, RuntimeLinkFlags(library)...), "-lm")
		if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
			test.Fatalf("mutant must compile: %v\n%s", err, output)
		}
		for attempt := 1; attempt <= 3; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			command := exec.CommandContext(ctx, binary)
			command.Env = parallelEnvironment("4", false)
			output, err := command.CombinedOutput()
			cancel()
			if err == nil ||
				!strings.Contains(string(output), "WARNING: ThreadSanitizer: data race") ||
				!strings.Contains(string(output), "adamic_slab_quarantine") {
				test.Fatalf("attempt %d: the shared ring escaped TSan: %v\n%.2000s", attempt, err, output)
			}
			summary := "data race"
			for _, line := range strings.Split(string(output), "\n") {
				if strings.Contains(line, "SUMMARY: ThreadSanitizer:") {
					summary = strings.TrimSpace(line)
					break
				}
			}
			test.Logf("attempt %d: shared ring caught: %s", attempt, summary)
		}
	})
}
