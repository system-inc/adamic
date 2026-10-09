import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `contains-over-first-not-nil` both ways. The unit cases parse a source string and hand the rule symbols built
 by position, one occurrence at every `first` and `firstIndex` name token, so each case says what the compiler
 would have resolved. The triggering cases are SwiftLint's documented examples for `contains_over_first_not_nil`,
 each asserting its line and column; the non-triggering cases are SwiftLint's own and the look-alikes the types
 tell apart. The end-to-end case runs a real package, so the symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct ContainsOverFirstNotNilTests {
    static let sequenceFirst = "s:STsE5first5where7ElementQzSgSbADKXE_tKF"
    static let collectionFirstIndex = "s:SlsE10firstIndex5where5IndexQzSgSb7ElementQzKXE_tKF"
    static let collectionFirstIndexOf = "s:SlsSQ7ElementRpzrlE10firstIndex2of5IndexQzSgAB_tF"
    static let oursFirst = "s:7Control5QueryV5first5whereSiSgSbSiXE_tF"

    /*
     The findings, as `line:column`, with every `first` token resolved to `first` and every `firstIndex` token to
     `firstIndex`. A nil symbol leaves that name unresolved.
     */
    static func findings(
        _ source: String,
        first: String? = sequenceFirst,
        firstIndex: String? = collectionFirstIndex,
    ) -> [String] {
        records(source, first: first, firstIndex: firstIndex).map { "\($0.line):\($0.column)" }
    }

    static func records(
        _ source: String,
        first: String? = sequenceFirst,
        firstIndex: String? = collectionFirstIndex,
    ) -> [FindingRecord] {
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
            if token.text == "first", let first {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: first,
                        name: "first(where:)",
                        isReference: true,
                    )
                )
            }
            if token.text == "firstIndex", let firstIndex {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: firstIndex,
                        name: "firstIndex(where:)",
                        isReference: true,
                    )
                )
            }
        }
        return ContainsOverFirstNotNil().findings(in: file, symbols: FileSymbols(occurrences))
    }

    /* SwiftLint's triggering examples for both methods and both comparisons, each at the start of the call. */
    @Test(arguments: ["first", "firstIndex"], ["!=", "=="])
    func incumbentTriggeringExamplesAreFound(method: String, comparison: String) {
        let source = """
            myList.\(method) { $0 % 2 == 0 } \(comparison) nil
            myList.\(method)(where: { $0 % 2 == 0 }) \(comparison) nil
            myList.map { $0 + 1 }.\(method)(where: { $0 % 2 == 0 }) \(comparison) nil
            myList.\(method)(where: someFunction) \(comparison) nil
            myList.map { $0 + 1 }.\(method) { $0 % 2 == 0 } \(comparison) nil
            (myList.\(method) { $0 % 2 == 0 }) \(comparison) nil

            """
        #expect(Self.findings(source) == ["1:1", "2:1", "3:1", "4:1", "5:1", "6:2"])
    }

    /* SwiftLint's non-triggering examples: the search's result kept, not compared with nil. */
    @Test(arguments: ["first", "firstIndex"])
    func incumbentNonTriggeringExamplesAreNotFound(method: String) {
        let source = """
            let \(method) = myList.\(method)(where: { $0 % 2 == 0 })
            let \(method) = myList.\(method) { $0 % 2 == 0 }

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* `firstIndex(of:)` is SwiftLint's shape too, by name, and its repair is `contains(_:)`; the rest repair to `contains(where:)`. */
    @Test func theMessageNamesTheRepairForTheCallFound() {
        let source = """
            let first = items.firstIndex(of: target) == nil
            let second = items.firstIndex(where: { $0 > 1 }) != nil
            let third = items.first { $0 > 1 } != nil

            """
        let messages = Self.records(source, firstIndex: Self.collectionFirstIndexOf).map(\.message)
        #expect(messages.count == 3)
        #expect(messages.first?.contains("Comparing firstIndex(of:) with nil") == true)
        #expect(messages.first?.contains("Use contains(_:) (or !contains(_:) for == nil)") == true)
        #expect(
            messages.dropFirst().allSatisfy { $0.contains("Use contains(where:) (or !contains(where:) for == nil)") }
        )
        #expect(messages.last?.contains("Comparing first(where:) with nil") == true)
        #expect(
            Self.records(source, firstIndex: Self.collectionFirstIndexOf).allSatisfy {
                $0.messageId == "containsOverFirstNotNil"
            }
        )
    }

    /* The reason the rule is typed: a query builder's own `first(where:)` reads the same and may have no `contains` beside it. */
    @Test func aFirstOfOursIsNotFound() {
        let source = """
            let first = query.first { $0 > 1 } != nil
            let second = query.firstIndex(where: { $0 > 1 }) == nil

            """
        #expect(Self.findings(source, first: Self.oursFirst, firstIndex: Self.oursFirst).isEmpty)
        #expect(Self.findings(source, first: nil, firstIndex: nil).isEmpty)
    }

    /* Every standard library search by these names is found: a set's, a dictionary's, an async sequence's and a lazy sequence's. */
    @Test(arguments: [
        "s:Sh10firstIndex2ofSh5IndexVyx_GSgx_tF", "s:SD10firstIndex5whereSD5IndexVyxq__GSgSbx3key_q_5valuet_tKXE_tKF",
        "s:ScisE5first5where7ElementQzSgSbADYaKXE_tYaKF",
    ])
    func otherStandardLibrarySearchesAreFound(symbol: String) {
        let source = """
            let first = values.firstIndex(of: target) != nil
            let second = try await stream.first { $0 > 1 } == nil
            let third = items.lazy.map { $0 + 1 }.first { $0 > 1 } != nil

            """
        #expect(Self.findings(source, first: symbol, firstIndex: symbol) == ["1:13", "2:24", "3:13"])
    }

    /* `nil` on the left reads the same question, a bare call in a collection's own extension is one, and a comparison inside a larger expression is still one, found by folding. */
    @Test func mirroredBareAndNestedComparisonsAreFound() {
        let source = """
            let first = nil != items.first { $0 > 1 }
            let second = ready && items.firstIndex(of: target) == nil
            let third = items.first(where: { $0 > 1 }) == nil ? "none" : "some"
            extension Array where Element == Int {
                var hasLarge: Bool { first { $0 > 1 } != nil }
                var hasSmall: Bool { self.firstIndex(where: { $0 < 1 }) != nil }
            }

            """
        #expect(Self.findings(source) == ["1:20", "2:23", "3:13", "5:26", "6:26"])
    }

    /* Not the shape: the `first` property, a comparison with something other than nil, `===`, and a search reached through an optional chain. */
    @Test func comparisonsThatAreNotTheShapeAreNotFound() {
        let source = """
            let first = items.first != nil
            let second = items.firstIndex(of: target) == items.startIndex
            let third = items.first { $0 > 1 } === nil
            let fourth = items?.first { $0 > 1 } != nil
            let fifth = owner?.items.firstIndex(of: target) == nil
            let sixth = owner?.groups[0].first(where: { $0 > 1 }) != nil
            let seventh = items.first { $0 > 1 }?.value != nil
            let eighth = items.contains { $0 > 1 }

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A force unwrap with no `?` before it is no optional chain, so the search after it is found. */
    @Test func aForceUnwrappedBaseIsFound() {
        #expect(Self.findings("let first = owner!.items.first { $0 > 1 } != nil\n") == ["1:13"])
    }

    @Test(arguments: [
        "let found = items.firstIndex(of: target)\n", "let found = items.contains(target) ? target : nil\n",
    ])
    func aFileWithNoFirstOrNoNilDoesNotApply(source: String) {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!ContainsOverFirstNotNil().applies(to: file))
    }

    /*
     End to end on a real package, symbols from the index the build wrote: an array's `first(where:)`,
     `firstIndex(where:)` and `firstIndex(of:)` compared with nil are found, and a query builder's own `first`
     and `firstIndex` are not.
     */
    static let packageSource = """
        struct Query {
            var terms: [Int]

            func first(where predicate: (Int) -> Bool) -> Int? {
                terms.last
            }

            func firstIndex(of term: Int) -> Int? {
                nil
            }
        }

        func checks(items: [Int], query: Query) -> [Bool] {
            let some = items.first { $0 > 1 } != nil
            let none = items.firstIndex(where: { $0 > 2 }) == nil
            let held = items.firstIndex(of: 3) != nil
            let ours = query.first { $0 > 4 } != nil
            let oursIndexed = query.firstIndex(of: 5) == nil
            return [some, none, held, ours, oursIndexed]
        }

        """

    @Test func anArraysSearchIsFlaggedAndAQuerysIsNot() async throws {
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
        let candidates = parsed.files.filter { ContainsOverFirstNotNil().applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            ContainsOverFirstNotNil().findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map
            { "\($0.line):\($0.column)" }
        }
        #expect(
            found == ["14:16", "15:16", "16:16"],
            "expected the array's three searches and not the query's: \(found)",
        )
    }
}
