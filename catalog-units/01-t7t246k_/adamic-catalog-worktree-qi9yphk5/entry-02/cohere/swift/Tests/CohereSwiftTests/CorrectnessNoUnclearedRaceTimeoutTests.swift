import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `cohere-swift/correctness-no-uncleared-race-timeout`, both directions. Neither proving ground
 races a timeout today, so the reporting fixtures are the real races that exist: scratchpad-apple's
 `VoiceRecorder.requestSpeechAuthorization` (a permission prompt raced against ten seconds) with its
 `group.cancelAll()` taken out, the TypeScript rule's own site (`PhiSocialGenerator`'s per-image render bound)
 written as Swift, the survey's shape, and every spelling the rule reads. Each asserts the line and column where
 the finding starts, so a finding on the wrong child fails rather than passing on its count.

 The silent fixtures are `VoiceRecorder` as it stands, every task group in both proving grounds verbatim
 (Presence's `CaptureVerify` gate and control, macOS's `ProviderProxy` statistics, Presence's `MotionSpotlight`
 probe), and the near miss each condition exists for.
 */
struct CorrectnessNoUnclearedRaceTimeoutTests {
    static func file(_ source: String) -> ParsedFile {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        return ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
    }

    /* The findings for a file the rule agrees to read, and none for one it declines, as the pipeline runs it. */
    static func findings(_ source: String) -> [FindingRecord] {
        let file = Self.file(source)
        let found = CorrectnessNoUnclearedRaceTimeout().findings(in: file)
        #expect(
            found.isEmpty || CorrectnessNoUnclearedRaceTimeout().applies(to: file),
            "the prefilter must never hide a finding",
        )
        guard CorrectnessNoUnclearedRaceTimeout().applies(to: file) else { return [] }
        #expect(found.allSatisfy { $0.fixes.isEmpty && $0.suggestions.isEmpty }, "the rule never fixes")
        return found
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    /* scratchpad-apple's `VoiceRecorder.requestSpeechAuthorization`, verbatim but for whether it cancels the loser (a comment holds the line otherwise, so positions agree). */
    static func voiceRecorder(cancelling: Bool) -> String {
        """
        private static func requestSpeechAuthorization() async -> Bool {
            await withTaskGroup(of: Bool.self) { group in
                group.addTask {
                    await withCheckedContinuation { continuation in
                        SFSpeechRecognizer.requestAuthorization { status in
                            continuation.resume(returning: status == .authorized)
                        }
                    }
                }
                group.addTask {
                    try? await Task.sleep(for: .seconds(authorizationTimeout))
                    return false
                }

                // Whichever finishes first wins, and the rest are cancelled.
                let first = await group.next() ?? false
                \(cancelling ? "group.cancelAll()" : "// (the cancelAll taken out)")
                return first
            }
        }
        """
    }

    // MARK: - Reported

    /* VoiceRecorder without its `cancelAll`: an answer in a second still waits ten for the timer. */
    @Test func voiceRecordersRaceWithoutCancelAllIsFound() {
        let found = Self.findings(Self.voiceRecorder(cancelling: false))
        #expect(Self.positions(found) == ["10:9"])
        #expect(found.map(\.endLine) == [13])
        #expect(found.map(\.messageId) == ["unclearedRaceTimeout"])
        #expect(found.map(\.rule) == ["cohere-swift/correctness-no-uncleared-race-timeout"])
    }

