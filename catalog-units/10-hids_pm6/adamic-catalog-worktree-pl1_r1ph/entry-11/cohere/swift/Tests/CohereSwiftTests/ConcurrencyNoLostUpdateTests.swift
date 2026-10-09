import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `concurrency-no-lost-update` both ways. Each case parses a source string and reads back the text of every finding's
 span, which is the whole assignment. The real sites come first: AhraOS Presence's `StudioModel` seeding loop, whose
 save-and-restore must stay silent, and the same loop with the restore turned into an update, which must not; the
 awaited writes beside it in Presence's `WardrobePanel` and `MotionSpotlightView`; and the accumulator in macOS's
 `ProviderProxy`. Then the shapes measured with swiftc (the reason a compound assignment is silent in Swift), the
 TypeScript rule's cases carried over where Swift has them, and the Swift-only reasons for silence (a `mutating`
 method, an `inout` parameter, a captured local).
 */
struct ConcurrencyNoLostUpdateTests {
    /* The text of every finding's span, after checking what every finding must be. */
    static func spans(_ source: String) -> [String] {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let rule = ConcurrencyNoLostUpdate()
        let findings = rule.findings(in: file)
        #expect(findings.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        for finding in findings {
            #expect(finding.rule == "cohere-swift/concurrency-no-lost-update")
            #expect(finding.messageId == "lostUpdate")
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
        }
        let lines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return findings.map { finding in
            guard let endLine = finding.endLine, let endColumn = finding.endColumn else { return "no end" }
            var bytes: [UInt8] = []
            for line in finding.line...endLine {
                let start = line == finding.line ? finding.column - 1 : 0
                let end = line == endLine ? endColumn - 1 : lines[line - 1].count
                if line != finding.line {
                    bytes.append(UInt8(ascii: "\n"))
                }
                bytes += lines[line - 1][start..<end]
            }
            return String(decoding: bytes, as: UTF8.self)
        }
    }

    /* Body lines inside an async method of an actor, with the state and the calls every case leans on. */
    static func subject(_ bodyLines: String, modifiers: String = "") -> String {
        """
        struct Tally {
            var count = 0
            var extra = 0
            func adding(_ value: Int) -> Tally { self }
        }

        final class Entry {
            var retries = 0
            var position = 0
        }

        actor Store {
            var count = 0
            var total = 0
            var other = 0
            var items: [Int] = []
            var table: [String: Int] = [:]
            var state = Tally()
            var cache: [Int]?
            var lastRunAt = 0.0
            let flag = true
            let xs: [Int] = []
            let stream = AsyncStream<Int> { _ in }

            func next() async -> Int { 1 }
            func price(_ value: Int = 0) async -> Int { 1 }
            func save(_ value: Int = 0) async {}
            func run() async throws {}
            func check() throws {}
            func checked() throws -> Int { 0 }
            func use(_ value: Int) {}

            \(modifiers)func subject(entry: Entry) async throws {
        \(bodyLines)
            }
        }

        """
    }

    // MARK: The real sites

    /*
     Presence's `StudioModel.swift:702` to `:723`, near verbatim: the appearance in front is kept, each preset is worn
     and photographed with an await, and the kept appearance is put back. The variations move the read into the
     `Task`, drop the loop's replacement, and change what is written back.
     */
    static func studioSeeding(keptInsideTask: Bool = false, replaces: Bool = true, restore: String = "kept") -> String {
        let kept = "let kept = self.appearance"
        return """
            @MainActor
            final class StudioModel {
                var appearance = 0
                var undoStack: [Int] = []
                func save() async -> URL? { nil }

                func seedCharacters(presets: [Int]) {
                    \(keptInsideTask ? "" : kept)
                    Task { @MainActor in
                        \(keptInsideTask ? kept : "")
                        for preset in presets {
                            \(replaces ? "self.appearance = preset" : "")
                            await self.save()
                        }
                        self.appearance = \(restore)
                        self.undoStack.removeAll()
                    }
                }
            }
            """
    }

    /* As it stands: a verbatim restore, after this function's own replacement, of a value read outside the `Task` that writes it. Silent three times over. */
    @Test func theStudioSeedingRestoreIsSilent() {
        #expect(Self.spans(Self.studioSeeding()).isEmpty)
    }

    /* With the kept value read inside the `Task`: still a verbatim restore after a replacement. */
    @Test func theStudioRestoreReadInsideTheTaskIsSilent() {
        #expect(Self.spans(Self.studioSeeding(keptInsideTask: true)).isEmpty)
        #expect(Self.spans(Self.studioSeeding(keptInsideTask: true, replaces: false)).isEmpty)
    }

    /* A derived restore is still a restore when the loop replaced the appearance before it awaited. */
    @Test func aDerivedRestoreAfterTheLoopsReplacementIsSilent() {
        #expect(Self.spans(Self.studioSeeding(keptInsideTask: true, restore: "kept + 1")).isEmpty)
    }

    /* The seeding loop with no replacement in it: the appearance it writes back was read before every photograph. */
    @Test func theStudioLoopWithoutItsReplacementIsAnUpdate() {
        #expect(
            Self.spans(Self.studioSeeding(keptInsideTask: true, replaces: false, restore: "kept + presets.count")) == [
                "self.appearance = kept + presets.count"
            ]
        )
    }

    /*
     Presence's `PresenceApplicationDelegate+GazeProbe.swift:203` to `:209`, near verbatim: a static saved, replaced,
     awaited across, and put back. The variation drops the replacement and writes back a value computed from the
     saved one, which is an update.
     */
    static func gazeProbe(replaces: Bool = true, restore: String = "usedMediaPipe") -> String {
        """
        extension PresenceApplicationDelegate {
            func runGazeProbe(seconds: Double, source: RecordedFrames) {
                Task { @MainActor in
                    let usedMediaPipe = LiveMirror.usesMediaPipe
                    \(replaces ? "LiveMirror.usesMediaPipe = true" : "")
                    await self.pane?.startLiveMirror(from: source)
                    await Task.sleepUnlessCancelled(for: .seconds(seconds))
                    let mirror = LiveMirror.source?.statistics
                    self.pane?.stopLiveMirror()
                    LiveMirror.usesMediaPipe = \(restore)
                    presenceLog("GAZE FOOTAGE \\(mirror)")
                }
            }
        }
        """
    }

    @Test func theGazeProbesRestoreIsSilent() {
        #expect(Self.spans(Self.gazeProbe()).isEmpty)
        #expect(Self.spans(Self.gazeProbe(restore: "!usedMediaPipe")).isEmpty)
        #expect(Self.spans(Self.gazeProbe(replaces: false)).isEmpty)
    }

    @Test func theGazeProbeWithoutItsReplacementIsAnUpdate() {
        #expect(
            Self.spans(Self.gazeProbe(replaces: false, restore: "!usedMediaPipe")) == [
                "LiveMirror.usesMediaPipe = !usedMediaPipe"
            ]
        )
    }

    /*
     Presence's `PresetColorProbe.swift:41` to `:68`, reduced: a parameter's object captured by a `Task`, its appearance
     kept, each preset worn and photographed, and the kept one put back. The variation is the update.
     */
    static func presetColorProbe(replaces: Bool = true, restore: String = "kept") -> String {
        """
        enum PresetColorProbe {
            static func runIfAsked(model: StudioModel) {
                Task { @MainActor in
                    await Task.sleepUnlessCancelled(for: .seconds(5))
                    guard let viewport = model.viewport else { return }
                    let kept = model.appearance
                    for preset in [Appearance()] + Randomiser.presets() {
                        \(replaces ? "model.appearance = preset" : "")
                        await Task.sleepUnlessCancelled(for: .seconds(2.5))
                        if let tiff = await viewport.snapshot()?.tiffRepresentation {
                            presenceLog("presence: PRESET COLORS photographed \\(tiff.count)")
                        }
                    }
                    model.appearance = \(restore)
                }
            }
        }
        """
    }

    @Test func thePresetColorProbesRestoreIsSilent() {
        #expect(Self.spans(Self.presetColorProbe()).isEmpty)
        #expect(Self.spans(Self.presetColorProbe(restore: "kept.tinted()")).isEmpty)
    }

    @Test func thePresetColorProbeWithoutItsReplacementIsAnUpdate() {
        #expect(
            Self.spans(Self.presetColorProbe(replaces: false, restore: "kept.tinted()")) == [
                "model.appearance = kept.tinted()"
            ]
        )
    }

    /* A method called on the object itself is not a replacement of one of its members, as in TypeScript. */
    @Test func aMethodOnTheObjectItselfIsNotAReplacement() {
        let source = Self.subject(
            """
                    let old = entry.retries
                    entry.reset()
                    await save()
                    entry.retries = old + 1
            """
        )
        #expect(Self.spans(source) == ["entry.retries = old + 1"])
    }

    /* Presence's `WardrobePanel.swift:81`: a value read from an await, not computed from what was there. */
    @Test func theWardrobesAwaitedPositionsAreSilent() {
        let source = """
            struct WardrobePanel: View {
                @State private var positions: [String] = []
                @State private var isReading = true
                var body: some View {
                    List(self.positions, id: \\.self) { Text($0) }
                        .task {
                            self.positions = await Kingdom.positions()
                            self.isReading = false
                        }
                }
            }
            """
        #expect(Self.spans(source).isEmpty)
    }

    /* Presence's `MotionSpotlightView.swift:649`, and the same write handed the old clip: the callee's answer either way. */
    @Test func theSpotlightsAwaitedClipIsSilent() {
        let source = """
            struct MotionSpotlightRow: View {
                let skeletons: SkeletonLibrary
                @State private var clip: SkeletonClip?
                var body: some View {
                    Image(systemName: "figure.dance")
                        .task(id: self.id) {
                            self.clip = await self.skeletons.clip(self.id, readingFrom: self.fileUrl)
                            self.clip = await self.skeletons.refine(self.clip)
                        }
                }
            }
            """
        #expect(Self.spans(source).isEmpty)
    }

    /* macOS's `ProviderProxy.swift:653`: a local accumulator, compound besides. */
    @Test func theProxysCancelledCounterIsSilent() {
        let source = """
            actor ProviderProxy {
                private func cancelAllInFlight() async -> Int {
                    var cancelled = 0
                    for entry in currentLanes() {
                        cancelled += await entry.lane.cancelAll()
                    }
                    return cancelled
                }
            }
            """
        #expect(Self.spans(source).isEmpty)
    }

    // MARK: Measured with swiftc

    /* `total = total + (await pause())` lost the other task's bump, implicit `self` and explicit alike. */
    @Test func aSumReadBeforeItsAwaitIsAnUpdate() {
        #expect(Self.spans(Self.subject("count = count + (await next())")) == ["count = count + (await next())"])
        #expect(
            Self.spans(Self.subject("self.total = self.total + (await price())")) == [
                "self.total = self.total + (await price())"
            ]
        )
    }

    /* `+=` reads its left side when the operator runs, after the await: the bump was kept, on an actor, a subscript, a `@MainActor` class and a global. */
    @Test func aCompoundAssignmentIsSilent() {
        #expect(Self.spans(Self.subject("self.total += await price()")).isEmpty)
        #expect(Self.spans(Self.subject("count -= await next()")).isEmpty)
        #expect(Self.spans(Self.subject(#"self.table["k", default: 0] += await next()"#)).isEmpty)
        let global = """
            nonisolated(unsafe) var global = 0
            func globalCompound() async { global += await pause() }
            """
        #expect(Self.spans(global).isEmpty)
    }

    /* `max(total, await pause())`, `items + [await pause()]` and `items.appending(await pause())` all lost it: a receiver and earlier arguments are read first. */
    @Test func readsBeforeAnAwaitInACallAreAnUpdate() {
        #expect(
            Self.spans(Self.subject("self.total = max(self.total, await price())")) == [
                "self.total = max(self.total, await price())"
            ]
        )
        #expect(
            Self.spans(Self.subject("self.items = self.items + [await next()]")) == [
                "self.items = self.items + [await next()]"
            ]
        )
        #expect(
            Self.spans(Self.subject("self.items = self.items.appending(await next())")) == [
                "self.items = self.items.appending(await next())"
            ]
        )
    }

    /* `items.append(await pause())` kept it: a mutating call is not an assignment, and its receiver is accessed after the argument. */
    @Test func aMutatingCallIsSilent() {
        #expect(Self.spans(Self.subject("self.items.append(await next())")).isEmpty)
    }

    /* `table["k"] = (table["k"] ?? 0) + (await pause())` lost it: a literal key is a member. */
    @Test func aLiteralKeyIsAMember() {
        let line = #"self.table["k"] = (self.table["k"] ?? 0) + (await next())"#
        #expect(Self.spans(Self.subject(line)) == [line])
        #expect(Self.spans(Self.subject(#"self.table["k"] = (self.table["other"] ?? 0) + (await next())"#)).isEmpty)
    }

    /* `total = await total + pause()` lost it too, and is not reported: where the suspension falls inside an awaited operand is not in the syntax. */
    @Test func aReadInsideAnAwaitedOperandIsAKnownMiss() {
        #expect(Self.spans(Self.subject("self.total = await self.total + price()")).isEmpty)
    }

    // MARK: Reporting

    @Test func aLetReadBeforeTheAwaitAndWrittenAfterIsAnUpdate() {
        let source = Self.subject(
            """
                    let old = self.count
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source) == ["self.count = old + 1"])
    }

    /* The message names the target as written, and the repair. */
    @Test func theMessageNamesTheTarget() {
        let source = Self.subject(
            """
                    let old = count
                    await save()
                    count = old + 1
            """
        )
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(
            ConcurrencyNoLostUpdate().findings(in: file).map(\.message) == [
                "This write loses updates. Its new value is computed from `count` as it was before an await, and the function is suspended in between, so anything else that changes `count` during the suspension (another call on this actor, another task on the main actor) is overwritten by this line. Read `count` after the suspension, or keep the whole read, compute and write on one side of it."
            ]
        )
    }

    /* `count` and `self.count` are one property, whichever spelling reads and whichever writes. */
    @Test func implicitAndExplicitSelfAreOnePath() {
        let source = Self.subject(
            """
                    let old = count
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source) == ["self.count = old + 1"])
    }

    /* A member of a parameter that is not `inout`: writing through it compiles only for an object the caller holds. */
    @Test func aParametersObjectIsShared() {
        #expect(
            Self.spans(Self.subject("entry.retries = entry.retries + (await next())")) == [
                "entry.retries = entry.retries + (await next())"
            ]
        )
    }

    @Test func aSuspensionOnOneBranchIsEnough() {
        let source = Self.subject(
            """
                    let old = self.count
                    if flag {
                        await save()
                    }
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source) == ["self.count = old + 1"])
    }

    @Test func aForAwaitLoopSuspends() {
        let before = Self.subject(
            """
                    let old = self.count
                    for await value in stream {
                        use(value)
                    }
                    self.count = old + 1
            """
        )
        #expect(Self.spans(before) == ["self.count = old + 1"])
        let inside = Self.subject(
            """
                    let old = self.count
                    for await value in stream {
                        self.count = old + value
                        break
                    }
            """
        )
        #expect(Self.spans(inside) == ["self.count = old + value"])
    }

    @Test func theReadPrecedesAForkThatSuspends() {
        #expect(
            Self.spans(Self.subject("self.count = self.count + (flag ? await next() : 1)")) == [
                "self.count = self.count + (flag ? await next() : 1)"
            ]
        )
        #expect(
            Self.spans(Self.subject("self.count = self.count + ((await next()) > 0 ? 1 : 2)")) == [
                "self.count = self.count + ((await next()) > 0 ? 1 : 2)"
            ]
        )
    }

    @Test func aMethodCalledOnTheStaleValue() {
        let source = Self.subject(
            """
                    let items = self.items
                    await save()
                    self.items = items.filter { $0 > 0 }
            """
        )
        #expect(Self.spans(source) == ["self.items = items.filter { $0 > 0 }"])
    }

    @Test func aChainOfLets() {
        let source = Self.subject(
            """
                    let old = self.count
                    let following = old + 1
                    await save()
                    self.count = following
            """
        )
        #expect(Self.spans(source) == ["self.count = following"])
    }

    /* Through a `do` whose `try await` completes, and into the `catch` of one that suspends and then throws. */
    @Test func throughADoThatSuspends() {
        let after = Self.subject(
            """
                    let old = self.count
                    do {
                        try await run()
                    } catch {
                        use(0)
                    }
                    self.count = old + 1
            """
        )
        #expect(Self.spans(after) == ["self.count = old + 1"])
        let inCatch = Self.subject(
            """
                    let old = self.count
                    do {
                        try await run()
                    } catch {
                        self.count = old + 1
                    }
            """
        )
        #expect(Self.spans(inCatch) == ["self.count = old + 1"])
    }

    @Test func aWriteToASiblingOrInsideTheTargetIsNotAReplacement() {
        let sibling = Self.subject(
            """
                    let old = self.count
                    self.other = 0
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(sibling) == ["self.count = old + 1"])
        let inside = Self.subject(
            """
                    let old = self.state
                    self.state.extra = 1
                    await save()
                    self.state = old.adding(2)
            """
        )
        #expect(Self.spans(inside) == ["self.state = old.adding(2)"])
    }

    @Test func aLetUsedAfterASuspensionInsideTheExpression() {
        let source = Self.subject(
            """
                    let old = self.count
                    self.count = (await next()) + old
            """
        )
        #expect(Self.spans(source) == ["self.count = (await next()) + old"])
    }

    /* The loop's back edge carries the read around: the second pass writes a value read before the first await. */
    @Test func aLoopWritingALetFromBeforeIt() {
        let source = Self.subject(
            """
                    let old = self.count
                    for value in xs {
                        self.count = old + value
                        await save()
                    }
            """
        )
        #expect(Self.spans(source) == ["self.count = old + value"])
    }

    @Test func aConditionalOnTheStaleValue() {
        let source = Self.subject(
            """
                    let old = self.count
                    await save()
                    self.count = (old + 1) > 0 ? old + 1 : 0
            """
        )
        #expect(Self.spans(source) == ["self.count = (old + 1) > 0 ? old + 1 : 0"])
    }

    /* A `Task` is judged on its own awaits, with `[weak self]` and `guard let self` reading the same object. */
    @Test func aTaskWithWeakSelf() {
        let source = """
            @MainActor
            final class Model {
                var count = 0
                func save() async {}
                func start() {
                    Task { [weak self] in
                        guard let self else { return }
                        let old = self.count
                        await self.save()
                        self.count = old + 1
                    }
                    Task { [weak self] in
                        self?.count = (self?.count ?? 0) + (await Model.remote())
                    }
                }
                static func remote() async -> Int { 1 }
            }
            """
        #expect(
            Self.spans(source) == ["self.count = old + 1", "self?.count = (self?.count ?? 0) + (await Model.remote())"]
        )
    }

    /* A global, and a static member through its type, are shared from a free function. */
    @Test func aGlobalAndAStaticAreShared() {
        let source = """
            nonisolated(unsafe) var global = 0
            enum Counter { nonisolated(unsafe) static var total = 0 }
            func pause() async -> Int { 1 }
            func bump() async {
                global = global + (await pause())
                Counter.total = Counter.total + (await pause())
            }
            """
        #expect(
            Self.spans(source) == [
                "global = global + (await pause())", "Counter.total = Counter.total + (await pause())",
            ]
        )
    }

    /* A `defer` runs at the end of its scope, after the await. */
    @Test func aDeferWritesAfterTheAwait() {
        let source = Self.subject(
            """
                    let old = self.count
                    defer { self.count = old + 1 }
                    await save()
            """
        )
        #expect(Self.spans(source) == ["self.count = old + 1"])
    }

    @Test func theTargetReadAsASubscriptKey() {
        #expect(
            Self.spans(Self.subject("self.count = xs[self.count] + (await next())")) == [
                "self.count = xs[self.count] + (await next())"
            ]
        )
    }

    /* A function named like the property is not a read of it. */
    @Test func aCalleeSpelledLikeTheTargetIsNotARead() {
        let source = """
            actor Counter {
                var count = 0
                func count(_ value: Int) -> Int { value }
                func next() async -> Int { 1 }
                func bump() async { count = count(await next()) }
            }
            """
        #expect(Self.spans(source).isEmpty)
    }

    // MARK: Silent

    @Test func aTimestampAfterAnAwaitIsSilent() {
        #expect(
            Self.spans(Self.subject("use(await next())\n        self.lastRunAt = Date().timeIntervalSince1970")).isEmpty
        )
    }

    /* Written back exactly as saved, cast or not. */
    @Test func aVerbatimRestoreIsSilent() {
        let source = Self.subject(
            """
                    let saved = self.count
                    try await run()
                    self.count = saved
            """
        )
        #expect(Self.spans(source).isEmpty)
        let cast = Self.subject(
            """
                    let saved = self.count
                    try await run()
                    self.count = saved as Int
            """
        )
        #expect(Self.spans(cast).isEmpty)
    }

    @Test func aReplacementBeforeTheSuspensionMakesTheWriteARestore() {
        let direct = Self.subject(
            """
                    let old = self.count
                    self.count = 0
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(direct).isEmpty)
        let compound = Self.subject(
            """
                    let old = self.count
                    self.count += 1
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(compound).isEmpty)
        let container = Self.subject(
            """
                    let old = self.state.count
                    self.state = Tally()
                    await save()
                    self.state.count = old + 1
            """
        )
        #expect(Self.spans(container).isEmpty)
        let mutated = Self.subject(
            """
                    let old = self.items
                    self.items.removeAll()
                    await save()
                    self.items = old + [1]
            """
        )
        #expect(Self.spans(mutated).isEmpty)
    }

    @Test func aReplacementOnOnePathOnlyLeavesTheOtherPathAnUpdate() {
        let source = Self.subject(
            """
                    let old = self.count
                    if flag {
                        self.count = 0
                    }
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source) == ["self.count = old + 1"])
    }

    @Test func theReadAfterTheSuspensionIsSilent() {
        #expect(Self.spans(Self.subject("await save()\n        self.count = self.count + 1")).isEmpty)
        #expect(Self.spans(Self.subject("self.count = (await next()) + self.count")).isEmpty)
    }

    /* A local is never judged: nobody else can see it, and a captured one cannot be written concurrently in Swift 6. */
    @Test func aLocalIsSilent() {
        #expect(
            Self.spans(Self.subject("var total = 0\n        total = total + (await next())\n        use(total)"))
                .isEmpty
        )
        let shadowing = Self.subject("var count = 0\n        count = count + (await next())\n        use(count)")
        #expect(Self.spans(shadowing).isEmpty)
        let captured = Self.subject(
            """
                    var total = 0
                    await withTaskGroup(of: Int.self) { group in
                        total = total + (await group.next() ?? 0)
                    }
                    use(total)
            """
        )
        #expect(Self.spans(captured).isEmpty)
        let localObject = Self.subject("let local = Entry()\n        local.retries = local.retries + (await next())")
        #expect(Self.spans(localObject).isEmpty)
    }

    @Test func theReadAndTheSuspensionOnDifferentBranchesAreSilent() {
        #expect(Self.spans(Self.subject("self.count = flag ? self.count + 1 : await next()")).isEmpty)
        #expect(Self.spans(Self.subject("self.cache = self.cache ?? [await next()]")).isEmpty)
        #expect(Self.spans(Self.subject("self.count = flag ? 1 : (await next())")).isEmpty)
    }

    @Test func theSuspendingBranchLeaves() {
        let returns = Self.subject(
            """
                    let old = self.count
                    if flag {
                        await save()
                        return
                    }
                    self.count = old + 1
            """
        )
        #expect(Self.spans(returns).isEmpty)
        let fails = Self.subject(
            """
                    let old = self.count
                    guard flag else {
                        await save()
                        fatalError("unreachable")
                    }
                    self.count = old + 1
            """
        )
        #expect(Self.spans(fails).isEmpty)
        let breaks = Self.subject(
            """
                    let old = self.count
                    outer: for value in xs {
                        for _ in xs {
                            await save()
                            break outer
                        }
                        self.count = old + value
                    }
            """
        )
        #expect(Self.spans(breaks).isEmpty)
        let switches = Self.subject(
            """
                    let old = self.count
                    switch flag {
                    case true:
                        await save()
                        return
                    case false:
                        break
                    }
                    self.count = old + 1
            """
        )
        #expect(Self.spans(switches).isEmpty)
    }

    /* A throw before the await reaches the `catch` without suspending. */
    @Test func aThrowBeforeTheAwaitIsSilent() {
        let source = Self.subject(
            """
                    let old = self.count
                    do {
                        try check()
                        await save()
                    } catch {
                        self.count = old + 1
                    }
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    /* A `try` covering more than one call may throw from the first, before the await beside it. */
    @Test func aThrowBeforeAnAwaitInTheSameTryIsSilent() {
        let source = Self.subject(
            """
                    let old = self.count
                    do {
                        use(try checked() + (await next()))
                    } catch {
                        self.count = old + 1
                    }
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    @Test func aLetDeclaredAgainOnEveryPassIsSilent() {
        let source = Self.subject(
            """
                    for value in xs {
                        let old = self.count
                        self.count = old + value
                        await save(value)
                    }
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    /* In a `mutating` method or an initializer `self` is held for the whole call, and an `inout` parameter likewise. */
    @Test func exclusiveAccessIsSilent() {
        let mutating = """
            struct Counter {
                var count = 0
                func next() async -> Int { 1 }
                mutating func bump() async {
                    count = count + (await next())
                    await withTaskGroup(of: Int.self) { group in
                        self.count = self.count + (await group.next() ?? 0)
                    }
                }
                init() async {
                    self.count = self.count + (await Counter.remote())
                }
                static func remote() async -> Int { 1 }
            }
            """
        #expect(Self.spans(mutating).isEmpty)
        let parameter = """
            func bump(entry: inout Entry, next: () async -> Int) async {
                entry.retries = entry.retries + (await next())
            }
            """
        #expect(Self.spans(parameter).isEmpty)
        let untyped = Self.subject(
            "let bump: (Entry) async -> Void = { entry in entry.retries = entry.retries + (await self.next()) }"
        )
        #expect(Self.spans(untyped).isEmpty)
    }

    @Test func aLetOfADifferentPropertyIsSilent() {
        let source = Self.subject(
            """
                    let old = self.other
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    /* A value copy of a path holding the target reads as an alias, as in TypeScript: a known miss for a struct. */
    @Test func aCopyOfTheContainingPathIsAKnownMiss() {
        let source = Self.subject(
            """
                    let snapshot = self.state
                    await save()
                    self.state.count = snapshot.count + 1
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    @Test func anAwaitedCallsResultIsTheCalleesAnswer() {
        let response = Self.subject(
            """
                    let response = await price(self.count)
                    self.count = response + 1
            """
        )
        #expect(Self.spans(response).isEmpty)
        let kept = Self.subject(
            """
                    let fetched = await price(self.count)
                    await save()
                    self.count = fetched + 1
            """
        )
        #expect(Self.spans(kept).isEmpty)
        #expect(Self.spans(Self.subject("self.count = await price(self.count + 1)")).isEmpty)
        #expect(Self.spans(Self.subject("self.count = try await run().hashValue")).isEmpty)
    }

    /* A closure reads the property when it runs, not when it is made, and an updater runs on the far side of the await. */
    @Test func aClosureReadsWhenItRuns() {
        #expect(Self.spans(Self.subject("self.count = await apply { self.count + 1 }")).isEmpty)
        #expect(Self.spans(Self.subject("self.items = wrap({ self.items.count + 1 }, await next())")).isEmpty)
    }

    @Test func aClosuresAwaitIsNotThisFunctions() {
        let source = Self.subject(
            """
                    let old = self.count
                    Task { await self.save() }
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    /* A write inside a closure from a read outside it is a known miss: the closure is its own function. */
    @Test func aWriteInsideAClosureIsAKnownMiss() {
        let source = Self.subject(
            """
                    let old = self.count
                    await save()
                    Task { self.count = old + 1 }
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    @Test func aLetStaleOnlyOnTheBranchThatDoesNotCarryIt() {
        let source = Self.subject(
            """
                    let old = self.count
                    let chosen = flag ? old : await next()
                    self.count = chosen + 1
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    @Test func aVariableIsNotACarrier() {
        let source = Self.subject(
            """
                    var old = self.count
                    old = 0
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    @Test func aPropertyOfSelfIsNotThatPropertyOfAParameter() {
        #expect(Self.spans(Self.subject("self.count = entry.retries + (await next())")).isEmpty)
    }

    /* `#if` hides which branch is compiled, and so whether the function replaced the target, so the function is not read. */
    @Test func aFunctionHoldingConditionalCompilationIsSilent() {
        let source = Self.subject(
            """
                    let old = self.count
                    #if DEBUG
                    self.count = 0
                    #endif
                    await save()
                    self.count = old + 1
            """
        )
        #expect(Self.spans(source).isEmpty)
    }

    @Test func aFunctionThatCannotSuspendIsSilent() {
        let source = """
            final class Model {
                var count = 0
                func bump() { let old = count; count = old + 1 }
            }
            """
        #expect(Self.spans(source).isEmpty)
    }
}
