package fuzz

import (
	"os"
	"path/filepath"
	"strings"
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

// Captured cells change size under ADAMIC_CANONICAL_CLOSURES. A runtime built
// without the program's feature switches corrupts this otherwise scalar state.
func TestFuzzerMatchesClosureRuntimeFeatures(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	source := "import { parallelMap } from 'adamic';\nconst items: readonly number[] = [1, 2];\n" +
		"function counter(): () => number { let total = 0; function next(): number { return ++total; } return next; }\n" +
		"const next = counter(); console.log(`${next()}/${next()}/${parallelMap(items, (item) => item * 2).join(',')}`);\n"
	outcome := checkout.Try(source, filepath.Join(directory, "program"))
	code, err := os.ReadFile(filepath.Join(directory, "program", "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(code), "#define ADAMIC_CANONICAL_CLOSURES 1\n") {
		t.Fatal("counter did not exercise canonical closure runtime layouts")
	}
	if outcome.Verdict != Agreed || string(outcome.Node.Stdout) != "1/2/2,4\n" {
		t.Fatalf("captured counter: %s %s\n%s\nNode: %#v", outcome.Verdict, outcome.Key, outcome.Detail, outcome.Node)
	}
}
