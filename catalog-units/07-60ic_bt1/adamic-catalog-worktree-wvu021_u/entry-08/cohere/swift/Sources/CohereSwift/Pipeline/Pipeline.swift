import Foundation

/*
 One run: describe the package, then fix and format, types, lint, unused, in that order, each phase's
 outcome stated whether or not it ran.

 The order is cohere's and the reasons are the same. Mutation runs first so everything downstream sees
 the repaired tree. A file that does not parse stops the run there, because rewriting or type-checking a
 tree the parser could not read produces nonsense. Compiler errors stop lint, because rule findings
 against code whose semantics are wrong are noise; compiler warnings do not, because warning-level code
 still means what it says.

 What counts as complete is narrower than "every phase ran". A phase the caller turned off (`--lint`
 alone, `--no-fix`, an opt-in report nobody asked for) withholds nothing they did not choose to withhold.
 A phase that was cut off, a phase that is not built yet, a file that could not be read, or a file the
 compiler left no record for all withhold findings nobody chose to give up, and any of them makes the
 run incomplete and its exit code 1.
 */
public struct Pipeline {
    private let options: CommandOptions
    private let writer: ContractWriter
    private let workingDirectory: URL
    private let runner: ProcessRunner

    public init(
        options: CommandOptions,
        writer: ContractWriter,
        workingDirectory: URL,
        runner: ProcessRunner = ProcessRunner(),
    ) {
        self.options = options
        self.writer = writer
        self.workingDirectory = workingDirectory
        self.runner = runner
    }

    /* A run that could not happen at all. Exit 2, no summary: the front door says nothing was checked. */
    public struct RunFailure: Error, CustomStringConvertible {
        public var description: String
    }

