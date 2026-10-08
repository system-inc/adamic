package native

import (
	"fmt"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"
)

// Not parallel: resets process-wide peak counters and cache environment.
// Eight independent cold builds compete for the same pool. Explicit worker counts
// keep this exercising contention even if the go-test default becomes sequential.
func TestClangPoolConcurrentBuilds(t *testing.T) {
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	directory := t.TempDir()
	library, err := RuntimeLibrary("", Options{Sanitize: true, Split: true})
	if err != nil {
		t.Fatal(err)
	}
	clangPeak.Store(0)
	start := make(chan struct{})
	errors := make(chan error, 8)
	var builds sync.WaitGroup
	for build := 0; build < 8; build++ {
		builds.Add(1)
		go func(build int) {
			defer builds.Done()
			var source strings.Builder
			source.WriteString("#include \"adamic.h\"\n")
			for unit := 0; unit < goruntime.NumCPU()+1; unit++ {
				fmt.Fprintf(&source, "// adamic-module \"build%d-unit%d.a\"\nstatic int adamic_probe_%d(void) { return %d; }\n", build, unit, unit, unit)
			}
			source.WriteString("int main(void) { return ")
			for unit := 0; unit < goruntime.NumCPU()+1; unit++ {
				fmt.Fprintf(&source, "adamic_probe_%d()+", unit)
			}
			fmt.Fprintf(&source, "0-%d; }\n", goruntime.NumCPU()*(goruntime.NumCPU()+1)/2)
			<-start
			errors <- buildUnitsWithLibrary(source.String(), filepath.Join(directory, fmt.Sprintf("build%d", build)), Options{Sanitize: true, Jobs: goruntime.NumCPU()}, library, nil)
		}(build)
	}
	close(start)
	builds.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	peak := clangPeak.Load()
	t.Logf("eight builds: peak clangs=%d CPUs=%d", peak, goruntime.NumCPU())
	if peak < 1 || peak > int64(goruntime.NumCPU()) {
		t.Fatalf("peak clangs=%d, want 1..%d", peak, goruntime.NumCPU())
	}
	if active := clangActive.Load(); active != 0 {
		t.Fatalf("leaked clangs: %d", active)
	}
}

func TestExternalPrototypeStructure(t *testing.T) {
	source := "int unrelated_runtime_entry(int value);\nint main(void) { return unrelated_runtime_entry(1); }\n"
	header, _, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(header, "int unrelated_runtime_entry(int value);") {
		t.Fatal("external ABI declaration lost")
	}
	for _, source := range []string{"int unrelated_runtime_entry(int value) { return value; }", "int unrelated_runtime_entry = 1;", "int (*unrelated_runtime_pointer)(int);", "int (*unrelated_runtime_pointer)(int callback(int));", "inline static int unrelated_runtime_entry(int value);"} {
		if _, _, err := splitC(source); err == nil {
			t.Fatalf("accepted external definition: %s", source)
		}
	}
}
