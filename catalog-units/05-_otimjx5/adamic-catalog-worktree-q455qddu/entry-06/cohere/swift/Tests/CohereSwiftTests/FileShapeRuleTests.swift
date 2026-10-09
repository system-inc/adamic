import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/* Fixtures for consistency-require-matching-file-name: a file named for its main type, and a purpose file with its helpers (Kirk's rulings of 2026-10-03). */
struct FileShapeRuleTests {
    static func file(_ name: String, _ source: String) -> ParsedFile {
        ParsedFile(
            url: URL(fileURLWithPath: "/fixture/\(name)"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
    }

    static func messages(_ rule: some FileRule, _ name: String, _ source: String) -> [String] {
        let subject = file(name, source)
        guard rule.applies(to: subject) else { return [] }
        return rule.findings(in: subject).map(\.messageId)
    }

    @Test func oneTypeWithNestedHelpersPasses() {
        let source =
            "struct Pane {\n    struct Layout {}\n    enum Mode { case a }\n}\nextension Pane { func draw() {} }\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "Pane.swift", source).isEmpty)
    }

    /* A generated file is named for its type before the marker, and still answers to the rule. */
    @Test func aGeneratedFileIsNamedBeforeItsMarker() {
        let source = "enum PolicyAbbreviations {}\n"
        let rule = ConsistencyRequireMatchingFileName()
        #expect(Self.messages(rule, "PolicyAbbreviations.generated.swift", source).isEmpty)
        #expect(Self.messages(rule, "Policy.generated.swift", source) == ["requireMatchingFileName"])
        #expect(Self.messages(rule, "PolicyAbbreviations.swift", source).isEmpty)
    }

    /* How many types a file holds is not judged; the name must lead to one of them, the main one. */
    @Test func aFileOfSeveralTypesIsNamedForTheMainOne() {
        let source = "struct Pane {}\nfinal class Box {}\nactor Worker {}\nprotocol Drawable {}\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "Pane.swift", source).isEmpty)
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Utilities.swift", source) == [
                "requireMatchingFileName"
            ]
        )
        /* The main type need not come first: the one the file is named for is its own. */
        #expect(
            Self.messages(
                ConsistencyRequireMatchingFileName(),
                "Pane.swift",
                "final class Helper {}\nstruct Pane {}\n",
            ).isEmpty
        )
    }

    /* A struct of `lines` lines, braces included, so a fixture can sit exactly at the limit or one past it. */
    static func structSource(_ name: String, lines: Int) -> String {
        "struct \(name) {\n" + (0..<(lines - 2)).map { "    var field\($0) = 0\n" }.joined() + "}\n"
    }

    /* What consistency-require-matching-file-name says of a purpose file holding `source` beside the extension it is named for. */
    static func besidePurpose(_ source: String) -> [String] {
        Self.messages(
            ConsistencyRequireMatchingFileName(),
            "Pane+Layout.swift",
            "extension Pane {\n    func layout() {}\n}\n" + source,
        )
    }

    /* In a purpose file a private helper stays at any size, and a protocol or a class beside it is a type the file is not named for. */
    @Test func aPurposeFilesPrivateHelperStaysAtAnySize() {
        #expect(Self.besidePurpose("private " + Self.structSource("Row", lines: 200)).isEmpty)
        #expect(Self.besidePurpose("private final class Helper {}\nfileprivate actor Worker {}\n").isEmpty)
        #expect(Self.besidePurpose("final class Box {}\n") == ["requireMatchingFileName"])
        #expect(Self.besidePurpose("protocol Drawable {}\n") == ["requireMatchingFileName"])
    }

    /* Small means at most 30 lines: 30 stays, 31 does not. */
    @Test func aSmallValueTypeStaysAtTheLimitAndNotPastIt() {
        #expect(Self.besidePurpose(Self.structSource("Row", lines: 30)).isEmpty)
        #expect(Self.besidePurpose(Self.structSource("Row", lines: 31)) == ["requireMatchingFileName"])
        #expect(Self.besidePurpose("enum Mode { case idle, busy }\n").isEmpty)
    }

    /* The comment directly above and the attributes count toward the 30; a comment a blank line away does not. */
    @Test func theAttachedCommentAndAttributesAreCounted() {
        let body = Self.structSource("Row", lines: 28)
        #expect(Self.besidePurpose("\n/* Why. */\n@frozen\n" + body).isEmpty)
        #expect(Self.besidePurpose("\n/*\n Why.\n */\n@frozen\n" + body) == ["requireMatchingFileName"])
        #expect(Self.besidePurpose("\n/// Why.\n/// More.\n@frozen\n" + body) == ["requireMatchingFileName"])
        #expect(Self.besidePurpose("\n/*\n Header.\n */\n\n@frozen\n" + body).isEmpty)
    }

    /* An extension of the small type in the same file is part of its size. */
    @Test func anExtensionOfTheSmallTypeCountsTowardIt() {
        let small = Self.structSource("Row", lines: 20)
        #expect(Self.besidePurpose(small + "extension Row {\n    var total: Int { 0 }\n}\n").isEmpty)
        /* Twenty lines of struct and eleven of extension: thirty-one in all. */
        let extensionSource = "extension Row {\n" + (0..<9).map { "    var extra\($0): Int { 0 }\n" }.joined() + "}\n"
        #expect(!Self.besidePurpose(small + extensionSource).isEmpty)
    }

    @Test func aTypeInBothBranchesOfAnIfIsOneType() {
        let source = "#if os(macOS)\nstruct Pane {}\n#else\nstruct Pane {}\n#endif\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "Pane.swift", source).isEmpty)
    }

    @Test func aTypeAliasIsNotAType() {
        let source = "typealias Panes = [Pane]\nstruct Pane {}\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "Pane.swift", source).isEmpty)
    }

    @Test func aFileNotNamedForItsTypeIsReported() {
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Utilities.swift", "final class Pane {}\n") == [
                "requireMatchingFileName"
            ]
        )
    }

    /* The words of each finding: the misnamed file under TypeScript's id, and an extension placed beside a type or alone. */
    @Test func eachFindingSaysWhereItBelongs() {
        let rule = ConsistencyRequireMatchingFileName()
        let misnamed = rule.findings(in: Self.file("Utilities.swift", "final class Pane {}\n"))
        #expect(
            misnamed.map(\.message) == [
                "`Pane` is the first type declared in `Utilities.swift`, and the file is not named for it. Name it `Pane.swift`, so a reader looking for `Pane` opens it without a search."
            ]
        )
        let beside = rule.findings(in: Self.file("Pane.swift", "struct Pane {}\nextension String {}\n"))
        #expect(
            beside.map(\.message) == [
                "An extension of String in the file for Pane. Extensions of another type belong in that type's own file, such as String+Purpose.swift, so everything added to String is found together."
            ]
        )
        let alone = rule.findings(in: Self.file("Helpers.swift", "extension String {}\n"))
        #expect(
            alone.map(\.message) == [
                "This file is Helpers.swift and extends String. Name it String+Purpose.swift after what it adds, so everything added to String is found together."
            ]
        )
    }

    @Test func anExtensionOfAnotherTypeBelongsInItsOwnFile() {
        let source = "struct Pane {}\nextension String { var trimmed: String { self } }\n"
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Pane.swift", source) == ["extensionOutsideItsFile"]
        )
    }

    @Test func aPrivateExtensionStaysBesideItsType() {
        let source = "struct App {}\nprivate extension View { var windowEnvironment: Int { 0 } }\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "App.swift", source).isEmpty)
    }

    @Test func aPrivateHelperExtensionStaysInAPurposeFile() {
        let source =
            "extension Terminal.SeededVtState { func wireCheckpoint() {} }\nfileprivate extension VtColor { init(value: Int) {} }\n"
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Terminal+AuthoritativeCheckpoint.swift", source)
                .isEmpty
        )
    }

    /* The control: with no visible face, a private extension is just a misnamed file. */
    @Test func aFileOfOnlyPrivateExtensionsIsStillNamedForWhatItExtends() {
        let source = "private extension VtColor { init(value: Int) {} }\n"
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Helpers.swift", source) == ["extensionOutsideItsFile"]
        )
    }

    /* The SwiftUI modifier idiom: the extension is the modifier's public face. */
    @Test func anExtensionExposingTheFilesTypeStaysBesideIt() {
        let source =
            "struct DelayWidthUntilIdle: ViewModifier {}\nextension View {\n    func delayWidthUntilIdle() -> some View { modifier(DelayWidthUntilIdle()) }\n}\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "DelayWidthUntilIdle.swift", source).isEmpty)
    }

    /* One member that does not name the type is enough to send the extension to its own file. */
    @Test func anExtensionWithUnrelatedMembersStillMoves() {
        let source =
            "struct Pane {}\nextension View {\n    func pane() -> Pane { Pane() }\n    func unrelated() {}\n}\n"
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Pane.swift", source) == ["extensionOutsideItsFile"]
        )
    }

    @Test func purposeFilesPass() {
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "String+Trimming.swift", "extension String {}\n")
                .isEmpty
        )
        #expect(
            Self.messages(
                ConsistencyRequireMatchingFileName(),
                "Array+Chunks.swift",
                "extension Array where Element == Int {}\n",
            ).isEmpty
        )
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Outer+Inner.swift", "extension Outer.Inner {}\n")
                .isEmpty
        )
        #expect(
            Self.messages(
                ConsistencyRequireMatchingFileName(),
                "Outer.Inner+Codable.swift",
                "extension Outer.Inner: Codable {}\n",
            ).isEmpty
        )
    }

    /* A purpose file's helpers stay beside the extension it is named for; a bigger internal type, or a file whose only face is the helper, still answers. */
    @Test func aPurposeFileMayHoldHelperTypes() {
        let privateHelper =
            "extension Prune {\n    func run() {}\n}\nprivate struct BinarySlice {\n    var low: Int\n}\nextension BinarySlice { var isEmpty: Bool { low == 0 } }\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "Prune+Search.swift", privateHelper).isEmpty)
        let smallValues =
            "extension StagePane {\n    func probe() {}\n}\nstruct ProbeResult {\n    var score: Double\n}\nenum ProbeStage { case warm, run }\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "StagePane+Probe.swift", smallValues).isEmpty)
        let internalClass = "extension StagePane {\n    func probe() {}\n}\nfinal class ProbeRunner {}\n"
        #expect(Self.messages(ConsistencyRequireMatchingFileName(), "StagePane+Probe.swift", internalClass).count == 1)
        let helperAsTheOnlyFace =
            "private extension StagePane {\n    func probe() {}\n}\nprivate struct ProbeResult {}\n"
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "StagePane+Probe.swift", helperAsTheOnlyFace).count == 1
        )
    }

    @Test func anExtensionOnlyFileNamedForSomethingElseIsReported() {
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "Helpers.swift", "extension String {}\n") == [
                "extensionOutsideItsFile"
            ]
        )
    }

    @Test func mainKeepsItsName() {
        #expect(
            Self.messages(ConsistencyRequireMatchingFileName(), "main.swift", "struct Tool {}\nTool().run()\n").isEmpty
        )
    }

    @Test func freeFunctionsHaveNoNameToMatch() {
        #expect(
            Self.messages(
                ConsistencyRequireMatchingFileName(),
                "Math.swift",
                "func clamp(_ value: Int) -> Int { value }\n",
            ).isEmpty
        )
    }
}
