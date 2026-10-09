import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `reduce-into` both ways. The unit cases parse a source string and hand the rule symbols built by position, the
 way the index records them: a reference at every `reduce` name, the literal initializer the compiler implies at
 the start of every array, dictionary and string literal, and a reference to the type and its initializer at
 every copy-on-write type name. Each case says what the compiler would have resolved. The triggering cases are
 SwiftLint's documented examples for `reduce_into`, each at the `reduce` name; the non-triggering cases are
 SwiftLint's own and the look-alikes the types tell apart. The end-to-end case runs a real package, so the
 symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct ReduceIntoTests {
    static let sequenceReduce = "s:STsE6reduceyqd__qd___qd__qd___7ElementQztKXEtKlF"
    static let sequenceReduceInto = "s:STsE6reduce4into_qd__qd__n_yqd__z_7ElementQztKXEtKlF"
    static let oursReduce = "s:7Control3BagV6reduceyS2S_S2S_SitXEtF"
    static let arrayLiteral = "s:Sa12arrayLiteralSayxGxd_tcfc"
    static let setLiteral = "s:Sh12arrayLiteralShyxGxd_tcfc"
    static let contiguousArrayLiteral = "s:s15ContiguousArrayV12arrayLiteralAByxGxd_tcfc"
    static let oursArrayLiteral = "s:7Control5StackV12arrayLiteralACSid_tcfc"
    static let dictionaryLiteral = "s:SD17dictionaryLiteralSDyxq_Gx_q_td_tcfc"
    static let stringInterpolation = "s:SS19stringInterpolationSSs013DefaultStringB0V_tcfc"
    static let substringLiteral = "s:Ss13stringLiteralSsSS_tcfc"
    static let oursStringLiteral = "s:7Control5LabelV13stringLiteralACSS_tcfc"
    static let standardTypes = [
        "Array": "s:Sa", "ContiguousArray": "s:s15ContiguousArrayV", "Dictionary": "s:SD", "Set": "s:Sh",
        "String": "s:SS",
    ]

    /*
     The findings, as `messageId@line:column`. Every `reduce` name resolves to `reduce`; every array literal
     records `arrayLiteral` and every dictionary literal `dictionaryLiteral`; a plain string literal records
     `stringLiteral` and an interpolated one `interpolation`; every name in `types` resolves to its symbol, with
     an initializer reference beside it as the index writes for `Array<Int>()`. A nil leaves that unrecorded.
     */
    static func findings(
        _ source: String,
        reduce: String? = sequenceReduce,
        arrayLiteral: String? = arrayLiteral,
        dictionaryLiteral: String? = dictionaryLiteral,
        stringLiteral: String? = nil,
        interpolation: String? = stringInterpolation,
        types: [String: String] = standardTypes,
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
        func record(_ symbol: String?, at position: AbsolutePosition, name: String, isImplicit: Bool) {
            guard let symbol else { return }
            let location = file.locations.location(for: position)
            occurrences.append(
                FileSymbols.Occurrence(
                    line: location.line,
                    column: location.column,
                    symbol: symbol,
                    name: name,
                    isReference: true,
                    isImplicit: isImplicit,
                )
            )
        }
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            let position = token.positionAfterSkippingLeadingTrivia
            if token.text == "reduce" {
                record(reduce, at: position, name: "reduce(_:_:)", isImplicit: false)
            }
            if let array = token.parent?.as(ArrayExprSyntax.self), token.id == array.leftSquare.id {
                record(arrayLiteral, at: position, name: "init(arrayLiteral:)", isImplicit: true)
            }
            if let dictionary = token.parent?.as(DictionaryExprSyntax.self), token.id == dictionary.leftSquare.id {
                record(dictionaryLiteral, at: position, name: "init(dictionaryLiteral:)", isImplicit: true)
            }
            if let literal = token.parent?.as(StringLiteralExprSyntax.self), token.id == literal.openingQuote.id {
                record(
                    literal.representedLiteralValue == nil ? interpolation : stringLiteral,
                    at: position,
                    name: "init(stringLiteral:)",
                    isImplicit: true,
                )
            }
            if token.parent?.is(DeclReferenceExprSyntax.self) == true, let type = types[token.text] {
                record(type, at: position, name: token.text, isImplicit: false)
                record("\(type)ycfc", at: position, name: "init()", isImplicit: false)
            }
        }
        return ReduceInto().findings(in: file, symbols: FileSymbols(occurrences)).map {
            "\($0.messageId)@\($0.line):\($0.column)"
        }
    }

    /* SwiftLint's `reduce_into` triggering examples, each at its `reduce` name. */
    @Test(arguments: [
        (#"let bar = values.reduce("abc") { $0 + "\($1)" }"#, "reduceInto@1:18"),
        ("values.reduce(Array<Int>()) { result, value in\n    result += [value]\n}", "reduceInto@1:8"),
        (
            "[1, 2, 3].reduce(Set<Int>()) { acc, value in\n    var result = acc\n    result.insert(value)\n    return result\n}",
            "reduceInto@1:11",
        ),
        (
            "let rows = violations.enumerated().reduce(\"\") { rows, indexAndViolation in\n    return rows + generateSingleRow(for: indexAndViolation.1, at: indexAndViolation.0 + 1)\n}",
            "reduceInto@1:36",
        ),
        (
            "zip(group, group.dropFirst()).reduce([]) { result, pair in\n    result + [pair.0 + pair.1]\n}",
            "reduceInto@1:31",
        ),
        (
            "let foo = values.reduce([String: Int]()) { result, value in\n    var result = result\n    result[\"\\(value)\"] = value\n    return result\n}",
            "reduceInto@1:18",
        ),
        (
            "let bar = values.reduce(Dictionary<String, Int>.init()) { result, value in\n    var result = result\n    result[\"\\(value)\"] = value\n    return result\n}",
            "reduceInto@1:18",
        ),
        (
            "let bar = values.reduce([Int](repeating: 0, count: 10)) { result, value in\n    return result + [value]\n}",
            "reduceInto@1:18",
        ),
        (
            "extension Data {\n    var hexString: String {\n        return reduce(\"\") { (output, byte) -> String in\n            output + String(format: \"%02x\", byte)\n        }\n    }\n}",
            "reduceInto@3:16",
        ),
    ])
    func incumbentTriggeringExamplesAreFound(source: String, expected: String) {
        #expect(Self.findings(source) == [expected])
    }

    /* SwiftLint's `reduce_into` non-triggering examples: `reduce(into:)` in every spelling, and a class as the initial value. */
    @Test func incumbentNonTriggeringExamplesAreNotFound() {
        let source = #"""
            let foo = values.reduce(into: "abc") { $0 += "\($1)" }
            values.reduce(into: Array<Int>()) { result, value in
                result.append(value)
            }
            let rows = violations.enumerated().reduce(into: "") { rows, indexAndViolation in
                rows.append(generateSingleRow(for: indexAndViolation.1, at: indexAndViolation.0 + 1))
            }
            zip(group, group.dropFirst()).reduce(into: []) { result, pair in
                result.append(pair.0 + pair.1)
            }
            let foo = values.reduce(into: [String: Int]()) { result, value in
                result["\(value)"] = value
            }
            let foo = values.reduce(into: Dictionary<String, Int>.init()) { result, value in
                result["\(value)"] = value
            }
            let foo = values.reduce(into: [Int](repeating: 0, count: 10)) { result, value in
                result.append(value)
            }
            let foo = values.reduce(MyClass()) { result, value in
                result.handleValue(value)
                return result
            }

            """#
        #expect(Self.findings(source, reduce: Self.sequenceReduceInto).isEmpty)
        #expect(Self.findings(source).isEmpty)
    }

    /* The reason the rule is typed: a `reduce` of ours reads the same and may copy nothing; an unresolved one is unknown. */
    @Test func aReduceNotKnownToBeTheStandardLibrarysIsNotFound() {
        let source = #"let first = bag.reduce("") { $0 + "\($1)" }"# + "\n"
        #expect(Self.findings(source, reduce: Self.oursReduce).isEmpty)
        #expect(Self.findings(source, reduce: nil).isEmpty)
    }

    /*
     An array literal is a copy-on-write value only where the compiler recorded one being built: an `OptionSet`'s
     `[]` records nothing, a literal type of ours records its own initializer, and an array's, a set's, a contiguous
     array's or a dictionary's records theirs.
     */
    @Test func anArrayOrDictionaryLiteralIsJudgedByTheTypeItBecame() {
        let source = """
            let first = values.reduce([]) { $0 + [$1] }
            let second = pairs.reduce([:]) { $0.merging($1) { $1 } }

            """
        #expect(Self.findings(source) == ["reduceInto@1:20", "reduceInto@2:20"])
        #expect(Self.findings(source, arrayLiteral: Self.setLiteral) == ["reduceInto@1:20", "reduceInto@2:20"])
        #expect(
            Self.findings(source, arrayLiteral: Self.contiguousArrayLiteral, dictionaryLiteral: nil) == [
                "reduceInto@1:20"
            ]
        )
        #expect(Self.findings(source, arrayLiteral: nil, dictionaryLiteral: nil).isEmpty)
        #expect(
            Self.findings(source, arrayLiteral: Self.oursArrayLiteral, dictionaryLiteral: Self.oursArrayLiteral).isEmpty
        )
    }

    /*
     A plain string literal that became a `String` records nothing, so it is read by what it holds: one character
     could be a `Character` and is not found, anything else is. One that records another type's initializer is not
     found, and an interpolated one is found when it records `String`'s interpolation.
     */
    @Test func aStringLiteralIsJudgedByTheTypeItBecameAndWhatItHolds() {
        let plain = #"""
            let first = names.reduce("") { $0 + $1 }
            let second = names.reduce("names: ") { $0 + $1 }
            let third = names.reduce(#"raw"#) { $0 + $1 }
            let fourth = names.reduce(",") { $0 + $1 }
            let fifth = names.reduce("\n") { $0 + $1 }

            """#
        #expect(Self.findings(plain) == ["reduceInto@1:19", "reduceInto@2:20", "reduceInto@3:19"])
        #expect(Self.findings(plain, stringLiteral: Self.substringLiteral).isEmpty)
        #expect(Self.findings(plain, stringLiteral: Self.oursStringLiteral).isEmpty)
        let interpolated = #"let first = values.reduce("\(prefix)") { $0 + "\($1)" }"# + "\n"
        #expect(Self.findings(interpolated) == ["reduceInto@1:20"])
        #expect(Self.findings(interpolated, interpolation: nil).isEmpty)
        #expect(Self.findings(interpolated, interpolation: Self.oursStringLiteral).isEmpty)
    }

    /* A named type is the standard library's only where its name resolved there: a `Set` of our own declaring is not. */
    @Test func aNamedTypeIsJudgedByWhatItsNameResolvedTo() {
        let source = """
            let first = values.reduce(Set<Int>()) { $0.union([$1]) }
            let second = values.reduce(Array()) { $0 + [$1] }

            """
        #expect(Self.findings(source) == ["reduceInto@1:20", "reduceInto@2:21"])
        #expect(Self.findings(source, types: ["Set": "s:7Control3SetV", "Array": "s:7Control5ArrayV"]).isEmpty)
        #expect(Self.findings(source, types: [:]).isEmpty)
    }

    /* Beyond SwiftLint's names: `String(...)`, `ContiguousArray(...)` and `[Int].init()` are copy-on-write initial values too. */
    @Test func stringContiguousArrayAndSugarInitAreFound() {
        let source = """
            let first = values.reduce(String()) { $0 + String($1) }
            let second = values.reduce(ContiguousArray<Int>()) { $0 + [$1] }
            let third = values.reduce([Int].init()) { $0 + [$1] }

            """
        #expect(Self.findings(source) == ["reduceInto@1:20", "reduceInto@2:21", "reduceInto@3:20"])
    }

    /* Initial values this rule cannot read as copy-on-write: a variable, a call of something else, `.init()` with its type inferred, a number. */
    @Test func initialValuesThatAreNotTheShapeAreNotFound() {
        let source = """
            let first = values.reduce(start) { $0 + [$1] }
            let second = values.reduce(makeStart()) { $0 + [$1] }
            let third = values.reduce(.init()) { (result: [Int], value: Int) in result + [value] }
            let fourth = values.reduce(0, +)
            let fifth = values.reduce(Swift.Array<Int>()) { $0 + [$1] }

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Two arguments without a trailing closure are the shape; one argument and no closure, or a labelled first argument, is not. */
    @Test func theArgumentCountIsSwiftLints() {
        let source = """
            let first = names.reduce("", +)
            let second = names.reduce("")
            let third = names.reduce(into: "", { $0 += $1 })

            """
        #expect(Self.findings(source) == ["reduceInto@1:19"])
    }

    /* A lazy sequence's `reduce` is `Sequence.reduce(_:_:)` and copies its accumulator the same way. */
    @Test func aLazySequencesReduceIsFound() {
        let source = "let first = values.lazy.reduce([]) { $0 + [$1] }\n"
        #expect(Self.findings(source) == ["reduceInto@1:25"])
    }

    @Test func aFileWithNoReduceDoesNotApply() {
        let source = "let first = values.map { $0 }\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!ReduceInto().applies(to: file))
    }

    /*
     End to end on a real package, symbols from the index the build wrote: an array's and a string's accumulators
     are found, and an option set's `[]`, a `reduce` of ours, `reduce(into:)` and a sum are not.
     */
    static let packageSource = """
        struct Flags: OptionSet {
            let rawValue: Int
        }

        struct Bag {
            func reduce(_ initial: String, _ next: (String, Int) -> String) -> String { initial }
        }

        func checks(values: [Int], flags: [Flags], bag: Bag) -> Int {
            let grown = values.reduce([]) { $0 + [$1] }
            let text = values.reduce("") { $0 + String($1) }
            let combined: Flags = flags.reduce([]) { $0.union($1) }
            let ours = bag.reduce("") { $0 + String($1) }
            let into = values.reduce(into: []) { $0.append($1) }
            let sum = values.reduce(0, +)
            return grown.count + text.count + combined.rawValue + ours.count + into.count + sum
        }

        """

    @Test func anArraysAndAStringsAccumulatorsAreFlaggedAndTheLookAlikesAreNot() async throws {
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
        let candidates = parsed.files.filter { ReduceInto().applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            ReduceInto().findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.messageId)@\($0.line)"
            }
        }
        #expect(
            found == ["reduceInto@10", "reduceInto@11"],
            "expected the array's and the string's accumulators and none of the look-alikes: \(found)",
        )
    }
}
