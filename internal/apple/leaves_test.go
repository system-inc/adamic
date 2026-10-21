//go:build darwin

package apple

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/apple/generate"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// The leaf table and the declarations left out for deprecation are generated beside the bindings
// from this Mac's SDK, and that's what the compiler reads (internal/apple/generate, leaves.go). These
// are the same two, checked in, so a new SDK that adds a reference path, or a declaration deprecated
// or put back, changes a reviewed file instead of changing what the cycle finder believes silently.
// ADAMIC_UPDATE_LEAVES=1 writes them from the SDK.
func TestLeafTableIsTheSDKs(t *testing.T) {
	output, err := generate.FromSDK(context.Background(), "macosx", "macos", []string{"AppKit", "CoreGraphics", "Foundation"})
	if err != nil {
		t.Fatal(err)
	}
	deprecated := []string{}
	for _, file := range output.Files {
		for _, line := range strings.Split(string(file.Content), "\n") {
			if strings.HasPrefix(line, "// Skipped ") && strings.HasSuffix(line, ": deprecated on macos.") {
				deprecated = append(deprecated, strings.TrimPrefix(line, "// Skipped ")+" ("+strings.TrimSuffix(file.Path, ".d.ts")+")")
			}
		}
	}
	sort.Strings(deprecated)
	generated := map[string][]byte{
		"leaves-macos.txt":     output.Leaves,
		"deprecated-macos.txt": []byte("# Declarations the bindings leave out because they're deprecated on macOS: Apple's \"don't use\"\n# (docs/apple.md). Generated with the leaf table; TestLeafTableIsTheSDKs holds it to the SDK.\n" + strings.Join(deprecated, "\n") + "\n"),
	}
	for name, content := range generated {
		if os.Getenv("ADAMIC_UPDATE_LEAVES") != "" {
			if err := os.WriteFile(name, content, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		checkedIn, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if string(checkedIn) != string(content) {
			t.Errorf("%s differs from what this SDK generates: review the difference, then ADAMIC_UPDATE_LEAVES=1 go test -run TestLeafTableIsTheSDKs ./internal/apple/", name)
		}
	}
	if !strings.Contains(string(output.Leaves), "\nleaf NSURL\n") || strings.Contains(string(output.Leaves), "\nleaf NSTimer\n") {
		t.Error("the ruling's leaves moved: Url should be one, Timer not")
	}
}

// The ruling's fixtures (#91ha8gs), against the bindings generated from this Mac's SDK: each program
// says on a want: line whether it's accepted or what refuses it.
func TestDelegateCyclesAgainstTheSDK(t *testing.T) {
	programs, err := filepath.Glob(filepath.Join("testdata", "cycles", "*.a"))
	if err != nil || len(programs) == 0 {
		t.Fatalf("no programs in testdata/cycles: %v", err)
	}
	for _, program := range programs {
		t.Run(filepath.Base(program), func(t *testing.T) {
			source, err := os.ReadFile(program)
			if err != nil {
				t.Fatal(err)
			}
			want := ""
			for _, line := range strings.Split(string(source), "\n") {
				if found, ok := strings.CutPrefix(line, "// want: "); ok {
					want = found
				}
			}
			if want == "" {
				t.Fatal("no want: line")
			}
			loaded, err := load.Load([]string{program})
			if err == nil {
				_, err = lower.Lower(context.Background(), loaded)
			}
			switch {
			case want == "accepted" && err != nil:
				t.Fatalf("refused: %v", err)
			case want != "accepted" && (err == nil || !strings.Contains(err.Error(), want)):
				t.Fatalf("want a refusal containing %q, got %v", want, err)
			}
		})
	}
}
