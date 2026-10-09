import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `identical-operands`, both directions. The failing cases are SwiftLint's documented
 triggering examples for every one of the eight comparison operators, each asserting the exact line and
 column of the comparison's first token. The passing cases are SwiftLint's documented non-triggering
 examples, the same eight times, and the near misses this rule adds on its own: calls, macros and `await`,
 which may give two values from one spelling.
 */
struct IdenticalOperandsTests {
    static let comparisonOperators = ["==", "!=", "===", "!==", ">", ">=", "<", "<="]

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
        return IdenticalOperands().findings(in: file).map { "\($0.line):\($0.column)" }
    }

    /* SwiftLint's triggering examples, one per line so each asserts its own column. */
    @Test(arguments: comparisonOperators)
    func incumbentTriggeringExamplesAreFound(comparison: String) {
        let source = """
            1 \(comparison) 1
            foo \(comparison) foo
            foo.aProperty \(comparison) foo.aProperty
            self.aProperty \(comparison) self.aProperty
            $0 \(comparison) $0
            a?.b \(comparison) a?.b
            if (elem \(comparison) elem) {}
            XCTAssertTrue(s3 \(comparison) s3)
            if let tab = tabManager.selectedTab, tab.webView \(comparison) tab.webView {}
            1    + 1 \(comparison)   1     +    1
            """
        #expect(Self.positions(source) == ["1:1", "2:1", "3:1", "4:1", "5:1", "6:1", "7:5", "8:15", "9:38", "10:1"])
    }

    /* SwiftLint's two `&&` examples: either side of a conjunction, across lines. */
    @Test func eitherSideOfAConjunctionIsFound() {
        let first = """
            func same(lhs: Item, rhs: Item) -> Bool {
                return lhs.foo == lhs.foo &&
                       lhs.bar == rhs.bar
            }
            """
        #expect(Self.positions(first) == ["2:12"])
        let second = """
            func same(lhs: Item, rhs: Item) -> Bool {
                return lhs.foo == rhs.foo &&
                       lhs.bar == lhs.bar
            }
            """
        #expect(Self.positions(second) == ["3:12"])
    }

    @Test(arguments: comparisonOperators)
    func incumbentNonTriggeringExamplesAreNot(comparison: String) {
        let source = """
            1 \(comparison) 2
            foo \(comparison) bar
            prefixedFoo \(comparison) foo
            foo.aProperty \(comparison) foo.anotherProperty
            self.aProperty \(comparison) self.anotherProperty
            let quoted = "1 \(comparison) 1"
            self.aProperty \(comparison) aProperty
            lhs.aProperty \(comparison) rhs.aProperty
            lhs.identifier \(comparison) rhs.identifier
            i \(comparison) index
            $0 \(comparison) 0
            keyValues?.count ?? 0 \(comparison) 0
            string \(comparison) string.lowercased()
            let num: Int? = 0
            _ = num != nil && num \(comparison) num?.byteSwapped
            num \(comparison) num!.byteSwapped
            1    + 1 \(comparison)   1     +    2
            f(  i :   2) \(comparison)   f (i: 3 )
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* SwiftLint's remaining non-triggering examples: angle brackets that are generics, and operands that only look alike. */
    @Test func incumbentLookalikesAreNot() {
        let source = """
            func evaluate(_ mode: CommandMode) -> Result<Options, CommandantError<CommandantError<()>>>
            let array = Array<Array<Int>>()
            guard Set(identifiers).count != identifiers.count else { return }
            expect("foo") == "foo"
            type(of: model).cachePrefix == cachePrefix
            histogram[156].0 == 0x003B8D96 && histogram[156].1 == 1
            [Wrapper(type: .three), Wrapper(type: .one)].sorted { "\\($0.type)" > "\\($1.type)"}
            array.sorted { "\\($0)" < "\\($1)" }
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* The old NaN test is found: `value.isNaN` says it in words. A reflexivity test is found: compare against a copy. */
    @Test func nanChecksAndReflexivityTestsAreFound() {
        let source = """
            if value != value { return }
            let isNumber = ratio == ratio
            #expect(item == item)
            """
        #expect(Self.positions(source) == ["1:4", "2:16", "3:9"])
    }

    /* Subscripts, key paths, string literals and nested comparisons are read values, so they are found; the inner comparison is the one reported. */
    @Test func readsAreFound() {
        let source = """
            let same = values[index] == values[index]
            let path = \\Item.name == \\Item.name
            let text = "a" < "a"
            let nested = (left == left) == flag
            """
        #expect(Self.positions(source) == ["1:12", "2:12", "3:12", "4:15"])
    }

    /*
     Where we differ from SwiftLint: two calls are two evaluations. Neighbouring elements, uniqueness, the
     current time, and an actor's value across a suspension can all differ between the two sides.
     */
    @Test func evaluationsThatMayDifferAreNot() {
        let source = """
             f(  i :   2) ==   f (i:
             2 )
            let neighbours = iterator.next() == iterator.next()
            #expect(UUID() != UUID())
            let later = Date() < Date()
            let moved = await counter.value == await counter.value
            let line = #line == #line
            let trailing = items.first { $0.isOn } == items.first { $0.isOn }
            """
        #expect(Self.positions(source).isEmpty)
    }

    /*
     A package's own operator folds at a guessed precedence, so its whole flat sequence is skipped, on
     either side of the comparison. Parentheses make their own sequence, and the comparison beside them is read.
     */
    @Test func sequencesWithUnknownOperatorsAreSkipped() {
        let source = """
            let first = left <=> right == left <=> right
            let second = left <=> right && value == value
            let third = value == value && left <=> right
            let fourth = (left <=> right) && value == value
            """
        #expect(Self.positions(source) == ["4:34"])
    }

    /* Operators outside the eight are not comparisons. */
    @Test func otherOperatorsAreNot() {
        let source = """
            let sum = value + value
            let both = flag && flag
            let either = flag || flag
            let fallback = optional ?? optional
            value = value
            let spaceship = left <=> left
            let range = start ..< start
            """
        #expect(Self.positions(source).isEmpty)
    }
}
