package tailwind

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
)

// The live table is verified against the engine, not against itself.
//
// A constructor that composes two halves can be self-consistent and wrong: the halves can agree with
// each other and disagree with Tailwind, which is the failure the whole descriptor slice is tested
// against rather than around. So the contract test builds a table with NewTable from a real design
// system and scores it against the engine's own readings in the fixture corpus, case for case, under
// the same rules descriptor_test.go uses for the measured table.
//
// Measured on both corpus repositories: 6,808 of 6,808 scored cases on ahra, and 6,806 of 6,806 on
// www-connected-app against that repository's own extraction. The second run is not automated here
// because it needs a fixture file generated against that repository, which is untracked for the same
// reason the table is. Reproduce it with:
//
//	node internal/lint/rules/tailwind/tools/generate_descriptor_table/fixtures.mjs \
//	    connected:app/_theme/styles/theme.css > /tmp/connected_fixtures.json
//
// The corpus repositories are read through internal/corpus. When one is unset the test skips naming
// its variable rather than passing, since a silent pass is what a broken lookup looks like, and when
// one is set but lacks its stylesheet or tailwindcss the test fails.

// corpusRepositories are the design systems the port is measured on.
//
// Both are named rather than one, because the two halves this file composes can only be told apart
// by a system whose theme differs, and a single repository would let a per-repository token pass as
// a framework fact.
var corpusRepositories = []struct {
	name     string
	spelling string
}{
	{name: "ahra", spelling: "ahra:app/_theme/styles/theme.css"},
	{name: "connected", spelling: "connected:app/_theme/styles/theme.css"},
}

// liveEntryPointFor resolves a corpus repository's stylesheet and the tailwindcss installed beside it: a
// skip naming the variable when the corpus is unset, and a failure when it is set but lacks either.
func liveEntryPointFor(t testing.TB, spelling string) (entryPoint, packageRoot string) {
	t.Helper()
	entryPoint = corpus.Resolve(t, spelling)
	packageRoot = findTailwindPackageRootForTest(filepath.Dir(entryPoint))
	if packageRoot == "" {
		t.Fatalf("%s holds no installed tailwindcss reachable from %s", spelling, entryPoint)
	}
	return entryPoint, packageRoot
}

// liveTableFor loads a repository's design system and builds its table, or skips.
func liveTableFor(t *testing.T, spelling string) (*LoadedDesignSystem, *Table) {
	t.Helper()
	entryPoint, packageRoot := liveEntryPointFor(t, spelling)
	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		t.Fatalf("loading the design system at %s: %v", entryPoint, err)
	}
	table := NewTable(system)
	if table == nil {
		t.Fatalf("NewTable returned nil for %s", entryPoint)
	}
	return system, table
}

// findTailwindPackageRootForTest walks up looking for an installed tailwindcss.
//
// A copy of the rules package's walk rather than an import, because internal/tailwind must not
// depend on internal/rules/tailwind; the dependency runs the other way.
func findTailwindPackageRootForTest(start string) string {
	directory := start
	for {
		candidate := filepath.Join(directory, "node_modules", "tailwindcss")
		if info, err := os.Stat(filepath.Join(candidate, "index.css")); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return ""
		}
		directory = parent
	}
}

