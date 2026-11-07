package main

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestBuildSelectsShippedReleaseLTO(test *testing.T) {
	test.Parallel()
	auditReleaseBuild(test, "shipped", nil, true)
}

func TestBuildCountedExcludesReleaseLTO(test *testing.T) {
	test.Parallel()
	auditReleaseBuild(test, "counted", []string{"--count"}, false)
}

func TestBuildSanitizedExcludesReleaseLTO(test *testing.T) {
	test.Parallel()
	auditReleaseBuild(test, "sanitized", []string{"--sanitize"}, false)
}

func auditReleaseBuild(test *testing.T, name string, arguments []string, thin bool) {
	test.Helper()
	source := requestSource(test, "console.log('hello');\n")
	calls := 0
	compile := func(source, output string, options native.Options) error {
		calls++
		if !options.Release {
			test.Fatal("command did not select shipped release policy")
		}
		for _, flags := range [][]string{native.Flags(options), native.LinkFlags(options)} {
			if slices.Contains(flags, "-flto=thin") != thin {
				test.Fatalf("%s build has wrong LTO policy: %q", name, flags)
			}
			if !slices.Contains(flags, "-ffp-contract=off") || !slices.Contains(flags, "-fno-optimize-sibling-calls") {
				test.Fatalf("semantic options lost: %q", flags)
			}
		}
		if slices.Contains(native.Flags(options), "-fuse-ld=lld") {
			test.Fatal("linker option reached runtime compilation")
		}
		// Use the real builder and its declared archive products. A wrapper compiler
		// would change the installation prefix in the runtime's portable key.
		return native.Build(source, output, options)
	}
	if code := buildWithNative(source, filepath.Join(test.TempDir(), name), arguments, compile); code != 0 {
		test.Fatalf("%s build exit %d", name, code)
	}
	if calls != 1 {
		test.Fatalf("%s reached the native builder %d times, want once", name, calls)
	}
}
