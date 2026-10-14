package main

import (
	"os"
	"path/filepath"
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
	if err := os.WriteFile(compiler, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	programs := []entry{{program: program{Path: "one"}}, {program: program{Path: "two"}}}
	classifyPrograms(dir, compiler, compiler, programs, 2, time.Second)
	for _, p := range programs {
		if p.Class != "accepted-by-both" || p.Base.WallSeconds <= 0 {
			t.Fatalf("parallel lowering failed: %+v", p)
		}
	}
}
