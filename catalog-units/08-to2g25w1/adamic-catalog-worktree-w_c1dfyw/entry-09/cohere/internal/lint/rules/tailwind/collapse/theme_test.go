package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind/vendored"
)

// The fixture is the theme the shipped Tailwind 4.3.3 engine resolved for every stylesheet in the
// corpus, captured by internal/lint/rules/tailwind/tools/generate_theme and checked in next to this test.
//
// Measured rather than transcribed, because the two are different claims and the ignored-key map is
// where they part. Reading `theme.ts` suggests `--font` names every key beginning `--font-`; asking
// the engine says `--font-weight-*` and `--font-size-*` are excluded from it. A hand-written
// expectation set would have encoded the plausible version and agreed with the engine on every
// namespace that has no ignored siblings, which is most of them.
//
// The corpus has two halves and needs both. The repository half is the exit criterion for this
// component: two real repositories on the same Tailwind, whose themes differ, diffed key by key
// against what the engine built from the same files. The synthetic half reaches the branches a real
// theme never does, `initial` deletion, namespace clearing, `@theme default` losing to a
// non-default value, prefixes, and the dot-to-underscore retry, which is roughly half of `theme.ts`
// and all of the half where a port fails silently.

type themeCorpus struct {
	TailwindVersion     string      `json:"tailwindVersion"`
	SyntheticCaseCount  int         `json:"syntheticCaseCount"`
	RepositoryCaseCount int         `json:"repositoryCaseCount"`
	Cases               []themeCase `json:"cases"`
}

// themeCase is one stylesheet and the theme the engine resolved from it.
//
// A synthetic case carries its stylesheet inline in Input. A repository case carries EntryPath
// instead, because the Go side must resolve the same `@import` graph from disk: the import
// following is part of what is under test, and inlining the entry file would test the loader
// against a single file it never has to resolve.
type themeCase struct {
	Name      string       `json:"name"`
	Source    string       `json:"source"`
	Input     string       `json:"input"`
	EntryPath string       `json:"entryPath"`
	Theme     themeFixture `json:"theme"`
}

type themeFixture struct {
	// Prefix is null in JSON when the theme carries none, which unmarshals to the empty string,
	// matching Theme.Prefix. See its doc comment for why the two collapse safely.
	Prefix             string                    `json:"prefix"`
	Size               int                       `json:"size"`
	Entries            []themeEntryFixture       `json:"entries"`
	NamespaceResults   []themeNamespaceFixture   `json:"namespaceResults"`
	Resolutions        []themeResolutionFixture  `json:"resolutions"`
	ResolveWithResults []themeResolveWithFixture `json:"resolveWithResults"`
	OptionQueries      []themeOptionFixture      `json:"optionQueries"`
}

type themeEntryFixture struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Options int    `json:"options"`
}

type themeNamespaceFixture struct {
	Namespace string `json:"namespace"`
	Entries   []struct {
		Key       string `json:"key"`
		KeyIsNull bool   `json:"keyIsNull"`
		Value     string `json:"value"`
	} `json:"entries"`
	KeysInNamespace []string `json:"keysInNamespace"`
}

// themeResolutionFixture is one `resolve` / `resolveValue` pair. Both are nullable, and null means
// the lookup found nothing, which is a different answer than the empty string: a theme entry
// declared with an empty value resolves successfully to "".
type themeResolutionFixture struct {
	CandidateValue       string   `json:"candidateValue"`
	CandidateValueIsNull bool     `json:"candidateValueIsNull"`
	ThemeKeys            []string `json:"themeKeys"`
	Resolve              *string  `json:"resolve"`
	ResolveValue         *string  `json:"resolveValue"`
}

type themeResolveWithFixture struct {
	CandidateValue string            `json:"candidateValue"`
	ThemeKeys      []string          `json:"themeKeys"`
	NestedKeys     []string          `json:"nestedKeys"`
	Value          *string           `json:"value"`
	Extra          map[string]string `json:"extra"`
}

type themeOptionFixture struct {
	Key        string  `json:"key"`
	Get        *string `json:"get"`
	HasDefault bool    `json:"hasDefault"`
	Options    int     `json:"options"`
}

