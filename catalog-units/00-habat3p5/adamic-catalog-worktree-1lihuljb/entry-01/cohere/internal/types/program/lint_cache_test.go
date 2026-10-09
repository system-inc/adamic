package program_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

func sampleLintCache() *program.LintCache {
	return &program.LintCache{
		Key: program.HashRuleSet([]string{"no-debugger", "no-empty"}),
		DesignSystem: &program.DesignSystemKey{
			Reads: []program.DesignSystemKeyRead{
				{Path: "/project/theme.css", Present: true, Hash: program.HashContent("@theme {}")},
				{Path: "/project/styles", Present: true, Directory: true, Hash: program.HashContent("ftheme.css")},
				{Path: "/project/node_modules/tailwindcss/index.css"},
			},
			Fingerprint: program.HashContent("a design system fingerprint"),
		},
		Entries: []program.LintCacheEntry{
			{
				Path:        "/project/source/dirty.ts",
				ContentHash: program.HashContent("debugger;\n"),
				Rules:       []string{"no-debugger", "no-empty"},
				TypedRules:  []string{"await-thenable"},
				// Not zero, so an encoder that dropped it could not match the zero a decode defaults to.
				TypeFingerprint: program.HashContent("the types debugger; can see"),
				ShapedRules:     []string{"no-floating-promises"},
				// Not zero and not the type fingerprint, so an encoder that dropped it or swapped the two
				// could not match.
				ShapeFingerprint: program.HashContent("the shapes debugger; can see"),
				DesignRules:      []string{"better-tailwindcss/no-unknown-classes"},
				// Not zero and not either other fingerprint, for the same reason.
				DesignFingerprint: program.HashContent("the design system debugger; was checked under"),
				DerivedRules:      []string{"structure/boundary-no-project-theme-value"},
				// Not zero and not any other fingerprint, for the same reason.
				DerivedFingerprint: program.HashContent("the derived debugger; was keyed on"),
				Listening:          []string{"no-debugger", "no-empty", "await-thenable", "no-floating-promises"},
				VisitedNodes:       5,
				// Two rules sharing a key and one rule with two, so an encoder that summed across either could
				// not match.
				Notes: program.RuleNotes{
					"no-empty":       {"Hub in /project/source/hub.ts": 2, "Other in /project/source/other.ts": 1},
					"await-thenable": {"Hub in /project/source/hub.ts": 3},
				},
				// A count, a zero pair and a skip, so an encoder that dropped any part or the zero pair could not match.
				Adamic: &program.AdamicRecord{
					Counts:  []program.AdamicCount{{Rule: "adamic/invariant-mutable", Findings: 2}, {Rule: "adamic/single-spread"}},
					Skipped: []string{"adamic/nominal-class"},
				},
				Findings: []program.LintCacheFinding{
					{
						RuleName: "no-debugger", Start: 0, End: 9,
						MessageId:          "unexpectedDebugger",
						MessageDescription: "Unexpected debugger statement.",
						FixCount:           1,
					},
					{
						RuleName: "no-empty", Start: 10, End: 12,
						MessageId: "emptyBlock",
						// Interpolated text, the case the id could never have rebuilt: this is the
						// shape of the 11 rules that build a description with fmt.Sprintf.
						MessageDescription: "Argument of unary negation should be assignable to number | bigint but is string instead.",
						SuggestionCount:    2,
					},
				},
			},
			{
				// A clean file is the case worth caching: most files produce nothing, and an
				// empty finding list is a real answer rather than an absence.
				Path:         "/project/source/clean.ts",
				ContentHash:  program.HashContent("export const value = 1;\n"),
				Rules:        []string{"no-debugger", "no-empty"},
				VisitedNodes: 8,
				Findings:     nil,
				// Measured with nothing to count: empty, not nil, and a round trip that turned it nil would make a
				// readiness run miss this file forever.
				Adamic: &program.AdamicRecord{},
				// A file whose directive silenced a cacheable rule's finding: the entry keeps which, so a replay
				// marks the directive applied.
				Directives: true,
				Withheld:   []program.LintCacheWithheld{{RuleName: "no-empty", Directive: 1}},
			},
		},
	}
}

// testIdentity is the build a test's tables are written and read by.
var testIdentity = program.CacheTableIdentity{SelfCommit: "test", CompilerCommit: "test", GoToolchain: "go-test", Platform: "test/test"}

