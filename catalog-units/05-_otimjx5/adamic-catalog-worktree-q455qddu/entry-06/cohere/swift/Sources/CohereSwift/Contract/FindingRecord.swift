/*
 One thing wrong with one file, from the compiler, a rule, or the formatter.

 Positions are 1-based lines and 1-based columns counted in UTF-8 bytes, which is what the TypeScript
 printer already prints and what swift-syntax, sourcekitd and the compiler's serialized diagnostics all
 use, so no position is converted anywhere between the parser and the reader's editor.
 */
public struct FindingRecord: Codable, Equatable, Sendable {
    /* Where a finding came from. The front door prints compiler findings in the compiler's shape and rule findings in the rule shape. */
    public enum Source: String, Codable, Sendable {
        case compiler
        case rule
        case format
    }

    /* How the finding was classified. Both severities are findings and both fail the run; the severity is kept so the line can say which the compiler called it. */
    public enum Severity: String, Codable, Sendable {
        case error
        case warning
    }

    /* One replacement, as 0-based UTF-8 byte offsets into the file as the engine read it. */
    public struct Edit: Codable, Equatable, Sendable {
        public var start: Int
        public var end: Int
        public var text: String

        public init(start: Int, end: Int, text: String) {
            self.start = start
            self.end = end
            self.text = text
        }
    }

    /* A repair offered and never applied, because choosing it is a judgment the reader makes. */
    public struct Suggestion: Codable, Equatable, Sendable {
        public var message: String
        public var fixes: [Edit]

        public init(message: String, fixes: [Edit]) {
            self.message = message
            self.fixes = fixes
        }
    }

    public var kind = "finding"
    public var source: Source
    public var file: String
    public var line: Int
    public var column: Int
    public var endLine: Int?
    public var endColumn: Int?
    public var severity: Severity
    public var rule: String
    public var messageId: String
    public var message: String
    public var fixes: [Edit]
    public var suggestions: [Suggestion]

    public init(
        source: Source,
        file: String,
        line: Int,
        column: Int,
        endLine: Int? = nil,
        endColumn: Int? = nil,
        severity: Severity,
        rule: String,
        messageId: String,
        message: String,
        fixes: [Edit] = [],
        suggestions: [Suggestion] = [],
    ) {
        self.source = source
        self.file = file
        self.line = line
        self.column = column
        self.endLine = endLine
        self.endColumn = endColumn
        self.severity = severity
        self.rule = rule
        self.messageId = messageId
        self.message = message
        self.fixes = fixes
        self.suggestions = suggestions
    }
}
