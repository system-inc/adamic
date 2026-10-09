/*
 A rule that needs to know what a name resolves to, not only how it is spelled: `count` on an `Array` and
 `count` on a type of ours that is not a collection read the same and mean different things.

 `applies(to:)` is the rule's syntactic prefilter, and it carries the cost model. Symbols are fetched only for
 files some typed rule applies to: from the build's index store when it describes the file as it stands, from
 sourcekitd in process when it does not. So the prefilter should be as narrow as the syntax allows, and must
 never be narrower than what the rule could flag. A file the rule applies to that no source of symbols could
 describe is reported as unchecked, never as clean.
 */
public protocol TypedFileRule: Sendable {
    /* `cohere-swift/<rule>`, the key a config uses and the tag a finding prints. */
    var name: String { get }

    /* Whether the rule is our own or a port, and of which tool's check; see `RuleOrigin`. */
    var origin: RuleOrigin { get }

    /* The incumbent's own id for a port, spelled as that tool spells it, and nil for a house rule. */
    var upstreamName: String? { get }

    /* Whether the file holds any syntax this rule could flag. False means the rule has nothing to say here, whatever the types. */
    func applies(to file: ParsedFile) -> Bool

    func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord]
}
