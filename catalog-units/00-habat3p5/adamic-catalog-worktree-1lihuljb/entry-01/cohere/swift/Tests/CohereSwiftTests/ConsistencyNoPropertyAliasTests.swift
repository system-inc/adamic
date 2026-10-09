import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `consistency-no-property-alias` both ways. Each case parses a source string and reads back the text of every
 finding's span, which is the binding (`tooltip = filterMode.tooltip`), or the reasons a silent line was silenced.

 The real sites come first: the two macOS findings (`KingdomSidebar`'s tooltip, `MetalTerminalPane`'s focus
 generation) and four of Presence's twelve, then the Presence and macOS lines that must stay silent and why (a clock
 read stored three times, a counter compared and stored, a take-and-clear, a write through the local after an
 `await`, a hoist before a hot loop, a value captured into a detached task). Then the TypeScript rule's own cases
 carried over where Swift has them, and the Swift-only reasons for silence.
 */
struct ConsistencyNoPropertyAliasTests {
    static func parsed(_ source: String) -> ParsedFile {
        ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
    }

    /* The text of every finding's span, after checking what every finding must be. */
    static func spans(_ source: String) -> [String] {
        let file = parsed(source)
        let rule = ConsistencyNoPropertyAlias()
        let findings = rule.findings(in: file)
        #expect(findings.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        for finding in findings {
            #expect(finding.rule == "cohere-swift/consistency-no-property-alias")
            #expect(finding.messageId == "noPropertyAlias")
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
        }
        let lines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return findings.map { finding in
            guard let endLine = finding.endLine, let endColumn = finding.endColumn, endLine == finding.line else {
                return "spans lines"
            }
            return String(decoding: lines[finding.line - 1][(finding.column - 1)..<(endColumn - 1)], as: UTF8.self)
        }
    }

    /* Why each alias-shaped binding in the source is silent, in source order; an empty list is a finding. */
    static func silences(_ source: String) -> [[ConsistencyNoPropertyAlias.Silence]] {
        ConsistencyNoPropertyAlias.candidates(in: Syntax(Parser.parse(source: source))).map(\.silences)
    }

    /* Lines inside a function of a class, with the members the cases lean on. */
    static func subject(
        _ bodyLines: String,
        signature: String = "func subject(options: Options, state: State)",
    ) -> String {
        """
        final class Subject {
            var count = 0
            var library = Library()
            var layer = Layer()
            var stream = AsyncStream<Int> { _ in }
            func save() async {}
            func use(_ value: Any) {}

            \(signature) {
        \(bodyLines)
            }
        }
        """
    }

    // MARK: The real sites, reported

    /* macOS `KingdomSidebar.swift:567`, near verbatim: a `@ViewBuilder` property passes the alias once, as an argument. */
    @Test func theSidebarTooltipIsReported() {
        let source = """
            struct KingdomSidebar: View {
                let filterMode: FilterMode

                @ViewBuilder
                private var livePanesToggle: some View {
                    let symbol = filterMode.iconSystemName
                    let tooltip = filterMode.tooltip
                    // All states render in the same `.secondary` tint.
                    IconButton(
                        systemName: symbol,
                        iconColor: .secondary,
                        iconSize: 12,
                        tooltip: tooltip,
                        action: cycleFilterMode
                    )
                }
            }
            """
        #expect(Self.spans(source) == ["tooltip = filterMode.tooltip"])
    }

    /* macOS `MetalTerminalPane.swift:312`: a static counter bumped, then read once into the call that carries it. */
    @Test func theFocusGenerationIsReported() {
        let source = """
            func updateNSView(_ nsView: PaneView) {
                PaneFocusArbiter.generation += 1
                let generation = PaneFocusArbiter.generation
                reassertFocus(on: nsView.terminalView, generation: generation, attemptsRemaining: 6)
            }
            """
        #expect(Self.spans(source) == ["generation = PaneFocusArbiter.generation"])
    }

    /* Presence `Announcement.swift:109`: a struct copied off a window, read once, with only a property read and arithmetic before it. */
    @Test func theAnnouncementSizeIsReported() {
        let source = """
            func place(_ panel: NSWindow, on screen: NSScreen) {
                let size = panel.frame.size
                let visible = screen.visibleFrame
                let origin = NSPoint(x: (visible.midX - size.width / 2).rounded(), y: visible.minY + Self.liftFromBottom)
                if panel.frame.origin != origin { panel.setFrameOrigin(origin) }
            }
            """
        #expect(Self.spans(source) == ["size = panel.frame.size"])
    }

    /* Presence `StagePane+MotionMenu.swift:425`: read once, as an `if` condition, through `self.`. */
    @Test func theMotionMenuMotionsAreReported() {
        let source = """
            func buildMenu(_ menu: NSMenu) {
                let motions = self.motionLibrary.motions
                if motions.isEmpty {
                    let empty = NSMenuItem(title: "No motions found", action: nil, keyEquivalent: "")
                    menu.addItem(empty)
                }
            }
            """
        #expect(Self.spans(source) == ["motions = self.motionLibrary.motions"])
    }

    /* Presence `StagePane+Things.swift:302`: declared inside a loop and read once in the same iteration, which is not a hoist. */
    @Test func anAliasInsideALoopReadInTheSameIterationIsReported() {
        let source = """
            func stepThings() {
                for var falling in self.things.falling {
                    let entity = falling.thing.entity
                    let position = entity.position
                    if simd_distance(position, falling.still) > StageThings.stillWithin {
                        falling.still = position
                        falling.stillSince = now
                    }
                }
            }
            """
        #expect(Self.spans(source) == ["entity = falling.thing.entity"])
    }

    /* Presence `StudioSearchProbe.swift:163`: read once inside a string interpolation in the call that logs it. */
    @Test func aReadInsideAnInterpolationIsReported() {
        let source = """
            func report(_ name: String) {
                let isShown = StudioSearchPopover.shared.isShown
                presenceLog("studio search probe: \\(name): \\(isShown ? "open" : "closed")")
                StudioSearchPopover.shared.close()
            }
            """
        #expect(Self.spans(source) == ["isShown = StudioSearchPopover.shared.isShown"])
    }

    // MARK: The real sites, silent

    /* macOS `ProviderProxyRequestTiming.swift:91`: one instant, stored three times. Three reaches would be three instants. */
    @Test func aClockReadStoredThreeTimesIsSilent() {
        let source = """
            func markBodyByte() {
                let now = clock.now
                if firstBodyByteAt == nil { firstBodyByteAt = now }
                if let previous = previousByteAt {
                    let gap = now - previous
                    if gap > maximumInterChunkGap { maximumInterChunkGap = gap }
                }
                previousByteAt = now
                lastBodyByteAt = now
            }
            """
        #expect(Self.spans(source).isEmpty)
        #expect(Self.silences(source) == [[.repeated]])
    }

    /* Presence `EffectRenderer.swift:314`: a locked counter compared and then stored. Two reaches could disagree. */
    @Test func aCounterComparedAndStoredIsSilent() {
        let source = """
            func draw() {
                let generation = EffectLibrary.shared.generation
                if generation != self.libraryGeneration {
                    self.libraryGeneration = generation
                }
            }
            """
        #expect(Self.silences(source) == [[.repeated]])
    }

    /* macOS `SessionDetail.swift:321`: a true alias, read three times with nothing between, silent as the price of the clock above. */
    @Test func aTrueAliasReadThreeTimesIsAKnownMiss() {
        let source = """
            func memberDisplayLabel(for position: KingdomPosition) -> String {
                let username = position.username
                let display = position.displayName
                if !display.isEmpty && display != username {
                    return "\\(display) (@\\(username))"
                }
                return "@\\(username)"
            }
            """
        #expect(Self.silences(source) == [[.repeated]])
    }

    /* Presence `ScreenRecorder.swift:232`: take and clear, then a shorthand `guard let` that reads the local. */
    @Test func aTakeAndClearIsSilent() {
        let source = Self.subject(
            """
                    let conform = self.conform
                    self.conform = nil
                    guard let conform else { return }
                    use(conform)
            """
        )
        #expect(Self.silences(source) == [[.runs]])
    }

    /* The same shorthand with nothing between: the reach is `guard let conform = self.conform`. */
    @Test func aShorthandReadWithNothingBetweenIsReported() {
        let source = Self.subject(
            """
                    let conform = self.conform
                    guard let conform else { return }
                    use(conform)
            """
        )
        #expect(Self.spans(source) == ["conform = self.conform"])
    }

    /* Presence `BodyProperties.swift:220`: written through after an `await` in the same statement. Kirk's lost-update snapshot. */
    @Test func aWriteThroughTheLocalAfterAnAwaitIsSilent() {
        let source = """
            func remove(_ choice: Choice) {
                Task { @MainActor in
                    let library = self.library
                    library.removal = await BodyRemoval.perform(choice)
                }
            }
            """
        #expect(Self.spans(source).isEmpty)
        #expect(Self.silences(source).first?.contains(.suspension) == true)
    }

    /* The suspension alone: an awaited property read, with no call and no assignment to stand in for it. */
    @Test func aReadAfterAnAwaitedPropertyIsSilent() {
        let source = Self.subject(
            """
                    let count = self.count
                    let total = await state.total
                    use(count)
            """,
            signature: "func subject(options: Options, state: Store) async",
        )
        #expect(Self.silences(source) == [[.suspension]])
    }

    /* The suspension alone, written to the local's object after the await in the same statement. */
    @Test func aWriteThroughAfterAnAwaitedPropertyIsSilent() {
        let source = Self.subject(
            """
                    let library = self.library
                    library.removal = await state.removal
            """,
            signature: "func subject(options: Options, state: Store) async",
        )
        #expect(Self.silences(source) == [[.suspension]])
    }

    /* A `for await` loop between: each step suspends. */
    @Test func aReadAfterAForAwaitIsSilent() {
        let source = Self.subject(
            """
                    let count = self.count
                    for await _ in stream {}
                    use(count)
            """,
            signature: "func subject(options: Options, state: State) async",
        )
        #expect(Self.silences(source).first?.contains(.suspension) == true)
    }

    /* Presence EarDSP `Texture.swift:131`, near verbatim: a class's array hoisted before a hot loop, on purpose. */
    @Test func theEarDspHoistIsSilent() {
        let source = """
            final class Texture {
                var magnitudes: [Float] = []
                private func frame(from start: Int) {
                    // Read through a local for the frame: each read of a class's array property is an
                    // exclusivity check.
                    let magnitudes = self.magnitudes
                    let scale = 1 / Double(window)
                    for bin in bins {
                        let magnitude = Double(magnitudes[bin]) * scale
                        summed[bin] += magnitude
                    }
                }
            }
            """
        #expect(Self.spans(source).isEmpty)
        #expect(Self.silences(source).first?.contains(.hoisted) == true)
    }

    /* The hoist alone, with nothing else between the declaration and the loop. */
    @Test func aHoistBeforeALoopIsSilent() {
        let source = Self.subject(
            """
                    let values = state.values
                    for index in 0..<4 {
                        total += values[index]
                    }
            """
        )
        #expect(Self.silences(source) == [[.hoisted]])
    }

    /* A read in a loop's sequence runs once, so it is not a hoist; the body is. */
    @Test func aReadInALoopsSequenceIsNotAHoist() {
        let source = Self.subject(
            """
                    let values = state.values
                    for value in values {
                        total += value
                    }
            """
        )
        #expect(Self.spans(source) == ["values = state.values"])
    }

    /* Presence `RigTakeReview.swift:210`: a value captured into a detached task, so `self` stays out of it. */
    @Test func aValueCapturedIntoADetachedTaskIsSilent() {
        let source = """
            func open(_ take: Take) {
                let folder = take.folder
                Task.detached(priority: .userInitiated) {
                    let loaded = Self.load(folder)
                    await MainActor.run { self.finishOpening(take, loaded) }
                }
            }
            """
        #expect(Self.silences(source).first == [.captured])
    }

    /* A capture list naming the local, a `defer`, a nested function and an `async let`: each reads it on another schedule. */
    @Test func everyOtherScheduleIsCaptured() {
        #expect(
            Self.silences(Self.subject("        let count = state.count\n        run { [count] in print(count) }"))
                .first?.contains(.captured) == true
        )
        #expect(
            Self.silences(Self.subject("        let count = state.count\n        defer { use(count) }")).first == [
                .captured
            ]
        )
        #expect(
            Self.silences(Self.subject("        let count = state.count\n        func nested() -> Int { count }")).first
                == [.captured]
        )
        #expect(
            Self.silences(
                Self.subject(
                    "        let count = state.count\n        async let doubled = count * 2",
                    signature: "func subject(state: State) async",
                )
            ).first == [.captured]
        )
    }

    // MARK: What runs between

    /* A call before the read, an argument before it, an assignment, an `inout`, a subscript and a macro: any of them could change the property. */
    @Test func codeThatRunsBeforeTheReadSilencesIt() {
        for between in [
            "refresh()", "_ = 1", "total += 1", "swap(&a, &b)", "let first = items[0]", "let file = #fileID",
        ] {
            let source = Self.subject("        let count = state.count\n        \(between)\n        use(count)")
            #expect(Self.silences(source) == [[.runs]], "\(between)")
        }
        #expect(Self.silences(Self.subject("        let count = state.count\n        use(make(), count)")) == [[.runs]])
    }

    /* A call that takes the read as its argument runs after it, and a receiver is read before its arguments. */
    @Test func callsThatContainTheReadDoNotSilenceIt() {
        #expect(
            Self.spans(Self.subject("        let count = state.count\n        use(String(count).count)")) == [
                "count = state.count"
            ]
        )
        #expect(
            Self.spans(Self.subject("        let library = self.library\n        library.reload(make())")) == [
                "library = self.library"
            ]
        )
    }

    /* Written through: the write lands after its right side runs, so a call there counts. */
    @Test func aWriteThroughTheLocalAfterACallIsSilent() {
        #expect(
            Self.silences(Self.subject("        let layer = self.layer\n        layer.contents = make()")) == [[.runs]]
        )
        #expect(
            Self.silences(Self.subject("        let layer = self.layer\n        layer?.contents = make()")) == [[.runs]]
        )
        #expect(
            Self.spans(Self.subject("        let layer = self.layer\n        layer.contentsFormat = .rgba")) == [
                "layer = self.layer"
            ]
        )
        #expect(
            Self.spans(Self.subject("        let layer = self.layer\n        use(layer.contents == make())")) == [
                "layer = self.layer"
            ]
        )
    }

    /* Property reads, operators and interpolation between are taken as safe, as TypeScript takes them. */
    @Test func readsAndOperatorsBetweenAreSafe() {
        let source = Self.subject(
            "        let count = state.count\n        let other = state.other + 1\n        let label = \"\\(other)\"\n        use(count)"
        )
        #expect(Self.spans(source) == ["count = state.count"])
    }

    /* The chain's first name declared again before the read: the reach would name another object. */
    @Test func aRedeclaredRootIsSilent() {
        let source = Self.subject(
            "        let count = state.count\n        let state = options.state\n        use(count)"
        )
        #expect(Self.silences(source).first == [.shadowed])
    }

    /* A `#if` that declares the name again hides which declaration a later read means. */
    @Test func aReadTheScopeReadingCannotPlaceIsSilent() {
        let source = Self.subject(
            "        let count = state.count\n        #if DEBUG\n        let count = 1\n        #endif\n        use(count)"
        )
        #expect(Self.silences(source).first?.contains(.unreadable) == true)
        #expect(Self.spans(source).isEmpty)
    }

    /* Never read: the compiler already says so. */
    @Test func anUnreadAliasIsSilent() {
        #expect(Self.silences(Self.subject("        let count = state.count")) == [[.unread]])
    }

    // MARK: TypeScript's cases, carried over

    /* TypeScript's real `Run.ts` and `TrackedPromise` shapes, a nested reach, a long one, a closure body and `self.`. */
    @Test func theTypeScriptFiringShapesAreReported() {
        #expect(
            Self.spans(Self.subject("        let onOutputLine = options.onOutputLine\n        return onOutputLine"))
                == ["onOutputLine = options.onOutputLine"]
        )
        #expect(
            Self.spans(Self.subject("        let promise = tracked.promise\n        return promise")) == [
                "promise = tracked.promise"
            ]
        )
        #expect(
            Self.spans(Self.subject("        let value = state.inner.value\n        return value")) == [
                "value = state.inner.value"
            ]
        )
        #expect(
            Self.spans(Self.subject("        let value = a.one.two.three.value\n        return value")) == [
                "value = a.one.two.three.value"
            ]
        )
        #expect(
            Self.spans("let run = { (options: Options) in\n    let timeout = options.timeout\n    return timeout\n}\n")
                == ["timeout = options.timeout"]
        )
        #expect(
            Self.spans(Self.subject("        let count = self.count\n        return count")) == ["count = self.count"]
        )
        #expect(
            Self.spans(Self.subject("        let value = (state).value\n        return value")) == [
                "value = (state).value"
            ]
        )
        #expect(
            Self.spans("var total: Int {\n    let count = state.count\n    return count\n}\n") == [
                "count = state.count"
            ]
        )
    }

    /* Several bindings in one `let`: each is judged, and a later binding's call runs before a read below it. */
    @Test func eachBindingOfOneLetIsJudged() {
        #expect(
            Self.spans(Self.subject("        let host = url.host, port = url.port\n        use(host)")) == [
                "host = url.host"
            ]
        )
        #expect(
            Self.silences(Self.subject("        let host = url.host, port = make()\n        use(host)")) == [[.runs]]
        )
    }

    /* The TypeScript clean cases Swift has: a different name, calls anywhere in the chain, a subscript, an optional at any depth. */
    @Test func theTypeScriptCleanShapesAreSilent() {
        for initializer in [
            "options.other", "Intl.DateTimeFormat().resolvedOptions().value", "make().value", "a.make().inner.value",
            "make()[key].value",
            "(make()).value", "record[name]", "options?.value", "account.data?.profile.value",
            "o.locales[index]?.value",
            "self.plainTerminalView.getTerminal().value",
        ] {
            #expect(
                Self.spans(Self.subject("        let value = \(initializer)\n        use(value)")).isEmpty,
                "\(initializer)",
            )
        }
    }

    /* File scope and a type's own members are out of scope, and so is a `let` with no value or a tuple pattern. */
    @Test func outsideAFunctionIsSilent() {
        #expect(Self.spans("let shared = Store.shared\nprint(shared)\n").isEmpty)
        #expect(Self.spans("if flag {\n    let shared = Store.shared\n    print(shared)\n}\n").isEmpty)
        #expect(Self.spans("struct Holder {\n    let shared = Store.shared\n}\n").isEmpty)
        #expect(Self.spans("func run() {\n    struct Local {\n        let shared = Store.shared\n    }\n}\n").isEmpty)
        #expect(Self.spans(Self.subject("        let value: Int\n        value = 1\n        use(value)")).isEmpty)
        #expect(
            Self.spans(Self.subject("        let (count, other) = (state.count, state.other)\n        use(count)"))
                .isEmpty
        )
    }

    // MARK: Swift's own shapes

    /* `if let`, `guard let` and `while let` narrow; `var` is a copy made to change; an annotation may convert; a force unwrap and a cast are not plain reads. */
    @Test func swiftsOwnSilentShapes() {
        for source in [
            "        if let count = state.count { use(count) }",
            "        guard let count = state.count else { return }\n        use(count)",
            "        while let count = state.count { use(count) }",
            "        var count = state.count\n        count += 1\n        use(count)",
            "        var count = state.count\n        use(count)",
            "        let count: Int = state.count\n        use(count)",
            "        let count = state!.count\n        use(count)",
            "        let count = (state as State).count\n        use(count)",
            "        let count = .count\n        use(count)",
            "        let count = try state.count\n        use(count)",
            "        lazy var count = state.count\n        use(count)",
        ] {
            #expect(Self.spans(Self.subject(source)).isEmpty, "\(source)")
        }
    }

    /* A function's name is not a property read. */
    @Test func aMethodReferenceWithArgumentNamesIsSilent() {
        #expect(Self.spans(Self.subject("        let run = state.run(_:)\n        use(run)")).isEmpty)
    }

    /* The message names the local, the reach and the repair. */
    @Test func theMessageNamesTheRepair() {
        let findings = ConsistencyNoPropertyAlias().findings(
            in: Self.parsed(Self.subject("        let count = self.count\n        use(count)"))
        )
        #expect(findings.count == 1)
        #expect(
            findings.first?.message
                == "`count` is a pure alias for `self.count`. It is read once, and nothing runs between its declaration and that read. Write `self.count` where `count` is read and delete the declaration, so the value keeps the object it came from and a naked local still means this scope made it."
        )
    }
}
