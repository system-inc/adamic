import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `first-where` both ways. The unit cases parse a source string and hand the rule symbols built by position, one
 occurrence at every `filter` name token and every `first` or `last` name token, so each case says what the
 compiler would have resolved; the symbols are the ones a real build's index recorded. The triggering cases are
 SwiftLint's documented examples for `first_where`, each asserting its line and column; the non-triggering
 cases are SwiftLint's own and the look-alikes the types tell apart. The end-to-end case runs a real package, so
 the symbols come from the index the build wrote. `LastWhereTests` reads through the same helpers.
 */
@Suite(.serialized)
struct FirstWhereTests {
    static let arrayFilter = "s:s14_ArrayProtocolPsE6filterySay7ElementQzGSbAEqd__YKXEqd__YKs5ErrorRd__lF"
    static let sequenceFilter = "s:STsE6filterySay7ElementQzGSbACqd__YKXEqd__YKs5ErrorRd__lF"
    static let rangeReplaceableFilter = "s:SmsE6filteryxSb7ElementQzqd__YKXEqd__YKs5ErrorRd__lF"
    static let substringFilter = "s:Ss6filterySSSbSJxYKXExYKs5ErrorRzlF"
    static let dictionaryFilter = "s:SD6filterySDyxq_GSbx3key_q_5valuet_tqd__YKXEqd__YKs5ErrorRd__lF"
    static let setFilter = "s:Sh6filteryShyxGSbxqd__YKXEqd__YKs5ErrorRd__lF"
    static let lazyFilter = "s:s20LazySequenceProtocolPsE6filterys0a6FilterB0Vy8ElementsQzGSb7ElementQzcF"
    static let asynchronousFilter = "s:Sci12_ConcurrencyE6filteryAA19AsyncFilterSequenceVyxGSb7ElementQzYacF"
    static let oursFilter = "s:7Control5QueryV6filteryACSbSiXEF"
    /* Realm's query `filter(_ predicateFormat: String, _ args: Any...)`, spelled as the index would spell a declaration in `RealmSwift`. */
    static let realmFilter = "s:10RealmSwift7ResultsV6filteryACyxGSS_ypdtF"
    static let collectionFirst = "s:SlsE5first7ElementQzSgvp"
    static let sequenceFirstWhere = "s:STsE5first5where7ElementQzSgSbADKXE_tKF"
    static let oursFirst = "s:7Control5QueryV5firstSiSgvp"
    /* Members an `extension Sequence` and an `extension Collection` of ours add: they name the type first, as the standard library's do, and our module next. */
    static let oursSequenceFilter = "s:ST7ControlE6filterySay7ElementQzGs7KeyPathCyADSbGF"
    static let oursCollectionFirst = "s:Sl7ControlE5first7ElementQzSgvp"

