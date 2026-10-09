import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 `correctness-no-identical-branches` both ways. Each case returns its findings as the source text each one covers, so a
 finding on the wrong `if` of a chain fails rather than passing on its count. The one real site on either proving
 ground comes first, macOS's `OscParser` chain, before and after its repair. The rest are modeled: ahra's Kling request
 builder, where the TypeScript rule was born, written as Swift; Presence's `dressBackdrop` chain, `BodyCatalog` decoder, `BodiesInspector` builder chain, `DanceMode` duration
 and `MotionSpotlightView` border, each with one branch copied over another and as they stand; and the proving
 grounds' `?? fallback : fallback` ternaries, which look like this shape to a regular expression and are not. Then the
 original's own cases, each one ported, then the places Swift needs its own exactness.
 */
struct CorrectnessNoIdenticalBranchesTests {
    /* Every finding as the text of its span, for a file the rule agrees to read, as the pipeline runs it. */
    static func findings(_ source: String) -> [String] {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let rule = CorrectnessNoIdenticalBranches()
        let found = rule.findings(in: file)
        #expect(found.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        let utf8 = Array(source.utf8)
        var lineStarts = [0]
        for (offset, byte) in utf8.enumerated() where byte == UInt8(ascii: "\n") {
            lineStarts.append(offset + 1)
        }
        return found.map { finding in
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
            #expect(finding.messageId == "identicalBranches")
            #expect(finding.rule == "cohere-swift/correctness-no-identical-branches")
            guard let endLine = finding.endLine, let endColumn = finding.endColumn else { return "no end" }
            let start = lineStarts[finding.line - 1] + finding.column - 1
            let end = lineStarts[endLine - 1] + endColumn - 1
            return String(decoding: utf8[start..<end], as: UTF8.self)
        }
    }

    // MARK: The real sites

    /*
     `modules/kling/KlingApi.ts:156` and `:187` in ahra, the TypeScript rule's first finding, written as the Swift a
     request builder would be: `highQuality` picked `"pro"` either way, and the fix sends `"std"` on the false side.
     The other ternaries beside it differ and stay silent.
     */
    static func klingApi(falseMode: String) -> String {
        """
        func buildPayload(uploadedUrl: String?, modelVersion: String, highQuality: Bool, hasTailImage: Bool, duration: String, prompt: String) -> [String: Any] {
            let is3 = modelVersion == "3.0"
            if let uploadedUrl {
                let modelType = is3 ? "m2v_aio2video" : highQuality ? "m2v_img2video_hq" : "m2v_img2video"
                let effectiveDuration = is3 ? (hasTailImage ? "5" : "7") : duration
                return [
                    "type": modelType,
                    "image": uploadedUrl,
                    "arguments": [
                        ["name": "prompt", "value": prompt],
                        ["name": "duration", "value": effectiveDuration],
                        ["name": "model_mode", "value": highQuality ? "pro" : \(falseMode)],
                    ],
                ]
            }
            else {
                let modelType = is3 ? "m2v_aio2video" : highQuality ? "m2v_txt2video_hq" : "m2v_txt2video"
                return [
                    "type": modelType,
                    "arguments": [
                        ["name": "prompt", "value": prompt],
                        ["name": "duration", "value": is3 ? "7" : duration],
                        ["name": "model_mode", "value": highQuality ? "pro" : \(falseMode)],
                    ],
                ]
            }
        }
        """
    }

