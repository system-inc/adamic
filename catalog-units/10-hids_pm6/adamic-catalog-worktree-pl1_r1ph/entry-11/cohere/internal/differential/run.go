// Running both gates for real, and proving the harness saw a difference before it is believed.
//
// Everything else in this package is pure: findings in, report out, easy to test and easy to trust.
// This file is where the harness touches the world, and it is therefore where the interesting lies
// live. A subprocess that fails to launch, a working directory that is not the tree, a config that
// was never read, a regex that stopped matching after an output format moved — every one of them
// produces a smaller finding list, and a smaller finding list on both sides is indistinguishable
// from agreement.
//
// So this file does not report a comparison. It reports a comparison plus the evidence that the
// comparison happened, and it plants a known violation to prove the pipeline can carry one end to
// end before any empty result is allowed to mean anything.
package differential

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// GateCommand is how to invoke one side.
type GateCommand struct {
	// Name is what appears in the report.
	Name string
	// Program and Arguments are the exact invocation.
	Program   string
	Arguments []string
	// Directory is the working directory, which for both gates is the tree being linted. Getting
	// this wrong is the single most likely way to produce a confident empty result: both gates
	// resolve their config and their file list from here.
	Directory string
}

// String renders the invocation so a reader can rerun it verbatim.
func (command GateCommand) String() string {
	return strings.TrimSpace(command.Program + " " + strings.Join(command.Arguments, " "))
}

// RunOptions is everything Run needs to compare the two gates over one tree.
type RunOptions struct {
	// Root is the tree both gates lint, and the directory paths are made relative to.
	Root string
	// Cohere and Gate are the two invocations.
	Cohere GateCommand
	Gate   GateCommand
	// CohereRules is every rule name compiled into cohere, and ConfiguredRules every rule the lint
	// config enables. Both are read by the caller rather than inferred here, for the reason stated
	// on Inputs: a rule that works and finds nothing is invisible in the output.
	CohereRules     map[string]bool
	ConfiguredRules map[string]bool
	// Controls are the planted violations that prove the pipeline can carry a difference. Running
	// with none is allowed and produces a report that says, in as many words, that it proved
	// nothing — rather than a clean one.
	Controls []Control
	// ExtraCohereRuleSettings are rules to enable for cohere's run only, merged over the tree's own
	// lint config and handed to cohere with `-lint-config`. The gate always runs against the
	// unmodified configuration.
	//
	// This exists for one narrow purpose and it is worth stating so nobody widens it casually.
	// A directional control needs a rule cohere can report and the gate structurally cannot, and
	// the shared config does not enable such a rule, so cohere would run it over no files and the
	// control would miss for a configuration reason rather than a harness one.
	//
	// Asymmetric configuration is otherwise exactly what this instrument must never do: comparing
	// two gates under different rule sets manufactures differences that say nothing about either
	// implementation. So the asymmetry is confined to rules the gate cannot express at all, where
	// there is no shared setting to diverge from, and every finding it produces is classified
	// not-ported rather than counted against agreement.
	ExtraCohereRuleSettings map[string]any
}

// Control is a violation planted where the harness knows in advance what should happen to it.
//
// The file is written into the tree, both gates run, and the harness checks that the expected
// outcome came out the far end. That is an end-to-end proof: the process launched, the output
// parsed, the paths normalized, the rule names collapsed across two decoration schemes, and the
// comparison ran. Any one of those failing silently makes the control miss.
//
// There are two shapes, and they prove different things.
//
// A one-sided control (ExpectedSide set) proves the harness can report a difference in a named
// direction. It is the stronger claim and it requires a real divergence to exist between the two
// gates, which cannot be manufactured on demand.
//
// A shared control (ExpectedShared) proves the pipeline carries a finding end to end and that both
// normalizations worked, without claiming anything about direction. It is the weaker claim, and it
// is honest about being weaker: ControlsProven ignores shared controls entirely, so a run carrying
// only these still reports that detection was never demonstrated.
type Control struct {
	// Name identifies the control in the report.
	Name string
	// RelativePath is where the file is written, relative to Root. It must be a path both gates
	// actually lint: a control placed somewhere the config ignores misses for a reason that has
	// nothing to do with the harness, and looks identical to a broken pipeline.
	RelativePath string
	// Contents is the file to write.
	Contents string
	// Rule is the rule expected to fire.
	Rule string
	// ExpectedSide is the gate that should see it alone. Ignored when ExpectedShared is set.
	ExpectedSide Side
	// ExpectedShared says both gates must report this finding at the same place, so it appears as a
	// match rather than as a difference. Set for a rule both sides implement and the config enables,
	// where a one-sided result would be the defect rather than the proof.
	ExpectedShared bool
	// ExpectsNothing marks a file that is planted only so another control's file has something to
	// resolve against. It asserts no finding of its own and is not evidence of anything.
	//
	// Declared explicitly rather than inferred from an empty Rule, because "asserts nothing" and
	// "somebody forgot to say what this asserts" are different states that must not share a
	// spelling. A control that quietly asserts nothing is a control that always passes, which is
	// the failure this package exists to refuse.
	ExpectsNothing bool
}