    public func run() async throws -> Int32 {
        let toolchain = EngineVersion.toolchain(runner: runner)
        try writer.write(EngineVersion.provenance(toolchain: toolchain))
        if options.showVersion {
            return 0
        }

        let configuration = try RuleConfiguration.load(
            packageRoot: options.root ?? workingDirectory,
            explicitPath: options.lintConfiguration,
        )
        if options.listRules || options.listRulesEnabled {
            for name in RuleRegistry.allNames {
                let severity = configuration.severity(of: name)
                if options.listRulesEnabled && severity == .off {
                    continue
                }
                try writer.write(RuleRecord(name: name, severity: severity.rawValue))
            }
            return 0
        }

        guard let root = options.root else {
            throw RunFailure(description: "no --root was given, so there is no package to check")
        }
        guard FileManager.default.fileExists(atPath: root.appendingPathComponent("Package.swift").path) else {
            throw RunFailure(description: "\(root.path) holds no Package.swift, so there is no package to check")
        }

        /*
         The abbreviation vocabulary, loaded before anything is checked. A compiled-in copy that would not load
         refuses the run with exit 2: a naming rule with no words to judge would report nothing and read as a
         clean tree, the silent green this tool exists to stop. A run with the rule turned off loads nothing,
         and its rule list holds an empty vocabulary that is never consulted.
         */
        var vocabulary = AbbreviationVocabulary()
        if configuration.severity(of: ConsistencyNoAbbreviatedIdentifier.ruleName) != .off
            && (options.runFix || options.runLint)
        {
            do {
                vocabulary = try AbbreviationVocabulary.compiledIn()
            }
            catch {
                throw RunFailure(
                    description:
                        "the naming rules' compiled-in words (policy/Abbreviations.json) could not be loaded, so nothing was checked: \(error)"
                )
            }
        }
        let fileRules = RuleRegistry.fileRules(vocabulary: vocabulary)

        let describeStart = Date()
        let package = try PackageModel.load(
            root: root,
            scratchPath: Self.scratchPath(for: root),
            runner: runner,
            toolchain: toolchain,
        )
        let fileSet = try FileSet.build(package: package)
        let scope = try FileScope.resolve(options: options, fileSet: fileSet, workingDirectory: workingDirectory)
        /*
         A `.swift-format` left at or above any file of ours refuses the run, whatever phases were asked for, as a
         leftover Prettier config refuses a TypeScript run: Swift has one house format, and a file still there is
         one that `Format.sh` or an editor reads and cohere never will.
         */
        let leftovers = HouseSwiftFormat.leftoverConfigurationFiles(
            above: fileSet.owned.map(\.url),
            boundary: repositoryRoot(of: root),
        )
        guard leftovers.isEmpty else {
            throw RunFailure(description: leftovers.map(HouseSwiftFormat.refusal(of:)).joined(separator: "\n"))
        }
        if !fileSet.note.isEmpty {
            FileHandle.standardError.write(Data("note: \(fileSet.note)\n".utf8))
        }
        try writer.write(
            projectRecord(
                package: package,
                fileSet: fileSet,
                scope: scope,
                elapsed: Self.milliseconds(since: describeStart),
            )
        )

        var complete = true

        /* Parsing is the fix phase's first step: it is what the fixers and the formatter would rewrite. */
        let fixStart = Date()
        var parsed = await SourceParser().parse(scope.files)
        for error in parsed.parseErrors {
            try writer.write(error)
        }
        /* After `project` and before the first `phase`, as the contract places them, so the front door can name every file nothing checked. */
        for crash in parsed.unreadable {
            complete = false
            try writer.write(UnreadableRecord(file: crash.file, error: crash.error))
        }
        let filesThatDoNotParse = Set(parsed.parseErrors.map(\.file)).count
        /* What the fixer's last walk found in each file whose final text is the text it walked, by path, so lint need not walk it again. */
        var reusableFindings: [String: [String: [FindingRecord]]] = [:]
        var nothingRewritten = false

        if !options.runFix {
            try writer.write(PhaseRecord(name: .fix, outcome: .skipped, detail: "not requested"))
        }
        else if filesThatDoNotParse > 0 {
            /* Nothing is rewritten in a run where any file does not parse: the bail below says why, and the fix line says nothing moved. */
            try writer.write(
                fixRecord(
                    scope: scope,
                    considered: parsed.files.count,
                    rewritten: [],
                    fixesApplied: 0,
                    refusals: [:],
                    reformatted: 0,
                    notFormatted: ["a file in scope does not parse, so nothing was rewritten": parsed.files.count],
                )
            )
            try writer.write(
                PhaseRecord(name: .fix, outcome: .ran, elapsedMilliseconds: Self.milliseconds(since: fixStart))
            )
        }
        else {
            /*
             Fixers first, then the formatter over the fixed text, so the formatter has the last word on layout
             and a fixer's output is always formatted. Under `--no-fix` nothing is fixed: the fixable findings
             surface in lint, and the formatter is asked only what it would change.
             */
            var fixesApplied = 0
            var refusals: [String: Int] = [:]
            var toFormat = parsed.files
            var fixerFindings: [String: [String: [FindingRecord]]] = [:]
            var fixedByFile: [String: [String: Int]] = [:]
            if options.mutate {
                let fixer = FileFixer(configuration: configuration, rules: fileRules, maximumPasses: options.fixPasses)
                /* Each file is fixed on its own, so the files are fixed side by side; results are gathered back in input order. */
                let files = parsed.files
                let results = await withTaskGroup(of: (Int, FileFixer.Result).self) { group in
                    for (index, file) in files.enumerated() {
                        group.addTask { (index, fixer.fix(file)) }
                    }
                    var collected = [FileFixer.Result?](repeating: nil, count: files.count)
                    for await (index, result) in group {
                        collected[index] = result
                    }
                    return collected
                }
                toFormat = zip(files, results).map { file, result in
                    guard let result else { return file }
                    fixesApplied += result.applied
                    if !result.appliedByRule.isEmpty {
                        fixedByFile[file.url.path] = result.appliedByRule
                    }
                    refusals.merge(result.refusalsByReason, uniquingKeysWith: +)
                    if let found = result.findingsOfFinalText {
                        fixerFindings[file.url.path] = found
                    }
                    return result.file
                }
            }
            let originals = Dictionary(
                parsed.files.map { ($0.url.path, $0.source) },
                uniquingKeysWith: { first, _ in first },
            )
            let formatting = await FormatPhase().run(toFormat)
            var rewritten: [FileSet.OwnedFile] = []
            var changedFiles: [FixRecord.ChangedFile] = []
            var reformatted = 0
            var notFormatted: [String: Int] = [:]
            for (file, outcome) in formatting.outcomes {
                var final = file.source
                var formattedHere = false
                switch outcome {
                    case .unchanged:
                        break
                    case let .changed(formatted):
                        if options.mutate {
                            final = formatted
                            reformatted += 1
                            formattedHere = true
                        }
                        else {
                            /* `--no-fix` writes nothing and reports what it would have changed, one finding per file, at the first line that moves. */
                            let line = FormatPhase.firstDifferingLine(file.source, formatted)
                            try writer.write(
                                FindingRecord(
                                    source: .format,
                                    file: file.url.path,
                                    line: line,
                                    column: 1,
                                    severity: .error,
                                    rule: "cohere-swift/consistency-require-formatting",
                                    messageId: "notFormatted",
                                    message: "not formatted the house way; a run without --no-fix rewrites it",
                                )
                            )
                        }
                    case let .failed(reason):
                        notFormatted["the formatter failed: \(reason)", default: 0] += 1
                        complete = false
                }
                /* The fixer walked exactly this text when the formatter left it as the fixer did, so its findings are lint's. */
                if final == file.source, let found = fixerFindings[file.url.path] {
                    reusableFindings[file.url.path] = found
                }
                /* One write per file, of the fixed and formatted text, and only when it differs from what was read. */
                if options.mutate, final != originals[file.url.path] {
                    try final.write(to: file.url, atomically: true, encoding: .utf8)
                    changedFiles.append(
                        FixRecord.ChangedFile(
                            file: file.url.path,
                            fixedBy: fixedByFile[file.url.path] ?? [:],
                            formatted: formattedHere,
                        )
                    )
                    rewritten.append(
                        FileSet.OwnedFile(
                            url: file.url,
                            targetName: file.targetName,
                            targetKind: file.targetKind,
                            packageRoot: file.packageRoot,
                        )
                    )
                }
            }
            /* Rewritten files are parsed again, so lint reads the text that is on disk now rather than the text that was. */
            if !rewritten.isEmpty {
                let reparsed = await SourceParser().parse(rewritten)
                let replacements = Dictionary(
                    reparsed.files.map { ($0.url.path, $0) },
                    uniquingKeysWith: { first, _ in first },
                )
                parsed.files = parsed.files.map { replacements[$0.url.path] ?? $0 }
            }
            nothingRewritten = rewritten.isEmpty
            try writer.write(
                fixRecord(
                    scope: scope,
                    considered: parsed.files.count,
                    rewritten: changedFiles,
                    fixesApplied: fixesApplied,
                    refusals: refusals,
                    reformatted: reformatted,
                    notFormatted: notFormatted,
                )
            )
            try writer.write(
                PhaseRecord(name: .fix, outcome: .ran, elapsedMilliseconds: Self.milliseconds(since: fixStart))
            )
        }

        if filesThatDoNotParse > 0 {
            let reason =
                "parsing bailed: \(filesThatDoNotParse) files do not parse, and checking a tree the parser could not read reports nonsense"
            try writer.write(
                PhaseRecord(
                    name: .types,
                    outcome: options.runTypes ? .notReached : .skipped,
                    detail: options.runTypes ? reason : "not requested",
                )
            )
            try writer.write(
                PhaseRecord(
                    name: .lint,
                    outcome: options.runLint ? .notReached : .skipped,
                    detail: options.runLint ? reason : "not requested",
                )
            )
            try writer.write(unusedPhase(notReached: reason))
            return try writer.finish(complete: false)
        }

        var typeErrors = 0
        if !options.runTypes {
            try writer.write(PhaseRecord(name: .types, outcome: .skipped, detail: "not requested"))
        }
        else {
            /* The whole package, whatever the scope: a change in one file changes what the rest of its module means. */
            var typesPhase = TypesPhase(
                package: package,
                files: fileSet.owned,
                scratchPath: Self.scratchPath(for: root),
                resolutionAllowed: !options.noFix,
                toolchain: toolchain,
                runner: runner,
            )
            typesPhase.parsed = Dictionary(
                parsed.files.map { ($0.url.path, $0.tree) },
                uniquingKeysWith: { first, _ in first },
            )
            let types = try typesPhase.run()
            for finding in types.findings {
                try writer.write(finding)
            }
            try writer.write(types.record)
            try writer.write(
                PhaseRecord(name: .types, outcome: .ran, elapsedMilliseconds: types.record.elapsedMilliseconds)
            )
            if !types.record.filesWithoutRecord.isEmpty {
                complete = false
            }
            typeErrors = types.findings.filter { $0.severity == .error }.count
        }

        if typeErrors > 0 && options.runLint {
            try writer.write(
                PhaseRecord(
                    name: .lint,
                    outcome: .notReached,
                    detail: "types bailed: \(typeErrors) type errors — lint findings against wrong semantics are noise",
                )
            )
            try writer.write(
                unusedPhase(
                    notReached:
                        "types bailed: \(typeErrors) type errors, and the index of a package that does not compile is not the package"
                )
            )
            return try writer.finish(complete: false)
        }

        if !options.runLint {
            try writer.write(PhaseRecord(name: .lint, outcome: .skipped, detail: "not requested"))
        }
        else {
            let lintStart = Date()
            var linter = Linter(configuration: configuration, fileRules: fileRules)
            linter.typedRules = RuleRegistry.typedRules
            let typedCandidates = parsed.files.filter { file in
                RuleRegistry.typedRules.contains { configuration.severity(of: $0.name) != .off && $0.applies(to: file) }
            }
            var symbolProvider = SymbolProvider(
                scratchPaths: Self.symbolScratchPaths(package: package, root: root),
                runner: runner,
            )
            symbolProvider.ownedModules = Self.ownedModules(of: package)
            linter.symbols = symbolProvider.symbols(for: typedCandidates)
            let lint = await linter.run(
                package: package,
                manifests: await manifests(of: package),
                files: parsed.files,
                reusable: reusableFindings,
            )
            for finding in lint.findings {
                try writer.write(finding)
            }
            /*
             Reported as reused only when every file's findings came from the fix phase and nothing was rewritten,
             because that is what the front door's line says. A run that reused some files and walked others
             still saves the walk, and says it ran.
             */
            let reusedEverything =
                nothingRewritten && typedCandidates.isEmpty && !parsed.files.isEmpty
                && parsed.files.allSatisfy { reusableFindings[$0.url.path] != nil }
            var record = lint.record
            record.elapsedMilliseconds = Self.milliseconds(since: lintStart)
            record.reusedFrom = reusedEverything ? "fix" : ""
            try writer.write(record)
            try writer.write(
                reusedEverything
                    ? PhaseRecord(
                        name: .lint,
                        outcome: .reused,
                        elapsedMilliseconds: record.elapsedMilliseconds,
                        detail: "the fix phase's walk (nothing was rewritten)",
                    )
                    : PhaseRecord(name: .lint, outcome: .ran, elapsedMilliseconds: record.elapsedMilliseconds)
            )
            if !record.crashes.isEmpty {
                complete = false
            }
        }

        if options.unused {
            try await runUnused(package: package, root: root, fileSet: fileSet, parsed: parsed.files)
        }
        else {
            try writer.write(PhaseRecord(name: .unused, outcome: .skipped, detail: Self.unusedNotRequested))
        }
        return try writer.finish(complete: complete)
    }