    /*
     The findings of `rule`, as `messageId@line:column`, with every `filter` token resolved to `filter` and
     every `first` and `last` token to `member`. A nil symbol leaves that name unresolved.
     */
    static func findings(
        _ source: String,
        rule: some TypedFileRule = FirstWhere(),
        filter: String? = arrayFilter,
        member: String? = collectionFirst,
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
            if token.text == "first" || token.text == "last", let member {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: member,
                        name: token.text,
                        isReference: true,
                    )
                )
                if member.hasSuffix("vp") {
                    occurrences.append(
                        FileSymbols.Occurrence(
                            line: location.line,
                            column: location.column,
                            symbol: member.dropLast(2) + "vg",
                            name: "getter:\(token.text)",
                            isReference: true,
                            isImplicit: true,
                        )
                    )
                }
            }
        }
        return rule.findings(in: file, symbols: FileSymbols(occurrences)).map {
            "\($0.messageId)@\($0.line):\($0.column)"
        }
    }

    /* SwiftLint's `first_where` triggering examples, each at the start of the `filter` call. */
    @Test func incumbentTriggeringExamplesAreFound() {
        let source = """
            _ = myList.filter { $0 % 2 == 0 }.first
            _ = myList.filter({ $0 % 2 == 0 }).first
            _ = myList.map { $0 + 1 }.filter({ $0 % 2 == 0 }).first
            _ = myList.map { $0 + 1 }.filter({ $0 % 2 == 0 }).first?.something()
            _ = myList.filter(someFunction).first
            _ = myList.filter({ $0 % 2 == 0 })
            .first
            _ = (myList.filter { $0 == 1 }).first
            _ = myListOfDict.filter { dict in dict["1"] }.first
            _ = myListOfDict.filter { $0["someString"] }.first

            """
        #expect(
            Self.findings(source) == [
                "firstWhere@1:5", "firstWhere@2:5", "firstWhere@3:5", "firstWhere@4:5", "firstWhere@5:5",
                "firstWhere@6:5", "firstWhere@8:6", "firstWhere@9:5", "firstWhere@10:5",
            ]
        )
    }

    /* SwiftLint's `first_where` non-triggering examples that are not the shape whatever the types. */
    @Test func incumbentNonTriggeringExamplesAreNotFound() {
        let source = """
            _ = kinds.filter(excludingKinds.contains).isEmpty && kinds.first == .identifier
            _ = myList.first(where: { $0 % 2 == 0 })
            _ = match(pattern: pattern).filter { $0.first == .identifier }
            _ = (myList.filter { $0 == 1 }.suffix(2)).first

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* SwiftLint's Realm examples, skipped there by the spelling of the argument and here because the `filter` is Realm's. */
    @Test func incumbentRealmExamplesAreNotFound() {
        let source = """
            _ = collection.filter("stringCol = '3'").first
            _ = realm?.objects(User.self).filter(NSPredicate(format: "email ==[c] %@", email)).first
            if let pause = timeTracker.pauses.filter("beginDate < %@", beginDate).first { print(pause) }

            """
        #expect(Self.findings(source, filter: Self.realmFilter).isEmpty)
    }

    /* Every standard library `filter` that builds a new collection: array, sequence, string, substring, dictionary and set alike. */
    @Test(arguments: [
        arrayFilter, sequenceFilter, rangeReplaceableFilter, substringFilter, dictionaryFilter, setFilter,
    ])
    func everyEagerFilterIsFound(filter: String) {
        #expect(Self.findings("_ = values.filter { _ in true }.first\n", filter: filter) == ["firstWhere@1:5"])
    }

    /* A lazy or asynchronous `filter` builds nothing and its `first` stops at the first match. */
    @Test(arguments: [lazyFilter, asynchronousFilter])
    func aLazyFilterIsNotFound(filter: String) {
        let source = """
            _ = items.lazy.filter { $0 > 1 }.first
            _ = items.lazy.filter { $0 > 1 }.first { $0 > 2 }

            """
        #expect(Self.findings(source, filter: filter).isEmpty)
        #expect(Self.findings(source, filter: filter, member: Self.sequenceFirstWhere).isEmpty)
    }

    /* The reason the rule is typed: a query builder's own `filter` and `first` read the same and allocate nothing. */
    @Test func aFilterOfOursIsNotFound() {
        let source = "_ = query.filter { $0 > 1 }.first\n"
        #expect(Self.findings(source, filter: Self.oursFilter, member: Self.oursFirst).isEmpty)
        #expect(Self.findings(source, filter: Self.oursFilter).isEmpty)
    }

    /* A `filter` or a `first` an extension of ours adds to a standard library protocol is ours, not the standard library's. */
    @Test func membersOurExtensionsAddToStandardLibraryProtocolsAreNotFound() {
        let source = "_ = items.filter(\\.isActive).first\n"
        #expect(Self.findings(source, filter: Self.oursSequenceFilter).isEmpty)
        #expect(Self.findings(source, member: Self.oursCollectionFirst).isEmpty)
    }

    /* A standard library `filter` whose `first` resolved to ours is not the shape; nor is one whose names resolved to nothing. */
    @Test func aFirstOrFilterNotKnownToBeTheStandardLibrarysIsNotFound() {
        let source = "_ = items.filter { $0 > 1 }.first\n"
        #expect(Self.findings(source, member: Self.oursFirst).isEmpty)
        #expect(Self.findings(source, member: nil).isEmpty)
        #expect(Self.findings(source, filter: nil).isEmpty)
    }

    /* `first(where:)` called on a filtered collection still builds it, as SwiftLint reads it: one `first(where:)` with both predicates is the repair. */
    @Test func firstWhereOnAFilteredCollectionIsFound() {
        let source = """
            _ = items.filter { $0 > 1 }.first { $0 > 2 }
            _ = items.filter { $0 > 1 }.first(where: { $0 > 2 })

            """
        #expect(Self.findings(source, member: Self.sequenceFirstWhere) == ["firstWhere@1:5", "firstWhere@2:5"])
    }

    /* Through optional chaining, force unwrapping and a comparison with nil, the `first` is still read from the filter. */
    @Test func readsAroundTheFirstAreFound() {
        let source = """
            _ = items?.filter { $0 > 1 }.first
            _ = items.filter { $0 > 1 }.first!
            _ = items.filter { $0 > 1 }.first != nil

            """
        #expect(Self.findings(source) == ["firstWhere@1:5", "firstWhere@2:5", "firstWhere@3:5"])
    }

    /* A bare `filter` in a collection's own extension, a parenthesised call, and `self.filter`. */
    @Test func bareParenthesisedAndSelfFiltersAreFound() {
        let source = """
            extension Array where Element == Int {
                var firstLarge: Int? { filter { $0 > 1 }.first }
                var firstSmall: Int? { (self.filter { $0 < 1 }).first }
            }

            """
        #expect(Self.findings(source) == ["firstWhere@2:28", "firstWhere@3:29"])
    }

    /* An unapplied `first(where:)` reference and a `first` read from something other than the filter call are not the shape. */
    @Test func nearMissesAreNotFound() {
        let source = """
            let pick = items.filter { $0 > 1 }.first(where:)
            _ = items.filter { $0 > 1 }.sorted().first
            _ = items.filter { $0 > 1 }.last
            _ = first.filter { $0 > 1 }

            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func aFileWithNoFilterOrNoFirstDoesNotApply() {
        for source in ["_ = items.first\n", "_ = items.filter { $0 > 1 }.last\n"] {
            let file = ParsedFile(
                url: URL(fileURLWithPath: "/Plain.swift"),
                targetName: "Control",
                targetKind: "regular",
                source: source,
                tree: Parser.parse(source: source),
                nodeCount: 0,
            )
            #expect(!FirstWhere().applies(to: file))
        }
    }

    /*
     One run on a real package of one file, which builds it and writes its index. The rule is not registered
     here, so it is run by hand on what the run left: the findings as `messageId@line`.
     */
    static func packageFindings(source: String, rule: some TypedFileRule) async throws -> [String] {
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
        try source.write(to: sources.appendingPathComponent("Control.swift"), atomically: true, encoding: .utf8)
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
        let candidates = parsed.files.filter { rule.applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        return candidates.flatMap { file in
            rule.findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.messageId)@\($0.line)"
            }
        }
    }

    /* End to end, symbols from the index: an array's and a dictionary's values' `filter` are found, and a query builder's own `filter` and `first` and a lazy `filter` are not. */
    @Test func anArraysFilterIsFlaggedAndAQuerysIsNot() async throws {
        let source = """
            struct Query {
                var terms: [Int]

                var first: Int? { terms.first }

                func filter(_ isIncluded: (Int) -> Bool) -> Query {
                    Query(terms: terms)
                }
            }

            func picks(items: [Int], names: [String: Int], query: Query) -> [Int?] {
                let array = items.filter { $0 > 1 }.first
                let values = names.values.filter { $0 > 2 }.first
                let ours = query.filter { $0 > 3 }.first
                let lazy = items.lazy.filter { $0 > 4 }.first
                return [array, values, ours, lazy]
            }

            """
        let found = try await Self.packageFindings(source: source, rule: FirstWhere())
        #expect(
            found == ["firstWhere@12", "firstWhere@13"],
            "expected the array's and the values' filter and not the query's or the lazy one: \(found)",
        )
    }
}
