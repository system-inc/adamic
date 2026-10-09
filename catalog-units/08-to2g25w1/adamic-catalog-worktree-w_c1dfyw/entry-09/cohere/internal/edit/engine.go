package edit

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// DefaultMaxPasses bounds convergence.
//
// A fix can expose a violation that was not visible before it landed, so one pass is not enough and
// a loop with no ceiling is not an option: two rules that undo each other would spin forever, on a
// tool whose entire premise is finishing in a quarter second. Ten is chosen to be far above what a
// healthy file needs and far below anything a person would wait through. A file that reaches the
// ceiling is reported rather than silently left half-fixed, because a file needing ten passes is
// telling you something about the rules rather than about the file.
const DefaultMaxPasses = 10

// FileResult is what the engine did to one file.
//
// Passes and Applied are separate numbers on purpose. A file fixed in one pass and a file fixed in
// six have the same diff and very different meanings — the second says two rules are arguing, or a
// rule proposes one repair at a time where it could propose all of them.
type FileResult struct {
	FileName string

	// Text is the file's contents after fixing. Equal to the original when nothing landed.
	Text string

	// Changed is whether anything actually landed. A file whose every proposal was refused is
	// unchanged, and that is different from a file that had nothing proposed.
	Changed bool

	// Passes is how many times the engine re-ran rules against this file, counting the pass that
	// found nothing left to do.
	Passes int

	// Applied is every fix that landed, across every pass.
	Applied []Proposal

	// Rejected is every proposal that did not, with a reason. Refusals are reported rather than
	// dropped: a fixer that silently declines half its work looks identical to one with less to do.
	Rejected []Rejection

	// Converged is whether the file reached a pass with nothing left to apply. False means the pass
	// budget ran out, and the file is left exactly as it was found: a fixpoint the engine did not
	// reach is not a result it may write.
	Converged bool

	// UnconvergedRules names the rules still proposing on the pass that exhausted the budget.
	//
	// The file is the symptom and the rule is the defect. A rule proposing a fix that does not
	// silence its own finding will be re-proposed every pass until the ceiling, so naming the file
	// alone sends a reader to look at code that is not wrong. Empty whenever Converged is true.
	UnconvergedRules []string

	// Transformed is whether the whole-text transform changed anything after the fixes converged.
	//
	// Reported separately from Changed because a formatter that silently did nothing and a file that
	// needed no formatting produce identical output otherwise, which is the same ambiguity the
	// coverage line exists to destroy.
	Transformed bool

	// TransformSkipped is whether the transform declined this file on purpose.
	//
	// A skip and a file that needed no formatting produce identical output otherwise, and they mean
	// opposite things: one says the formatter never looked, the other says it looked and approved.
	// Collapsing them is how a formatter with a known corruption bug on one file type comes to read
	// as a formatter that had nothing to do.
	TransformSkipped bool

	// TransformSkipReason is why, in the transform's own words. Empty when it did not say.
	TransformSkipReason string

	// Failed is set when the engine could not process the file at all — unreadable, unwritable, or
	// already unparseable before anything was applied.
	//
	// Distinct from Changed being false, which is a file the engine looked at and correctly left
	// alone. A zero-value FileResult is indistinguishable from a clean file, so a caller that
	// recorded a failure without setting this would report the failure as a success.
	Failed bool
}

// Propose is asked for the fixes that apply to a file's current text.
//
// The engine re-asks after every pass rather than reusing the first answer, and that is not an
// optimization to remove later — it is the correctness requirement. Offsets from the previous pass
// are measured against text that no longer exists, so reusing them is exactly the silent corruption
// this package exists to prevent. Re-running rules against the rewritten text is how every surviving
// proposal is guaranteed to be measured against the bytes it will be applied to.
type Propose func(fileName string, text string) ([]Proposal, error)

// FixText drives one file to a fixpoint in memory, without touching disk.
//
// The loop is: ask for proposals against the current text, resolve overlaps, apply back to front,
// parse the result, and keep it only if it parses. Repeat until a pass proposes nothing that lands
// or the budget runs out.
//
// Refusal is per pass and whole. When the rewritten text does not parse, the entire pass is
// discarded and the previous text stands — not the subset of fixes that were individually fine.
// Bisecting to find which fix broke it would be a nicer result and a worse guarantee: it means
// writing text assembled from a combination no rule proposed and nothing verified. The file is left
// alone and every proposal in the pass is reported as refused, so the reader learns which file and
// which rules to look at.
func FixText(fileName string, text string, propose Propose, maxPasses int) (FileResult, error) {
	result, _, err := fixText(fileName, text, nil, propose, maxPasses)
	return result, err
}

