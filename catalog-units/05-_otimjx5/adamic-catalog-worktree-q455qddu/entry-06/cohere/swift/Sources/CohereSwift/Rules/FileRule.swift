/*
 A rule that reads one file's tree and reports what is wrong with it.

 Rules see only the shared parse and report findings; they never read the disk, run a tool, or write.
 That keeps a rule cheap enough to add without justifying it, and it means the coverage the pipeline
 reports (which files a rule was offered, which it reported on) is the whole story of what it did.
 */
public protocol FileRule: Sendable {
    /* `cohere-swift/<rule>`, the key a config uses and the tag a finding prints. */
    var name: String { get }

    /* Whether the rule is our own or a port, and of which tool's check; see `RuleOrigin`. */
    var origin: RuleOrigin { get }

    /* The incumbent's own id for a port, spelled as that tool spells it, and nil for a house rule. */
    var upstreamName: String? { get }

    /* Whether the rule looks at this file at all. A rule that declines every file is reported as silent, never as clean. */
    func applies(to file: ParsedFile) -> Bool

    func findings(in file: ParsedFile) -> [FindingRecord]
}

extension FileRule {
    public func applies(to file: ParsedFile) -> Bool { true }
}