    static let unusedNotRequested = "not requested — this is a report, ask for it with --unused"

    /* The unused phase of a run cut off before it: not reached when it was asked for, not requested otherwise. */
    private func unusedPhase(notReached reason: String) -> PhaseRecord {
        options.unused
            ? PhaseRecord(name: .unused, outcome: .notReached, detail: reason)
            : PhaseRecord(name: .unused, outcome: .skipped, detail: Self.unusedNotRequested)
    }

    /*
     The unused report, over the whole package whatever the scope, since an import or a private declaration is
     used or not by its file alone but the index that answers is the whole build's. It reads the index the types phase just refreshed,
     or the last build's, and a file that index does not describe as it stands is counted as not checked
     rather than guessed at. Its records are a report, not findings: the writer does not count them, so they
     never fail a run, the way the TypeScript unused report never does.
     */
    private func runUnused(package: PackageModel, root: URL, fileSet: FileSet, parsed: [ParsedFile]) async throws {
        let start = Date()
        let stores: [IndexStore]
        let demangler: SwiftDemangler
        do {
            let library = try IndexStore.toolchainLibraryPath(runner: runner)
            stores = Self.symbolScratchPaths(package: package, root: root).compactMap {
                try? IndexStore(libraryPath: library, storePath: IndexStore.storePath(scratchPath: $0))
            }
            demangler = try SwiftDemangler.shared(runner: runner)
        }
        catch {
            try writer.write(
                PhaseRecord(
                    name: .unused,
                    outcome: .notReached,
                    detail: "the index store or the demangler could not be loaded: \(error)",
                )
            )
            return
        }
        guard !stores.isEmpty else {
            try writer.write(
                PhaseRecord(
                    name: .unused,
                    outcome: .notReached,
                    detail:
                        "no build has written an index store for this package yet; a run with the types phase writes one",
                )
            )
            return
        }
        /* Every file of ours, the out-of-scope ones parsed here: an `@_exported import` anywhere decides what a module re-exports. */
        let parsedPaths = Set(parsed.map(\.url.path))
        let missing = fileSet.owned.filter { !parsedPaths.contains($0.url.path) }
        let files = missing.isEmpty ? parsed : parsed + (await SourceParser().parse(missing)).files
        let imports = UnusedImports(stores: stores, demangler: demangler).run(files: files)
        try writeReport(
            rule: UnusedImports.ruleName,
            findings: imports.findings,
            filesChecked: imports.filesChecked,
            filesUnchecked: imports.filesUnchecked,
            checked: imports.importsChecked,
            skipped: imports.importsExported > 0
                ? ["re-exported with @_exported, which is API": imports.importsExported] : [:],
            since: start,
        )
        let declarationsStart = Date()
        let declarations = UnusedDeclarations(stores: stores, package: package).run(files: files)
        try writeReport(
            rule: UnusedDeclarations.ruleName,
            findings: declarations.findings,
            filesChecked: declarations.filesChecked,
            filesUnchecked: declarations.filesUnchecked,
            checked: declarations.declarationsChecked,
            skipped: declarations.skipped,
            since: declarationsStart,
        )
        try writer.write(
            PhaseRecord(name: .unused, outcome: .ran, elapsedMilliseconds: Self.milliseconds(since: start))
        )
    }

