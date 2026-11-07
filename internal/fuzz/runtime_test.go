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
	checkout, err := PrepareBuilt("../..", buildcache.Adamic(test), filepath.Join(directory, "checkout"))
	if err != nil {
		test.Fatal(err)
	}
	for _, name := range []string{"library", "program"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			if name == "library" {
				library, err := native.RuntimeLibrary("", native.Options{Sanitize: true})
				if err != nil {
					test.Fatal(err)
				}
				checkoutLibrary, err := native.RuntimeLibraryForSource(
					filepath.Join(checkout.Root, "internal", "native", "runtime"),
					"",
					native.Options{Sanitize: true},
				)
				if err != nil {
					test.Fatal(err)
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