// Run executes both gates over the tree, plants any controls, and returns the compared report.
//
// The order is deliberate: controls are planted first so that both gates see them in the same
// single run. Running the gates twice, once clean and once planted, would double the cost and
// introduce a window in which the tree changed underneath the two runs — and on a tree with several
// authors in it, that window is not theoretical.
func Run(ctx context.Context, options RunOptions) (Report, error) {
	planted, cleanup, err := plantControls(options)
	// cleanup is deferred before the error check on purpose: plantControls may have written some
	// files before failing, and leaving a planted violation behind in a shared tree would be a
	// defect handed to whoever runs the gate next.
	defer cleanup()
	if err != nil {
		return Report{}, err
	}

	cohereCommand, removeConfig, err := cohereCommandWithExtraRules(options)
	defer removeConfig()
	if err != nil {
		return Report{}, err
	}

	verifyOutput, cohereErr := runGate(ctx, cohereCommand)
	if cohereErr != nil {
		return Report{}, fmt.Errorf("running cohere: %w", cohereErr)
	}
	gateOutput, gateErr := runGate(ctx, options.Gate)
	if gateErr != nil {
		return Report{}, fmt.Errorf("running the gate: %w", gateErr)
	}

	cohereParsed, err := ParseCohere(verifyOutput, options.Root)
	if err != nil {
		return Report{}, fmt.Errorf("parsing cohere output: %w", err)
	}
	gateParsed, err := ParseGate(gateOutput, options.Root)
	if err != nil {
		return Report{}, fmt.Errorf("parsing gate output: %w", err)
	}

	// An unparsed line is a finding the diff cannot see, so it is a hard failure rather than a
	// note. Continuing past it would compare two partial lists and report the difference between
	// them as though it were the difference between the gates.
	if len(cohereParsed.UnparsedLines) > 0 {
		return Report{}, fmt.Errorf(
			"cohere printed %d lines this harness could not parse, so its findings are incomplete; first: %q",
			len(cohereParsed.UnparsedLines), cohereParsed.UnparsedLines[0],
		)
	}
	if len(gateParsed.UnparsedLines) > 0 {
		return Report{}, fmt.Errorf(
			"the gate printed %d lines this harness could not parse, so its findings are incomplete; first: %q",
			len(gateParsed.UnparsedLines), gateParsed.UnparsedLines[0],
		)
	}

	cohereFilesWalked, cohereRulesRun := cohereCoverageFrom(cohereParsed.SummaryLines)
	// The gate prints no coverage line of its own, so its population is the file count cohere
	// reported for the same tree. That is a borrowed number and it is marked as such here rather
	// than presented as the gate's own claim: it establishes that the tree was non-empty, which is
	// the property the vacuity guard needs, and nothing more.
	gateFilesWalked := cohereFilesWalked

	report := Compare(Inputs{
		CohereFindings:   cohereParsed.Findings,
		GateFindings:     gateParsed.Findings,
		CoherePopulation: Population{Findings: len(cohereParsed.Findings), FilesWalked: cohereFilesWalked, Rules: cohereRulesRun},
		GatePopulation:   Population{Findings: len(gateParsed.Findings), FilesWalked: gateFilesWalked},
		CohereRules:      options.CohereRules,
		ConfiguredRules:  options.ConfiguredRules,
		// The known gate defects are compiled in rather than passed at the command line: an
		// acknowledgement that can be supplied per-run can be supplied by whoever wants a green
		// result, and this list is reviewed like any other code.
		Acknowledged:   KnownGateDefects,
		ReadSourceLine: sourceLineReaderAt(options.Root),
	})

	report.Provenance = Provenance{
		CohereFilesLinted: cohereFilesWalked,
		GateFilesLinted:   gateFilesWalked,
		CohereRulesRun:    cohereRulesRun,
		GateRulesRun:      len(options.ConfiguredRules),
		CohereCommand:     options.Cohere.String(),
		CohereVersion:     cohereVersionOf(ctx, options.Cohere),
		GateCommand:       options.Gate.String(),
		ControlsRun:       checkControls(planted, report, cohereParsed.Findings, gateParsed.Findings),
	}

	// Marked after the comparison rather than during it, because Compare must stay unaware of
	// controls: a comparison that knew which findings were planted could special-case them, and the
	// controls exist precisely to test the path a real finding takes.
	markPlantedDifferences(&report, planted)

	return report, nil
}

