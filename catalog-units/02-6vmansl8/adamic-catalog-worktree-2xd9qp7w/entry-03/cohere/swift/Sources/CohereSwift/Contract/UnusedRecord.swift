/*
 One thing the `--unused` report found: code that was written and is never used. The Swift counterpart of a
 line in the TypeScript unused report.

 It carries a finding's place and words but is not a `finding`: the report is not a gate, so the writer does
 not count it and it never changes the exit code. An import nothing uses is safe to remove and never breaks a
 build, so failing a build over one would make the report something people route around rather than read.
 */
public struct UnusedRecord: Codable, Equatable, Sendable {
    public var kind = "unused"
    public var file: String
    public var line: Int
    public var column: Int
    public var endLine: Int?
    public var endColumn: Int?
    public var rule: String
    public var messageId: String
    public var message: String
    /* The code reported, as written and short, for the report's one line per item: `import AhraOsCore`. */
    public var subject: String
    /* The removal, offered and never applied: `--unused` reports, it does not rewrite. */
    public var suggestions: [FindingRecord.Suggestion]

    public init(
        file: String,
        line: Int,
        column: Int,
        endLine: Int? = nil,
        endColumn: Int? = nil,
        rule: String,
        messageId: String,
        message: String,
        subject: String,
        suggestions: [FindingRecord.Suggestion] = [],
    ) {
        self.file = file
        self.line = line
        self.column = column
        self.endLine = endLine
        self.endColumn = endColumn
        self.rule = rule
        self.messageId = messageId
        self.message = message
        self.subject = subject
        self.suggestions = suggestions
    }

    /* The report's form of a finding a rule built at a node. */
    public init(_ finding: FindingRecord, subject: String) {
        self.init(
            file: finding.file,
            line: finding.line,
            column: finding.column,
            endLine: finding.endLine,
            endColumn: finding.endColumn,
            rule: finding.rule,
            messageId: finding.messageId,
            message: finding.message,
            subject: subject,
            suggestions: finding.suggestions,
        )
    }
}
