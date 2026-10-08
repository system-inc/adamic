package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var viewWriteStorageFixtures = []string{"boolean", "boolean-reader", "number", "string", "object", "boolean-nullish", "boolean-mixed", "boolean-required", "boolean-dictionary", "boolean-required-undefined"}

func TestViewWriteStorageCoherence(t *testing.T) {
	t.Parallel()
	for _, name := range viewWriteStorageFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/optional_widening/storage_coherence", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".expected")
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			want := node
			if diff := disagreement(run{stdout: expected}, want); diff != "" {
				t.Fatal("Node record: " + diff)
			}
			if diagnostic, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".stderr"); err == nil {
				want = run{exitCode: 70, stderr: diagnostic}
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			t.Logf("expected backend: stdout=%q stderr=%q exit=%d", want.stdout, want.stderr, want.exitCode)
			t.Logf("Node: stdout=%q stderr=%q exit=%d", node.stdout, node.stderr, node.exitCode)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			sanitized, _ := nativelyUncached(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"native-sanitized", sanitized},
				{"native-release", releasedUncached(t, program)},
				{"javascript", onJavaScriptBackend(t, program)},
			} {
				t.Logf("%s: stdout=%q stderr=%q exit=%d", backend.name, backend.got.stdout, backend.got.stderr, backend.got.exitCode)
				if diff := disagreement(want, backend.got); diff != "" {
					t.Errorf("%s: %s", backend.name, diff)
				}
			}
		})
	}
}

func init() {
	for _, name := range viewWriteStorageFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/lower/testdata/optional_widening/storage_coherence/" + name + ".a", lowers: true, checked: name == "boolean-required-undefined"})
	}
}

// This subset remains measurable when unrelated inherited fixtures block the full count writer.
func TestViewWriteStorageCounts(t *testing.T) {
	t.Parallel()
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range viewWriteStorageFixtures {
		path := "internal/lower/testdata/optional_widening/storage_coherence/" + name + ".a"
		row := counted(t, path, false, nil, false, false)
		if !strings.Contains(string(recorded), row+"\n") {
			t.Errorf("counts changed: %s", row)
		}
	}
}