// checkControls asks, for each planted violation, whether the comparison actually reported it.
//
// This is the end-to-end assertion and the reason the whole file exists. It deliberately looks in
// report.Differences rather than in either gate's raw findings: a control that one gate found but
// the comparison dropped is exactly the bug worth catching, and checking the raw output instead
// would sail straight past it.
func checkControls(planted []Control, report Report, cohereFindings []Finding, gateFindings []Finding) []ControlResult {
	results := make([]ControlResult, 0, len(planted))
	for _, control := range planted {
		result := ControlResult{
			Name:         control.Name,
			ExpectedSide: control.ExpectedSide,
			Rule:         control.Rule,
		}

		// A support file is planted, cleaned up, and never counted. It contributes no evidence, so
		// it is dropped from the results entirely rather than recorded as a passing control: a
		// result that always says "detected" would inflate the count of controls that fired and
		// make a run look better proven than it is.
		if control.ExpectsNothing {
			continue
		}

		if control.ExpectedShared {
			results = append(results, checkSharedControl(control, result, cohereFindings, gateFindings))
			continue
		}

		for _, difference := range report.Differences {
			if !controlMatches(control, difference.Finding) {
				continue
			}
			if difference.OnlyOn != control.ExpectedSide {
				result.Detail = fmt.Sprintf("found, but attributed to %s rather than %s", difference.OnlyOn, control.ExpectedSide)
				break
			}
			result.Detected = true
			result.Detail = ""
			break
		}

		if !result.Detected && result.Detail == "" {
			result.Detail = fmt.Sprintf("no difference reported for %s in %s", control.Rule, control.RelativePath)
		}
		results = append(results, result)
	}
	return results
}

// checkSharedControl asks whether both gates reported the planted finding.
//
// It reads the raw per-side findings rather than the comparison's output, because the property
// being tested is that the finding survived each side's parse and normalization and then matched.
// Asking the difference list could only ever say the control was absent from it, which is true both
// when both sides saw it and when neither did — the two outcomes this control exists to separate.
func checkSharedControl(control Control, result ControlResult, cohereFindings []Finding, gateFindings []Finding) ControlResult {
	seenByCohere := false
	for _, finding := range cohereFindings {
		if controlMatches(control, finding) {
			seenByCohere = true
			break
		}
	}
	seenByGate := false
	for _, finding := range gateFindings {
		if controlMatches(control, finding) {
			seenByGate = true
			break
		}
	}

	switch {
	case seenByCohere && seenByGate:
		result.Detected = true
	case !seenByCohere && !seenByGate:
		result.Detail = fmt.Sprintf("neither gate reported %s in %s, so the plant never reached either one", control.Rule, control.RelativePath)
	case seenByCohere:
		result.Detail = fmt.Sprintf("only cohere reported %s in %s, and both were expected to", control.Rule, control.RelativePath)
	default:
		result.Detail = fmt.Sprintf("only the gate reported %s in %s, and both were expected to", control.Rule, control.RelativePath)
	}
	return result
}

