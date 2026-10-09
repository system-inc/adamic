import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `cohere-swift/unowned-variable-capture`, both directions. The capture cases are SwiftLint's
 `unowned_variable_capture` examples, triggering and not, with the exact column it marks. The property
 cases are cohere-only, including SwiftLint's own non-triggering `unowned var value: First`, which this rule
 flags on purpose. The near misses are `weak` in both positions, captures with no specifier, and the word
 `unowned` used as an ordinary identifier, label, or member, which the parser never reads as the keyword.
 */
struct UnownedVariableCaptureTests {
    static func findings(_ source: String) -> [FindingRecord] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        return UnownedVariableCapture().findings(in: file)
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    /* SwiftLint's five triggering examples, at the column it marks. */
    @Test func unownedCapturesAreFound() {
        let source = """
            foo { [unowned self] in _ }
            foo { [unowned bar] in _ }
            foo { [unowned(safe) self] in _ }
            foo { [bar, unowned self] in _ }
            foo { [unowned(unsafe) self] in _ }
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["1:8", "2:8", "3:8", "4:13", "5:8"])
        #expect(found.allSatisfy { $0.messageId == "unownedCapture" })
    }

    /* Captures with an initializer, mixed with weak ones, across lines, and nested closures each report their own. */
    @Test func unownedCapturesInLargerShapesAreFound() {
        let source = """
            let handler = { [weak self, unowned model = self.model] (value: Int) in
                model.apply(value)
                queue.async { [unowned self] in self.finish() }
            }
            items.map { [
                unowned owner
            ] item in owner.wrap(item) }
            """
        #expect(Self.positions(Self.findings(source)) == ["1:29", "3:20", "6:5"])
    }

    /* Cohere-only: SwiftLint lists the first as non-triggering. Stored properties, `unowned(unsafe)`, and a local. */
    @Test func unownedPropertiesAreFound() {
        let source = """
            final class First {}
            final class Second {
                unowned var value: First
                private unowned let owner: First
                unowned(unsafe) var fast: First
                unowned(safe) let checked: First
                init(value: First) {
                    self.value = value
                    unowned let local = value
                }
            }
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["3:5", "4:13", "5:5", "6:5", "9:9"])
        #expect(found.allSatisfy { $0.messageId == "unownedProperty" })
    }

    /* SwiftLint's non-triggering capture examples, minus the property case flagged above. */
    @Test func weakAndPlainCapturesAreNot() {
        let source = """
            foo { [weak self] in _ }
            foo { [weak self] param in _ }
            foo { [weak bar] in _ }
            foo { [weak bar] param in _ }
            foo { bar in _ }
            foo { $0 }
            foo { [bar, baz = qux] in _ }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* `weak` in the property position, and `unowned` as a name, label, member, string, and comment. */
    @Test func unownedAsAWordIsNot() {
        let source = """
            final class Holder {
                weak var delegate: Delegate?
                let unowned = 1
                var description: String { "unowned self" }
                func mark(unowned: Bool) {}
            }
            /* [unowned self] in a comment */
            holder.mark(unowned: holder.unowned == 1)
            let unownedCount = references.filter { $0.isUnowned }.count
            """
        #expect(Self.findings(source).isEmpty)
    }
}
