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
	table := DynamicKeyedArrayPrototypeNames.Names()
	for _, name := range names {
		if name != "length" && !slices.Contains(table, name) {
			t.Errorf("Node's Array.prototype has %q, which DynamicKeyedArrayPrototypeNames lacks", name)
		}
	}
	for index, name := range table {
		if name == "length" || !slices.Contains(names, name) {
			t.Errorf("DynamicKeyedArrayPrototypeNames has %q, which isn't one of Node's Array.prototype names an array inherits", name)
		}
		if slices.Contains(table[:index], name) {
			t.Errorf("DynamicKeyedArrayPrototypeNames has %q twice", name)
		}
	}
}

// Each of DynamicKeyedArrayPrototypeNames' lengths is the pinned Node's Array.prototype[name].length,
// both ways: natively a run-time key read hands out a closure carrying the table's length
// (dynamic_keyed_names.h, union.c), where the JavaScript backend hands out Node's own function, so a
// length the table has wrong reads wrong natively alone. Node is the toolchain's, and a run without
// it fails. Mutant: one length changed in the table.
func TestDynamicKeyedArrayPrototypeLengthsAreNodes(t *testing.T) {
	t.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	output, err := toolchain.Command(node, "-e", "console.log(JSON.stringify(Object.fromEntries(Object.getOwnPropertyNames(Array.prototype).filter((name) => name !== 'length').map((name) => [name, Array.prototype[name].length]))))").Output()
	if err != nil {
		t.Fatal(err)
	}
	var lengths map[string]int
	if err := json.Unmarshal(output, &lengths); err != nil {
		t.Fatal(err)
	}
	if len(lengths) == 0 {
		t.Fatal("Node's Array.prototype has no functions")
	}
	table := map[string]int{}
	for _, function := range DynamicKeyedArrayPrototypeNames {
		table[function.Name] = function.Length
	}
	for name, length := range lengths {
		if held, found := table[name]; !found || held != length {
			t.Errorf("Node's Array.prototype.%s has length %d, which DynamicKeyedArrayPrototypeNames doesn't hold for it", name, length)
		}
	}
	for name, length := range table {
		if held, found := lengths[name]; !found || held != length {
			t.Errorf("DynamicKeyedArrayPrototypeNames holds %s with length %d, which isn't Node's", name, length)
		}
	}
}

// DynamicKeyedRegExpStringIteratorNames is the pinned Node's own list both ways: the string names of
// a matchAll iterator's prototype and of Iterator.prototype above it, in that order. A name Node has
// and the table lacks would read undefined where Node hands out a function, and one the table has and
// Node lacks would hand out a function where Node reads undefined. The table is whole only while the
// iterator has no own property and Object.prototype is next above, so both are held too. Node is the
// toolchain's, and a run without it fails. Mutant: one name dropped from the table.
func TestDynamicKeyedRegExpStringIteratorNamesAreNodes(t *testing.T) {
	t.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	output, err := toolchain.Command(node, "-e", "const iterator = ''.matchAll(/(?:)/g); const prototype = Object.getPrototypeOf(iterator); const above = Object.getPrototypeOf(prototype); console.log(JSON.stringify({own: Object.getOwnPropertyNames(iterator), names: [...Object.getOwnPropertyNames(prototype), ...Object.getOwnPropertyNames(above)], objectAbove: Object.getPrototypeOf(above) === Object.prototype}))").Output()
	if err != nil {
		t.Fatal(err)
	}
	var chain struct {
		Own         []string `json:"own"`
		Names       []string `json:"names"`
		ObjectAbove bool     `json:"objectAbove"`
	}
	if err := json.Unmarshal(output, &chain); err != nil {
		t.Fatal(err)
	}
	if len(chain.Own) != 0 || !chain.ObjectAbove {
		t.Fatalf("Node's matchAll iterator has own names %q, or Object.prototype isn't above Iterator.prototype (%v)", chain.Own, chain.ObjectAbove)
	}
	table := DynamicKeyedRegExpStringIteratorNames.Names()
	if !slices.Equal(chain.Names, table) {
		t.Errorf("Node's matchAll iterator inherits %q, where DynamicKeyedRegExpStringIteratorNames holds %q", chain.Names, table)
	}
}

// Each of DynamicKeyedRegExpStringIteratorNames' lengths is the pinned Node's, read off a matchAll
// iterator, so constructor's is Iterator's, through its getter: natively the closure carries the
// table's length, where the JavaScript backend hands out Node's own function, so a length the table
// has wrong reads wrong natively alone. Node is the toolchain's, and a run without it fails. Mutant:
// one length changed in the table.
func TestDynamicKeyedRegExpStringIteratorLengthsAreNodes(t *testing.T) {
	t.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	output, err := toolchain.Command(node, "-e", "const iterator = ''.matchAll(/(?:)/g); const prototype = Object.getPrototypeOf(iterator); const names = [...Object.getOwnPropertyNames(prototype), ...Object.getOwnPropertyNames(Object.getPrototypeOf(prototype))]; console.log(JSON.stringify(Object.fromEntries(names.map((name) => [name, iterator[name].length]))))").Output()
	if err != nil {
		t.Fatal(err)
	}
	var lengths map[string]int
	if err := json.Unmarshal(output, &lengths); err != nil {
		t.Fatal(err)
	}
	if len(lengths) == 0 {
		t.Fatal("Node's matchAll iterator inherits no functions")
	}
	table := map[string]int{}
	for _, function := range DynamicKeyedRegExpStringIteratorNames {
		table[function.Name] = function.Length
	}
	for name, length := range lengths {
		if held, found := table[name]; !found || held != length {
			t.Errorf("Node's matchAll iterator's %s has length %d, which DynamicKeyedRegExpStringIteratorNames doesn't hold for it", name, length)
		}
	}
	for name, length := range table {
		if held, found := lengths[name]; !found || held != length {
			t.Errorf("DynamicKeyedRegExpStringIteratorNames holds %s with length %d, which isn't Node's", name, length)
		}
	}
}
