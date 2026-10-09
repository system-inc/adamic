// Tests for the text-to-structure layer, fed output the two gates actually printed.
//
// This is a separate file from differential_test.go because it defends a different failure. Those
// tests prove the comparison can detect a difference given findings. These prove findings exist to
// compare at all.
//
// The gap matters more than it looks. Every parse failure in this layer is silent by construction:
// a regex that stops matching yields zero findings, and zero findings on one side makes every
// finding on the other look one-sided while the diff stays perfectly well-formed. A path
// normalization bug does the same thing — cohere prints absolute paths, the gate prints relative
// ones, and if they fail to reconcile then two identical findings never match and the harness
// reports total disagreement in a shape that looks like a working result.
//
// So the inputs below are copied from real runs on the ahra tree at 01:16 on 2026-08-23, not
// written to match the patterns. A parser tested only against its author's idea of the format is a
// parser tested against nothing.
package differential

import "testing"

// Not a machine path: the tree root the copied output names, trimmed from it as text; no file is read.
const ahraTreeRoot = "/Users/kirkouimet/Projects/ahra"

// TestParsesRealVerifyOutput uses cohere's own stdout, coverage lines and all, so the summary
// lines have to be recognized as summaries rather than counted as unparsed findings.
func TestParsesRealVerifyOutput(test *testing.T) {
	test.Parallel()
	// Not a machine path: cohere's output copied verbatim, whose tree root the parser trims.
	output := `graph built in 1.752s — 9973 files in the program, 3407 of them ours
/Users/kirkouimet/Projects/ahra/modules/mcp/McpApi.ts:240:16 - Identifier "params" should not be abbreviated. Use "parameters" or a more descriptive name. [consistency-no-abbreviated-identifier/noParams]
lint: 1 findings — 22 rules over 3407 files, 2098253 nodes visited, in 1.036s
  note: rule consistency-no-utils-folder listened to no files — it ran, but it never looked
  suppressed: 279 findings silenced by a disable comment, 264 of them without a stated reason
  config: rule consistency-require-type-suffix scoped off for 6 files by the config`

	result, err := ParseCohere(output, ahraTreeRoot)
	if err != nil {
		test.Fatalf("parsing real cohere output: %v", err)
	}
	if len(result.Findings) != 1 {
		test.Fatalf("want 1 finding from real output, got %d (unparsed: %v)", len(result.Findings), result.UnparsedLines)
	}
	if len(result.UnparsedLines) != 0 {
		test.Errorf("real output produced unparsed lines, each of which is a finding the harness cannot see: %v", result.UnparsedLines)
	}

	finding := result.Findings[0]
	if finding.File != "modules/mcp/McpApi.ts" {
		test.Errorf("path is %q, want it relative to the tree root — an absolute-versus-relative mismatch makes every finding look one-sided", finding.File)
	}
	if finding.Rule != "consistency-no-abbreviated-identifier" {
		test.Errorf("rule is %q, want the bare name with the message id stripped", finding.Rule)
	}
	if finding.Line != 240 {
		test.Errorf("line is %d, want 240", finding.Line)
	}
}

// TestParsesRealGateOutput uses the gate's own stdout. The parenthesized route-group directories
// are real in this tree and are the shape most likely to break a naive split on punctuation.
func TestParsesRealGateOutput(test *testing.T) {
	test.Parallel()
	output := `app/(os-layout)/finance/_components/FinanceLedgerTable.tsx:49:5: error structure(react-component-no-multiple-primary): File contains a component with 108 lines.
libraries/structure/source/components/files/FileCarousel.tsx:36:8: error structure(react-component-no-multiple-primary): File contains a component with 184 lines.`

	result, err := ParseGate(output, ahraTreeRoot)
	if err != nil {
		test.Fatalf("parsing real gate output: %v", err)
	}
	if len(result.Findings) != 2 {
		test.Fatalf("want 2 findings from real output, got %d (unparsed: %v)", len(result.Findings), result.UnparsedLines)
	}
	if len(result.UnparsedLines) != 0 {
		test.Errorf("real output produced unparsed lines: %v", result.UnparsedLines)
	}
	if result.Findings[0].Rule != "react-component-no-multiple-primary" {
		test.Errorf("rule is %q, want the bare name with the plugin prefix stripped", result.Findings[0].Rule)
	}
	if result.Findings[0].File != "app/(os-layout)/finance/_components/FinanceLedgerTable.tsx" {
		test.Errorf("relative path mangled to %q", result.Findings[0].File)
	}
}

