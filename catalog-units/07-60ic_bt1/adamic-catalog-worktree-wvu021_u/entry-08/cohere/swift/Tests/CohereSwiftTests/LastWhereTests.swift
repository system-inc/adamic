import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 `last-where` both ways, through `FirstWhereTests`' helpers: symbols built by position for the unit cases,
 and a real package for the end-to-end case. The triggering cases are SwiftLint's documented examples for
 `last_where`, each asserting its line and column; the non-triggering cases are SwiftLint's own, the look-alikes
 the types tell apart, and the one this rule adds: a `filter` whose receiver may have no `last(where:)`.
 */
@Suite(.serialized)
struct LastWhereTests {
    static let bidirectionalLast = "s:SKsE4last7ElementQzSgvp"
    static let bidirectionalLastWhere = "s:SKsE4last5where7ElementQzSgSbADKXE_tKF"
    static let oursLast = "s:7Control5QueryV4lastSiSgvp"
    /* A `last` an `extension BidirectionalCollection` of ours adds: the standard library's type first, our module next. */
    static let oursBidirectionalLast = "s:SK7ControlE4last7ElementQzSgvp"

    /* The findings of `last-where`, as `messageId@line:column`, with every `filter` token resolved to `filter` and every `last` token to `member`. */
    static func findings(
        _ source: String,
        filter: String? = FirstWhereTests.arrayFilter,
        member: String? = bidirectionalLast,
    ) -> [String] {
        FirstWhereTests.findings(source, rule: LastWhere(), filter: filter, member: member)
    }

