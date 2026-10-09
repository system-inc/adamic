package optionschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// samplesFile is testdata/samples.json, written by internal/lint/tools/option_schemas/extract.mjs: option
// lists drawn from each upstream schema in both directions, each with ajv's verdict as ESLint 10
// configures ajv.
type samplesFile struct {
	GeneratedFrom map[string]string `json:"generatedFrom"`
	Samples       []struct {
		Rule     string            `json:"rule"`
		Elements []json.RawMessage `json:"elements"`
		Keyword  string            `json:"keyword"`
		Path     string            `json:"path"`
		ESLint   string            `json:"eslint"`
	} `json:"samples"`
}

// corpusDirectory is ESLint's own test rows for core, every one of which ESLint ran with its options.
const corpusDirectory = "../registry/testdata/eslint-corpus/eslint-10.8.1"

// The gate is two counts, each required to be zero (#pd2chkx condition 5): refusing a list ajv accepts,
// which would stop a run on a config ESLint loads, and accepting a list ajv refuses, which is the hole
// this package closes. Reported separately, because the first is the costly direction.
func TestValidatorAgreesWithAjv(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("testdata", "samples.json"))
	if err != nil {
		t.Fatal(err)
	}
	var samples samplesFile
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	if !equalVersions(samples.GeneratedFrom, GeneratedFrom()) {
		t.Fatalf("the samples were drawn from %v and the embedded schemas from %v; regenerate both together", samples.GeneratedFrom, GeneratedFrom())
	}

	var falseRefusals, falseAccepts []string
	accepted, refused := 0, 0
	for _, sample := range samples.Samples {
		schema, found, err := For(sample.Rule)
		if err != nil {
			t.Fatalf("%s: %v", sample.Rule, err)
		}
		if !found {
			t.Fatalf("%s has a sample and no schema", sample.Rule)
		}
		violation := schema.Validate(sample.Elements)
		encoded, _ := json.Marshal(sample.Elements)
		switch sample.ESLint {
		case "accepts":
			accepted++
			if violation != nil {
				falseRefusals = append(falseRefusals, fmt.Sprintf("%s %s: %v", sample.Rule, encoded, violation))
			}
		case "refuses":
			refused++
			if violation == nil {
				falseAccepts = append(falseAccepts, fmt.Sprintf("%s %s (%s at %s)", sample.Rule, encoded, sample.Keyword, sample.Path))
			}
		default:
			t.Fatalf("%s: unknown verdict %q", sample.Rule, sample.ESLint)
		}
	}

	corpusRows := 0
	entries, err := os.ReadDir(corpusDirectory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		ruleName, isRows := strings.CutSuffix(entry.Name(), ".json")
		if !isRows || entry.Name() == "known-gaps.json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(corpusDirectory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var rows []struct {
			Options []json.RawMessage `json:"options"`
		}
		if err := json.Unmarshal(data, &rows); err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		schema, found, err := For(ruleName)
		if err != nil {
			t.Fatalf("%s: %v", ruleName, err)
		}
		for _, row := range rows {
			if len(row.Options) == 0 {
				continue
			}
			corpusRows++
			if !found {
				continue
			}
			if violation := schema.Validate(row.Options); violation != nil {
				encoded, _ := json.Marshal(row.Options)
				falseRefusals = append(falseRefusals, fmt.Sprintf("%s corpus row %s: %v", ruleName, encoded, violation))
			}
		}
	}

	t.Logf("samples: ajv accepts %d and refuses %d; corpus option rows: %d", accepted, refused, corpusRows)
	report := func(label string, failures []string) {
		if len(failures) == 0 {
			return
		}
		sort.Strings(failures)
		shown := failures
		if len(shown) > 40 {
			shown = shown[:40]
		}
		t.Errorf("%s: %d\n  %s", label, len(failures), strings.Join(shown, "\n  "))
	}
	report("refuses what ajv accepts", falseRefusals)
	report("accepts what ajv refuses", falseAccepts)

	// The counts are part of the gate: a sample file that shrank would read as agreement.
	if accepted < 1900 || refused < 8000 || corpusRows != 6641 {
		t.Errorf("the gate ran on %d accepted samples, %d refused and %d corpus rows, fewer than it was built on", accepted, refused, corpusRows)
	}
}

