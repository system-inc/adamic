package program

import "testing"

// An entry whose withheld findings do not fit the file's directives is a miss, never repaired (#kdee854). The bytes
// are the entry's, so a mismatch means the entry is wrong, and replaying it would mark a directive the walk would
// not, or none. Each kind of doubt is a miss, and the entry that fits replays, so the misses are about the doubt.
func TestAnEntryWhoseWithheldFindingsDoNotFitIsAMiss(t *testing.T) {
	t.Parallel()
	text := "// cohere-disable-next-line no-var\nvar a = 1;\n// cohere-disable-next-line\nvar b = 2;\n"
	entry := func(withheld ...LintCacheWithheld) LintCacheEntry {
		return LintCacheEntry{
			Path:        "/project/a.ts",
			ContentHash: HashContent(text),
			Rules:       []string{"no-var", "no-debugger"},
			Directives:  true,
			Withheld:    withheld,
		}
	}
	lookup := func(subject LintCacheEntry) bool {
		t.Helper()
		reuse, keys := adamicFixtureReuse(t, subject)
		_, hits := reuse.lookup(subject.Path, keys, text)
		return hits.pure
	}
	if !lookup(entry(LintCacheWithheld{RuleName: "no-var", Directive: 0}, LintCacheWithheld{RuleName: "no-var", Directive: 1})) {
		t.Fatal("an entry whose withheld findings fit missed, so the misses below prove nothing")
	}
	for _, doubt := range []struct {
		name     string
		withheld LintCacheWithheld
	}{
		{"a directive index past the file's directives", LintCacheWithheld{RuleName: "no-var", Directive: 2}},
		{"a negative directive index", LintCacheWithheld{RuleName: "no-var", Directive: -1}},
		{"a directive that names another rule", LintCacheWithheld{RuleName: "no-debugger", Directive: 0}},
		{"a rule the entry did not run", LintCacheWithheld{RuleName: "no-empty", Directive: 1}},
	} {
		if lookup(entry(doubt.withheld)) {
			t.Errorf("an entry with %s replayed", doubt.name)
		}
	}
}
