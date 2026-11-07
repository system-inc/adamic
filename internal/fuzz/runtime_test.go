package fuzz

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

func TestFuzzerSharesRuntimeLibrary(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	checkout, prepareBuiltError := PrepareBuilt("../..", buildcache.Adamic(test), filepath.Join(directory, "checkout"))
	if prepareBuiltError != nil {
		test.Fatal(prepareBuiltError)
	}
	for _, name := range []string{"library", "program"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			if name == "library" {
				library, runtimeLibraryError := native.RuntimeLibrary("", native.Options{Sanitize: true})
				if runtimeLibraryError != nil {
					test.Fatal(runtimeLibraryError)
				}
				checkoutLibrary, runtimeLibraryForSourceError := native.RuntimeLibraryForSource(
					filepath.Join(checkout.Root, "internal", "native", "runtime"),
					"",
					native.Options{Sanitize: true},
				)
				if runtimeLibraryForSourceError != nil {
					test.Fatal(runtimeLibraryForSourceError)
				}
				if checkoutLibrary != library {
					test.Fatalf("fuzzer uses %q, Build uses %q", checkoutLibrary, library)
				}
				return
			}
			outcome := checkout.Try("console.log('shared runtime');\n", filepath.Join(directory, "program"))
			if outcome.Verdict != Agreed {
				test.Fatalf("%s: %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
			}
			if string(outcome.Native.Stdout) != "shared runtime\n" || outcome.Native.ExitCode != 0 {
				test.Fatalf("native: %#v", outcome.Native)
			}
		})
	}
}
