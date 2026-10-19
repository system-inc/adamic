package main

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestBuildSelectsShippedReleaseLTO(t *testing.T) {
	t.Parallel()
	auditReleaseBuild(t, "shipped", nil, true)
}

func TestBuildCountedExcludesReleaseLTO(t *testing.T) {
	t.Parallel()
	auditReleaseBuild(t, "counted", []string{"--count"}, false)
}

func TestBuildSanitizedExcludesReleaseLTO(t *testing.T) {
	t.Parallel()
	auditReleaseBuild(t, "sanitized", []string{"--sanitize"}, false)
}

func auditReleaseBuild(t *testing.T, name string, arguments []string, thin bool) {
	t.Helper()
	source := requestSource(t, "console.log('hello');\n")
	calls := 0
	compile := func(source, output string, options native.Options) error {
		calls++
		if !options.Release {
			t.Fatal("command did not select shipped release policy")
		}
		for _, flags := range [][]string{native.Flags(options), native.LinkFlags(options)} {
			if slices.Contains(flags, "-flto=thin") != thin {
				t.Fatalf("%s build has wrong LTO policy: %q", name, flags)
			}
			if !slices.Contains(flags, "-ffp-contract=off") || !slices.Contains(flags, "-fno-optimize-sibling-calls") {
				t.Fatalf("semantic options lost: %q", flags)
			}
		}
		if slices.Contains(native.Flags(options), "-fuse-ld=lld") {
			t.Fatal("linker option reached runtime compilation")
		}
		// Use the real builder and its declared archive products. A wrapper compiler
		// would change the installation prefix in the runtime's portable key.
		return native.Build(source, output, options)
	}
	if code := buildWithNative(source, filepath.Join(t.TempDir(), name), arguments, compile); code != 0 {
		t.Fatalf("%s build exit %d", name, code)
	}
	if calls != 1 {
		t.Fatalf("%s reached the native builder %d times, want once", name, calls)
	}
}