    /* The message names the group and the repair; checked against a literal. */
    @Test func theMessageNamesTheRepair() {
        let found = Self.findings(Self.voiceRecorder(cancelling: false))
        #expect(
            found.map(\.message) == [
                "This child only sleeps and then gives up, racing the group's other work for the first result, but nothing cancels the group once that result is in. A task group waits for every child before it returns, so when the work wins the caller still sits out the whole timeout, and a timeout that returns rather than throws bounds nothing, since the group then waits for the work too. Call group.cancelAll() once the first result is in (a defer { group.cancelAll() } at the top of the group's body covers every path), so the losing child is cancelled and its sleep ends at once."
            ]
        )
    }

    /* The TypeScript rule's site, `PhiSocialGenerator`'s ten-minute render bound, as Swift: a caught timeout returns, and so waits. */
    @Test func theRenderBoundWrittenInSwiftIsFound() {
        let source = """
            func renderImage(prompt: String) async -> ImageResult? {
                let imageTimeout = Duration.seconds(10 * 60)
                do {
                    return try await withThrowingTaskGroup(of: ImageResult.self) { renders in
                        renders.addTask { try await generateImage(prompt) }
                        renders.addTask {
                            try await Task.sleep(for: imageTimeout)
                            throw ImageRenderTimeout(seconds: 600)
                        }
                        guard let imageResult = try await renders.next() else { throw ImageRenderTimeout(seconds: 600) }
                        return imageResult
                    }
                }
                catch {
                    log(error)
                    return nil
                }
            }
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["6:13"])
        #expect(found.first?.message.contains("renders.cancelAll()") == true)
    }

    /* The survey's shape: a generic timeout helper whose first result is returned straight out of the group. */
    @Test func aTimeoutHelperWithoutCancelAllIsFound() {
        let source = """
            func withTimeout<Value: Sendable>(_ limit: Duration, _ operation: @escaping @Sendable () async throws -> Value) async throws -> Value {
                try await withThrowingTaskGroup(of: Value.self) { group in
                    group.addTask { try await operation() }
                    group.addTask { try await Task.sleep(for: limit); throw TimeoutError() }
                    return try await group.next()!
                }
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["4:9"])
    }

    /* Returning from inside an `if`, before the body's end, is not cancelling: measured, the group still waits. */
    @Test func anEarlyReturnFromInsideAnIfIsFound() {
        let source = """
            try await withThrowingTaskGroup(of: String.self) { group in
                group.addTask { try await work() }
                group.addTask { try await Task.sleep(for: .seconds(2)); throw TimeoutError() }
                if let first = try await group.next() { return first }
                return nil
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["3:5"])
    }

    /* The spellings: `$0`, `body:`, `_Concurrency.`, `addTaskUnlessCancelled`, `priority:`, `operation:`, `nextResult()`, `nanoseconds:`, a bare `return`. */
    @Test func everySpellingIsFound() {
        let shorthand = """
            await withTaskGroup(of: Int?.self) {
                $0.addTask { await count() }
                $0.addTask { try? await Task.sleep(nanoseconds: 5_000_000_000); return nil }
                return await $0.next() ?? nil
            }
            """
        #expect(Self.positions(Self.findings(shorthand)) == ["3:5"])
        let labeled = """
            await _Concurrency.withTaskGroup(of: Void.self, body: { group in
                group.addTaskUnlessCancelled(priority: .high) { await refresh() }
                group.addTask(priority: .low, operation: {
                    try? await (Task.sleep(for: .seconds(1)))
                    return
                })
                _ = await group.nextResult()
            })
            """
        #expect(Self.positions(Self.findings(labeled)) == ["3:5"])
        let typed = """
            await withThrowingTaskGroup(of: Data.self) { (group: inout ThrowingTaskGroup<Data, any Error>) in
                group.addTask { try await download() }
                group.addTask { await Task.sleep(UInt64(limit)); throw URLError(.timedOut) }
                return try await group.next() ?? Data()
            }
            """
        #expect(Self.positions(Self.findings(typed)) == ["3:5"])
    }

    /* Two timers racing each other: the shorter wins and the longer is waited out, so both are reported, as two TypeScript timers are. */
    @Test func twoRacingTimeoutsAreBothFound() {
        let source = """
            await withTaskGroup(of: String.self) { group in
                group.addTask { try? await Task.sleep(for: .seconds(1)); return "slow" }
                group.addTask { try? await Task.sleep(for: .seconds(2)); return "slower" }
                return await group.next() ?? ""
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:5", "3:5"])
    }

    /* A child running its own group, also named `group`, is another binding; the outer race is still read. */
    @Test func aChildsOwnGroupOfTheSameNameIsAnotherBinding() {
        let source = """
            try await withThrowingTaskGroup(of: [Int].self) { group in
                group.addTask {
                    await withTaskGroup(of: Int.self) { group in
                        for page in 0..<4 { group.addTask { await fetch(page) } }
                        var all: [Int] = []
                        for await value in group { all.append(value) }
                        return all
                    }
                }
                group.addTask { try await Task.sleep(for: .seconds(30)); throw TimeoutError() }
                return try await group.next() ?? []
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["10:5"])
    }

    /* A body that throws on one path still returns on the other: the guard's throw cancels, the return waits. */
    @Test func aBodyThatReturnsAfterAGuardThrowIsFound() {
        let source = """
            try await withThrowingTaskGroup(of: Reply.self) { group in
                group.addTask { try await ask() }
                group.addTask { try await Task.sleep(for: .seconds(5)); throw TimeoutError() }
                guard let reply = try await group.next() else { throw TimeoutError() }
                return reply
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["3:5"])
    }

    /* A last `if` with no `else` can fall through and end the body normally, so it is not a throw on every path. */
    @Test func aLastIfWithoutElseCanFallThrough() {
        let source = """
            try await withThrowingTaskGroup(of: Void.self) { group in
                group.addTask { try await upload() }
                group.addTask { try await Task.sleep(for: .seconds(5)); throw TimeoutError() }
                let first: Void? = try await group.next()
                if first == nil { throw TimeoutError() }
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["3:5"])
    }

    /* A timeout child added in a loop is still a timer racing the work. */
    @Test func aWorkChildAddedInALoopStillRaces() {
        let source = """
            await withTaskGroup(of: Mirror?.self) { group in
                for mirror in mirrors { group.addTask { await probe(mirror) } }
                group.addTask { try? await Task.sleep(for: .seconds(3)); return nil }
                return await group.next() ?? nil
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["3:5"])
    }

    // MARK: - Silent

    /* VoiceRecorder as it stands: the loser is cancelled. */
    @Test func voiceRecorderAsItStandsIsClean() {
        #expect(Self.findings(Self.voiceRecorder(cancelling: true)).isEmpty)
    }

    /* `defer { group.cancelAll() }` at the top covers every path. */
    @Test func aDeferredCancelAllIsClean() {
        let source = """
            try await withThrowingTaskGroup(of: String.self) { group in
                defer { group.cancelAll() }
                group.addTask { try await work() }
                group.addTask { try await Task.sleep(for: .seconds(2)); throw TimeoutError() }
                return try await group.next()
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A known miss: a `cancelAll` on one path silences the group, as any read of a TypeScript handle does. */
    @Test func aCancelAllOnOnePathIsAKnownMiss() {
        let source = """
            try await withThrowingTaskGroup(of: String.self) { group in
                group.addTask { try await work() }
                group.addTask { try await Task.sleep(for: .seconds(2)); throw TimeoutError() }
                if let first = try await group.next() {
                    group.cancelAll()
                    return first
                }
                return nil
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Presence's `CaptureVerify.gate`, both groups verbatim: the sleeps pace the submissions in the body, and `for await` collects every hold. */
    @Test func captureVerifysPacedGroupsAreClean() {
        let source = """
            let holds = await withTaskGroup(of: Hold?.self) { group in
                for (sourceIndex, source) in sources.enumerated() {
                    group.addTask {
                        let clock = StageClock()
                        do {
                            _ = try await CapturePipeline.run(
                                source: source, folder: folder(sourceIndex, source, in: out), settings: CaptureSettings()
                            ) { stage, _ in clock.report(stage) }
                        }
                        catch {
                            presenceLog("presence: VERIFY gate #\\(sourceIndex) \\(source.lastPathComponent) threw: \\(error)")
                            return nil
                        }
                        guard let entered = clock.first, let left = clock.last else { return nil }
                        return Hold(asked: sourceIndex, name: source.lastPathComponent, entered: entered, left: left)
                    }
                    // Asked 50 ms apart, so "the order asked" is an order and not a race.
                    await Task.sleepUnlessCancelled(for: .milliseconds(50))
                }
                var holds: [Hold] = []
                for await hold in group { if let hold { holds.append(hold) } }
                return holds
            }
            let controlHolds = await withTaskGroup(of: Hold.self) { group in
                for job in 0..<3 {
                    group.addTask {
                        await leaky.enter()
                        let entered = CFAbsoluteTimeGetCurrent()
                        await Task.sleepUnlessCancelled(for: .milliseconds(500))
                        let left = CFAbsoluteTimeGetCurrent()
                        await leaky.leave()
                        return Hold(asked: job, name: "job", entered: entered, left: left)
                    }
                    await Task.sleepUnlessCancelled(for: .milliseconds(50))
                }
                var holds: [Hold] = []
                for await hold in group { holds.append(hold) }
                return holds
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* macOS's `ProviderProxy.aggregateStatistics`, verbatim but for a sleep added so the prefilter lets the rule read it: every lane is collected. */
    @Test func providerProxysStatisticsGroupIsClean() {
        let source = """
            let accountStatistics = await withTaskGroup(of: (Int, ProviderProxyAccountStatistics).self) { group in
                for (index, entry) in lanes.enumerated() {
                    group.addTask { (index, await entry.lane.snapshot(resolveSessionId: resolver)) }
                }
                var byLane = [ProviderProxyAccountStatistics?](repeating: nil, count: lanes.count)
                for await (index, statistics) in group {
                    byLane[index] = statistics
                }
                try? await Task.sleep(for: .seconds(1))
                return byLane.compactMap(\\.self)
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Presence's `MotionSpotlight` probe, verbatim but for a timeout child: children added in a nested function name the group there, which the rule does not follow. */
    @Test func motionSpotlightsNestedAddIsClean() {
        let source = """
            await withTaskGroup(of: (String, String, String, Double).self) { group in
                var next = 0
                func add() {
                    guard next < files.count else { return }
                    let file = files[next]
                    next += 1
                    group.addTask {
                        let start = Date()
                        let verdict = Self.figureVerdict(file.fileUrl)
                        return (file.id, file.category, verdict, Date().timeIntervalSince(start))
                    }
                }
                for _ in 0..<8 { add() }
                group.addTask { try? await Task.sleep(for: .seconds(60)); return ("", "", "timeout", 60) }
                let first = await group.next()
                return first
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* `next()` in a loop, `for try await` with a return (a known miss), and a second `next()`: none of them is one first result. */
    @Test func readsThatAreNotOneFirstResultAreClean() {
        let whileLoop = """
            try await withThrowingTaskGroup(of: String?.self) { group in
                group.addTask { try await work() }
                group.addTask { try await Task.sleep(for: .seconds(2)); return nil }
                while let result = try await group.next() { if let result { return result } }
                return nil
            }
            """
        #expect(Self.findings(whileLoop).isEmpty)
        let forAwait = """
            try await withThrowingTaskGroup(of: String.self) { group in
                group.addTask { try await work() }
                group.addTask { try await Task.sleep(for: .seconds(2)); throw TimeoutError() }
                for try await result in group { return result }
                return nil
            }
            """
        #expect(Self.findings(forAwait).isEmpty)
        let both = """
            try await withThrowingTaskGroup(of: String.self) { group in
                group.addTask { try await work() }
                group.addTask { try await Task.sleep(for: .seconds(2)); return "" }
                let first = try await group.next()
                let second = try await group.next()
                return (first ?? "") + (second ?? "")
            }
            """
        #expect(Self.findings(both).isEmpty)
    }

    /* Children that are not a timer: a sleep that then works (a hedged request), a sleep alone (a minimum duration), a throw that awaits, and a timeout from a helper. */
    @Test func childrenThatAreNotATimerAreClean() {
        for child in [
            "try await Task.sleep(for: .milliseconds(200)); return try await fetch(mirror)",
            "try? await Task.sleep(for: .seconds(1))",
            "try await Task.sleep(for: .seconds(1)); throw await failure()",
            "try await Task.sleep(for: .seconds(1)); log(\"late\"); throw TimeoutError()",
            "try await timeout(after: .seconds(1))",
            "try await ContinuousClock().sleep(for: .seconds(1)); throw TimeoutError()",
            "try await Task.sleep(for: .seconds(1)) { tick() }; throw TimeoutError()",
        ] {
            let source = """
                try await withThrowingTaskGroup(of: String.self) { group in
                    group.addTask { try await fetch(primary) }
                    group.addTask { \(child) }
                    return try await group.next() ?? ""
                }
                """
            #expect(Self.findings(source).isEmpty, "\(child)")
        }
    }

    /* A body that throws on every path after the read cancels on its way out (measured), and one that stops the program never waits, so the timer is not waited for. */
    @Test func aBodyThatAlwaysThrowsAfterTheReadIsClean() {
        for ending in [
            "throw Finished(first)",
            "if first == nil { throw TimeoutError() } else { throw Finished(first) }",
            "switch first { case nil: throw TimeoutError()\n    default: throw Finished(first) }",
            "do { throw Finished(first) } catch { throw error }",
            "if first == nil { throw TimeoutError() } else { fatalError(\"raced \\(String(describing: first))\") }",
        ] {
            let source = """
                try await withThrowingTaskGroup(of: String.self) { group in
                    group.addTask { try await work() }
                    group.addTask { try await Task.sleep(for: .seconds(2)); throw TimeoutError() }
                    let first = try await group.next()
                    \(ending)
                }
                """
            #expect(Self.findings(source).isEmpty, "\(ending)")
        }
    }

    /* Other uses of the group are ways it could be cancelled, and silence it: waiting for all, asking if it is empty, handing it on, and rebinding its name. */
    @Test func otherUsesOfTheGroupAreClean() {
        for use in [
            "await group.waitForAll()",
            "if group.isEmpty { log(\"empty\") }",
            "finish(&group)",
            "log(\"\\(group)\")",
            "let group = Other(); group.reset()",
            "[1].forEach { _ in group.cancelAll() }",
        ] {
            let source = """
                await withTaskGroup(of: Int?.self) { group in
                    group.addTask { await count() }
                    group.addTask { try? await Task.sleep(for: .seconds(1)); return nil }
                    let first = await group.next() ?? nil
                    \(use)
                    return first
                }
                """
            #expect(Self.findings(source).isEmpty, "\(use)")
        }
    }

    /* Not a race: one child before the read (the second added after it), a discarding group, a group named `_`, and a member that shares the group's name. */
    @Test func groupsThatDoNotRaceAreClean() {
        let addedAfter = """
            await withTaskGroup(of: Int?.self) { group in
                group.addTask { await count() }
                let first = await group.next() ?? nil
                group.addTask { try? await Task.sleep(for: .seconds(1)); return nil }
                return first
            }
            """
        #expect(Self.findings(addedAfter).isEmpty)
        let discarding = """
            try await withThrowingDiscardingTaskGroup { group in
                group.addTask { try await serve() }
                group.addTask { try await Task.sleep(for: .seconds(1)); throw TimeoutError() }
            }
            """
        #expect(Self.findings(discarding).isEmpty)
        let unnamed = """
            await withTaskGroup(of: Void.self) { _ in
                await Task.sleep(1)
            }
            """
        #expect(Self.findings(unnamed).isEmpty)
        let oneChild = """
            await withTaskGroup(of: Int?.self) { group in
                group.addTask { try? await Task.sleep(for: .seconds(1)); return nil }
                return await group.next() ?? nil
            }
            """
        #expect(Self.findings(oneChild).isEmpty)
    }

    /* A name bound again is another group, and a group from a function of ours, which may cancel itself, is not read. */
    @Test func anotherGroupIsNotRead() {
        let rebound = """
            await withTaskGroup(of: Int?.self) { group in
                group.addTask { await count() }
                group.addTask { try? await Task.sleep(for: .seconds(1)); return nil }
                let group = spare
                return await group.next() ?? nil
            }
            """
        #expect(Self.findings(rebound).isEmpty)
        let helper = """
            await withCancellingTaskGroup(of: Int?.self) { group in
                group.addTask { await count() }
                group.addTask { try? await Task.sleep(for: .seconds(1)); return nil }
                return await group.next() ?? nil
            }
            """
        #expect(Self.findings(helper).isEmpty)
    }

    /* `other.group` names a member, not the binding, and does not silence the race; a function named like a group function on another receiver is not one. */
    @Test func aMemberNamedGroupIsNotTheBinding() {
        let member = """
            await withTaskGroup(of: Int?.self) { group in
                group.addTask { await count(settings.group) }
                group.addTask { try? await Task.sleep(for: .seconds(1)); return nil }
                return await group.next() ?? nil
            }
            """
        #expect(Self.positions(Self.findings(member)) == ["3:5"])
        let otherReceiver = """
            await scheduler.withTaskGroup(of: Int?.self) { group in
                group.addTask { await count() }
                group.addTask { try? await Task.sleep(for: .seconds(1)); return nil }
                return await group.next() ?? nil
            }
            """
        #expect(Self.findings(otherReceiver).isEmpty)
    }

    /* The prefilter declines a file with no task group or no sleep. */
    @Test func thePrefilterDeclinesFilesWithoutTheShape() {
        #expect(!CorrectnessNoUnclearedRaceTimeout().applies(to: Self.file("let value = try await work()")))
        #expect(
            !CorrectnessNoUnclearedRaceTimeout().applies(to: Self.file("await withTaskGroup(of: Int.self) { _ in }"))
        )
        #expect(CorrectnessNoUnclearedRaceTimeout().applies(to: Self.file(Self.voiceRecorder(cancelling: true))))
    }
}
