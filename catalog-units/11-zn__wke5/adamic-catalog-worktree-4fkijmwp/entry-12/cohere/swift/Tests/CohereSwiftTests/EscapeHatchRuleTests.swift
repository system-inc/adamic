import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/* Fixture pairs for print by target kind, escape hatches with reasons, discarded try?, failure messages and TODO comments. */
struct EscapeHatchRuleTests {
    static func file(_ source: String, kind: String = "library") -> ParsedFile {
        ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: kind,
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
    }

    static func lines(_ rule: some FileRule, _ source: String, kind: String = "library") -> [Int] {
        let subject = file(source, kind: kind)
        guard rule.applies(to: subject) else { return [] }
        return rule.findings(in: subject).map(\.line)
    }

    /* A typed rule given no symbols: what syntax alone decides. */
    static func lines(_ rule: some TypedFileRule, _ source: String, kind: String = "library") -> [Int] {
        let subject = file(source, kind: kind)
        guard rule.applies(to: subject) else { return [] }
        return rule.findings(in: subject, symbols: FileSymbols([])).map(\.line)
    }

    @Test func printIsDecidedByTargetKind() {
        let tool = "import Foundation\nprint(\"usage: tool <file>\")\n"
        let app = "import SwiftUI\nprint(\"tapped\")\n"
        #expect(Self.lines(ConsistencyNoPrint(), tool, kind: "library") == [2])
        #expect(
            Self.lines(ConsistencyNoPrint(), tool, kind: "executable").isEmpty,
            "a command-line tool's stdout is its interface",
        )
        #expect(
            Self.lines(ConsistencyNoPrint(), tool, kind: "application") == [2],
            "an app's model file imports only Foundation and is still app code",
        )
        #expect(Self.lines(ConsistencyNoPrint(), app, kind: "test").isEmpty)
    }

    @Test func aMethodNamedPrintIsNotPrint() {
        #expect(Self.lines(ConsistencyNoPrint(), "printer.print(page)\nlet print = 3\n").isEmpty)
        #expect(Self.lines(ConsistencyNoPrint(), "debugPrint(value)\n") == [1])
    }

    /* The finding is filed under TypeScript's id for the same defect, and says which kind of target nobody reads. */
    @Test func printSaysWhoseStdoutNobodyReads() {
        let library = ConsistencyNoPrint().findings(in: Self.file("print(\"loaded\")\n"))
        let app = ConsistencyNoPrint().findings(in: Self.file("print(\"tapped\")\n", kind: "application"))
        #expect(library.map(\.messageId) == ["consistencyNoConsole"])
        #expect(
            library.map(\.message) == [
                "Do not call 'print'. In a library it writes to a stdout nobody reads. Log through os.Logger, so the message lands in the unified log with a subsystem, a category and a level."
            ]
        )
        #expect(
            app.map(\.message) == [
                "Do not call 'print'. In an app it writes to a stdout nobody reads. Log through os.Logger, so the message lands in the unified log with a subsystem, a category and a level."
            ]
        )
    }

    @Test func escapeHatchesNeedACommentDirectlyAbove() {
        let source = """
            final class Bare: @unchecked Sendable {}

            /* The buffer is only touched on the render queue. */
            final class Explained: @unchecked Sendable {}

            /* A comment two paragraphs up is about something else. */

            nonisolated(unsafe) var orphaned = 0
            // The callback is registered once, before any thread starts.
            nonisolated(unsafe) var explained = 0
            @preconcurrency import Darwin
            """
        #expect(Self.lines(ConcurrencyRequireEscapeHatchReason(), source) == [1, 8, 11])
    }

    @Test func plainSendableAndNonisolatedAreNotHatches() {
        #expect(
            Self.lines(
                ConcurrencyRequireEscapeHatchReason(),
                "struct Value: Sendable {}\nnonisolated func work() {}\n",
            ).isEmpty
        )
    }

    @Test func discardedTryOptionalsAreFound() {
        let source = """
            func cleanUp() {
                try? FileManager.default.removeItem(at: url)
                _ = try? load()
                let kept = try? load()
                if let value = try? load() { use(value) }
            }
            """
        #expect(Self.lines(CorrectnessNoDiscardedTryOptional(), source) == [2, 3])
    }

    /* A discarded sleep is consistency-no-hand-rolled-delay's, which names the house primitive, so this rule leaves it and every other try? on the same lines is still found. */
    @Test func aDiscardedSleepIsTheDelayRulesNotThisOnes() {
        let source = """
            func pause() async {
                try? await Task.sleep(for: .seconds(1))
                _ = try? await Task.sleep(nanoseconds: 500)
                try? await Task.sleep(until: .now + .seconds(1), clock: .continuous)
                try? await save()
            }
            """
        #expect(Self.lines(CorrectnessNoDiscardedTryOptional(), source) == [4, 5])
        #expect(Self.lines(ConsistencyNoHandRolledDelay(), source) == [2, 3])
    }

    /* A sole `try?` is its body's value wherever the body returns one (the shapes @system_cohere_swift_ahraos_presence found on Presence), and a discard wherever it does not. */
    @Test func aSoleTryOptionalThatIsItsBodysValueIsNotADiscard() {
        let values = """
            let looks = data.flatMap { try? decoder.decode([Look].self, from: $0) } ?? []
            let ledges = candidates.compactMap { try? self.fit(ledge: $0) }
            func latency() -> Double? { try? OutputLatency.read() }
            var duration: Double? { try? asset.duration() }
            var cached: Data? {
                get { try? Data(contentsOf: url) }
            }
            let chosen = if fast { try? quick() } else { try? slow() }
            let picked = switch mode {
            case .fast: try? quick()
            default: nil
            }
            """
        #expect(Self.lines(CorrectnessNoDiscardedTryOptional(), values).isEmpty)
        let discards = """
            func cleanUp() { try? FileManager.default.removeItem(at: url) }
            func reset() -> Void { try? store.clear() }
            if stale { try? cache.purge() }
            switch mode {
            case .fast: try? quick()
            default: break
            }
            let runs = items.map { item in
                try? item.write()
                return item
            }
            """
        #expect(Self.lines(CorrectnessNoDiscardedTryOptional(), discards) == [1, 2, 3, 5, 9])
    }

    @Test func failuresWithoutMessagesAreFound() {
        let source = """
            fatalError()
            fatalError("")
            fatalError("the cache was built before the store opened")
            preconditionFailure()
            assertionFailure("unreachable: every case returns above")
            """
        #expect(Self.lines(FatalErrorMessage(), source) == [1, 2, 4])
    }

    @Test func todoCommentsAreFound() {
        let source = """
            // TODO: handle resize
            /* FIXME the offset is off by one */
            // keep a todo list in the task tree instead
            // TODOS are words, but TODOLIST is not a marker
            let value = 1
            """
        #expect(Self.lines(Todo(), source) == [1, 2])
    }
}

