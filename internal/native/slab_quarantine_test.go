package native

import (
	"bytes"
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
func TestSlabQuarantineCatchesAStaleReadOfAReusedSlot(t *testing.T) {
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
	started := time.Now()
	compiler, library := runtimeWithFlags(t, slabs)
	t.Logf("shared slab runtime archive: %s", time.Since(started))
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
			archive = slabQuarantineOffLibrary(t, compiler, library, build.flags)
		}
		binary := buildHarnessWithLibrary(t, harness, append(append([]string{}, build.flags...), "-I", filepath.Dir(library)), compiler, archive)
		t.Logf("%s build: %s", build.name, time.Since(started))
		started = time.Now()
		command := exec.Command(binary)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		t.Logf("%s process: %s", build.name, time.Since(started))
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
	compiler, library := runtimeWithFlags(t, flags)
	return buildHarnessWithLibrary(t, source, flags, compiler, library)
}

func runtimeWithFlags(t *testing.T, flags []string) (string, string) {
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
	return compiler, library
}

// The off define affects only the units including slab_quarantine.h. Reuse every
// other object from the exact same sanitized slab archive, including its headers.
func slabQuarantineOffLibrary(t *testing.T, compiler, original string, flags []string) string {
	t.Helper()
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(original)
	directory := t.TempDir()
	var objects []string
	for _, file := range files {
		if !strings.HasSuffix(file.name, ".c") {
			continue
		}
		object := filepath.Join(root, strings.TrimSuffix(file.name, ".c")+".o")
		if strings.Contains(string(file.contents), `#include "slab_quarantine.h"`) {
			object = filepath.Join(directory, filepath.Base(object))
			arguments := append(append([]string{}, flags...), "-I", root, "-c", filepath.Join(root, file.name), "-o", object)
			if output, err := exec.Command(compiler, arguments...).CombinedOutput(); err != nil {
				t.Fatalf("quarantine mutant must compile: %v\n%s", err, output)
			}
		}
		objects = append(objects, object)
	}
	library := filepath.Join(directory, "runtime.a")
	if output, err := exec.Command(archiverName(compiler), append([]string{"rcs", library}, objects...)...).CombinedOutput(); err != nil {
		t.Fatalf("quarantine mutant archive: %v\n%s", err, output)
	}
	return library
}

func buildHarnessWithLibrary(t *testing.T, source string, flags []string, compiler, library string) string {
	t.Helper()
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
