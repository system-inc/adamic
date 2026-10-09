import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `sorted-first-last` both ways. The unit cases parse a source string and hand the rule symbols built by
 position, one occurrence at every `sorted`, `first` and `last` name token, so each case says what the compiler
 would have resolved. The triggering cases are SwiftLint's documented examples for `sorted_first_last`, each
 asserting its message id, line and column; the non-triggering cases are SwiftLint's own and the look-alikes
 the types tell apart. The end-to-end case runs a real package, so the symbols come from the index the build
 wrote.
 */
@Suite(.serialized)
struct SortedFirstLastTests {
    typealias Resolution = (symbol: String, name: String)

    static let sortedAscending: Resolution = ("s:STsSL7ElementRpzrlE6sortedSayABGyF", "sorted()")
    static let sortedBy: Resolution = ("s:STsE6sorted2bySay7ElementQzGSbAD_ADtKXE_tKF", "sorted(by:)")
    static let collectionFirst: Resolution = ("s:SlsE5first7ElementQzSgvp", "first")
    static let bidirectionalLast: Resolution = ("s:SKsE4last7ElementQzSgvp", "last")
    static let oursSorted: Resolution = ("s:7Control6LedgerV6sortedACyF", "sorted()")
    static let oursFirst: Resolution = ("s:7Control6LedgerV5firstSiSgvp", "first")
    static let oursLast: Resolution = ("s:7Control6LedgerV4lastSiSgvp", "last")
    static let oursKeyPathSorted: Resolution = (
        "s:ST7ControlE6sorted2bySay7ElementQzGs7KeyPathCyAEqd__G_tSLRd__lF", "sorted(by:)",
    )
    static let foundationSorted: Resolution = (
        "s:ST10FoundationE6sorted5usingSay7ElementQzGqd___tAA14SortComparatorRd__8ComparedQyd__AERSlF",
        "sorted(using:)",
    )

    /* The standard library's `sorted()` where the call is written with empty parentheses, and `sorted(by:)` otherwise. */
    static func standardSorted(_ token: TokenSyntax) -> Resolution? {
        let open = token.nextToken(viewMode: .sourceAccurate)
        let close = open?.nextToken(viewMode: .sourceAccurate)
        return open?.tokenKind == .leftParen && close?.tokenKind == .rightParen ? sortedAscending : sortedBy
    }

