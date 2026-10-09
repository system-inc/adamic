import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for the four force rules, both directions. Each failing case asserts the exact line and
 column, so a finding pointed at the wrong token fails rather than passing on its count. The passing cases
 are the near misses the rules must not confuse with the real thing: `!=`, prefix `!`, `try?`, `as?`, and an
 optional type spelled with `?`.
 */
struct ForceRuleTests {
    static func findings(_ rule: some FileRule, _ source: String) -> [FindingRecord] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        return rule.findings(in: file)
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    @Test func forceUnwrapsAreFound() {
        let source = """
            let first = optional!
            let second = call()!.member
            let third = chain?.value!
            """
        #expect(Self.positions(Self.findings(ForceUnwrapping(), source)) == ["1:21", "2:20", "3:25"])
    }

    @Test func forceUnwrapNearMissesAreNot() {
        let source = """
            let different = left != right
            let negated = !flag
            if let value = optional { use(value) }
            let defaulted = optional ?? 0
            """
        #expect(Self.findings(ForceUnwrapping(), source).isEmpty)
    }

    @Test func forceTriesAreFound() {
        let source = """
            let data = try! load()
            let decoded = try! decode(try! load())
            """
        #expect(Self.positions(Self.findings(ForceTry(), source)) == ["1:15", "2:18", "2:30"])
    }

    @Test func forceTryNearMissesAreNot() {
        let source = """
            let data = try load()
            let maybe = try? load()
            """
        #expect(Self.findings(ForceTry(), source).isEmpty)
    }

    /* Alone and inside a larger expression; the unfolded tree spells both as `UnresolvedAsExprSyntax`. */
    @Test func forceCastsAreFoundFoldedOrNot() {
        let source = """
            let view = item as! NSView
            let width = (item as! NSView).frame.width + margin
            let sum = base + (value as! Int) * 2
            """
        #expect(Self.positions(Self.findings(ForceCast(), source)) == ["1:19", "2:21", "3:27"])
    }

    @Test func forceCastNearMissesAreNot() {
        let source = """
            let view = item as? NSView
            let number = 3 as Double
            if item is NSView { }
            """
        #expect(Self.findings(ForceCast(), source).isEmpty)
    }

    /* The id and text come from the catalog; pinned here as literals so a generator bug cannot pass. */
    @Test func forceCastMessage() {
        let found = Self.findings(ForceCast(), "let view = item as! NSView")
        #expect(found.map(\.messageId) == ["forceCast"])
        #expect(
            found.map(\.message) == [
                "as! crashes the process when the value is not that type. Use as? and say what happens when it is not."
            ]
        )
    }

    @Test func implicitlyUnwrappedOptionalsAreFound() {
        let source = """
            var window: NSWindow!
            func handle(_ callback: (String!) -> Void) {}
            """
        #expect(Self.positions(Self.findings(ImplicitlyUnwrappedOptional(), source)) == ["1:13", "2:26"])
    }

    /* A finding points at the node, not at the whitespace or comments before it, which belong to the line above. */
    @Test func positionsSkipLeadingTrivia() {
        let source = """
            var window:
                NSWindow!
            """
        #expect(Self.positions(Self.findings(ImplicitlyUnwrappedOptional(), source)) == ["2:5"])
    }

    @Test func plainOptionalsAreNot() {
        let source = """
            var window: NSWindow?
            let forced = optional!
            """
        #expect(Self.findings(ImplicitlyUnwrappedOptional(), source).isEmpty)
    }

    @Test func everyForceRuleIsRegistered() {
        let names = Set(RuleRegistry.allNames)
        #expect(
            names.isSuperset(of: [
                ForceUnwrapping().name, ForceTry().name, ForceCast().name, ImplicitlyUnwrappedOptional().name,
            ])
        )
    }
}
