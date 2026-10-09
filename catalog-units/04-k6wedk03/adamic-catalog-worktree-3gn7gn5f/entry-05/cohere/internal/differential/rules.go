// Asking the measured binary which rules it has, rather than assuming they match this one's.
//
// The classification split turns on whether cohere implements a rule, and that fact was read from
// `registry.All()` — the registry linked into the harness, not into the binary under measurement.
// When the two differ, every classification describes the wrong tool, and it does so silently.
//
// It happened. A run at 03:22 measured a binary built before `react-component-no-multiple-primary`
// landed. The harness reported 128 findings as `both-active`, meaning "both sides have this rule
// enabled and disagree," when the truth was that the measured binary had no such rule at all. That
// is a defect in the port by one reading and a coverage gap by the other, and the words point at
// different people.
//
// A binary without a rule and a binary whose rule found nothing are indistinguishable from the
// finding list alone, which is why this cannot be inferred and has to be asked.
package differential

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ruleTimingPattern matches a rule's own line in `-explain` output:
//
//	no-compare-neg-zero  0.02ms, 54 nodes
//
// Two spaces of indent and the name first, which distinguishes these from the coverage notes that
// share the indent but begin with a word and a colon.
var ruleTimingPattern = regexp.MustCompile(`^  ([a-z][a-z0-9-]*)  [0-9.]+m?s,`)

// declinedRulesPattern captures the rules that looked at the file and opted out. They are
// implemented, so they belong in the set, and they appear nowhere else in the output.
var declinedRulesPattern = regexp.MustCompile(`^\s*declined this file \(the rule looked and opted out\): (.+)$`)

// unconfiguredRulePattern captures a rule that exists but the config never enabled, which is the
// third way a compiled rule can be absent from the timing table.
var unconfiguredRulePattern = regexp.MustCompile(`^\s*([a-z][a-z0-9-]*) did not run:`)

// ruleCountPattern reads the rule count from the coverage line, which is the total this parse must
// account for:
//
//	lint: 128 findings — 33 rules over 3407 files, ...
var ruleCountPattern = regexp.MustCompile(`^lint: \d+ findings? — (\d+) rules? over`)

// CompiledRulesOf asks the binary which rules it has.
//
// `-rules` is the purpose-built answer: one name per line, sorted, nothing to reconcile. It is
// tried first and `-explain` is the fallback, because a binary older than `-rules` still answers
// the question, just less directly. Both are asked of the binary rather than derived from the
// registry linked into this harness, which is the entire point: when the two builds differ, a
// classification read from the wrong registry describes the wrong tool and says so confidently.
func CompiledRulesOf(ctx context.Context, command GateCommand, explainFile string) (map[string]bool, error) {
	if names, err := rulesFromRulesFlag(ctx, command); err == nil {
		return names, nil
	}
	return rulesFromExplain(ctx, command, explainFile)
}

// rulesFromRulesFlag reads `cohere -rules`, which prints one rule name per line.
//
// An empty list is refused rather than returned. A binary that printed nothing would produce a
// harness believing cohere implements no rules at all, under which every gate finding classifies
// not-ported and every real disagreement is excused. That is the vacuous pass this package exists
// to refuse, arriving through the door meant to prevent it.
func rulesFromRulesFlag(ctx context.Context, command GateCommand) (map[string]bool, error) {
	probe := command
	probe.Arguments = []string{"-rules"}

	output, err := runGate(ctx, probe)
	if err != nil {
		return nil, fmt.Errorf("%s does not answer -rules: %w", command.Program, err)
	}

	names, err := namesFromRuleLines(output)
	if err != nil {
		return nil, fmt.Errorf("%s %w", command.Program, err)
	}
	return names, nil
}

// namesFromRuleLines parses `-rules` stdout, which is bare rule names one per line.
//
// Split out from the invocation so the contract can be tested without a binary, because it was
// being tested by accident: `cohere -rules` gained a provenance note and this kept working only
// because the note goes to stderr while the names go to stdout. Had it landed on stdout, the guard
// would have rejected correct output.
//
// The guard stays strict rather than learning to skip prose. A line this does not understand is a
// contract change, and absorbing it quietly is how a format drift becomes a short rule list, where
// every missing name becomes a rule the harness believes cohere lacks and every real disagreement
// on it gets excused as a coverage gap.
func namesFromRuleLines(output string) (map[string]bool, error) {
	names := map[string]bool{}
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.ContainsAny(trimmed, " \t") {
			return nil, fmt.Errorf("-rules printed %q, which is not a bare rule name", trimmed)
		}
		names[NormalizeRuleName(trimmed)] = true
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("-rules printed no rule names")
	}
	return names, nil
}

// rulesFromExplain recovers the rule list from `-explain` output, for binaries predating `-rules`.
//
// `-explain` names all three states a compiled rule can be in: it listened, it declined the file,
// or the config never enabled it. Their sum is checked against the coverage line's own count, so a
// parse that silently captured a subset fails loudly instead of producing a plausible short list.
func rulesFromExplain(ctx context.Context, command GateCommand, explainFile string) (map[string]bool, error) {
	probe := command
	probe.Arguments = []string{"-explain", explainFile}

	output, err := runGate(ctx, probe)
	if err != nil {
		// A binary predating `-explain` cannot answer, and that is a real state rather than a
		// broken one: the harness is measuring an older build. It is an error rather than a silent
		// fallback to this harness's own registry, because that fallback is exactly the defect
		// being fixed, and a caller that wants the old behavior should have to ask for it.
		return nil, fmt.Errorf(
			"%s cannot report its own rules, so classification would describe this harness rather than the binary being measured: %w",
			command.Program, err,
		)
	}

	names := map[string]bool{}
	declaredCount := 0

	for _, line := range strings.Split(output, "\n") {
		if match := ruleCountPattern.FindStringSubmatch(strings.TrimRight(line, "\r")); match != nil {
			declaredCount, _ = strconv.Atoi(match[1])
			continue
		}
		if match := ruleTimingPattern.FindStringSubmatch(line); match != nil {
			names[NormalizeRuleName(match[1])] = true
			continue
		}
		if match := declinedRulesPattern.FindStringSubmatch(line); match != nil {
			for _, name := range strings.Split(match[1], ",") {
				if trimmed := strings.TrimSpace(name); trimmed != "" {
					names[NormalizeRuleName(trimmed)] = true
				}
			}
			continue
		}
		if match := unconfiguredRulePattern.FindStringSubmatch(line); match != nil {
			names[NormalizeRuleName(match[1])] = true
		}
	}

	if declaredCount == 0 {
		return nil, fmt.Errorf(
			"%s printed no lint coverage line while explaining %s, so its rule list cannot be trusted",
			command.Program, explainFile,
		)
	}
	// The binary states its own total, so the parse can be checked against it rather than believed.
	// This is the guard that makes a short list impossible to mistake for a small binary.
	if len(names) != declaredCount {
		return nil, fmt.Errorf(
			"parsed %d rule names from %s but it reports running %d rules, so the parse is incomplete and the classification would be wrong",
			len(names), command.Program, declaredCount,
		)
	}

	return names, nil
}