// roundTripLintCache sends a findings cache through the cache table's file and back, the only way one
// reaches disk.
func roundTripLintCache(t *testing.T, cache *program.LintCache) *program.LintCache {
	t.Helper()
	table := program.NewCacheTable()
	table.Findings = cache
	directory := t.TempDir()
	if err := program.WriteCacheTable(directory, table, testIdentity, program.CacheTableSections{Findings: true}, program.PathAnchor{}); err != nil {
		t.Fatalf("writing: %v", err)
	}
	read, err := program.ReadCacheTable(directory, testIdentity, program.CacheTableSections{Findings: true}, program.PathAnchor{})
	if err != nil {
		t.Fatalf("reading what we just wrote: %v", err)
	}
	if read.Findings == nil {
		t.Fatal("the findings section came back absent")
	}
	return read.Findings
}

// TestLintCacheRoundTripsEveryField compares every field rather than spot-checking, because a field
// the encoder forgets decodes to empty while everything around it looks correct. The format was once
// raw offsets into a blob, where that was easy; it is JSON now, where a missing tag or a lowercase
// field does the same thing silently. I shipped exactly that bug twice while writing the
// resolution cache, and once more here before this test existed. The encoding is gob now, which cannot
// forget a field it was given, but the interning of rule lists is still written by hand.
//
// It then happened a fourth time, in the way this test could not see. The comparisons below were
// written out by hand, so `MessageDescription` was added to the struct and this test kept passing
// without ever looking at it — a test named for every field, checking the fields someone
// remembered. TestLintCacheFindingHasNoUncheckedFields now makes that structural: it fails when a
// field exists that the comparison does not name, so the next field cannot be added silently.
func TestLintCacheRoundTripsEveryField(t *testing.T) {
	t.Parallel()
	original := sampleLintCache()
	decoded := roundTripLintCache(t, original)

	if decoded.Key != original.Key {
		t.Errorf("key did not survive: %x against %x", decoded.Key, original.Key)
	}
	if !reflect.DeepEqual(decoded.DesignSystem, original.DesignSystem) || original.DesignSystem == nil {
		t.Errorf("the design system key did not survive: %+v against %+v", decoded.DesignSystem, original.DesignSystem)
	}
	if len(decoded.Entries) != len(original.Entries) {
		t.Fatalf("entries: %d back from %d", len(decoded.Entries), len(original.Entries))
	}

	for index, want := range original.Entries {
		got := decoded.Entries[index]
		if got.Path != want.Path {
			t.Errorf("entry %d Path: %q against %q", index, got.Path, want.Path)
		}
		if got.ContentHash != want.ContentHash {
			t.Errorf("entry %d ContentHash: %x against %x", index, got.ContentHash, want.ContentHash)
		}
		if strings.Join(got.Rules, ",") != strings.Join(want.Rules, ",") {
			t.Errorf("entry %d Rules: %v against %v", index, got.Rules, want.Rules)
		}
		if strings.Join(got.TypedRules, ",") != strings.Join(want.TypedRules, ",") {
			t.Errorf("entry %d TypedRules: %v against %v", index, got.TypedRules, want.TypedRules)
		}
		if got.TypeFingerprint != want.TypeFingerprint {
			t.Errorf("entry %d TypeFingerprint: %x against %x", index, got.TypeFingerprint, want.TypeFingerprint)
		}
		if strings.Join(got.ShapedRules, ",") != strings.Join(want.ShapedRules, ",") {
			t.Errorf("entry %d ShapedRules: %v against %v", index, got.ShapedRules, want.ShapedRules)
		}
		if got.ShapeFingerprint != want.ShapeFingerprint {
			t.Errorf("entry %d ShapeFingerprint: %x against %x", index, got.ShapeFingerprint, want.ShapeFingerprint)
		}
		if strings.Join(got.DesignRules, ",") != strings.Join(want.DesignRules, ",") {
			t.Errorf("entry %d DesignRules: %v against %v", index, got.DesignRules, want.DesignRules)
		}
		if strings.Join(got.DerivedRules, ",") != strings.Join(want.DerivedRules, ",") {
			t.Errorf("entry %d DerivedRules: %v against %v", index, got.DerivedRules, want.DerivedRules)
		}
		if got.DerivedFingerprint != want.DerivedFingerprint {
			t.Errorf("entry %d DerivedFingerprint: %x against %x", index, got.DerivedFingerprint, want.DerivedFingerprint)
		}
		if got.DesignFingerprint != want.DesignFingerprint {
			t.Errorf("entry %d DesignFingerprint: %x against %x", index, got.DesignFingerprint, want.DesignFingerprint)
		}
		if strings.Join(got.Listening, ",") != strings.Join(want.Listening, ",") {
			t.Errorf("entry %d Listening: %v against %v", index, got.Listening, want.Listening)
		}
		if got.VisitedNodes != want.VisitedNodes {
			t.Errorf("entry %d VisitedNodes: %d against %d", index, got.VisitedNodes, want.VisitedNodes)
		}
		if !reflect.DeepEqual(got.Notes, want.Notes) {
			t.Errorf("entry %d Notes: %v against %v", index, got.Notes, want.Notes)
		}
		if !reflect.DeepEqual(got.Adamic, want.Adamic) || want.Adamic == nil {
			t.Errorf("entry %d Adamic: %+v against %+v", index, got.Adamic, want.Adamic)
		}
		if got.Directives != want.Directives {
			t.Errorf("entry %d Directives: %t against %t", index, got.Directives, want.Directives)
		}
		if !reflect.DeepEqual(got.Withheld, want.Withheld) {
			t.Errorf("entry %d Withheld: %+v against %+v", index, got.Withheld, want.Withheld)
		}
		if len(got.Findings) != len(want.Findings) {
			t.Fatalf("entry %d findings: %d back from %d", index, len(got.Findings), len(want.Findings))
		}
		for findingIndex, wantFinding := range want.Findings {
			gotFinding := got.Findings[findingIndex]
			if gotFinding.RuleName != wantFinding.RuleName {
				t.Errorf("entry %d finding %d RuleName: %q against %q",
					index, findingIndex, gotFinding.RuleName, wantFinding.RuleName)
			}
			if gotFinding.Start != wantFinding.Start || gotFinding.End != wantFinding.End {
				t.Errorf("entry %d finding %d range: %d-%d against %d-%d",
					index, findingIndex, gotFinding.Start, gotFinding.End, wantFinding.Start, wantFinding.End)
			}
			if gotFinding.MessageId != wantFinding.MessageId {
				t.Errorf("entry %d finding %d MessageId: %q against %q",
					index, findingIndex, gotFinding.MessageId, wantFinding.MessageId)
			}
			if gotFinding.MessageDescription != wantFinding.MessageDescription {
				t.Errorf("entry %d finding %d MessageDescription: %q against %q",
					index, findingIndex, gotFinding.MessageDescription, wantFinding.MessageDescription)
			}
			if gotFinding.FixCount != wantFinding.FixCount {
				t.Errorf("entry %d finding %d FixCount: %d against %d",
					index, findingIndex, gotFinding.FixCount, wantFinding.FixCount)
			}
			if gotFinding.SuggestionCount != wantFinding.SuggestionCount {
				t.Errorf("entry %d finding %d SuggestionCount: %d against %d",
					index, findingIndex, gotFinding.SuggestionCount, wantFinding.SuggestionCount)
			}
		}
	}
}

