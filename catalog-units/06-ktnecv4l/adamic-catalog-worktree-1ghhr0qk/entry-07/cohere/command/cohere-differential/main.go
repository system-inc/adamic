// Command cohere-differential runs both gates over one tree and reports where they disagree.
//
// This is the acceptance instrument for the whole project. cohere replaces the gate only when it
// can be shown to find what the gate finds and stay silent where the gate is silent, and until this
// command existed that claim was established by a person running two commands and comparing counts
// by eye — which is not a harness, does not survive that person, and cannot answer the question it
// appears to answer.
//
// The question it appears to answer is "do the two gates agree." The question it has to answer
// first is "did both gates run, and can this harness see a difference at all." Those are not the
// same, and an empty diff satisfies the first while proving nothing about the second. So this
// command plants a known violation on every run and refuses a verdict when the plant did not come
// out the far end.
//
//	cohere-differential                          compare, planting the default controls
//	cohere-differential -root ~/Projects/ahra    the tree to compare over
//	cohere-differential -cohere-binary <path>    which cohere to measure
//	cohere-differential -no-controls             skip planting, and say so in the report
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/system-inc/cohere/internal/differential"
	"github.com/system-inc/cohere/internal/lint/configuration"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "cohere-differential: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	options, err := parseArguments(os.Args[1:])
	if err != nil {
		return err
	}

	lintConfig, err := configuration.Load(filepath.Join(options.root, ".oxlintrc.json"))
	if err != nil {
		return fmt.Errorf("loading the lint config: %w", err)
	}

	if err := checkControlsAreLintable(lintConfig, options.controls, cohereOnlyRuleSettings(options.root)); err != nil {
		return err
	}

	// Asked of the binary being measured, not of this one. `registry.All()` is the registry linked
	// into the harness, and when the two builds differ every classification describes the wrong
	// tool. That produced a real wrong answer: 128 findings reported as both-active, meaning both
	// sides had the rule and disagreed, when the measured binary had no such rule at all.
	cohereRules, err := differential.CompiledRulesOf(
		context.Background(), options.cohereCommand, ruleListProbeFile,
	)
	if err != nil {
		return err
	}

	report, err := differential.Run(context.Background(), differential.RunOptions{
		Root:                    options.root,
		Cohere:                  options.cohereCommand,
		Gate:                    options.gateCommand,
		CohereRules:             cohereRules,
		ConfiguredRules:         configuredRuleNames(lintConfig),
		Controls:                options.controls,
		ExtraCohereRuleSettings: cohereOnlyRuleSettings(options.root),
	})
	if err != nil {
		return err
	}

	rendered := &strings.Builder{}
	differential.Write(rendered, report)
	fmt.Print(rendered.String())

	// The exit code comes from the same value the reader sees rather than being recomputed, so the
	// two cannot disagree. Agreed() is false for a disagreement and equally false for a run that
	// proved nothing, which is correct: neither is a result anyone should build on.
	if !report.Agreed() {
		os.Exit(1)
	}
	return nil
}

// ruleListProbeFile is the file `-explain` is pointed at to enumerate the measured binary's rules.
//
// Any file in the program works, because every compiled rule is accounted for in that output
// whether it listened to the file, declined it, or was never configured. This one is chosen for
// being unremarkable: a file no rule finds interesting still names all of them.
const ruleListProbeFile = "modules/mcp/McpApi.ts"