// tailwindPackageRoots maps a repository case's name to the tailwindcss install its theme resolves
// `@import "tailwindcss"` against.
//
// Both are 4.3.3, which is the point: the two repositories differ by their own `@theme` content on
// an identical framework, so a table generated from either one is wrong for the other. That is the
// defect this whole component exists to fix, and the corpus is what proves it is real rather than
// theorized. Both are the vendored snapshot of 4.3.3, which is that identical framework, on any machine
// (#sycrdr6).
var tailwindPackageRoots = map[string]string{
	"ahra":              vendored.TailwindPackageRoot(),
	"www-connected-app": vendored.TailwindPackageRoot(),
}

func loadThemeCorpus(t *testing.T) themeCorpus {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "theme_fixtures.json"))
	if err != nil {
		t.Fatalf("read theme fixture: %v", err)
	}

	var fixture themeCorpus
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode theme fixture: %v", err)
	}
	if len(fixture.Cases) == 0 {
		t.Fatal("theme fixture holds no cases")
	}
	return fixture
}

// buildThemeForCase builds a Theme the way the case demands: a synthetic case is parsed from its
// inline stylesheet, a repository case is loaded from disk through its whole import graph.
//
// A repository case's stylesheet is spelled inside its corpus, so the case skips naming the variable
// when that corpus is unset, and fails when it is set and lacks the stylesheet. It skips loudly: a
// silently absent repository half would turn the exit criterion into a synthetic-only suite that
// still prints a green line, which is exactly the failure mode this slice has been bitten by.
func buildThemeForCase(t *testing.T, aCase themeCase) (*Theme, []SkippedDirective, bool) {
	t.Helper()

	if aCase.Source == "repository" {
		theme, skipped := loadRepositoryTheme(t, aCase)
		return theme, skipped, true
	}

	nodes, err := ParseCSS(aCase.Input)
	if err != nil {
		t.Fatalf("parse synthetic stylesheet: %v", err)
	}
	loader := &themeLoader{theme: NewTheme(), visiting: make(map[string]bool)}
	if err := loader.ingest(nodes, aCase.Name); err != nil {
		t.Fatalf("ingest synthetic stylesheet: %v", err)
	}
	return loader.theme, loader.skipped, true
}

// loadRepositoryTheme loads a repository case's theme from its corpus, through the vendored tailwindcss
// its name maps to.
func loadRepositoryTheme(t *testing.T, aCase themeCase) (*Theme, []SkippedDirective) {
	t.Helper()
	packageRoot, ok := tailwindPackageRoots[aCase.Name]
	if !ok {
		t.Fatalf("repository case %q has no tailwindcss package root registered in tailwindPackageRoots", aCase.Name)
	}
	if _, err := os.Stat(filepath.Join(packageRoot, "index.css")); err != nil {
		t.Fatalf("the vendored tailwindcss for %q is not at %s: %v", aCase.Name, packageRoot, err)
	}
	entryPath := corpus.Resolve(t, aCase.EntryPath)
	theme, skipped, err := LoadThemeFromFile(entryPath, NodeStylesheetResolver(packageRoot))
	if err != nil {
		t.Fatalf("load %s: %v", entryPath, err)
	}
	return theme, skipped
}

// coveredRepositoryCases is how many of the fixture's repository cases have their corpus set, which is
// how many a run must check: the rest skip, naming their variable.
func coveredRepositoryCases(t *testing.T, fixture themeCorpus) int {
	t.Helper()
	covered := 0
	for _, aCase := range fixture.Cases {
		if aCase.Source != "repository" {
			continue
		}
		repositoryCorpus, _, spelled, err := corpus.Spelled(aCase.EntryPath)
		if err != nil || !spelled {
			t.Fatalf("repository case %q records %q, which is not spelled <corpus>:<path in it>: %v", aCase.Name, aCase.EntryPath, err)
		}
		if repositoryCorpus.Covered() {
			covered++
		}
	}
	return covered
}

