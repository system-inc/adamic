import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `cohere-swift/consistency-no-hand-rolled-delay`, both directions. The reporting fixtures are
 the real sites: the two the survey found in Presence's `CapturePhone/Sources/PhoneRig.swift` as they stand, and
 the spellings each project's primitive replaced, taken from the commits that introduced it (Presence's 47f22c42,
 macOS's 00f1b8e), then the do/catch the other rule's repair would write. Each asserts the line and column where
 the finding starts, so a finding on the wrong node fails rather than passing on its count.

 The silent fixtures are both projects' `sleepUnlessCancelled` verbatim (the definition is not a use), macOS's
 `DelayWidthUntilIdle` debounce (a catch that returns), Presence's `WindowClimb` (a `try` that passes the
 cancellation up), and the near misses each condition exists for.
 */
struct ConsistencyNoHandRolledDelayTests {
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
        let found = ConsistencyNoHandRolledDelay().findings(in: file)
        #expect(
            found.isEmpty || ConsistencyNoHandRolledDelay().applies(to: file),
            "the prefilter must never hide a finding",
        )
        guard ConsistencyNoHandRolledDelay().applies(to: file) else { return [] }
        #expect(found.allSatisfy { $0.fixes.isEmpty && $0.suggestions.isEmpty }, "the rule never fixes")
        return found
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    // MARK: - Reported

    /* PhoneRig.swift:129: a listener that failed waits a second before starting again. */
    @Test func phoneRigsRestartPauseIsFound() {
        let source = """
            listener.stateUpdateHandler = { [weak self] state in
                if case .failed = state {
                    Task { @MainActor in
                        try? await Task.sleep(for: .seconds(1))
                        self?.listen()
                    }
                }
            }
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["4:13"])
        #expect(found.map(\.endColumn) == [52])
        #expect(found.map(\.messageId) == ["handRolledDelay"])
        #expect(found.map(\.rule) == ["cohere-swift/consistency-no-hand-rolled-delay"])
    }

    /* PhoneRig.swift:260: the flash holds for 120 milliseconds. */
    @Test func phoneRigsFlashPauseIsFound() {
        let source = """
            case .flash:
                self.flashing = true
                Task { @MainActor in
                    try? await Task.sleep(for: .milliseconds(120))
                    self.flashing = false
                }
            """
        #expect(Self.positions(Self.findings(source)) == ["4:9"])
    }

    /* The messages name the primitive, both its spellings, and say the do/catch is not the repair; checked against literals. */
    @Test func theMessagesNameThePrimitive() {
        let tryOptional = Self.findings("func f() async {\n    try? await Task.sleep(for: .seconds(1))\n    work()\n}")
        #expect(
            tryOptional.map(\.message) == [
                "This try? await Task.sleep is a hand-rolled Task.sleepUnlessCancelled. Task.sleep throws only CancellationError, so the try? does nothing but let the pause end early when the task is cancelled, which is all the primitive does. Write await Task.sleepUnlessCancelled with the same duration (declared once per project, in Task+SleepUnlessCancelled.swift, taking for: or nanoseconds:), so the pause reads as a plain pause and the one place that says why cancellation is let go says it for every call. A do/catch with an empty catch around the sleep is the same pause spelled longer, not the repair."
            ]
        )
        let doCatch = Self.findings(
            "func f() async {\n    do { try await Task.sleep(for: .seconds(1)) } catch {}\n    work()\n}"
        )
        #expect(
            doCatch.map(\.message) == [
                "This do/catch around Task.sleep, with nothing in the catch, is the house's Task.sleepUnlessCancelled written out by hand: Task.sleep throws only CancellationError, so the catch does nothing but let the pause end early when the task is cancelled, which is all the primitive does. Write await Task.sleepUnlessCancelled with the same duration (declared once per project, in Task+SleepUnlessCancelled.swift, taking for: or nanoseconds:), so the pause reads as a plain pause and the one place that says why cancellation is let go says it for every call."
            ]
        )
        #expect(doCatch.map(\.messageId) == ["handRolledDelayDoCatch"])
    }

    /* macOS's AhraOsHttpClient before 00f1b8e: a liveness poll's pause, its duration on the next line. */
    @Test func aPollsPauseSpreadOverLinesIsFound() {
        let source = """
            let livenessMonitor = Task {
                while !Task.isCancelled {
                    try? await Task.sleep(
                        nanoseconds: UInt64(Self.livenessCheckSeconds * 1_000_000_000)
                    )
                    if Task.isCancelled { return }
                }
            }
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["3:9"])
        #expect(found.map(\.endLine) == [5])
    }

    /* macOS before 00f1b8e (`if flushNanos > 0`) and Presence before 47f22c42 (`if wait > 0`): the only statement of an `if` is a statement. */
    @Test func aPauseAloneInAnIfIsFound() {
        let source = """
            func flush(flushNanos: UInt64, wait: Double) async {
                if flushNanos > 0 { try? await Task.sleep(nanoseconds: flushNanos) }
                if wait > 0 { try? await Task.sleep(for: .seconds(wait)) }
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:25", "3:19"])
    }

    /* Presence's MotionSpotlight before 47f22c42: a duration read from a property, and a `guard`'s and a loop's statement. */
    @Test func aPropertyDurationInAGuardAndALoopIsFound() {
        let source = """
            func preview(entry: Entry) async {
                guard entry.isReady else {
                    try? await Task.sleep(for: entry.previewDelay)
                    return
                }
                for _ in 0..<3 {
                    try? await Task.sleep(for: entry.previewDelay)
                }
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["3:9", "7:9"])
    }

    /* SwiftTerm's iOSTerminalView: the vendored terminal waits with nanoseconds. */
    @Test func aVendoredNanosecondPauseIsFound() {
        let source = """
            Task { @MainActor in
                try? await Task.sleep(nanoseconds: 100_000_000)
                self.becomeFirstResponder()
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:5"])
    }

    @Test func aDiscardAssignmentIsFound() {
        let source = """
            func f() async {
                _ = try? await Task.sleep(for: .seconds(1))
                work()
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:9"])
    }

    @Test func theConcurrencyModulesTaskAndParenthesesAreFound() {
        let source = """
            func f() async {
                try? await _Concurrency.Task.sleep(for: .seconds(1))
                try? await (Task.sleep(for: .seconds(1)))
                try? (await Task.sleep(for: .seconds(1)))
                try? await Task
                    .sleep(for: .seconds(1))
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:5", "3:5", "4:5", "5:5"])
    }

    /* A file whose only pause is spread over lines never spells `Task.sleep`, so the prefilter must not ask for it. */
    @Test func aPauseWhoseNameIsSpreadOverLinesIsFound() {
        let source = """
            func f() async {
                try? await Task
                    .sleep(for: .seconds(1))
                work()
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:5"])
    }

    /* What `correctness-no-discarded-try-optional`'s repair writes for the PhoneRig flash: the same pause, spelled longer. */
    @Test func anEmptyCatchAroundTheSleepIsFound() {
        let source = """
            Task { @MainActor in
                do {
                    try await Task.sleep(for: .milliseconds(120))
                }
                catch {
                    // Cancelled: the flash is over either way.
                }
                self.flashing = false
            }
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["2:5"])
        #expect(found.map(\.endLine) == [7])
        #expect(found.map(\.messageId) == ["handRolledDelayDoCatch"])
    }

    /* The do/catch as the only statement of a closure is still a statement, and a label changes nothing. */
    @Test func aDoCatchAloneInAClosureAndALabeledOneAreFound() {
        let source = """
            let pause = { do { try await Task.sleep(for: .seconds(1)) } catch {} }
            func f() async {
                settle: do { try await Task.sleep(nanoseconds: 500_000_000) } catch {}
                work()
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["1:15", "3:13"])
    }

    /* A function with a fixed duration is a use with a name around it, not a definition. */
    @Test func aHelperWithAFixedDurationIsFound() {
        let source = """
            func settle() async {
                try? await Task.sleep(for: .seconds(1))
            }
            func settleAgain() async {
                do { try await Task.sleep(for: .seconds(1)) } catch {}
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:5", "5:5"])
    }

    /* Presence's GazeCalibrationSession before 47f22c42: the duration is computed from the parameter, not the parameter. */
    @Test func aDurationComputedFromAParameterIsFound() {
        let source = """
            func wait(_ seconds: Double) async { try? await Task.sleep(for: .milliseconds(Int(seconds * 1000))) }
            """
        #expect(Self.positions(Self.findings(source)) == ["1:38"])
    }

    @Test func aParameterDurationBesideOtherStatementsIsFound() {
        let source = """
            func pause(for duration: Duration) async {
                log("pausing")
                try? await Task.sleep(for: duration)
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["3:5"])
    }

    /* The duration is the outer function's parameter, not the one the pause is the body of. */
    @Test func anOuterFunctionsParameterIsFound() {
        let source = """
            func outer(duration: Duration) -> () async -> Void {
                func inner() async {
                    do { try await Task.sleep(for: duration) } catch {}
                }
                return inner
            }
            func other(label: String) async {
                do { try await Task.sleep(for: defaultPause) } catch {}
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["3:9", "8:5"])
    }

    // MARK: - Silent

    /* macOS's Task+SleepUnlessCancelled.swift, verbatim: the definition of the primitive. */
    @Test func macOsPrimitiveIsSilent() {
        let source = """
            import Foundation

            extension Task where Success == Never, Failure == Never {
                public static func sleepUnlessCancelled(nanoseconds: UInt64) async {
                    do {
                        try await Task.sleep(nanoseconds: nanoseconds)
                    }
                    catch {
                        /*
                         * Task.sleep throws only CancellationError, and only to say the wait
                         * ended early.
                         */
                    }
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Presence's Task+SleepUnlessCancelled.swift, verbatim: the parameter's name inside the body is its second name. */
    @Test func presencePrimitiveIsSilent() {
        let source = """
            extension Task where Success == Never, Failure == Never {
                /// Sleeps for `duration`, returning early, with nothing thrown, if the task is cancelled.
                static func sleepUnlessCancelled(for duration: Duration) async {
                    do {
                        try await Task.sleep(for: duration)
                    }
                    catch {
                        // Cancelled: the only error Task.sleep throws, and the pause is simply over.
                    }
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* The same primitive as a closure, by a named parameter, a typed one, and `$0`. */
    @Test func closureDefinitionsAreSilent() {
        let source = """
            let pause: (Duration) async -> Void = { duration in do { try await Task.sleep(for: duration) } catch {} }
            let typed = { (nanoseconds: UInt64) async in do { try await Task.sleep(nanoseconds: nanoseconds) } catch {} }
            let shorthand: (Duration) async -> Void = { do { try await Task.sleep(for: $0) } catch {} }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A second definition by `try?` is a duplicate helper, left alone as the TypeScript rule leaves `function sleep(ms)`. */
    @Test func aSecondDefinitionIsSilent() {
        let source = """
            func pause(for duration: Duration) async { try? await Task.sleep(for: duration) }
            func pause(_ seconds: UInt64) async { _ = try? await Task.sleep(nanoseconds: seconds) }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* macOS's DelayWidthUntilIdle.swift:124: a cancelled debounce returns before it commits, which the primitive would not. */
    @Test func aCatchThatReturnsIsSilent() {
        let source = """
            debounceTask = Task { @MainActor in
                do {
                    try await Task.sleep(for: idleAfter)
                }
                catch {
                    return // cancelled by a newer change
                }
                if let pending = pendingWidth {
                    frozenWidth = pending
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Presence's WindowClimb.swift:124 and :340: the cancellation is passed up, so the climb stops. */
    @Test func aTryThatPassesCancellationUpIsSilent() {
        let source = """
            private static func settled(_ body: VRMEntity) async throws {
                for _ in 0..<30 where still < 6 {
                    try await Task.sleep(for: .seconds(0.1))
                }
                if self.settles { try await Task.sleep(for: .seconds(0.3)) }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A `try?` whose value is read uses the cut-short as a signal. */
    @Test func aTryOptionalWhoseValueIsReadIsSilent() {
        let source = """
            func f() async {
                if (try? await Task.sleep(for: .seconds(1))) == nil { return }
                let finished = try? await Task.sleep(nanoseconds: 1_000)
                guard (try? await Task.sleep(for: .seconds(1))) != nil else { return }
                use(finished)
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* The `try?` that is a closure's only statement is the closure's value, which takes types to judge. */
    @Test func aPauseAloneInAClosureIsSilent() {
        let source = """
            Task { try? await Task.sleep(for: .seconds(1)) }
            let wait = { try? await Task.sleep(nanoseconds: 1_000) }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func argumentsThePrimitiveDoesNotTakeAreSilent() {
        let source = """
            func f(deadline: ContinuousClock.Instant) async {
                try? await Task.sleep(for: .seconds(1), tolerance: .milliseconds(10))
                try? await Task.sleep(for: .seconds(1), clock: .suspending)
                try? await Task.sleep(until: deadline)
                try? await Task.sleep(until: deadline, clock: .continuous)
                try? await Task.sleep(1_000_000)
                try? await Task.sleep(seconds: 1)
                try? await Task.sleep()
                do { try await Task.sleep(for: .seconds(1), tolerance: .zero) } catch {}
                do { try await Task.sleep(until: deadline) } catch {}
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func otherSleepsAreSilent() {
        let source = """
            func f(clock: some Clock<Duration>) async {
                try? await clock.sleep(for: .seconds(1))
                try? await ContinuousClock().sleep(for: .seconds(1))
                try? await Task<Never, Never>.sleep(for: .seconds(1))
                try? await Other.Task.sleep(for: .seconds(1))
                try? await Task.pause(for: .seconds(1))
                try? await task.sleep(for: .seconds(1))
                Thread.sleep(forTimeInterval: 1)
                try! await Task.sleep(for: .seconds(1))
                await Task.sleepUnlessCancelled(for: .seconds(1))
                await Task.sleepUnlessCancelled(nanoseconds: 500_000_000)
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A timer nothing can cancel is the TypeScript shape's own semantics, and not the Swift primitive's. */
    @Test func anUncancellableTimerIsSilent() {
        let source = """
            func f() async {
                await withCheckedContinuation { continuation in
                    DispatchQueue.main.asyncAfter(deadline: .now() + 1) { continuation.resume() }
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func doCatchesThatDoMoreAreSilent() {
        let source = """
            func f() async throws {
                do {
                    try await Task.sleep(for: .seconds(1))
                    try await send()
                } catch {}
                do { try await Task.sleep(for: .seconds(1)) } catch { log("cancelled") }
                do { try await Task.sleep(for: .seconds(1)) } catch is CancellationError {}
                do { try await Task.sleep(for: .seconds(1)) } catch let error {}
                do { try await Task.sleep(for: .seconds(1)) } catch is CancellationError {} catch {}
                do throws(CancellationError) { try await Task.sleep(for: .seconds(1)) } catch {}
                do { try? await Task.sleep(for: .seconds(1)) } catch {}
                do { await Task.sleepUnlessCancelled(for: .seconds(1)) } catch {}
            }
            """
        /* The `try?` inside the `do` on line 11 is a pause statement of its own, so it is the one finding, and the `do` around it is not. */
        #expect(Self.positions(Self.findings(source)) == ["11:10"])
    }

    @Test func aFileWithoutTheWordsIsDeclined() {
        let file = Self.file("func f() async { await pause() }")
        #expect(!ConsistencyNoHandRolledDelay().applies(to: file))
    }
}