// configuredRuleNames is every rule the lint config switches on.
//
// The config keys carry their plugin prefix (`nexus/consistency-no-enum`) and cohere reports bare
// names, so both are normalized to the same form here. Skipping that would make every configured
// rule look unconfigured, and every real disagreement would then be filed as not-configured and
// silently excused.
//
// Rules set to off are excluded rather than counted, because a rule the config turned off ran over
// no files, which is exactly what not-configured means.
//
// **The rules block is not the whole configuration.** Forty rules are enforced by the `plugins` declarations
// and named in no rules block, proven with planted violations rather than read from a file. Reading
// only `lintConfig.Rules` therefore reported a denominator of 159 when the real one is 214, and it
// did so in the summary line a reader uses to decide whether cohere is safe to trust yet.
//
// An understated denominator here makes a clean result mean less than it claims, in the one harness
// the whole migration is judged on.
//
// This once merged a captured catalog of what the replaced tools enforced, so that rules arriving
// from `plugins` declarations were counted even though no rules block names them. That catalog is
// gone: cohere passed it, and a denominator that can only be met stops asserting anything once it
// has been. The plugin-default knowledge it carried lives in `configuration.PluginDefaultRules`,
// which the resolver already applies before this function sees the config, so those rules arrive
// here as ordinary configured rules rather than as a second source to union in.
func configuredRuleNames(lintConfig *configuration.Config) map[string]bool {
	names := map[string]bool{}
	for name, setting := range lintConfig.Rules {
		normalized := differential.NormalizeRuleName(name)
		// A rule turned off ran over no files, so it is not part of what this comparison covers.
		if setting.Severity == configuration.SeverityOff {
			continue
		}
		names[normalized] = true
	}
	return names
}

type arguments struct {
	root          string
	cohereCommand differential.GateCommand
	gateCommand   differential.GateCommand
	controls      []differential.Control
}

func parseArguments(rest []string) (arguments, error) {
	parsed := arguments{root: defaultRoot}
	cohereBinary := "cohere"
	plantControls := true

	for index := 0; index < len(rest); index++ {
		switch rest[index] {
		case "-root", "--root":
			if index+1 >= len(rest) {
				return parsed, fmt.Errorf("-root needs a directory")
			}
			index++
			parsed.root = rest[index]
		case "-cohere-binary", "--cohere-binary":
			if index+1 >= len(rest) {
				return parsed, fmt.Errorf("-cohere-binary needs a path")
			}
			index++
			cohereBinary = rest[index]
		case "-no-controls", "--no-controls":
			plantControls = false
		default:
			return parsed, fmt.Errorf("unknown argument %q", rest[index])
		}
	}

	absoluteRoot, err := filepath.Abs(parsed.root)
	if err != nil {
		return parsed, fmt.Errorf("resolving the root %s: %w", parsed.root, err)
	}
	parsed.root = absoluteRoot

	parsed.cohereCommand = differential.GateCommand{
		Name:      "cohere",
		Program:   cohereBinary,
		Arguments: []string{"-lint"},
		Directory: parsed.root,
	}
	parsed.gateCommand = differential.GateCommand{
		Name:      "gate",
		Program:   "node",
		Arguments: []string{gateRunnerPath, "--no-cache"},
		Directory: parsed.root,
	}

	if plantControls {
		parsed.controls = defaultControls()
	}

	return parsed, nil
}