    /*
     The findings, as `messageId@line:column`, with every `sorted` token resolved by `sorted` and every `first`
     and `last` token to the given declaration, whatever follows it. A nil resolution leaves that name
     unresolved.
     */
    static func findings(
        _ source: String,
        sorted: (TokenSyntax) -> Resolution? = standardSorted,
        first: Resolution? = collectionFirst,
        last: Resolution? = bidirectionalLast,
    ) -> [String] {
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
            let resolution: Resolution? =
                switch token.text {
                    case "sorted": sorted(token)
                    case "first": first
                    case "last": last
                    default: nil
                }
            guard let resolution else { continue }
            occurrences.append(
                FileSymbols.Occurrence(
                    line: location.line,
                    column: location.column,
                    symbol: resolution.symbol,
                    name: resolution.name,
                    isReference: true,
                )
            )
            if token.text != "sorted" {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: "s:SlsE5first7ElementQzSgvg",
                        name: "getter:\(token.text)",
                        isReference: true,
                        isImplicit: true,
                    )
                )
            }
        }
        return SortedFirstLast().findings(in: file, symbols: FileSymbols(occurrences)).map {
            "\($0.messageId)@\($0.line):\($0.column)"
        }
    }

    /* SwiftLint's triggering examples, each at the start of the chain. */
    @Test func incumbentTriggeringExamplesAreFound() {
        let source = """
            myList.sorted().first
            myList.sorted(by: { $0.description < $1.description }).first
            myList.sorted(by: >).first
            myList.map { $0 + 1 }.sorted().first
            myList.sorted(by: someFunction).first
            myList.map { $0 + 1 }.sorted { $0.description < $1.description }.first
            myList.sorted().last
            myList.sorted().last?.something()
            myList.sorted(by: { $0.description < $1.description }).last
            myList.map { $0 + 1 }.sorted().last
            myList.sorted(by: someFunction).last
            myList.map { $0 + 1 }.sorted { $0.description < $1.description }.last
            myList.map { $0 + 1 }.sorted { $0.first < $1.first }.last

            """
        #expect(
            Self.findings(source) == [
                "sortedFirst@1:1", "sortedFirst@2:1", "sortedFirst@3:1", "sortedFirst@4:1", "sortedFirst@5:1",
                "sortedFirst@6:1",
                "sortedLast@7:1", "sortedLast@8:1", "sortedLast@9:1", "sortedLast@10:1", "sortedLast@11:1",
                "sortedLast@12:1", "sortedLast@13:1",
            ]
        )
    }

    /*
     SwiftLint's non-triggering examples. Every `first` and `last` here is handed the property's symbol and
     every `sorted` the standard library's, so what keeps them out is the shape: another label, another name,
     or a `first` that is called.
     */
    @Test func incumbentNonTriggeringExamplesAreNotFound() {
        let source = """
            let min = myList.min()
            let min = myList.min(by: { $0 < $1 })
            let min = myList.min(by: >)
            let max = myList.max()
            let max = myList.max(by: { $0 < $1 })
            let message = messages.sorted(byKeyPath: #keyPath(Message.timestamp)).last
            let message = messages.sorted(byKeyPath: "timestamp", ascending: false).first
            myList.sorted().firstIndex(of: key)
            myList.sorted().lastIndex(of: key)
            myList.sorted().firstIndex(where: someFunction)
            myList.sorted().lastIndex(where: someFunction)
            myList.sorted().firstIndex { $0 == key }
            myList.sorted().lastIndex { $0 == key }
            myList.sorted().first(where: someFunction)
            myList.sorted().last(where: someFunction)
            myList.sorted().first { $0 == key }
            myList.sorted().last { $0 == key }

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* The reason the rule is typed: a ledger's own `sorted`, `first` and `last` read the same and sort nothing. */
    @Test func aSortedOfOursIsNotFound() {
        let source = """
            let first = ledger.sorted().first
            let second = ledger.sorted().last

            """
        #expect(
            Self.findings(source, sorted: { _ in Self.oursSorted }, first: Self.oursFirst, last: Self.oursLast).isEmpty
        )
        #expect(Self.findings(source, sorted: { _ in Self.oursSorted }).isEmpty)
    }

    /*
     A `sorted(by:)` taking a key path, declared in an extension of `Sequence` of ours, passes the label check
     and is spelled like the standard library's (`s:ST...`), but is ours. This rule's own module check keeps it
     out, and since this case was found, so does `isStandardLibrary`, which reads the extending module.
     */
    @Test func aKeyPathSortedOfOursOnSequenceIsNotFound() {
        let source = "let first = scores.sorted(by: \\.name).first\n"
        let ours = FileSymbols.Occurrence(
            line: 1,
            column: 1,
            symbol: Self.oursKeyPathSorted.symbol,
            name: Self.oursKeyPathSorted.name,
            isReference: true,
        )
        #expect(
            !ours.isStandardLibrary,
            "an extension of ours on a standard library protocol read as the standard library's",
        )
        #expect(Self.findings(source, sorted: { _ in Self.oursKeyPathSorted }).isEmpty)
    }

    /* Foundation's `sorted(using:)` sorts a standard library sequence, but has no `min(using:)` to move to; the label keeps it out. */
    @Test func foundationsSortedUsingIsNotFound() {
        let source = "let first = scores.sorted(using: KeyPathComparator(\\Score.name)).first\n"
        #expect(Self.findings(source, sorted: { _ in Self.foundationSorted }).isEmpty)
        #expect(Self.findings(source, sorted: { _ in Self.sortedBy }).isEmpty)
    }

    /* A standard library `sorted` whose end resolved to ours is not the shape; nor is one whose names resolved to nothing. */
    @Test func aNameNotKnownToBeTheStandardLibrarysIsNotFound() {
        let source = """
            let first = items.sorted().first
            let second = items.sorted().last

            """
        #expect(Self.findings(source, first: Self.oursFirst, last: Self.oursLast).isEmpty)
        #expect(Self.findings(source, first: nil, last: nil).isEmpty)
        #expect(Self.findings(source, sorted: { _ in nil }).isEmpty)
        #expect(Self.findings(source, first: Self.bidirectionalLast, last: Self.collectionFirst).isEmpty)
    }

    /* A lazy chain still reaches `Sequence.sorted`, which builds and sorts an array, so it is found like any other. */
    @Test func aLazyChainIsFound() {
        let source = """
            let first = items.lazy.map { $0 + 1 }.sorted().first
            let second = items.lazy.filter { $0 > 1 }.sorted(by: >).last

            """
        #expect(Self.findings(source) == ["sortedFirst@1:13", "sortedLast@2:14"])
    }

    /* A bare `sorted` in a sequence's own extension, `self.sorted`, a parenthesised call, an optional chain and a force unwrap. */
    @Test func bareSelfParenthesisedAndChainedReadsAreFound() {
        let source = """
            extension Sequence where Element: Comparable {
                var smallest: Element? { sorted().first }
                var largest: Element? { self.sorted().last }
                var also: Element? { (sorted()).first }
            }
            let fourth = items?.sorted().first
            let fifth = items.sorted().last!.name
            let sixth = [3, 1, 2]
                .sorted()
                .first

            """
        #expect(
            Self.findings(source) == [
                "sortedFirst@2:30", "sortedLast@3:29", "sortedFirst@4:26", "sortedFirst@6:14", "sortedLast@7:13",
                "sortedFirst@8:13",
            ]
        )
    }

    /* Reads that are not the shape: an end by subscript or prefix, two pairs of parentheses, and an unapplied `first(where:)`. */
    @Test func readsThatAreNotTheShapeAreNotFound() {
        let source = """
            let first = items.sorted()[0]
            let second = items.sorted().prefix(1)
            let third = ((items.sorted())).first
            let fourth = items.sorted().first(where:)
            let fifth = items.sorted(by: >, extra).first

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* The findings on a source whose every `sorted`, `first` and `last` resolves to the standard library's. */
    static func records(_ source: String) -> [FindingRecord] {
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
            let resolution: Resolution? =
                switch token.text {
                    case "sorted": standardSorted(token)
                    case "first": collectionFirst
                    case "last": bidirectionalLast
                    default: nil
                }
            guard let resolution else { continue }
            occurrences.append(
                FileSymbols.Occurrence(
                    line: location.line,
                    column: location.column,
                    symbol: resolution.symbol,
                    name: resolution.name,
                    isReference: true,
                )
            )
        }
        return SortedFirstLast().findings(in: file, symbols: FileSymbols(occurrences))
    }

    /* The source with every finding's suggested edits applied, last first so each offset still points where it did. */
    static func repaired(_ source: String) -> String {
        var bytes = Array(source.utf8)
        for edit in records(source).flatMap({ $0.suggestions.flatMap(\.fixes) }).sorted(by: { $0.start > $1.start }) {
            bytes.replaceSubrange(edit.start..<edit.end, with: Array(edit.text.utf8))
        }
        return String(decoding: bytes, as: UTF8.self)
    }

    /*
     The repair follows the comparator's direction: a descending `>`, alone or the one comparison of a closure,
     makes the read the other one of `min` and `max` with `<`, so "newest" reads `max { $0.created < $1.created }`.
     A forward comparator, no comparator, and a compound predicate keep what they say.
     */
    @Test func aDescendingComparatorIsRepairedAsTheOtherEndWithTheForwardOne() {
        let source = """
            let newest = posts.sorted { $0.created > $1.created }.first
            let oldest = posts.sorted { $0.created > $1.created }.last
            let largest = values.sorted(by: >).first
            let smallest = values.sorted(by: >).last
            let returned = posts.sorted(by: { first, second in return first.created > second.created }).first
            let ascending = values.sorted(by: { $0 < $1 }).first
            let plain = values.sorted().last
            let compound = posts.sorted { $0.rank > $1.rank || $0.name < $1.name }.first
            let chained = [3, 1, 2]
                .sorted(by: >)
                .first

            """
        let expected = """
            let newest = posts.max { $0.created < $1.created }
            let oldest = posts.min { $0.created < $1.created }
            let largest = values.max(by: <)
            let smallest = values.min(by: <)
            let returned = posts.max(by: { first, second in return first.created < second.created })
            let ascending = values.min(by: { $0 < $1 })
            let plain = values.max()
            let compound = posts.min { $0.rank > $1.rank || $0.name < $1.name }
            let chained = [3, 1, 2]
                .max(by: <)

            """
        #expect(Self.repaired(source) == expected)
    }

    /* The message says the repair the suggestion makes, and the tie the `last` side still has to check. */
    @Test func theMessageNamesTheRepair() throws {
        let source = """
            let newest = posts.sorted { $0.created > $1.created }.first
            let smallest = values.sorted(by: >).last
            let ascending = values.sorted(by: { $0 < $1 }).first
            let plain = values.sorted().last

            """
        let findings = Self.records(source)
        #expect(
            findings.map { $0.suggestions.map(\.message) } == [
                ["Use max(by:) with < for >"], ["Use min(by:) with < for >"], ["Use min(by:)"], ["Use max()"],
            ]
        )
        #expect(
            findings.first?.message
                == "Sorting a collection to read its first element builds and sorts a whole new array to keep one element. Use max(by:) with the predicate's > turned to <, which reads as the largest by the forward order rather than the smallest by a reversed one: it says what is meant, and it walks the elements once."
        )
        let smallest = try #require(findings.dropFirst().first)
        #expect(
            smallest.message.contains(
                "Use min(by:) with the predicate's > turned to <, which reads as the smallest by the forward order rather than the largest by a reversed one"
            )
        )
        #expect(
            smallest.message.contains(
                "Among equally ordered elements min returns the first where sorted().last returned the last"
            )
        )
        #expect(findings.dropFirst(2).first?.message.contains("Use min(by:) with the same predicate:") == true)
        #expect(findings.last?.message.contains("Use max():") == true)
    }

    @Test func aFileWithNoSortedDoesNotApply() {
        let source = "let first = items.first\nlet last = items.last\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!SortedFirstLast().applies(to: file))
    }

    /*
     End to end on a real package, symbols from the index the build wrote: an array's, a set's, a dictionary's,
     a string's and a lazy chain's `sorted` are found at either end, and a ledger's own `sorted`, a key path
     `sorted(by:)` of ours on `Sequence`, Foundation's `sorted(using:)` and a `first(where:)` are not.
     */
    static let packageSource = """
        import Foundation

        struct Score {
            var value: Int
            var name: String
        }

        struct Ledger {
            func sorted() -> Ledger { self }
            var first: Int? { nil }
            var last: Int? { nil }
        }

        extension Sequence {
            func sorted<Key: Comparable>(by key: KeyPath<Element, Key>) -> [Element] {
                sorted { $0[keyPath: key] < $1[keyPath: key] }
            }
        }

        func checks(values: [Int], scores: [Score], ledger: Ledger, set: Set<Int>, table: [String: Int], text: String) {
            let first = values.sorted().first
            let second = values.sorted(by: >).last
            let third = scores.sorted { $0.value < $1.value }.first
            let fourth = values.lazy.map { $0 + 1 }.sorted().first
            let fifth = set.sorted().last
            let sixth = table.sorted { $0.value < $1.value }.first
            let seventh = text.sorted().last
            let eighth = ledger.sorted().first
            let ninth = scores.sorted(by: \\.name).first
            let tenth = scores.sorted(using: KeyPathComparator(\\Score.name)).first
            let eleventh = values.sorted().first(where: { $0 > 1 })
            let twelfth = (values.sorted()).last
            _ = (first, second, third, fourth, fifth, sixth, seventh, eighth, ninth, tenth, eleventh, twelfth)
        }

        """

    @Test func aStandardLibrarySortIsFlaggedAndOursIsNot() async throws {
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
        let candidates = parsed.files.filter { SortedFirstLast().applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            SortedFirstLast().findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.messageId)@\($0.line)"
            }
        }
        #expect(
            found == [
                "sortedFirst@21", "sortedLast@22", "sortedFirst@23", "sortedFirst@24", "sortedLast@25",
                "sortedFirst@26", "sortedLast@27", "sortedLast@32",
            ],
            "expected the standard library's sorts and not the ledger's, the key path one of ours, sorted(using:) or first(where:): \(found)",
        )
    }
}
