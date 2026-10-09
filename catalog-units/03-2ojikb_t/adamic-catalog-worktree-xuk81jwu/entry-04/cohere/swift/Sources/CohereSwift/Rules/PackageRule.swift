/*
 A rule about the package as a whole rather than any one file, such as which language mode its targets
 compile in. It gets the package model and the parsed manifest, so a finding can point at the line of
 `Package.swift` that would fix it.
 */
public protocol PackageRule: Sendable {
    var name: String { get }

    /* Whether the rule is our own or a port, and of which tool's check; see `RuleOrigin`. */
    var origin: RuleOrigin { get }

    /* The incumbent's own id for a port, spelled as that tool spells it, and nil for a house rule. */
    var upstreamName: String? { get }

    func findings(in package: PackageModel, manifest: ParsedFile?) -> [FindingRecord]
}