// controlMatches is whether a finding is the specific one a control planted.
//
// Suffix rather than exact path, because the two gates print paths differently and the parser
// normalizes them against the tree root, so the stored path is relative while the control names a
// relative path of its own.
//
// The conjunction is load-bearing and both halves have to stay. Dropping the path check leaves a
// rule check that answers correctly on this tree only because every rule a control declares happens
// to appear in no other file. That is a property of today's controls rather than of this function,
// and it holds until the first planted rule also fires somewhere real, at which point a finding
// gets attributed to a control that did not plant it. A mutation that removed the path half was
// read as equivalent because the totals matched; the totals matched for the wrong reason.
//
// This is deliberately narrower than markPlantedDifferences, which asks only whether a finding
// landed in a planted file. The two questions are different: this one verifies a named control
// fired, and a control that fired somewhere else has not fired. That one asks whether the harness
// caused a finding at all, where the rule is irrelevant because the file would not exist otherwise.
func controlMatches(control Control, finding Finding) bool {
	if finding.Rule != control.Rule {
		return false
	}
	return strings.HasSuffix(filepath.ToSlash(finding.File), filepath.ToSlash(control.RelativePath))
}

// plantControls writes each control into the tree and returns a cleanup that removes them.
//
// It refuses to overwrite an existing file. On a tree with several authors in it, clobbering
// somebody's work to run a self-test would be a worse defect than the one the self-test is looking
// for, and a control that silently replaced a real file would also be linted for the wrong reasons.
func plantControls(options RunOptions) ([]Control, func(), error) {
	planted := make([]Control, 0, len(options.Controls))
	written := make([]string, 0, len(options.Controls))
	// createdDirectories is the directories this plant brought into existence, deepest first, so
	// cleanup can unwind exactly what it made. Removing the file and leaving its directory behind
	// is still a mutation of a tree several people are working in, and it shows up in their
	// `git status` as an unexplained empty directory. An earlier run of this harness did precisely
	// that and the leftover was found by someone else.
	createdDirectories := make([]string, 0, len(options.Controls))

	cleanup := func() {
		for _, path := range written {
			// A control left behind would be linted by the next person to run either gate, so a
			// failure to remove one is worth saying out loud rather than swallowing.
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "differential: could not remove planted control %s: %v\n", path, err)
			}
		}
		// Deepest first, and only the ones this plant created. Remove refuses on a non-empty
		// directory, which is the guard that matters: if anything else landed in there while the
		// gates ran, it stays and so does the directory.
		for index := len(createdDirectories) - 1; index >= 0; index-- {
			if err := os.Remove(createdDirectories[index]); err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "differential: left the directory %s in place: %v\n", createdDirectories[index], err)
			}
		}
	}

	for _, control := range options.Controls {
		absolutePath := filepath.Join(options.Root, control.RelativePath)
		if _, err := os.Stat(absolutePath); err == nil {
			return planted, cleanup, fmt.Errorf(
				"control %q would overwrite the existing file %s; choose a path that does not exist",
				control.Name, control.RelativePath,
			)
		}
		missing, err := missingDirectories(filepath.Dir(absolutePath), options.Root)
		if err != nil {
			return planted, cleanup, fmt.Errorf("planning the directory for control %q: %w", control.Name, err)
		}
		if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
			return planted, cleanup, fmt.Errorf("creating the directory for control %q: %w", control.Name, err)
		}
		createdDirectories = append(createdDirectories, missing...)
		if err := os.WriteFile(absolutePath, []byte(control.Contents), 0o644); err != nil {
			return planted, cleanup, fmt.Errorf("writing control %q: %w", control.Name, err)
		}
		written = append(written, absolutePath)
		planted = append(planted, control)
	}

	return planted, cleanup, nil
}