// TestLintCacheLookupDistinguishesCleanFromUnknown is the correctness heart of this cache.
//
// A clean file and an unknown file both have zero findings. If Lookup cannot tell them apart,
// every uncached file reads as clean and the tool reports a green tree it never linted. That
// is the exact silent failure this whole domain exists to prevent, and it is one boolean away.
func TestLintCacheLookupDistinguishesCleanFromUnknown(t *testing.T) {
	t.Parallel()
	cache := roundTripLintCache(t, sampleLintCache())
	key := program.HashRuleSet([]string{"no-debugger", "no-empty"})
	rules := []string{"no-debugger", "no-empty"}

	// Not parallel: Lookup builds the shared cache's path index on its first call with the matching key,
	// lazily and unlocked, so two subtests looking up at once would race to build it.
	t.Run("a cached clean file is a hit with no findings", func(t *testing.T) {
		entry, hit := cache.Lookup("/project/source/clean.ts",
			program.HashContent("export const value = 1;\n"), key, rules)
		if !hit {
			t.Fatal("a file that was linted and found clean reported as a miss, so its result " +
				"is recomputed forever and the cache never pays for the common case")
		}
		if len(entry.Findings) != 0 {
			t.Fatalf("expected no findings, got %d", len(entry.Findings))
		}
	})

	// Not parallel: Lookup builds the shared cache's path index on its first call with the matching key,
	// lazily and unlocked, so two subtests looking up at once would race to build it.
	t.Run("an unknown file is a miss, not a clean hit", func(t *testing.T) {
		_, hit := cache.Lookup("/project/source/never-seen.ts",
			program.HashContent("whatever"), key, rules)
		if hit {
			t.Fatal("a file the cache has never seen reported as a hit; every unlinted file " +
				"would read as clean and the tool would report a green tree it never checked")
		}
	})

	// Not parallel: Lookup builds the shared cache's path index on its first call with the matching key,
	// lazily and unlocked, so two subtests looking up at once would race to build it.
	t.Run("changed contents under a known path is a miss", func(t *testing.T) {
		_, hit := cache.Lookup("/project/source/dirty.ts",
			program.HashContent("something completely different\n"), key, rules)
		if hit {
			t.Fatal("an edited file matched its stale entry by path; the cache would replay " +
				"yesterday's findings for today's code")
		}
	})

	t.Run("a changed key invalidates every entry", func(t *testing.T) {
		t.Parallel()
		_, hit := cache.Lookup("/project/source/clean.ts",
			program.HashContent("export const value = 1;\n"),
			program.HashRuleSet([]string{"no-debugger"}), rules)
		if hit {
			t.Fatal("a changed key left the cache serving results produced under the old one; a new " +
				"binary, config or rule set must invalidate everything at once")
		}
	})

	// An override can change which rules reach one file while the key, built from the whole config's
	// bytes, would also change. This guards the narrower case of the rule list itself differing, so a
	// file never replays findings from a rule set it is not offered now.
	// Not parallel: Lookup builds the shared cache's path index on its first call with the matching key,
	// lazily and unlocked, so two subtests looking up at once would race to build it.
	t.Run("a different applied rule list is a miss", func(t *testing.T) {
		_, hit := cache.Lookup("/project/source/clean.ts",
			program.HashContent("export const value = 1;\n"), key, []string{"no-debugger"})
		if hit {
			t.Fatal("an entry recorded under two rules was served to a file offered one")
		}
	})

	t.Run("a nil cache is always a miss", func(t *testing.T) {
		t.Parallel()
		var nilCache *program.LintCache
		if _, hit := nilCache.Lookup("anything", program.HashContent("x"), key, rules); hit {
			t.Fatal("a nil cache reported a hit")
		}
	})
}

