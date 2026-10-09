/*
 What the rules did, with the population beside the result. A rule that never listened to a file and a
 rule that listened to every file and found nothing produce the same empty finding list, and only the
 coverage fields here tell them apart.
 */
public struct LintRecord: Codable, Equatable, Sendable {
    /* A file a rule could not finish, so nothing that rule would have said about it was said. */
    public struct Crash: Codable, Equatable, Sendable {
        public var file: String
        public var error: String

        public init(file: String, error: String) {
            self.file = file
            self.error = error
        }
    }

    public var kind = "lint"
    public var findings: Int
    public var rulesRun: Int
    public var filesWalked: Int
    public var nodesVisited: Int
    public var elapsedMilliseconds: Int
    public var reusedFrom: String
    public var rulesSilent: [String]
    public var rulesWatchedAndQuiet: Int
    public var crashes: [Crash]
    public var rulesScopedOff: [String: Int]
    public var rulesNotConfigured: [String]
    public var configurationNote: String

    /* The contract spells this field `configNote`, which the Go front door reads; the Swift name is a full word and the wire name stays the contract's. */
    enum CodingKeys: String, CodingKey {
        case kind, findings, rulesRun, filesWalked, nodesVisited, elapsedMilliseconds, reusedFrom, rulesSilent,
            rulesWatchedAndQuiet, crashes, rulesScopedOff, rulesNotConfigured
        case configurationNote = "configNote"
    }

    public init(
        findings: Int,
        rulesRun: Int,
        filesWalked: Int,
        nodesVisited: Int,
        elapsedMilliseconds: Int,
        reusedFrom: String,
        rulesSilent: [String],
        rulesWatchedAndQuiet: Int,
        crashes: [Crash],
        rulesScopedOff: [String: Int],
        rulesNotConfigured: [String],
        configurationNote: String,
    ) {
        self.findings = findings
        self.rulesRun = rulesRun
        self.filesWalked = filesWalked
        self.nodesVisited = nodesVisited
        self.elapsedMilliseconds = elapsedMilliseconds
        self.reusedFrom = reusedFrom
        self.rulesSilent = rulesSilent
        self.rulesWatchedAndQuiet = rulesWatchedAndQuiet
        self.crashes = crashes
        self.rulesScopedOff = rulesScopedOff
        self.rulesNotConfigured = rulesNotConfigured
        self.configurationNote = configurationNote
    }
}
