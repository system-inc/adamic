package tailwind

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestEveryTableIsTailwindsOrHasAStatedReason is the inventory this package owes its reader.
//
// The port removed the tables that did not need to exist. What remains splits in two, and the split
// is the whole claim: a table is either Tailwind's own data, which we mirror, or it answers a
// question the port's own machinery structurally cannot, which is stated per table below.
//
// # Tailwind's own data
//
// Upstream carries these and there is no algorithm behind them to port. `property-order.ts` at 4.3.3
// is a hand-written array of 359 property names, with comments in it about how to make `inset-x-0`
// come before `top-0`, and `getPropertySort` calls `.indexOf` on it. Mirroring it is the faithful
// thing; reading it at lint time would be the same table with extra steps.
//
//	PropertyOrder                  359   property-order.ts
//	FrameworkVariantRegistrations   88   variants.ts registrations
//	FrameworkStaticDeclarations    890   staticUtility calls
//	FrameworkFunctionalUtilities    33   functionalUtility registrations
//	FrameworkMultiDeclarationUtilities
//	                               152   functionalUtility registrations
//	VariantOrder                   145   the order numbers those registrations sort at
//	FrameworkNamespaces             14   the theme namespaces a root consumes, all Tailwind's own
//	baseReadings                   374   readings of Tailwind's own utilities, held by the differential
//	baseStatics                    869   readings of Tailwind's own statics, same
//	GapRootDescriptions             57   how each closure-registered root resolves a value, read
//	                                     from utilities.ts at the tag. Tailwind's own registration
//	                                     data, in the same sense as the 185 in the two tables above:
//	                                     these 57 are registered as bare closures rather than through
//	                                     `functionalUtility`, so the shape had to be written out
//	                                     instead of extracted, and gaproots_test.go holds all 57
//	                                     against `candidatesToCss` on a live design system.
//	valueOperators                   6   the operators the arbitrary-value printer unspaces, the
//	                                     literal set in `printArbitraryValue` at 4.3.3
//
// # Ours, each answering something a reading cannot
//
// A reading is a sorted set of property positions and a count. `getPropertySort` reads
// `node.property` and increments a counter, and never reads a value. That single fact is what made
// this port tractable, deleting roughly 85% of `utilities.ts`, and it is also the reason these two
// tables cannot be derived from what the port carries.
//
//	baseDescriptors          78   the ordered type list each root infers against. The stated reason
//	                              used to be that the reading partitions on the resolved type and no
//	                              registration carries both halves; the emitting half is ported now
//	                              and that claim is measured rather than argued: 396 of 399 cells
//	                              agree with the emitters across all three axes, and the 3 that do
//	                              not are one dead cell no class can reach. So the readings are
//	                              redundant and the table is not, because `InferDataType` returns the
//	                              first match in `TypeList` and the order decides which cell a class
//	                              lands in. `bg` infers `3px` as a position and `border` as a length
//	                              from the same value. An emitter answers every type it is asked and
//	                              so cannot rank them; see
//	                              TestTypeListOrderDecidesTheCellAndEmittersCannotSupplyIt.
//	CollapseFamilies         45   which pairs canonicalize into a third root. Upstream does ship
//	                              shorthand data, which this entry used to deny: two maps keyed on
//	                              CSS property, `padding-inline: [padding-left, padding-right]` and
//	                              its kin, covering inset, margin, padding, scroll-margin,
//	                              scroll-padding, border width/style/colour, gap, overflow,
//	                              overscroll-behavior and ten logical block/inline pairs. They do not
//	                              derive this. This is keyed on utility roots, `pl + pr => px`, and
//	                              bridging the two key spaces is `canonicalizeCandidates`, the
//	                              signature-equivalence search enumerate.mjs prices at eight to ten
//	                              thousand lines that changes every Tailwind minor. It also holds
//	                              families no CSS shorthand covers: `h + w => size`, `skew`,
//	                              `translate`, the `rounded-*` corners, the `mask-*` edges and
//	                              `border-spacing`, which collapse through a utility relationship
//	                              rather than a property one.
//
// `RootDeclaredProperties` and its three companions were here until #mz0m6k8, carried for exactly the
// reason the paragraph below rejected. The emitting half is now ported, so the answer is computed by
// `DeclaredPropertiesFor` and the four tables, 1,183 entries and 62 KB, are deleted.
//
// `ComposingRoots` went the same way in #gb4bkgc, and its reason had read as the strongest of the
// four. A reading cannot answer whether two values of a root layer or overwrite, which is true and
// was never the question: the emitters can, once they stop flattening the constant each layering
// family writes to its shared property. 81 of its 82 rows answered and all 81 agreed before it was
// deleted, and the 82nd was not a utility. Its answers are kept as a fixture in composes_test.go.
//
// # Why not port the emitting half and delete the rest
//
// Measured rather than assumed. `utilities.ts` holds 582 `decl()` call sites and 374 of them compute
// their value, which means porting `color-mix`, `withAlpha`, `calc` and the 729-reference `--tw-*`
// var chain. That is the surface this port deleted on purpose. It would cost more than the four
// tables together and buy nothing: the differential holds at 95,931 answered with 17 disagreements,
// all named and tested.
//
// # What this test actually asserts
//
// The counts, so a table growing or shrinking without its reason being revisited fails here. The
// reasons themselves are prose and cannot be asserted, which is why each table also carries its own
// test: the collapse families assert their shorthand structure, the descriptor rows are compared
// class for class by the differential, and the upstream diff catches Tailwind's own data moving.
func TestEveryTableIsTailwindsOrHasAStatedReason(t *testing.T) {
	t.Parallel()
	upstream := map[string]int{
		"PropertyOrder":                 len(PropertyOrder),
		"FrameworkVariantRegistrations": len(FrameworkVariantRegistrations),
		"FrameworkStaticDeclarations":   len(FrameworkStaticDeclarations),
		"FrameworkFunctionalUtilities":  len(FrameworkFunctionalUtilities),
		// Spelled in full, because the hand-written list this replaced said
		// `FrameworkMultiDeclaration` and nothing noticed: a name in a list nobody parses can be
		// wrong forever.
		"FrameworkMultiDeclarationUtilities": len(FrameworkMultiDeclarationUtilities),
		"VariantOrder":                       len(VariantOrder),
		"FrameworkNamespaces":                len(FrameworkNamespaces),
		"baseReadings":                       len(baseReadings),
		"baseStatics":                        len(baseStatics),
		"GapRootDescriptions":                len(GapRootDescriptions),
		"valueOperators":                     len(valueOperators),
	}
	ours := map[string]int{
		"baseDescriptors":  len(baseDescriptors),
		"CollapseFamilies": len(CollapseFamilies),
	}

	for name, count := range upstream {
		if count == 0 {
			t.Errorf("%s is empty; a table mirroring Tailwind's own data cannot be", name)
		}
	}
	for name, count := range ours {
		if count == 0 {
			t.Errorf("%s is empty; if it is genuinely no longer needed, delete it and its entry here "+
				"rather than leaving a zero that reads like a working table", name)
		}
	}

	// The count of tables, not their contents. A fifth table of ours appearing without an entry in
	// the doc comment above is the thing this catches: the inventory going stale is how a table ends
	// up carried for no stated reason, which is what this package started with.
	if len(ours) != 2 {
		t.Errorf("the inventory lists %d tables of our own; the doc comment above accounts for 2, so "+
			"one has been added or removed without its reason being written down", len(ours))
	}

	// Every table in the package is accounted for above, and this half is read out of the source
	// rather than restated.
	//
	// The first version of this test asserted "exactly four tables of ours" over a list it wrote
	// itself, so `KnownRoots` and `KnownStatics` sat in the same file, dead since 58fb982, and passed.
	// The second version named the tables in a slice, which was the same defect one step removed: the
	// slice was written by hand from the two maps above it, so a table nobody typed into it could
	// never be caught. It listed 8 while the package held 35.
	//
	// This parses the package and finds every top-level composite-literal table, so a table added to
	// a file is a failure here whether or not anyone remembered this test. That is the difference
	// between a measurement and a restatement, which is the distinction this package keeps rebuilding
	// the hard way.
	declared := tablesDeclaredInPackage(t)

	accounted := make(map[string]bool, len(upstream)+len(ours))
	for name := range upstream {
		accounted[name] = true
	}
	for name := range ours {
		accounted[name] = true
	}

	var unaccounted []string
	for name := range declared {
		if !accounted[name] && !tablesExemptFromTheInventory[name] {
			unaccounted = append(unaccounted, fmt.Sprintf("%s (%s, %d entries)", name, declared[name].file, declared[name].entries))
		}
	}
	sort.Strings(unaccounted)
	for _, name := range unaccounted {
		t.Errorf("%s is a table in this package with no entry in the inventory above", name)
	}

	for name := range accounted {
		if _, isDeclared := declared[name]; !isDeclared {
			t.Errorf("%s has an inventory entry and is not declared in this package, so the two have drifted", name)
		}
	}

	t.Logf("Tailwind's own: %v", upstream)
	t.Logf("ours, each with a stated reason: %v", ours)
}