// TestRuleSetHashIsOrderSensitive pins a property that looks like pedantry and is not. A reordered
// rule set can change which findings exist, because a fix applied by an earlier rule changes the text
// a later rule reads. This doc once said rule order decides the order findings come back in; it does
// not, the walk sorts them, and HashRuleSet's own comment says so.
func TestRuleSetHashIsOrderSensitive(t *testing.T) {
	t.Parallel()
	forward := program.HashRuleSet([]string{"alpha", "beta"})
	backward := program.HashRuleSet([]string{"beta", "alpha"})
	if forward == backward {
		t.Fatal("the rule set hash ignores order, so a reordered configuration reads as the " +
			"same rule set and replays findings in an order it would not produce")
	}

	t.Run("concatenation cannot collide", func(t *testing.T) {
		t.Parallel()
		// Without a separator, {"ab","c"} and {"a","bc"} hash identically. That is a real
		// collision between two different rule sets, and it would serve one's cache to the
		// other.
		if program.HashRuleSet([]string{"ab", "c"}) == program.HashRuleSet([]string{"a", "bc"}) {
			t.Fatal("rule names are concatenated without a separator, so different rule sets " +
				"collide and one set's cache is served to another")
		}
	})
}

// TestLintCacheFindingHasNoUncheckedFields makes the round-trip comparison structural.
//
// TestLintCacheRoundTripsEveryField enumerates fields by hand, which means it checks the fields
// somebody remembered rather than the fields that exist. That is not hypothetical: MessageDescription
// was added to LintCacheFinding and that test kept passing, having never looked at it. A test named
// for every field silently stopped covering one.
//
// So this asserts the field set itself. Adding a field to LintCacheFinding fails here until it is
// named below AND compared above, which is the cheapest way to make "the encoder forgot a field"
// impossible to ship — the exact bug this format has produced four times.
func TestLintCacheFindingHasNoUncheckedFields(t *testing.T) {
	t.Parallel()
	// Every field the round-trip test compares. Kept in sync by this test failing, not by discipline.
	compared := map[string]struct{}{
		"RuleName":           {},
		"Start":              {},
		"End":                {},
		"MessageId":          {},
		"MessageDescription": {},
		"FixCount":           {},
		"SuggestionCount":    {},
	}

	findingType := reflect.TypeOf(program.LintCacheFinding{})
	if findingType.NumField() == 0 {
		t.Fatal("LintCacheFinding has no fields, so this check is looking at the wrong type")
	}

	for index := range findingType.NumField() {
		name := findingType.Field(index).Name
		if _, checked := compared[name]; !checked {
			t.Errorf("LintCacheFinding.%s is not compared by TestLintCacheRoundTripsEveryField: "+
				"add it there and name it here, or an encoder that drops it decodes to empty while "+
				"every field around it looks correct", name)
		}
		delete(compared, name)
	}

	// A name left over means the list drifted the other way: this check would keep passing while
	// pointing at a field that no longer exists, which makes it decorative.
	for name := range compared {
		t.Errorf("this check names %q but LintCacheFinding has no such field, so the list is stale", name)
	}
}

