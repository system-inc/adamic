package docsdata

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/cohere/internal/benchresults"
	"github.com/system-inc/cohere/internal/docsdata/capture"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// moduleRoot is the directory the generated paths are relative to: go test runs in the package's own
// directory, two below it.
const moduleRoot = "../.."

func sourceInputs(t *testing.T) Inputs {
	t.Helper()
	inputs, err := SourceInputs(moduleRoot)
	if err != nil {
		t.Fatal(err)
	}
	return inputs
}

func built(t *testing.T, inputs Inputs) map[string][]byte {
	t.Helper()
	files, err := Build(inputs)
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func staleAgainstCommitted(t *testing.T, files map[string][]byte) []string {
	t.Helper()
	stale, err := Stale(moduleRoot, files)
	if err != nil {
		t.Fatal(err)
	}
	return stale
}

// TestTheRenderedFilesAreCurrent: docs/data is what Build renders now from the committed examples, every
// file but cli.json, which needs the binary built. It is -check-rendering's depth: the examples themselves
// are recaptured only by the generator's -check, which takes minutes and runs in the release's module check.
func TestTheRenderedFilesAreCurrent(t *testing.T) {
	t.Parallel()
	if stale := staleAgainstCommitted(t, built(t, sourceInputs(t))); len(stale) > 0 {
		t.Errorf("stale: %v; run go run ./internal/docsdata/tools/generate", stale)
	}
}

// TestCheckSeesARegistryChange is -check's contract for the registry, in both directions: the committed
// files pass as they are, and fail once a rule is removed or added.
func TestCheckSeesARegistryChange(t *testing.T) {
	t.Parallel()
	inputs := sourceInputs(t)
	if stale := staleAgainstCommitted(t, built(t, inputs)); len(stale) > 0 {
		t.Fatalf("the committed files are already stale (%v), so a change could not be told apart", stale)
	}

	removed := inputs
	removed.Registrations = inputs.Registrations[1:]
	if stale := staleAgainstCommitted(t, built(t, removed)); !slices.Contains(stale, RulesPath) {
		t.Errorf("removing %s from the registry left rules.json current: %v", inputs.Registrations[0].Rule.Name, stale)
	}

	added := inputs
	added.Registrations = append(slices.Clone(inputs.Registrations), rule.Registration{
		Rule: rule.Rule{Name: "nexus/consistency-no-docsdata-test-rule", Run: func(rule.Context, any) rule.Listeners { return nil }},
	})
	if stale := staleAgainstCommitted(t, built(t, added)); !slices.Contains(stale, RulesPath) {
		t.Errorf("adding a rule to the registry left rules.json current: %v", stale)
	}
}

// TestCheckSeesATierChange: a rule set resolving differently makes sets.json and rules.json stale.
func TestCheckSeesATierChange(t *testing.T) {
	t.Parallel()
	inputs := sourceInputs(t)
	changedSet := inputs.SetNames[0]
	var changedRule string
	loadSet := inputs.LoadSet
	inputs.LoadSet = func(name string) (*configuration.Config, error) {
		resolved, err := loadSet(name)
		if err != nil || name != changedSet {
			return resolved, err
		}
		names := make([]string, 0, len(resolved.Rules))
		for ruleName := range resolved.Rules {
			names = append(names, ruleName)
		}
		sort.Strings(names)
		changedRule = names[0]
		setting := resolved.Rules[changedRule]
		setting.Severity = configuration.SeverityError
		if resolved.Rules[changedRule].Severity == configuration.SeverityError {
			setting.Severity = configuration.SeverityWarn
		}
		resolved.Rules[changedRule] = setting
		return resolved, nil
	}
	stale := staleAgainstCommitted(t, built(t, inputs))
	for _, path := range []string{SetsPath, RulesPath} {
		if !slices.Contains(stale, path) {
			t.Errorf("changing %s's severity in %s left %s current: %v", changedRule, changedSet, path, stale)
		}
	}
}

// TestEveryRegisteredRuleHasOneRow: rules.json names every rule each engine registers exactly once, under
// that engine's language, and nothing else. The Swift registry is read from swift/Rules.json, which a
// Swift test holds to the Swift engine's registry.
func TestEveryRegisteredRuleHasOneRow(t *testing.T) {
	t.Parallel()
	committed, err := os.ReadFile(filepath.Join(moduleRoot, RulesPath))
	if err != nil {
		t.Fatal(err)
	}
	var rules Rules
	if err := json.Unmarshal(committed, &rules); err != nil {
		t.Fatal(err)
	}
	swiftRegistry, err := os.ReadFile(filepath.Join(moduleRoot, "swift", "Rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	var swiftEntries []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(swiftRegistry, &swiftEntries); err != nil {
		t.Fatal(err)
	}
	registered := map[string][]string{"TypeScript": registry.Names()}
	for _, entry := range swiftEntries {
		registered["Swift"] = append(registered["Swift"], entry.Name)
	}

	rows := map[string]map[string]int{}
	for _, row := range rules.Rules {
		if rows[row.Language] == nil {
			rows[row.Language] = map[string]int{}
		}
		rows[row.Language][row.Name]++
	}
	for _, language := range []string{"TypeScript", "Swift"} {
		if len(registered[language]) == 0 {
			t.Fatalf("no %s rule is registered, so the rows were checked against nothing", language)
		}
		names := map[string]bool{}
		for _, name := range registered[language] {
			names[name] = true
			if rows[language][name] != 1 {
				t.Errorf("%s is a registered %s rule and has %d %s rows in rules.json", name, language, rows[language][name], language)
			}
		}
		for name := range rows[language] {
			if !names[name] {
				t.Errorf("rules.json has a %s row for %s, which that engine does not register", language, name)
			}
		}
	}
	for language := range rows {
		if registered[language] == nil {
			t.Errorf("rules.json has rows in %s, which no engine registers", language)
		}
	}
	if rules.Count != len(rules.Rules) {
		t.Errorf("rules.json says count %d over %d rows", rules.Count, len(rules.Rules))
	}
}

// TestCheckSeesASwiftRegistryChange: a Swift rule removed from swift/Rules.json makes rules.json stale, and
// a Swift rule with an unknown origin, a misnamed house rule, or a verdict naming a missing Swift rule
// refuses the build.
func TestCheckSeesASwiftRegistryChange(t *testing.T) {
	t.Parallel()
	inputs := sourceInputs(t)
	if stale := staleAgainstCommitted(t, built(t, inputs)); len(stale) > 0 {
		t.Fatalf("the committed files are already stale (%v), so a change could not be told apart", stale)
	}
	var entries []map[string]any
	if err := json.Unmarshal(inputs.SwiftRules, &entries); err != nil {
		t.Fatal(err)
	}
	withEntries := func(change func([]map[string]any) []map[string]any) Inputs {
		copied := make([]map[string]any, 0, len(entries))
		for _, entry := range entries {
			clone := map[string]any{}
			for key, value := range entry {
				clone[key] = value
			}
			copied = append(copied, clone)
		}
		encoded, err := json.Marshal(change(copied))
		if err != nil {
			t.Fatal(err)
		}
		changed := inputs
		changed.SwiftRules = encoded
		return changed
	}

	removed := withEntries(func(entries []map[string]any) []map[string]any {
		// The last row, so a rule no verdict needs is the likelier one removed; a verdict-carrying rule
		// would refuse the build instead, which is the case below.
		for index := len(entries) - 1; index >= 0; index-- {
			if entries[index]["origin"] != houseOrigin {
				return append(entries[:index], entries[index+1:]...)
			}
		}
		t.Fatal("no ported Swift rule to remove")
		return nil
	})
	if stale := staleAgainstCommitted(t, built(t, removed)); !slices.Contains(stale, RulesPath) {
		t.Errorf("removing a Swift rule left rules.json current: %v", stale)
	}

	for name, change := range map[string]func([]map[string]any) []map[string]any{
		"an unknown origin": func(entries []map[string]any) []map[string]any {
			entries[0]["origin"] = "swiftlint-ish"
			return entries
		},
		"a house rule outside the scheme": func(entries []map[string]any) []map[string]any {
			return append(entries, map[string]any{"name": swiftNamespace + "/docsdata-test-rule", "origin": houseOrigin, "typeAware": false})
		},
		"a verdict naming a missing Swift rule": func(entries []map[string]any) []map[string]any {
			for index, entry := range entries {
				if entry["name"] == swiftNamespace+"/consistency-no-bare-throw" {
					return append(entries[:index], entries[index+1:]...)
				}
			}
			t.Fatal("the bare-throw rule, which a verdict names, is not in swift/Rules.json")
			return nil
		},
	} {
		if _, err := Build(withEntries(change)); err == nil {
			t.Errorf("%s built", name)
		}
	}
}

// TestEveryHouseRuleFollowsTheNamingScheme: every house rule builds with a category and no port carries
// one, and a house rule registered under a name outside the scheme refuses the build.
func TestEveryHouseRuleFollowsTheNamingScheme(t *testing.T) {
	t.Parallel()
	inputs := sourceInputs(t)
	var rules Rules
	if err := json.Unmarshal(built(t, inputs)[RulesPath], &rules); err != nil {
		t.Fatal(err)
	}
	house := 0
	for _, row := range rules.Rules {
		switch {
		case row.Origin == houseOrigin && row.Category == "":
			t.Errorf("%s is a house rule with no category", row.Name)
		case row.Origin != houseOrigin && row.Category != "":
			t.Errorf("%s is a port and carries the category %q", row.Name, row.Category)
		}
		if row.Origin == houseOrigin {
			house++
		}
	}
	if house == 0 {
		t.Fatal("no house rule was built, so the scheme was checked against nothing")
	}

	misnamed := inputs
	misnamed.Registrations = append(slices.Clone(inputs.Registrations), rule.Registration{
		Rule: rule.Rule{Name: "nexus/docsdata-test-rule", Run: func(rule.Context, any) rule.Listeners { return nil }},
	})
	if _, err := Build(misnamed); err == nil || !strings.Contains(err.Error(), "nexus/docsdata-test-rule") {
		t.Errorf("a house rule named outside the scheme built: %v", err)
	}
}

// TestBuildIsDeterministic: two builds give the same bytes, so a regeneration diffs to nothing.
func TestBuildIsDeterministic(t *testing.T) {
	t.Parallel()
	inputs := sourceInputs(t)
	first, second := built(t, inputs), built(t, inputs)
	if len(first) != len(second) {
		t.Fatalf("%d files, then %d", len(first), len(second))
	}
	for path, contents := range first {
		if !bytes.Equal(contents, second[path]) {
			t.Errorf("%s differs between two builds", path)
		}
	}
}

// TestEveryFieldHasASourceNote: rules.json says where each field of a row comes from, and names no field
// a row does not have.
func TestEveryFieldHasASourceNote(t *testing.T) {
	t.Parallel()
	encoded, err := json.Marshal(RuleRow{Options: &RuleOptions{}, Swift: &SwiftVerdict{}, Sets: []RuleSetSeverity{{}}, MessageIds: []string{""},
		Namespace: "-", UpstreamName: "-", Category: "-", FixKind: "-", TypeScriptRules: []string{""}})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	for field := range fields {
		if ruleSourceNotes[field] == "" {
			t.Errorf("rules.json rows carry %q and sourceNotes does not say where it comes from", field)
		}
	}
	for field := range ruleSourceNotes {
		if _, present := fields[field]; !present {
			t.Errorf("sourceNotes describes %q, which a row does not carry", field)
		}
	}
}

// TestTheHelpParserReadsWhatTheFlagPackageWrites: a flag set with every shape the binary uses, printed by
// the flag package itself, parses back to exactly what the set holds.
func TestTheHelpParserReadsWhatTheFlagPackageWrites(t *testing.T) {
	t.Parallel()
	set := flag.NewFlagSet("cohere", flag.ContinueOnError)
	set.String("tsconfig", "tsconfig.json", "the tsconfig, quoted \"here\"")
	set.String("directory", "", "the root (default: the nearest tsconfig.json)")
	set.Bool("fix", false, "apply fixes only")
	set.Bool("on", true, "a boolean that defaults on")
	set.Int("fix-passes", 10, "how many passes")
	set.Duration("wait", 2*time.Second, "how long")
	set.String("profile", "", "write a profile to `file`, for pprof")
	set.String("explain", "", "a usage that runs\nonto a second line")

	var printed bytes.Buffer
	set.SetOutput(&printed)
	set.PrintDefaults()
	_, parsed, err := parseFlagHelp(printed.Bytes())
	if err != nil {
		t.Fatalf("%v\n%s", err, printed.String())
	}

	var want []CliFlag
	set.VisitAll(func(each *flag.Flag) {
		argument, usage := flag.UnquoteUsage(each)
		entry := CliFlag{Name: each.Name, Argument: argument, Usage: usage}
		if !map[string]bool{"": true, "false": true, "0": true, "0s": true}[each.DefValue] {
			entry.Default = each.DefValue
		}
		want = append(want, entry)
	})
	if len(parsed) != len(want) {
		t.Fatalf("parsed %d flags, the set holds %d:\n%s", len(parsed), len(want), printed.String())
	}
	for index := range want {
		if parsed[index] != want[index] {
			t.Errorf("flag %d: parsed %+v, want %+v", index, parsed[index], want[index])
		}
	}
}

// TestTheVerbHelpKeepsItsSynopsis: a verb's usage lines and summary are told apart from its flags.
func TestTheVerbHelpKeepsItsSynopsis(t *testing.T) {
	t.Parallel()
	help := "usage: cohere rename <file>:<line>:<column> <newName>\n" +
		"       cohere rename <name> <newName>\n\n" +
		"Renames a symbol.\nNothing is written without --write.\n\n" +
		"  -write\n    \tapply the rename\n"
	cli, err := parseCli([]byte("  -fix\n    \tapply fixes only\n"), map[string][]byte{"rename": []byte(help)})
	if err != nil {
		t.Fatal(err)
	}
	verb := cli.Verbs[0]
	if strings.Join(verb.Usage, "|") != "cohere rename <file>:<line>:<column> <newName>|cohere rename <name> <newName>" {
		t.Errorf("usage %q", verb.Usage)
	}
	if verb.Summary != "Renames a symbol. Nothing is written without --write." {
		t.Errorf("summary %q", verb.Summary)
	}
	if len(verb.Flags) != 1 || verb.Flags[0].Name != "write" {
		t.Errorf("flags %+v", verb.Flags)
	}
}

// TestTheChangelogParsesEveryRelease: every release heading in CHANGELOG.md becomes one entry, and a
// second-level heading that is not a release refuses the parse.
func TestTheChangelogParsesEveryRelease(t *testing.T) {
	t.Parallel()
	changelog, err := os.ReadFile(filepath.Join(moduleRoot, "CHANGELOG.md"))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := parseChangelog(changelog)
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Count("\n"+string(changelog), "\n## "); len(entries) != want {
		t.Errorf("%d entries for %d release headings", len(entries), want)
	}
	for _, entry := range entries {
		if strings.TrimSpace(entry.Body) == "" {
			t.Errorf("%s has an empty body", entry.Version)
		}
	}
	if headingAnchor("1.0.0 (unreleased)") != "100-unreleased" {
		t.Errorf("anchor %q", headingAnchor("1.0.0 (unreleased)"))
	}
	if _, err := parseChangelog([]byte("# Changelog\n\n## Notes\n")); err == nil {
		t.Error("a second-level heading that is not a release parsed")
	}
}

// TestPickExamplesPrefersWhatAReaderCanRun: the shortest case run without options and without other
// files wins, a rule with an asserted fix shows its fixed source, and the message ids and fix kind cover
// every asserted case, those run with options included.
func TestPickExamplesPrefersWhatAReaderCanRun(t *testing.T) {
	t.Parallel()
	finding := func(id string, fix bool) []capture.Finding {
		return []capture.Finding{{Line: 1, Column: 1, MessageId: id, Fix: fix}}
	}
	picked, _ := PickExamples([]capture.Record{
		{Rule: "r", File: "a.ts", Source: "x", Options: json.RawMessage(`{}`), Outcome: capture.OutcomeFindings, Findings: finding("withOptions", false)},
		{Rule: "r", File: "a.ts", Source: "xx", OtherFiles: 1, Outcome: capture.OutcomeFindings, Findings: finding("other", false)},
		{Rule: "r", File: "a.ts", Source: "longer one", Outcome: capture.OutcomeFixed, Findings: finding("fixed", true), FixedSource: "fixed"},
		{Rule: "r", File: "a.ts", Source: "xyz", Outcome: capture.OutcomeFindings, Findings: finding("plain", false)},
		{Rule: "r", File: "b.ts", Source: "clean, longer", Outcome: capture.OutcomeClean},
		{Rule: "r", File: "a.ts", Source: "clean", Outcome: capture.OutcomeClean},
	})
	examples := picked["r"]
	if examples.Firing == nil || examples.Firing.FixedSource != "fixed" {
		t.Errorf("a rule with an asserted fix should show it: %+v", examples.Firing)
	}
	if examples.Clean == nil || examples.Clean.Source != "clean" {
		t.Errorf("clean %+v", examples.Clean)
	}
	if strings.Join(examples.MessageIds(), ",") != "fixed,other,plain,withOptions" || examples.FixKind() != "fix" {
		t.Errorf("ids %v, fix kind %q", examples.MessageIds(), examples.FixKind())
	}

	unfixed, _ := PickExamples([]capture.Record{
		{Rule: "r", File: "a.ts", Source: "x", Options: json.RawMessage(`{}`), Outcome: capture.OutcomeFindings},
		{Rule: "r", File: "a.ts", Source: "xx", OtherFiles: 1, Outcome: capture.OutcomeFindings},
		{Rule: "r", File: "a.ts", Source: "xyz", Outcome: capture.OutcomeFindings},
	})
	if unfixed["r"].Firing.Source != "xyz" {
		t.Errorf("the case a reader can run as shown should win: %+v", unfixed["r"].Firing)
	}
}

// TestACaseRunWithOptionsIsNeverShown: a rule whose only asserted cases ran with options has no example,
// and its message ids still count.
func TestACaseRunWithOptionsIsNeverShown(t *testing.T) {
	t.Parallel()
	picked, withOptions := PickExamples([]capture.Record{
		{Rule: "r", File: "a.ts", Source: "x", Options: json.RawMessage(`{"Mode": "TypeAnnotation"}`), Outcome: capture.OutcomeFindings,
			Findings: []capture.Finding{{MessageId: "only"}}},
		{Rule: "r", File: "a.ts", Source: "y", Options: json.RawMessage(`{"Mode": "TypeAnnotation"}`), Outcome: capture.OutcomeClean},
	})
	examples := picked["r"]
	if examples.Firing != nil || examples.Clean != nil {
		t.Errorf("a case run with options was shown: %+v", examples)
	}
	if strings.Join(examples.MessageIds(), ",") != "only" || withOptions != 2 {
		t.Errorf("ids %v, %d with options", examples.MessageIds(), withOptions)
	}
}

// TestPickExamplesCountsACaseOnce: one case asserted twice, as findings and then as its fixed source, or
// as findings by two tests, is counted once per outcome, and the fixed record is the one shown.
func TestPickExamplesCountsACaseOnce(t *testing.T) {
	t.Parallel()
	findings := capture.Record{Rule: "r", File: "a.ts", Source: "x", Options: json.RawMessage(`{"a":1}`), Outcome: capture.OutcomeFindings}
	_, withOptions := PickExamples([]capture.Record{findings, findings})
	if withOptions != 1 {
		t.Errorf("one case asserted twice counted %d times", withOptions)
	}

	asserted := capture.Record{Rule: "r", File: "a.ts", Source: "debugger;", Outcome: capture.OutcomeFindings,
		Findings: []capture.Finding{{MessageId: "debugger", Fix: true}}}
	fixed := asserted
	fixed.Outcome, fixed.FixedSource = capture.OutcomeFixed, "\n"
	picked, _ := PickExamples([]capture.Record{asserted, fixed, asserted})
	if picked["r"].Firing == nil || picked["r"].Firing.FixedSource != "\n" {
		t.Errorf("the fixed record should be the one shown: %+v", picked["r"].Firing)
	}
}

// TestBenchmarksHoldEveryRecordNewestFirst: benchmarks.json is every committed record, as recorded, newest
// first, with a later record placed ahead of the committed ones.
func TestBenchmarksHoldEveryRecordNewestFirst(t *testing.T) {
	t.Parallel()
	inputs := sourceInputs(t)
	committed, err := benchresults.Read(filepath.Join(moduleRoot, benchresults.Directory))
	if err != nil {
		t.Fatal(err)
	}
	if len(committed) == 0 {
		t.Fatal("bench/results holds no record, so this test would compare nothing")
	}
	later := committed[0]
	later.RecordedAt = "2999-01-01T00:00:00Z"
	inputs.Benchmarks = append(slices.Clone(committed), later)

	var benchmarks Benchmarks
	if err := json.Unmarshal(built(t, inputs)[BenchmarksPath], &benchmarks); err != nil {
		t.Fatal(err)
	}
	if len(benchmarks.Records) != len(committed)+1 {
		t.Fatalf("%d records published for %d given", len(benchmarks.Records), len(committed)+1)
	}
	if benchmarks.Records[0].RecordedAt != later.RecordedAt {
		t.Errorf("the newest record is not first: %s", benchmarks.Records[0].RecordedAt)
	}
	for index := 1; index < len(benchmarks.Records); index++ {
		if benchmarks.Records[index-1].RecordedAt < benchmarks.Records[index].RecordedAt {
			t.Errorf("records %d and %d are oldest first", index-1, index)
		}
	}
	if benchmarks.SourceNote == "" {
		t.Error("benchmarks.json says nothing about where its numbers come from")
	}
}

// TestAnUnearnedQuietNumberRefusesTheBuild: a record whose mode claims a quiet summary its runs did not earn
// is refused by Build, not published.
func TestAnUnearnedQuietNumberRefusesTheBuild(t *testing.T) {
	t.Parallel()
	inputs := sourceInputs(t)
	committed, err := benchresults.Read(filepath.Join(moduleRoot, benchresults.Directory))
	if err != nil {
		t.Fatal(err)
	}
	if len(committed) == 0 {
		t.Fatal("bench/results holds no record, so this test would plant nothing")
	}
	tampered := committed[0]
	tampered.Modes = map[benchresults.Mode]benchresults.ModeSummary{}
	for mode, summary := range committed[0].Modes {
		tampered.Modes[mode] = summary
	}
	cold := tampered.Modes[benchresults.ModeCold]
	cold.Quiet = &benchresults.QuietSummary{Count: 1, Best: 0.1, Median: 0.1, Worst: 0.1}
	tampered.Modes[benchresults.ModeCold] = cold
	inputs.Benchmarks = []benchresults.Record{tampered}
	if _, err := Build(inputs); err == nil {
		t.Error("a quiet number the runs did not earn was published")
	}
}