    @Test func klingApiFiresOnBothBuildersBeforeTheFix() {
        #expect(
            Self.findings(Self.klingApi(falseMode: "\"pro\"")) == [
                #"highQuality ? "pro" : "pro""#, #"highQuality ? "pro" : "pro""#,
            ]
        )
    }

    @Test func klingApiIsSilentAfterTheFix() {
        #expect(Self.findings(Self.klingApi(falseMode: "\"std\"")).isEmpty)
    }

    /*
     Presence `Studio/StudioViewport.swift:1196`, `dressBackdrop()`, on 2026-10-03: a three-arm chain in a `Void`
     method. As written every arm differs. With the last arm copied from the middle one, only `isDressedForEffects`
     is ignored, so the finding starts at its `if`.
     */
    static func dressBackdrop(lastArm: String) -> String {
        """
        private func dressBackdrop() {
            let color = self.backdrop.color
            if self.isDressedForEffects, self.backdrop == .showcase, let showcase = self.showcase {
                self.rendering?.renderer.cameraSettings.colorBackground = .color(
                    CGColor(red: 0, green: 0, blue: 0, alpha: 0))
                showcase.element = self.showcaseElement
                self.effects?.backing = .painted(showcase)
            }
            else if self.isDressedForEffects {
                self.rendering?.renderer.cameraSettings.colorBackground = .color(
                    CGColor(red: 0, green: 0, blue: 0, alpha: 0))
                self.effects?.backing = .color(self.backdropLight())
            }
            else {
        \(lastArm)
            }
        }
        """
    }

    @Test func dressBackdropFiresWhenTheLastArmIsACopy() {
        let copied = """
                    // Copied from the arm above and never edited.
                    self.rendering?.renderer.cameraSettings.colorBackground =
                        .color(CGColor(red: 0, green: 0, blue: 0, alpha: 0))
                    self.effects?.backing = .color(self.backdropLight())
            """
        let expected = """
            if self.isDressedForEffects {
                    self.rendering?.renderer.cameraSettings.colorBackground = .color(
                        CGColor(red: 0, green: 0, blue: 0, alpha: 0))
                    self.effects?.backing = .color(self.backdropLight())
                }
                else {
                    // Copied from the arm above and never edited.
                    self.rendering?.renderer.cameraSettings.colorBackground =
                        .color(CGColor(red: 0, green: 0, blue: 0, alpha: 0))
                    self.effects?.backing = .color(self.backdropLight())
                }
            """
        #expect(Self.findings(Self.dressBackdrop(lastArm: copied)) == [expected])
    }

    @Test func dressBackdropIsSilentAsWritten() {
        let written = """
                    // A developed frame is her alone on clear, the colour laid under by the development.
                    self.rendering?.renderer.cameraSettings.colorBackground =
                        self.rendering?.develops == true
                        ? .color(CGColor(red: 0, green: 0, blue: 0, alpha: 0)) : .color(color.cgColor)
                    self.effects?.backing = nil
            """
        #expect(Self.findings(Self.dressBackdrop(lastArm: written)).isEmpty)
    }

    /*
     Presence `Studio/Bodies/BodyCatalog.swift:97`, a decoder's chain of `if let` attempts, in an initializer. Its
     arms use the names they bind, so the bindings keep the chain silent. With the `[String]` arm's body copied from
     the fallback, the branch no longer reads `list`, the binding cannot reach it, and the attempt chooses nothing.
     */
    static func bodyCatalog(listArm: String) -> String {
        """
        init(from decoder: any Decoder) throws {
            let container = try decoder.singleValueContainer()
            if let flag = try? container.decode(Bool.self) {
                self = .flag(flag)
            }
            else if let number = try? container.decode(Int.self) {
                self = .number(number)
            }
            else if let list = try? container.decode([String].self) {
                \(listArm)
            }
            else {
                self = .text(try container.decode(String.self))
            }
        }
        """
    }

    @Test func bodyCatalogFiresWhenABoundArmIsACopyOfTheFallback() {
        #expect(
            Self.findings(Self.bodyCatalog(listArm: "self = .text(try container.decode(String.self))")) == [
                """
                if let list = try? container.decode([String].self) {
                        self = .text(try container.decode(String.self))
                    }
                    else {
                        self = .text(try container.decode(String.self))
                    }
                """
            ]
        )
    }

    @Test func bodyCatalogIsSilentAsWritten() {
        #expect(Self.findings(Self.bodyCatalog(listArm: "self = .list(list)")).isEmpty)
    }

    /*
     Presence `Studio/Bodies/BodiesInspector.swift:37`, a `@ViewBuilder` chain choosing what the inspector shows.
     With the last two arms the same, a builder still gives each arm its own identity, so flipping between them
     rebuilds the view: the chain is left alone. The very same arms in a `Void` method choose nothing, and fire.
     */
    static func bodiesInspector(declaration: String) -> String {
        """
        \(declaration) {
            if let folder = self.library.selectedFolder {
                FolderDetail(library: self.library, folder: folder)
                    .id(folder.id)
            }
            else if self.library.selection == BodiesNode.bundledId {
                Explainer(
                    symbol: "shippingbox",
                    title: "Built In",
                    text: "These ship inside the app."
                )
            }
            else {
                Explainer(
                    symbol: "shippingbox",
                    title: "Built In",
                    text: "These ship inside the app."
                )
            }
        }
        """
    }

    @Test func bodiesInspectorChainUnderAViewBuilderIsSilent() {
        #expect(
            Self.findings(Self.bodiesInspector(declaration: "@ViewBuilder\nprivate var selected: some View")).isEmpty
        )
        #expect(Self.findings(Self.bodiesInspector(declaration: "var body: some View")).isEmpty)
    }

    @Test func bodiesInspectorChainInAVoidMethodFires() {
        #expect(Self.findings(Self.bodiesInspector(declaration: "func showSelection()")).count == 1)
    }

    /*
     macOS `Sources/AhraOsServices/OscParser.swift:402` on 2026-10-03, the one finding on either proving ground: the
     last two arms of the OSC parameter chain both `reset()`, each with its own comment, so the terminator test
     chooses nothing. Merged into one `else` that keeps both reasons, the chain is silent.
     */
    static func oscParameter(tail: String) -> String {
        """
        mutating func feed(_ byte: UInt8) {
            switch state {
            case .oscParameter:
                if byte >= 0x30 && byte <= 0x39 { // digit
                    parameterValue = parameterValue * 10 + Int(byte - 0x30)
                }
                else if byte == 0x3b { // ';' → parameter done
                    payloadBytes.removeAll(keepingCapacity: true)
                    state = .oscString
                }
        \(tail)
            default:
                break
            }
        }
        """
    }

    @Test func oscParameterChainFiresOnItsTwoResets() {
        let tail = """
                    else if byte == 0x07 || byte == 0x18 || byte == 0x1a {
                        // Terminator/abort before any ';', malformed OSC, drop it.
                        reset()
                    }
                    else {
                        // Non-numeric parameter char we don't model, abandon.
                        reset()
                    }
            """
        #expect(
            Self.findings(Self.oscParameter(tail: tail)) == [
                """
                if byte == 0x07 || byte == 0x18 || byte == 0x1a {
                            // Terminator/abort before any ';', malformed OSC, drop it.
                            reset()
                        }
                        else {
                            // Non-numeric parameter char we don't model, abandon.
                            reset()
                        }
                """
            ]
        )
    }

    @Test func oscParameterChainIsSilentMerged() {
        let tail = """
                    else {
                        // A terminator or abort before any ';' is malformed, and a non-numeric char is not modeled: abandon either way.
                        reset()
                    }
            """
        #expect(Self.findings(Self.oscParameter(tail: tail)).isEmpty)
    }

    /* Presence `Body/DanceMode.swift:98`, a getter that is one ternary: as written it differs; copied, the probe flag does nothing. */
    @Test func danceModeJoinDuration() {
        #expect(Self.findings("var joinDuration: TimeInterval { Self.joinProbe ? 0 : 0.55 }").isEmpty)
        #expect(
            Self.findings("var joinDuration: TimeInterval { Self.joinProbe ? 0.55 : 0.55 }") == [
                "Self.joinProbe ? 0.55 : 0.55"
            ]
        )
    }

    /* Presence `Stage/MotionSpotlightView.swift:352`: a ternary is one value of one type, so it is judged inside a builder too. */
    @Test func motionSpotlightBorderInsideAViewBuilder() {
        let source = """
            var body: some View {
                VStack {
                    Text("Motions")
                }
                .overlay(
                    RoundedRectangle(cornerRadius: 14, style: .continuous)
                        .strokeBorder(Color.primary.opacity(self.isPopover ? 0 : 0.12), lineWidth: 1)
                )
            }
            """
        #expect(Self.findings(source).isEmpty)
        #expect(
            Self.findings(source.replacingOccurrences(of: "? 0 :", with: "? 0.12 :")) == [
                "self.isPopover ? 0.12 : 0.12"
            ]
        )
    }

    /*
     The survey's fourteen lookalikes in macOS and Presence: a fallback after `??` sits right before the `:`, so a
     regular expression sees `x : x`. The fold groups `??` above the ternary, as the compiler does, and none matches.
     */
    @Test func nilCoalescingFallbacksAreNotThisShape() {
        let source = """
            let step = arguments.count > 3 ? Int(arguments[3]) ?? 2 : 2
            let stars = marked ? ratings[body.id] ?? 0 : 0
            let offset = moving?.id == clip.id ? moving?.offset ?? .zero : .zero
            let origin = trimming?.id == clip.id ? trimming?.origin ?? clip : clip
            let range = isFirstPerson ? firstPersonIndexRange ?? indexRange : indexRange
            let local = joint < slots.count ? slots[joint].map { pose[$0] } ?? identity : identity
            let padded = key.count >= 36 ? key : key + String(repeating: " ", count: 36 - key.count)
            let dance = wanted == "first" ? dances.first : dances.first { $0.id == wanted }
            let fill = beat == 0 ? Color.white : Color.white.opacity(0.7)
            """
        #expect(Self.findings(source).isEmpty)
    }

    // MARK: The original's cases, ported

    @Test func ifAndElseDoingTheSameThing() {
        #expect(
            Self.findings(
                "func run(ready: Bool) {\n    if ready {\n        start()\n    }\n    else {\n        start()\n    }\n}"
            ) == [
                "if ready {\n        start()\n    }\n    else {\n        start()\n    }"
            ]
        )
    }

    /* Structure, not text: line breaks, spacing and comments inside the branches do not count. */
    @Test func branchesDifferingOnlyInWhitespaceAndComments() {
        let source = """
            func run(ready: Bool, count: Int) {
                if ready {
                    // keep it
                    save("count", count + 1)
                }
                else {
                    save( "count" ,
                        count + 1 /* again */ )
                }
            }
            """
        #expect(Self.findings(source).count == 1)
    }

    @Test func branchesThatOnlyExit() {
        let source = """
            func run(ready: Bool, items: [Int]) throws {
                if ready { return } else { return }
                for item in items {
                    if item > 1 { continue } else { continue }
                    if item > 2 { break } else { break }
                    if item > 3 { throw Failure.bad } else { throw Failure.bad }
                }
            }
            """
        #expect(
            Self.findings(source) == [
                "if ready { return } else { return }",
                "if item > 1 { continue } else { continue }",
                "if item > 2 { break } else { break }",
                "if item > 3 { throw Failure.bad } else { throw Failure.bad }",
            ]
        )
    }

    /* Every branch of the chain is the same, so the finding covers the whole chain once. */
    @Test func anElseIfChainWhoseEveryBranchIsTheSame() {
        #expect(
            Self.findings(
                "func run(a: Bool, b: Bool) {\n    if a { start() }\n    else if b { start() }\n    else { start() }\n}"
            ) == [
                "if a { start() }\n    else if b { start() }\n    else { start() }"
            ]
        )
    }

    /* Only `b` is ignored, so the finding starts at its `if`. */
    @Test func anElseIfChainWhoseLastConditionIsIgnored() {
        #expect(
            Self.findings(
                "func run(a: Bool, b: Bool) {\n    if a { stop() }\n    else if b { start() }\n    else { start() }\n}"
            ) == [
                "if b { start() }\n    else { start() }"
            ]
        )
    }

    @Test func ternaryChains() {
        #expect(Self.findings(#"let mode = a ? "pro" : b ? "pro" : "pro""#) == [#"a ? "pro" : b ? "pro" : "pro""#])
        #expect(Self.findings(#"let mode = a ? "std" : (b ? "pro" : "pro")"#) == [#"b ? "pro" : "pro""#])
        #expect(
            Self.findings(#"let mode = a ? ("pro") : (b ? "pro" : "pro")"#) == [#"a ? ("pro") : (b ? "pro" : "pro")"#]
        )
    }

    @Test func parenthesesAroundOneTernaryBranch() {
        #expect(Self.findings("let next = a ? (count + 1) : count + 1") == ["a ? (count + 1) : count + 1"])
    }

    @Test func emptyCollectionsAndLiteralsWrittenTwoWays() {
        let source = """
            let list = ready ? [] : [ ]
            let entry = ready
                ? ["name": "prompt", "value": prompt]
                : [
                    "name": "prompt",
                    "value": prompt
                ]
            """
        #expect(Self.findings(source).count == 2)
    }

    /* The rare `if` whose first branch holds the same `if` as its `else`: `a` is ignored, and the whole statement is the finding. */
    @Test func anIfWhoseElseIsTheSameIfAsItsBranch() {
        #expect(
            Self.findings(
                "func run(a: Bool, c: Bool) {\n    if a { if c { p() } else { q() } } else if c { p() } else { q() }\n}"
            ) == [
                "if a { if c { p() } else { q() } } else if c { p() } else { q() }"
            ]
        )
    }

    /* Each condition still chooses between `start` and `stop`: this is `if a || b` written long. */
    @Test func twoEqualBranchesFollowedByADifferentOne() {
        #expect(
            Self.findings(
                "func run(a: Bool, b: Bool) {\n    if a { start() }\n    else if b { start() }\n    else { stop() }\n}"
            ).isEmpty
        )
    }

    @Test func chainsWithNoFinalElseAndUnequalNeighbours() {
        let source = """
            func run(a: Bool, b: Bool, c: Bool) {
                if a { start() }
                else if b { start() }
                if a { start() }
                else if b { stop() }
                else if c { start() }
                if a { start() }
                start()
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Empty branches are another rule's, and commented ones carry the only difference an empty branch can carry. */
    @Test func emptyBranches() {
        let source = """
            func run(ready: Bool) {
                if ready {} else {}
                if ready {
                    // handled by the caller
                }
                else {
                    // nothing to do
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func tokensThatDiffer() {
        let source = """
            let sign = ready ? -1 : +1
            let name = ready ? item.name : item?.name
            let pattern = ready ? /a b/ : /a  b/
            let text = ready ? "a /* b */ c" : "a  c"
            let quoted = ready ? "pro" : #"pro"#
            let list = ready ? [prompt] : [prompt,]
            func run(ready: Bool, count: Int) {
                if ready { let value = 1; use(value) } else { var value = 1; use(value) }
                if ready { save("count", count + 1) } else { save("count", count - 1) }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* `switch` is declined by decision, as the original declines it. */
    @Test func switchWhoseEveryCaseIsTheSame() {
        let source = """
            func run(kind: String) {
                switch kind {
                case "a": start()
                default: start()
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    // MARK: Where Swift needs its own exactness

    /* Swift narrows nothing through a Boolean: `value` is the same `Any` in both branches, so the check is dead. */
    @Test func aTypeCheckConditionDoesNotNarrow() {
        #expect(
            Self.findings("func run(value: Any) {\n    if value is String { show(value) } else { show(value) }\n}")
                .count == 1
        )
    }

    /* `if let name` gives the first branch a new `name` of another type, so the same call can mean two things. */
    @Test func aBindingTheFirstBranchReadsKeepsItSilent() {
        let source = """
            func run(name: String?, state: Result<Int, Error>) {
                if let name { show(name) } else { show(name) }
                if let `default` = fallback { show(`default`) } else { show(`default`) }
                if case .success(let value) = state { show(value) } else { show(value) }
                if case let .success(value) = state { show(value) } else { show(value) }
                if let self { reload(self) } else { reload(self) }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A binding the branches never read cannot reach them, and a pattern that binds nothing changes no name. */
    @Test func aBindingTheBranchesDoNotReadIsJudged() {
        let source = """
            func run(cache: [String: Int], key: String, state: Phase) {
                if let cached = cache[key] { reset() } else { reset() }
                if case .ready = state { reset() } else { reset() }
            }
            """
        #expect(Self.findings(source).count == 2)
    }

    /* The widening stops at an `if let` whose branch reads its binding: there `value` is another declaration. */
    @Test func aBindingStopsTheWidening() {
        #expect(
            Self.findings(
                "func run(value: Int?, b: Bool) {\n    if let value { use(value) } else if b { use(value) } else { use(value) }\n}"
            ) == [
                "if b { use(value) } else { use(value) }"
            ]
        )
    }

    /* A bound name beside a Boolean clause still keeps it silent, and so does a pattern's spelling of a name it does not bind: a miss written down. */
    @Test func aPatternNameTheBranchSpellsIsLeftAloneEvenWhenNotBound() {
        let source = """
            func run(cache: [String: Int], key: String, state: Phase, expected: Int) {
                if let cached = cache[key], cached > 0 { record(cached) } else { record(cached) }
                if case .loaded(expected) = state { show(expected) } else { show(expected) }
                if case .loaded = state { log(.loaded) } else { log(.loaded) }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func availabilityConditionsAreLeftAlone() {
        let source = """
            func run(view: NSView) {
                if #available(macOS 15, *) { view.refresh() } else { view.refresh() }
                if #unavailable(macOS 15) { view.refresh() } else { view.refresh() }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* `#line` is where it is written, so the two copies are two values; any macro may read its position. */
    @Test func branchesExpandingAMacroAreLeftAlone() {
        let source = """
            func run(verbose: Bool) {
                if verbose { log(#line) } else { log(#line) }
                let column = verbose ? #column : #column
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A default argument that captures the caller's line is not a difference the condition chooses. */
    @Test func copiedCrashesStillChooseNothing() {
        #expect(
            Self.findings(
                "func run(strict: Bool) {\n    if strict { fatalError(\"unreachable\") } else { fatalError(\"unreachable\") }\n}"
            ).count == 1
        )
    }

    /* Swift reads a line break: `load` then `(next)` is two statements, where `load(next)` is one call, with the same tokens. */
    @Test func theSameTokensInADifferentTree() {
        let source = """
            func run(ready: Bool) {
                if ready {
                    load
                    (next)
                }
                else {
                    load(next)
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A multi-line string keeps its indentation past the closing delimiter in its token, so these two values differ. */
    @Test func multiLineStringsDifferingOnlyInIndentation() {
        let different = "let text = ready ? \"\"\"\n    a\n    \"\"\" : \"\"\"\n    a\n  \"\"\""
        #expect(Self.findings(different).isEmpty)
        let same = "let text = ready ? \"\"\"\n    a\n    \"\"\" : \"\"\"\n      a\n      \"\"\""
        #expect(Self.findings(same).count == 1)
    }

    @Test func aTrailingSemicolonIsADifferentToken() {
        #expect(Self.findings("func run(ready: Bool) {\n    if ready { start(); } else { start() }\n}").isEmpty)
    }

    /* A closure with no `return` may be a builder's: SwiftUI gives each arm its own identity. */
    @Test func statementsAResultBuilderMayTransformAreLeftAlone() {
        let source = """
            var body: some View {
                VStack {
                    if self.isEditing { Editor(text: self.text) } else { Editor(text: self.text) }
                }
            }
            @ViewBuilder func row() -> some View {
                if compact { Label("Row") } else { Label("Row") }
            }
            func mode(highQuality: Bool) -> String {
                if highQuality { "pro" } else { "pro" }
            }
            let handler = { (ready: Bool) in
                if ready { start() } else { start() }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A `return`, a `Void` result, an initializer or a setter turns the transform off or never had one, and an `if` expression is a value. */
    @Test func statementsNoBuilderCanTransformAreJudged() {
        let source = """
            let handler = { (ready: Bool) -> Int in
                if ready { start() } else { start() }
                return 1
            }
            func mode(highQuality: Bool) -> String {
                if highQuality { log() } else { log() }
                return "pro"
            }
            func modes() -> Void {
                if highQuality { log() } else { log() }
            }
            var title: String {
                get { if compact { log() } else { log() }; return "a" }
                set { if compact { log() } else { log() } }
            }
            let task = Task {
                let mode = if highQuality { "pro" } else { "pro" }
            }
            if ready { start() } else { start() }
            """
        #expect(Self.findings(source).count == 7)
    }

    /* A ternary among operators the standard table does not know is grouped at a guess, so it is skipped. */
    @Test func aTernaryBesideAnUnknownOperatorIsLeftAlone() {
        #expect(Self.findings("let order = a <=> b ? left : left").isEmpty)
        #expect(Self.findings("let order = (a <=> b) ? left : left") == ["(a <=> b) ? left : left"])
    }

    /* A `#if` is a compile-time choice, never read as a conditional. */
    @Test func compilationConditionsAreNotRead() {
        let source = """
            #if os(macOS)
            let spacing = 8
            #else
            let spacing = 8
            #endif
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func theMessageNamesTheRepair() {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: "let a = b ? 1 : 1",
            tree: Parser.parse(source: "let a = b ? 1 : 1"),
            nodeCount: 0,
        )
        let messages = CorrectnessNoIdenticalBranches().findings(in: file).map(\.message)
        #expect(
            messages == [
                "Every branch of this conditional does the same thing, so its condition chooses nothing: the code runs the same way whether it is true or false. Usually one branch was meant to differ and a copy was never edited (highQuality ? \"pro\" : \"pro\"). Write the branch that was meant, or, if both really are the same, drop the conditional and keep one copy."
            ]
        )
    }
}
