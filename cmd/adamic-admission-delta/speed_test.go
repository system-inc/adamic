package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestCompilerCacheIncludesSha(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	a, keyError := buildcache.Key(root, compilerInputs("base-sha"))
	if keyError != nil {
		test.Fatal(keyError)
	}
	b, keyError := buildcache.Key(root, compilerInputs("head-sha"))
	if keyError != nil {
		test.Fatal(keyError)
	}
	if a == b {
		test.Fatal("revision SHA missing from compiler cache key")
	}
	again, keyError := buildcache.Key(root, compilerInputs("base-sha"))
	if keyError != nil || a != again {
		test.Fatal("cache key is not stable", keyError)
	}
}

func TestParallelLoweringOnly(test *testing.T) {
	test.Parallel()
	dir := test.TempDir()
	compiler := filepath.Join(dir, "compiler")
	// Both inputs must start before either finishes. A serial worker times out.
	script := "#!/bin/sh\n[ \"$1\" = admission-lower ] || { echo 'panic: C emission reached' >&2; exit 1; }\n: > \"$2.started\"\nwhile [ ! -f one.started ] || [ ! -f two.started ]; do :; done\n"
	writeExecutable(test, compiler, script)
	programs := []entry{{program: program{Path: "one"}}, {program: program{Path: "two"}}}
	classifyPrograms(dir, compiler, compiler, programs, 2, time.Second)
	for _, program := range programs {
		if program.Class != "accepted-by-both" || program.Base.WallSeconds <= 0 {
			test.Fatalf("parallel lowering failed: %+v", program)
		}
	}
}

func TestExecutableForkIsolation(test *testing.T) {
	test.Parallel()
	dir := test.TempDir()
	// The forks run this test binary with no test selected, so the pressure needs no host tool.
	self, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	var forks sync.WaitGroup
	ready := make(chan struct{}, 8)
	failures := make(chan error, 8)
	for worker := 0; worker < 8; worker++ {
		forks.Add(1)
		go func() {
			defer forks.Done()
			ready <- struct{}{}
			for timeoutContext.Err() == nil {
				if runError := exec.CommandContext(
					timeoutContext,
					self,
					"-test.run=^$",
				).Run(); runError != nil && timeoutContext.Err() == nil {
					failures <- runError
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
		if err := timeoutContext.Err(); err != nil {
			test.Fatal("fork stress deadline", err)
		}
		path := filepath.Join(dir, "compiler")
		writeExecutable(test, path, "#!/bin/sh\necho ready\n")
		observation := execute(dir, time.Second, path)
		if observation.Exit != 0 || observation.Error != "" || observation.Stdout != "ready\n" {
			test.Fatalf("executable %d under concurrent forks: %+v", program, observation)
		}
	}
	cancel()
	forks.Wait()
	close(failures)
	for err := range failures {
		test.Errorf("fork pressure failed: %v", err)
	}
}

// Not parallel: compares the builder's policy with a runner's using test.Setenv.
func TestCompilerCacheIgnoresBuildPolicy(test *testing.T) {
	root := test.TempDir()
	test.Setenv("GOTOOLCHAIN", "auto")
	test.Setenv("GOFLAGS", "-trimpath -buildvcs=false -p=4")
	built, keyError := buildcache.Key(root, compilerInputs("revision"))
	if keyError != nil {
		test.Fatal(keyError)
	}
	test.Setenv("GOTOOLCHAIN", "local")
	test.Setenv("GOFLAGS", "-buildvcs=false")
	read, keyError := buildcache.Key(root, compilerInputs("revision"))
	if keyError != nil || built != read {
		test.Fatalf("build and read policies key different compilers: %s != %s: %v", built, read, keyError)
	}
	test.Setenv("GOFLAGS", "-tags=other")
	changed, keyError := buildcache.Key(root, compilerInputs("revision"))
	if keyError != nil || changed == read {
		test.Fatal("output-changing build flags lost", keyError)
	}
}
