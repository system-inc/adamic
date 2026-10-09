import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `is-disjoint` both ways. The unit cases parse a source string and hand the rule symbols built by position, one
 occurrence at every `intersection` name token, so each case says what the compiler would have resolved. The
 triggering cases are SwiftLint's documented examples for `is_disjoint`, each at the `intersection` name; the
 non-triggering cases are SwiftLint's own and the look-alikes the types tell apart. The end-to-end case runs a
 real package, so the symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct IsDisjointTests {
    static let setIntersection = "s:Sh12intersectionyShyxGABF"
    static let setSequenceIntersection = "s:Sh12intersectionyShyxGqd__7ElementQyd__RszSTRd__lF"
    static let setAlgebraIntersection = "s:s10SetAlgebraP12intersectionyxxF"
    static let optionSetIntersection = "s:s9OptionSetPsE12intersectionyxxF"
    static let rangeSetIntersection = "s:s8RangeSetV12intersectionyAByxGADF"
    static let indexSetIntersection = "s:10Foundation8IndexSetV12intersectionyA2CF"
    static let characterSetIntersection = "s:10Foundation12CharacterSetV12intersectionyA2CF"
    static let rectangleIntersection = "s:So6CGRectV12CoreGraphicsE12intersectionyA2BF"
    static let oursIntersection = "s:7Control3BagV12intersectionySaySiGACF"
    static let oursSetExtensionIntersection = "s:Sh7ControlE12intersection5otherShyxGAC_tF"

    /* The findings, as `messageId@line:column`, with every `intersection` token resolved to `intersection`. A nil leaves it unresolved. */
    static func findings(_ source: String, intersection: String? = setIntersection) -> [FindingRecord] {
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
        for token in file.tree.tokens(viewMode: .sourceAccurate) where token.text == "intersection" {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            if let intersection {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: intersection,
                        name: "intersection(_:)",
                        isReference: true,
                    )
                )
            }
        }
        return IsDisjoint().findings(in: file, symbols: FileSymbols(occurrences))
    }

    static func places(_ source: String, intersection: String? = setIntersection) -> [String] {
        findings(source, intersection: intersection).map { "\($0.messageId)@\($0.line):\($0.column)" }
    }

    /* SwiftLint's `is_disjoint` triggering examples, each at the `intersection` name. */
    @Test func incumbentTriggeringExamplesAreFound() {
        let source = """
            _ = Set(syntaxKinds).intersection(commentAndStringKindsSet).isEmpty
            let isObjc = !objcAttributes.intersection(dictionary.enclosedSwiftAttributes).isEmpty

            """
        #expect(Self.places(source) == ["isDisjoint@1:22", "isDisjoint@2:30"])
    }

    /* SwiftLint's `is_disjoint` non-triggering examples: `isDisjoint(with:)` itself, and an intersection not asked whether it is empty. */
    @Test func incumbentNonTriggeringExamplesAreNotFound() {
        let source = """
            _ = Set(syntaxKinds).isDisjoint(with: commentAndStringKindsSet)
            let isObjc = !objcAttributes.isDisjoint(with: dictionary.enclosedSwiftAttributes)
            _ = Set(syntaxKinds).intersection(commentAndStringKindsSet)
            _ = !objcAttributes.intersection(dictionary.enclosedSwiftAttributes)

            """
        #expect(Self.places(source).isEmpty)
    }

    /*
     The reason the rule is typed: an `intersection` of ours, one our module adds to `Set`, or a rectangle's may
     have no `isDisjoint(with:)`; an unresolved one is unknown.
     */
    @Test func anIntersectionWithoutTheRepairIsNotFound() {
        let source = "let first = bag.intersection(other).isEmpty\n"
        #expect(Self.places(source, intersection: Self.oursIntersection).isEmpty)
        #expect(Self.places(source, intersection: Self.rectangleIntersection).isEmpty)
        #expect(Self.places(source, intersection: Self.oursSetExtensionIntersection).isEmpty)
        #expect(Self.places(source, intersection: nil).isEmpty)
    }

    /* Every `intersection` whose type has `isDisjoint(with:)`: a set's with a set or a sequence, a generic `SetAlgebra`'s, an `OptionSet`'s, and Foundation's `IndexSet`'s and `CharacterSet`'s. */
    @Test(arguments: [
        setIntersection, setSequenceIntersection, setAlgebraIntersection, optionSetIntersection, indexSetIntersection,
        characterSetIntersection,
    ])
    func everyIntersectionWithTheRepairIsFound(intersection: String) {
        let found = Self.findings("let first = left.intersection(right).isEmpty\n", intersection: intersection)
        #expect(found.map { "\($0.messageId)@\($0.line):\($0.column)" } == ["isDisjoint@1:18"])
        #expect(found.allSatisfy { $0.message.contains("isDisjoint(with:)") })
    }

    /* A `RangeSet`'s repair is spelled `isDisjoint(_:)`, and the message says so. */
    @Test func aRangeSetsRepairIsNamedAsItIsSpelled() {
        let found = Self.findings(
            "let first = ranges.intersection(other).isEmpty\n",
            intersection: Self.rangeSetIntersection,
        )
        #expect(found.count == 1)
        #expect(found.allSatisfy { $0.message.contains("isDisjoint(_:)") && !$0.message.contains("isDisjoint(with:)") })
    }

    /* A parenthesised call, as SwiftLint reads it, and a bare `intersection` in a set's own extension. */
    @Test func parenthesisedAndBareIntersectionsAreFound() {
        let source = """
            let first = (left.intersection(right)).isEmpty
            extension Set {
                func sharesNothing(with other: Set) -> Bool { intersection(other).isEmpty }
            }

            """
        #expect(Self.places(source) == ["isDisjoint@1:19", "isDisjoint@3:51"])
    }

    /* Reads that are not the shape: the count of an intersection, a comparison with an empty set, a property named `intersection`. */
    @Test func readsThatAreNotTheShapeAreNotFound() {
        let source = """
            let first = left.intersection(right).count == 0
            let second = left.intersection(right) == []
            let third = shape.intersection.isEmpty
            let fourth = left.intersection(right).first?.isEmpty

            """
        #expect(Self.places(source).isEmpty)
    }

    @Test func aFileWithNoIntersectionDoesNotApply() {
        let source = "let first = items.isEmpty\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!IsDisjoint().applies(to: file))
    }

    /*
     End to end on a real package, symbols from the index the build wrote: a set's, an `IndexSet`'s and a generic
     `SetAlgebra`'s intersections are found, and a bag's own `intersection` and a rectangle's are not.
     */
    static let packageSource = """
        import CoreGraphics
        import Foundation

        struct Bag {
            func intersection(_ other: Bag) -> [Int] { [] }
        }

        func checks(left: Set<Int>, right: Set<Int>, bag: Bag, indexes: IndexSet, other: IndexSet, rectangle: CGRect) -> [Bool] {
            let sets = left.intersection(right).isEmpty
            let ours = bag.intersection(bag).isEmpty
            let foundation = indexes.intersection(other).isEmpty
            let rectangles = rectangle.intersection(rectangle).isEmpty
            return [sets, ours, foundation, rectangles]
        }

        func generic<Value: SetAlgebra>(left: Value, right: Value) -> Bool {
            !left.intersection(right).isEmpty
        }

        """

    @Test func aSetsIntersectionIsFlaggedAndABagsAndARectanglesAreNot() async throws {
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
        let candidates = parsed.files.filter { IsDisjoint().applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            IsDisjoint().findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.messageId)@\($0.line)"
            }
        }
        #expect(
            found == ["isDisjoint@9", "isDisjoint@11", "isDisjoint@17"],
            "expected the set's, the index set's and the generic's intersections and not the bag's or the rectangle's: \(found)",
        )
    }
}
