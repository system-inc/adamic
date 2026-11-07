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
func TestDynamicKeyedArrayPrototypeNamesAreNodes(test *testing.T) {
	test.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	output, err := toolchain.Command(
		node,
		"-e",
		"console.log(JSON.stringify(Object.getOwnPropertyNames(Array.prototype)))",
	).Output()
	if err != nil {
		test.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(output, &names); err != nil {
		test.Fatal(err)
	}
	if !slices.Contains(names, "length") {
		test.Fatalf("Node's Array.prototype has no own length: %q", names)
	}
	table := DynamicKeyedArrayPrototypeNames.Names()
	for _, name := range names {
		if name != "length" && !slices.Contains(table, name) {
			test.Errorf("Node's Array.prototype has %q, which DynamicKeyedArrayPrototypeNames lacks", name)
		}
	}
	for index, name := range table {
		if name == "length" || !slices.Contains(names, name) {
			test.Errorf(
				"DynamicKeyedArrayPrototypeNames has %q, which isn't one of Node's Array.prototype names an array inherits",
				name,
			)
		}
		if slices.Contains(table[:index], name) {
			test.Errorf("DynamicKeyedArrayPrototypeNames has %q twice", name)
		}
	}
}

// Each of DynamicKeyedArrayPrototypeNames' lengths is the pinned Node's Array.prototype[name].length,
// both ways: natively a run-time key read hands out a closure carrying the table's length
// (dynamic_keyed_names.h, union.c), where the JavaScript backend hands out Node's own function, so a
// length the table has wrong reads wrong natively alone. Node is the toolchain's, and a run without
// it fails. Mutant: one length changed in the table.
func TestDynamicKeyedArrayPrototypeLengthsAreNodes(test *testing.T) {
	test.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	output, err := toolchain.Command(
		node,
		"-e",
		"console.log(JSON.stringify(Object.fromEntries(Object.getOwnPropertyNames(Array.prototype).filter((name) => name !== 'length').map((name) => [name, Array.prototype[name].length]))))",
	).Output()
	if err != nil {
		test.Fatal(err)
	}
	var lengths map[string]int
	if err := json.Unmarshal(output, &lengths); err != nil {
		test.Fatal(err)
	}
	if len(lengths) == 0 {
		test.Fatal("Node's Array.prototype has no functions")
	}
	table := map[string]int{}
	for _, function := range DynamicKeyedArrayPrototypeNames {
		table[function.Name] = function.Length
	}
	for name, length := range lengths {
		if held, found := table[name]; !found || held != length {
			test.Errorf(
				"Node's Array.prototype.%s has length %d, which DynamicKeyedArrayPrototypeNames doesn't hold for it",
				name,
				length,
			)
		}
	}
	for name, length := range table {
		if held, found := lengths[name]; !found || held != length {
			test.Errorf("DynamicKeyedArrayPrototypeNames holds %s with length %d, which isn't Node's", name, length)
		}
	}
}

// DynamicKeyedRegExpStringIteratorNames is the pinned Node's own list both ways: the string names of
// a matchAll iterator's prototype and of Iterator.prototype above it, in that order. A name Node has
// and the table lacks would read undefined where Node hands out a function, and one the table has and
// Node lacks would hand out a function where Node reads undefined. The table is whole only while the
// iterator has no own property and Object.prototype is next above, so both are held too. Node is the
// toolchain's, and a run without it fails. Mutant: one name dropped from the table.
func TestDynamicKeyedRegExpStringIteratorNamesAreNodes(test *testing.T) {
	test.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	output, err := toolchain.Command(
		node,
		"-e",
		"const iterator = ''.matchAll(/(?:)/g); const prototype = Object.getPrototypeOf(iterator); const above = Object.getPrototypeOf(prototype); console.log(JSON.stringify({own: Object.getOwnPropertyNames(iterator), names: [...Object.getOwnPropertyNames(prototype), ...Object.getOwnPropertyNames(above)], objectAbove: Object.getPrototypeOf(above) === Object.prototype}))",
	).Output()
	if err != nil {
		test.Fatal(err)
	}
	var chain struct {
		Own         []string `json:"own"`
		Names       []string `json:"names"`
		ObjectAbove bool     `json:"objectAbove"`
	}
	if err := json.Unmarshal(output, &chain); err != nil {
		test.Fatal(err)
	}
	if len(chain.Own) != 0 || !chain.ObjectAbove {
		test.Fatalf(
			"Node's matchAll iterator has own names %q, or Object.prototype isn't above Iterator.prototype (%v)",
			chain.Own,
			chain.ObjectAbove,
		)
	}
	table := DynamicKeyedRegExpStringIteratorNames.Names()
	if !slices.Equal(chain.Names, table) {
		test.Errorf(
			"Node's matchAll iterator inherits %q, where DynamicKeyedRegExpStringIteratorNames holds %q",
			chain.Names,
			table,
		)
	}
}

// Each of DynamicKeyedRegExpStringIteratorNames' lengths is the pinned Node's, read off a matchAll
// iterator, so constructor's is Iterator's, through its getter: natively the closure carries the
// table's length, where the JavaScript backend hands out Node's own function, so a length the table
// has wrong reads wrong natively alone. Node is the toolchain's, and a run without it fails. Mutant:
// one length changed in the table.
func TestDynamicKeyedRegExpStringIteratorLengthsAreNodes(test *testing.T) {
	test.Parallel()
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	output, err := toolchain.Command(
		node,
		"-e",
		"const iterator = ''.matchAll(/(?:)/g); const prototype = Object.getPrototypeOf(iterator); const names = [...Object.getOwnPropertyNames(prototype), ...Object.getOwnPropertyNames(Object.getPrototypeOf(prototype))]; console.log(JSON.stringify(Object.fromEntries(names.map((name) => [name, iterator[name].length]))))",
	).Output()
	if err != nil {
		test.Fatal(err)
	}
	var lengths map[string]int
	if err := json.Unmarshal(output, &lengths); err != nil {
		test.Fatal(err)
	}
	if len(lengths) == 0 {
		test.Fatal("Node's matchAll iterator inherits no functions")
	}
	table := map[string]int{}
	for _, function := range DynamicKeyedRegExpStringIteratorNames {
		table[function.Name] = function.Length
	}
	for name, length := range lengths {
		if held, found := table[name]; !found || held != length {
			test.Errorf(
				"Node's matchAll iterator's %s has length %d, which DynamicKeyedRegExpStringIteratorNames doesn't hold for it",
				name,
				length,
			)
		}
	}
	for name, length := range table {
		if held, found := lengths[name]; !found || held != length {
			test.Errorf(
				"DynamicKeyedRegExpStringIteratorNames holds %s with length %d, which isn't Node's",
				name,
				length,
			)
		}
	}
}

// dynamicKeyedStatsChain is how the pinned Node's fs.Stats is laid out: its own names, and of its
// prototype, Stats.prototype, the accessors and then the functions, StatsBase.prototype's functions
// after them less its constructor, which Stats.prototype's shadows, and any name on either prototype
// that is neither. Two premises make the tables whole: the prototype is fs.Stats.prototype (read
// without calling the deprecated fs.Stats), and Object.prototype is next above StatsBase.prototype.
const dynamicKeyedStatsChain = "const fs = require('node:fs'); const stats = fs.statSync('.'); const prototype = Object.getPrototypeOf(stats); const base = Object.getPrototypeOf(prototype); const describe = (object, name) => Object.getOwnPropertyDescriptor(object, name); const functions = (object) => Object.getOwnPropertyNames(object).filter((name) => typeof describe(object, name).value === 'function'); const neither = (object) => Object.getOwnPropertyNames(object).filter((name) => typeof describe(object, name).value !== 'function' && describe(object, name).get === undefined); console.log(JSON.stringify({own: Object.getOwnPropertyNames(stats), accessors: Object.getOwnPropertyNames(prototype).filter((name) => describe(prototype, name).get !== undefined), functions: [...functions(prototype), ...functions(base).filter((name) => name !== 'constructor')], neither: [...neither(prototype), ...neither(base)], shadowed: functions(prototype).includes('constructor') && functions(base).includes('constructor'), statsPrototype: prototype === fs.Stats.prototype, objectAbove: Object.getPrototypeOf(base) === Object.prototype, lengths: Object.fromEntries([...functions(prototype), ...functions(base).filter((name) => name !== 'constructor')].map((name) => [name, stats[name].length]))}))"

type dynamicKeyedStatsLayout struct {
	Own            []string       `json:"own"`
	Accessors      []string       `json:"accessors"`
	Functions      []string       `json:"functions"`
	Neither        []string       `json:"neither"`
	Shadowed       bool           `json:"shadowed"`
	StatsPrototype bool           `json:"statsPrototype"`
	ObjectAbove    bool           `json:"objectAbove"`
	Lengths        map[string]int `json:"lengths"`
}

// dynamicKeyedStatsLayoutOfNode runs dynamicKeyedStatsChain on the toolchain's Node, and fails the
// test unless its premises hold. A run without Node fails rather than passing having compared nothing.
func dynamicKeyedStatsLayoutOfNode(test *testing.T) dynamicKeyedStatsLayout {
	test.Helper()
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	output, err := toolchain.Command(node, "-e", dynamicKeyedStatsChain).Output()
	if err != nil {
		test.Fatal(err)
	}
	var layout dynamicKeyedStatsLayout
	if err := json.Unmarshal(output, &layout); err != nil {
		test.Fatal(err)
	}
	if !layout.StatsPrototype ||
		!layout.ObjectAbove ||
		!layout.Shadowed ||
		len(layout.Neither) != 0 ||
		len(layout.Functions) == 0 {
		test.Fatalf("Node's Stats isn't laid out as the tables hold it: %+v", layout)
	}
	return layout
}

// DynamicKeyedStatsNames is the pinned Node's own list both ways: the functions of a Stats'
// prototype and then of StatsBase.prototype above it, in that order. A name Node has and the table
// lacks would read undefined where Node hands out a function, and one the table has and Node lacks
// would hand out a function where Node reads undefined. Mutant: one name dropped from the table.
func TestDynamicKeyedStatsNamesAreNodes(test *testing.T) {
	test.Parallel()
	layout := dynamicKeyedStatsLayoutOfNode(test)
	if table := DynamicKeyedStatsNames.Names(); !slices.Equal(layout.Functions, table) {
		test.Errorf(
			"Node's Stats inherits the functions %q, where DynamicKeyedStatsNames holds %q",
			layout.Functions,
			table,
		)
	}
}

// Each of DynamicKeyedStatsNames' lengths is the pinned Node's, read off a Stats, so constructor's is
// Stats' own: natively the closure carries the table's length, where the JavaScript backend hands out
// Node's own function, so a length the table has wrong reads wrong natively alone. Mutant: one length
// changed in the table.
func TestDynamicKeyedStatsLengthsAreNodes(test *testing.T) {
	test.Parallel()
	layout := dynamicKeyedStatsLayoutOfNode(test)
	table := map[string]int{}
	for _, function := range DynamicKeyedStatsNames {
		table[function.Name] = function.Length
	}
	for name, length := range layout.Lengths {
		if held, found := table[name]; !found || held != length {
			test.Errorf(
				"Node's Stats' %s has length %d, which DynamicKeyedStatsNames doesn't hold for it",
				name,
				length,
			)
		}
	}
	for name, length := range table {
		if held, found := layout.Lengths[name]; !found || held != length {
			test.Errorf("DynamicKeyedStatsNames holds %s with length %d, which isn't Node's", name, length)
		}
	}
}

// statsHeldNames are the fields node_fs_file.c's stat shape holds that Node's Stats has too: its
// stat_fields, less the private _fsFileMode. A run-time key read answers these from the shape.
var statsHeldNames = []string{"size", "mtimeMs", "mtime", "atime"}

// DynamicKeyedStatsFieldNames is exactly the pinned Node's Stats fields, its own names and then
// Stats.prototype's date accessors, less the four native holds, in order: a field Node has and the
// list lacks would read undefined natively where Node answers, and one the list has that native
// holds would stop where it could answer. Each of the four is one of Node's. Mutant: one name dropped
// from the list.
func TestDynamicKeyedStatsFieldNamesAreNodes(test *testing.T) {
	test.Parallel()
	layout := dynamicKeyedStatsLayoutOfNode(test)
	fields := append(slices.Clone(layout.Own), layout.Accessors...)
	for _, held := range statsHeldNames {
		if !slices.Contains(fields, held) {
			test.Errorf("native's Stats holds %s, which isn't one of Node's Stats fields %q", held, fields)
		}
	}
	unheld := slices.DeleteFunc(fields, func(name string) bool { return slices.Contains(statsHeldNames, name) })
	if !slices.Equal(unheld, DynamicKeyedStatsFieldNames) {
		test.Errorf(
			"Node's Stats has the fields %q beyond native's four, where DynamicKeyedStatsFieldNames holds %q",
			unheld,
			DynamicKeyedStatsFieldNames,
		)
	}
}

// A run-time key read of a Stats is closed against the pinned Node's Stats both ways (Adamic's
// conditions on #zhrjf2t): every own key of a real fs.statSync result, every accessor and every
// function of Stats.prototype and StatsBase.prototype is in exactly one of native's held fields
// (statsHeldNames), DynamicKeyedStatsNames and DynamicKeyedStatsFieldNames, so no name of Node's
// reaches the undefined a run-time key read gives any other key, and no name in the three is one Node
// lacks. Object.prototype's names, above StatsBase.prototype, answer as on any object. Mutant: ino
// dropped from DynamicKeyedStatsFieldNames, so Node's ino would read undefined natively.
func TestDynamicKeyedStatsChainIsClosed(test *testing.T) {
	test.Parallel()
	layout := dynamicKeyedStatsLayoutOfNode(test)
	chain := slices.Concat(layout.Own, layout.Accessors, layout.Functions, layout.Neither)
	tables := map[string][]string{
		"statsHeldNames":              statsHeldNames,
		"DynamicKeyedStatsNames":      DynamicKeyedStatsNames.Names(),
		"DynamicKeyedStatsFieldNames": DynamicKeyedStatsFieldNames,
	}
	for _, name := range chain {
		holders := []string{}
		for table, names := range tables {
			if slices.Contains(names, name) {
				holders = append(holders, table)
			}
		}
		if len(holders) != 1 {
			test.Errorf(
				"Node's Stats has %s, which %d of the held fields, the functions and the stop list hold (%q), not one",
				name,
				len(holders),
				holders,
			)
		}
	}
	for table, names := range tables {
		for _, name := range names {
			if !slices.Contains(chain, name) {
				test.Errorf("%s holds %s, which isn't one of Node's Stats names", table, name)
			}
		}
	}
}
