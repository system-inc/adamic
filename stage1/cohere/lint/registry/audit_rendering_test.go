package registry

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDeterministicRegeneration(test *testing.T) {
	test.Parallel()
	root := copyRules(test)
	descriptors, generateError := Generate(root)
	if generateError != nil {
		test.Fatal(generateError)
	}
	ts, goSource := Render(descriptors)
	verifyAuditRegistryRendering(test, descriptors)
	for _, name := range []string{"registry.ts", "registry.go"} {
		path := filepath.Join(root, ".generated", name)
		before, statError := os.Stat(path)
		if statError != nil {
			test.Fatal(statError)
		}
		if _, generateError := Generate(root); generateError != nil {
			test.Fatal(generateError)
		}
		after, statError := os.Stat(path)
		if statError != nil {
			test.Fatal(statError)
		}
		if !before.ModTime().Equal(after.ModTime()) {
			test.Fatalf("unchanged %s rewritten", name)
		}
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		expected := ts
		if name == "registry.go" {
			expected = goSource
		}
		if !bytes.Equal(data, expected) {
			test.Fatal("generation changed bytes")
		}
	}
}

// Preserve the generator's ordering, then compare both complete products. These
// two real registrations exercise ordered versus unordered rules and the parent
// argument used by cohere's fused walk.
func verifyAuditRegistryRendering(test *testing.T, descriptors []Descriptor) {
	test.Helper()
	var selected []Descriptor
	for _, d := range descriptors {
		if d.Name == "no-debugger" || d.Name == "adamic/no-type-predicate" {
			selected = append(selected, d)
		}
	}
	if len(selected) != 2 {
		test.Fatalf("audit registry has %d rules, want 2", len(selected))
	}
	ts, goSource := Render(selected)
	for _, output := range []struct {
		name string
		data []byte
	}{{"registry.ts.txt", ts}, {"registry.go.txt", goSource}} {
		want, readError := os.ReadFile(filepath.Join("testdata", "audit", output.name))
		if readError != nil {
			test.Fatal(readError)
		}
		if !bytes.Equal(output.data, want) {
			test.Fatalf("%s full rendering differs: got\n%s\nwant\n%s", output.name, output.data, want)
		}
	}
}
