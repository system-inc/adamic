import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `contains-over-filter` both ways. The unit cases parse a source string and hand the rule symbols built by
 position, one occurrence at every `filter` and `count` name token, so each case says what the compiler would
 have resolved. The triggering cases are SwiftLint's documented examples for `contains_over_filter_count` and
 `contains_over_filter_is_empty`, each asserting its message id, line and column; the non-triggering cases are
 SwiftLint's own and the look-alikes the types tell apart. The end-to-end case runs a real package, so the
 symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct ContainsOverFilterTests {
    static let arrayFilter = "s:s14_ArrayProtocolPsE6filterySay7ElementQzGSbAEKXEKF"
    static let arrayCount = "s:Sa5countSivp"
    static let oursFilter = "s:7Control5QueryV6filteryACSbSiXEF"
    static let oursCount = "s:7Control5QueryV5countSivp"
    static let lazyFilter = "s:s20LazySequenceProtocolPsE6filtery0a6FilterB0Vy8ElementsQzGSb7ElementQzcF"

    /*
     The findings, as `messageId@line:column`, with every `filter` token resolved to `filter` and every
     `count` token to `count`. A nil symbol leaves that name unresolved.
     */
    static func findings(_ source: String, filter: String? = arrayFilter, count: String? = arrayCount) -> [String] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        var occurrences: [FileSymbols.Occurrence] = []
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            if token.text == "filter", let filter {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: filter,
                        name: "filter(_:)",
                        isReference: true,
                    )
                )
            }
            if token.text == "count", let count {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: count,
                        name: "count",
                        isReference: true,
                    )
                )
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: "s:Sa5countSivg",
                        name: "getter:count",
                        isReference: true,
                        isImplicit: true,
                    )
                )
            }
        }
        return ContainsOverFilter().findings(in: file, symbols: FileSymbols(occurrences)).map {
            "\($0.messageId)@\($0.line):\($0.column)"
        }
    }

    /* SwiftLint's `contains_over_filter_count` triggering examples for `>`, `==` and `!=`, each at the start of the comparison. */
    @Test(arguments: [">", "==", "!="])
    func incumbentCountTriggeringExamplesAreFound(comparison: String) {
        let source = """
            let first = myList.filter(where: { $0 % 2 == 0 }).count \(comparison) 0
            let second = myList.filter { $0 % 2 == 0 }.count \(comparison) 0
            let third = myList.filter(where: someFunction).count \(comparison) 0

            """
        #expect(Self.findings(source) == ["filterCount@1:13", "filterCount@2:14", "filterCount@3:13"])
    }

    /* SwiftLint's `contains_over_filter_count` non-triggering examples: a count compared with one, `01` which is not zero, and `contains` itself. */
    @Test(arguments: [">", "==", "!="])
    func incumbentCountNonTriggeringExamplesAreNotFound(comparison: String) {
        let source = """
            let first = myList.filter(where: { $0 % 2 == 0 }).count \(comparison) 1
            let second = myList.filter { $0 % 2 == 0 }.count \(comparison) 1
            let third = myList.filter(where: { $0 % 2 == 0 }).count \(comparison) 01
            let fourth = myList.contains(where: { $0 % 2 == 0 })
            let fifth = !myList.contains(where: { $0 % 2 == 0 })
            let sixth = myList.contains(10)

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* SwiftLint's `contains_over_filter_is_empty` triggering examples, each at the start of the `isEmpty` read, after a `!`. */
    @Test func incumbentIsEmptyTriggeringExamplesAreFound() {
        let source = """
            let first = myList.filter(where: { $0 % 2 == 0 }).isEmpty
            let second = !myList.filter(where: { $0 % 2 == 0 }).isEmpty
            let third = myList.filter { $0 % 2 == 0 }.isEmpty
            let fourth = myList.filter(where: someFunction).isEmpty

            """
        #expect(
            Self.findings(source) == [
                "filterIsEmpty@1:13", "filterIsEmpty@2:15", "filterIsEmpty@3:13", "filterIsEmpty@4:14",
            ]
        )
    }

    /* SwiftLint's `contains_over_filter_is_empty` non-triggering examples. */
    @Test func incumbentIsEmptyNonTriggeringExamplesAreNotFound() {
        let source = """
            let first = myList.filter(where: { $0 % 2 == 0 }).count > 1
            let second = myList.filter { $0 % 2 == 0 }.count == 1
            let third = myList.contains(where: { $0 % 2 == 0 })
            let fourth = !myList.contains(where: { $0 % 2 == 0 })
            let fifth = myList.contains(10)

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* The reason the rule is typed: a query builder's own `filter`, `count` and `isEmpty` read the same and allocate nothing. */
    @Test func aFilterOfOursIsNotFound() {
        let source = """
            let first = query.filter { $0 > 1 }.isEmpty
            let second = query.filter { $0 > 1 }.count > 0

            """
        #expect(Self.findings(source, filter: Self.oursFilter, count: Self.oursCount).isEmpty)
        #expect(Self.findings(source, filter: Self.oursFilter).isEmpty)
    }

    /* A standard library `filter` whose `count` resolved to ours is not the shape; nor is one whose names resolved to nothing. */
    @Test func aCountOrFilterNotKnownToBeTheStandardLibrarysIsNotFound() {
        let source = "let first = items.filter { $0 > 1 }.count > 0\n"
        #expect(Self.findings(source, count: Self.oursCount).isEmpty)
        #expect(Self.findings(source, count: nil).isEmpty)
        #expect(Self.findings(source, filter: nil).isEmpty)
        #expect(Self.findings("let first = items.filter { $0 > 1 }.isEmpty\n", filter: nil).isEmpty)
    }

    /* A lazy filter builds nothing and its `isEmpty` stops at the first match, so neither shape is found on it. */
    @Test func aLazyFilterIsNotFound() {
        let source = """
            let first = items.lazy.filter { $0 > 1 }.isEmpty
            let second = items.lazy.filter { $0 > 1 }.count > 0

            """
        #expect(Self.findings(source, filter: Self.lazyFilter).isEmpty)
    }

    /* Every standard library collection's `filter` builds a new collection: dictionary, set and string alike. */
    @Test(arguments: [
        ("s:SD6filteryAByxq_GSbx3key_q_5valuet_tKXEKF", "s:SD5countSivp"),
        ("s:Sh6filteryShyxGSbxKXEKF", "s:Sh5countSivp"), ("s:SS6filteryS2SSbSJKXEKF", "s:SS5countSivp"),
    ])
    func otherStandardLibraryCollectionsAreFound(filter: String, count: String) {
        let source = """
            let first = values.filter { _ in true }.isEmpty
            let second = values.filter { _ in true }.count != 0

            """
        #expect(Self.findings(source, filter: filter, count: count) == ["filterIsEmpty@1:13", "filterCount@2:14"])
    }

    /* Zero on the left reads the same question, and a comparison inside a larger expression is still one, found by folding. */
    @Test func mirroredAndNestedComparisonsAreFound() {
        let source = """
            let first = 0 < items.filter { $0 > 1 }.count
            let second = 0 == items.filter { $0 > 1 }.count
            let third = 0 != items.filter { $0 > 1 }.count
            let fourth = ready && items.filter { $0 > 1 }.count > 0
            let fifth = items.filter { $0 > 1 }.count == 0 ? "none" : "some"

            """
        #expect(
            Self.findings(source) == [
                "filterCount@1:13", "filterCount@2:14", "filterCount@3:13", "filterCount@4:23", "filterCount@5:13",
            ]
        )
    }

    /* Comparisons that are not the shape: zero the wrong way round, a sum compared with zero, and the same question asked with one. */
    @Test func comparisonsThatAreNotTheShapeAreNotFound() {
        let source = """
            let first = 0 > items.filter { $0 > 1 }.count
            let second = total + items.filter { $0 > 1 }.count == 0
            let third = items.filter { $0 > 1 }.count >= 1
            let fourth = items.filter { $0 > 1 }.count < 1
            let fifth = items.filter { $0 > 1 }.count >= 0
            let sixth = items.filter { $0 > 1 }.first == nil
            let seventh = items.filter { $0 > 1 }.count + 0

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Zero as SwiftLint reads it: any radix and underscores. */
    @Test func zeroInAnySpellingIsFound() {
        let source = """
            let first = items.filter { $0 > 1 }.count > 0x0
            let second = items.filter { $0 > 1 }.count > 0_0
            let third = items.filter { $0 > 1 }.count > 00
            let fourth = items.filter { $0 > 1 }.count > 0b0

            """
        #expect(
            Self.findings(source) == ["filterCount@1:13", "filterCount@2:14", "filterCount@3:13", "filterCount@4:14"]
        )
    }

    /* A bare `filter` in a collection's own extension, a parenthesised call, and `self.filter`. */
    @Test func bareParenthesisedAndSelfFiltersAreFound() {
        let source = """
            extension Array where Element == Int {
                var hasLarge: Bool { !filter { $0 > 1 }.isEmpty }
                var hasSmall: Bool { (self.filter { $0 < 1 }).count > 0 }
            }

            """
        #expect(Self.findings(source) == ["filterIsEmpty@2:27", "filterCount@3:26"])
    }

    @Test func aFileWithNoFilterDoesNotApply() {
        let source = "let first = items.count > 0\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!ContainsOverFilter().applies(to: file))
    }

    /*
     End to end on a real package, symbols from the index the build wrote: an array's `filter` is found in both
     shapes, and a query builder's own `filter`, `count` and `isEmpty` and a lazy `filter` are not.
     */
    static let packageSource = """
        struct Query {
            var terms: [Int]

            var count: Int { terms.count }
            var isEmpty: Bool { terms.isEmpty }

            func filter(_ isIncluded: (Int) -> Bool) -> Query {
                Query(terms: terms)
            }
        }

        func checks(items: [Int], query: Query) -> [Bool] {
            let none = items.filter { $0 > 1 }.isEmpty
            let some = items.filter { $0 > 2 }.count > 0
            let ours = query.filter { $0 > 3 }.isEmpty
            let oursCounted = query.filter { $0 > 4 }.count != 0
            let lazy = items.lazy.filter { $0 > 5 }.isEmpty
            return [none, some, ours, oursCounted, lazy]
        }

        """

    @Test func anArraysFilterIsFlaggedAndAQuerysIsNot() async throws {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-typed-\(UUID().uuidString)",
            isDirectory: true,
        )
        let sources = root.appendingPathComponent("Sources/Control", isDirectory: true)
        try FileManager.default.createDirectory(at: sources, withIntermediateDirectories: true)
        try PipelineControlTests.manifest.write(
            to: root.appendingPathComponent("Package.swift"),
            atomically: true,
            encoding: .utf8,
        )
        try Self.packageSource.write(
            to: sources.appendingPathComponent("Control.swift"),
            atomically: true,
            encoding: .utf8,
        )

        /* One run builds the package and writes its index. The rule is not registered here, so it is run by hand on what the run left. */
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix"],
            workingDirectory: root,
        )
        _ = try await Pipeline(options: options, writer: ContractWriter { _ in }, workingDirectory: root).run()
        let package = try PackageModel.load(
            root: root,
            scratchPath: Pipeline.scratchPath(for: root),
            runner: ProcessRunner(),
        )
        let parsed = await SourceParser().parse(try FileSet.build(package: package).owned)
        let candidates = parsed.files.filter { ContainsOverFilter().applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            ContainsOverFilter().findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.messageId)@\($0.line)"
            }
        }
        #expect(
            found == ["filterIsEmpty@13", "filterCount@14"],
            "expected the array's two shapes and not the query's or the lazy filter's: \(found)",
        )
    }
}
