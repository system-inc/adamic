package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// The table under test, built at test time from the same measurements the generator consumes.
//
// # Why the table is loaded rather than generated into the package
//
// The generated Go is per repository, and checking a copy in would do the exact thing this port
// exists to undo: bake one repository's tokens into a file claiming to describe Tailwind. That is
// not a stylistic position, it is measured. Generating against ~/Projects/ahra and
// ~/Projects/connected/www-connected-app produces byte-identical descriptor arrays and different
// namespace sets, and a synthetic theme whose only colour token is `--color-weird-1` yields a
// `--color-weird` namespace that changes the reading on 20 roots.
//
// So the suite reads the same JSON the generator reads and builds the table in memory. That tests
// the model and the lookup, which is what this task owns, and leaves the generator's Go-printing
// tested separately by the fact that its output compiles and produces this same structure.
//
// The fixture and the table come from one extraction, so a drift between them is impossible by
// construction rather than by discipline.

var (
	testTableOnce  sync.Once
	testTableValue *Table
	testTableError error
)

func testTable(t *testing.T) *Table {
	t.Helper()
	testTableOnce.Do(func() {
		testTableValue, testTableError = loadTestTable()
	})
	if testTableError != nil {
		t.Fatalf("building the test table: %v", testTableError)
	}
	return testTableValue
}

type extractedReading struct {
	Order []int `json:"order"`
	Count int   `json:"count"`
}

type extractedRoot struct {
	Root                                 string                       `json:"root"`
	TypeList                             []string                     `json:"typeList"`
	ReadingByType                        map[string]*extractedReading `json:"readingByType"`
	Fallback                             *extractedReading            `json:"fallback"`
	ReadingByNamespace                   map[string]*extractedReading `json:"readingByNamespace"`
	ReadingByLiteral                     map[string]*extractedReading `json:"readingByLiteral"`
	Empty                                *extractedReading            `json:"empty"`
	EmptyWithModifier                    *extractedReading            `json:"emptyWithModifier"`
	EmptyWithThemedModifier              *extractedReading            `json:"emptyWithThemedModifier"`
	ReadingByTypeWithModifier            map[string]*extractedReading `json:"readingByTypeWithModifier"`
	FallbackWithModifier                 *extractedReading            `json:"fallbackWithModifier"`
	ReadingByNamespaceWithModifier       map[string]*extractedReading `json:"readingByNamespaceWithModifier"`
	ReadingByTypeWithThemedModifier      map[string]*extractedReading `json:"readingByTypeWithThemedModifier"`
	FallbackWithThemedModifier           *extractedReading            `json:"fallbackWithThemedModifier"`
	ReadingByNamespaceWithThemedModifier map[string]*extractedReading `json:"readingByNamespaceWithThemedModifier"`
}

type extractedTableFile struct {
	TailwindVersion string                       `json:"tailwindVersion"`
	Roots           []extractedRoot              `json:"roots"`
	Statics         map[string]*extractedReading `json:"statics"`
}

type extractedContextFile struct {
	TailwindVersion     string              `json:"tailwindVersion"`
	Namespaces          []string            `json:"namespaces"`
	KeysByNamespace     map[string][]string `json:"keysByNamespace"`
	PropertyOrder       []string            `json:"propertyOrder"`
	PerDeclarationRoots []string            `json:"perDeclarationRoots"`
	// UnlistedStatics are statics the engine compiles and getClassList() does not advertise; see
	// context.mjs. `order-none` is the only one on this design system, and it is a real class an
	// author can write, so a table without it declines a class the engine reads.
	UnlistedStatics map[string]*extractedReading `json:"unlistedStatics"`
}

