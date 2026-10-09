package differential

import (
	"fmt"
	"strings"
)

// Provenance is what actually ran, recorded so that an empty difference list can be told apart
// from a harness that never looked.
//
// This is the whole reason the package exists rather than a shell script with a `diff` in it. The
// commissioning task put it plainly: two gates that both report zero, one because it checked and
// one because it ran on no files, produce identical output. Four vacuous probes were caught in one
// night, and every one of them produced a clean, confident, wrong answer. So the harness records
// its own population and refuses to render a verdict that its population does not support.
//
// The two guards below fail in opposite directions and a sweep needs both. FilesLinted catches a
// detector that was handed nothing to examine — the case a planted-violation control sails straight
// past, because a control file is itself a file and a harness linting only that one file would
// still flag it. ControlDetected catches a detector that cannot detect at all, which a healthy
// file count says nothing about.
//
// # What is proven, and how it got there
//
// Both directions now fire, against ~/Projects/ahra, through `cmd/cohere-differential`. Each rests
// on a structural asymmetry rather than a configurable one, which is what keeps a control from
// quietly ceasing to discriminate:
//
//	import-require-path-alias      cohere has it; the oxlint plugin does not define it at all
//	exhaustive-deps                the gate has it; cohere implements no react-hooks rule at all
//
// The gate-side control used to be `consistency-organize-imports`, on the ground that cohere did not
// implement it. Cohere implements it now, at `internal/rules/structure/consistency_organize_imports.go`,
// so that asymmetry is gone and this comment named a control that had already been replaced. The
// replacement is not a documentation detail: a control resting on "not ported yet" expires the moment
// somebody ports it, and the expiry is silent, because a control that has stopped discriminating still
// reports Detected against whichever side happens to flag first.
//
// `ControlsProven` was made to key on `ExpectedSide` rather than on a rule name for exactly this
// reason, so the port could not flip the verdict without anyone noticing. That predicate is why the
// stale name here cost nothing. It is still worth reading as the warning it is.
//
// The replacement's own ground, and the one caveat on it, are recorded at the control itself in
// `cmd/cohere-differential/main.go`. Read it before treating this control as permanent: it rests on
// nobody porting a 4,477-line rule, which is a judgment about effort rather than a fact about the
// tools, and the note there names the sturdier ground to move to if that stops holding.
//
// A third, shared control asserts a finding both sides must report. It proves the pipeline end to
// end without claiming direction, and ControlsProven ignores it deliberately so the weaker proof
// cannot pass as the stronger claim.
//
// This took a while to reach, and the history is worth keeping because the intermediate states all
// looked like completion. A run with a shared control alone reported a confident population and an
// empty diff; a run with the cohere-only control alone would have been half-blind in a way nothing
// in the output revealed, since it could miss everything the gate sees and cohere does not.
//
// The verdict itself has been shown able to go negative, which a clean result cannot establish: the
// react rule's line threshold was skewed from 60 to 90 in a throwaway build, and the report flipped
// to `disagrees` with the new differences classified both-active rather than not-ported.
type Provenance struct {
	// CohereFilesLinted and GateFilesLinted are how many files each side actually walked. A zero
	// here means the run proved nothing regardless of what the diff says.
	CohereFilesLinted int
	GateFilesLinted   int

	// CohereRulesRun and GateRulesRun are how many rules each side had loaded.
	CohereRulesRun int
	GateRulesRun   int

	// CohereCommand and GateCommand are the exact invocations, so a reader can rerun them.
	CohereCommand string
	GateCommand   string

	// CohereVersion is what the cohere binary says it is, captured from the binary that actually
	// ran rather than from the source tree beside it.
	//
	// This exists because of a real wrong answer. A run at 03:22 reported 128 disagreements on
	// `react-component-no-multiple-primary`; the rule had landed at 02:46 and the binary being
	// measured was built at 01:18, so the differences were a stale binary lacking a rule that had
	// already shipped. The same comparison against a binary built from head reported agreement on
	// all 128. Nothing in the output distinguished those two runs, because the invocation string is
	// a path and a path says nothing about what is inside it.
	//
	// A binary is the one input to this instrument that changes without leaving a trace in the
	// report, so its identity is recorded next to the numbers it produced.
	CohereVersion string

	// ControlsRun records the planted-violation controls that were exercised this run, if any.
	// Empty means the harness's ability to detect a difference was not demonstrated, which is a
	// materially weaker claim and is reported as such.
	ControlsRun []ControlResult
}

// ControlResult is one planted violation and whether the harness saw it.
//
// A control is a file with a known defect that exactly one side can see. Running it proves the
// pipeline end to end: the gate ran, the parse worked, the comparison ran, and a real difference
// came out the other side. Without it, "no differences" is a claim about the codebase that is
// indistinguishable from a claim about a broken harness.
type ControlResult struct {
	// Name identifies the control in the report.
	Name string

	// ExpectedSide is the gate that should see this violation. A control that only cohere can see
	// and a control that only the gate can see prove different halves of the pipeline, and a
	// harness that runs one direction only is half-proven — a parse bug on the unexercised side
	// would still report clean.
	ExpectedSide Side

	// Rule is the rule expected to fire.
	Rule string

	// Detected is whether the harness actually reported this as a difference found by
	// ExpectedSide.
	Detected bool

	// Detail is what the harness saw instead, when Detected is false.
	Detail string
}