// TestThemeMatchesEngine is the differential: every answer the port gives, against the answer the
// engine gave for the same question.
//
// The counts are printed on success rather than only on failure. A suite that compared twelve
// answers and a suite that compared two hundred thousand are indistinguishable from a green line,
// and this component's whole claim is about scale: it is checked against two real design systems,
// not against a reading.
func TestThemeMatchesEngine(t *testing.T) {
	t.Parallel()
	fixture := loadThemeCorpus(t)

	var totalEntries, totalNamespaces, totalNamespaceEntries, totalKeysInNamespace int
	var totalResolutions, totalResolveWith, totalOptionQueries, totalNested int
	var repositoriesChecked, syntheticChecked int

	for _, aCase := range fixture.Cases {
		// Not parallel: its subtests add to totals the test checks after them, and a parallel subtest would run
		// only after the test had returned, so the check would pass on nothing.
		t.Run(aCase.Name, func(t *testing.T) {
			theme, _, ok := buildThemeForCase(t, aCase)
			if !ok {
				return
			}

			if theme.Prefix != aCase.Theme.Prefix {
				t.Errorf("Prefix = %q, engine says %q", theme.Prefix, aCase.Theme.Prefix)
			}
			if theme.Size() != aCase.Theme.Size {
				t.Errorf("Size() = %d, engine says %d", theme.Size(), aCase.Theme.Size)
			}

			// Entries, compared positionally. Insertion order is part of the contract; comparing
			// as sets would pass a port that returned the right keys in an arbitrary order, and Go
			// map iteration is randomized per run, so that port would be nondeterministically
			// wrong in production while green here.
			entries := theme.Entries()
			if len(entries) != len(aCase.Theme.Entries) {
				t.Errorf("Entries() returned %d entries, engine returned %d", len(entries), len(aCase.Theme.Entries))
			}
			for index := range aCase.Theme.Entries {
				if index >= len(entries) {
					t.Errorf("Entries()[%d] missing; engine has %q", index, aCase.Theme.Entries[index].Key)
					continue
				}
				want := aCase.Theme.Entries[index]
				got := entries[index]
				if got.Key != want.Key || got.Value != want.Value || int(got.Options) != want.Options {
					t.Errorf(
						"Entries()[%d] = {%q, %q, %d}, engine says {%q, %q, %d}",
						index, got.Key, got.Value, got.Options, want.Key, want.Value, want.Options,
					)
				}
			}
			totalEntries += len(aCase.Theme.Entries)

			// Namespaces. Both reads, because they differ in three ways that each change a real
			// count: the null key, sub-variables, and ignored-key filtering.
			for _, wantNamespace := range aCase.Theme.NamespaceResults {
				gotEntries := theme.Namespace(wantNamespace.Namespace)
				if len(gotEntries) != len(wantNamespace.Entries) {
					t.Errorf(
						"Namespace(%q) returned %d entries, engine returned %d",
						wantNamespace.Namespace, len(gotEntries), len(wantNamespace.Entries),
					)
				}
				for index := range wantNamespace.Entries {
					if index >= len(gotEntries) {
						break
					}
					want := wantNamespace.Entries[index]
					got := gotEntries[index]
					if got.KeyIsNull != want.KeyIsNull || got.Key != want.Key || got.Value != want.Value {
						t.Errorf(
							"Namespace(%q)[%d] = {null=%v, %q, %q}, engine says {null=%v, %q, %q}",
							wantNamespace.Namespace, index,
							got.KeyIsNull, got.Key, got.Value,
							want.KeyIsNull, want.Key, want.Value,
						)
					}
				}
				totalNamespaceEntries += len(wantNamespace.Entries)

				gotKeys := theme.KeysInNamespaces([]string{wantNamespace.Namespace})
				if !equalStringSlices(gotKeys, wantNamespace.KeysInNamespace) {
					t.Errorf(
						"KeysInNamespaces([%q]) = %s, engine says %s",
						wantNamespace.Namespace, summarizeStrings(gotKeys), summarizeStrings(wantNamespace.KeysInNamespace),
					)
				}
				totalKeysInNamespace += len(wantNamespace.KeysInNamespace)
			}
			totalNamespaces += len(aCase.Theme.NamespaceResults)

			// Resolution. Both the `var()`-building Resolve and the literal ResolveValue, since a
			// port that confused them would answer every inline entry correctly and every other
			// one with the wrong half of the pair.
			for _, want := range aCase.Theme.Resolutions {
				candidatePresent := !want.CandidateValueIsNull

				gotResolve, gotResolveOK := theme.Resolve(want.CandidateValue, candidatePresent, want.ThemeKeys, ThemeOptionNone)
				assertOptionalString(
					t, gotResolve, gotResolveOK, want.Resolve,
					fmt.Sprintf("Resolve(%s, %v)", describeCandidate(want.CandidateValue, candidatePresent), want.ThemeKeys),
				)

				gotValue, gotValueOK := theme.ResolveValue(want.CandidateValue, candidatePresent, want.ThemeKeys)
				assertOptionalString(
					t, gotValue, gotValueOK, want.ResolveValue,
					fmt.Sprintf("ResolveValue(%s, %v)", describeCandidate(want.CandidateValue, candidatePresent), want.ThemeKeys),
				)
			}
			totalResolutions += len(aCase.Theme.Resolutions)

			// ResolveWith and its nested keys, the `-*--nested` suffix form.
			for _, want := range aCase.Theme.ResolveWithResults {
				gotValue, gotExtra, gotOK := theme.ResolveWith(want.CandidateValue, want.ThemeKeys, want.NestedKeys)
				assertOptionalString(
					t, gotValue, gotOK, want.Value,
					fmt.Sprintf("ResolveWith(%q, %v, %v)", want.CandidateValue, want.ThemeKeys, want.NestedKeys),
				)
				if !gotOK {
					continue
				}
				if len(gotExtra) != len(want.Extra) {
					t.Errorf(
						"ResolveWith(%q, %v, %v) extra has %d keys, engine has %d",
						want.CandidateValue, want.ThemeKeys, want.NestedKeys, len(gotExtra), len(want.Extra),
					)
				}
				for nestedKey, wantNested := range want.Extra {
					gotNested, present := gotExtra[nestedKey]
					if !present {
						t.Errorf(
							"ResolveWith(%q, %v, %v) extra missing %q; engine says %q",
							want.CandidateValue, want.ThemeKeys, want.NestedKeys, nestedKey, wantNested,
						)
						continue
					}
					if gotNested != wantNested {
						t.Errorf(
							"ResolveWith(%q, %v, %v) extra[%q] = %q, engine says %q",
							want.CandidateValue, want.ThemeKeys, want.NestedKeys, nestedKey, gotNested, wantNested,
						)
					}
				}
				totalNested += len(want.Extra)
			}
			totalResolveWith += len(aCase.Theme.ResolveWithResults)

			// Get, HasDefault and GetOptions. These read the options bitfield, which is what
			// decides whether Resolve returns a `var()` or a literal, so a port that stored the
			// values correctly and the bits wrongly would pass every count above.
			for _, want := range aCase.Theme.OptionQueries {
				gotGet, gotGetOK := theme.Get([]string{want.Key})
				assertOptionalString(t, gotGet, gotGetOK, want.Get, fmt.Sprintf("Get([%q])", want.Key))

				if got := theme.HasDefault(want.Key); got != want.HasDefault {
					t.Errorf("HasDefault(%q) = %v, engine says %v", want.Key, got, want.HasDefault)
				}
				if got := theme.GetOptions(want.Key); int(got) != want.Options {
					t.Errorf("GetOptions(%q) = %d, engine says %d", want.Key, got, want.Options)
				}
			}
			totalOptionQueries += len(aCase.Theme.OptionQueries)

			if aCase.Source == "repository" {
				repositoriesChecked++
			} else {
				syntheticChecked++
			}
		})
	}

	// Coverage. A count that has never been shown to be large is not a large count, and a suite
	// that skipped its repository half would otherwise report the same green line as one that ran
	// it. Every repository case whose corpus is set must have run; the rest skipped naming their
	// variable, so the run says what it did not compare.
	if covered := coveredRepositoryCases(t, fixture); repositoriesChecked != covered {
		t.Errorf("%d repository cases ran and %d have their corpus set: the exit criterion for this component is a comparison against two real design systems", repositoriesChecked, covered)
	}
	if syntheticChecked != fixture.SyntheticCaseCount {
		t.Errorf("ran %d of %d synthetic cases", syntheticChecked, fixture.SyntheticCaseCount)
	}

	totalComparisons := totalEntries + totalNamespaceEntries + totalKeysInNamespace +
		totalResolutions*2 + totalResolveWith + totalNested + totalOptionQueries*3

	t.Logf(
		"tailwind %s: %d repository and %d synthetic design systems; %d theme entries, "+
			"%d namespaces holding %d entries and %d namespace keys, %d resolutions, "+
			"%d resolveWith answers carrying %d nested values, %d option queries; %d compared answers",
		fixture.TailwindVersion, repositoriesChecked, syntheticChecked,
		totalEntries, totalNamespaces, totalNamespaceEntries, totalKeysInNamespace,
		totalResolutions, totalResolveWith, totalNested, totalOptionQueries,
		totalComparisons,
	)
}