    /* SwiftLint's `last_where` triggering examples, each at the start of the `filter` call. */
    @Test func incumbentTriggeringExamplesAreFound() {
        let source = """
            _ = myList.filter { $0 % 2 == 0 }.last
            _ = myList.filter({ $0 % 2 == 0 }).last
            _ = myList.map { $0 + 1 }.filter({ $0 % 2 == 0 }).last
            _ = myList.map { $0 + 1 }.filter({ $0 % 2 == 0 }).last?.something()
            _ = myList.filter(someFunction).last
            _ = myList.filter({ $0 % 2 == 0 })
            .last
            _ = (myList.filter { $0 == 1 }).last

            """
        #expect(
            Self.findings(source) == [
                "lastWhere@1:5", "lastWhere@2:5", "lastWhere@3:5", "lastWhere@4:5", "lastWhere@5:5", "lastWhere@6:5",
                "lastWhere@8:6",
            ]
        )
    }

    /* SwiftLint's `last_where` non-triggering examples that are not the shape whatever the types. */
    @Test func incumbentNonTriggeringExamplesAreNotFound() {
        let source = """
            _ = kinds.filter(excludingKinds.contains).isEmpty && kinds.last == .identifier
            _ = myList.last(where: { $0 % 2 == 0 })
            _ = match(pattern: pattern).filter { $0.last == .identifier }
            _ = (myList.filter { $0 == 1 }.suffix(2)).last

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* SwiftLint's Realm example, skipped there by the spelling of the argument and here because the `filter` is Realm's. */
    @Test func incumbentRealmExampleIsNotFound() {
        #expect(
            Self.findings("_ = collection.filter(\"stringCol = '3'\").last\n", filter: FirstWhereTests.realmFilter)
                .isEmpty
        )
    }

    /* The `filter`s whose receiver is known to be bidirectional: an array's, a string's (or any range-replaceable collection's) and a substring's. */
    @Test(arguments: [
        FirstWhereTests.arrayFilter, FirstWhereTests.rangeReplaceableFilter, FirstWhereTests.substringFilter,
    ])
    func everyBidirectionalFilterIsFound(filter: String) {
        #expect(Self.findings("_ = values.filter { _ in true }.last\n", filter: filter) == ["lastWhere@1:5"])
    }

    /*
     `Sequence`'s `filter` returns an array, which has a `last`, from a receiver that may have no `last(where:)`:
     `dictionary.values.filter { }.last` compiles and `dictionary.values.last(where:)` does not. Not found.
     */
    @Test func aSequencesFilterIsNotFound() {
        let source = """
            _ = dictionary.values.filter { $0 > 1 }.last
            _ = zip(left, right).filter { $0.0 > 1 }.last

            """
        #expect(Self.findings(source, filter: FirstWhereTests.sequenceFilter).isEmpty)
    }

    /* A lazy or asynchronous `filter` builds nothing and its `last` already searches from the end. */
    @Test(arguments: [FirstWhereTests.lazyFilter, FirstWhereTests.asynchronousFilter])
    func aLazyFilterIsNotFound(filter: String) {
        let source = """
            _ = items.lazy.filter { $0 > 1 }.last
            _ = items.lazy.filter { $0 > 1 }.last { $0 > 2 }

            """
        #expect(Self.findings(source, filter: filter).isEmpty)
        #expect(Self.findings(source, filter: filter, member: Self.bidirectionalLastWhere).isEmpty)
    }

    /* The reason the rule is typed: a query builder's own `filter` and `last` read the same and allocate nothing. */
    @Test func aFilterOfOursIsNotFound() {
        let source = "_ = query.filter { $0 > 1 }.last\n"
        #expect(Self.findings(source, filter: FirstWhereTests.oursFilter, member: Self.oursLast).isEmpty)
        #expect(Self.findings(source, filter: FirstWhereTests.oursFilter).isEmpty)
        #expect(Self.findings(source, member: Self.oursLast).isEmpty)
        #expect(Self.findings(source, member: nil).isEmpty)
    }

    /* A `filter` or a `last` an extension of ours adds to a standard library protocol is ours, not the standard library's. */
    @Test func membersOurExtensionsAddToStandardLibraryProtocolsAreNotFound() {
        let source = "_ = items.filter(\\.isActive).last\n"
        #expect(Self.findings(source, filter: FirstWhereTests.oursSequenceFilter).isEmpty)
        #expect(Self.findings(source, member: Self.oursBidirectionalLast).isEmpty)
    }

    /* `last(where:)` called on a filtered collection still builds it, as SwiftLint reads it. */
    @Test func lastWhereOnAFilteredCollectionIsFound() {
        let source = """
            _ = items.filter { $0 > 1 }.last { $0 > 2 }
            _ = items.filter { $0 > 1 }.last(where: { $0 > 2 })

            """
        #expect(Self.findings(source, member: Self.bidirectionalLastWhere) == ["lastWhere@1:5", "lastWhere@2:5"])
    }

    /* A bare `filter` in a collection's own extension, and a parenthesised `self.filter`. */
    @Test func bareAndParenthesisedFiltersAreFound() {
        let source = """
            extension Array where Element == Int {
                var lastLarge: Int? { filter { $0 > 1 }.last }
                var lastSmall: Int? { (self.filter { $0 < 1 }).last }
            }

            """
        #expect(Self.findings(source) == ["lastWhere@2:27", "lastWhere@3:28"])
    }

    @Test func aFileWithNoFilterOrNoLastDoesNotApply() {
        for source in ["_ = items.last\n", "_ = items.filter { $0 > 1 }.first\n"] {
            let file = ParsedFile(
                url: URL(fileURLWithPath: "/Plain.swift"),
                targetName: "Control",
                targetKind: "regular",
                source: source,
                tree: Parser.parse(source: source),
                nodeCount: 0,
            )
            #expect(!LastWhere().applies(to: file))
        }
    }

    /* End to end, symbols from the index: an array's `filter` is found, and a dictionary's values', a query builder's own and a lazy one are not. */
    @Test func anArraysFilterIsFlaggedAndAQuerysIsNot() async throws {
        let source = """
            struct Query {
                var terms: [Int]

                var last: Int? { terms.last }

                func filter(_ isIncluded: (Int) -> Bool) -> Query {
                    Query(terms: terms)
                }
            }

            func picks(items: [Int], names: [String: Int], query: Query) -> [Int?] {
                let array = items.filter { $0 > 1 }.last
                let values = names.values.filter { $0 > 2 }.last
                let ours = query.filter { $0 > 3 }.last
                let lazy = items.lazy.filter { $0 > 4 }.last
                return [array, values, ours, lazy]
            }

            """
        let found = try await FirstWhereTests.packageFindings(source: source, rule: LastWhere())
        #expect(
            found == ["lastWhere@12"],
            "expected the array's filter and not the values', the query's or the lazy one: \(found)",
        )
    }
}
