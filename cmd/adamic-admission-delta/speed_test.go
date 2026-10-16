package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestCompilerCacheIncludesSHA(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a, err := buildcache.Key(root, compilerInputs("base-sha"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := buildcache.Key(root, compilerInputs("head-sha"))
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("revision SHA missing from compiler cache key")
	}
	again, err := buildcache.Key(root, compilerInputs("base-sha"))
	if err != nil || a != again {
		t.Fatal("cache key is not stable", err)
	}
}

func TestParallelLoweringOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	compiler := filepath.Join(dir, "compiler")
	// Both inputs must start before either finishes. A serial worker times out.
	script := "#!/bin/sh\n[ \"$1\" = admission-lower ] || { echo 'panic: C emission reached' >&2; exit 1; }\n: > \"$2.started\"\nwhile [ ! -f one.started ] || [ ! -f two.started ]; do :; done\n"
	writeExecutable(t, compiler, script)
	programs := []entry{{program: program{Path: "one"}}, {program: program{Path: "two"}}}
	classifyPrograms(dir, compiler, compiler, programs, 2, time.Second)
	for _, p := range programs {
		if p.Class != "accepted-by-both" || p.Base.WallSeconds <= 0 {
			t.Fatalf("parallel lowering failed: %+v", p)
		}
	}
}

func TestExecutableForkIsolation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Resolve the host tool without naming a machine path in the test.
	truth, err := exec.LookPath("true")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	var forks sync.WaitGroup
	ready := make(chan struct{}, 8)
	failures := make(chan error, 8)
	for worker := 0; worker < 8; worker++ {
		forks.Add(1)
		go func() {
			defer forks.Done()
			ready <- struct{}{}
			for ctx.Err() == nil {
				if err := exec.CommandContext(ctx, truth).Run(); err != nil && ctx.Err() == nil {
					failures <- err
					return
				}
			}
		}()
	}
	defer func() { cancel(); forks.Wait() }()
	for worker := 0; worker < 8; worker++ {
		<-ready
	}
	for program := 0; program < 200; program++ {
		if err := ctx.Err(); err != nil {
			t.Fatal("fork stress deadline", err)
		}
		path := filepath.Join(dir, "compiler")
		writeExecutable(t, path, "#!/bin/sh\necho ready\n")
		o := execute(dir, time.Second, path)
		if o.Exit != 0 || o.Error != "" || o.Stdout != "ready\n" {
			t.Fatalf("executable %d under concurrent forks: %+v", program, o)
		}
	}
	cancel()
	forks.Wait()
	close(failures)
	for err := range failures {
		t.Errorf("fork pressure failed: %v", err)
	}
}

// Not parallel: compares the builder's policy with a runner's using t.Setenv.
func TestCompilerCacheIgnoresBuildPolicy(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GOTOOLCHAIN", "auto")
	t.Setenv("GOFLAGS", "-trimpath -buildvcs=false -p=4")
	built, err := buildcache.Key(root, compilerInputs("revision"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("GOFLAGS", "-buildvcs=false")
	read, err := buildcache.Key(root, compilerInputs("revision"))
	if err != nil || built != read {
		t.Fatalf("build and read policies key different compilers: %s != %s: %v", built, read, err)
	}
	t.Setenv("GOFLAGS", "-tags=other")
	changed, err := buildcache.Key(root, compilerInputs("revision"))
	if err != nil || changed == read {
		t.Fatal("output-changing build flags lost", err)
	}
}