// TestLintCacheEntryHasNoUncheckedFields is the same structural guard over the entry, which gained
// three fields at once when coverage started being replayed with findings: Rules, Listening and
// VisitedNodes. An entry field the round trip does not compare can be dropped by the encoder and still
// pass, and a dropped VisitedNodes would change the coverage line of every replayed run.
func TestLintCacheEntryHasNoUncheckedFields(t *testing.T) {
	t.Parallel()
	compared := map[string]struct{}{
		"Path": {}, "ContentHash": {}, "Rules": {}, "TypedRules": {}, "TypeFingerprint": {}, "Listening": {},
		"VisitedNodes": {}, "Findings": {}, "ShapedRules": {}, "ShapeFingerprint": {}, "DesignRules": {}, "DesignFingerprint": {},
		"DerivedRules": {}, "DerivedFingerprint": {}, "Notes": {}, "Adamic": {}, "Directives": {}, "Withheld": {},
	}
	entryType := reflect.TypeOf(program.LintCacheEntry{})
	for index := range entryType.NumField() {
		name := entryType.Field(index).Name
		if _, checked := compared[name]; !checked {
			t.Errorf("LintCacheEntry.%s is not compared by TestLintCacheRoundTripsEveryField", name)
		}
		delete(compared, name)
	}
	for name := range compared {
		t.Errorf("this check names %q but LintCacheEntry has no such field, so the list is stale", name)
	}
}

// TestLintCacheStoreReplacesRatherThanAppends pins the property Lookup depends on.
//
// A path stored twice must leave one entry, not two. Two entries make Lookup's answer depend on
// which copy the index happened to keep, and the stale one would win or lose by insertion order,
// which is exactly the kind of nondeterminism that reads as a flaky cache rather than a bug.
func TestLintCacheStoreReplacesRatherThanAppends(t *testing.T) {
	t.Parallel()
	cache := &program.LintCache{Key: program.HashRuleSet([]string{"no-debugger"})}
	path := "/project/source/edited.ts"
	rules := []string{"no-debugger"}

	first := program.HashContent("debugger;\n")
	cache.Store(program.LintCacheEntry{Path: path, ContentHash: first, Rules: rules, Findings: []program.LintCacheFinding{{
		RuleName: "no-debugger", Start: 0, End: 9,
		MessageId: "unexpectedDebugger", MessageDescription: "Unexpected debugger statement.",
	}}})

	// The same file, edited clean. The second store must supersede the first.
	second := program.HashContent("export const value = 1;\n")
	cache.Store(program.LintCacheEntry{Path: path, ContentHash: second, Rules: rules})

	if len(cache.Entries) != 1 {
		t.Fatalf("entries: %d, so Store appended instead of replacing", len(cache.Entries))
	}

	// The superseded content hash must no longer be a hit, or the cache would serve findings for
	// bytes that are gone.
	if _, hit := cache.Lookup(path, first, cache.Key, rules); hit {
		t.Error("the replaced content hash is still a hit, so stale findings survived a Store")
	}

	entry, hit := cache.Lookup(path, second, cache.Key, rules)
	if !hit {
		t.Fatal("the current content hash is a miss, so the replacement did not take")
	}
	if len(entry.Findings) != 0 {
		t.Errorf("findings: %d, want 0 for the cleaned file", len(entry.Findings))
	}
}

