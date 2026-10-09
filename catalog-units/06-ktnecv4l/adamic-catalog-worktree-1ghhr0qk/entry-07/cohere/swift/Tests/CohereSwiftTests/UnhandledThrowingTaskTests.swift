import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `unhandled-throwing-task` both ways. The unit cases parse a source string and hand the rule symbols built
 by position, as the index records them: the concurrency `Task` (`s:ScT`) at every `Task` a call names, and
 at the name that starts the task (`Task` itself, or `detached`, `immediate`, `init`) the declaration the
 compiler picked, the one where the failure type is `any Error` when the closure can throw and the one where
 it is `Never` when it cannot. So each case says what the compiler would have resolved, and the rule's syntax
 is judged only for where the task's value goes. The triggering cases are SwiftLint's documented examples for
 `unhandled_throwing_task`, each asserting its line and column; the non-triggering cases are SwiftLint's own
 and the look-alikes the types tell apart. The end-to-end case runs a real package, so the symbols come from
 the index the build wrote.
 */
@Suite(.serialized)
struct UnhandledThrowingTaskTests {
    static let task = "s:ScT"
    static let ours = "s:7Control4JobsO4TaskV"

    /* What the compiler resolved a task's starting name to, as the probe package recorded each. */
    enum Resolution {
        case throwing
        case never
        case ours
        case unresolved
    }

    static let throwingStarters = [
        "Task": "s:ScT12_Concurrencys5Error_pRs_rlE4name8priority9operationScTyxsAB_pGSSSg_ScPSgxyYaKYAcntcfc",
        "init": "s:ScT12_Concurrencys5Error_pRs_rlE4name8priority9operationScTyxsAB_pGSSSg_ScPSgxyYaKYAcntcfc",
        "detached":
            "s:ScT12_Concurrencys5Error_pRs_rlE8detached4name8priority9operationScTyxsAB_pGSSSg_ScPSgxyYaKYAcntFZ",
        "immediate":
            "s:ScT12_Concurrencys5Error_pRs_rlE9immediate4name8priority18executorPreference9operationScTyxsAB_pGSSSg_ScPSgSch_pSgnxyYaKYAcntFZ",
        "immediateDetached":
            "s:ScT12_Concurrencys5Error_pRs_rlE17immediateDetached4name8priority18executorPreference9operationScTyxsAB_pGSSSg_ScPSgSch_pSgnxyYaKYAcntFZ",
    ]

    static let neverStarters = [
        "Task": "s:ScT12_Concurrencys5NeverORs_rlE4name8priority9operationScTyxACGSSSg_ScPSgxyYaYAcntcfc",
        "init": "s:ScT12_Concurrencys5NeverORs_rlE4name8priority9operationScTyxACGSSSg_ScPSgxyYaYAcntcfc",
        "detached": "s:ScT12_Concurrencys5NeverORs_rlE8detached4name8priority9operationScTyxACGSSSg_ScPSgxyYaYAcntFZ",
        "immediate":
            "s:ScT12_Concurrencys5NeverORs_rlE9immediate4name8priority18executorPreference9operationScTyxACGSSSg_ScPSgSch_pSgnxyYaYAcntFZ",
    ]