// Trustworthy is whether this run's population supports drawing any conclusion from the diff.
//
// Deliberately strict, and deliberately not a judgment about the findings. A run over zero files
// is untrustworthy no matter how clean it looks, and a run whose controls did not fire is
// untrustworthy no matter how many files it walked.
func (provenance Provenance) Trustworthy() (bool, []string) {
	reasons := []string{}

	if provenance.CohereFilesLinted == 0 {
		reasons = append(reasons, "cohere linted 0 files, so its result is empty rather than clean")
	}
	if provenance.GateFilesLinted == 0 {
		reasons = append(reasons, "the gate linted 0 files, so its result is empty rather than clean")
	}
	if provenance.CohereRulesRun == 0 {
		reasons = append(reasons, "cohere ran 0 rules")
	}
	if provenance.GateRulesRun == 0 {
		reasons = append(reasons, "the gate ran 0 rules")
	}

	// A file count that is implausibly small is its own tell. The ahra tree is thousands of files;
	// a run reporting a handful walked something other than the tree, and that is the failure mode
	// where a clean answer is most convincing and most wrong.
	if provenance.CohereFilesLinted > 0 && provenance.CohereFilesLinted < minimumPlausibleFileCount {
		reasons = append(reasons, fmt.Sprintf(
			"cohere linted only %d files, which is too few to be the tree — check the directory it was pointed at",
			provenance.CohereFilesLinted,
		))
	}

	// No controls at all is its own failure, and it has to be named separately from a control that
	// ran and missed. Looping over an empty slice finds nothing wrong with it, so a run that never
	// planted anything would otherwise pass this guard vacuously — the exact shape of bug this
	// package exists to catch, one level up, inside the catcher. Found by a test, not by reading.
	if !provenance.ControlsProven() {
		reasons = append(reasons, "no control fired in both directions, so the harness has not been shown able to detect a difference")
	}

	for _, control := range provenance.ControlsRun {
		if !control.Detected {
			reasons = append(reasons, fmt.Sprintf(
				"control %q did not fire: %s — the harness has not been shown able to detect a difference",
				control.Name, control.Detail,
			))
		}
	}

	return len(reasons) == 0, reasons
}

// ControlsProven is whether both directions of detection were demonstrated.
//
// Reported separately from Trustworthy because they answer different questions. Trustworthy asks
// whether this run's population was real. This asks whether the harness was shown able to detect a
// difference at all — and both directions, because a harness that can only see one side's findings
// reports a clean diff for every defect on the other.
// A shared control leaves ExpectedSide empty and therefore satisfies neither direction, which is
// deliberate and load-bearing rather than incidental. A shared control proves the pipeline carries
// a finding end to end; it says nothing about whether a one-sided finding would survive, and
// letting it count here would turn the weaker proof into the stronger claim silently.
func (provenance Provenance) ControlsProven() bool {
	sawVerifyDirection := false
	sawGateDirection := false
	for _, control := range provenance.ControlsRun {
		if !control.Detected {
			return false
		}
		if control.ExpectedSide == SideCohere {
			sawVerifyDirection = true
		}
		if control.ExpectedSide == SideGate {
			sawGateDirection = true
		}
	}
	return sawVerifyDirection && sawGateDirection
}

// minimumPlausibleFileCount is the floor below which a run over this codebase is assumed to have
// been pointed at the wrong thing.
//
// A threshold rather than an exact expected count on purpose: the tree grows, and a harness that
// has to be edited every time a file lands is a harness that gets its guard commented out. The
// number only has to be high enough that a misconfigured run cannot slip under it, and low enough
// that it never fires on a real one.
const minimumPlausibleFileCount = 100

// Describe renders provenance for the report, as prose a reader can check.
func (provenance Provenance) Describe() string {
	builder := &strings.Builder{}

	fmt.Fprintf(builder, "cohere: %d files, %d rules — %s\n",
		provenance.CohereFilesLinted, provenance.CohereRulesRun, provenance.CohereCommand)
	if provenance.CohereVersion != "" {
		fmt.Fprintf(builder, "        %s\n", provenance.CohereVersion)
	}
	fmt.Fprintf(builder, "gate:   %d files, %d rules — %s\n",
		provenance.GateFilesLinted, provenance.GateRulesRun, provenance.GateCommand)

	if len(provenance.ControlsRun) == 0 {
		fmt.Fprintf(builder, "controls: none run — this harness has NOT been shown able to detect a difference on this run\n")
		return builder.String()
	}

	directions := map[Side]bool{}
	for _, control := range provenance.ControlsRun {
		status := "detected"
		if !control.Detected {
			status = "missed: " + control.Detail
		}
		// A shared control has no expected side, so naming one would be a lie and leaving the field
		// blank renders as `( should see ...)`. It says what it actually asserts instead.
		expectation := fmt.Sprintf("%s should see %s", control.ExpectedSide, control.Rule)
		if control.ExpectedSide == "" {
			expectation = fmt.Sprintf("both should see %s", control.Rule)
		} else {
			directions[control.ExpectedSide] = true
		}
		fmt.Fprintf(builder, "control %q (%s): %s\n", control.Name, expectation, status)
	}

	// The closing line says which proof is missing rather than always claiming one direction ran.
	// "only one direction exercised" printed over a run with no directional control at all was a
	// true-sounding sentence about something that had not happened.
	if !provenance.ControlsProven() {
		switch {
		case len(directions) == 0:
			fmt.Fprintf(builder, "controls: no directional control ran, so a one-sided difference in either direction would still report clean\n")
		case len(directions) == 1:
			for side := range directions {
				fmt.Fprintf(builder, "controls: only the %s direction was exercised, so a defect on the other side would still report clean\n", side)
			}
		default:
			fmt.Fprintf(builder, "controls: both directions were exercised but not every control fired\n")
		}
	}

	return builder.String()
}
