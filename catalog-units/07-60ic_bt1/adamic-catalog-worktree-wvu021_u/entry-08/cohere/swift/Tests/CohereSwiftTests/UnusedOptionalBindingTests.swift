import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `unused-optional-binding`, both directions. The failing cases are SwiftLint's own
 triggering examples plus `var`, `while`, nested and labeled tuples, a type annotation and the empty tuple,
 each with its exact position on the pattern. The passing cases are SwiftLint's non-triggering examples
 and the pattern matches that look like a binding but are not one. The message cases pin which repair each
 right side gets, and that a `try?` in a condition is reported here and not by `correctness-no-discarded-try-optional`.
 */
struct UnusedOptionalBindingTests {
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
        return UnusedOptionalBinding().findings(in: file)
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    @Test func incumbentTriggeringExamplesAreFound() {
        let source = """
            if let _ = Foo.optionalValue {}
            if let a = Foo.optionalValue, let _ = Foo.optionalValue2 {}
            guard let a = Foo.optionalValue, let _ = Foo.optionalValue2 {}
            if let (first, second) = getOptionalTuple(), let _ = Foo.optionalValue {}
            if let (first, _) = getOptionalTuple(), let _ = Foo.optionalValue {}
            if let (_, second) = getOptionalTuple(), let _ = Foo.optionalValue {}
            if let (_, _, _) = getOptionalTuple(), let bar = Foo.optionalValue {}
            func foo() { if let _ = bar {} }
            """
        #expect(
            Self.positions(Self.findings(source)) == ["1:8", "2:35", "3:38", "4:50", "5:45", "6:46", "7:8", "8:21"]
        )
    }

    @Test func otherBindingShapesAreFound() {
        let source = """
            if var _ = value {}
            while let _ = iterator.next() {}
            guard let _ = value else { return }
            if let ((_, _), _) = nested {}
            if let (_) = value {}
            if let (_: _, _) = labeled {}
            if let _: Int = value {}
            if let () = empty {}
            if let _ = value, let _ = other {}
            if let _ = try work() {}
            if let _ = await work() {}
            """
        #expect(
            Self.positions(Self.findings(source)) == [
                "1:8", "2:11", "3:11", "4:8", "5:8", "6:8", "7:8", "8:8", "9:8", "9:23", "10:8", "11:8",
            ]
        )
    }

    @Test func incumbentNonTriggeringExamplesAreNot() {
        let source = """
            if let bar = Foo.optionalValue {}
            if let (_, second) = getOptionalTuple() {}
            if let (_, asd, _) = getOptionalTuple(), let bar = Foo.optionalValue {}
            if foo() { let _ = bar() }
            if foo() { _ = bar() }
            if case .some(_) = self {}
            if let point = state.find({ _ in true }) {}
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Pattern matches, shorthand bindings, a nested tuple that binds a name, and a backticked name. */
    @Test func nearMissesAreNot() {
        let source = """
            if case let _ = value {}
            if case let _? = value {}
            if case _? = value {}
            for case let _? in list {}
            switch value { case let _?: break; default: break }
            if let value {}
            if let ((_, used), _) = nested {}
            if let `_` = value {}
            let _ = value
            guard value != nil else { return }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func eachRightSideGetsItsRepair() {
        let source = """
            if let _ = value {}
            if let _ = try? work() {}
            if let _ = item as? NSView {}
            if let _ = try work() {}
            if let _ = item as NSView? {}
            """
        let messageIds = Self.findings(source).map(\.messageId)
        #expect(
            messageIds == [
                "unusedOptionalBinding", "unusedOptionalBindingTry", "unusedOptionalBindingCast",
                "unusedOptionalBinding", "unusedOptionalBinding",
            ]
        )
    }

    /* `correctness-no-discarded-try-optional` leaves a `try?` in a condition alone, so this rule is its only report. */
    @Test func tryOptionalInAConditionIsReportedOnce() {
        let source = """
            if let _ = try? work() {}
            """
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(CorrectnessNoDiscardedTryOptional().findings(in: file, symbols: FileSymbols([])).isEmpty)
        #expect(Self.positions(Self.findings(source)) == ["1:8"])
    }
}
