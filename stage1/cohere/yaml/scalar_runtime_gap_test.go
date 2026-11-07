package yaml

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestSharedSliceAppendMatchesNode(test *testing.T) {
	test.Parallel()
	binary := yamlSharedSliceBuild.binary(test, "gaps/sharedSliceAppend.ts")
	sanitized := yamlSharedSliceSanitizedBuild.binary(test, "gaps/sharedSliceAppend.ts")
	entry, absolutePathError := filepath.Abs("gaps/sharedSliceAppend.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	expected := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, "0")
	if string(expected) != "a\nx\n" {
		test.Fatalf("Node %q", expected)
	}
	for _, offset := range []string{"0", "48"} {
		// Not parallel: a parallel subtest waits for a slot of its own, and the wait counts in its parent's time.
		test.Run(offset, func(test *testing.T) {
			expected := run(test, "", nil, "node", "--disable-warning=ExperimentalWarning", runner, entry, offset)
			if string(expected) != "a\nx\n" {
				test.Fatalf("Node %q", expected)
			}
			for _, executable := range []string{binary, sanitized} {
				actual := run(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, executable, offset)
				if !bytes.Equal(actual, expected) {
					test.Fatalf("%s: native %q Node %q", executable, actual, expected)
				}
			}
		})
	}
}