// runGate executes one side and returns its stdout.
//
// A non-zero exit is expected and is not an error: both gates exit 1 when they find something, which
// is the common case. Only a failure to run at all is an error — and it must be one, because a gate
// that never launched produces empty output that reads exactly like a clean result.
func runGate(ctx context.Context, command GateCommand) (string, error) {
	execution := exec.CommandContext(ctx, command.Program, command.Arguments...)
	execution.Dir = command.Directory

	// Stderr is captured and surfaced rather than discarded. `exec.Output` drops it, and a gate
	// that crashed mid-run also exits non-zero, so the crash is indistinguishable from a gate that
	// simply found something. Its explanation would be on stderr and nowhere else.
	//
	// This is the same trap that cost three people an evening in a different costume: a probe
	// whose loud failure was silenced, leaving an empty result that read as an answer. The rule is
	// to never suppress stderr on something whose empty output you intend to trust, and the gate's
	// stdout is exactly that.
	var standardError strings.Builder
	execution.Stderr = &standardError

	output, err := execution.Output()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return "", fmt.Errorf("%s did not run: %w%s", command, err, stderrTail(standardError.String()))
		}

		// It ran and exited non-zero, which is what a gate with findings does. But a gate that
		// produced no findings at all and still failed did not find nothing, it broke, and those
		// two are the same exit code with different stdout.
		if strings.TrimSpace(string(output)) == "" {
			return "", fmt.Errorf(
				"%s exited %d and printed nothing, so it failed rather than found nothing%s",
				command, exitError.ExitCode(), stderrTail(standardError.String()),
			)
		}
		return string(output), nil
	}
	return string(output), nil
}

// stderrTail renders a failed gate's stderr for an error message, or nothing when it stayed quiet.
//
// Trimmed to the last few lines because a crashing linter can emit a great deal, and the part that
// says what happened is at the end. The whole point is that this text reaches a person: a gate
// failure diagnosed from an exit code alone sends the reader to the harness rather than to the
// gate.
func stderrTail(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) > maximumStderrLines {
		lines = lines[len(lines)-maximumStderrLines:]
	}
	return "; its stderr ended with: " + strings.Join(lines, " | ")
}

// maximumStderrLines is how much of a failed gate's stderr reaches the error message.
const maximumStderrLines = 5

// cohereCoverageLinePattern reads the population out of cohere's own coverage line:
//
//	lint: 1 findings — 23 rules over 3407 files, 2098302 nodes visited, in 813ms
var cohereCoverageLinePattern = regexp.MustCompile(`^lint: \d+ findings? — (\d+) rules? over (\d+) files?`)

// cohereCoverageFrom reads files-walked and rules-run out of cohere's summary lines.
//
// Returning zeros when the line is absent is deliberate and is not a fallback. A missing coverage
// line means cohere did not finish its lint phase, and zero is precisely the value that makes the
// vacuity guard refuse the comparison — which is the correct outcome for a run that did not
// complete. Defaulting to a plausible number here would be the vacuous pass this package exists to
// prevent, introduced by the file that reports the evidence.
func cohereCoverageFrom(summaryLines []string) (int, int) {
	for _, line := range summaryLines {
		match := cohereCoverageLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			continue
		}
		rulesRun, _ := strconv.Atoi(match[1])
		filesWalked, _ := strconv.Atoi(match[2])
		return filesWalked, rulesRun
	}
	return 0, 0
}

// missingDirectories lists the directories from root down to directory that do not exist yet.
//
// Asked before MkdirAll rather than inferred afterward, because afterward every directory on the
// path exists and there is no way to tell which ones were already there. Removing a directory the
// tree already had would be a worse mutation than leaving one behind.
//
// The walk stops at root and never above it, so a control cannot cause a directory outside the
// tree being linted to be removed. Ordered shallowest first; the caller unwinds in reverse.
func missingDirectories(directory string, root string) ([]string, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absoluteDirectory, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}

	var missing []string
	for current := absoluteDirectory; strings.HasPrefix(current, absoluteRoot) && current != absoluteRoot; {
		if _, err := os.Stat(current); err == nil {
			break
		}
		missing = append(missing, current)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	// Reverse into shallowest-first order, which is the order they get created in.
	for left, right := 0, len(missing)-1; left < right; left, right = left+1, right-1 {
		missing[left], missing[right] = missing[right], missing[left]
	}
	return missing, nil
}