// TestThemeRepositoriesDiffer is the measurement the whole port argues from.
//
// cohere shipped generated tables built from one repository and linted three with them. That is
// only a defect if two repositories on the same Tailwind actually resolve different themes, so this
// asserts it directly rather than leaving it as the premise of a design document. If this test ever
// passes trivially because the two themes became identical, the assertion below fails rather than
// quietly succeeding.
func TestThemeRepositoriesDiffer(t *testing.T) {
	t.Parallel()
	fixture := loadThemeCorpus(t)

	type resolved struct {
		name string
		keys map[string]string
		size int
	}
	var themes []resolved

	for _, aCase := range fixture.Cases {
		if aCase.Source != "repository" {
			continue
		}
		theme, _ := loadRepositoryTheme(t, aCase)
		keys := make(map[string]string, theme.Size())
		for _, entry := range theme.Entries() {
			keys[entry.Key] = entry.Value
		}
		themes = append(themes, resolved{name: aCase.Name, keys: keys, size: theme.Size()})
	}

	if len(themes) < 2 {
		t.Fatalf("need two repositories to compare, have %d", len(themes))
	}

	first, second := themes[0], themes[1]

	var onlyInFirst, onlyInSecond, differingValues int
	for key, value := range first.keys {
		otherValue, present := second.keys[key]
		switch {
		case !present:
			onlyInFirst++
		case otherValue != value:
			differingValues++
		}
	}
	for key := range second.keys {
		if _, present := first.keys[key]; !present {
			onlyInSecond++
		}
	}

	if onlyInFirst == 0 && onlyInSecond == 0 && differingValues == 0 {
		t.Errorf(
			"%s and %s resolved identical themes; the argument for this port is that they do not, "+
				"so either the corpus regressed to one repository or the loader is not reading the repository in front of it",
			first.name, second.name,
		)
	}

	t.Logf(
		"%s resolved %d entries, %s resolved %d: %d keys only in %s, %d only in %s, %d shared keys with different values",
		first.name, first.size, second.name, second.size,
		onlyInFirst, first.name, onlyInSecond, second.name, differingValues,
	)
}