// defaultControls are the planted violations that prove the pipeline carries a difference.
//
// # Why there is a shared control and not a cohere-only one
//
// The first version of this planted an enum and expected cohere alone to see it. It missed, and the
// miss was correct: `consistency-no-enum` is implemented on both sides and enabled in the config, so
// both gates saw the plant and it matched. The control was wrong, not the harness — and the harness
// caught it, refused a verdict, and printed the difference list as not-a-measurement, which is the
// behavior it was built for.
//
// Chasing that further produced a more useful conclusion than a working control would have. There
// is no cohere-only control available on this tree today, and the reason is structural rather than
// an oversight:
//
//   - Every rule cohere implements was ported from the gate, so both sides have it.
//   - A rule in cohere but absent from the config runs over no files in cohere either, so it fires
//     on neither side.
//   - Cohere honors `ignorePatterns` and the config overrides, so a plant in an ignored path or
//     under a scoping override is correctly silent on both sides.
//
// A cohere-only difference can therefore only arise from a genuine behavioral divergence, and one
// cannot be planted to order without first knowing of a real one. That is a good property of a
// faithful port and a real limit on this instrument, so it is written here rather than worked
// around with a control that tests something other than what it claims.
//
// What is planted instead is a shared control: a violation both gates must see. It does not prove
// direction, and `ControlsProven` still returns false because of that. It does prove the pipeline
// end to end — process launched, output parsed, paths normalized, rule names collapsed across two
// decoration schemes, comparison run — which is the layer where a silent total mismatch lives. A
// plant that both sides report as one shared finding, rather than as two one-sided ones, is
// positive evidence that path normalization and rule-name normalization both worked.
//
// The gate-only direction needs no plant: the gate reports findings from rules cohere has not
// ported, so that direction has a natural population on this tree.
func defaultControls() []differential.Control {
	return []differential.Control{{
		Name:         "shared-enum",
		RelativePath: filepath.Join("code-quality", "differential-control", "PlantedEnum.ts"),
		Contents: "// Planted by cohere-differential to prove the harness can see a difference.\n" +
			"// Removed automatically when the run finishes. If you are reading this in a working\n" +
			"// tree, a differential run was interrupted and this file is safe to delete.\n" +
			"export enum PlantedControlOrder {\n    First = 'First',\n}\n",
		Rule: "consistency-no-enum",
		// Both gates implement and enable this rule, so both must see it. Expecting it shared is
		// the claim being tested; expecting it one-sided would be testing a divergence that does
		// not exist.
		ExpectedShared: true,
	}, {
		// The cohere-only control. Verified against the tree at 03:17 rather than assumed:
		//
		//   registry.go:52                 cohere has import-require-path-alias compiled in
		//   NexusLintConfiguration.ts:31   the ESLINT plugin defines it
		//   OxlintNexusPlugin.mjs          the OXLINT plugin does not, at all
		//   .oxlintrc.json                 no mention of the rule
		//
		// The gate we diff against runs oxlint, so it is structurally incapable of producing this
		// finding, and no configuration change on that side can make it. That is what makes this a
		// sturdy directional control rather than one resting on a defect somebody might repair: a
		// control built on a bug stops discriminating the moment the bug is fixed.
		//
		// The rule needs its options, because the shared config does not enable it and cohere runs
		// an unconfigured rule over no files. It is marked Required in the registry, so an
		// unconfigured run refuses loudly instead of coming back clean, and the harness cannot get
		// a false pass from forgetting this.
		Name:         "cohere-only-path-alias",
		RelativePath: filepath.Join("code-quality", "differential-control", "deep", "nested", "PlantedImport.ts"),
		Contents: `// Planted by cohere-differential to prove the harness can see a one-sided
// difference. Removed automatically when the run finishes.
import { PlantedTarget } from '../../target/PlantedTarget';

export const PlantedUse = PlantedTarget;
`,
		Rule:         "import-require-path-alias",
		ExpectedSide: differential.SideCohere,
	}, {
		// The target of the planted import. It exists so the import resolves and so the rule has an
		// aliased directory to climb into; it carries no violation of its own and is expected to
		// produce nothing, which is why it declares no Rule and no side.
		Name:         "cohere-only-path-alias-target",
		RelativePath: filepath.Join("code-quality", "differential-control", "target", "PlantedTarget.ts"),
		// PascalCase because it is exported: the tree's own constant-casing rule flags an exported
		// camelCase constant, and a support file that trips a rule stops being support. The first
		// version was `plantedTarget` and it produced a real finding of its own, which is exactly
		// the measurement perturbation these files must not cause.
		Contents:       "export const PlantedTarget = 'planted';\n",
		ExpectsNothing: true,
	}, {
		// The gate-only control, and the mirror image of the cohere-only one.
		//
		//   .oxlintrc.json    react/exhaustive-deps is enabled at error
		//   oxlint plugins    it comes from oxlint's built-in react plugin
		//   registry          cohere implements no react-hooks rule at all
		//
		// This replaced a control keyed on consistency-organize-imports, and the replacement is the
		// whole point. That rule was unported when the control was written, which made the
		// asymmetry look structural, and it was temporary: the rule sat on a porting slice, so a
		// correct port would have made the control stop discriminating and ControlsProven return
		// false. The differential would have printed "no verdict" and nothing in the porting work
		// would have looked wrong.
		//
		// A control resting on the migration being incomplete expires on purpose, which is the same
		// defect as a control resting on a bug somebody might fix. This one does not: the 16
		// react-hooks rules are one 3.9 MB bundle carrying the React Compiler, not a rule set anyone
		// is porting, so cohere cannot produce this finding no matter how far parity goes.
		//
		// One honest caveat, measured rather than assumed after the fact. This asymmetry is durable
		// but not structural: `exhaustive_deps.rs` does exist in oxc at 4,477 lines, and 14 of the 16
		// react-hooks rules have an oxc implementation. So the argument here is "nobody will port a
		// 4,477-line rule", which is a judgment about effort rather than a fact about the tools, and
		// it is the weakest of the three reasons a control can rest on.
		//
		// A control resting on a bug expires when the bug is fixed. One resting on the migration
		// being incomplete expires when the migration succeeds. One resting on a rule being
		// impractical expires when somebody finds it practical.
		//
		// If this ever needs replacing, the sturdier ground is the eight enabled `better-tailwindcss`
		// rules: the gate produces them through `OxlintTailwindPlugin.mjs`, a JavaScript plugin
		// bridge cohere does not have, which is the same shape as the cohere-only control rather than
		// a file somebody could translate.
		//
		// The violation is a hook whose dependency array omits a value it reads.
		Name:         "gate-only-exhaustive-deps",
		RelativePath: filepath.Join("code-quality", "differential-control", "PlantedEffect.ts"),
		Contents: `// Planted by cohere-differential to prove the harness can see a one-sided
// difference from the gate. Removed automatically when the run finishes.
import * as React from 'react';

export function PlantedEffect(properties: { label: string }) {
    const [seen, setSeen] = React.useState('');
    React.useEffect(function () {
        setSeen(properties.label);
    }, []);
    return seen;
}
`,
		Rule:         "exhaustive-deps",
		ExpectedSide: differential.SideGate,
	}}
}