// cohereCommandWithExtraRules gives cohere a config carrying the extra rules, when there are any.
//
// The merged config is written beside the tree's own rather than over it, and removed afterward, so
// the gate and every other process in this worktree keep reading the unmodified file. Editing the
// real config in place would change what a sibling's `s c` does mid-run, which on a tree with
// several authors is a defect rather than a shortcut.
//
// With no extra rules this returns the command untouched, so the ordinary path allocates nothing
// and writes nothing.
func cohereCommandWithExtraRules(options RunOptions) (GateCommand, func(), error) {
	noCleanup := func() {}
	if len(options.ExtraCohereRuleSettings) == 0 {
		return options.Cohere, noCleanup, nil
	}

	// Deliberately `.oxlintrc.json` rather than cohere's own `CohereSettings.json`. This harness
	// runs the real oxlint binary to compare against, and oxlint reads its own config and would
	// reject ours: `CohereSettings.json` enables rules oxlint does not implement, `no-octal` among
	// them, and an unknown rule name makes oxlint refuse to start rather than skip the line. The
	// two files split for that reason and this side must keep pointing at oxlint's.
	configPath := filepath.Join(options.Root, ".oxlintrc.json")
	existing, err := os.ReadFile(configPath)
	if err != nil {
		return options.Cohere, noCleanup, fmt.Errorf("reading the lint config to extend it: %w", err)
	}

	var document map[string]any
	if err := json.Unmarshal(existing, &document); err != nil {
		return options.Cohere, noCleanup, fmt.Errorf("parsing the lint config to extend it: %w", err)
	}

	rules, _ := document["rules"].(map[string]any)
	if rules == nil {
		rules = map[string]any{}
	}
	for name, setting := range options.ExtraCohereRuleSettings {
		// Refusing rather than overwriting. A rule the shared config already configures is one the
		// gate may also run, so overriding it here would compare the two sides under different
		// settings for a rule they both have, which is the asymmetry this must never introduce.
		if _, alreadyConfigured := rules[name]; alreadyConfigured {
			return options.Cohere, noCleanup, fmt.Errorf(
				"rule %q is already configured in the tree's lint config, so enabling it only for cohere would compare the two gates under different settings",
				name,
			)
		}
		rules[name] = setting
	}
	document["rules"] = rules

	extended, err := json.Marshal(document)
	if err != nil {
		return options.Cohere, noCleanup, fmt.Errorf("encoding the extended lint config: %w", err)
	}

	// Beside the original, because the config's own directory is what relative ignore patterns and
	// override globs resolve against. A config in a temporary directory elsewhere would silently
	// change which files those patterns match.
	extendedPath := filepath.Join(options.Root, ".oxlintrc.differential.json")
	if _, err := os.Stat(extendedPath); err == nil {
		return options.Cohere, noCleanup, fmt.Errorf(
			"%s already exists, so a differential run is in flight or one was interrupted; not overwriting it", extendedPath,
		)
	}
	if err := os.WriteFile(extendedPath, extended, 0o644); err != nil {
		return options.Cohere, noCleanup, fmt.Errorf("writing the extended lint config: %w", err)
	}

	cleanup := func() {
		if err := os.Remove(extendedPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "differential: could not remove %s: %v\n", extendedPath, err)
		}
	}

	command := options.Cohere
	command.Arguments = append(append([]string{}, command.Arguments...), "-lint-config", ".oxlintrc.differential.json")
	return command, cleanup, nil
}

// cohereVersionOf asks the binary that ran what it is.
//
// Asked of the binary rather than derived from the source tree beside it, because those are exactly
// the two things that drift apart: a cached or previously-built binary sits at a path whose name
// says nothing about which rules are compiled into it. A run at 03:22 reported 128 disagreements
// that were entirely a binary predating the rule they were about.
//
// A failure here is recorded as unknown rather than raised. The version is context for a reader,
// not an input to any verdict, and refusing to compare because a binary declined to introduce
// itself would be the harness failing over something that changes nothing about the comparison.
func cohereVersionOf(ctx context.Context, command GateCommand) string {
	probe := command
	probe.Arguments = []string{"-version"}

	output, err := runGate(ctx, probe)
	if err != nil {
		return "version unknown: " + err.Error()
	}

	return versionLineFrom(output)
}