// Every embedded schema compiles, so a rule's first config load can never be the place a schema fails.
func TestEveryEmbeddedSchemaCompiles(t *testing.T) {
	t.Parallel()

	for _, name := range RuleNames() {
		if _, _, err := For(name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	if count := len(RuleNames()); count != 384 {
		t.Errorf("the embedded schemas cover %d rules, want 384", count)
	}
}

func equalVersions(left map[string]string, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for name, version := range left {
		if right[name] != version {
			return false
		}
	}
	return true
}

// pluginOrigins are the origins docs/data/rules.json gives a rule ported from ESLint core or an ESLint
// plugin, which are the rules extract.mjs reads a schema for.
var pluginOrigins = map[string]bool{
	"core": true, "typescript-eslint": true, "react": true, "react-hooks": true, "next": true,
	"tailwind": true, "eslint-comments": true, "boundaries": true,
}

// The embedded schemas cover exactly the rules cohere ports from ESLint, as docs/data/rules.json lists
// them (which docsdata's own check keeps current with the registry). A port added without regenerating
// has no schema and would be checked by nothing, and a schema for a rule cohere dropped is dead weight.
func TestSchemasCoverEveryPortedRule(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "data", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rules struct {
		Rules []struct {
			Name   string `json:"name"`
			Origin string `json:"origin"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(data, &rules); err != nil {
		t.Fatal(err)
	}
	ported := map[string]bool{}
	for _, row := range rules.Rules {
		if pluginOrigins[row.Origin] {
			ported[row.Name] = true
		}
	}
	embedded := map[string]bool{}
	for _, name := range RuleNames() {
		embedded[name] = true
		if !ported[name] {
			t.Errorf("%s has an embedded schema and is not a ported rule; regenerate with extract.mjs", name)
		}
	}
	for name := range ported {
		if !embedded[name] {
			t.Errorf("%s is ported and has no embedded schema; regenerate with extract.mjs", name)
		}
	}
}

// A bare severity is an empty option list, which the config layer does not check because no upstream
// schema refuses it. This holds that true for every embedded schema.
func TestEverySchemaAcceptsTheBareSeverity(t *testing.T) {
	t.Parallel()

	for _, name := range RuleNames() {
		schema, found, err := For(name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if found {
			if violation := schema.Validate(nil); violation != nil {
				t.Errorf("%s refuses the bare severity: %v", name, violation)
			}
		}
	}
}

// The extension list is a set of decisions, so its size is a bare equality: adding one is a change to
// this number, made on purpose (#pd2chkx condition 6).
func TestExtensionsAreExactlyTheDeclaredOnes(t *testing.T) {
	t.Parallel()

	if len(Extensions) != 1 {
		t.Fatalf("there are %d extensions, and exactly 1 is declared", len(Extensions))
	}
	for _, extension := range Extensions {
		if extension.Reason == "" {
			t.Errorf("%s's extension %q carries no reason", extension.Rule, extension.Key)
		}
		if _, found, _ := For(extension.Rule); !found {
			t.Errorf("%s has an extension and no schema", extension.Rule)
		}
	}
}

// boundaries/dependencies' elements load through its extension, are refused without it, and the rest of
// the element is still checked.
func TestTheBoundariesExtensionAcceptsElementsAndNothingElse(t *testing.T) {
	t.Parallel()

	withElements := []json.RawMessage{json.RawMessage(`{"default":"disallow","elements":[{"type":"api","pattern":"api/**"}]}`)}
	if err := Check("boundaries/dependencies", withElements); err != nil {
		t.Errorf("elements are refused: %v", err)
	}
	schema, _, _ := For("boundaries/dependencies")
	if err := schema.Validate(withElements); err == nil {
		t.Errorf("upstream's schema accepts elements without the extension, so the extension is not needed")
	}
	otherKey := []json.RawMessage{json.RawMessage(`{"default":"disallow","elements":[],"notAnUpstreamKey":true}`)}
	if err := Check("boundaries/dependencies", otherKey); err == nil {
		t.Errorf("a key outside the extension is accepted")
	}
}

// A refusal names the element as a config author counts it, the path inside it, and what upstream
// expected (#pd2chkx condition 8).
func TestARefusalNamesTheElementThePathAndTheExpectation(t *testing.T) {
	t.Parallel()

	err := Check("@typescript-eslint/no-deprecated", []json.RawMessage{json.RawMessage(`{"allow":[{"from":"File","name":"x"}]}`)})
	if err == nil {
		t.Fatal("a mis-cased from is accepted")
	}
	for _, want := range []string{"element 1 at allow[0]", "upstream's schema, which ESLint validates against, expects"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal %q does not say %q", err, want)
		}
	}
}