// TestLiveTableAgreesWithTheEngineOverTheFixtureCorpus is the contract.
//
// The live table is scored against the engine's own readings, exactly as descriptor_test.go scores
// the measured table, over the same corpus and with the same rules: a case with no engine reading is
// reported and never counted as agreement, and a decline is only free on a per-declaration root.
//
// Scoring against the engine rather than against the measured table is deliberate. Two tables built
// from one extraction can agree with each other and disagree with Tailwind, and the whole point of
// this constructor is that its checked-in half was measured to be a fact about the framework. The
// engine is the only thing that can refuse it.
func TestLiveTableAgreesWithTheEngineOverTheFixtureCorpus(t *testing.T) {
	t.Parallel()
	fixtures := loadDescriptorFixtures(t)
	_, live := liveTableFor(t, corpusRepositories[0].spelling)

	if fixtures.TailwindVersion != live.TailwindVersion {
		t.Fatalf("the fixtures are Tailwind %s and the live table is Tailwind %s", fixtures.TailwindVersion, live.TailwindVersion)
	}

	agreed, mismatched, declined, withoutReading := 0, 0, 0, 0
	for _, one := range fixtures.Cases {
		candidate := one.parsed()
		got, answered := live.Lookup(candidate)
		if !answered {
			declined++
			// A decline is only free where the model states it cannot answer. Anywhere else it is
			// the failure that blocked #4q5dsn3: a rule going silent and scoring as agreement.
			if one.Reading != nil {
				descriptor, found := live.Descriptors[one.Root]
				if !found || !descriptor.PerDeclaration {
					mismatched++
					if mismatched <= 20 {
						t.Errorf("%s: the live table declined but the engine read %v#%d",
							one.ClassName, one.Reading.Order, one.Reading.Count)
					}
				}
			}
			continue
		}
		if one.Reading == nil {
			// No engine reading is neither an agreement nor a disagreement. See the same branch in
			// descriptor_test.go for why these exist and why they are counted rather than skipped.
			withoutReading++
			continue
		}
		want := Reading{Order: one.Reading.Order, Count: one.Reading.Count}
		if !got.Equal(want) {
			mismatched++
			if mismatched <= 20 {
				t.Errorf("%s (root %q, %s %q, modifier %s %q): the live table read %v#%d, the engine said %v#%d",
					one.ClassName, one.Root, one.ValueKind, one.Value, one.ModifierKind, one.ModifierValue,
					got.Order, got.Count, want.Order, want.Count)
			}
			continue
		}
		agreed++
	}

	if mismatched > 0 {
		t.Errorf("%d of %d cases disagreed with the engine", mismatched, len(fixtures.Cases))
	}
	scored := agreed + mismatched
	t.Logf("the live table agreed on %d of %d scored cases over %d roots", agreed, scored, fixtures.Counts.Roots)
	t.Logf("  %d cases had no engine reading and were reported rather than scored", withoutReading)
	t.Logf("  %d lookups declined", declined)

	// Volume assertions, in the extractor's spirit: a suite that agreed on everything because it
	// tested almost nothing reports the same green line as one that did the work.
	if withoutReading > len(fixtures.Cases)/2 {
		t.Errorf("%d of %d cases had no reading; the corpus is mostly unscored", withoutReading, len(fixtures.Cases))
	}
	if agreed < 5000 {
		t.Errorf("only %d agreements; the corpus is too small for the number to mean anything", agreed)
	}
}

// TestLiveTableCarriesTheRepositoryTheme asserts the per-repository half is actually present.
//
// This is the test the whole task turns on. A table that silently described bare Tailwind while
// claiming to describe this repository would answer most classes correctly and be wrong on exactly
// the tokens the repository added, which is the failure the port exists to remove and the one no
// aggregate agreement number would surface.
func TestLiveTableCarriesTheRepositoryTheme(t *testing.T) {
	t.Parallel()
	for _, repository := range corpusRepositories {
		t.Run(repository.name, func(t *testing.T) {
			t.Parallel()
			system, table := liveTableFor(t, repository.spelling)
			checkTableCarriesTheTheme(t, repository.name, system, table)
		})
	}
}