/*
 `correctness-no-discarded-try-optional` on a closure's sole `try?`, which the types decide. The unit cases parse a source
 string and hand the rule symbols built by position, one occurrence at every token named in `resolving`, with
 the symbols and names the index recorded for these declarations on a real build; the signatures come from the
 toolchain's own demangler. The end-to-end case runs a real package, so the symbols come from the index too.
 */
extension EscapeHatchRuleTests {
    typealias Declaration = (symbol: String, name: String)

    static let forEach: Declaration = ("s:STsE7forEachyyy7ElementQzKXEKF", "forEach(_:)")
    static let dispatchAsync: Declaration = (
        "s:So17OS_dispatch_queueC8DispatchE5async5group3qos5flags7executeySo0a1_b1_F0CSg_AC0D3QoSVAC0D13WorkItemFlagsVyyXLtF",
        "async(group:qos:flags:execute:)",
    )
    static let dispatchSync: Declaration = (
        "s:So17OS_dispatch_queueC8DispatchE4sync7executexxyKXE_tKlF", "sync(execute:)",
    )
    static let taskType: Declaration = ("s:ScT", "Task")
    static let taskInit: Declaration = (
        "s:ScT12_Concurrencys5NeverORs_rlE4name8priority9operationScTyxACGSSSg_ScPSgxyYaYAcntcfc",
        "init(name:priority:operation:)",
    )
    static let taskDetached: Declaration = (
        "s:ScT12_Concurrencys5NeverORs_rlE8detached4name8priority9operationScTyxACGSSSg_ScPSgxyYaYAcntFZ",
        "detached(name:priority:operation:)",
    )
    static let compactMap: Declaration = ("s:STsE10compactMapySayqd__Gqd__Sg7ElementQzKXEKlF", "compactMap(_:)")
    static let optionalFlatMap: Declaration = (
        "s:Sq7flatMapyqd_0_SgABxqd__YKXEqd__YKs5ErrorRd__Ri_d_0_r0_lF", "flatMap(_:)",
    )
    static let map: Declaration = ("s:SlsE3mapySayqd__Gqd__7ElementQzqd_0_YKXEqd_0_YKs5ErrorRd_0_r0_lF", "map(_:)")
    static let withAnimation: Declaration = ("s:7SwiftUI13withAnimationyxAA0D0VSg_xyKXEtKlF", "withAnimation(_:_:)")
    static let buttonAction: Declaration = ("s:7SwiftUI6ButtonV6action5labelACyxGyyc_xyXEtcfc", "init(action:label:)")
    static let buttonTitled: Declaration = (
        "s:7SwiftUI6ButtonVA2A4TextVRszrlE_6actionACyAEGAA18LocalizedStringKeyV_yyctcfc", "init(_:action:)",
    )
    static let onAppear: Declaration = ("s:7SwiftUI4ViewPAAE8onAppear7performQryycSg_tF", "onAppear(perform:)")
    static let task: Declaration = (
        "s:7SwiftUI4ViewPAAE4task4name8priority4file4line_QrSSSg_ScPSSSiyyYaYAcntF", "task(name:priority:file:line:_:)",
    )
    static let addObserver: Declaration = (
        "c:objc(cs)NSNotificationCenter(im)addObserverForName:object:queue:usingBlock:",
        "addObserver(forName:object:queue:using:)",
    )
    static let run: Declaration = ("s:7Control6RunnerV3runyyyyXEF", "run(_:)")
    static let perform: Declaration = ("s:7Control6RunnerV7perform5label_ySS_yyXEtF", "perform(label:_:)")
    static let value: Declaration = ("s:7Control6RunnerV5valueyxxyXElF", "value(_:)")
    static let both: Declaration = ("s:7Control6RunnerV4both5first6secondyyyXE_SiSgyXEtF", "both(first:second:)")
    static let optional: Declaration = ("s:7Control6RunnerV8optional10completionyyycSg_tF", "optional(completion:)")
    static let variadic: Declaration = ("s:7Control6RunnerV8variadicyyyycd_tF", "variadic(_:)")
    static let multiple: Declaration = ("s:7Control6RunnerV8multiple1a1byyyXE_SiSgyXEtF", "multiple(a:b:)")
    /* A `forEach` of ours that returns what its body returns, so its closure's `try?` is the value. */
    static let oursForEach: Declaration = ("s:7Control5BatchV7forEachySayxGxSiKXEKlF", "forEach(_:)")

