// Package fix applies the repairs rules propose, and refuses the ones that would corrupt a file.
//
// This is the only component in cohere that writes to source. Everything else reads: the parser
// reads bytes, the checker reads the parse, a rule reads the tree and proposes. A proposal is
// inert until this package decides it lands, which is what lets a rule be written by anyone and
// still be safe — a rule cannot damage a file it has no way to write to.
//
// The gate this replaces ran oxlint in check mode even during `lint fix`, and its own comment said
// why: an unaudited fixer writing across every file is the widest blast radius in the tool. That
// was avoidance rather than a solution. The solution is four guards, each answering a way a fixer
// silently corrupts source rather than failing:
//
//   - Overlap. Two rules want the same bytes. One wins by a stated rule, or neither lands.
//   - Offsets. The first applied fix shifts every later one. Applying back to front means no
//     surviving fix is ever measured against text that moved under it.
//   - Refusal. The rewritten text is parsed before it is written, and discarded if it does not
//     parse. This is the guard that makes autofix safe at all.
//   - Atomicity. Temp file plus rename, so a reader sees the old file or the new one and never a
//     half-written one. Five agents work in this tree at once.
//
// The safety property the rest rests on: a fix here lands in a working tree, is read in review,
// and becomes a commit. A bad rewrite is visible before it ships. That is the reason refusal can
// be the only hard stop rather than one of many.
package edit