// fixText is FixText, also returning the guard's tree of the text it settled on: the starting text's, or
// the last accepted pass's. Nil where the guard built none (a file the TypeScript parser does not own), and
// for a file whose passes ran out and were discarded.
//
// seed is a tree of text someone already parsed, or nil: the starting guard takes it rather than parsing where
// it is the tree the guard would build (seedFor), and every later guard parses as always.
//
// One tree is held at a time. Keeping the starting tree through every pass, for the rare file that runs out
// and reverts, held two whole trees and a third being built on a file with fixes, and a 2.8 MB bundle is a
// file with fixes (#xn1k1gz). The formatter parses a reverted file again, as it did before it was handed
// trees at all. Found by @system_cohere_lint_fix in review.
func fixText(fileName string, text string, seed *ast.SourceFile, propose Propose, maxPasses int) (FileResult, *ast.SourceFile, error) {
	if maxPasses <= 0 {
		maxPasses = DefaultMaxPasses
	}

	result := FileResult{FileName: fileName, Text: text}

	// The starting text must parse, or nothing downstream means anything. A file that is already
	// broken is not this package's problem to report — the types phase does that, loudly — but it is
	// this package's problem not to make worse, and "the result parses" is a guarantee that says
	// nothing if the input did not.
	var parses bool
	var reason string
	var currentTree *ast.SourceFile
	if seedFor(fileName, text, seed) {
		parses, reason, currentTree = verdictOf(seed)
	} else {
		parses, reason, currentTree = parsesWithTree(fileName, text)
	}
	if !parses {
		return result, nil, fmt.Errorf("%s does not parse before any fix is applied (%s)", fileName, reason)
	}

	current := text

	// lastProposals holds the most recent pass's proposals so that a run which exhausts the budget
	// can name the rules still proposing at the ceiling rather than only the file they landed on.
	lastProposals := []Proposal{}

	for pass := 1; pass <= maxPasses; pass++ {
		result.Passes = pass

		proposals, err := propose(fileName, current)
		if err != nil {
			return result, nil, fmt.Errorf("collecting fixes for %s on pass %d: %w", fileName, pass, err)
		}
		lastProposals = proposals
		if len(proposals) == 0 {
			result.Converged = true
			break
		}

		plan := resolveOverlaps(proposals)
		rewritten, plan := applyToText(current, plan)

		result.Rejected = append(result.Rejected, plan.Rejected...)

		if len(plan.Applied) == 0 {
			// Everything proposed was refused. Re-running would produce the same refusals forever,
			// so this is a fixpoint even though findings remain.
			result.Converged = true
			break
		}

		// Let go of the last pass's tree before building this one's. A pass refused below leaves the file
		// with no tree, and the formatter parses it, as it did before it was handed one.
		currentTree = nil
		parses, reason, rewrittenTree := parsesWithTree(fileName, rewritten)
		if !parses {
			for _, proposal := range plan.Applied {
				result.Rejected = append(result.Rejected, Rejection{
					Proposal:         proposal,
					Reason:           fmt.Sprintf("%s (%s)", ReasonParseFailure, reason),
					BreaksParseAlone: breaksParseAlone(fileName, current, proposal),
				})
			}
			// The pass is discarded whole and the loop stops. Trying again would re-collect the same
			// proposals against the same text and refuse them the same way.
			result.Converged = true
			break
		}

		current = rewritten
		currentTree = rewrittenTree
		result.Applied = append(result.Applied, plan.Applied...)
		result.Changed = true
	}

	if !result.Converged {
		// The budget ran out with work still landing, so the whole run is discarded and the file is
		// left as it was found.
		//
		// The earlier behavior kept the last pass that parsed, on the reasoning that it was a valid
		// file. Parsing is not the same property as being finished: a rule that does not silence its
		// own finding gets applied once per pass, and ten passes of a brace-wrapping fix produce
		// `{ { { { ... } } } }`, which parses cleanly and is nobody's code. Valid was carrying the
		// weight of acceptable.
		//
		// The failure directions are not symmetric, which is what decides it. Refusing costs a run
		// where autofix did nothing and said why. Writing costs a developer a file worse than they
		// left it, carrying a rejection line beside it that a hurried review passes over — and a
		// mangled file that parses survives review far more easily than an unfixed one does.
		result.Rejected = append(result.Rejected, Rejection{
			Proposal: Proposal{RuleName: "fix-engine"},
			Reason:   fmt.Sprintf("%s (%d passes)", ReasonPassesReached, maxPasses),
		})
		result.UnconvergedRules = ruleNamesOf(lastProposals)
		result.Applied = nil
		result.Changed = false
		result.Text = text
		return result, nil, nil
	}

	result.Text = current
	return result, currentTree, nil
}