// TestThemeLoaderReportsSkippedDirectives pins which directives the loader knows it is not acting
// on.
//
// `@config` can, upstream, contribute theme values through a JavaScript config that cohere never
// runs. That is a real gap in this component, and the honest handling is to report it: a
// repository that started declaring theme values in TypeScript would produce a theme quietly short
// by that many keys, and this is what makes that arrive as a visible directive rather than as a
// wrong count. See the boundary note in themeloader.go.
func TestThemeLoaderReportsSkippedDirectives(t *testing.T) {
	t.Parallel()
	fixture := loadThemeCorpus(t)

	ran := 0
	for _, aCase := range fixture.Cases {
		if aCase.Source != "repository" {
			continue
		}
		_, skipped := loadRepositoryTheme(t, aCase)

		// Both corpus repositories carry exactly one `@config`, pointing at a TypeScript file used
		// for content globs. Anything else appearing here is a change in what the repository asks
		// of Tailwind and must be read rather than absorbed.
		if len(skipped) != 1 {
			t.Errorf("%s: expected 1 skipped directive, got %d: %v", aCase.Name, len(skipped), skipped)
			continue
		}
		if skipped[0].Name != "@config" {
			t.Errorf("%s: skipped %s, expected @config", aCase.Name, skipped[0].Name)
		}
		ran++
	}

	if ran == 0 {
		t.Fatal("the fixture holds no repository case, so this test checked nothing")
	}
	t.Logf("%d repositories each skip exactly one @config directive", ran)
}

