// Acknowledged differences are the ones we decided on purpose.
//
// The harness exists to catch a port that drifted from the gate, so every difference it finds is a
// defect in one of the two until someone says otherwise. But the gate is software, and software has
// bugs: when a rule is ported to its stated semantics and the gate cannot see a case it claims to
// enforce, the two disagree and cohere is the one that is right.
//
// "Cohere and the gate agree" is the acceptance test for a faithful port, not the definition of a
// correct rule. Those are the same thing right up until the gate has a bug, and then they are
// opposites. Without somewhere to record that, the only ways forward are to reproduce the bug so the
// numbers match, or to let the harness sit red and lose the signal it exists to give.
package differential

import (
	"fmt"
	"strings"
)

// AcknowledgedDifference is one finding we expect only one side to report, and why.
//
// # Keyed to a file, a rule, a side and an anchor, never to a line number
//
// Keyed to the one finding rather than to the rule alone. A rule-wide excuse would also swallow the
// next genuine drift in that rule, which is the failure this whole package is built to prevent: an
// excuse broad enough to be convenient is an excuse broad enough to hide a defect.
//
// The entries were first keyed by line, and a line is the wrong key for a table that lives in another
// repository from the code it describes. Any edit above a site moves it, so by 2026-10-04 most of the
// table named lines that no longer held its finding (Map.tsx 747, 765 and 784, Button.tsx 222, the
// loop sites at 305, 548 and 689, a misused spread in a file since renamed), and every such entry went
// on excusing whatever happened to land on that line next (#cn8sthd). So an entry names text from its
// finding's own source line instead, the enclosing declaration's name when it sits there or the
// expression the rule reports, which moves with the code and survives edits elsewhere in the file.
type AcknowledgedDifference struct {
	File string
	Rule string
	// Side is which gate is expected to report it alone.
	Side Side
	// Anchor is text on the line the finding sits on, compared against that line with its
	// indentation trimmed. It must pick out exactly one of the side's findings for this rule in this
	// file: an entry matching none excuses nothing and is reported stale, and one matching several is
	// reported ambiguous and excuses none of them, since it would otherwise excuse more than its
	// reason was written about.
	Anchor string
	// Reason is why this difference is correct, in a sentence a reader can check.
	//
	// Required rather than optional. An acknowledgement with no reason is indistinguishable from a
	// suppression someone added to make a red build green, which is the shape the suppression note
	// in the main output already exists to count.
	Reason string
}

// Key names an acknowledgement in output, the four things that pick out its finding.
func (acknowledged AcknowledgedDifference) Key() string {
	return fmt.Sprintf("%s:%s:%s:%q", acknowledged.File, acknowledged.Rule, acknowledged.Side, acknowledged.Anchor)
}

// KnownGateDefects are the differences where cohere is right and the gate is wrong.
//
// Two directions. `SideCohere` is a true positive the gate cannot see. `SideGate` is a false positive
// the gate reports and cohere deliberately does not, where a rule was made to tell with the type
// checker rather than given an allowance (cohere's parity doctrine: never worse than ESLint, rule by
// rule, and differing only in its favour). Each gate-side entry names the rule document (`.md` beside
// the rule) that records the condition, and the fixture holding the site as a must-stay-silent case.
//
// Each one names a defect in the tool being replaced, so each is a reason the migration is worth
// doing rather than a cost of it. They are listed here, in source, rather than passed in at the
// command line: an acknowledgement that can be supplied per-run can be supplied by whoever wants a
// green result, and this list is reviewed like any other code.
//
// Empty since 2026-10-04 (#cn8sthd). Every one of the 23 entries it held had stopped naming a
// difference: the ESLint side now carries cohere's rulings through Nexus's twins and wrappers, the
// gate's own rule bugs were fixed in Nexus, and the rest of the sites were rewritten so neither engine
// reports them. Both engines read 0 on all 19 files the table named. An entry is added back when a
// real difference is decided, with its anchor and its reason.
var KnownGateDefects = []AcknowledgedDifference{}

// acknowledgementMatch is what one acknowledgement matched among a run's differences.
type acknowledgementMatch struct {
	acknowledgement AcknowledgedDifference
	// differenceIndexes are the positions in the difference list whose finding it names.
	differenceIndexes []int
}

// SourceLineReader returns the text of a line in a file of the tree being compared, 1-based, and
// false when the file or the line cannot be read.
type SourceLineReader func(file string, line int) (string, bool)

// matchAcknowledgements pairs each acknowledgement with the differences its anchor names.
//
// Only a difference on the acknowledgement's side, in its file and for its rule, can match, and only
// when its source line contains the anchor. Without a line reader nothing matches, which reports
// every acknowledgement stale: a comparison that cannot read the code cannot tell an anchor that
// moved from one that holds, and a stale report is the honest answer to that.
func matchAcknowledgements(
	acknowledged []AcknowledgedDifference,
	differences []Difference,
	readSourceLine SourceLineReader,
) []acknowledgementMatch {
	matches := make([]acknowledgementMatch, 0, len(acknowledged))
	for _, acknowledgement := range acknowledged {
		match := acknowledgementMatch{acknowledgement: acknowledgement}
		for index, difference := range differences {
			if difference.OnlyOn != acknowledgement.Side ||
				difference.Finding.File != acknowledgement.File ||
				difference.Finding.Rule != acknowledgement.Rule ||
				readSourceLine == nil {
				continue
			}
			text, readable := readSourceLine(difference.Finding.File, difference.Finding.Line)
			if readable && strings.Contains(strings.TrimSpace(text), acknowledgement.Anchor) {
				match.differenceIndexes = append(match.differenceIndexes, index)
			}
		}
		matches = append(matches, match)
	}
	return matches
}