// tableInPackage is one top-level composite-literal table found in this package's source.
type tableInPackage struct {
	file    string
	entries int
}

// tablesDeclaredInPackage parses this package and returns every top-level map or slice literal.
//
// Reading the source rather than a list, because a list is what the two previous versions of this
// check were and neither could see a table nobody added to it. A parse sees what is there.
//
// Composite literals only. A table is a set of entries written down, which is the thing that goes
// stale and the thing this inventory is about; a `var x = someCall()` is code and is not.
func tablesDeclaredInPackage(t *testing.T) map[string]tableInPackage {
	t.Helper()

	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("listing the package's files: %v", err)
	}

	declared := make(map[string]tableInPackage)
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, declaration := range parsed.Decls {
			general, isGeneral := declaration.(*ast.GenDecl)
			if !isGeneral || general.Tok != token.VAR {
				continue
			}
			for _, specification := range general.Specs {
				value, isValue := specification.(*ast.ValueSpec)
				if !isValue {
					continue
				}
				for index, name := range value.Names {
					if index >= len(value.Values) {
						continue
					}
					literal, isLiteral := value.Values[index].(*ast.CompositeLit)
					if !isLiteral {
						continue
					}
					switch literal.Type.(type) {
					case *ast.MapType, *ast.ArrayType:
						declared[name.Name] = tableInPackage{file: path, entries: len(literal.Elts)}
					}
				}
			}
		}
	}

	if len(declared) == 0 {
		t.Fatal("the parse found no tables at all, so this check measured nothing")
	}
	return declared
}

