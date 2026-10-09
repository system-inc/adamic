import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `cohere-swift/duplicated-key-in-dictionary-literal`, both directions. The failing cases include
 SwiftLint's four triggering examples and assert the exact line and column of each repeat. The passing
 cases include SwiftLint's non-triggering examples, plus the near misses: different spellings of what might
 be the same value, calls and subscripts that may differ per evaluation, NaN, and `KeyValuePairs`.
 */
struct DuplicatedKeyInDictionaryLiteralTests {
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
        return DuplicatedKeyInDictionaryLiteral().findings(in: file).map { "\($0.line):\($0.column)" }
    }

    @Test func repeatedIntegerKeyIsFound() {
        let source = """
            let values = [
                1: "1",
                2: "2",
                1: "one"
            ]
            """
        #expect(Self.positions(source) == ["4:5"])
    }

    @Test func repeatedStringKeyIsFound() {
        let source = """
            let values = [
                "1": 1,
                "2": 2,
                "2": 2
            ]
            """
        #expect(Self.positions(source) == ["4:5"])
    }

    @Test func repeatedNameKeyIsFound() {
        let source = """
            let values = [
                foo: "1",
                bar: "2",
                baz: "3",
                foo: "4",
                zaz: "5"
            ]
            """
        #expect(Self.positions(source) == ["5:5"])
    }

    @Test func repeatedImplicitMemberKeyIsFound() {
        let source = """
            let values = [
                .one: "1",
                .two: "2",
                .three: "3",
                .one: "1",
                .four: "4",
                .five: "5"
            ]
            """
        #expect(Self.positions(source) == ["5:5"])
    }

    /* Every repeat after the first is reported, each against the key it repeats, never the first occurrence. */
    @Test func everyLaterRepeatIsFound() {
        let source = """
            let values = [.one: 1, .two: 2, .one: 3, .two: 4, .one: 5]
            """
        #expect(Self.positions(source) == ["1:33", "1:42", "1:51"])
    }

    @Test func otherLiteralAndReferenceKeysAreFound() {
        let source = """
            let booleans = [true: 1, false: 2, true: 3]
            let optionals = [nil: 1, nil: 2] as [Int?: Int]
            let negatives = [-1: "a", 1: "b", -1: "c"]
            let floats = [1.5: "a", 1.5: "b"]
            let qualified = [Kind.one: 1, Kind.one: 2]
            let chained = [self.base.name: 1, self.base.name: 2]
            let paths = [\\Item.name: 1, \\Item.name: 2]
            let parenthesized = [(foo): 1, (foo): 2]
            let interpolated = ["\\(prefix)-a": 1, "\\(prefix)-a": 2]
            let raw = [#"a"#: 1, #"a"#: 2]
            let chains = [owner?.name: 1, owner?.name: 2]
            let forced = [owner!.name: 1, owner!.name: 2]
            """
        #expect(
            Self.positions(source) == [
                "1:36", "2:26", "3:35", "4:25", "5:31", "6:35", "7:29", "8:32", "9:39", "10:22", "11:31", "12:31",
            ]
        )
    }

    /* Spacing and comments are not part of the key, so these are the same spelling. */
    @Test func triviaDoesNotHideARepeat() {
        let source = """
            let values = [
                Kind.one: 1,
                Kind . one /* again */: 2
            ]
            """
        #expect(Self.positions(source) == ["3:5"])
    }

    /* Each literal is its own scope: the same key in an outer and an inner literal is not a repeat, but a repeat inside the inner one is. */
    @Test func nestedLiteralsAreCheckedSeparately() {
        let source = """
            let values = ["a": ["a": 1, "b": 2], "b": ["a": 3, "a": 4]]
            """
        #expect(Self.positions(source) == ["1:52"])
    }

    @Test func distinctKeysAreNot() {
        let source = """
            let numbers = [1: "1", 2: "2"]
            let strings = ["1": 1, "2": 2]
            let names = [foo: "1", bar: "2"]
            let members = [.one: 1, .two: 2]
            let empty: [String: Int] = [:]
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* SwiftLint's own exemptions: a call or a macro may give a different value each time it is evaluated. */
    @Test func callsAndMacrosAreNot() {
        let source = """
            let identifiers = [UUID(): "1", UUID(): "2"]
            let lines = [#line: "1", #line: "2"]
            let made = [make(1): "a", make(1): "b"]
            let members = [Kind.make(): 1, Kind.make(): 2]
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* Keys whose value is not decided by their spelling: subscripts, operators, casts, calls inside an interpolation or a key path. */
    @Test func evaluatedKeysAreNot() {
        let source = """
            let subscripts = [items[0]: 1, items[0]: 2]
            let sums = [base + 1: 1, base + 1: 2]
            let casts = [value as Int: 1, value as Int: 2]
            let interpolatedCalls = ["\\(make())": 1, "\\(make())": 2]
            let paths = [\\Item.items[0]: 1, \\Item.items[0]: 2]
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* Different spellings are never claimed equal, even when the values are, because that needs evaluation. */
    @Test func differentSpellingsOfOneValueAreNot() {
        let source = """
            let numbers = [1: "a", 0x1: "b", 1.0: "c"]
            let strings = ["A": 1, "\\u{41}": 2, #"A"#: 3]
            let members = [.one: 1, Kind.one: 2]
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* NaN is never equal to itself, so two NaN keys are two entries and nothing crashes. */
    @Test func notANumberKeysAreNot() {
        let source = """
            let values = [Double.nan: 1, Double.nan: 2, .signalingNaN: 3, .signalingNaN: 4]
            """
        #expect(Self.positions(source).isEmpty)
    }

    /* `KeyValuePairs` keeps repeated keys on purpose. */
    @Test func keyValuePairsAreNot() {
        let source = """
            let annotated: KeyValuePairs<String, Int> = ["a": 1, "a": 2]
            let cast = ["a": 1, "a": 2] as KeyValuePairs<String, Int>
            let mirror = Mirror(self, children: ["a": 1, "a": 2])
            func headers() -> KeyValuePairs<String, String> {
                ["Set-Cookie": first, "Set-Cookie": second]
            }
            """
        #expect(Self.positions(source).isEmpty)
    }
}
