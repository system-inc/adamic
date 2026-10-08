package native

import (
	"os"
	"os/exec"
	"path/filepath"
	// goruntime is Go's runtime package: this package's own runtime is the embedded C.
	goruntime "runtime"
	"strings"
	"testing"
)

// fusedHarness computes what a fused multiply-add would get wrong. The operands come from volatile
// variables, so clang can't fold the arithmetic away before it chooses the instructions.
const fusedHarness = `#include <stdio.h>

int main(void) {
	volatile double tenth = 0.1, ten = 10, one = 1, third = 1.0 / 3;
	double a = tenth, b = ten, c = one, x = third;
	printf("%.17g %.17g %.17g\n", a * b - c, x * 3 - c, x * x - x * x);
	return 0;
}
`

// fusingProcessor is the -march that gives clang a fused multiply-add on this machine: none needed on
// arm64, where every processor has one (every Apple silicon Mac); Haswell, the first x86 with FMA, on
// an x86 that has it. It says why when there's no FMA to test against.
func fusingProcessor() (cpu string, why string, ok bool) {
	switch goruntime.GOARCH {
	case "arm64":
		return "", "", true
	case "amd64":
		switch goruntime.GOOS {
		case "linux":
			cpuinfo, err := os.ReadFile("/proc/cpuinfo")
			if err != nil {
				return "", "no /proc/cpuinfo to say whether this x86 has FMA: " + err.Error(), false
			}
			if !strings.Contains(string(cpuinfo), " fma ") {
				return "", "this x86 has no FMA (/proc/cpuinfo doesn't list fma)", false
			}
			return "haswell", "", true
		case "darwin":
			output, err := exec.Command("sysctl", "-n", "hw.optional.fma").Output()
			if err != nil || strings.TrimSpace(string(output)) != "1" {
				return "", "this Intel Mac has no FMA (sysctl hw.optional.fma isn't 1)", false
			}
			return "haswell", "", true
		}
	}
	return "", "no way known to tell whether this " + goruntime.GOOS + "/" + goruntime.GOARCH + " has FMA", false
}

// JavaScript rounds each multiply and each add on its own, so a native program must never fuse them,
// on a processor that could. This compiles for one that can and runs it here. First it proves the
// harness can see fusion on this machine at all, by building it with contraction allowed and getting
// the fused answers; a test that can't fail proves nothing, so that has to differ, or this fails.
func TestArithmeticIsNeverFused(t *testing.T) {
	t.Parallel()
	cpu, why, ok := fusingProcessor()
	if !ok {
		// Never on arm64, where it matters most: there FMA is always there and this always runs.
		// census: not-applicable Requires an FMA-capable arm64 or amd64 CPU; amd64 capability is read from /proc/cpuinfo on Linux or sysctl hw.optional.fma on macOS.
		t.Skip(why)
	}
	source := filepath.Join(t.TempDir(), "fused.c")
	if err := os.WriteFile(source, []byte(fusedHarness), 0o644); err != nil {
		t.Fatal(err)
	}
	control := filepath.Join(t.TempDir(), "contracted")
	arguments := []string{"-std=c11", "-O2", "-ffp-contract=fast", "-o", control, source}
	if cpu != "" {
		arguments = append(arguments, "-march="+cpu)
	}
	if combined, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		t.Fatalf("clang %v: %v\n%s", arguments, err, combined)
	}
	if got := runWithInput(t, "", control); got == "0 0 0\n" {
		t.Fatalf("with contraction allowed (-ffp-contract=fast, -march=%q) the harness still printed %q: it can't see fusion on this machine, so it proves nothing", cpu, got)
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "fused")
		if err := Build(fusedHarness, binary, Options{Sanitize: sanitize, cpu: cpu}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != "0 0 0\n" {
			t.Errorf("sanitize %v: got %q, want \"0 0 0\\n\", as JavaScript computes it", sanitize, got)
		}
	}
	if t.Failed() {
		return
	}
	t.Logf("%s/%s, -march=%q: contraction allowed fuses, and Build never does, with or without the sanitizers", goruntime.GOOS, goruntime.GOARCH, cpu)
}