// cohereOnlyRuleSettings enables, for cohere's run alone, the rules the gate cannot express.
//
// Only rules absent from the gate's plugin belong here. A rule both sides implement must stay on
// the shared config, because configuring it differently per side manufactures differences that say
// nothing about either implementation. `Run` refuses a rule the tree's config already configures,
// which is the mechanical version of that boundary rather than a comment asking for care.
func cohereOnlyRuleSettings(root string) map[string]any {
	// The alias is scoped to the control's own directory rather than to a real source root.
	//
	// A first version aliased `modules`, which is a real directory, and the rule fired 95 times
	// across the tree. Every one of those was a true finding about code nobody has asked this rule
	// to judge, and they drowned the single finding the control exists to produce. A control must
	// perturb the measurement by exactly the one finding it plants; anything else is the
	// instrument changing what it measures.
	//
	// Aliasing the control directory means the only relative import that can climb into an aliased
	// directory is the one in the planted file.
	return map[string]any{
		"nexus/import-require-path-alias": []any{
			"error",
			map[string]any{
				"repositoryRoot": root,
				"aliases": []any{
					map[string]any{
						"directory": filepath.Join("code-quality", "differential-control"),
						"alias":     "@differential-control",
					},
				},
			},
		},
	}
}

// gateRunnerPath is the gate as it actually ships, run through its own cached runner rather than by
// invoking oxlint directly. Comparing against a hand-rolled oxlint invocation would measure a gate
// nobody runs.
var gateRunnerPath = filepath.Join(
	"libraries", "structure", "libraries", "nexus", "code-quality", "oxlint", "RunCachedOxlint.ts",
)

