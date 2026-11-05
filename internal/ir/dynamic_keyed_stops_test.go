package ir

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// DynamicKeyedArrayPrototypeNames is the pinned Node's own list, less length, both ways: a name Node
// has and the table lacks would read undefined where Node hands out a method, and one the table has
// and Node lacks would stop where Node reads undefined. Node is the toolchain's, and a run without it
// fails rather than passing having compared nothing. Mutant: one name dropped from the table.
func TestDynamicKeyedArrayPrototypeNamesAreNodes(t *testing.T) {
	t.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	output, err := toolchain.Command(node, "-e", "console.log(JSON.stringify(Object.getOwnPropertyNames(Array.prototype)))").Output()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(output, &names); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(names, "length") {
		t.Fatalf("Node's Array.prototype has no own length: %q", names)
	}
	for _, name := range names {
		if name != "length" && !slices.Contains(DynamicKeyedArrayPrototypeNames, name) {
			t.Errorf("Node's Array.prototype has %q, which DynamicKeyedArrayPrototypeNames lacks", name)
		}
	}
	for index, name := range DynamicKeyedArrayPrototypeNames {
		if name == "length" || !slices.Contains(names, name) {
			t.Errorf("DynamicKeyedArrayPrototypeNames has %q, which isn't one of Node's Array.prototype names an array inherits", name)
		}
		if slices.Contains(DynamicKeyedArrayPrototypeNames[:index], name) {
			t.Errorf("DynamicKeyedArrayPrototypeNames has %q twice", name)
		}
	}
}