// breaksParseAlone reports whether one fix, applied by itself to the text its pass started from, already
// leaves a file that does not parse.
//
// Asked only of a pass the guard refused, to name the fix that broke it, and the text it builds is read and
// dropped. The refusal stays whole, as FixText's comment says: this diagnoses the pass and never writes a
// combination no rule proposed.
func breaksParseAlone(fileName string, text string, proposal Proposal) bool {
	alone, plan := applyToText(text, Plan{Applied: []Proposal{proposal}})
	if len(plan.Applied) == 0 {
		return false
	}
	parses, _ := Parses(fileName, alone)
	return !parses
}

// ruleNamesOf reduces proposals to the distinct rules behind them, in first-seen order.
//
// Order is deterministic rather than sorted so the rule that proposed first is named first, which
// is usually the one a reader wants. Distinct because a rule proposing forty fixes on the final
// pass is one broken rule, not forty.
func ruleNamesOf(proposals []Proposal) []string {
	seen := map[string]bool{}
	names := []string{}
	for _, proposal := range proposals {
		if seen[proposal.RuleName] {
			continue
		}
		seen[proposal.RuleName] = true
		names = append(names, proposal.RuleName)
	}
	return names
}

// FixFile drives a file to a fixpoint and writes it, atomically, only if something landed.
//
// Reading the file here rather than taking the text from the already-parsed program is deliberate.
// The program's copy was read when the graph was built, and in a tree where several agents edit at
// once that copy can be minutes stale — applying offsets from a stale parse to a file that moved
// underneath is precisely the corruption shape this package exists to refuse. Re-reading costs one
// syscall and removes the whole class.
//
// A file with nothing to apply is not written at all. Rewriting identical bytes would update the
// modification time, which invalidates every downstream cache keyed on it and makes a run that
// changed nothing look like a run that changed everything.
func FixFile(fileName string, propose Propose, maxPasses int) (FileResult, error) {
	return FixAndTransformFile(fileName, propose, nil, maxPasses)
}

// Transform is a whole-text rewrite applied after fixes have converged — formatting, in practice.
//
// It is deliberately not a Propose. A proposal is a claim to some bytes, and the engine arbitrates
// between competing claims; a transform is the new value of all of them, and there is nothing to
// arbitrate. Feeding a whole-document rewrite through resolveOverlaps asks that machinery a question
// it was not built to answer, and both available answers are wrong: either a correctness fix loses
// to a whitespace change because the rule names sort that way, or the formatting is silently skipped
// wherever a fix touched. Neither is a tradeoff anyone would choose, and a priority field would only
// pick the wrong one on purpose.
//
// Order is fix-then-transform and it is not arbitrary. Fixes are semantic and change what the
// correct formatting is, so formatting first would leave every later fix mis-formatted and want a
// re-format, which is a loop. A total, idempotent transform running last converges in one shot
// regardless of what the fixes did.
//
// Identified by @system_cohere_format, which read this package and found the failure mode before
// anything was built against it.
//
// parsed is a tree of text that the engine's own guard already built, or nil. A transform that parses with
// typescript-go may take it rather than parse the same bytes again, which every formatted TypeScript file
// did: 0.18 GB a cold ahra run (#dk2502g). It is the guard's tree, parsed with the guard's script kind (JS
// for a .js file, where the formatter parses TSX), so a transform takes it only where it is the tree it
// would build, and a wrapper passes it through untouched.
type Transform func(fileName string, text string, parsed *ast.SourceFile) (string, error)

// ErrSkipped is what a transform returns to decline a file rather than to fail on it.
//
// A transform that cannot handle a file type, or handles it incorrectly today, must be able to say
// so in a way the engine records rather than swallows. Returning the text unchanged would work and
// would be wrong: it is indistinguishable from "already correctly formatted", so a formatter with a
// known corruption bug on markdown would report the whole tree as clean.
//
// Wrap it to carry a reason, which is what reaches the coverage line:
//
//	if isMarkdown(fileName) {
//		return "", fmt.Errorf("%w: markdown doubles a standalone tilde", fix.ErrSkipped)
//	}
var ErrSkipped = errors.New("the transform skipped this file")

// skipReasonOf pulls a transform's own words out of a skip error, without the sentinel's text.
//
// The reason is what a reader acts on. "Skipped" alone tells them a number; "skipped: markdown
// doubles a standalone tilde" tells them which files and why, which is the difference between a
// count and a finding.
func skipReasonOf(err error) string {
	message := err.Error()
	prefix := ErrSkipped.Error() + ": "
	if after, found := strings.CutPrefix(message, prefix); found {
		return after
	}
	if message == ErrSkipped.Error() {
		return ""
	}
	return message
}