// defaultRoot is the tree this instrument was commissioned to measure.
var defaultRoot = filepath.Join(os.Getenv("HOME"), "Projects", "ahra")

// checkControlsAreLintable refuses to run when a control was placed where its rule cannot fire.
//
// A control exists to prove the harness can see a difference, so a control that misses is supposed
// to mean the harness is blind. But a control written to a path the config ignores, or naming a
// rule the config never enables, misses for a reason that has nothing to do with the harness and
// produces exactly the same output. The evidence and the defect are indistinguishable, which makes
// the control worse than no control: it reports a failure the reader will attribute to the wrong
// thing.
//
// So the precondition is checked against the config rather than assumed from the path, and it is a
// refusal to start rather than a note in the report. The comment on Control.RelativePath already
// said the path must be one both gates lint. A comment is carefully safe; this is mechanically
// safe, and mechanical is the kind that survives the next editor.
//
// The check runs before either gate is launched, so a misplaced control costs a second rather than
// two full lint runs and a misleading verdict.
func checkControlsAreLintable(lintConfig *configuration.Config, controls []differential.Control, extraVerifyRules map[string]any) error {
	// The rules supplied to cohere alone, reduced to bare names. A control naming one of these is
	// enabled for cohere's run even though the tree's own config says nothing about it, so the
	// check below has to know about them or it rejects exactly the directional control it should
	// be admitting. The ignore check still applies to them: an ignored path is ignored by both
	// sides regardless of which config names the rule.
	suppliedToVerify := map[string]bool{}
	for name := range extraVerifyRules {
		suppliedToVerify[differential.NormalizeRuleName(name)] = true
	}

	for _, control := range controls {
		// A support file asserts no finding, so there is no rule to check it can fire. The ignore
		// check below would also be meaningless for it: whether the tree lints it changes nothing
		// about what it proves, which is nothing.
		if control.ExpectsNothing {
			continue
		}

		resolved := lintConfig.Resolve(filepath.Join(lintConfig.Root, control.RelativePath))

		if resolved.Ignored {
			return fmt.Errorf(
				"control %q is placed at %s, which the lint config ignores via %q, so it would miss for a reason unrelated to the harness; move it somewhere both gates lint",
				control.Name, control.RelativePath, resolved.IgnoredBy,
			)
		}

		// The rule has to be enabled for that specific path, not merely present in the config: an
		// override can scope a rule off for exactly the directory a control was written to, and
		// that override is invisible from the base rule list.
		if !resolved.Enabled(pluginQualified(lintConfig, control.Rule)) && !suppliedToVerify[control.Rule] {
			return fmt.Errorf(
				"control %q expects rule %s to fire at %s, but the lint config does not enable it there, so the control would miss for a reason unrelated to the harness",
				control.Name, control.Rule, control.RelativePath,
			)
		}
	}
	return nil
}

// pluginQualified finds the config's own spelling of a bare rule name.
//
// The config keys rules as `plugin/rule-name` and a control names the bare rule, so a direct lookup
// would miss every time and this guard would reject every control. Searching for the key whose
// normalized form matches keeps the control declarations free of plugin prefixes, which is the
// right place for that knowledge to live: a rule that moves between plugins should not break a
// control that never mentioned one.
//
// Returning the bare name when nothing matches is deliberate. It makes Enabled report unconfigured,
// which is the correct answer for a rule the config genuinely does not mention, and it keeps this
// helper from being the thing that decides a control is valid.
func pluginQualified(lintConfig *configuration.Config, bareRuleName string) string {
	for name := range lintConfig.Rules {
		if differential.NormalizeRuleName(name) == bareRuleName {
			return name
		}
	}
	return bareRuleName
}