// TestLintCacheStoreRecordsCleanFiles is the case worth caching, and the one easiest to skip.
//
// Most files produce nothing. If Store dropped empty results, every clean file would be a miss
// forever and the cache would save nothing on exactly the population it exists for, while looking
// like it worked.
func TestLintCacheStoreRecordsCleanFiles(t *testing.T) {
	t.Parallel()
	cache := &program.LintCache{Key: program.HashRuleSet([]string{"no-debugger"})}
	contentHash := program.HashContent("export const value = 1;\n")
	rules := []string{"no-debugger"}
	cache.Store(program.LintCacheEntry{Path: "/project/source/clean.ts", ContentHash: contentHash, Rules: rules})

	entry, hit := cache.Lookup("/project/source/clean.ts", contentHash, cache.Key, rules)
	if !hit {
		t.Fatal("a stored clean file is a miss, so clean files would never be cached")
	}
	if len(entry.Findings) != 0 {
		t.Errorf("findings: %d, want 0", len(entry.Findings))
	}
}

// TestCacheableRulesExcludesImpureRules pins the split the cache's correctness rests on.
//
// A rule is cacheable on content alone only if its findings are a function of the linted file's
// bytes and of what the key already holds, which includes the tsconfig chain, so a rule reading only
// the compiler options is. Every other declaration below describes a rule whose answer can change
// while that hash does not, and replaying such a rule serves zero findings forever on a file that now
// has one.
func TestCacheableRulesExcludesImpureRules(t *testing.T) {
	t.Parallel()
	pure := rule.Rule{Name: "pure"}
	options := rule.Rule{Name: "options", ProgramReads: rule.ReadsCompilerOptions}
	otherFiles := rule.Rule{Name: "other-files", ProgramReads: rule.ReadsOtherFiles}
	resolution := rule.Rule{Name: "resolution", ProgramReads: rule.ReadsModuleResolution}
	needsChecker := rule.Rule{Name: "needs-checker", NeedsTypeChecker: true}
	both := rule.Rule{Name: "both", ProgramReads: rule.ReadsOtherFiles, NeedsTypeChecker: true}

	cacheable, uncacheable := program.CacheableRules(
		[]rule.Rule{pure, options, otherFiles, resolution, needsChecker, both})

	names := make([]string, 0, len(cacheable))
	for _, subject := range cacheable {
		names = append(names, subject.Name)
	}
	if strings.Join(names, ",") != "pure,options" {
		t.Errorf("cacheable = %v, want exactly [pure options]: anything reading outside its own file "+
			"and the key can have its answer changed by an edit the content hash cannot see", names)
	}
	if len(uncacheable) != 4 {
		t.Errorf("uncacheable = %d, want 4", len(uncacheable))
	}
}

// TestCacheableRulesDefaultsToExcluding is the asymmetry, stated as a test rather than a comment.
//
// Including a rule wrongly serves stale findings silently and forever; excluding one wrongly costs
// a cache miss. Those are not comparable, so a rule carrying any impurity declaration is excluded
// even when it also looks pure by every other measure.
func TestCacheableRulesDefaultsToExcluding(t *testing.T) {
	t.Parallel()
	// A rule that declares an impurity and nothing else must still be turned away, so a future
	// declaration added to rule.Rule cannot quietly become cacheable by omission here.
	for _, subject := range []rule.Rule{
		{Name: "reads-other-files", ProgramReads: rule.ReadsOtherFiles},
		{Name: "reads-module-resolution", ProgramReads: rule.ReadsModuleResolution},
		{Name: "needs-checker", NeedsTypeChecker: true},
	} {
		cacheable, _ := program.CacheableRules([]rule.Rule{subject})
		if len(cacheable) != 0 {
			t.Errorf("%s was admitted to the cache despite declaring an impurity", subject.Name)
		}
	}
}

// TestCacheableRulesHandlesAnEmptySet pins that no rules means no cacheable rules.
//
// An empty input returning a nil cacheable set matters because the caller uses emptiness to decide
// whether to consult the cache at all, and a non-nil empty slice and a nil one must behave the same
// at that decision.
func TestCacheableRulesHandlesAnEmptySet(t *testing.T) {
	t.Parallel()
	cacheable, uncacheable := program.CacheableRules(nil)
	if len(cacheable) != 0 || len(uncacheable) != 0 {
		t.Errorf("empty input produced %d cacheable and %d uncacheable", len(cacheable), len(uncacheable))
	}
}