// FixAndTransformFile drives a file to a fixpoint, applies a whole-text transform, and writes the
// result — once, atomically, behind the same parse guard.
//
// One phase, one parse guard, one write. Two components that both wrote the same file in one run is
// precisely the failure this package exists to prevent, so formatting arrives here rather than
// writing on its own.
//
// The transform runs even when no fix landed, because a file can be correctly written and badly
// formatted. It is guarded exactly as a fix pass is: a transform that errors, or whose output does
// not parse, is discarded whole and the file is left alone.
func FixAndTransformFile(fileName string, propose Propose, transform Transform, maxPasses int) (FileResult, error) {
	result, err := CheckFile(fileName, propose, transform, maxPasses)
	if err != nil {
		return result, err
	}

	// Changed alone is the right condition here, and it is deliberately not joined by Converged.
	//
	// A run that exhausted the pass budget has already cleared Changed and reverted Text upstream, in
	// FixText, so a non-converged file cannot reach this line with anything to write. Re-checking
	// Converged would be a second gate on a decision already made, and two gates on one property is
	// how they drift apart. Said here because reading this function alone makes the guard look
	// absent: two readers concluded exactly that from this line, and the guard is one branch above.
	if !result.Changed {
		return result, nil
	}

	if err := WriteAtomically(fileName, result.Text); err != nil {
		// The file on disk is untouched: the rename is the only step that changes it, and a failure
		// before it leaves the original intact. Report the fixes as not-applied rather than applied,
		// so the count matches what a reader would find in the tree.
		result.Changed = false
		return result, fmt.Errorf("writing %s: %w", fileName, err)
	}

	return result, nil
}

// CheckFile is FixAndTransformFile without the write: the file is read now, driven through the same
// fixpoint and transform, and left untouched, with what a run would write in FileResult.Text.
//
// It is `--no-fix`'s answer to "what would change". Every decision is the writing run's own, reached
// through the same function the writing run calls before it writes, so a file this reports as
// unchanged is a file `--fix` would leave alone, and the two cannot drift apart.
func CheckFile(fileName string, propose Propose, transform Transform, maxPasses int) (FileResult, error) {
	return CheckFileSeeded(fileName, nil, propose, transform, maxPasses)
}

// CheckFileSeeded is CheckFile given a tree of the file that someone already parsed, the program's, or nil.
// The file is still read now, and the tree is taken only if it is of exactly the bytes read and is the tree
// the guard would build (seedFor), so a tree of a stale copy, or of another script kind, is parsed past as if
// it were not there. Taken, it is the starting guard's tree and the transform's (#dk2502g).
func CheckFileSeeded(fileName string, seed *ast.SourceFile, propose Propose, transform Transform, maxPasses int) (FileResult, error) {
	text, err := readFile(fileName)
	if err != nil {
		return FileResult{FileName: fileName}, err
	}
	return fixAndTransformText(fileName, text, seed, propose, transform, maxPasses)
}

// FixAndTransformText is FixAndTransformFile on text in hand, writing nothing: the fixpoint, then the
// transform behind the same parse guard, with the result in FileResult.Text.
//
// An editor formatting a buffer before it saves needs exactly what a run would write for that file,
// computed from text that is not on disk yet. Splitting the write off rather than copying the rest is
// what keeps the two answers the same answer.
//
// The transform's output is linted again whenever it changed the text, because a printer can write
// what a rule repairs. Breaking a one-line block-bodied arrow across lines is the case that found it:
// 51 arrows in 23 of www-phi-health's files met consistency-no-multiline-arrow-function, whose fix then
// waited for the next run, so one `--fix --format` left files its own `--no-fix` check flagged. When a
// fixer fires on the printed text, the fixes land and the result is formatted again, up to
// FormatFixRoundLimit rounds. A file still changing after that is left exactly as it was found and
// reported as not converged, naming the rules and the formatter, never written half-settled.
func FixAndTransformText(fileName string, text string, propose Propose, transform Transform, maxPasses int) (FileResult, error) {
	return fixAndTransformText(fileName, text, nil, propose, transform, maxPasses)
}

// CheckTextSeeded is CheckFileSeeded on text the caller already holds, read from nothing: the caller has shown
// the text is the file's, as the format speculation does when the program's copy still matches the file's
// stat (#q6dey77). The seed is taken under the same test as CheckFileSeeded's, against this text.
func CheckTextSeeded(fileName string, text string, seed *ast.SourceFile, propose Propose, transform Transform,
	maxPasses int) (FileResult, error) {
	return fixAndTransformText(fileName, text, seed, propose, transform, maxPasses)
}