    /* One rule's part of the unused report: its records in file order, then the coverage record that counts them. */
    private func writeReport(
        rule: String,
        findings: [(finding: FindingRecord, subject: String)],
        filesChecked: Int,
        filesUnchecked: [String: String],
        checked: Int,
        skipped: [String: Int],
        since start: Date,
    ) throws {
        let ordered = findings.sorted {
            ($0.finding.file, $0.finding.line, $0.finding.column) < (
                $1.finding.file, $1.finding.line, $1.finding.column,
            )
        }
        for found in ordered {
            try writer.write(UnusedRecord(found.finding, subject: found.subject))
        }
        var notChecked: [String: Int] = [:]
        for reason in filesUnchecked.values {
            notChecked[reason, default: 0] += 1
        }
        try writer.write(
            UnusedCoverageRecord(
                rule: rule,
                filesChecked: filesChecked,
                filesNotChecked: notChecked,
                checked: checked,
                skipped: skipped,
                found: ordered.count,
                elapsedMilliseconds: Self.milliseconds(since: start),
            )
        )
    }

    private func projectRecord(
        package: PackageModel,
        fileSet: FileSet,
        scope: FileScope,
        elapsed: Int,
    ) -> ProjectRecord {
        let counts = Dictionary(grouping: fileSet.owned, by: \.targetName).mapValues(\.count)
        return ProjectRecord(
            root: package.root.path,
            package: package.name,
            elapsedMilliseconds: elapsed,
            filesInPackage: fileSet.filesInPackage.count,
            filesOurs: fileSet.owned.count,
            filesInScope: scope.files.count,
            scopeDescription: scope.description,
            /* A local package's targets are named `<package>/<target>`, so a reader can tell Presence's own targets from VRMKit's. */
            targets: package.allPackages.flatMap { member in
                member.targets.map { target in
                    let name = member.root == package.root ? target.name : "\(member.name)/\(target.name)"
                    return ProjectRecord.Target(
                        name: name,
                        kind: target.kind,
                        files: counts[target.name] ?? 0,
                        languageMode: target.languageMode,
                    )
                }
            },
            excluded: fileSet.excluded,
        )
    }

