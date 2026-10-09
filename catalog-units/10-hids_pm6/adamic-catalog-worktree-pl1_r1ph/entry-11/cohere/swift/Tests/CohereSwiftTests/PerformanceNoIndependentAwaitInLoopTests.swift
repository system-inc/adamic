import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `performance-no-independent-await-in-loop` both ways. Each case parses a source string and reads back the text of
 every finding's span, which is the loop header. The real sites come first: the two loops in AhraOS's macOS app
 the survey found, before and after their repair, the accumulator beside them that must stay silent, and the
 loops in Presence that await one at a time on purpose. Then every other shape the rule flags and every shape it
 leaves alone, the TypeScript rule's cases carried over where Swift has them, and the Swift-only reasons for
 silence (actor isolation, shared state, an uncaught `try`).
 */
struct PerformanceNoIndependentAwaitInLoopTests {
    /* The words a finding carries, pinned here so a generator that drops or mangles them fails. */
    static let message =
        "This loop awaits once per item, so the items run one after another, yet nothing carries from one iteration to the next: each await is the item's own work, no local of the function is reassigned and no shared state is written, nothing leaves the loop on an awaited result or a thrown error, and nothing prints, logs or sleeps between iterations. Start them together with `withTaskGroup` (collect each result with its index when the order matters, since a group hands results back as they finish), with `async let` when the items are a fixed few, or with a bounded group when the collection can be large or the far side rate-limits."

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
        let rule = PerformanceNoIndependentAwaitInLoop()
        let findings = rule.findings(in: file)
        #expect(findings.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        for finding in findings {
            #expect(finding.rule == "cohere-swift/performance-no-independent-await-in-loop")
            #expect(finding.messageId == "independentAwaitInLoop")
            #expect(finding.message == Self.message)
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

    /* Body lines inside an async method of an actor, with the declarations every case leans on. */
    static func subject(_ bodyLines: String) -> String {
        """
        actor Lane {
            let name: String
            var queueDepth: Int { 0 }
            init(name: String) { self.name = name }
            func snapshot() -> Int { 0 }
            func snapshot(resolveSessionId: (String) -> String?) -> Int { 0 }
            func cancelAll() -> Int { 0 }
            func refresh() throws -> Int { 0 }
        }

        struct Entry {
            let name: String
            let lane: Lane
        }

        actor Subject {
            var lanes: [Entry] = []
            var cache: [Int] = []
            var appearance = 0

            func save() async {}
            func currentLanes() -> [Entry] { lanes }

            func run(entries: [Entry], tasks: [Task<Int, Never>], attempts: Int, resolver: @escaping (String) -> String?) async throws {
        \(bodyLines)
            }
        }

        """
    }

    // MARK: The real sites

    /*
     `ahraos-macos` `Sources/AhraOsServices/ProviderProxy.swift` on 2026-10-03, its three loops over the account
     lanes. Each lane is its own actor, and the first two await each one's answer before asking the next. The third
     adds into an outer count, which is loop-carried state, as the TypeScript rule's accumulator is.
     */
    static let providerProxy = """
        actor ProviderProxyAccountLane {
            func snapshot(resolveSessionId: (String) -> String?) -> Int { 0 }
            func accountConfiguration() -> String { "" }
            func cancelAll() -> Int { 0 }
        }

        actor ProviderProxy {
            private var lanes: [(name: String, lane: ProviderProxyAccountLane)] = []
            private var resolveSessionId: ((String) -> String?)?

            private func currentLanes() -> [(name: String, lane: ProviderProxyAccountLane)] { lanes }

            /// Lanes are queried in defaults order so the array is stable across reads.
            private func aggregateStatistics() async -> [Int] {
                let resolver = resolveSessionId ?? { _ in nil }
                var accountStatistics: [Int] = []
                for entry in currentLanes() {
                    accountStatistics.append(await entry.lane.snapshot(resolveSessionId: resolver))
                }
                return accountStatistics
            }

            private func settingsView() async -> [String] {
                var accounts: [String] = []
                for entry in currentLanes() {
                    accounts.append(await entry.lane.accountConfiguration())
                }
                return accounts
            }

            private func cancelAllInFlight() async -> Int {
                var cancelled = 0
                for entry in currentLanes() {
                    cancelled += await entry.lane.cancelAll()
                }
                return cancelled
            }
        }

        """