// checkTableCarriesTheTheme is the composition TestLiveTableCarriesTheRepositoryTheme asserts, shared with
// its public twin: namespaces longest-first and drawn from the theme, and every `@utility` block composed
// into the table, functional ones marked PerDeclaration.
func checkTableCarriesTheTheme(t *testing.T, name string, system *LoadedDesignSystem, table *Table) {
	t.Helper()

	if len(table.Namespaces) == 0 {
		t.Fatal("the table carries no namespaces, so every bare value resolves through inference alone")
	}
	// Longest-first is the engine's precedence and a contract of the field.
	for index := 1; index < len(table.Namespaces); index++ {
		if len(table.Namespaces[index-1]) < len(table.Namespaces[index]) {
			t.Fatalf("namespaces are not longest-first at %d: %q before %q",
				index, table.Namespaces[index-1], table.Namespaces[index])
		}
	}

	// Every namespace the table names must actually hold keys from this theme, and every
	// key must be one the theme reports. Otherwise the table's namespaces came from
	// somewhere other than the repository.
	for _, namespace := range table.Namespaces {
		keys := table.KeysByNamespace[namespace]
		if len(keys) == 0 {
			t.Fatalf("namespace %q holds no keys", namespace)
		}
		fromTheme := map[string]bool{}
		for _, key := range system.Theme().KeysInNamespaces([]string{namespace}) {
			fromTheme[key] = true
		}
		for key := range keys {
			if !fromTheme[key] {
				t.Fatalf("namespace %q holds key %q, which this repository's theme does not", namespace, key)
			}
		}
	}

	// The repository's own contributions, by count rather than by name, so the assertion
	// holds on a repository this test has never seen.
	staticRoots, functionalRoots := 0, 0
	// Each kind separately rather than a switch, because a root can be declared both ways:
	// `@utility fade-in` and `@utility fade-in-*` are two blocks naming one root, and
	// sixteen roots in this repository have that shape. A switch counted each such root once
	// and checked only whichever kind happened to win.
	for root, kinds := range system.utilityRoots {
		if kinds[UtilityKindStatic] {
			if _, found := table.Statics[root]; !found {
				t.Errorf("static `@utility %s` is missing from the table's statics", root)
			}
			staticRoots++
		}
		if kinds[UtilityKindFunctional] {
			descriptor, found := table.Descriptors[root]
			if !found {
				t.Errorf("functional `@utility %s` is missing from the table's descriptors", root)
				continue
			}
			if !descriptor.PerDeclaration {
				t.Errorf("functional `@utility %s` is not marked PerDeclaration, so the table would answer it from a row rather than declining to the evaluator", root)
			}
			functionalRoots++
		}
	}
	if staticRoots+functionalRoots == 0 {
		t.Fatal("this repository declares no `@utility` blocks, so this test proved nothing about composition")
	}
	t.Logf("%s: %d namespaces, %d static and %d functional `@utility` roots composed in",
		name, len(table.Namespaces), staticRoots, functionalRoots)
}

// TestLiveTableBaseHalfCarriesNoRepositoryTokens is the other side of the same claim.
//
// The checked-in half must be a fact about Tailwind. Every namespace it keys a reading on has to be
// one of the framework's, or the file is carrying one repository's tokens under a name that says
// otherwise, which is the exact bug descriptor.go's comment on Table refuses.
func TestLiveTableBaseHalfCarriesNoRepositoryTokens(t *testing.T) {
	t.Parallel()
	checked := 0
	for root, descriptor := range baseDescriptors {
		for _, axis := range []AxisReadings{descriptor.Absent, descriptor.Alpha, descriptor.Themed} {
			for namespace := range axis.ByNamespace {
				if namespace == namespaceColorKeyword || namespace == namespaceNone {
					continue
				}
				if !FrameworkNamespaces[namespace] {
					t.Errorf("%s keys a reading on namespace %q, which is not one of the framework's", root, namespace)
				}
				checked++
			}
		}
	}
	if checked == 0 {
		t.Fatal("no namespace buckets were checked, so this test proved nothing")
	}
	t.Logf("%d namespace buckets in the base table, all keyed on the %d framework namespaces", checked, len(FrameworkNamespaces))
}

