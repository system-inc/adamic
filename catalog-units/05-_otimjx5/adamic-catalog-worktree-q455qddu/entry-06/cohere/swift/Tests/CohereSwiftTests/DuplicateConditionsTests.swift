import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `duplicate-conditions`, both directions. The failing cases are every triggering example
 SwiftLint documents for `duplicate_conditions`, each asserting the exact line and column SwiftLint marks:
 the first token of an `if` branch's condition list, or of a `case` item. The passing cases are every
 non-triggering example it documents, then the near misses this rule adds: calls, macros and `await`,
 which may answer differently each time, and branches that call between two copies.
 */
struct DuplicateConditionsTests {
    static func positions(_ source: String) -> [String] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        return DuplicateConditions().findings(in: file).map { "\($0.line):\($0.column)" }
    }

    @Test func repeatedIfConditionIsFound() {
        let source = """
            if x < 5 {
                foo()
            } else if y == "s" {
                bar()
            } else if x < 5 {
                baz()
            }
            """
        #expect(Self.positions(source) == ["1:4", "5:11"])
    }

    @Test func repeatedConditionInNestedChainIsFound() {
        let source = """
            if z {
                if x < 5 {
                    foo()
                } else if y == "s" {
                    bar()
                } else if x < 5 {
                    baz()
                }
            }
            """
        #expect(Self.positions(source) == ["2:8", "6:15"])
    }

    /* Clauses compare as a set, so their order does not matter. */
    @Test func reorderedClausesAreFound() {
        let source = """
            if x < 5, y == "s" {
                foo()
            } else if x < 10 {
                bar()
            } else if y == "s", x < 5 {
                baz()
            }
            """
        #expect(Self.positions(source) == ["1:4", "5:11"])
    }

    @Test func caseSharedAcrossSwitchCasesIsFound() {
        let source = """
            switch x {
            case "a", "b":
                foo()
            case "c", "a":
                bar()
            }
            """
        #expect(Self.positions(source) == ["2:6", "4:11"])
    }

    @Test func repeatedCaseWithSameWhereIsFound() {
        let source = """
            switch x {
            case "a" where y == "s":
                foo()
            case "a" where y == "s":
                bar()
            }
            """
        #expect(Self.positions(source) == ["2:6", "4:6"])
    }

    @Test func repeatedOptionalBindingIsFound() {
        let source = """
            if let xyz = maybeXyz {
                foo()
            } else if let xyz = maybeXyz {
                bar()
            }
            """
        #expect(Self.positions(source) == ["1:4", "3:11"])
    }

    @Test func repeatedBindingListIsFound() {
        let source = """
            if let x = maybeAbc, let z = x.maybeY {
                foo()
            } else if let x = maybeAbc, let z = x.maybeY {
                bar()
            }
            """
        #expect(Self.positions(source) == ["1:4", "3:11"])
    }

    @Test func repeatedAvailabilityIsFound() {
        let source = """
            if #available(macOS 10.15, *) {
                foo()
            } else if #available(macOS 10.15, *) {
                bar()
            }
            """
        #expect(Self.positions(source) == ["1:4", "3:11"])
    }

    @Test func repeatedCasePatternConditionIsFound() {
        let source = """
            if case .p = x {
                foo()
            } else if case .p = x {
                bar()
            }
            """
        #expect(Self.positions(source) == ["1:4", "3:11"])
    }

    @Test func everyCopyIsFound() {
        let source = """
            if x < 5 {}
            else if x < 5 {}
            else if x < 5 {}
            """
        #expect(Self.positions(source) == ["1:4", "2:9", "3:9"])
    }

    /* SwiftLint compares source text and misses this; the tokens are the same, so it is the same test. */
    @Test func spacingAndCommentsDoNotHideACopy() {
        let source = """
            if x < 5 {
                foo()
            } else if x<5 /* again */ {
                bar()
            }
            """
        #expect(Self.positions(source) == ["1:4", "3:11"])
    }

    /* Enum case patterns are spelled like calls and run no code, in a switch and in `if case`. */
    @Test func enumCasePatternsAreCompared() {
        let source = """
            switch result {
            case .success(let value):
                use(value)
            case .failure(let error):
                report(error)
            case Result.success(let value):
                use(value)
            case .success(let value):
                use(value)
            }
            if case .some(.ready(let value)) = state {
                use(value)
            } else if case .some(.ready(let value)) = state {
                use(value)
            }
            """
        #expect(Self.positions(source) == ["2:6", "8:6", "11:4", "13:11"])
    }

    /* A call in a branch between two pairs splits the chain; copies on one side of it are still found. */
    @Test func copiesOnOneSideOfACallAreFound() {
        let source = """
            if queue.isEmpty {
                wait()
            } else if queue.isEmpty {
                wait()
            } else if queue.refill() {
                drain()
            } else if queue.isFull {
                drain()
            } else if queue.isFull {
                drain()
            }
            """
        #expect(Self.positions(source) == ["1:4", "3:11", "7:11", "9:11"])
    }

    @Test func incumbentNonTriggeringExamplesAreNot() {
        let source = """
            if x < 5 {
                foo()
            } else if y == "s" {
                bar()
            }
            if x < 5 {
                foo()
            }
            if x < 5 {
                bar()
            }
            if x < 5, y == "s" {
                foo()
            } else if x < 5 {
                bar()
            }
            switch x {
            case "a":
                foo()
                bar()
            }
            switch x {
            case "a" where y == "s":
                foo()
            case "a" where y == "t":
                bar()
            }
            if let x = maybeAbc {
                foo()
            } else if let x = maybePqr {
                bar()
            }
            if let x = maybeAbc, let z = x.maybeY {
                foo()
            } else if let x = maybePqr, let z = x.maybeY {
                bar()
            }
            if case .p = x {
                foo()
            } else if case .q = x {
                bar()
            }
            if true {
                if true { foo() }
            }
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* Each evaluation is a new call and may answer differently, so both branches can run. SwiftLint flags these. */
    @Test func conditionsThatCallAreNot() {
        let source = """
            if reader.next() == "{" {
                openBlock()
            } else if reader.next() == "{" {
                openNestedBlock()
            }
            if let line = input.readLine() {
                use(line)
            } else if let line = input.readLine() {
                use(line)
            }
            if case .ready = makeState() {
                start()
            } else if case .ready = makeState() {
                start()
            }
            if await worker.isIdle {
                assign()
            } else if await worker.isIdle {
                assign()
            }
            if #isEnabled(flag) {
                go()
            } else if #isEnabled(flag) {
                go()
            }
            switch value {
            case makeLimit():
                clamp()
            case makeLimit():
                clamp()
            case .some(let other) where other > threshold():
                use(other)
            case .some(let other) where other > threshold():
                use(other)
            case .wrapped(compute()):
                use(value)
            case .wrapped(compute()):
                use(value)
            }
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* The call between the copies runs before the second test and may change what it reads. */
    @Test func copiesSeparatedByACallAreNot() {
        let source = """
            if queue.isEmpty {
                wait()
            } else if queue.refill() {
                drain()
            } else if queue.isEmpty {
                wait()
            }
            switch token {
            case expected:
                accept()
            case _ where advance():
                skip()
            case expected:
                accept()
            }
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* Near misses: a branch nested in a body is a new chain, a subset is not a copy, and two `#if` clauses never compile together. */
    @Test func nearMissesAreNot() {
        let source = """
            if x < 5 {
                foo()
            } else {
                if x < 5 { bar() }
            }
            if x < 5, y < 5 {
                foo()
            } else if x < 5, y < 5, z < 5 {
                bar()
            }
            switch platform {
            #if os(macOS)
            case .desktop:
                foo()
            #else
            case .desktop:
                bar()
            #endif
            default:
                baz()
            }
            switch pair {
            case (.a, .b):
                foo()
            case (.b, .a):
                bar()
            }
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* An `if` used as a value is still a chain, and `case` items in one `#if` clause still compare with each other. */
    @Test func chainsInOtherPositionsAreFound() {
        let source = """
            let label = if x < 5 { "low" } else if x < 5 { "lower" } else { "high" }
            switch platform {
            #if os(macOS)
            case .desktop:
                foo()
            case .desktop:
                bar()
            #endif
            default:
                baz()
            }
            """
        #expect(Self.positions(source) == ["1:16", "1:40", "4:6", "6:6"])
    }
}