// fixAndTransformText is FixAndTransformText with a seed for the starting guard (fixText).
func fixAndTransformText(fileName string, text string, seed *ast.SourceFile, propose Propose, transform Transform,
	maxPasses int) (FileResult, error) {
	result, tree, err := fixText(fileName, text, seed, propose, maxPasses)
	if err != nil || transform == nil {
		return result, err
	}
	if !result.Converged {
		// The fixes ran out of passes and were discarded, so the transform formats the text as found,
		// once, as it always has. Re-linting what it printed would only meet the same arguing rules.
		applyTransform(&result, fileName, transform, tree)
		return result, nil
	}

	for round := 1; ; round++ {
		changed, printedTree := applyTransform(&result, fileName, transform, tree)
		if !changed {
			return result, nil
		}
		tree = printedTree

		// Only a file the TypeScript parser reads has rules to ask. A markdown, css or json file the
		// printer changed has nothing to re-lint, and handing it to the TypeScript parser crashed the
		// run on the first `.md` a writing run formatted.
		if !TypeScriptParsable(fileName) {
			return result, nil
		}

		// The printer changed the text, so ask the rules about what it printed.
		proposals, err := propose(fileName, result.Text)
		if err != nil {
			return result, fmt.Errorf("collecting fixes for %s after formatting: %w", fileName, err)
		}
		if len(proposals) == 0 {
			return result, nil
		}

		if round == FormatFixRoundLimit {
			// Still changing at the bound: two of them, a rule and the printer, are undoing each
			// other. Writing the last round would call a file settled that the next run rewrites.
			unsettled := Rejection{
				Proposal: Proposal{RuleName: transformRuleName},
				Reason:   fmt.Sprintf("%s (%d rounds)", ReasonFormatFixUnsettled, FormatFixRoundLimit),
			}
			return FileResult{
				FileName:         fileName,
				Text:             text,
				Passes:           result.Passes,
				Rejected:         append(result.Rejected, unsettled),
				UnconvergedRules: append(ruleNamesOf(proposals), transformRuleName),
			}, nil
		}

		// The proposals in hand were measured against exactly this text, so the next fixpoint starts
		// from them rather than linting the same text a second time.
		used := false
		startingWith := func(name string, current string) ([]Proposal, error) {
			if !used {
				used = true
				return proposals, nil
			}
			return propose(name, current)
		}
		next, nextTree, err := fixText(fileName, result.Text, nil, startingWith, maxPasses)
		if err != nil {
			return result, err
		}
		result.Passes += next.Passes
		result.Rejected = append(result.Rejected, next.Rejected...)
		if !next.Converged {
			// The fix budget ran out on the printed text, which leaves the file as found, as FixText
			// does for a file that never reached formatting.
			return FileResult{
				FileName:         fileName,
				Text:             text,
				Passes:           result.Passes,
				Rejected:         result.Rejected,
				UnconvergedRules: next.UnconvergedRules,
			}, nil
		}
		if next.Changed {
			result.Text = next.Text
			result.Applied = append(result.Applied, next.Applied...)
			tree = nextTree
		}
	}
}

// FormatFixRoundLimit bounds how many times a file goes back through fix and format when the
// printer's output trips a fixer. The case that found it settles in two rounds, a fix and then a
// format the fix leaves nothing for; a third that still changes something means a rule and the
// printer are undoing each other.
const FormatFixRoundLimit = 3

// applyTransform runs the transform once over result.Text and records what happened, reporting
// whether it changed the text, and the guard's tree of the text it printed. A skip, a failure, a result
// that does not parse, and text already in the transform's shape all leave the text as it was and report
// false.
//
// tree is the guard's tree of result.Text, or nil, and is offered to the transform only while it is still
// that text's.
func applyTransform(result *FileResult, fileName string, transform Transform, tree *ast.SourceFile) (bool, *ast.SourceFile) {
	if tree != nil && tree.Text() != result.Text {
		tree = nil
	}
	transformed, transformError := transform(fileName, result.Text, tree)
	switch {
	case errors.Is(transformError, ErrSkipped):
		// The transform declined this file on purpose — a file type it does not handle, or one it
		// handles incorrectly today. Recorded rather than treated as a failure, because the two
		// mean different things to a reader: a failure says the formatter broke, a skip says it
		// chose not to look, and a run where a formatter skipped four hundred files must not read
		// as a run where four hundred files were already correctly formatted.
		result.TransformSkipped = true
		result.TransformSkipReason = skipReasonOf(transformError)
		return false, nil

	case transformError != nil:
		// A transform that failed says nothing about the fixes, which already converged and
		// already passed the guard. They are kept and the transform is reported as refused, so a
		// formatter that is broken today does not also block every correctness fix in the tree.
		result.Rejected = append(result.Rejected, Rejection{
			Proposal: Proposal{RuleName: transformRuleName},
			Reason:   fmt.Sprintf("%s (%s)", ReasonTransformFailed, transformError),
		})
		return false, nil

	case transformed == result.Text:
		// Already in the shape the transform wants. Not an error and not a change.
		return false, nil
	}

	// The guard applies to the transform exactly as it applies to a fix pass. A whole-text
	// rewrite has a wider blast radius than any single fix, so it earns the check more, not
	// less.
	parses, reason, printedTree := parsesWithTree(fileName, transformed)
	if !parses {
		result.Rejected = append(result.Rejected, Rejection{
			Proposal: Proposal{RuleName: transformRuleName},
			Reason:   fmt.Sprintf("%s (%s)", ReasonParseFailure, reason),
		})
		return false, nil
	}
	result.Text = transformed
	result.Transformed = true
	result.Changed = true
	return true, printedTree
}

