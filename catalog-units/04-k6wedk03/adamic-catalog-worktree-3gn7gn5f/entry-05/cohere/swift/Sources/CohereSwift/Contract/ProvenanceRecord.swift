/*
 What built this engine and what it stands on. Always the first record of a run, and the whole output
 of `--version`, so every finding a reader carries away can be traced to a toolchain, a parser and a
 formatter, and the front door can refuse an engine that speaks a different contract.
 */
public struct ProvenanceRecord: Codable, Equatable, Sendable {
    public var kind = "provenance"
    public var contract: Int
    public var engine: String
    public var version: String
    public var commit: String
    public var sourceTreeModified: Bool
    public var toolchain: String
    public var swiftSyntax: String
    public var swiftFormat: String

    public init(
        contract: Int,
        engine: String,
        version: String,
        commit: String,
        sourceTreeModified: Bool,
        toolchain: String,
        swiftSyntax: String,
        swiftFormat: String,
    ) {
        self.contract = contract
        self.engine = engine
        self.version = version
        self.commit = commit
        self.sourceTreeModified = sourceTreeModified
        self.toolchain = toolchain
        self.swiftSyntax = swiftSyntax
        self.swiftFormat = swiftFormat
    }
}