    /*
     The findings, as `messageId@line`, with every token named in `resolving` resolved to its declarations, and
     with a result builder recorded at every `{` when `builder` is set.
     */
    static func typedFindings(_ source: String, resolving: [String: [Declaration]], builder: Bool = false) -> [String] {
        let subject = file(source)
        var occurrences: [FileSymbols.Occurrence] = []
        for token in subject.tree.tokens(viewMode: .sourceAccurate) {
            let location = subject.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            for declaration in resolving[token.text] ?? [] {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: declaration.symbol,
                        name: declaration.name,
                        isReference: true,
                    )
                )
            }
            if builder, token.tokenKind == .leftBrace {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: "s:7SwiftUI11ViewBuilderV10buildBlockyxxAA0C0RzlFZ",
                        name: "buildBlock(_:)",
                        isReference: true,
                        isImplicit: true,
                    )
                )
            }
        }
        let rule = CorrectnessNoDiscardedTryOptional()
        guard rule.applies(to: subject) else { return [] }
        return rule.findings(in: subject, symbols: FileSymbols(occurrences)).map { "\($0.messageId)@\($0.line)" }
    }

    @Test func aSoleTryOptionalInAClosureThatReturnsNothingIsADiscard() {
        let source = """
            items.forEach { try? save($0) }
            items.forEach { item in try? save(item) }
            DispatchQueue.main.async { try? write() }
            Button(action: { try? write() }) { Text("x") }
            Button("y") { try? write() }
            view.onAppear { try? write() }
            view.task { try? await send() }
            """
        let found = Self.typedFindings(
            source,
            resolving: [
                "forEach": [Self.forEach], "async": [Self.dispatchAsync], "Button": [Self.buttonAction],
                "onAppear": [Self.onAppear], "task": [Self.task],
            ],
        )
        #expect(
            found == [
                "discardedTryOptionalInClosure@1", "discardedTryOptionalInClosure@2", "discardedTryOptionalInClosure@3",
                "discardedTryOptionalInClosure@4",
            ] + ["discardedTryOptionalInClosure@6", "discardedTryOptionalInClosure@7"],
            "\(found)",
        )
        let titled = Self.typedFindings(
            "Button(\"y\") { try? write() }\n",
            resolving: ["Button": [Self.taskType, Self.buttonTitled]],
        )
        #expect(
            titled == ["discardedTryOptionalInClosure@1"],
            "a trailing closure after a positional title binds to the action: \(titled)",
        )
    }

    /* A generic result is the closure's value: what `map`, `compactMap`, `flatMap`, `sync` and `withAnimation` return. */
    @Test func aSoleTryOptionalWhoseClosureReturnsAGenericIsAValue() {
        let source = """
            let plans = items.compactMap { try? fit($0) }
            let looks = data.flatMap { try? decode($0) }
            let results = items.map { try? save($0) }
            let read = DispatchQueue.main.sync { try? load() }
            DispatchQueue.main.sync { try? write() }
            withAnimation { try? write() }
            """
        let found = Self.typedFindings(
            source,
            resolving: [
                "compactMap": [Self.compactMap], "flatMap": [Self.optionalFlatMap], "map": [Self.map],
                "sync": [Self.dispatchSync], "withAnimation": [Self.withAnimation],
            ],
        )
        #expect(found.isEmpty, "\(found)")
    }

    /* A `Task`'s operation returns the task's value, so its `try?` is dropped exactly when the task's handle is. */
    @Test func aTaskWhoseHandleIsDroppedDropsItsTryOptional() {
        let source = """
            Task { try? await send() }
            Task.detached { try? await send() }
            _ = Task { try? await send() }
            let task = Task { try? await send() }
            let value = await Task.detached { try? read() }.value
            runner.run { Task { try? await send() } }
            let tasks = items.map { _ in Task { try? await send() } }
            func start() -> Task<Void?, Never> { Task { try? await send() } }
            """
        let found = Self.typedFindings(
            source,
            resolving: [
                "Task": [Self.taskType, Self.taskInit], "detached": [Self.taskDetached], "run": [Self.run],
                "map": [Self.map],
            ],
        )
        #expect(
            found == [
                "discardedTryOptionalInClosure@1", "discardedTryOptionalInClosure@2", "discardedTryOptionalInClosure@3",
                "discardedTryOptionalInClosure@6",
            ],
            "\(found)",
        )
    }

    /* Labeled, positional, trailing, optional and additional trailing closures bind as the compiler binds them; where the binding is unsure, nothing is flagged. */
    @Test func theClosureIsMatchedToItsParameter() {
        let source = """
            runner.run { try? write() }
            runner.run({ try? write() })
            runner.perform { try? write() }
            runner.perform(label: "x") { try? write() }
            runner.optional { try? write() }
            runner.optional(completion: { try? write() })
            runner.multiple { try? write() } b: { try? load() }
            runner.value { try? load() }
            runner.both { try? load() }
            runner.variadic({ try? write() }, { try? write() })
            """
        let found = Self.typedFindings(
            source,
            resolving: [
                "run": [Self.run], "perform": [Self.perform], "optional": [Self.optional], "multiple": [Self.multiple],
                "value": [Self.value], "both": [Self.both], "variadic": [Self.variadic],
            ],
        )
        #expect(found == (1...7).map { "discardedTryOptionalInClosure@\($0)" }, "\(found)")
    }

    /* A written signature or annotation decides without the types. */
    @Test func aWrittenReturnTypeDecides() {
        let source = """
            let work: () -> Void = { try? write() }
            let later: (@Sendable () -> Void)? = { try? write() }
            let explicit = { () -> Void in try? write() }
            let inferred = { try? write() }
            let typed: () -> Int? = { try? load() }
            let returning = { () -> Int? in try? load() }
            """
        let found = Self.typedFindings(source, resolving: [:])
        #expect(
            found == [
                "discardedTryOptionalInClosure@1", "discardedTryOptionalInClosure@2", "discardedTryOptionalInClosure@3",
            ],
            "\(found)",
        )
    }

    /* A callee not recorded, not Swift's, a look-alike of ours that returns a value, or a closure that is a result builder's body: not flagged. */
    @Test func whatTheTypesCannotTellIsNotFlagged() {
        let source = """
            items.forEach { try? save($0) }
            center.addObserver(forName: name, object: nil, queue: nil) { _ in try? write() }
            """
        #expect(Self.typedFindings(source, resolving: [:]).isEmpty)
        #expect(
            Self.typedFindings(source, resolving: ["forEach": [Self.oursForEach], "addObserver": [Self.addObserver]])
                .isEmpty
        )
        #expect(
            Self.typedFindings(source, resolving: ["forEach": [Self.forEach, Self.compactMap]]).isEmpty,
            "two declarations at one name are not one answer",
        )
        #expect(Self.typedFindings(source, resolving: ["forEach": [Self.forEach]], builder: true).isEmpty)
    }

    /* The statement shapes are found as before, beside the closure ones, in source order. */
    @Test func statementAndClosureFindingsInterleaveInOrder() {
        let source = """
            func clean() {
                items.forEach { try? save($0) }
                try? write()
                _ = try? load()
            }
            """
        let found = Self.typedFindings(source, resolving: ["forEach": [Self.forEach]])
        #expect(
            found == ["discardedTryOptionalInClosure@2", "discardedTryOptional@3", "discardedTryOptional@4"],
            "\(found)",
        )
    }

    @Test func demangledSignaturesAreReadCarefully() throws {
        let asynchronous = try #require(
            CorrectnessNoDiscardedTryOptional.Signature(
                demangled:
                    "(extension in Dispatch):__C.OS_dispatch_queue.async(group: __C.OS_dispatch_group?, qos: Dispatch.DispatchQoS, flags: Dispatch.DispatchWorkItemFlags, execute: @escaping @convention(block) () -> ()) -> ()",
                name: "async(group:qos:flags:execute:)",
            )
        )
        #expect(asynchronous.labels == ["group", "qos", "flags", "execute"])
        #expect(
            asynchronous.parameters.map(CorrectnessNoDiscardedTryOptional.Signature.kind(of:)) == [
                .other, .other, .other, .function(returning: "()"),
            ]
        )
        #expect(asynchronous.result == "()")
        let mapping = try #require(
            CorrectnessNoDiscardedTryOptional.Signature(
                demangled:
                    "(extension in Swift):Swift.Collection.map<A, B where B1: Swift.Error>((A.Element) throws(B1) -> A1) throws(B1) -> [A1]",
                name: "map(_:)",
            )
        )
        #expect(
            mapping.parameters.map(CorrectnessNoDiscardedTryOptional.Signature.kind(of:)) == [
                .function(returning: "A1")
            ]
        )
        let task = try #require(
            CorrectnessNoDiscardedTryOptional.Signature(
                demangled:
                    "(extension in _Concurrency):Swift.Task< where B == Swift.Never>.init(name: Swift.String?, priority: Swift.TaskPriority?, operation: __owned @isolated(any) () async -> A) -> Swift.Task<A, Swift.Never>",
                name: "init(name:priority:operation:)",
            )
        )
        #expect(
            task.parameters.map(CorrectnessNoDiscardedTryOptional.Signature.kind(of:)) == [
                .other, .other, .function(returning: "A"),
            ]
        )
        #expect(CorrectnessNoDiscardedTryOptional.Signature.taskSuccess(task.result) == "A")
        #expect(
            CorrectnessNoDiscardedTryOptional.Signature(
                demangled: "Control.Runner.run(() -> ()) -> ()",
                name: "run(_:_:)",
            ) == nil,
            "parameters that do not match the name's labels",
        )
        #expect(
            CorrectnessNoDiscardedTryOptional.Signature(
                demangled: "Control.Runner.run(label: () -> ()) -> ()",
                name: "run(_:)",
            ) == nil,
            "a printed label that is not the name's",
        )
        #expect(
            CorrectnessNoDiscardedTryOptional.Signature(
                demangled: "run #1 (() -> ()) -> () in Control.start() -> ()",
                name: "run(_:)",
            ) == nil,
            "a local function is not found by name",
        )
        #expect(
            CorrectnessNoDiscardedTryOptional.Signature(demangled: "Control.Runner.run(() -> ()", name: "run(_:)")
                == nil,
            "text that does not balance",
        )
    }

    @Test(arguments: [
        ("(A.Element) throws -> ()", CorrectnessNoDiscardedTryOptional.Signature.Kind.function(returning: "()")),
        ("(() -> ())?", .function(returning: "()")),
        ("@escaping @Sendable () -> ()", .function(returning: "()")),
        ("sending @escaping @isolated(any) () async throws -> A", .function(returning: "A")),
        ("() -> () -> ()", .function(returning: "() -> ()")),
        ("Swift.String?", .other),
        ("__C.OS_dispatch_group?", .other),
        ("(Swift.Int, Swift.String)", .other),
        ("[Swift.String : Any]", .other),
        ("A", .unknown),
        ("A1", .unknown),
        ("A.Element", .unknown),
        ("Any", .unknown),
        ("any Swift.Sendable", .unknown),
        ("@autoclosure () -> A?", .unknown),
        ("() -> ()...", .unknown),
    ])
    func parameterTypesAreClassified(type: String, kind: CorrectnessNoDiscardedTryOptional.Signature.Kind) {
        #expect(CorrectnessNoDiscardedTryOptional.Signature.kind(of: type) == kind)
    }

    /*
     End to end on a real package, symbols from the index the build wrote and the rule run by the pipeline:
     the standard library's `forEach` drops its closure's `try?`, and a `forEach` of ours that returns what
     its body returns keeps it, as `compactMap` does.
     */
    static let discardPackageSource = """
        struct Batch {
            var values: [Int]

            func forEach<Each>(_ body: (Int) throws -> Each) rethrows -> [Each] {
                try values.map(body)
            }
        }

        func save(_ value: Int) throws {}

        func keep(items: [Int], batch: Batch) -> Int {
            try? save(0)
            items.forEach { try? save($0) }
            let decoded = items.compactMap { try? save($0) }
            let saved = batch.forEach { try? save($0) }
            return saved.count + decoded.count
        }

        """

    @Test func theStandardLibrarysForEachDropsItsTryOptionalAndOursDoesNot() async throws {
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
        try Self.discardPackageSource.write(
            to: sources.appendingPathComponent("Control.swift"),
            atomically: true,
            encoding: .utf8,
        )
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix"],
            workingDirectory: root,
        )
        var stream = Data()
        _ = try await Pipeline(options: options, writer: ContractWriter { stream.append($0) }, workingDirectory: root)
            .run()
        var found: [String] = []
        for line in stream.split(separator: UInt8(ascii: "\n")) {
            let record = try #require(try JSONSerialization.jsonObject(with: Data(line)) as? [String: Any])
            if record["kind"] as? String == "finding",
                record["rule"] as? String == CorrectnessNoDiscardedTryOptional().name
            {
                found.append("\(record["messageId"] as? String ?? "")@\(record["line"] as? Int ?? 0)")
            }
        }
        #expect(
            found == ["discardedTryOptional@12", "discardedTryOptionalInClosure@13"],
            "expected the statement and the standard library's forEach, not compactMap's or ours: \(found)",
        )
    }
}
