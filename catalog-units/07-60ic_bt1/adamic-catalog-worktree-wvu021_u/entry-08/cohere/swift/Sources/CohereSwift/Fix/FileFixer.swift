import SwiftParser
import SwiftSyntax

/*
 Runs every enabled rule that proposes fixes over one file, applies the fixes, parses the result, and goes
 again until nothing more is proposed or the pass limit is reached.

 Passes repeat because one fix can expose another: the rewritten text can match a rule the old text did
 not. The limit exists because two fixers that undo each other would otherwise loop forever.
 A file that still has fixes at the limit is not hidden: its remaining findings surface in lint like any
 other finding.

 Every pass runs every enabled rule, so the last pass already holds what lint would find in the text
 the fixer ends on. That is returned (`findingsOfFinalText`), and lint reuses it for a file the formatter
 then leaves alone, rather than walking the same text twice. Only the last pass's findings qualify, and
 only when no fix landed after them: a pass that applied fixes moves the text past what it found.

 Rules see only the tree, so a fixed text is parsed fresh before the next pass. A pass that would produce
 text the parser rejects is discarded, and the file keeps the last text that parsed. A fixer that breaks a
 file must never be the reason the file stops compiling.
 */
struct FileFixer {
    /* What fixing one file did. */
    struct Result {
        var file: ParsedFile
        var applied: Int
        /* The applied fixes by the rule that made them, across every pass; its values add up to `applied`. */
        var appliedByRule: [String: Int]
        var refusalsByReason: [String: Int]
        /* Each enabled rule that applies, with what it found in `file` as returned; nil when the text moved after the last walk. */
        var findingsOfFinalText: [String: [FindingRecord]]?
    }

    let configuration: RuleConfiguration
    let rules: [any FileRule]
    let maximumPasses: Int

    func fix(_ file: ParsedFile) -> Result {
        var current = file
        var applied = 0
        var appliedByRule: [String: Int] = [:]
        var refusals: [String: Int] = [:]
        var findingsOfFinalText: [String: [FindingRecord]]?
        for _ in 0..<maximumPasses {
            var findingsByRule: [String: [FindingRecord]] = [:]
            var proposals: [FixApplier.Proposal] = []
            for rule in rules where configuration.severity(of: rule.name) != .off && rule.applies(to: current) {
                let found = rule.findings(in: current)
                findingsByRule[rule.name] = found
                proposals.append(contentsOf: FixApplier.proposals(from: found))
            }
            findingsOfFinalText = findingsByRule
            guard !proposals.isEmpty else { break }
            let result = FixApplier.apply(proposals, to: current.source)
            refusals["overlaps another fix", default: 0] += result.refusedOverlapping
            refusals["invalid range", default: 0] += result.refusedInvalidRange
            guard result.applied > 0, result.text != current.source else { break }
            let tree = Parser.parse(source: result.text)
            if tree.hasError {
                refusals["the fixed text would not parse", default: 0] += result.applied
                break
            }
            let counter = NodeCounter(viewMode: .sourceAccurate)
            counter.walk(tree)
            current = ParsedFile(
                url: current.url,
                targetName: current.targetName,
                targetKind: current.targetKind,
                source: result.text,
                tree: tree,
                nodeCount: counter.count,
                packageRoot: current.packageRoot,
            )
            applied += result.applied
            appliedByRule.merge(result.appliedByRule, uniquingKeysWith: +)
            /* The text moved past what this pass found; only a later pass over it can stand for lint. */
            findingsOfFinalText = nil
        }
        return Result(
            file: current,
            applied: applied,
            appliedByRule: appliedByRule,
            refusalsByReason: refusals.filter { $0.value > 0 },
            findingsOfFinalText: findingsOfFinalText,
        )
    }
}