// TestLiveTableDeclinesOnlyWhereTheEvaluatorAnswers is the price of declining, measured.
//
// The live table declines on every repository `@utility` root, where the measured table carries a
// row for the ones that are not per-declaration. That is a deliberate loss and this test is what
// stops it being a silent one: every case the live table declines and the measured table answers
// must be recoverable from the `@utility` evaluator, or have no engine reading at all.
//
// Without this, "the evaluator picks it up" is a claim in a comment. A comment cannot fail.
func TestLiveTableDeclinesOnlyWhereTheEvaluatorAnswers(t *testing.T) {
	t.Parallel()
	fixtures := loadDescriptorFixtures(t)
	measured := testTable(t)
	system, live := liveTableFor(t, corpusRepositories[0].spelling)

	evaluator := system.Utilities()
	if evaluator == nil {
		t.Fatal("this repository declares no `@utility` blocks, so this test proved nothing")
	}

	moved, recovered, unscored := 0, 0, 0
	for _, one := range fixtures.Cases {
		candidate := one.parsed()
		if _, liveAnswered := live.Lookup(candidate); liveAnswered {
			continue
		}
		if _, measuredAnswered := measured.Lookup(candidate); !measuredAnswered {
			continue
		}
		moved++
		if one.Reading == nil {
			// The measured table answered a case the engine gave no reading for, so nothing is lost
			// by declining it: it was never scorable.
			unscored++
			continue
		}
		reading, answered := evaluator.Reading(candidate)
		if !answered {
			t.Errorf("%s: the live table declined, the measured table answered, and the evaluator cannot answer it either", one.ClassName)
			continue
		}
		want := Reading{Order: one.Reading.Order, Count: one.Reading.Count}
		if !reading.Equal(want) {
			t.Errorf("%s: the evaluator read %v#%d where the engine said %v#%d",
				one.ClassName, reading.Order, reading.Count, want.Order, want.Count)
			continue
		}
		recovered++
	}

	if moved == 0 {
		t.Fatal("no cases moved from answered to declined, so this test proved nothing")
	}
	t.Logf("%d cases the live table declines and the measured table answers: %d recovered exactly by the evaluator, %d had no engine reading",
		moved, recovered, unscored)
}

// TestLiveTableIsNilWithoutADesignSystem pins the failing-safe behaviour.
//
// A bare framework table would answer every class confidently and be wrong on every repository
// token. Nil forces the caller to decline instead.
func TestLiveTableIsNilWithoutADesignSystem(t *testing.T) {
	t.Parallel()
	if table := NewTable(nil); table != nil {
		t.Fatalf("NewTable(nil) returned a table; a caller with no design system must decline rather than read a framework-only one")
	}
}

// BenchmarkNewTable is what the table costs on top of the design system it is built from.
//
// The relevant comparison is `518fec5`'s 78.758µs for the design system build itself. This runs on
// the same counted path, once per program, so the question a reader has is whether adding it changed
// the order of magnitude of a per-run cost, not whether it is fast in isolation.
func BenchmarkNewTable(benchmark *testing.B) {
	entryPoint, packageRoot := liveEntryPointFor(benchmark, corpusRepositories[0].spelling)
	system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
	if err != nil {
		benchmark.Fatalf("loading the design system: %v", err)
	}

	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for range benchmark.N {
		if table := NewTable(system); table == nil {
			benchmark.Fatal("NewTable returned nil")
		}
	}
}

// BenchmarkLoadDesignSystemWithTable is the whole per-run cost the rules actually pay.
//
// Load plus table, which is what `loadDesignSystemForProgram` does once per program.
func BenchmarkLoadDesignSystemWithTable(benchmark *testing.B) {
	entryPoint, packageRoot := liveEntryPointFor(benchmark, corpusRepositories[0].spelling)

	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for range benchmark.N {
		system, err := LoadDesignSystem(LoadOptions{EntryPoint: entryPoint, TailwindPackageRoot: packageRoot})
		if err != nil {
			benchmark.Fatalf("loading the design system: %v", err)
		}
		if table := NewTable(system); table == nil {
			benchmark.Fatal("NewTable returned nil")
		}
	}
}