// Summary is what a whole fix run did, across every file.
//
// Every field here answers a question a bare "fixed 40 files" cannot. A run that changed forty
// files must not look like a run that changed none, and a run that refused four hundred fixes must
// not look like a run that had none to refuse.
type Summary struct {
	FilesConsidered int
	FilesChanged    int
	FixesApplied    int
	FixesRefused    int

	// FilesTransformSkipped is how many files the transform declined on purpose.
	//
	// Printed alongside the reformat count, because a run where the formatter skipped a whole file
	// type must not read as a run where that file type was already clean.
	FilesTransformSkipped int

	// TransformSkipReasons counts skips by the reason the transform gave, so a reader learns which
	// file types were left alone and why rather than only how many.
	TransformSkipReasons map[string]int

	// FilesTransformed is how many files the whole-text transform rewrote after fixes converged.
	//
	// Counted separately from FilesChanged because the two answer different questions. A run where
	// formatting silently did nothing and a run where nothing needed formatting produce the same
	// FilesChanged, and telling those apart is the same discipline that makes the coverage line
	// worth printing at all.
	FilesTransformed int

	// RefusalsByReason counts refusals by their stated reason, so "we refused four hundred fixes"
	// becomes "three hundred and ninety were overlaps between two rules and ten broke the parse",
	// which is the difference between a number and a finding.
	RefusalsByReason map[string]int

	// FilesByPasses counts files by how many passes they took. A tail here is the signal that some
	// rule proposes one repair where it could propose several, or that two rules are arguing.
	FilesByPasses map[int]int

	// FilesNotConverged names the files that hit the pass ceiling. Named rather than counted,
	// because the answer to "which file needed ten passes" is the only useful next step.
	//
	// A file here was left untouched on disk. Hitting the ceiling discards the run for that file.
	FilesNotConverged []string

	// UnconvergedRules names every rule still proposing when some file hit the ceiling, distinct
	// across the run.
	//
	// This is the actionable half and the file list is not: a rule whose fix does not silence its
	// own finding will do it on every file it matches, so a reader given only filenames goes and
	// reads code that is not the problem.
	UnconvergedRules []string

	// FilesRefused names the files where a pass was discarded because the rewrite did not parse, with the
	// rule that broke each. These are the interesting failures: a rule proposed something that looked fine
	// and was not, which is a bug in that rule's fix.
	FilesRefused []RefusedFile

	// FilesFailed names the files the engine could not process at all — unreadable, unwritable, or
	// already unparseable before anything was applied.
	//
	// Separate from FilesRefused and FilesNotConverged because the three send a reader somewhere
	// different. Refused means a rule proposed something wrong. Not converged means two rules are
	// arguing. Failed means the file never entered the loop, and folding it into either of the
	// others reports a real number under the wrong heading — which reads as a finding about the
	// rules when it is actually a finding about the file.
	FilesFailed []string

	// ChangedFiles names every file counted in FilesChanged, with what changed it, sorted by name.
	//
	// The count says how much; this says where and why, which is what `--no-fix` reports as findings:
	// a file that would change is a finding against that file, and the reader needs the rule or the
	// formatter that would change it to know whether to run the fixer or read the code.
	ChangedFiles []ChangedFile

	// Checked is set when the run computed what it would write and wrote nothing, so the summary line
	// says "would be rewritten" rather than claiming a rewrite that never happened.
	Checked bool

	// NotTransformed names every file the transform declined or failed on, with its reason, sorted by
	// name. The counts above say how many; a caller whose verdict is the transform alone needs to know
	// which, since a file the formatter could not read is a file nobody checked.
	NotTransformed []NotTransformedFile
}

// RefusedFile is one file whose rewrite did not parse, so the engine wrote nothing to it.
//
// The count alone ("1 refused (1 the rewritten file does not parse)") named neither the file nor the rule,
// even under --verbose, so a broken fix could not be traced to its rule (#v1ah2qq).
type RefusedFile struct {
	FileName string

	// Reason is the refusal with the parser's own message.
	Reason string

	// Breaking names the rules whose fix, applied alone, already does not parse: the defect. Empty when
	// only the pass's fixes together broke it, and then Rules is where to look.
	Breaking []string

	// Rules names every rule with a fix in the refused pass, or "format" when the whole-text transform's
	// output did not parse.
	Rules []string
}