// TestThemeInsertionOrderSurvivesRedefinition pins the JavaScript `Map.set` behaviour that a Go
// port is most likely to get wrong, in isolation from the corpus.
//
// The corpus covers it, but only incidentally and only where a repository happens to redefine a
// framework key. Pinned here it names the rule: overwriting keeps position, deleting and re-adding
// does not. A port that appends on every write reorders every redefinition, which is what a
// repository's `@theme` does to nearly every framework default it overrides.
func TestThemeInsertionOrderSurvivesRedefinition(t *testing.T) {
	t.Parallel()
	theme := NewTheme()
	for _, entry := range [][2]string{{"--color-a", "red"}, {"--color-b", "blue"}, {"--color-c", "green"}} {
		if err := theme.Add(entry[0], entry[1], ThemeOptionNone); err != nil {
			t.Fatalf("Add(%q): %v", entry[0], err)
		}
	}

	if err := theme.Add("--color-a", "rebeccapurple", ThemeOptionNone); err != nil {
		t.Fatalf("redefine: %v", err)
	}
	assertThemeKeyOrder(t, theme, "overwrite keeps position", "--color-a", "--color-b", "--color-c")
	if value, _ := theme.Get([]string{"--color-a"}); value != "rebeccapurple" {
		t.Errorf("--color-a = %q after redefinition, want rebeccapurple", value)
	}

	// Deleting genuinely gives up the position, so re-adding lands at the end.
	if err := theme.Add("--color-a", "initial", ThemeOptionNone); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := theme.Add("--color-a", "red", ThemeOptionNone); err != nil {
		t.Fatalf("re-add: %v", err)
	}
	assertThemeKeyOrder(t, theme, "delete then re-add moves to the end", "--color-b", "--color-c", "--color-a")
}

// TestThemeKeyOrderCompacts checks that the order slice does not grow without bound under repeated
// deletion, and that compaction preserves order.
//
// Order is tracked in a slice that deletion does not compact in place, so without the amortized
// reclaim in compactKeyOrder a stylesheet that cleared a namespace repeatedly would grow it
// forever. The bound is the property worth pinning; the exact threshold is an implementation
// choice and is not asserted.
func TestThemeKeyOrderCompacts(t *testing.T) {
	t.Parallel()
	theme := NewTheme()

	const rounds = 500
	for round := 0; round < rounds; round++ {
		key := fmt.Sprintf("--churn-%d", round)
		if err := theme.Add(key, "x", ThemeOptionNone); err != nil {
			t.Fatalf("Add: %v", err)
		}
		if err := theme.Add(key, "initial", ThemeOptionNone); err != nil {
			t.Fatalf("delete: %v", err)
		}
	}
	if err := theme.Add("--survivor", "y", ThemeOptionNone); err != nil {
		t.Fatalf("Add survivor: %v", err)
	}

	if theme.Size() != 1 {
		t.Errorf("Size() = %d after %d add/delete rounds, want 1", theme.Size(), rounds)
	}
	assertThemeKeyOrder(t, theme, "only the survivor remains", "--survivor")

	// The slice is bounded by twice the live size plus the in-flight dead entries, never by the
	// number of writes. A port without compaction reaches `rounds` here.
	if len(theme.keyOrder) > 2*theme.Size()+2 {
		t.Errorf("keyOrder holds %d slots for %d live keys after %d rounds; compaction is not reclaiming", len(theme.keyOrder), theme.Size(), rounds)
	}
}

// TestThemeAddRejectsInvalidNamespaceClear pins that a clear directive with a value other than
// `initial` is an error rather than an entry.
//
// Upstream throws. A port that stored `--color-*: red` as a key named `--color-*` would answer
// every namespace read with one extra bogus member, and no corpus stylesheet contains the input
// that would reveal it.
func TestThemeAddRejectsInvalidNamespaceClear(t *testing.T) {
	t.Parallel()
	theme := NewTheme()
	if err := theme.Add("--color-*", "red", ThemeOptionNone); err == nil {
		t.Fatal("Add(\"--color-*\", \"red\") returned no error; the engine throws")
	}
	if theme.Size() != 0 {
		t.Errorf("Size() = %d after a rejected clear directive, want 0", theme.Size())
	}
}

