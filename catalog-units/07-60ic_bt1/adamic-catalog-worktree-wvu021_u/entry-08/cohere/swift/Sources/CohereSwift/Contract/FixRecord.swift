/*
 What the fix and format phase did. The front door fills an `edit.Summary` from it and prints that
 summary's own line, so the Swift and TypeScript fix lines cannot drift apart.

 Formatting keeps its own counts, separate from fixes, because a formatter that declined every file and
 a formatter that found every file already correct otherwise print the same zero.
 */
public struct FixRecord: Codable, Equatable, Sendable {
    /* One file the phase wrote: the fixes that landed in it by rule, and whether the formatter changed it. */
    public struct ChangedFile: Codable, Equatable, Sendable {
        public var file: String
        public var fixedBy: [String: Int]
        public var formatted: Bool

        public init(file: String, fixedBy: [String: Int], formatted: Bool) {
            self.file = file
            self.fixedBy = fixedBy
            self.formatted = formatted
        }
    }

    public var kind = "fix"
    public var filesConsidered: Int
    public var filesRewritten: Int
    public var fixesApplied: Int
    public var fixesRefused: Int
    public var refusalsByReason: [String: Int]
    public var filesReformatted: Int
    public var filesNotFormatted: Int
    public var notFormattedReasons: [String: Int]
    public var formatScope: String
    /*
     Every file written, once each, so its length is `filesRewritten`. Additive within contract 3 (Contract.md): this
     engine always writes it, and a stream from an engine before it has none, which decodes as nil.
     */
    public var changedFiles: [ChangedFile]?

    public init(
        filesConsidered: Int,
        filesRewritten: Int,
        fixesApplied: Int,
        fixesRefused: Int,
        refusalsByReason: [String: Int],
        filesReformatted: Int,
        filesNotFormatted: Int,
        notFormattedReasons: [String: Int],
        formatScope: String,
        changedFiles: [ChangedFile]?,
    ) {
        self.filesConsidered = filesConsidered
        self.filesRewritten = filesRewritten
        self.fixesApplied = fixesApplied
        self.fixesRefused = fixesRefused
        self.refusalsByReason = refusalsByReason
        self.filesReformatted = filesReformatted
        self.filesNotFormatted = filesNotFormatted
        self.notFormattedReasons = notFormattedReasons
        self.formatScope = formatScope
        self.changedFiles = changedFiles
    }
}