// NotTransformedFile is one file the whole-text transform left as it was without having shaped it.
type NotTransformedFile struct {
	FileName string

	// Reason is the transform's own words for a skip, or the failure it reported.
	Reason string

	// Failed is a transform that broke rather than declined.
	Failed bool
}

// ChangedFile is one file a run rewrote, or under `--no-fix` would rewrite, and what rewrote it.
type ChangedFile struct {
	FileName string

	// Changers names the fixing rules whose repairs landed, distinct and in the order they first
	// landed, followed by "format" when the whole-text transform changed the result.
	Changers []string

	// FixesByRule counts the repairs that landed, by rule, so a reader of `prefer-const ×2` learns how
	// much each rule changed and not only that it did. The transform is not in it; it is in Changers.
	FixesByRule map[string]int
}

// Summarize folds per-file results into a run summary.
func Summarize(results []FileResult) Summary {
	summary := Summary{
		RefusalsByReason:     map[string]int{},
		FilesByPasses:        map[int]int{},
		TransformSkipReasons: map[string]int{},
	}

	unconvergedRuleSeen := map[string]bool{}

	for _, result := range results {
		summary.FilesConsidered++
		if result.Changed {
			summary.FilesChanged++
			changers := ruleNamesOf(result.Applied)
			if result.Transformed {
				changers = append(changers, transformRuleName)
			}
			fixesByRule := map[string]int{}
			for _, proposal := range result.Applied {
				fixesByRule[proposal.RuleName]++
			}
			summary.ChangedFiles = append(summary.ChangedFiles, ChangedFile{FileName: result.FileName, Changers: changers, FixesByRule: fixesByRule})
		}
		if result.Transformed {
			summary.FilesTransformed++
		}
		if result.TransformSkipped {
			summary.FilesTransformSkipped++
			summary.TransformSkipReasons[result.TransformSkipReason]++
			summary.NotTransformed = append(summary.NotTransformed, NotTransformedFile{FileName: result.FileName, Reason: result.TransformSkipReason})
		}
		summary.FixesApplied += len(result.Applied)
		summary.FixesRefused += len(result.Rejected)
		summary.FilesByPasses[result.Passes]++

		if result.Failed {
			summary.FilesFailed = append(summary.FilesFailed, result.FileName)
		} else if !result.Converged {
			// Only a file that actually entered the loop can be said to have not converged.
			summary.FilesNotConverged = append(summary.FilesNotConverged, result.FileName)
			for _, ruleName := range result.UnconvergedRules {
				if !unconvergedRuleSeen[ruleName] {
					unconvergedRuleSeen[ruleName] = true
					summary.UnconvergedRules = append(summary.UnconvergedRules, ruleName)
				}
			}
		}

		var refused *RefusedFile
		for _, rejection := range result.Rejected {
			reason := reasonKey(rejection.Reason)
			summary.RefusalsByReason[reason]++
			if reason == ReasonParseFailure {
				if refused == nil {
					summary.FilesRefused = append(summary.FilesRefused, RefusedFile{FileName: result.FileName, Reason: rejection.Reason})
					refused = &summary.FilesRefused[len(summary.FilesRefused)-1]
				}
				if !slices.Contains(refused.Rules, rejection.Proposal.RuleName) {
					refused.Rules = append(refused.Rules, rejection.Proposal.RuleName)
				}
				if rejection.BreaksParseAlone && !slices.Contains(refused.Breaking, rejection.Proposal.RuleName) {
					refused.Breaking = append(refused.Breaking, rejection.Proposal.RuleName)
				}
			}
			if rejection.Proposal.RuleName == transformRuleName {
				summary.NotTransformed = append(summary.NotTransformed, NotTransformedFile{FileName: result.FileName, Reason: rejection.Reason, Failed: true})
			}
		}
	}
	sort.SliceStable(summary.NotTransformed, func(first, second int) bool {
		return summary.NotTransformed[first].FileName < summary.NotTransformed[second].FileName
	})

	sort.Strings(summary.FilesNotConverged)
	sort.Slice(summary.FilesRefused, func(first, second int) bool {
		return summary.FilesRefused[first].FileName < summary.FilesRefused[second].FileName
	})
	sort.Strings(summary.FilesFailed)
	sort.Slice(summary.ChangedFiles, func(first, second int) bool {
		return summary.ChangedFiles[first].FileName < summary.ChangedFiles[second].FileName
	})
	return summary
}

