package fuzz

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestFuzzerSharesRuntimeLibrary(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"library", "program"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if name == "library" {
				library, err := native.RuntimeLibrary("", native.Options{Sanitize: true})
				if err != nil {
					t.Fatal(err)
				}
				if checkout.runtime != library {
					t.Fatalf("fuzzer uses %q, Build uses %q", checkout.runtime, library)
				}
				return
			}
			outcome := checkout.Try("console.log('shared runtime');\n", filepath.Join(directory, "program"))
			if outcome.Verdict != Agreed {
				t.Fatalf("%s: %s\n%s", outcome.Verdict, outcome.Key, outcome.Detail)
			}
			if string(outcome.Native.Stdout) != "shared runtime\n" || outcome.Native.ExitCode != 0 {
				t.Fatalf("native: %#v", outcome.Native)
			}
		})
	}
}
