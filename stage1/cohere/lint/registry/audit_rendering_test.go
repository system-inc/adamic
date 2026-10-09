package registry

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDeterministicRegeneration(t *testing.T) {
	t.Parallel()
	root := copyRules(t)
	descriptors, err := Generate(root)
	if err != nil {
		t.Fatal(err)
	}
	ts, goSource := Render(descriptors)
	verifyAuditRegistryRendering(t, descriptors)
	for _, name := range []string{"registry.ts", "registry.go"} {
		path := filepath.Join(root, ".generated", name)
		before, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Generate(root); err != nil {
			t.Fatal(err)
		}
		after, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if !before.ModTime().Equal(after.ModTime()) {
			t.Fatalf("unchanged %s rewritten", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		expected := ts
		if name == "registry.go" {
			expected = goSource
		}
		if !bytes.Equal(data, expected) {
			t.Fatal("generation changed bytes")
		}
	}
}

// Preserve the generator's ordering, then compare both complete products. These
// two real registrations exercise ordered versus unordered rules and the parent
// argument used by cohere's fused walk.
func verifyAuditRegistryRendering(t *testing.T, descriptors []Descriptor) {
	t.Helper()
	var selected []Descriptor
	for _, d := range descriptors {
		if d.Name == "no-debugger" || d.Name == "adamic/no-type-predicate" {
			selected = append(selected, d)
		}
	}
	if len(selected) != 2 {
		t.Fatalf("audit registry has %d rules, want 2", len(selected))
	}
	ts, goSource := Render(selected)
	for _, output := range []struct {
		name string
		data []byte
	}{{"registry.ts.txt", ts}, {"registry.go.txt", goSource}} {
		want, err := os.ReadFile(filepath.Join("testdata", "audit", output.name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(output.data, want) {
			t.Fatalf("%s full rendering differs: got\n%s\nwant\n%s", output.name, output.data, want)
		}
	}
}
