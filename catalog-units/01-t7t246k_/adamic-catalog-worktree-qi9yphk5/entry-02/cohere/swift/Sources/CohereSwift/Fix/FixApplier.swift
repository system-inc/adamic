/*
 Applies rule fixes to a file's text.

 Edits are UTF-8 byte ranges into the text the rules read, applied back to front so an earlier edit never
 moves a later one. Two edits that overlap cannot both be right about the same bytes, so the first one wins
 and the other is refused and counted. It is not dropped silently: the fix line reports refusals by reason,
 so a fixer that keeps losing to another is visible rather than inert.
 */
enum FixApplier {
    /* An edit a rule proposed, kept with the rule's name, so what landed can be told by the rule that made it. */
    struct Proposal: Equatable {
        var rule: String
        var edit: FindingRecord.Edit
    }

    /* The rewritten text and what happened to every edit offered. */
    struct Result: Equatable {
        var text: String
        var applied: Int
        /* The applied edits by the rule that proposed them; its values add up to `applied`. */
        var appliedByRule: [String: Int]
        var refusedOverlapping: Int
        var refusedInvalidRange: Int
    }

    /* Every fix the findings carry, each with its finding's rule. */
    static func proposals(from findings: [FindingRecord]) -> [Proposal] {
        findings.flatMap { finding in finding.fixes.map { Proposal(rule: finding.rule, edit: $0) } }
    }

    static func apply(_ proposals: [Proposal], to text: String) -> Result {
        var bytes = Array(text.utf8)
        var applied = 0
        var appliedByRule: [String: Int] = [:]
        var refusedOverlapping = 0
        var refusedInvalidRange = 0

        /* Accepted in source order, so "first one wins" means the earliest in the file, the same every run. */
        var accepted: [Proposal] = []
        for proposal in proposals.sorted(by: { ($0.edit.start, $0.edit.end) < ($1.edit.start, $1.edit.end) }) {
            let edit = proposal.edit
            guard edit.start >= 0, edit.start <= edit.end, edit.end <= bytes.count else {
                refusedInvalidRange += 1
                continue
            }
            if let previous = accepted.last, edit.start < previous.edit.end {
                refusedOverlapping += 1
                continue
            }
            accepted.append(proposal)
        }
        for proposal in accepted.reversed() {
            bytes.replaceSubrange(proposal.edit.start..<proposal.edit.end, with: Array(proposal.edit.text.utf8))
            applied += 1
            appliedByRule[proposal.rule, default: 0] += 1
        }
        return Result(
            text: String(decoding: bytes, as: UTF8.self),
            applied: applied,
            appliedByRule: appliedByRule,
            refusedOverlapping: refusedOverlapping,
            refusedInvalidRange: refusedInvalidRange,
        )
    }
}
