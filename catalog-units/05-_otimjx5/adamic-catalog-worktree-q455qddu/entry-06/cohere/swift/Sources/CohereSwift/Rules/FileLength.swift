import Foundation

/*
 A file holds at most 2,000 lines.

 Kirk's ruling (2026-10-03, #r3hfpe8). The size of a file is what a reader pays for, not how many types it
 holds: a file with one type and a fifteen-line helper beside it is easy to hold, and a 2,426-line file with
 one type is not. So size is checked on its own. A file past the limit has seams in it, and the repair is to
 split along them: an extension per concern (`Type+Purpose.swift`), or a type of its own (the orbit camera
 out of a 2,241-line viewport, for one).

 The threshold is Kirk's: 2,000. The line counts of every owned file in the two Swift repos on 2026-10-03
 (#r3hfpe8) put it in context. Of 1,469 files the median was about 100 lines; 141 were over 400 (SwiftLint's
 default), 20 over 1,000, 6 over 1,500 and 3 over 2,000. The limit reports only the files no reader can hold,
 rather than every substantial view or service, which a lower limit would turn into splits made to satisfy
 a number.

 Lines are counted the way SwiftLint's `file_length` counts them in its default configuration: every line,
 comments and blank lines included, with a final newline ending the last line rather than starting another.
 A comment is part of what a reader has to hold, and a count that skipped comments would reward moving
 explanation out of the file, the opposite of what the house wants. SwiftLint reports at the last line; this
 rule reports once at line 1, column 1, zero width, because the file as a whole is what is too long.
 */
public struct FileLength: FileRule {
    public let name = "cohere-swift/file-length"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "file_length"

    public static let maximumLines = 2_000

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        /* Kirk's convention (2026-10-03), shared with TypeScript: a generated file is named `*.generated.swift`, gets every other rule and the formatter, and is exempt from this one alone, because its length is the generator's, not a reader's. */
        guard !file.url.lastPathComponent.hasSuffix(".generated.swift") else { return [] }
        let lines = Self.lineCount(of: file.source)
        guard lines > Self.maximumLines else { return [] }
        return [
            FindingRecord(
                source: .rule,
                file: file.url.path,
                line: 1,
                column: 1,
                endLine: 1,
                endColumn: 1,
                severity: .error,
                rule: name,
                messageId: "tooManyLines",
                message:
                    "This file is \(lines) lines, over the \(Self.maximumLines) a file may hold. Split it along its seams, an extension per concern in Type+Purpose.swift or a helper type in its own file, so each part is one a reader can hold whole.",
            )
        ]
    }

    /* Line feeds, plus one for a last line that does not end in one. `\r\n` ends a line once. */
    static func lineCount(of source: String) -> Int {
        let utf8 = source.utf8
        let lineFeeds = utf8.reduce(into: 0) { count, byte in
            if byte == UInt8(ascii: "\n") {
                count += 1
            }
        }
        return lineFeeds + (utf8.last.map { $0 == UInt8(ascii: "\n") ? 0 : 1 } ?? 0)
    }
}
