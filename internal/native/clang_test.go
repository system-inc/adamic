package native

import (
	"errors"
	"fmt"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// Not parallel: clears the process-wide jobserver override and observes the shared pool.
func TestClangPoolBoundsCompiles(t *testing.T) {
	t.Setenv(JobserverVariable, "")
	budget := max(1, goruntime.GOMAXPROCS(0))
	if cap(clangTokens) != budget {
		t.Fatalf("pool %d budget %d", cap(clangTokens), budget)
	}
	arrived := make(chan struct{}, budget+3)
	release := make(chan struct{})
	failures := make(chan error, budget+3)
	var group sync.WaitGroup
	for range budget + 3 {
		group.Go(func() { failures <- withCompilerToken(func() error { arrived <- struct{}{}; <-release; return nil }) })
	}
	closed := false
	defer func() {
		if !closed {
			close(release)
		}
		group.Wait()
	}()
	for range budget {
		select {
		case <-arrived:
		case <-time.After(2 * time.Second):
			t.Fatal("pool did not fill its budget")
		}
	}
	overflow := false
	select {
	case <-arrived:
		overflow = true
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	closed = true
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if overflow {
		t.Fatal("independent compiles exceeded the shared clang budget")
	}
	if clangActive.Load() != 0 {
		t.Fatal("leaked active clang slot")
	}
}

// Not parallel: fills and observes the process-wide pool's failure path.
func TestClangFailureReturnsBudget(t *testing.T) {
	t.Setenv(JobserverVariable, "")
	failed := errors.New("compile failed")
	for range cap(clangTokens) {
		if err := withCompilerToken(func() error { return failed }); !errors.Is(err, failed) {
			t.Fatalf("lost compile error: %v", err)
		}
	}
	done := make(chan error, 1)
	go func() { done <- withCompilerToken(func() error { return nil }) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("failed compiler kept its budget slot")
	}
	if clangActive.Load() != 0 {
		t.Fatal("failure leaked active clang slot")
	}
}

// Not parallel: changes the unit-cache environment and resets process-wide peak counters.
func TestClangPoolConcurrentBuilds(t *testing.T) {
	t.Setenv(JobserverVariable, "")
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	directory := t.TempDir()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	version, err := compilerVersion(compiler)
	if err != nil {
		t.Fatal(err)
	}
	library, err := cachedRuntime([]runtimeFile{{"runtime.c", []byte("int runtime_probe(void){return 1;}\n")}}, Flags(Options{Sanitize: true}), compiler, string(version), filepath.Join(directory, "runtime-cache"), 1)
	if err != nil {
		t.Fatal(err)
	}
	clangPeak.Store(0)
	var group sync.WaitGroup
	failures := make(chan error, 3)
	for build := range 3 {
		group.Go(func() {
			var source strings.Builder
			source.WriteString("#include <stdio.h>\n")
			for unit := range 3 {
				fmt.Fprintf(&source, "// adamic-module \"build%d-unit%d.a\"\nstatic int probe%d(void){return %d;}\n", build, unit, unit, unit+build)
			}
			source.WriteString("int main(void){printf(\"%d\\n\",probe0()+probe1()+probe2());return 0;}\n")
			binary := filepath.Join(directory, fmt.Sprintf("build%d", build))
			err := buildUnitsWithLibrary(source.String(), binary, Options{Sanitize: true, Jobs: 4}, library, "")
			if err == nil {
				got := runWithInput(t, "", binary)
				truth := runWithInput(t, "", "node", "-e", fmt.Sprintf("console.log(%d)", 3+3*build))
				if got != truth {
					err = fmt.Errorf("build %d: native %q Node %q", build, got, truth)
				}
			}
			failures <- err
		})
	}
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	peak := clangPeak.Load()
	t.Logf("three independent split builds: peak clang slots=%d budget=%d", peak, cap(clangTokens))
	if peak < 1 || peak > int64(cap(clangTokens)) || clangActive.Load() != 0 {
		t.Fatalf("bad pool peak=%d active=%d", peak, clangActive.Load())
	}
}
