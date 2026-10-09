import Foundation
import SwiftSyntax

/*
 No `TODO` or `FIXME` comments: unfinished work belongs in the task tree, where it has an owner, a
 priority and a place in a plan, not in a comment that is read only when someone happens to open the file.

 The Swift form of our TypeScript `no-warning-comments`. Matched as whole words at the start of a
 comment's text, so prose that mentions "a todo list" is not a finding.
 */
public struct Todo: FileRule {
    public let name = "cohere-swift/todo"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "todo"

    static let markers = ["TODO", "FIXME"]

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        var findings: [FindingRecord] = []
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            var offset = token.position.utf8Offset
            for piece in token.leadingTrivia {
                defer { offset += piece.sourceLength.utf8Length }
                guard let text = Self.commentText(piece), let marker = Self.marker(in: text) else { continue }
                let location = file.locations.location(for: AbsolutePosition(utf8Offset: offset))
                findings.append(
                    FindingRecord(
                        source: .rule,
                        file: file.url.path,
                        line: location.line,
                        column: location.column,
                        severity: .error,
                        rule: name,
                        messageId: "todoComment",
                        message:
                            "A \(marker) comment is unfinished work that only someone opening this file will see. Put it in the task tree, where it has an owner and a priority, and delete the comment.",
                    )
                )
            }
        }
        return findings
    }

    static func commentText(_ piece: TriviaPiece) -> String? {
        switch piece {
            case let .lineComment(text), let .docLineComment(text), let .blockComment(text), let .docBlockComment(text):
                return text
            default:
                return nil
        }
    }

    /* `TODO` or `FIXME` as the first word of the comment, or of any line of a block comment. */
    static func marker(in comment: String) -> String? {
        for line in comment.split(separator: "\n", omittingEmptySubsequences: false) {
            let body = line.drop { $0 == "/" || $0 == "*" || $0 == " " || $0 == "\t" }
            for marker in markers where body.hasPrefix(marker) {
                let after = body.dropFirst(marker.count).first
                if after == nil || !(after?.isLetter ?? false) {
                    return marker
                }
            }
        }
        return nil
    }
}