// Refusals is a line for each file whose rewrite did not parse: the file, the rule to look at, and the
// parser's message. Printed under the summary line, which counts them, so the count can be traced.
func (s Summary) Refusals() []string {
	lines := make([]string, 0, len(s.FilesRefused))
	for _, refused := range s.FilesRefused {
		blame := fmt.Sprintf("the fix from %s breaks it", strings.Join(refused.Breaking, " and from "))
		switch {
		case len(refused.Rules) == 1 && refused.Rules[0] == transformRuleName:
			blame = "the formatter's output breaks it"
		case len(refused.Breaking) == 0:
			blame = fmt.Sprintf("no one fix breaks it alone, the fixes from %s together do", strings.Join(refused.Rules, ", "))
		}
		lines = append(lines, fmt.Sprintf("refused: %s: %s; nothing was written to it. %s", refused.FileName, refused.Reason, blame))
	}
	return lines
}

// reasonKey collapses a reason that carries detail down to its category.
//
// A parse failure reason embeds the compiler's message so a reader can act on it, which makes every
// one of them a distinct string and would turn a tally into a list of four hundred singletons.
func reasonKey(reason string) string {
	for _, known := range []string{ReasonParseFailure, ReasonOverlap, ReasonInvalidRange, ReasonNoProgress, ReasonPassesReached, ReasonFormatFixUnsettled,
		ReasonTransformFailed} {
		if len(reason) >= len(known) && reason[:len(known)] == known {
			return known
		}
	}
	return reason
}

// String renders the summary line that follows a fix run.
//
// It always prints, and it always prints the population alongside the result. The gate this tool
// replaces printed a green checkmark over zero files linted for days, because an empty file list is
// indistinguishable from a clean tree. The same ambiguity is available to a fixer, and this is what
// closes it.
func (s Summary) String() string {
	// A checked run states the same numbers in the conditional, because "3 files rewritten" over a tree
	// nobody wrote to is a count claiming work that never happened.
	rewritten, applied, reformatted := "rewritten", "applied", "reformatted"
	if s.Checked {
		rewritten, applied, reformatted = "would be rewritten", "would apply", "would be reformatted"
	}
	line := fmt.Sprintf(
		"fix: %d of %d files %s, %d fixes %s, %d refused",
		s.FilesChanged, s.FilesConsidered, rewritten, s.FixesApplied, applied, s.FixesRefused,
	)

	// The refusal breakdown attaches to the refusal count and must stay adjacent to it. It was
	// written when nothing sat between the two, and adding the reformat and skip clauses moved it
	// away — producing "2 not formatted (2 markdown) (1 overlaps another fix)", where
	// the second parenthetical reads as a second skip reason. A true number under the wrong heading,
	// found by reading the rendered line rather than by any assertion, because every fixture checked
	// for the presence of its own substring and none checked what the whole line said.
	if len(s.RefusalsByReason) > 0 {
		reasons := make([]string, 0, len(s.RefusalsByReason))
		for reason := range s.RefusalsByReason {
			reasons = append(reasons, reason)
		}
		sort.Strings(reasons)

		line += " ("
		for index, reason := range reasons {
			if index > 0 {
				line += ", "
			}
			line += fmt.Sprintf("%d %s", s.RefusalsByReason[reason], reason)
		}
		line += ")"
	}

	// Formatting reports its own number rather than hiding inside the rewrite count, so a run where
	// the formatter did nothing is distinguishable from one where nothing needed formatting.
	if s.FilesTransformed > 0 {
		line += fmt.Sprintf(", %d %s", s.FilesTransformed, reformatted)
	}

	// Skips print whenever there are any, and they name the reason. A formatter that declined four
	// hundred files and one that found four hundred files already correct produce the same reformat
	// count, and only this line tells them apart.
	if s.FilesTransformSkipped > 0 {
		line += fmt.Sprintf(", %d not formatted", s.FilesTransformSkipped)
		reasons := make([]string, 0, len(s.TransformSkipReasons))
		for reason := range s.TransformSkipReasons {
			if reason != "" {
				reasons = append(reasons, reason)
			}
		}
		sort.Strings(reasons)
		for _, reason := range reasons {
			line += fmt.Sprintf(" (%d %s)", s.TransformSkipReasons[reason], reason)
		}
	}

	maximumPasses := 0
	for passes := range s.FilesByPasses {
		if passes > maximumPasses {
			maximumPasses = passes
		}
	}
	if maximumPasses > 1 {
		line += fmt.Sprintf(", up to %d passes", maximumPasses)
	}

	if len(s.FilesNotConverged) > 0 {
		line += fmt.Sprintf(", %d files did not converge", len(s.FilesNotConverged))
	}

	// Failures print even though they were already written to stderr as they happened. A summary
	// that omits them reads as a clean run to anyone reading only the last line, which is most
	// readers most of the time.
	if len(s.FilesFailed) > 0 {
		line += fmt.Sprintf(", %d files could not be processed", len(s.FilesFailed))
	}

	return line
}
