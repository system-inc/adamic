import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `correctness-no-default-for-owned-enum` both ways. The unit cases parse a source string and hand the rule symbols built by
 position, one occurrence at every name token the case gives a symbol for, so each case says what the compiler
 would have resolved. The symbols are the ones a probe package's index recorded. There is no SwiftLint rule to
 take examples from, so the cases are the shapes the rule reads and every miss its header names. The end-to-end
 case runs a real package, so the symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct CorrectnessNoDefaultForOwnedEnumTests {
    static let running = "s:7Control6StatusO7runningyA2CmF"
    static let failed = "s:7Control6StatusO6failedyACSScACmF"
    static let done = "s:7Control6StatusO4doneyA2CmF"
    static let status = "s:7Control6StatusO"
    static let preferred = "s:7Control6StatusO9preferredACvpZ"
    static let leaf = "s:7Control4TreeO4leafyA2C4LeafOcACmF"
    static let red = "s:7Control4TreeO4LeafO3redyA2EmF"
    static let optionalSome = "s:Sq4someyxSgxcABmlF"
    static let patternOperator = "s:s2teoiySbx_xtSQRzlF"
    static let part = "s:7Control5PieceO4partyAcA4PartOcACmF"
    static let make = "s:7Control6StatusO4makeACyFZ"
    static let fallbackName = "s:7Control12fallbackNameSSyF"
    static let quaternionInitializer = "s:So10simd_quatda2ix2iy2iz1rABSd_S3dtcfc"

    /*
     The findings, as `line:column message`, with each token whose text is a key resolved to that symbol, under
     the name `declarationNames` gives it (an initializer's `init(ix:iy:iz:r:)`) or else its own text. A token in
     `matchedByOperator` also gets the implicit `~=` the index records where a pattern is compared, not matched.
     */
    static func findings(
        _ source: String,
        symbols names: [String: String],
        matchedByOperator: Set<String> = [],
        ownedModules: Set<String> = ["Control"],
        declarationNames: [String: String] = [:],
    ) -> [String] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Control",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        var occurrences: [FileSymbols.Occurrence] = []
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            if let symbol = names[token.text] {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: symbol,
                        name: declarationNames[token.text] ?? token.text,
                        isReference: true,
                    )
                )
            }
            /* As the index records `Status.done`: the type a second time, at the element's own name. */
            if let dot = token.previousToken(viewMode: .sourceAccurate), dot.text == ".",
                let type = dot.previousToken(viewMode: .sourceAccurate), type.text.first?.isUppercase == true,
                let symbol = names[type.text]
            {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: symbol,
                        name: type.text,
                        isReference: true,
                    )
                )
            }
            if matchedByOperator.contains(token.text) {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: patternOperator,
                        name: "~=(_:_:)",
                        isReference: true,
                        isImplicit: true,
                    )
                )
            }
        }
        return CorrectnessNoDefaultForOwnedEnum().findings(
            in: file,
            symbols: FileSymbols(occurrences, ownedModules: ownedModules),
        ).map { "\($0.line):\($0.column) \($0.message)" }
    }

    static let statusSymbols = ["running": running, "failed": failed, "done": done, "preferred": preferred]
    static let statusSymbolsWithType = statusSymbols.merging(["Status": status]) { _, written in written }

    static func message(_ enumName: String) -> String {
        "This default answers for every case of \(enumName), including any added later, so the compiler can no longer say this switch does not handle a new one. List the remaining cases instead."
    }

    /* The plain shape: leading-dot elements, one with a payload bound, and a `default` at its keyword. */
    @Test func aDefaultOverAnOwnedEnumIsFound() {
        let source = """
            func describe(status: Status) -> String {
                switch status {
                case .running: "running"
                case .failed(let reason): reason
                default: "other"
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbols) == ["5:5 \(Self.message("Status"))"])
    }

    /* Every way a pattern names its element: `let` outside, a payload, a trailing `?` on an optional subject, and the type written beside a leading dot. */
    @Test func everyPatternSpellingIsRead() {
        let source = """
            func describe(status: Status, maybe: Status?) {
                switch status {
                case let .failed(reason): print(reason)
                case Status.done, .running: break
                default: break
                }
                switch maybe {
                case .running?: break
                case nil: break
                default: break
                }
            }

            """
        #expect(
            Self.findings(source, symbols: Self.statusSymbolsWithType) == [
                "5:5 \(Self.message("Status"))", "10:5 \(Self.message("Status"))",
            ]
        )
    }

    /* A `let` binding, a `where` guard and a static member of the enum are still patterns on the enum once one leading-dot element anchors it. */
    @Test func otherPatternsBesideAnElementStillCount() {
        let source = """
            func describe(status: Status, ready: Bool) {
                switch status {
                case .running where ready: break
                case .preferred: break
                case let other where other == .done: break
                default: break
                }
            }

            """
        #expect(
            Self.findings(source, symbols: Self.statusSymbols, matchedByOperator: ["preferred"]) == [
                "6:5 \(Self.message("Status"))"
            ]
        )
    }

    /* A nested enum is named with its parents, a generic enum's element ends in its signature, word substitutions spell the whole name, and a private enum's discriminator is read past. */
    @Test(arguments: [
        ("s:7Control6HolderC4ModeO4fastyA2EmF", "Holder.Mode"),
        ("s:7Control3BoxO4someyACyxGxcAEmlF", "Box"),
        ("s:7Control0A5StateO2onyA2CmF", "ControlState"),
        ("s:7Control6StatusO07runningB0yACSi_tcACmF", "Status"),
        ("s:7Control5OuterV5InnerO5alphayA2EmF", "Outer.Inner"),
        (
            "s:7Control9StagePaneC14HandProbePhase33_B31C49341E9607FD1243C82440F7E8AALLO6movingyA2FmF",
            "StagePane.HandProbePhase",
        ),
    ])
    func theEnumIsNamedFromItsElementsSymbol(element: String, enumName: String) {
        let source = """
            func describe(value: Subject) {
                switch value {
                case .element: break
                default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: ["element": element]) == ["4:5 \(Self.message(enumName))"])
    }

    /* The two `default`s of a switch nested in another are each judged by their own switch's patterns, and a `#if` clause's cases count. */
    @Test func nestedSwitchesAndConditionalCasesAreEachJudged() {
        let source = """
            func describe(status: Status, count: Int) {
                switch status {
                case .running:
                    switch count {
                    case 0: break
                    default: break
                    }
                #if DEBUG
                case .done: break
                #endif
                default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbols) == ["11:5 \(Self.message("Status"))"])
    }

    /* Patterns that name no element: values, ranges, bindings, `where` alone, and a `default` alone. */
    @Test func aSwitchWhosePatternsNameNoElementIsNotFound() {
        let source = """
            func describe(count: Int, name: String) {
                switch count {
                case 0: break
                case 1...9: break
                case let large where large > 100: break
                default: break
                }
                switch name {
                case "running": break
                default: break
                }
                switch count {
                default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbols).isEmpty)
    }

    /*
     DesignerRail's "not mine" shape: several cases to values, every other case to one constant. Each arm a
     literal, a raw value, a bare case or static member, or a constructor of constants, written alone or
     returned, so the `default` is the value a case added later would get anyway.
     */
    @Test func aProjectionToConstantsKeepsItsDefault() {
        let source = """
            extension Status {
                var category: String? {
                    switch self {
                    case .running, .done: self.rawValue
                    case .failed: "Failed"
                    default: nil
                    }
                }
                var slot: Slot? {
                    switch self {
                    case .running: .socks
                    case .done: Garments.Slot.shoes
                    default: nil
                    }
                }
                var isActive: Bool {
                    switch self {
                    case .running: return true
                    default: return false
                    }
                }
                var piece: Piece? {
                    switch self {
                    case .running: .part(.mouth)
                    case .done: .feelings
                    default: nil
                    }
                }
                var rotation: Rotator {
                    switch self {
                    case .running: Rotator(pitch: 1, yaw: -1, roll: 0.5)
                    case .done: .init(pitch: 0, yaw: 0, roll: 1)
                    default: Rotator(pitch: 0, yaw: 0, roll: 0)
                    }
                }
                var quaternion: simd_quatd {
                    switch self {
                    case .running: SIMD4<Double>(0, 0, 0, 1).quaternion
                    default: simd_quatd(ix: 0, iy: 0, iz: 0, r: 1)
                    }
                }
                var limits: [Int] {
                    switch self {
                    case .running: [1, 2]
                    case .done: []
                    default: [0]
                    }
                }
            }

            """
        let symbols = Self.statusSymbols.merging(["part": Self.part, "simd_quatd": Self.quaternionInitializer]) {
            _,
            written in written
        }
        /* `quaternion` reads a member of a constructed value, which is not a constructor, so its default is still found. */
        let initializer = ["simd_quatd": "init(ix:iy:iz:r:)"]
        #expect(
            Self.findings(source, symbols: symbols, declarationNames: initializer) == ["39:9 \(Self.message("Status"))"]
        )
        let constructed = source.replacingOccurrences(
            of: "SIMD4<Double>(0, 0, 0, 1).quaternion",
            with: "simd_quatd(ix: 1, iy: 0, iz: 0, r: 0)",
        )
        #expect(Self.findings(constructed, symbols: symbols, declarationNames: initializer).isEmpty)
        /* The same lowercase call, resolved to something that is not an initializer, is a function call. */
        #expect(Self.findings(constructed, symbols: symbols) == ["39:9 \(Self.message("Status"))"])
    }

    /* HumanoidBone's VRM 0.x name: a few cases spelled otherwise, every other case its own raw value, read bare or through the subject. */
    @Test func aRawValuePassthroughKeepsItsDefault() {
        let source = """
            extension Status {
                var legacyName: String {
                    switch self {
                    case .running: "busy"
                    case .done: "finished"
                    default: rawValue
                    }
                }
            }
            func legacyName(of status: Status) -> String {
                switch status {
                case .running: "busy"
                default: status.rawValue
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbols).isEmpty)
    }

    /*
     A switch that decides behavior lists every case, whatever its `default` says: a default calling a function, a
     constant default beside an arm that calls one, a branch, a payload read, an interpolation, two statements, a
     bare `rawValue` on a subject that is not `self`, and a member call the index resolved to a static function.
     */
    @Test func aSwitchThatDecidesBehaviorStillReportsItsDefault() {
        let source = """
            func decide(status: Status, ready: Bool, count: Int) -> String? {
                switch status {
                case .running: "busy"
                default: fallbackName()
                }
                switch status {
                case .running: fallbackName()
                default: nil
                }
                switch status {
                case .running: ready ? "busy" : "idle"
                default: nil
                }
                switch status {
                case .failed(let reason): reason
                default: nil
                }
                switch status {
                case .running: "\\(count) running"
                default: nil
                }
                switch status {
                case .running:
                    let label = "busy"
                    return label
                default: return nil
                }
                switch status {
                case .running: "busy"
                default: rawValue
                }
                switch status {
                case .running: .make()
                default: nil
                }
            }

            """
        let symbols = Self.statusSymbols.merging(["fallbackName": Self.fallbackName, "make": Self.make]) { _, written in
            written
        }
        #expect(
            Self.findings(source, symbols: symbols) == [
                "4:5 \(Self.message("Status"))", "8:5 \(Self.message("Status"))", "12:5 \(Self.message("Status"))",
                "16:5 \(Self.message("Status"))",
                "20:5 \(Self.message("Status"))", "26:5 \(Self.message("Status"))", "30:5 \(Self.message("Status"))",
                "34:5 \(Self.message("Status"))",
            ]
        )
    }

    /* A tuple subject's `default` stands for the combinations left out. */
    @Test func aTupleSubjectIsNotFound() {
        let source = """
            func describe(left: Status, right: Status) {
                switch (left, right) {
                case (.running, .running): break
                default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbols).isEmpty)
    }

    /* Elements from two enums, a payload's and an optional's, where the default may stand for either. */
    @Test func elementsFromMoreThanOneEnumAreNotFound() {
        let source = """
            func describe(tree: Tree, maybe: Status?) {
                switch tree {
                case .leaf(.red): break
                default: break
                }
                switch maybe {
                case .some(.running): break
                default: break
                }
            }

            """
        #expect(
            Self.findings(
                source,
                symbols: ["leaf": Self.leaf, "red": Self.red, "some": Self.optionalSome, "running": Self.running],
            ).isEmpty
        )
    }

    /* `@unknown default` still warns on a case not listed, so it is a different statement. */
    @Test func anUnknownDefaultIsNotFound() {
        let source = """
            func describe(status: Status) {
                switch status {
                case .running: break
                @unknown default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbols).isEmpty)
    }

    /* A dependency's or the standard library's enum may gain cases we do not own. */
    @Test func anEnumThisPackageDoesNotOwnIsNotFound() {
        let source = """
            func describe(status: Status, maybe: Int?) {
                switch status {
                case .running: break
                default: break
                }
                switch maybe {
                case .some: break
                default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbols, ownedModules: ["Vendor"]).isEmpty)
        #expect(
            Self.findings(source, symbols: ["some": Self.optionalSome], ownedModules: ["Control", "Swift"]).isEmpty
        )
    }

    /*
     A custom `~=` lets an enum's element match an `Int`, which needs its `default`. Written with its type, the
     element anchors nothing; written with a leading dot, the index records the `~=` that compared it.
     */
    @Test func aPatternComparedByAPatternOperatorIsNotFound() {
        let source = """
            func describe(number: Int) {
                switch number {
                case Status.running: break
                default: break
                }
                switch number {
                case .running: break
                default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: Self.statusSymbolsWithType, matchedByOperator: ["running"]).isEmpty)
    }

    /* Symbols this reading does not follow: an `@objc` enum's Clang name, an enum inside another module's type, a static member, a name never resolved. */
    @Test func symbolsThatDoNotReadAsAnOwnedElementAreNotFound() {
        let source = """
            func describe(value: Subject) {
                switch value {
                case .element: break
                default: break
                }
            }

            """
        #expect(Self.findings(source, symbols: ["element": "c:@M@Control@E@Bridged@BridgedOne"]).isEmpty)
        #expect(Self.findings(source, symbols: ["element": "s:10Foundation4DateV7ControlE4KindO5earlyyA2FmF"]).isEmpty)
        #expect(Self.findings(source, symbols: ["element": Self.preferred]).isEmpty)
        #expect(Self.findings(source, symbols: ["element": "s:7Control5PointV4zeroACvpZ"]).isEmpty)
        #expect(Self.findings(source, symbols: [:]).isEmpty)
    }

    @Test(arguments: [
        ("s:7Control6StatusO7runningyA2CmF", "s:7Control6StatusO", "Status"),
        ("s:7Control4TreeO4LeafO3redyA2EmF", "s:7Control4TreeO4LeafO", "Tree.Leaf"),
        ("s:7Control0A5StateO2onyA2CmF", "s:7Control0A5StateO", "ControlState"),
    ])
    func anElementsSymbolReadsAsItsEnum(element: String, enumSymbol: String, enumName: String) {
        #expect(
            CorrectnessNoDefaultForOwnedEnum.element(element)
                == CorrectnessNoDefaultForOwnedEnum.Element(enumSymbol: enumSymbol, enumName: enumName)
        )
    }

    @Test(arguments: [
        "s:7Control6StatusO", "s:7Control6StatusO9preferredACvpZ", "s:7Control6StatusO4makeACyFZ",
        "s:7Control5PointV4zeroACvpZ", "s:Sq4someyxSgxcABmlF", "s:7Control6StatusO007running", "c:@M@Control@E@Bridged",
    ])
    func otherSymbolsDoNotReadAsAnElement(symbol: String) {
        #expect(CorrectnessNoDefaultForOwnedEnum.element(symbol) == nil)
    }

    @Test func aFileWithNoDefaultDoesNotApply() {
        let source = "func describe(status: Status) {\n    switch status {\n    case .running: break\n    }\n}\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!CorrectnessNoDefaultForOwnedEnum().applies(to: file))
    }

    /*
     End to end on a real package, symbols from the index the build wrote: a `default` over our own enum is found,
     and the `default`s over an `Int` matched by a custom `~=`, over a tuple, over an optional's two enums and over
     the standard library's `Optional` are not.
     */
    static let packageSource = """
        import simd

        enum Status {
            case running
            case failed(String)
            case done
        }

        func ~= (pattern: Status, value: Int) -> Bool { false }

        func describe(status: Status, number: Int, maybe: Status?, count: Int?) -> [Int] {
            var results: [Int] = []
            switch status {
            case .running: results.append(1)
            case .failed(let reason): results.append(reason.count)
            default: results.append(0)
            }
            switch number {
            case .running: results.append(1)
            default: results.append(0)
            }
            switch (status, number) {
            case (.done, 0): results.append(1)
            default: results.append(0)
            }
            switch maybe {
            case .some(.done): results.append(1)
            default: results.append(0)
            }
            switch count {
            case .some(let value): results.append(value)
            default: results.append(0)
            }
            return results
        }

        enum Piece {
            case part(Int)
            case whole

            static func make() -> Piece { .whole }
        }

        func fallback() -> Piece? { nil }

        func projections(status: Status) -> [Piece?] {
            let piece: Piece? =
                switch status {
                case .running: .part(1)
                case .done: .whole
                default: nil
                }
            let rotation: simd_quatd =
                switch status {
                case .running: simd_quatd(ix: 1, iy: 0, iz: 0, r: 0)
                default: simd_quatd(ix: 0, iy: 0, iz: 0, r: 1)
                }
            let made: Piece? =
                switch status {
                case .running: .make()
                default: nil
                }
            let called: Piece? =
                switch status {
                case .running: .whole
                default: fallback()
                }
            _ = rotation
            return [piece, made, called]
        }

        """

    @Test func aDefaultOverOurEnumIsFlaggedAndTheLookAlikesAreNot() async throws {
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
        let candidates = parsed.files.filter { CorrectnessNoDefaultForOwnedEnum().applies(to: $0) }
        var provider = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        )
        provider.ownedModules = Pipeline.ownedModules(of: package)
        let symbols = provider.symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            CorrectnessNoDefaultForOwnedEnum().findings(
                in: file,
                symbols: symbols.symbols[file.url.path] ?? FileSymbols([]),
            ).map { "\($0.line):\($0.column)" }
        }
        #expect(
            found == ["16:5", "61:9", "66:9"],
            "expected the defaults over Status that decide, and not its projections or the look-alikes: \(found)",
        )
    }
}