import (
	"sort"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// Proposal is one rule's suggested edit, carrying who proposed it.
//
// A Fix on its own does not say which rule wanted it, and every decision this package makes needs
// to name a rule: an overlap report that says "two fixes collided at byte 400" tells a reader
// nothing they can act on, while one that names both rules tells them which pair to look at.
type Proposal struct {
	RuleName string
	Fix      rule.Fix
}

// Rejection is a proposal that did not land, and why.
//
// Refusals are reported rather than dropped. A fixer that silently declines half its work looks
// exactly like one that had less to do, and this project's history is a list of times that
// ambiguity cost days.
type Rejection struct {
	Proposal Proposal
	Reason   string

	// ConflictsWith names the rule whose fix won the bytes, when the reason is an overlap. Empty
	// otherwise.
	//
	// It can equal the losing proposal's own RuleName. A diagnostic reporting two fixes has them
	// flattened into independent proposals, so a rule's fixes compete with each other exactly as
	// they compete with another rule's, and a reader seeing the same name on both sides is looking
	// at that case rather than at a bug.
	ConflictsWith string

	// BreaksParseAlone is set on a parse failure's refusal when this fix, applied by itself to the text
	// its pass started from, already does not parse. It names the fix that broke the file. When no fix of
	// a refused pass breaks it alone, the pass's fixes broke it together, and none is marked.
	BreaksParseAlone bool
}

// Reasons a proposal is rejected. Stated as constants because they are read by tests and printed
// in reports, and a reason that drifts between the two stops being checkable.
const (
	// ReasonOverlap says "another fix" rather than "another rule" because resolveOverlaps is purely
	// positional: it tracks the span last claimed and never looks at who proposed it. When one
	// diagnostic reports two fixes, ProposalsFrom has already flattened them into siblings, so the
	// winner can be the same rule as the loser. Saying "another rule" sent a reader hunting for a
	// second rule that does not exist. ConflictsWith carries the accurate answer either way.
	//
	// It stays a bare constant rather than naming the winner inline, and the reason is not that a
	// longer string would break the tally: reasonKey prefix-matches, so `ReasonOverlap + " from X"`
	// would collapse correctly. ReasonParseFailure is the precedent for when detail belongs in the
	// string, and its comment says why — the compiler's message is actionable and has nowhere else
	// to live. An overlap has no such payload. ConflictsWith is already a structured field holding
	// the identity, so assembling it into the prose would duplicate a field into a string that a
	// prefix-matcher then strips back off. Detail belongs in the reason only when no field carries it.
	ReasonOverlap       = "overlaps another fix"
	ReasonInvalidRange  = "range is outside the file or inverted"
	ReasonParseFailure  = "the rewritten file does not parse"
	ReasonNoProgress    = "the fix replaces text with itself"
	ReasonPassesReached = "the pass budget was exhausted before the file converged"

	// ReasonTransformFailed is the whole-text transform erroring rather than producing bad output.
	// Distinct from a parse failure: one means the formatter broke, the other means it produced
	// something that is not a program.
	ReasonTransformFailed = "the whole-text transform failed"

	// ReasonFormatFixUnsettled is a file whose formatted text kept tripping a fixer, and whose fixed
	// text kept changing under the formatter, through every round FixAndTransformText allows: a rule
	// and the printer undoing each other. The file is left as it was found.
	ReasonFormatFixUnsettled = "fix and format did not settle"
)

// transformRuleName is what a transform's refusals are attributed to.
//
// A transform has no rule behind it, and an empty name in a refusal reads as a bug in the reporting
// rather than as a fact about the run.
const transformRuleName = "format"

// Plan is a resolved set of edits: what will be applied to a file, and what was refused.
//
// Applied is ordered by position ascending, which is the order a reader expects even though the
// edits are performed back to front.
type Plan struct {
	Applied  []Proposal
	Rejected []Rejection
}

// resolveOverlaps decides which proposals may be applied together.
//
// The rule is deterministic and stated rather than emergent: sort by start position, then by end
// position, then by rule name, and keep a proposal only if it begins at or after the end of the
// last one kept. Everything else is refused with the winner named.
//
// Determinism is the whole requirement. Two rules fighting over the same bytes is a real
// disagreement about what the code should say, and no ordering resolves it correctly — but an
// ordering that varies run to run turns one disagreement into a file that changes shape depending
// on map iteration order, which is far worse than picking the earlier one and saying so. The
// refusal is reported, so a pair of rules that fight constantly shows up as a number rather than
// as a mystery.
//
// A pure insertion (an empty range) is a real edit at a point and is treated as occupying no
// bytes, so two insertions at the same point do not both land — the second is refused rather than
// silently interleaved with the first, because the order they would apply in is exactly the
// ambiguity this function exists to remove.
func resolveOverlaps(proposals []Proposal) Plan {
	ordered := make([]Proposal, len(proposals))
	copy(ordered, proposals)
	sort.SliceStable(ordered, func(first, second int) bool {
		firstFix, secondFix := ordered[first].Fix, ordered[second].Fix
		if firstFix.Range.Pos() != secondFix.Range.Pos() {
			return firstFix.Range.Pos() < secondFix.Range.Pos()
		}
		if firstFix.Range.End() != secondFix.Range.End() {
			return firstFix.Range.End() < secondFix.Range.End()
		}
		return ordered[first].RuleName < ordered[second].RuleName
	})

	plan := Plan{}

	// lastEnd is the first byte not yet claimed. It starts below zero rather than at zero so a
	// legitimate insertion at position zero is not mistaken for an overlap with nothing.
	lastEnd := -1
	lastRuleName := ""

	for _, proposal := range ordered {
		start, end := proposal.Fix.Range.Pos(), proposal.Fix.Range.End()

		if start > lastEnd || (start == lastEnd && start != end && lastEnd >= 0) {
			plan.Applied = append(plan.Applied, proposal)
			lastEnd = end
			lastRuleName = proposal.RuleName
			continue
		}

		// A fix that starts exactly where the previous one ended is adjacent rather than
		// overlapping, and adjacent edits compose safely. The case above admits it; this one
		// catches the genuine collisions, including two insertions at one point.
		if start >= lastEnd && start != end {
			plan.Applied = append(plan.Applied, proposal)
			lastEnd = end
			lastRuleName = proposal.RuleName
			continue
		}

		plan.Rejected = append(plan.Rejected, Rejection{
			Proposal:      proposal,
			Reason:        ReasonOverlap,
			ConflictsWith: lastRuleName,
		})
	}

	return plan
}

// applyToText performs a resolved plan against source text, back to front.
//
// Back to front is the whole trick, and it is worth stating why rather than trusting the reader to
// reconstruct it. Every fix carries offsets measured against the text as the rule saw it. Applying
// the earliest fix first changes the length of everything after it, so every later fix is then
// pointing at text that moved — by a delta that depends on all the edits before it. Applying from
// the end means each edit only ever disturbs bytes that no remaining fix refers to, so no offset
// is ever recomputed and there is no arithmetic to get wrong. This is exactly where an off-by-one
// corrupts a file quietly instead of failing loudly.
//
// Ranges are validated against the text here rather than trusted, because a rule computes them
// against the AST it was handed and a caller can hand this function text from a different pass.
func applyToText(text string, plan Plan) (string, Plan) {
	validated := Plan{Rejected: plan.Rejected}

	for _, proposal := range plan.Applied {
		start, end := proposal.Fix.Range.Pos(), proposal.Fix.Range.End()
		if start < 0 || end < start || end > len(text) {
			validated.Rejected = append(validated.Rejected, Rejection{
				Proposal: proposal,
				Reason:   ReasonInvalidRange,
			})
			continue
		}
		if text[start:end] == proposal.Fix.Text {
			// A fix that replaces text with itself is not an error, but counting it as applied
			// would make a run that changed nothing report as a run that changed something, and
			// would keep a convergence loop spinning forever on a rule that proposes it every pass.
			validated.Rejected = append(validated.Rejected, Rejection{
				Proposal: proposal,
				Reason:   ReasonNoProgress,
			})
			continue
		}
		validated.Applied = append(validated.Applied, proposal)
	}

	rewritten := text
	for index := len(validated.Applied) - 1; index >= 0; index-- {
		fix := validated.Applied[index].Fix
		rewritten = rewritten[:fix.Range.Pos()] + fix.Text + rewritten[fix.Range.End():]
	}

	return rewritten, validated
}

// Resolve turns a set of proposals into a plan without applying it.
//
// Exposed so a caller can see what would land — `--no-fix` reporting, and the tests that prove the
// overlap rule picks who it says it picks.
func Resolve(proposals []Proposal) Plan {
	return resolveOverlaps(proposals)
}

// ProposalsFrom collects the fixes out of a set of diagnostics.
//
// Suggestions are deliberately not collected. The split is intent rather than confidence: a fix
// preserves what the code means and may be applied unattended, while a suggestion changes it and
// needs a human to agree. Removing an await on a non-promise is a fix; filling in a missing switch
// case is a suggestion. A fixer that quietly applied suggestions would be changing behavior
// without anyone choosing it.
//
// The fixes of one diagnostic are flattened into independent proposals, and a rule author writing a
// multi-fix report should know what that costs. Two fixes reported together are not treated as one
// atomic edit: overlap resolution judges them separately, so it may admit one and refuse the other.
// For a pair that only makes sense together — wrapping a span in an opening and a closing token, the
// obvious example — that is a half-application, and the result can be well-formed enough to parse
// while meaning something nobody proposed.
//
// No shipped rule does this today. Every ReportNodeWithFixes site emits exactly one fix, so the
// hazard has no live caller and this is a warning rather than a bug. A rule that genuinely needs two
// edits to land together needs grouping here first; it cannot get it by reporting them side by side.
//
// The door this used to come through is now closed. adaptFixes in internal/rules/upstream carried an
// upstream fix slice of any length into one diagnostic, so an adapted rule could reach this hazard
// without anyone writing one of ours. That adapter and every rule that went through it are gone: the
// six tsgolint rules were absorbed onto our own interface and each one's fixes are now written here,
// in this repo, by someone who can read this comment. So the next multi-fix report will be
// deliberate, which is the condition under which grouping should be built rather than warned about.
func ProposalsFrom(diagnostics []rule.Diagnostic) []Proposal {
	proposals := []Proposal{}
	for _, diagnostic := range diagnostics {
		for _, singleFix := range diagnostic.Fixes {
			proposals = append(proposals, Proposal{RuleName: diagnostic.RuleName, Fix: singleFix})
		}
	}
	return proposals
}