func loadTestTable() (*Table, error) {
	var tableFile extractedTableFile
	if err := readJSON(filepath.Join("testdata", "descriptor_table.json"), &tableFile); err != nil {
		return nil, err
	}
	var contextFile extractedContextFile
	if err := readJSON(filepath.Join("testdata", "descriptor_context.json"), &contextFile); err != nil {
		return nil, err
	}

	perDeclaration := make(map[string]bool, len(contextFile.PerDeclarationRoots))
	for _, root := range contextFile.PerDeclarationRoots {
		perDeclaration[root] = true
	}

	table := &Table{
		TailwindVersion: tableFile.TailwindVersion,
		Descriptors:     make(map[string]*Descriptor, len(tableFile.Roots)),
		Statics:         make(map[string]Reading, len(tableFile.Statics)),
		Namespaces:      contextFile.Namespaces,
		KeysByNamespace: make(map[string]map[string]bool, len(contextFile.KeysByNamespace)),
		PropertyOrder:   make(map[string]int, len(contextFile.PropertyOrder)),
	}

	for namespace, keys := range contextFile.KeysByNamespace {
		set := make(map[string]bool, len(keys))
		for _, key := range keys {
			set[key] = true
		}
		table.KeysByNamespace[namespace] = set
	}
	for index, property := range contextFile.PropertyOrder {
		table.PropertyOrder[property] = index
	}
	// The unlisted statics first, so the registry's own measurement wins any collision.
	for name, value := range contextFile.UnlistedStatics {
		if value == nil {
			continue
		}
		table.Statics[name] = Reading{Order: value.Order, Count: value.Count}
	}
	for name, value := range tableFile.Statics {
		if value == nil {
			continue
		}
		table.Statics[name] = Reading{Order: value.Order, Count: value.Count}
	}

	for index := range tableFile.Roots {
		entry := &tableFile.Roots[index]
		descriptor := &Descriptor{
			Root:           entry.Root,
			PerDeclaration: perDeclaration[entry.Root],
			ByLiteral:      readingMap(entry.ReadingByLiteral),
			Absent: AxisReadings{
				ByType:      typeReadingMap(entry.ReadingByType, entry.Fallback),
				ByNamespace: readingMap(entry.ReadingByNamespace),
				Fallback:    readingOrZero(entry.Fallback),
				Empty:       readingPointer(entry.Empty),
			},
			Alpha: AxisReadings{
				ByType:      typeReadingMap(entry.ReadingByTypeWithModifier, entry.FallbackWithModifier),
				ByNamespace: readingMap(entry.ReadingByNamespaceWithModifier),
				Fallback:    readingOrZero(entry.FallbackWithModifier),
				Empty:       readingPointer(entry.EmptyWithModifier),
			},
			Themed: AxisReadings{
				ByType:      typeReadingMap(entry.ReadingByTypeWithThemedModifier, entry.FallbackWithThemedModifier),
				ByNamespace: readingMap(entry.ReadingByNamespaceWithThemedModifier),
				Fallback:    readingOrZero(entry.FallbackWithThemedModifier),
				Empty:       readingPointer(entry.EmptyWithThemedModifier),
			},
		}
		// The order is carried verbatim, never sorted. See TestTypeListOrderIsCarriedNotSorted.
		for _, typeName := range entry.TypeList {
			descriptor.TypeList = append(descriptor.TypeList, DataType(typeName))
		}
		table.Descriptors[entry.Root] = descriptor
	}

	return table, nil
}

// readingMap keeps every entry, because a miss in a namespace map continues the bare-value
// precedence rather than ending it. See AxisReadings.ByNamespace for the 647 buckets that proved it.
func readingMap(entries map[string]*extractedReading) map[string]Reading {
	result := map[string]Reading{}
	for key, value := range entries {
		if value == nil {
			continue
		}
		result[key] = Reading{Order: value.Order, Count: value.Count}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func typeReadingMap(entries map[string]*extractedReading, fallback *extractedReading) map[DataType]Reading {
	result := map[DataType]Reading{}
	for key, value := range entries {
		if value == nil || sameExtractedReading(value, fallback) {
			continue
		}
		result[DataType(key)] = Reading{Order: value.Order, Count: value.Count}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func sameExtractedReading(left, right *extractedReading) bool {
	if left == nil || right == nil {
		return left == right
	}
	if left.Count != right.Count || len(left.Order) != len(right.Order) {
		return false
	}
	for index, value := range left.Order {
		if value != right.Order[index] {
			return false
		}
	}
	return true
}

func readingOrZero(value *extractedReading) Reading {
	if value == nil {
		return Reading{}
	}
	return Reading{Order: value.Order, Count: value.Count}
}

func readingPointer(value *extractedReading) *Reading {
	if value == nil {
		return nil
	}
	return &Reading{Order: value.Order, Count: value.Count}
}

func readJSON(path string, into any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, into)
}
