package yaml

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Not parallel: native.Build writes the shared user cache (adamic/runtime or adamic/units).
func TestSharedSliceAppendMatchesNode(t *testing.T) {
	t.Parallel()
	entry, err := filepath.Abs("gaps/sharedSliceAppend.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	expected := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "0")
	if string(expected) != "a\nx\n" {
		t.Fatalf("Node %q", expected)
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "shared")
	if err := native.Build(native.C(lowered), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	sanitized := filepath.Join(t.TempDir(), "shared-sanitized")
	if err := native.Build(native.C(lowered), sanitized, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, offset := range []string{"0", "48"} {
		t.Run(offset, func(t *testing.T) {
			t.Parallel()
			expected := run(t, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, offset)
			if string(expected) != "a\nx\n" {
				t.Fatalf("Node %q", expected)
			}
			for _, executable := range []string{binary, sanitized} {
				actual := run(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, executable, offset)
				if !bytes.Equal(actual, expected) {
					t.Fatalf("%s: native %q Node %q", executable, actual, expected)
				}
			}
		})
	}
}