    /*
     The findings, as `line:column`. Each call naming `Task` takes the next resolution, in source order (the
     last repeats), at its starting name: the last of `Task`, `init`, `detached`, `immediate` and
     `immediateDetached` it spells. `extra` adds occurrences by position, for a result builder's records.
     */
    static func findings(
        _ source: String,
        _ resolutions: Resolution...,
        extra: [(line: Int, column: Int, symbol: String)] = [],
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
        var occurrences: [FileSymbols.Occurrence] = extra.map {
            FileSymbols.Occurrence(line: $0.line, column: $0.column, symbol: $0.symbol, name: "", isReference: true)
        }
        var index = 0
        for call in Self.taskCalls(in: file.tree) {
            let resolution = resolutions.isEmpty ? .throwing : resolutions[min(index, resolutions.count - 1)]
            index += 1
            let names = call.calledExpression.tokens(viewMode: .sourceAccurate).filter {
                Self.throwingStarters[$0.text] != nil
            }
            guard let starter = names.last else { continue }
            for token in names where token.text == "Task" {
                let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: resolution == .ours ? Self.ours : Self.task,
                        name: "Task",
                        isReference: true,
                    )
                )
            }
            let location = file.locations.location(for: starter.positionAfterSkippingLeadingTrivia)
            let symbol: String? =
                switch resolution {
                    case .throwing: Self.throwingStarters[starter.text]
                    case .never: Self.neverStarters[starter.text]
                    case .ours: "s:7Control4JobsO4TaskVyAEyyYaKccfc"
                    case .unresolved: nil
                }
            if let symbol {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: symbol,
                        name: "init(name:priority:operation:)",
                        isReference: true,
                    )
                )
            }
        }
        return UnhandledThrowingTask().findings(in: file, symbols: FileSymbols(occurrences)).map {
            "\($0.line):\($0.column)"
        }
    }

    /* Every call whose called expression names `Task`, outermost first. */
    static func taskCalls(in tree: SourceFileSyntax) -> [FunctionCallExprSyntax] {
        final class Collector: SyntaxVisitor {
            var calls: [FunctionCallExprSyntax] = []
            override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
                if node.calledExpression.tokens(viewMode: .sourceAccurate).contains(where: { $0.text == "Task" }) {
                    calls.append(node)
                }
                return .visitChildren
            }
        }
        let collector = Collector(viewMode: .sourceAccurate)
        collector.walk(tree)
        return collector.calls
    }

    /* SwiftLint's triggering examples, each closure one the compiler gives an `any Error` failure type. */
    @Test(arguments: [
        "Task {\n  try await myThrowingFunction()\n}\n",
        "Task {\n  let text = try myThrowingFunction()\n  return text\n}\n",
        "Task {\n  do {\n    try myThrowingFunction()\n  }\n}\n",
        "Task {\n  do {\n    try myThrowingFunction()\n  } catch let e as FooError {\n    print(e)\n  }\n}\n",
        "Task {\n  do {\n    throw FooError.bar\n  }\n}\n",
        "Task {\n  throw FooError.bar\n}\n",
        "Task<_, _> {\n  throw FooError.bar\n}\n",
        "Task<Void,_> {\n  throw FooError.bar\n}\n",
        "Task {\n  do {\n    try foo()\n  } catch {\n    try bar()\n  }\n}\n",
        "Task {\n  do {\n    try foo()\n  } catch {\n    throw BarError()\n  }\n}\n",
    ])
    func incumbentTriggeringExamplesAreFound(source: String) {
        #expect(Self.findings(source, .throwing) == ["1:1"])
    }

    /* SwiftLint's last triggering example: a task that is the only statement of a function returning nothing. */
    @Test func aTaskInAFunctionReturningNothingIsFound() {
        let source = """
            func doTask() {
              Task {
                try await someThrowingFunction()
              }
            }

            """
        #expect(Self.findings(source, .throwing) == ["2:3"])
    }

    /* SwiftLint's non-triggering examples whose closures cannot throw: the compiler picks the `Never` declaration. */
    @Test(arguments: [
        "Task<Void, Never> {\n  try await myThrowingFunction()\n}\n",
        "Task {\n  try? await myThrowingFunction()\n}\n",
        "Task {\n  try! await myThrowingFunction()\n}\n",
        "Task {\n  do {\n    try myThrowingFunction()\n  } catch let e {\n    print(e)\n  }\n}\n",
        "func someFunction() throws {\n  Task {\n    anotherFunction()\n    do {\n      try myThrowingFunction()\n    } catch {\n      print(error)\n    }\n  }\n\n  try something()\n}\n",
        "Task {\n  return Result {\n      try someThrowingFunc()\n  }\n}\n",
    ])
    func incumbentNonThrowingExamplesAreNotFound(source: String) {
        #expect(Self.findings(source, .never).isEmpty)
    }

    /* SwiftLint's non-triggering examples whose tasks can throw but are held, read or returned, and one with a failure type written by hand. */
    @Test(arguments: [
        "Task<Void, String> {\n  let text = try myThrowingFunction()\n  return text\n}\n",
        "let task = Task {\n  try await myThrowingFunction()\n}\n",
        "var task = Task {\n  try await myThrowingFunction()\n}\n",
        "try await Task {\n  try await myThrowingFunction()\n}.value\n",
        "executor.task = Task {\n  try await isolatedOpen(.init(executor.asUnownedSerialExecutor()))\n}\n",
        "let result = await Task {\n  throw CancellationError()\n}.result\n",
        "func makeTask() -> Task<String, Error> {\n  return Task {\n    try await someThrowingFunction()\n  }\n}\n",
        "func makeTask() -> Task<String, Error> {\n  // Implicit return\n  Task {\n    try await someThrowingFunction()\n  }\n}\n",
    ])
    func incumbentHeldTasksAreNotFound(source: String) {
        #expect(Self.findings(source, .throwing).isEmpty)
    }

    /*
     The compiler's answer, not the spelling, says whether the closure throws: an operation passed by name, a
     narrow `catch` and `for try await` throw with no `try` SwiftLint would read as unhandled, and a `try` in a
     nested closure does not reach the task.
     */
    @Test func whetherTheClosureThrowsIsTheCompilersAnswer() {
        let throwing = """
            Task(operation: work)
            Task {
                do {
                    try await load()
                } catch is CancellationError {}
            }
            Task {
                for try await line in lines {
                    print(line)
                }
            }

            """
        #expect(Self.findings(throwing, .throwing) == ["1:1", "2:1", "7:1"])
        let nested = """
            Task {
                let parse = { try decode() }
                await use(parse)
            }

            """
        #expect(Self.findings(nested, .never).isEmpty)
    }

    /* Every other way to start a task: SwiftLint reads only `Task`, and these drop the error the same way. */
    @Test func everyStarterIsFound() {
        let source = """
            Task.detached { try await load() }
            Task.detached(priority: .low) { try await load() }
            Task.immediate { try await load() }
            Task.immediateDetached { try await load() }
            Task.init { try await load() }
            _Concurrency.Task { try await load() }
            Task<Void, _>.detached { try await load() }
            Task(priority: .high) { try await load() }

            """
        #expect(Self.findings(source, .throwing) == ["1:1", "2:1", "3:1", "4:1", "5:1", "6:1", "7:1", "8:1"])
        #expect(Self.findings(source, .never).isEmpty)
    }

    /* A failure type written by hand, `any Error` or anything but `_`, is the writer's choice, as SwiftLint reads it. */
    @Test func aFailureTypeWrittenByHandIsNotFound() {
        let source = """
            Task<Void, any Error> { try await load() }
            Task<Void, Error>.detached { try await load() }
            Task<_, Error> { try await load() }

            """
        #expect(Self.findings(source, .throwing).isEmpty)
    }

    /* The reason the rule is typed: a `Task` of ours, and one the compiler resolved to nothing, are never found. */
    @Test func aTaskOfOursOrUnresolvedIsNotFound() {
        let source = """
            Task { try await load() }
            Task.detached { try await load() }

            """
        #expect(Self.findings(source, .ours).isEmpty)
        #expect(Self.findings(source, .unresolved).isEmpty)
    }

    /* Statements whose value goes nowhere: each of several, and a lone one in a body that returns nothing. */
    @Test func droppedStatementsAreFound() {
        let source = """
            final class Loader {
                var state = 0 {
                    didSet { Task { try await load() } }
                }
                var mirrored: Int {
                    get { state }
                    set { Task { try await load() } }
                }
                init() { Task { try await load() } }
                deinit { Task { try await load() } }
                func voidly() -> Void { Task { try await load() } }
                func empty() -> () { Task { try await load() } }
                func several() -> Int {
                    Task { try await load() }
                    return 1
                }
                func bodies(items: [Int], ready: Bool) {
                    for _ in items { Task { try await load() } }
                    while ready { Task { try await load() } }
                    repeat { Task { try await load() } } while ready
                    do { Task { try await load() } }
                    guard ready else { Task { try await load() }; return }
                    defer { Task { try await load() } }
                    if ready { Task { try await load() } } else if items.isEmpty { Task { try await load() } } else { Task { try await load() } }
                    switch items.count {
                    case 0: Task { try await load() }
                    default: Task { try await load() }
                    }
                    #if DEBUG
                    Task { try await load() }
                    #endif
                    let closure = { () -> Void in Task { try await load() } }
                    let longer = {
                        Task { try await load() }
                        print("started")
                    }
                    _ = (closure, longer)
                }
            }
            Task { try await load() }

            """
        let found = Self.findings(source, .throwing)
        #expect(
            found == [
                "3:18", "7:15", "9:14", "10:14", "11:29", "12:26", "14:9", "18:26", "19:23", "20:18", "21:14", "22:28",
                "23:17",
                "24:20", "24:72", "24:107", "26:17", "27:18", "30:9", "32:39", "34:13", "40:1",
            ]
        )
    }

    /* Statements whose value may be returned or is held: SwiftLint flags the argument, the array, the closure and the builder. */
    @Test func heldTasksAreNotFound() {
        let source = """
            var tasks: [Task<Void, any Error>] = []
            tasks.append(Task { try await load() })
            let pair = [Task { try await load() }, Task { try await load() }]
            let mapped = urls.map { url in Task { try await fetch(url) } }
            _ = Task { try await load() }
            current = Task { try await load() }
            Task { try await load() }.cancel()
            var pending: Task<Void, any Error> { Task { try await load() } }
            var guarded: Task<Void, any Error> {
                get { Task { try await load() } }
            }
            func make() -> Task<Void, any Error> { Task { try await load() } }
            func choose(ready: Bool) -> Task<Void, any Error>? {
                if ready { Task { try await load() } } else { nil }
            }
            func pick(count: Int) -> Task<Void, any Error> {
                switch count {
                case 0: Task { try await load() }
                default: Task { try await load() }
                }
            }
            let chosen = if ready { Task { try await load() } } else { Task { try await load() } }
            Button("Save") { Task { try await save() } }

            """
        #expect(Self.findings(source, .throwing).isEmpty)
        /* The outer task returns the inner one and cannot throw itself; the inner one is its value. */
        #expect(Self.findings("Task { Task { try await load() } }\n", .never, .throwing).isEmpty)
    }

    /*
     A result builder's body collects its statements: the compiler records `buildBlock` at the body's brace,
     and `buildExpression`, when the builder has one, at the statement's first token.
     */
    @Test func aResultBuildersComponentIsNotFound() {
        let source = """
            @Collect
            func collected() -> [Task<Void, any Error>] {
                Task { try await load() }
                Task { try await load() }
            }

            """
        let buildBlock = "s:7Control7CollectO10buildBlockySayScTyyts5Error_pGGAGd_tFZ"
        let buildExpression = "s:7Control7CollectO15buildExpressionySayScTyyts5Error_pGGAFFZ"
        #expect(Self.findings(source, .throwing) == ["3:5", "4:5"])
        #expect(Self.findings(source, .throwing, extra: [(2, 45, buildBlock)]).isEmpty)
        #expect(Self.findings(source, .throwing, extra: [(3, 5, buildExpression), (4, 5, buildExpression)]).isEmpty)
        let nested = """
            func make() {
                collecting {
                    Task { try await load() }
                    Task { try await load() }
                }
                Task { try await load() }
                print("done")
            }

            """
        #expect(Self.findings(nested, .throwing, extra: [(2, 16, buildBlock)]) == ["6:5"])
    }

    @Test func aFileWithNoTaskDoesNotApply() {
        let source = "func load() async throws {}\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!UnhandledThrowingTask().applies(to: file))
    }

    /*
     End to end on a real package, symbols from the index the build wrote: a dropped throwing task is found,
     started by `Task` and by `Task.detached`, and a `Task` of our own declaring, a task that cannot throw, one
     that handles its errors, and one that is held are not.
     */
    static let packageSource = """
        enum Jobs {
            struct Task {
                init(_ body: @escaping () async throws -> Void) {}
            }

            static func run() {
                Task { try await load() }
            }
        }

        func load() async throws {}

        func quietly() async {}

        func start() async throws {
            Task { try await load() }
            Task.detached { try await load() }
            Task { await quietly() }
            Task {
                do {
                    try await load()
                } catch {}
            }
            Task { try? await load() }
            let held = Task { try await load() }
            try await held.value
        }

        """

    @Test func aDroppedThrowingTaskIsFlaggedAndATaskOfOursIsNot() async throws {
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
        let candidates = parsed.files.filter { UnhandledThrowingTask().applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            UnhandledThrowingTask().findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.messageId)@\($0.line):\($0.column)"
            }
        }
        #expect(
            found == ["unhandledThrowingTask@16:5", "unhandledThrowingTask@17:5"],
            "expected the two dropped throwing tasks and not ours, the quiet one, the handled ones or the held one: \(found)",
        )
    }
}