    /* The fix line's numbers. Rewrites, fixes and reformats are separate counts, so a run where only the formatter moved reads differently from one where a fixer did. */
    private func fixRecord(
        scope: FileScope,
        considered: Int,
        rewritten: [FixRecord.ChangedFile],
        fixesApplied: Int,
        refusals: [String: Int],
        reformatted: Int,
        notFormatted: [String: Int],
    ) -> FixRecord {
        FixRecord(
            filesConsidered: considered,
            filesRewritten: rewritten.count,
            fixesApplied: fixesApplied,
            fixesRefused: refusals.values.reduce(0, +),
            refusalsByReason: refusals,
            filesReformatted: reformatted,
            filesNotFormatted: notFormatted.values.reduce(0, +),
            notFormattedReasons: notFormatted,
            formatScope: scope.everything ? "every file in the package" : scope.description,
            changedFiles: rewritten,
        )
    }

    /*
     The top of the git repository holding the package, or the package itself outside git: the highest
     directory a leftover `.swift-format` is looked for in. Found by the `.git` on disk, the same walk the file
     set uses.
     */
    private func repositoryRoot(of root: URL) -> URL {
        IgnoreRules.Repository.containing(root)?.root ?? root
    }

    /* The modules of every target this package owns, vendored packages left out, spelled as the compiler spells a module: non-identifier characters made `_`. */
    static func ownedModules(of package: PackageModel) -> Set<String> {
        let owned = package.allPackages.filter { $0.root == package.root || !package.isVendored($0) }
        return Set(
            owned.flatMap(\.targets).map { target in
                String(target.name.map { $0.isLetter || $0.isNumber || $0 == "_" ? $0 : "_" })
            }
        )
    }