// versionLineFrom picks the line of `cohere -version` output that identifies the build.
//
// The first line is the version name, and on a local build it is the constant "cohere dev" for
// every binary ever compiled from this tree. Taking it produces a field that is always populated,
// always plausible, and never distinguishes anything, which is worse than an empty one because it
// looks like provenance. The first attempt at this did exactly that, and two binaries eleven
// commits apart rendered identically while producing different results.
//
// The commit is what identifies which rules are compiled in, so the line naming it is what is kept.
//
// That is the `commit:` line, which names the binary's own commit. The `compiler:` line's fallback,
// "unknown (built from cohere <sha>)", named it too, but only on a build that did not stamp its
// compiler, and reading it was reading the right number off the wrong line. The launcher's
// committed-tree builds stamp the compiler, so that fallback disappeared and every global binary
// rendered as "cohere dev" (measured 2026-10-02 on 7b434bb4105b). It is still read when there is no
// `commit:` line, for binaries that predate it.
func versionLineFrom(output string) string {
	lines := strings.Split(output, "\n")

	commit := ""
	compilerFallback := ""
	note := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if commit == "" && strings.HasPrefix(trimmed, "commit:") {
			commit = trimmed
			continue
		}
		if compilerFallback == "" && strings.Contains(trimmed, "built from") {
			compilerFallback = trimmed
			continue
		}
		// A local build says so in a `note:` line, and the note is what qualifies the commit. A
		// dirty-tree build's note says its rules are whatever was on disk when it compiled; a clean
		// committed build's note says its commit reproduces it. Either way it has to travel with the
		// commit, because the commit alone reads as provenance: it is a real sha, it resolves, and it
		// says nothing about whether uncommitted rules were linked in beside it. The match is on the
		// `note:` prefix rather than the wording, so a reworded note is still carried.
		//
		// Three people quoted a rule count from a dev build in one night, each correct about the
		// binary and wrong about what they said it measured, with this note one command away from
		// every invocation they made. A disclosure that has to be sought is a disclosure that gets
		// skipped, so it is carried rather than left to be looked up.
		if note == "" && strings.HasPrefix(trimmed, "note:") {
			note = strings.TrimSpace(strings.TrimPrefix(trimmed, "note:"))
		}
	}

	if commit == "" {
		commit = compilerFallback
	}
	if commit != "" {
		if note != "" {
			return commit + " — " + note
		}
		return commit
	}

	// A release build states its version on the first line instead of naming a commit, so that is
	// the fallback rather than the preference.
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return "version unknown: the binary printed nothing"
}

// markPlantedDifferences flags the differences this run caused rather than observed.
//
// Matched by file rather than by rule, and that distinction was a defect for one revision. Matching
// on the control's declared rule leaves any OTHER rule that fires on the same planted file counted
// as observed, and one did: the cohere-only control's file trips the gate's import-ordering rule as
// well as the rule it was written for, so the summary read "1 observed, 2 planted" when both were
// this harness's own footprint.
//
// A file this run wrote did not exist a moment ago, so every finding in it is the harness looking
// at itself, whatever rule produced it. Support files are included for exactly that reason: they
// assert nothing, but a rule may still fire on them, and such a finding is no more observed than
// any other.
func markPlantedDifferences(report *Report, planted []Control) {
	for index := range report.Differences {
		for _, control := range planted {
			if sharesPath(control.RelativePath, report.Differences[index].Finding.File) {
				report.Differences[index].Planted = true
				break
			}
		}
	}
}

// sharesPath reports whether a finding landed in the file a control planted.
func sharesPath(controlPath string, findingPath string) bool {
	return strings.HasSuffix(filepath.ToSlash(findingPath), filepath.ToSlash(controlPath))
}

// ObservedDifferences returns the differences that are about the codebase rather than about the
// harness, which is the number a reader means when they ask how far apart the two gates are.
func (report Report) ObservedDifferences() []Difference {
	observed := make([]Difference, 0, len(report.Differences))
	for _, difference := range report.Differences {
		if !difference.Planted {
			observed = append(observed, difference)
		}
	}
	return observed
}

// sourceLineReaderAt reads lines of files under a root, each file read once, which is how an
// acknowledgement's anchor is checked against the finding it names.
func sourceLineReaderAt(root string) SourceLineReader {
	linesByFile := map[string][]string{}
	return func(file string, line int) (string, bool) {
		lines, read := linesByFile[file]
		if !read {
			content, err := os.ReadFile(filepath.Join(root, file))
			if err == nil {
				lines = strings.Split(string(content), "\n")
			}
			linesByFile[file] = lines
		}
		if line < 1 || line > len(lines) {
			return "", false
		}
		return lines[line-1], true
	}
}