// TestEscapeRoundTrip checks escapeCSSIdentifier and unescapeCSSIdentifier against the cases whose
// answers a reasonable guess gets wrong.
//
// These run on every `var()` the theme builds and on every key GetOptions is asked about, and the
// rules are not intuitive: a lone `-` is escaped, a leading digit becomes a hex escape with a
// trailing space, and a digit in second position is escaped only after a `-`.
func TestEscapeRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		value  string
		escape string
	}{
		{"--color-red-500", "--color-red-500"},
		{"-", "\\-"},
		{"--", "--"},
		{"--color-a/b", "--color-a\\/b"},
		{"--spacing-1.5", "--spacing-1\\.5"},
		{"0abc", "\\30 abc"},
		{"-0abc", "-\\30 abc"},
		{"--a b", "--a\\ b"},
		{"--emoji-\U0001F49C", "--emoji-\U0001F49C"},
		{"--aéb", "--aéb"},
	}

	for _, aCase := range cases {
		if got := escapeCSSIdentifier(aCase.value); got != aCase.escape {
			t.Errorf("escapeCSSIdentifier(%q) = %q, want %q", aCase.value, got, aCase.escape)
		}
		if got := unescapeCSSIdentifier(aCase.escape); got != aCase.value {
			t.Errorf("unescapeCSSIdentifier(%q) = %q, want %q", aCase.escape, got, aCase.value)
		}
	}

	// Unescape-only cases: inputs the escaper never produces but a stylesheet can contain.
	unescapeOnly := []struct{ escaped, want string }{
		{"--a\\62 c", "--abc"},
		{"--a\\000062c", "--abc"},
		{"--a\\0 b", "--a�b"},
		{"--a\\d800 b", "--a�b"},
		{"--a\\110000 b", "--a�b"},
		{"--a\\\\b", "--a\\b"},
		{"--a\\", "--a\\"},
		{"--a\\\U0001F49Cb", "--a\U0001F49Cb"},
	}
	for _, aCase := range unescapeOnly {
		if got := unescapeCSSIdentifier(aCase.escaped); got != aCase.want {
			t.Errorf("unescapeCSSIdentifier(%q) = %q, want %q", aCase.escaped, got, aCase.want)
		}
	}

	t.Logf("%d escape round trips and %d unescape-only cases", len(cases), len(unescapeOnly))
}

// assertThemeKeyOrder compares the theme's live key order against an expected sequence.
func assertThemeKeyOrder(t *testing.T, theme *Theme, what string, want ...string) {
	t.Helper()

	var got []string
	for _, entry := range theme.Entries() {
		got = append(got, entry.Key)
	}
	if !equalStringSlices(got, want) {
		t.Errorf("%s: key order is %v, want %v", what, got, want)
	}
}

// assertOptionalString compares a (value, ok) pair against the fixture's nullable string.
//
// The distinction between "resolved to the empty string" and "did not resolve" is what this exists
// to preserve. A theme entry declared with an empty value resolves successfully to "", and a port
// that returned "" for a miss would agree with the engine on every hit and disagree on every miss
// while looking identical in a printout.
func assertOptionalString(t *testing.T, got string, ok bool, want *string, what string) {
	t.Helper()

	if want == nil {
		if ok {
			t.Errorf("%s = %q, engine resolved nothing", what, got)
		}
		return
	}
	if !ok {
		t.Errorf("%s resolved nothing, engine says %q", what, *want)
		return
	}
	if got != *want {
		t.Errorf("%s = %q, engine says %q", what, got, *want)
	}
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// summarizeStrings renders a slice for an error message, truncated so that a disagreement on a
// 558-key namespace reports its shape rather than scrolling the failure off the screen.
func summarizeStrings(values []string) string {
	const limit = 8
	if len(values) <= limit {
		return fmt.Sprintf("%d%v", len(values), values)
	}
	return fmt.Sprintf("%d[%s ...]", len(values), strings.Join(values[:limit], " "))
}

// describeCandidate renders a candidate value for an error message, keeping the null case
// distinguishable from the empty-string case.
func describeCandidate(value string, present bool) string {
	if !present {
		return "null"
	}
	return fmt.Sprintf("%q", value)
}