    /* Every scratch a build of this package writes an index store into: the root's, and each local package's that the types phase builds for its tests. */
    static func symbolScratchPaths(package: PackageModel, root: URL) -> [URL] {
        let scratch = Self.scratchPath(for: root)
        return [scratch]
            + package.localPackages.map {
                scratch.appendingPathComponent("local/\($0.root.lastPathComponent)", isDirectory: true)
            }
    }

    /* Each owned package's manifest, parsed, so package rules can point at the line that would fix them. A manifest that cannot be read is absent, and its findings point at line 1. */
    private func manifests(of package: PackageModel) async -> [String: ParsedFile] {
        let owned = package.allPackages.filter { $0.root == package.root || !package.isVendored($0) }
        let files = owned.map {
            FileSet.OwnedFile(
                url: $0.root.appendingPathComponent("Package.swift"),
                targetName: $0.root.path,
                targetKind: "manifest",
                packageRoot: $0.root,
            )
        }
        let result = await SourceParser().parse(files)
        return Dictionary(result.files.map { ($0.targetName, $0) }, uniquingKeysWith: { first, _ in first })
    }

    /*
     Where SwiftPM writes when the engine runs it: the project's own `.cache/cohere/swift`, beside the dispatcher's
     table (#wb3xz1h), so deleting that one folder is always a correct answer and a worktree gets its own. It is
     never `.build`, so a run never takes the developer's lock, and SwiftPM and the engine's walks skip hidden
     directories, so nothing here is ever read as source. `--no-fix` writes no source file; this cache is cohere's.
     */
    public static func scratchPath(for root: URL) -> URL {
        root.appendingPathComponent(".cache/cohere/swift", isDirectory: true)
    }

    static func milliseconds(since start: Date) -> Int {
        Int((Date().timeIntervalSince(start) * 1000).rounded())
    }
}