    @Test func providerProxysTwoLaneLoopsAreFlaggedAndItsAccumulatorIsNot() {
        #expect(Self.spans(Self.providerProxy) == ["for entry in currentLanes()", "for entry in currentLanes()"])
    }

    /* The repair: every lane asked at once, each answer kept at its lane's index so the order the comment promises holds. */
    @Test func providerProxysRepairIsClean() {
        let source = """
            actor ProviderProxyAccountLane {
                func snapshot(resolveSessionId: @Sendable (String) -> String?) -> Int { 0 }
            }

            actor ProviderProxy {
                private var lanes: [(name: String, lane: ProviderProxyAccountLane)] = []

                private func aggregateStatistics(resolver: @escaping @Sendable (String) -> String?) async -> [Int] {
                    let entries = lanes
                    return await withTaskGroup(of: (Int, Int).self) { group in
                        for (index, entry) in entries.enumerated() {
                            group.addTask { (index, await entry.lane.snapshot(resolveSessionId: resolver)) }
                        }
                        var statistics = [Int](repeating: 0, count: entries.count)
                        for await (index, snapshot) in group {
                            statistics[index] = snapshot
                        }
                        return statistics
                    }
                }
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    /*
     `ahraos-presence` `Sources/AhraOSPresence/Studio/StudioModel.swift:715` on 2026-10-03: each preset is put on and
     photographed in turn, because the photograph reads what is on screen. The TypeScript rule would call writing
     `this.appearance` a sink and report it; here the write is shared state and the await is on `self`.
     */
    @Test func studioModelsSeedingIsSilent() {
        let source = """
            @MainActor
            final class StudioModel {
                var appearance = 0
                var undoStack: [Int] = []

                func save() async {}

                func seed(presets: [Int]) {
                    let kept = self.appearance
                    Task { @MainActor in
                        for preset in presets {
                            self.appearance = preset
                            await self.save()
                        }
                        self.appearance = kept
                        self.undoStack.removeAll()
                        presenceLog("presence: seeded \\(presets.count) starting characters")
                    }
                }
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    /*
     `Studio/StudioModel.swift:804`: a part worn on each body through a static function, the result written into
     `self.wornParts` and skipped when the part was taken off meanwhile.
     */
    @Test func studioModelsWearingIsSilent() {
        let source = """
            final class Body {}
            struct WornPart {
                static func wear(_ url: String, on body: Body) async throws -> WornPart { WornPart() }
                func remove() {}
            }

            @MainActor
            final class StudioModel {
                var wearings: [String: Int] = [:]
                var wornParts: [String: [WornPart]] = [:]

                func wear(url: String, bodies: [Body], wearing: Int) {
                    Task {
                        for body in bodies {
                            do {
                                let worn = try await WornPart.wear(url, on: body)
                                guard self.wearings[url] == wearing else {
                                    worn.remove()
                                    continue
                                }
                                self.wornParts[url, default: []].append(worn)
                            }
                            catch {
                                presenceLog("presence: could not wear \\(url): \\(error)")
                            }
                        }
                    }
                }
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    /*
     `Studio/Capture/CaptureWorkspace.swift:179`: each dropped provider's URL read through a checked continuation.
     The awaited call is a free function, which the syntax cannot tell from one isolated to this actor: a miss.
     */
    @Test func captureWorkspacesProviderLoopIsSilent() {
        let source = """
            import AppKit

            @MainActor
            final class CaptureWorkspace {
                func accept(_ providers: [NSItemProvider]) {
                    Task { @MainActor in
                        var urls: [URL] = []
                        for provider in providers {
                            if let url = await withCheckedContinuation({ continuation in
                                _ = provider.loadObject(ofClass: URL.self) { url, _ in continuation.resume(returning: url) }
                            }) {
                                urls.append(url)
                            }
                        }
                        open(urls)
                    }
                }

                func open(_ urls: [URL]) {}
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    /*
     `Tests/LibraryTests/PresenceStoreCardParityTests.swift:113`: two stores asked the same question per clip. Both
     stores are one actor each, the same on every pass, and the loop counts ties into an outer `var`.
     */
    @Test func storeParityLoopIsSilent() {
        let source = """
            actor LibrarySimilar {
                func nearest(to code: String) -> [Int] { [] }
            }

            func compare(sample: [String], fromOld: LibrarySimilar, fromNew: LibrarySimilar) async {
                var ties = 0
                for code in sample {
                    let fromOldNearest = await fromOld.nearest(to: code)
                    let fromNewNearest = await fromNew.nearest(to: code)
                    ties += zip(fromOldNearest, fromNewNearest).filter { $0 == $1 }.count
                }
                print(ties)
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    /* `Stage/WindowClimb.swift:721`: each sentence said once the last has settled, paced and logged. */
    @Test func windowClimbsSentencesAreSilent() {
        let source = """
            @MainActor
            final class WindowClimb {
                func wait(until time: Double) async {}
                static func say(_ sentence: String) -> Bool { true }

                func climb(next: [String]) async {
                    func settled() async {}
                    for sentence in next {
                        await settled()
                        await wait(until: 1)
                        presenceLog("climb probe: said")
                        if !Self.say(sentence) { presenceLog("climb probe: not understood") }
                    }
                }
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    /* `Stage/StagePane+TimelineProbes.swift:275`: each generation awaited and the probe abandoned on the first failure. */
    @Test func timelineProbesGenerationsAreSilent() {
        let source = """
            @MainActor
            final class StagePane {
                func generationIds() -> [Int] { [] }
                func generate(_ id: Int, completion: @escaping (Bool) -> Void) {}

                func probe() async {
                    for id in self.generationIds() {
                        let result = await withCheckedContinuation { done in
                            self.generate(id) { done.resume(returning: $0) }
                        }
                        guard result else {
                            presenceLog("gap energy probe: could not write \\(id)")
                            return
                        }
                    }
                }
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    /*
     `ahraos-macos` `Sources/AhraOsServices/PtyPool+SessionVerbs.swift:621` awaits only inside a `Task` it starts per
     session, and `ProviderProxy.swift:979` is a `for try await` over a byte stream. Neither is this shape.
     */
    @Test func macosLoopsThatAreNotThisShapeAreSilent() {
        let source = """
            final class Pty { func kill() {} }

            actor PtyPool {
                var entries: [Int: Pty] = [:]

                func forceReapIfStuck(sessionId: Int) async {}

                func resetAll() {
                    let ids = Array(entries.keys)
                    for sessionId in ids {
                        guard let pty = entries[sessionId] else { continue }
                        pty.kill()
                        Task { [weak self] in
                            await self?.forceReapIfStuck(sessionId: sessionId)
                        }
                    }
                }

                func read(byteStream: AsyncThrowingStream<UInt8, any Error>) async throws -> [UInt8] {
                    var bytes: [UInt8] = []
                    for try await byte in byteStream {
                        bytes.append(byte)
                    }
                    return bytes
                }
            }

            """
        #expect(Self.spans(source).isEmpty)
    }

    // MARK: What it flags

    @Test func everyFlaggedShapeIsFlagged() {
        let cases: [(name: String, body: String, header: String)] = [
            (
                "a dictionary filled by subscript",
                """
                var statistics: [String: Int] = [:]
                for entry in entries {
                    statistics[entry.name] = await entry.lane.snapshot()
                }
                """, "for entry in entries",
            ),
            (
                "a dictionary filled by updateValue and a set by insert",
                """
                var statistics: [String: Int] = [:]
                var depths = Set<Int>()
                for entry in entries {
                    statistics.updateValue(await entry.lane.snapshot(), forKey: entry.name)
                    depths.insert(await entry.lane.queueDepth)
                }
                """, "for entry in entries",
            ),
            (
                "the indexed form over indices",
                """
                var results: [Int] = []
                for index in entries.indices {
                    results.append(await entries[index].lane.snapshot())
                }
                """, "for index in entries.indices",
            ),
            (
                "the indexed form over a range to count",
                """
                var results: [Int] = []
                for index in 0..<entries.count {
                    results.append(await entries[index].lane.snapshot())
                }
                """, "for index in 0..<entries.count",
            ),
            (
                "an enumerated walk filling by index",
                """
                var results: [Int: Int] = [:]
                for (index, entry) in entries.enumerated() {
                    results[index] = await entry.lane.snapshot()
                }
                """, "for (index, entry) in entries.enumerated()",
            ),
            (
                "a loop-invariant outer value read by the await",
                """
                var results: [Int] = []
                for entry in entries {
                    results.append(await entry.lane.snapshot(resolveSessionId: resolver))
                }
                """, "for entry in entries",
            ),
            (
                "an await through a body let bound to the item's chain",
                """
                var results: [Int] = []
                for entry in entries {
                    let lane = entry.lane
                    results.append(await lane.snapshot())
                }
                """, "for entry in entries",
            ),
            (
                "a where clause on the item",
                """
                var results: [Int] = []
                for entry in entries where !entry.name.isEmpty {
                    results.append(await entry.lane.snapshot())
                }
                """, "for entry in entries where !entry.name.isEmpty",
            ),
            (
                "a validation throw that reads no awaited value",
                """
                var results: [Int] = []
                for entry in entries {
                    if entry.name.isEmpty { throw CancellationError() }
                    results.append(await entry.lane.snapshot())
                }
                """, "for entry in entries",
            ),
            (
                "a continue on an awaited value, which only skips this item",
                """
                var results: [Int] = []
                for entry in entries {
                    let value = await entry.lane.snapshot()
                    if value < 0 { continue }
                    results.append(value)
                }
                """, "for entry in entries",
            ),
            (
                "a do and catch that records the failure and moves on",
                """
                var results: [Int] = []
                var failures: [String] = []
                for entry in entries {
                    do {
                        results.append(try await entry.lane.refresh())
                    } catch {
                        failures.append("\\(error)")
                        continue
                    }
                }
                """, "for entry in entries",
            ),
            (
                "a break that leaves only an inner switch",
                """
                var results: [String] = []
                for entry in entries {
                    switch await entry.lane.snapshot() {
                    case 1: break
                    default: results.append(entry.name)
                    }
                }
                """, "for entry in entries",
            ),
            (
                "a throw caught inside the body",
                """
                var failures: [String] = []
                for entry in entries {
                    do {
                        let value = await entry.lane.snapshot()
                        if value < 0 { throw CancellationError() }
                    } catch let error {
                        failures.append("\\(error)")
                    }
                }
                """, "for entry in entries",
            ),
            (
                "a body-local var reassigned within the iteration",
                """
                var results: [Int] = []
                for entry in entries {
                    var value = await entry.lane.snapshot()
                    value += 1
                    results.append(value)
                }
                """, "for entry in entries",
            ),
            (
                "a property read on each item's actor",
                """
                var depths: [Int] = []
                for entry in entries {
                    depths.append(await entry.lane.queueDepth)
                }
                """, "for entry in entries",
            ),
            (
                "an await whose result is dropped",
                """
                for entry in entries {
                    _ = await entry.lane.cancelAll()
                }
                """, "for entry in entries",
            ),
            (
                "a failure logged as an error from the catch",
                """
                var results: [Int] = []
                for entry in entries {
                    do {
                        results.append(try await entry.lane.refresh())
                    } catch {
                        logger.error("refresh failed: \\(error)")
                    }
                }
                """, "for entry in entries",
            ),
            (
                "a sink by prefix whose receiver the body never reads back",
                """
                var document = Document()
                for entry in entries {
                    document.addPage(await entry.lane.snapshot())
                }
                """, "for entry in entries",
            ),
            (
                "a debug block with no effect in it",
                """
                var results: [Int] = []
                for entry in entries {
                    #if DEBUG
                    let checked = entry.name
                    #endif
                    results.append(await entry.lane.snapshot())
                }
                """, "for entry in entries",
            ),
        ]
        for testCase in cases {
            #expect(Self.spans(Self.subject(testCase.body)) == [testCase.header], "\(testCase.name)")
        }
    }

    /*
     A function in this file that reports only from its catch, or only an error through a log receiver, does not make
     the loop calling it ordered.
     */
    @Test func aCalleeThatOnlyReportsFailureDoesNotSilence() {
        let source = """
            actor Lane {
                func refresh() async -> Int {
                    do {
                        return try load()
                    } catch {
                        print("refresh failed")
                        return 0
                    }
                }
                func check() -> Int {
                    if Bool.random() { logger.warning("empty lane") }
                    return 0
                }
                func load() throws -> Int { 0 }
            }

            func refreshAll(lanes: [Lane]) async -> [Int] {
                var results: [Int] = []
                for lane in lanes {
                    results.append(await lane.refresh())
                    results.append(await lane.check())
                }
                return results
            }

            """
        #expect(Self.spans(source) == ["for lane in lanes"])
    }

    /* One finding per loop, at the innermost loop that waits, however many awaits it holds. */
    @Test func theInnermostLoopIsReportedOnce() {
        let source = Self.subject(
            """
            var results: [Int] = []
            for group in [entries, entries] {
                for entry in group {
                    results.append(await entry.lane.snapshot())
                    results.append(await entry.lane.cancelAll())
                }
            }
            """
        )
        #expect(Self.spans(source) == ["for entry in group"])
    }

    /* The span is the header alone, from `for` to the end of the sequence, even across lines. */
    @Test func theSpanIsTheHeader() {
        let source = Self.subject(
            """
            var results: [Int] = []
            for entry
                in entries
            {
                results.append(await entry.lane.snapshot())
            }
            """
        )
        #expect(Self.spans(source) == ["for entry\n    in entries"])
    }

    @Test func theMessageNamesTheRepairsAndCarriesNoEmDash() {
        let message = Self.message
        for phrase in ["withTaskGroup", "with its index", "async let", "bounded"] {
            #expect(message.contains(phrase), "the message should mention \(phrase)")
        }
        #expect(!message.contains("\u{2014}"))
    }

    // MARK: What it leaves alone

    @Test func loopsThatAreNotWalksOverACollectionAreSilent() {
        let bodies = [
            "for try await entry in stream { _ = await entry.lane.snapshot() }",
            "for await entry in stream { _ = await entry.lane.snapshot() }",
            "for attempt in 0..<3 { _ = await entries[attempt].lane.snapshot() }",
            "for attempt in 0..<attempts { _ = await entries[attempt].lane.snapshot() }",
            "for index in 0...entries.count { _ = await entries[index].lane.snapshot() }",
            "for index in 1...3 { _ = await entries[index].lane.snapshot() }",
            "for index in 0...entries.count - 1 { _ = await entries[index].lane.snapshot() }",
            "for index in (0...) { _ = await entries[index].lane.snapshot() }",
            "for index in stride(from: 0, to: entries.count, by: 2) { _ = await entries[index].lane.snapshot() }",
            "for _ in entries { _ = await entries[0].lane.snapshot() }",
            "while let entry = entries.first { _ = await entry.lane.snapshot() }",
            "repeat { _ = await entries[0].lane.snapshot() } while entries.isEmpty",
        ]
        for body in bodies {
            #expect(Self.spans(Self.subject(body)).isEmpty, "\(body)")
        }
    }

    @Test func loopsThatDoNotThemselvesWaitAreSilent() {
        let bodies = [
            /* An await only on the sequence runs once. */
            "var names: [String] = []\nfor entry in await loadEntries() { names.append(entry.name) }",
            /* An await only inside a closure the body writes, which runs when called. */
            "for entry in entries { Task { _ = await entry.lane.snapshot() } }",
            /* An await only inside a nested while belongs to that loop. */
            "for entry in entries { while entry.name.isEmpty { _ = await entry.lane.snapshot() } }",
            /* Awaiting tasks started before the loop serializes nothing. */
            "var results: [Int] = []\nfor task in tasks { results.append(await task.value) }",
            "var results: [Int] = []\nfor index in tasks.indices { results.append(await tasks[index].value) }",
        ]
        for body in bodies {
            #expect(Self.spans(Self.subject(body)).isEmpty, "\(body)")
        }
    }

    /* Stricter than TypeScript: an await not rooted at the item may queue on one actor however the loop is written. */
    @Test func awaitsThatAreNotTheItemsOwnWorkAreSilent() {
        let bodies = [
            /* The same actor, `self`, on every pass, explicitly or through implicit `self`. */
            "for entry in entries { _ = entry.name; await self.save() }",
            "for entry in entries { _ = entry.name; await save() }",
            /* One outer actor, the same on every pass. */
            "let lane = Lane(name: \"\")\nvar results: [Int] = []\nfor entry in entries { results.append(await lane.snapshot(resolveSessionId: { _ in entry.name })) }",
            /* A static member or a free function, whose isolation the syntax cannot see. */
            "var results: [Int] = []\nfor entry in entries { results.append(await Subject.load(entry)) }",
            "var results: [Int] = []\nfor entry in entries { results.append(await withCheckedContinuation { $0.resume(returning: entry.name.count) }) }",
            /* An await that covers a call not rooted at the item, so the syntax cannot say which call suspends. */
            "var results: [Int] = []\nfor entry in entries { await results.append(entry.lane.snapshot()) }",
            "var results: [Int] = []\nfor entry in entries { results.append(await entry.lane.snapshot(resolveSessionId: Resolver.make())) }",
            /* One item-rooted await and one that is not. */
            "var results: [Int] = []\nfor entry in entries { results.append(await entry.lane.snapshot()); await save() }",
            /* An `async let` awaited by its name: the await is on a name, not on a chain reaching the item. */
            "var results: [Int] = []\nfor entry in entries { async let first = entry.lane.snapshot(); results.append(await first) }",
            "var descriptions: [String] = []\nfor entry in entries { async let first = entry.lane.snapshot(); descriptions.append(await first.description) }",
            /* A nested `for await` consumes a stream inside the iteration. */
            "var lines: [String] = []\nfor entry in entries { _ = await entry.lane.snapshot(); for await line in stream { lines.append(line) } }",
            /* An await under `#if DEBUG` is still the loop's own. */
            "var results: [Int] = []\nfor entry in entries {\n#if DEBUG\n_ = await Subject.load(entry)\n#endif\nresults.append(await entry.lane.snapshot()) }",
            /* A body `var` bound to the item's chain may be reassigned before the await. */
            "var results: [Int] = []\nfor entry in entries { var lane = entry.lane; lane = Lane(name: \"\"); results.append(await lane.snapshot()) }",
        ]
        for body in bodies {
            #expect(Self.spans(Self.subject(body)).isEmpty, "\(body)")
        }
    }

    @Test func loopCarriedStateIsSilent() {
        let bodies = [
            /* An outer count added to: `ProviderProxy.cancelAllInFlight`. */
            "var cancelled = 0\nfor entry in entries { cancelled += await entry.lane.cancelAll() }",
            /* An outer variable reassigned. */
            "var last = 0\nfor entry in entries { last = await entry.lane.snapshot() }",
            /* An outer property added to. */
            "var summary = (total: 0, count: 0)\nfor entry in entries { summary.total += await entry.lane.snapshot() }",
            /* Stricter than TypeScript: shared state written, through `self` or implicitly. */
            "for entry in entries { self.appearance = await entry.lane.snapshot() }",
            "for entry in entries { appearance = await entry.lane.snapshot() }",
            "for entry in entries { cache.append(await entry.lane.snapshot()) }",
            "for entry in entries { self.cache.append(await entry.lane.snapshot()) }",
            /* Stricter than TypeScript: a method of `self`, explicit or implicit, may write what every pass shares. */
            "for entry in entries { self.note(entry.name); _ = await entry.lane.snapshot() }",
            "for entry in entries { note(entry.name); _ = await entry.lane.snapshot() }",
            /* A method on a local `var` may be `mutating`. */
            "var counter = Counter()\nfor entry in entries { counter.increment(); _ = await entry.lane.snapshot() }",
            /* A collection read back in the body. */
            "var results: [Int] = []\nfor entry in entries { let value = await entry.lane.snapshot(); results.append(value + results.count) }",
            /* A record consulted before it is filled. */
            "var seen: [Int: Bool] = [:]\nfor entry in entries { let value = await entry.lane.snapshot(); if seen[value] != nil { continue }; seen[value] = true }",
            /* A queue that grows while it is walked. */
            "var queue = entries\nfor entry in queue { if await entry.lane.snapshot() > 0 { queue.append(entry) } }",
            /* A where clause reading what the body fills. */
            "var results: [Int] = []\nfor entry in entries where results.count < 3 { results.append(await entry.lane.snapshot()) }",
            /* A consuming call on an outer receiver, here an object held by a `let`. */
            "let stack = Stack()\nfor entry in entries { _ = stack.popLast(); _ = await entry.lane.snapshot() }",
            /* An outer value passed inout. */
            "var buffer: [Int] = []\nfor entry in entries { Recorder.record(&buffer, await entry.lane.snapshot()) }",
            /* A dictionary filled through a member path the body cannot be read for. */
            "var nested: [[Int]] = [[]]\nfor entry in entries { nested[0][0] += await entry.lane.snapshot() }",
            /* An element of the walked collection written in place writes the collection. */
            "var lanes = entries\nfor index in lanes.indices { let depth = await lanes[index].lane.snapshot(); lanes[index] = Entry(name: \"\\(depth)\", lane: Lane(name: \"\")) }",
        ]
        for body in bodies {
            #expect(Self.spans(Self.subject(body)).isEmpty, "\(body)")
        }
    }

    @Test func anExitOnAnAwaitedValueIsSilent() {
        let bodies = [
            /* Find-first. */
            "for entry in entries { if await entry.lane.snapshot() > 0 { return } }",
            /* A break on a value derived from the await. */
            "for entry in entries { let value = await entry.lane.snapshot(); let done = value > 9; if done { break } }",
            /* A throw on a derived value. */
            "for entry in entries { let value = await entry.lane.snapshot(); if value < 0 { throw CancellationError() } }",
            /* A guard on an awaited value. */
            "for entry in entries { guard await entry.lane.snapshot() > 0 else { return } }",
            /* An `if case` on an awaited value. */
            "for entry in entries { if case 1 = await entry.lane.snapshot() { return } }",
            /* Stricter than TypeScript: an uncaught try stops the loop at the first failure. */
            "var results: [Int] = []\nfor entry in entries { results.append(try await entry.lane.refresh()) }",
            /* A do whose catches do not handle every error. */
            "var results: [Int] = []\nfor entry in entries { do { results.append(try await entry.lane.refresh()) } catch is CancellationError { } }",
            /* A return from the catch of a do that awaits. */
            "for entry in entries { do { _ = try await entry.lane.refresh() } catch { return } }",
            /* A rethrow from the catch of a do that awaits. */
            "for entry in entries { do { _ = try await entry.lane.refresh() } catch { throw error } }",
            /* A return inside a do that awaits first and swallows failure. */
            "for entry in entries { do { _ = try await entry.lane.refresh(); return } catch { } }",
            /* A flag set from the caught error and read after the do. */
            "for entry in entries { var failed = false; do { _ = try await entry.lane.refresh() } catch { failed = error is CancellationError }; if failed { return } }",
            /* A value an inner loop carries backward from an await, so the taint needs a second pass to reach it. */
            "for entry in entries { _ = await entry.lane.snapshot(); var previous = 0; var latest = 0; for other in entries { previous = latest; latest = await other.lane.snapshot() }; if previous > 9 { return } }",
            /* An unconditional break reached only past an awaited continue. */
            "for entry in entries { let value = await entry.lane.snapshot(); if value < 0 { continue }; break }",
            /* A return under a case whose pattern reads an awaited value. */
            "for entry in entries { let value = await entry.lane.snapshot(); switch entry.name.count { case value: return; default: break } }",
        ]
        for body in bodies {
            #expect(Self.spans(Self.subject(body)).isEmpty, "\(body)")
        }
        /* A labeled continue and a labeled break out of the inner loop leave it on an awaited value. */
        let labeled = [
            "outer: for group in [entries] { for entry in group { if await entry.lane.snapshot() > 0 { continue outer } } }",
            "scan: for entry in entries { switch await entry.lane.snapshot() { case 1: break scan; default: break } }",
        ]
        for body in labeled {
            #expect(Self.spans(Self.subject(body)).isEmpty, "\(body)")
        }
    }

    /*
     An ordered effect in the body. Each is reached through a receiver, because a bare lower-case call is already
     silence for another reason (it may be a method of `self`); the bare names are proved one call down, below.
     */
    @Test func anOrderedSideEffectIsSilent() {
        let bodies = [
            "for entry in entries { logger.info(\"\\(entry.name)\"); _ = await entry.lane.snapshot() }",
            "for entry in entries { poolLog.debug(\"\\(entry.name)\"); _ = await entry.lane.snapshot() }",
            "for entry in entries { Thread.sleep(forTimeInterval: 0.25); _ = await entry.lane.snapshot() }",
            "for entry in entries { _ = await entry.lane.snapshot(); progress.increment() }",
            "for entry in entries { _ = await entry.lane.snapshot(); Reporter.reportProgress(1) }",
            "for entry in entries { TraceLog.write(entry.name); _ = await entry.lane.snapshot() }",
            "for entry in entries { FileHandle.standardOutput.write(Data(entry.name.utf8)); _ = await entry.lane.snapshot() }",
            "for entry in entries { continuation.yield(await entry.lane.snapshot()) }",
            "for entry in entries { DispatchQueue.main.asyncAfter(deadline: .now()) {}; _ = await entry.lane.snapshot() }",
            /* Inside a closure the iteration runs. */
            "for entry in entries { _ = await entry.lane.snapshot(); [entry].forEach { logger.info(\"\\($0.name)\") } }",
            /* Progress output from a catch still shows in order. */
            "var results: [Int] = []\nfor entry in entries { do { results.append(try await entry.lane.refresh()) } catch { FileHandle.standardOutput.write(Data(\"skipped\".utf8)) } }",
            /* Under `#if DEBUG`, an effect is still one. */
            "for entry in entries {\n#if DEBUG\nlogger.info(\"\\(entry.name)\")\n#endif\n_ = await entry.lane.snapshot() }",
        ]
        for body in bodies {
            #expect(Self.spans(Self.subject(body)).isEmpty, "\(body)")
        }
    }

    /* A bare call to a global function is silence before its name is read, so the bare names are read one call down. */
    @Test func everyBareOrderedNameIsReadOneCallDown() {
        for effect in [
            "print(\"rendering\")", "debugPrint(1)", "dump(1)", "NSLog(\"rendering\")", "os_log(\"rendering\")",
            "fputs(\"rendering\", stdout)",
            "puts(\"rendering\")", "fflush(stdout)", "usleep(1000)", "presenceLog(\"rendering\")", "log(\"rendering\")",
            "printTree(1)",
            "reportProgress(1)", "updateSpinner()", "delayBriefly()", "waitForIdle()", "pauseBriefly()",
            "throttleRequests()", "backoffAfterFailure()",
            "rateLimitRequests()", "try? await Task.sleep(for: .seconds(1))", "await Task.yield()",
        ] {
            #expect(Self.spans(Self.oneCallDown(effect)).isEmpty, "\(effect)")
        }
        /* The narrowed list: a failure report and a request timer one call down do not make the loop ordered. */
        for quiet in [
            "logger.error(\"failed\")", "logger.warning(\"slow\")", "poolLog.fault(\"lost\")",
            "fputs(\"failed\", stderr)",
            "FileHandle.standardError.write(Data())", "DispatchQueue.main.asyncAfter(deadline: .now()) {}",
        ] {
            #expect(Self.spans(Self.oneCallDown(quiet)) == ["for lane in lanes"], "\(quiet)")
        }
    }

    /* A loop whose awaited method, declared in this file, makes the given call. */
    static func oneCallDown(_ calleeLine: String) -> String {
        """
        actor Lane {
            func render() async -> Int {
                \(calleeLine)
                return 0
            }
        }

        func renderAll(lanes: [Lane]) async -> [Int] {
            var results: [Int] = []
            for lane in lanes {
                results.append(await lane.render())
            }
            return results
        }

        """
    }

    /* The control: the same loop, its callee making no call at all, is flagged, so the silences above are the calls. */
    @Test func theOneCallDownControlIsFlagged() {
        #expect(Self.spans(Self.oneCallDown("let unused = 1")) == ["for lane in lanes"])
    }

    /* The prefilter: a file without both words is not read at all. */
    @Test func thePrefilterNeedsBothWords() {
        let rule = PerformanceNoIndependentAwaitInLoop()
        let without = "func f() { for item in [1] { print(item) } }"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/A.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: without,
            tree: Parser.parse(source: without),
            nodeCount: 0,
        )
        #expect(!rule.applies(to: file))
    }
}