// Tables the inventory deliberately does not carry an entry for, each with the reason.
//
// The inventory is about tables that answer a question upstream answers by running code, because
// those are the ones that go stale against a repository or a Tailwind version. Three kinds do not:
//
// The ported handle bodies. `frameworkEmitters`, `frameworkMultiEmitters`,
// `frameworkMultiLiteralEmitters` and `gapEmitters` are the port's own source, one closure per root
// read at the tag. They are code written as a map, not data recording what code produced, and
// TestEveryFrameworkRootHasAnEmitter already holds them against the registration tables.
//
// The CSS grammar. `lengthUnits`, `angleUnits`, `mathFunctions`, `imageFunctions`,
// `gradientFunctions`, `colorFunctions`, `namedColors`, `genericNames`, `absoluteSizes`,
// `backgroundPositionKeywords`, `percentSuffix` and `colorKeywords` are facts about CSS rather than
// about Tailwind or about a repository. They move when the CSS specification moves, which is not the
// drift this inventory watches, and upstream carries the same lists for the same reason.
//
// Small local constants. `ignoredThemeKeys`, `bareValueDataTypes`, `themedModifierValues`,
// `composedAggregateValues` and `FrameworkVariantComparisonGroups` are each read by one function in
// the file that declares them and are part of that function's definition.
//
// An entry here is a claim that a table cannot go stale against a repository. Adding one to silence
// a failure rather than because that claim is true is the failure mode, and it is not defended
// against: measured, adding a two-entry table to this package and an exemption for it turns the
// check green. That is the honest limit of this test. It catches a table added and forgotten, which
// is what happened to `KnownRoots`, `KnownStatics` and the three ordering tables, and it does not
// catch a table added deliberately with a false reason.
//
// Each is grouped under a reason above rather than listed bare, so the claim is at least written
// down where review can see it.
var tablesExemptFromTheInventory = map[string]bool{
	"frameworkEmitters":             true,
	"frameworkMultiEmitters":        true,
	"frameworkMultiLiteralEmitters": true,
	"gapEmitters":                   true,

	"lengthUnits":                true,
	"angleUnits":                 true,
	"mathFunctions":              true,
	"imageFunctions":             true,
	"gradientFunctions":          true,
	"colorFunctions":             true,
	"namedColors":                true,
	"genericNames":               true,
	"absoluteSizes":              true,
	"backgroundPositionKeywords": true,
	"percentSuffix":              true,
	"colorKeywords":              true,

	"ignoredThemeKeys":                 true,
	"bareValueDataTypes":               true,
	"themedModifierValues":             true,
	"composedAggregateValues":          true,
	"FrameworkVariantComparisonGroups": true,
}
