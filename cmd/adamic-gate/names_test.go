package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"sort"
	"strings"
	"testing"
)

// Every test has one globally unique PascalCase identifier: its function name minus Test, with a numbered shard
// (TestFoo_0842) standing for its family Foo. The census refuses an identifier that lives in two or more packages,
// and one that says nothing about what it checks. TestProduct_ functions declare products, not tests, so they are
// exempt from both.
const (
	nameCensusCollides      = "lives in two or more packages"
	nameCensusUndescriptive = "is not descriptive"
	nameCensusStale         = "is pinned but no longer an offender"
)

// nameCensusGeneric are words that name no behavior. Each is also shorter than eight characters today; the list
// keeps them refused if the length floor ever moves.
var nameCensusGeneric = map[string]bool{"Works": true, "Basic": true, "Misc": true, "Simple": true, "Stuff": true, "Example": true, "Foo": true, "Bar": true, "Smoke": true, "Sanity": true, "All": true, "It": true}

// nameCensusPinned is the offenders on the day the census landed, sorted by identifier. It is an exact set that only
// shrinks: a new offender fails, and so does an entry that no longer offends.
var nameCensusPinned = []nameCensusEntry{
	{Identifier: "Ceiling", Rule: nameCensusUndescriptive, Packages: []string{"internal/childguard"}},
	{Identifier: "Child", Rule: nameCensusUndescriptive, Packages: []string{"internal/childguard"}},
	{Identifier: "CompileProfiles", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/lint/rules/no-underscore-dangle", "stage1/cohere/lint/rules/no-unsafe-negation", "stage1/cohere/lint/rules/no-unsafe-optional-chaining", "stage1/cohere/lint/rules/typescript-no-this-alias"}},
	{Identifier: "EachGapStandsWhereGapsMdSaysItDoes", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/css", "stage1/cohere/gitignore", "stage1/cohere/graphql", "stage1/cohere/mediaquery", "stage1/cohere/selector", "stage1/cohere/static_single_assignment", "stage1/cohere/suppression", "stage1/cohere/values"}},
	{Identifier: "JsxNode", Rule: nameCensusUndescriptive, Packages: []string{"stage1/typescript/parser"}},
	{Identifier: "Mutants", Rule: nameCensusUndescriptive, Packages: []string{"stage1/cohere/tsprinter"}},
	{Identifier: "MutantsUnion", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/lint", "stage1/cohere/tsprinter"}},
	{Identifier: "NoToken", Rule: nameCensusUndescriptive, Packages: []string{"cmd/adamic-gocacheprog"}},
	{Identifier: "Parse", Rule: nameCensusUndescriptive, Packages: []string{"internal/regexp"}},
	{Identifier: "Performance", Rule: nameCensusCollides, Packages: []string{"stage1/typescript/parser", "stage1/typescript/scanner"}},
	{Identifier: "ProfileArtifacts", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/lint", "stage1/typescript/scanner"}},
	{Identifier: "ProfileSnapshotsAgree", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/json", "stage1/cohere/lint", "stage1/typescript/scanner"}},
	{Identifier: "RuntimeFeaturesMatchNode", Rule: nameCensusCollides, Packages: []string{"cloud/reports/decode-ascii", "cmd/adamic-test262", "internal/fuzz"}},
	{Identifier: "Stalled", Rule: nameCensusUndescriptive, Packages: []string{"internal/childguard"}},
	{Identifier: "ThePortParsesAsGoCohereDoes", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/css", "stage1/cohere/formatfiles", "stage1/cohere/graphql", "stage1/cohere/mediaquery", "stage1/cohere/selector", "stage1/cohere/suppression", "stage1/cohere/values"}},
	{Identifier: "ThePortParsesAsGoCohereDoesUnion", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/css", "stage1/cohere/formatfiles"}},
	{Identifier: "Throughput", Rule: nameCensusCollides, Packages: []string{"stage1/cohere/estree", "stage1/cohere/lint"}},
	{Identifier: "ViewMixedUnionUnknownAndUnavailable", Rule: nameCensusCollides, Packages: []string{"internal/javascript", "internal/native"}},
}

type nameCensusTest struct {
	Package  string // directory from the repository root
	File     string // path from the repository root, for the report only
	Function string
}

type nameCensusEntry struct {
	Identifier string
	Rule       string
	Packages   []string
}

type nameCensusRefusal struct {
	Rule    string
	Message string
}

func TestEveryTestHasOneDescriptiveName(t *testing.T) {
	t.Parallel()
	root := parallelCheckRoot(t)
	var tests []nameCensusTest
	err := parallelCheckEachTest(root, func(rel string, fn *ast.FuncDecl, _ string, _ *token.FileSet, _ []string) {
		tests = append(tests, nameCensusTest{Package: path.Dir(rel), File: rel, Function: fn.Name.Name})
	})
	if err != nil {
		t.Fatal(err)
	}
	refusals := nameCensus(tests, nameCensusPinned)
	t.Logf("census of %d top-level tests; %d pinned offenders", len(tests), len(nameCensusPinned))
	for _, refusal := range refusals {
		t.Error(refusal.Message)
	}
}

// The census's pure core, fed names no tree holds: each offender is refused for its own rule, and the clean case,
// shards of one family in one package, and a pinned offender pass.
func TestNameCensusRefusesWhatItShould(t *testing.T) {
	t.Parallel()
	clean := []nameCensusTest{{Package: "stage1/a", Function: "TestParsesTemplateLiterals"}, {Package: "stage1/a", Function: "TestLowersOptionalChains_0001"}, {Package: "stage1/a", Function: "TestLowersOptionalChains_0002"}, {Package: "stage1/b", Function: "TestResolvesImportCycles"}}
	pin := nameCensusEntry{Identifier: "ResolvesImportCycles", Rule: nameCensusCollides, Packages: []string{"stage1/a", "stage1/b"}}
	cases := []struct {
		name   string
		tests  []nameCensusTest
		pinned []nameCensusEntry
		want   []string
		reason string
	}{
		{name: "clean", tests: clean},
		{name: "two packages", tests: append(clean, nameCensusTest{Package: "stage1/c", Function: "TestParsesTemplateLiterals"}), want: []string{nameCensusCollides}},
		{name: "shard family meets a plain test", tests: append(clean, nameCensusTest{Package: "stage1/c", Function: "TestLowersOptionalChains"}), want: []string{nameCensusCollides}},
		{name: "bare number", tests: append(clean, nameCensusTest{Package: "stage1/c", Function: "Test_0842"}), want: []string{nameCensusUndescriptive}, reason: "empty"},
		{name: "all digits", tests: append(clean, nameCensusTest{Package: "stage1/c", Function: "Test20260913"}), want: []string{nameCensusUndescriptive}, reason: "all digits"},
		{name: "generic word", tests: append(clean, nameCensusTest{Package: "stage1/c", Function: "TestSanity"}), want: []string{nameCensusUndescriptive}, reason: "generic word"},
		{name: "too short", tests: append(clean, nameCensusTest{Package: "stage1/c", Function: "TestParses_0003"}), want: []string{nameCensusUndescriptive}, reason: "shorter than 8"},
		{name: "stale pin", tests: clean, pinned: []nameCensusEntry{pin}, want: []string{nameCensusStale}, reason: "take it off the list"},
		{name: "pinned collision", tests: append(clean, nameCensusTest{Package: "stage1/a", Function: "TestResolvesImportCycles"}), pinned: []nameCensusEntry{pin}},
		{name: "product duplicate", tests: append(clean, nameCensusTest{Package: "stage1/a", Function: "TestProduct_Oracle"}, nameCensusTest{Package: "stage1/c", Function: "TestProduct_Oracle"})},
	}
	for _, c := range cases {
		refusals := nameCensus(c.tests, c.pinned)
		var rules []string
		for _, refusal := range refusals {
			rules = append(rules, refusal.Rule)
			if c.reason != "" && !strings.Contains(refusal.Message, c.reason) {
				t.Errorf("%s: refusal %q does not say %q", c.name, refusal.Message, c.reason)
			}
		}
		if strings.Join(rules, "|") != strings.Join(c.want, "|") {
			t.Errorf("%s: refused for %q, want %q", c.name, rules, c.want)
		}
	}
}

// nameCensusIdentifier is a test's identifier: its name minus Test and minus a trailing _ plus digits.
func nameCensusIdentifier(function string) string {
	identifier := strings.TrimPrefix(function, "Test")
	if cut := strings.LastIndexByte(identifier, '_'); cut >= 0 && cut < len(identifier)-1 && strings.Trim(identifier[cut+1:], "0123456789") == "" {
		identifier = identifier[:cut]
	}
	return identifier
}

// nameCensusUndescribed says why an identifier names nothing, or returns "" when it is descriptive.
func nameCensusUndescribed(identifier string) string {
	switch {
	case identifier == "":
		return "empty once Test and the shard number are stripped"
	case strings.Trim(identifier, "0123456789") == "":
		return "all digits"
	case nameCensusGeneric[identifier]:
		return "a generic word"
	case len(identifier) < 8:
		return "shorter than 8 characters"
	}
	return ""
}

// nameCensus is the census over (package, function) pairs: every offender not pinned and every pin that no longer
// offends, in identifier order.
func nameCensus(tests []nameCensusTest, pinned []nameCensusEntry) []nameCensusRefusal {
	byIdentifier := make(map[string][]nameCensusTest)
	for _, test := range tests {
		if strings.HasPrefix(test.Function, "TestProduct_") {
			continue
		}
		identifier := nameCensusIdentifier(test.Function)
		byIdentifier[identifier] = append(byIdentifier[identifier], test)
	}
	identifiers := make([]string, 0, len(byIdentifier))
	for identifier := range byIdentifier {
		identifiers = append(identifiers, identifier)
	}
	sort.Strings(identifiers)
	key := func(entry nameCensusEntry) string {
		return entry.Identifier + "\t" + entry.Rule + "\t" + strings.Join(entry.Packages, " ")
	}
	pins := make(map[string]bool)
	for _, entry := range pinned {
		pins[key(entry)] = true
	}
	offending := make(map[string]bool)
	var refusals []nameCensusRefusal
	for _, identifier := range identifiers {
		seenPackages, seenFiles := make(map[string]bool), make(map[string]bool)
		var packages, where []string
		for _, test := range byIdentifier[identifier] {
			if !seenPackages[test.Package] {
				seenPackages[test.Package] = true
				packages = append(packages, test.Package)
			}
			// One test per file names it; a shard family can hold hundreds.
			if !seenFiles[test.File] {
				seenFiles[test.File] = true
				where = append(where, fmt.Sprintf("%s (%s)", test.Function, test.File))
			}
		}
		sort.Strings(packages)
		var offenses []nameCensusEntry
		reasons := make(map[string]string)
		if len(packages) > 1 {
			offenses = append(offenses, nameCensusEntry{Identifier: identifier, Rule: nameCensusCollides, Packages: packages})
			reasons[nameCensusCollides] = strings.Join(packages, ", ")
		}
		if reason := nameCensusUndescribed(identifier); reason != "" {
			offenses = append(offenses, nameCensusEntry{Identifier: identifier, Rule: nameCensusUndescriptive, Packages: packages})
			reasons[nameCensusUndescriptive] = reason
		}
		for _, offense := range offenses {
			offending[key(offense)] = true
			if !pins[key(offense)] {
				refusals = append(refusals, nameCensusRefusal{Rule: offense.Rule, Message: fmt.Sprintf("%s: identifier %q %s (%s)", strings.Join(where, ", "), identifier, offense.Rule, reasons[offense.Rule])})
			}
		}
	}
	for _, entry := range pinned {
		if !offending[key(entry)] {
			refusals = append(refusals, nameCensusRefusal{Rule: nameCensusStale, Message: fmt.Sprintf("identifier %q (%s: %s) %s: renamed: take it off the list", entry.Identifier, entry.Rule, strings.Join(entry.Packages, ", "), nameCensusStale)})
		}
	}
	return refusals
}
