import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for file-length: the limit itself and one past it, the way a final newline is counted,
 and the one finding's place and wording.
 */
struct FileLengthTests {
    static func findings(_ source: String, fileName: String = "Pane.swift") -> [FindingRecord] {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/\(fileName)"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let rule = FileLength()
        guard rule.applies(to: file) else { return [] }
        return rule.findings(in: file)
    }

    /* `count` lines of code, each ending in a newline, as an editor saves them. */
    static func source(lines count: Int) -> String {
        (0..<count).map { "let value\($0) = \($0)\n" }.joined()
    }

    @Test func aFileAtTheLimitPasses() {
        #expect(Self.findings(Self.source(lines: FileLength.maximumLines)).isEmpty)
    }

    @Test func aFileOnePastTheLimitIsReportedOnceAtTheTop() throws {
        let findings = Self.findings(Self.source(lines: FileLength.maximumLines + 1))
        #expect(findings.count == 1)
        let finding = try #require(findings.first)
        #expect("\(finding.line):\(finding.column)-\(finding.endLine ?? 0):\(finding.endColumn ?? 0)" == "1:1-1:1")
        #expect(finding.messageId == "tooManyLines")
        #expect(finding.message.hasPrefix("This file is 2001 lines, over the 2000 a file may hold."))
    }

    /* The limit is 2,000, the number Kirk chose; a change to it should be a decision, not a drift. */
    @Test func theLimitIsTwoThousand() {
        #expect(FileLength.maximumLines == 2_000)
    }

    /* A generated file, named `*.generated.swift`, is exempt from this rule alone; a name that only contains the word is not. */
    @Test func aGeneratedFileIsExemptByItsName() {
        let long = Self.source(lines: FileLength.maximumLines + 1)
        #expect(Self.findings(long, fileName: "Schema.generated.swift").isEmpty)
        #expect(Self.findings(long, fileName: "GeneratedSchema.swift").count == 1)
        #expect(Self.findings(long, fileName: "Schema.generated.swift.swift").count == 1)
    }

    /* A final newline ends the last line; it does not start another. Without one, the last line still counts. */
    @Test func linesAreCountedAsAnEditorShowsThem() {
        #expect(FileLength.lineCount(of: "") == 0)
        #expect(FileLength.lineCount(of: "let a = 1") == 1)
        #expect(FileLength.lineCount(of: "let a = 1\n") == 1)
        #expect(FileLength.lineCount(of: "let a = 1\n\nlet b = 2") == 3)
        #expect(FileLength.lineCount(of: "let a = 1\r\nlet b = 2\r\n") == 2)
    }

    /* Comments and blank lines are part of what a reader holds, so they count. */
    @Test func commentsAndBlankLinesCount() {
        let source =
            "/*\n" + String(repeating: " A line of explanation.\n\n", count: FileLength.maximumLines / 2) + "*/\n"
        #expect(Self.findings(source).count == 1)
    }

    @Test func theRuleIsRegistered() {
        #expect(RuleRegistry.allNames.contains(FileLength().name))
    }
}
