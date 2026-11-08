package native_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// recordGraphSource is a graph program whose Object.values and Object.entries over a Record<string, Node>
// are stored into the cycle, so graphFlows plans both arrays (#y7wexf5).
const recordGraphSource = `interface Node { name: string; next: Node[]; pairs: [string, Node][] }
function valuesOf(record: Record<string, Node>): Node[] {
 return Object.values(record);
}
function entriesOf(record: Record<string, Node>): [string, Node][] {
 return Object.entries(record);
}
const a: Node = { name: 'a' + '', next: [], pairs: [] };
const b: Node = { name: 'b' + '', next: [a], pairs: [] };
a.next.push(b);
const record: Record<string, Node> = { b: b };
a.next = valuesOf(record);
a.pairs = entriesOf(record);
console.log(` + "`${a.next.length} ${a.pairs.length}`" + `);
`

var (
	// recordGraphArray is the line recordCall makes a values or entries array with (emit_records.go).
	recordGraphArray = regexp.MustCompile(`(?m)^\s*adamic_array \*\s*(\w+) = adamic_array_make\(0, [^;]*\);$`)
	// recordGraphPairAdopted is an entries pair adopted as it's made.
	recordGraphPairAdopted = regexp.MustCompile(`(\w+) = adamic_graph_adopt_owned\((\w+), adamic_object_allocation_size\((\w+)\)\);`)
)

// recordGraphBumps is every way emitted C raises a value's count: a retain, and a graph hold or escape of
// it as the value held, each perhaps through a cast.
func recordGraphBumps(name string) *regexp.Regexp {
	value := `(?:\([^()]*\)\s*)?` + regexp.QuoteMeta(name) + `\s*\)`
	return regexp.MustCompile(`adamic_retain\(\s*` + value + `|adamic_graph_(?:hold|escape)\([^;]*,\s*` + value)
}

// TestRecordGraphArraysAdoptedUnretained pins the order adamic_graph_adopt_owned relies on: a values or
// entries array is adopted with nothing but its own reference, since adamic_graph_adopt refuses an
// allocation whose count isn't 1 (runtime graph_regions.c:104). Between the array's making and its
// adoption, nothing may retain it; the adoption must be there at all; and each entries pair is adopted
// as it's made, as adamic_map_entries adopts a graph map's (map.c:297).
func TestRecordGraphArraysAdoptedUnretained(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(recordGraphSource), 0600); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) == 0 {
		t.Fatal("no graph types selected: the program pins nothing of graph adoption")
	}
	code := native.C(program)
	made := recordGraphArray.FindAllStringSubmatchIndex(code, -1)
	values, entries := 0, 0
	for _, iterate := range regexp.MustCompile(`adamic_record_iterate\(`).FindAllStringIndex(code, -1) {
		// The record's walk follows the array it fills: the last array made before it is that one.
		start, name := -1, ""
		for _, match := range made {
			if match[0] < iterate[0] {
				start, name = match[0], code[match[2]:match[3]]
			}
		}
		if start < 0 {
			t.Fatalf("a record walk at byte %d with no array made before it", iterate[0])
		}
		adopt := name + " = adamic_graph_adopt_owned(" + name + ", sizeof *" + name + ");"
		end := strings.Index(code[iterate[0]:], adopt)
		if end < 0 {
			t.Errorf("the record array %s is never adopted: no %q after its walk", name, adopt)
			continue
		}
		between := code[start : iterate[0]+end]
		if bump := recordGraphBumps(name).FindString(between); bump != "" {
			t.Errorf("the record array %s is retained before it's adopted (%s), so adamic_graph_adopt would refuse it:\n%s", name, bump, between)
		}
		if !strings.Contains(between, "->tuple = true;") {
			values++
			continue
		}
		entries++
		pairs := recordGraphPairAdopted.FindAllStringSubmatch(between, -1)
		adopted := false
		for _, pair := range pairs {
			adopted = adopted || pair[1] == pair[2] && pair[2] == pair[3] && strings.Contains(between, pair[1]+"->tuple = true;")
		}
		if !adopted {
			t.Errorf("the entries array %s's pairs aren't adopted as they're made:\n%s", name, between)
		}
	}
	// A function the emitter writes twice (a region's copy) walks twice; each copy is held the same.
	if values == 0 || entries == 0 {
		t.Errorf("found %d values and %d entries arrays adopted in the C, want each at least once", values, entries)
	}
}