// TestTheTwoGatesAgreeOnOneFindingAfterParsing is the end-to-end control for this layer, and the
// one that would have caught a path or name normalization bug on its own.
//
// The same defect is fed in both formats. If parsing is right, the two produce equal keys and the
// comparison reports no difference. If either path or rule normalization is wrong, the keys differ
// and one finding becomes two one-sided differences — which is exactly what a total mismatch looks
// like, and it looks like working output.
func TestTheTwoGatesAgreeOnOneFindingAfterParsing(test *testing.T) {
	test.Parallel()
	// Not a machine path: cohere's output copied verbatim, whose tree root the parser trims.
	cohereResult, err := ParseCohere(
		`/Users/kirkouimet/Projects/ahra/app/Component.tsx:198:1 - File contains a component with 96 lines. [react-component-no-multiple-primary/tooMany]`,
		ahraTreeRoot,
	)
	if err != nil {
		test.Fatalf("parsing cohere side: %v", err)
	}
	gateResult, err := ParseGate(
		`app/Component.tsx:198:1: error structure(react-component-no-multiple-primary): File contains a component with 96 lines.`,
		ahraTreeRoot,
	)
	if err != nil {
		test.Fatalf("parsing gate side: %v", err)
	}

	if len(cohereResult.Findings) != 1 || len(gateResult.Findings) != 1 {
		test.Fatalf("setup did not parse: cohere %d, gate %d", len(cohereResult.Findings), len(gateResult.Findings))
	}

	if cohereResult.Findings[0].Key() != gateResult.Findings[0].Key() {
		test.Fatalf("the same defect keyed differently across the two formats:\n  cohere %q\n  gate   %q\nnormalization is wrong, and this presents as total disagreement rather than as an error",
			cohereResult.Findings[0].Key(), gateResult.Findings[0].Key())
	}

	report := Compare(Inputs{
		CohereFindings:   cohereResult.Findings,
		GateFindings:     gateResult.Findings,
		CoherePopulation: Population{Findings: 1, FilesWalked: 3407, Rules: 22},
		GatePopulation:   Population{Findings: 1, FilesWalked: 3407, Rules: 181},
		CohereRules:      map[string]bool{"react-component-no-multiple-primary": true},
		ConfiguredRules:  map[string]bool{"react-component-no-multiple-primary": true},
	})
	if len(report.Differences) != 0 {
		test.Errorf("one defect reported in both formats produced %d differences, want 0: %+v", len(report.Differences), report.Differences)
	}
}

// TestUnparsedLinesAreSurfacedNotDropped is this layer's own vacuity guard. A format change that
// silently stops matching must not present as a clean run.
func TestUnparsedLinesAreSurfacedNotDropped(test *testing.T) {
	test.Parallel()
	result, err := ParseCohere("this line is neither a finding nor a summary", "")
	if err != nil {
		test.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 0 {
		test.Fatalf("invented %d findings from an unrecognized line", len(result.Findings))
	}
	if len(result.UnparsedLines) != 1 {
		test.Error("an unrecognized line was dropped silently — a finding the harness cannot read must not look like a finding that is not there")
	}
}

// TestPluginSlashRuleNamesSurviveNormalization pins a real hazard in the shared normalizer.
//
// NormalizeRuleName strips everything after the first slash, to turn cohere's
// `rule-name/messageId` into `rule-name`. Oxlint's own catalog writes rules as `plugin/rule-name`
// — `typescript/no-explicit-any` — where the part after the slash is the rule and the part before
// it is the plugin. The two conventions collide, and the collision is silent: the name still
// parses, still compares, and simply refers to the wrong thing.
//
// This test used to document the boundary rather than assert the collision was handled, and it
// named the exact moment it would bite: "if a future reader wires rule catalogs from the config
// into Inputs.ConfiguredRules." That happened at 01:48 when cohere-differential was first pointed
// at the real tree. Every config key collapsed to its plugin, the configured-rule set became three
// junk names, and every rule read as unconfigured, so the per-rule table labelled an enabled rule
// not-configured. It surfaced as one wrong word and would have excused every real disagreement.
//
// It is handled now. The two conventions are told apart by shape rather than by a plugin list, so
// a new plugin needs no edit: rule names are hyphenated and message ids are camelCase, so the
// segment containing a hyphen is the rule name. The assertions below now pin the fixed behavior.
func TestPluginSlashRuleNamesSurviveNormalization(test *testing.T) {
	test.Parallel()
	// The forms that actually appear in gate and cohere output today, which must keep working.
	fromOutput := map[string]string{
		"structure(react-component-no-multiple-primary)": "react-component-no-multiple-primary",
		"nexus(consistency-require-constant-casing)":     "consistency-require-constant-casing",
		"consistency-no-abbreviated-identifier/noParams": "consistency-no-abbreviated-identifier",
		"consistency-no-shouting":                        "consistency-no-shouting",
	}
	for raw, want := range fromOutput {
		if got := NormalizeRuleName(raw); got != want {
			test.Errorf("NormalizeRuleName(%q) = %q, want %q", raw, got, want)
		}
	}

	// The config form, which now survives. A regression here means the configured-rule set has gone
	// back to holding plugin names, and every genuine disagreement would be excused as
	// not-configured.
	fromConfig := map[string]string{
		"typescript/no-explicit-any":                    "no-explicit-any",
		"nexus/consistency-no-enum":                     "consistency-no-enum",
		"structure/react-component-no-multiple-primary": "react-component-no-multiple-primary",
	}
	for raw, want := range fromConfig {
		if got := NormalizeRuleName(raw); got != want {
			test.Errorf("NormalizeRuleName(%q) = %q, want %q", raw, got, want)
		}
	}
}
