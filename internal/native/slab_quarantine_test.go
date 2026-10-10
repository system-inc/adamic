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
func TestQuarantineCatchesAStaleReadOfAReusedSlot(t *testing.T) {
	t.Parallel()
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
	for _, build := range []struct {
		name  string
		flags []string
	}{
		{"without quarantine", append(append([]string{}, slabs...), "-DADAMIC_SLAB_QUARANTINE_OFF")},
		{"with quarantine", slabs},
	} {
		binary := buildWithFlags(t, harness, build.flags)
		command := exec.Command(binary)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		report := firstSanitizerLine(stderr.String())
		if build.name == "without quarantine" {
			if err != nil || stdout.String() != "slot reused\nstale length 40\n" || report != "" {
				t.Fatalf("%s: want the slot reused and the stale read unreported, got %v\nstdout %q\nstderr %.600s", build.name, err, stdout.String(), stderr.String())
			}
		} else if err == nil || stdout.String() != "slot quarantined\n" || !strings.Contains(report, "ERROR: AddressSanitizer: use-after-poison") {
			t.Fatalf("%s: want another slot and the stale read reported as a use after poison, got %v\nstdout %q\nstderr %.600s", build.name, err, stdout.String(), stderr.String())
		}
		t.Logf("%s: exit %v, stdout %q, first sanitizer line %q", build.name, err, stdout.String(), report)
	}
}

// buildWithFlags builds a harness against the embedded runtime compiled with exactly flags, which
// lets a test set a define Options has no field for.
func buildWithFlags(t *testing.T, source string, flags []string) string {
	t.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := exec.LookPath(compilerName(Options{}))
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(compiler, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("clang --version: %v\n%s", err, version)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	library, err := cachedRuntime(files, flags, compiler, string(version), filepath.Join(cache, "adamic", "runtime"))
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	main, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "harness")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	arguments := append(append([]string{}, flags...), "-I", filepath.Dir(library), "-o", binary, main)
	arguments = append(append(arguments, RuntimeLinkFlags(library)...), "-lm")
	if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("clang: %v\n%s", err, output)
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
func TestQuarantineIsPerThreadUnderTSan(t *testing.T) {
	t.Parallel()
	if !parallelTSan(t) {
		t.Skip("TSan is unavailable on this platform; see capability probe")
	}
	options := Options{ThreadSanitize: true, Slabs: true}
	t.Run("memory", func(t *testing.T) {
		t.Parallel()
		binary := parallelHarness(t, "memory.c", options)
		if stdout, _ := parallelRun(t, binary, "4"); stdout != "memory clean\n" {
			t.Fatalf("got %q", stdout)
		}
	})
	t.Run("map", func(t *testing.T) {
		t.Parallel()
		binary := parallelHarness(t, "map.c", options)
		one, _ := parallelRun(t, binary, "1", "strings")
		many, _ := parallelRun(t, binary, "4", "strings")
		if one != many {
			t.Fatalf("one worker %q, four workers %q", one, many)
		}
	})
	t.Run("shared_ring_mutant", func(t *testing.T) {
		t.Parallel()
		directory := t.TempDir()
		files, err := fs.ReadDir(runtime, "runtime")
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			contents, err := fs.ReadFile(runtime, "runtime/"+file.Name())
			if err != nil {
				t.Fatal(err)
			}
			if file.Name() == "slab_quarantine.c" {
				shared := strings.ReplaceAll(string(contents), "static _Thread_local", "static")
				if strings.Count(string(contents), "static _Thread_local") != 2 || strings.Contains(shared, "_Thread_local") {
					t.Fatal("mutant seam changed: want exactly the ring and its index thread-local")
				}
				contents = []byte(shared)
			}
			if err := os.WriteFile(filepath.Join(directory, file.Name()), contents, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		library, err := RuntimeLibrary(directory, options)
		if err != nil {
			t.Fatal(err)
		}
		source, err := os.ReadFile(filepath.Join("testdata", "parallel", "memory.c"))
		if err != nil {
			t.Fatal(err)
		}
		main, binary := filepath.Join(directory, "main.c"), filepath.Join(directory, "mutant")
		if err := os.WriteFile(main, []byte(parallelSource(string(source))), 0o600); err != nil {
			t.Fatal(err)
		}
		arguments := append(Flags(options), "-I", filepath.Dir(library), main, "-o", binary)
		arguments = append(append(arguments, RuntimeLinkFlags(library)...), "-lm")
		if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
			t.Fatalf("mutant must compile: %v\n%s", err, output)
		}
		for attempt := 1; attempt <= 3; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			command := exec.CommandContext(ctx, binary)
			command.Env = parallelEnvironment("4", false)
			output, err := command.CombinedOutput()
			cancel()
			if err == nil || !strings.Contains(string(output), "WARNING: ThreadSanitizer: data race") || !strings.Contains(string(output), "adamic_slab_quarantine") {
				t.Fatalf("attempt %d: the shared ring escaped TSan: %v\n%.2000s", attempt, err, output)
			}
			summary := "data race"
			for _, line := range strings.Split(string(output), "\n") {
				if strings.Contains(line, "SUMMARY: ThreadSanitizer:") {
					summary = strings.TrimSpace(line)
					break
				}
			}
			t.Logf("attempt %d: shared ring caught: %s", attempt, summary)
		}
	})
}
