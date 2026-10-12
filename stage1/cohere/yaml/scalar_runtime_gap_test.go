package yaml

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestSharedSliceAppendMatchesNode(t *testing.T) {
	t.Parallel()
	binary := yamlSharedSliceBuild.binary(t, "gaps/sharedSliceAppend.ts")
	sanitized := yamlSharedSliceSanitizedBuild.binary(t, "gaps/sharedSliceAppend.ts")
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
	for _, offset := range []string{"0", "48"} {
		// Not parallel: a parallel subtest waits for a slot of its own, and the wait counts in its parent's time.
		t.Run(offset, func(t *testing.T) {
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
