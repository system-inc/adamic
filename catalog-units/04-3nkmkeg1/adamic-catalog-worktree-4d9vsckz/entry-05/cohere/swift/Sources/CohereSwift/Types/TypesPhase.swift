import Foundation
import SwiftParser
import SwiftSyntax
import Synchronization

/*
 The compiler's own errors and warnings for every file we own, read from what the build recorded rather
 than from what it printed.

 The build runs into the engine's cache (`--scratch-path`), never the project's `.build`. It builds test
 targets too, because their files are ours and a test that does not compile is a finding like any other.
 Then every serialized diagnostics file in each owned target's build directories is read, each diagnostic
 is placed by its own file path, and duplicates are dropped: an executable target is compiled twice (once
 normally, once testable for its tests), and both copies report the same warning.

 A file is covered only when its own `.dia` exists and is newer than the source. A missing or older
 record means the compiler did not vouch for the file as it stands, usually because the build stopped
 before reaching it. That file is listed by name and the run is incomplete, never quietly clean.

 Measured on Swift 6.4's build system: records live at
 `out/Intermediates.noindex/<Package>.build/Debug/<Target>-{p,t}[...].build/Objects-normal/<arch>/<File>.dia`.
 A layout the phase cannot find any record in is a loud failure, because the build system moved its
 output once already (the index store went to `out/v5` in 6.4) and may move this too.
 */
struct TypesPhase {
    /* What the phase found and what it could not vouch for. */
    struct Result {
        var findings: [FindingRecord]
        var record: TypesRecord
        /* True when any compiler error was found: lint is then not reached. Warnings do not bail. */
        var hasErrors: Bool
    }

    let package: PackageModel
    let files: [FileSet.OwnedFile]
    let scratchPath: URL
    let resolutionAllowed: Bool
    /* The toolchain `swift --version` named; an unknown one never reuses a build, so a toolchain change can never be missed. */
    let toolchain: String
    let runner: ProcessRunner
    /* Trees the pipeline already parsed, by path, so fingerprinting a file does not parse it twice. Any file missing is parsed here. */
    var parsed: [String: SourceFileSyntax] = [:]

    /*
     The packages to build: the root, plus each local package that holds test files of ours. Building the root
     compiles a local package's libraries but never its tests, because a dependency's test targets are not part
     of the root's graph. Measured on Presence: without these builds, RigSolve's, CaptureWire's and VRMKit's 26
     test files had no compiler record at all.
     */
    private var builds: [(root: URL, scratchPath: URL)] {
        let testTargetNames = Set(files.filter { $0.targetKind == "test" }.map(\.targetName))
        let locals = package.localPackages.filter { local in
            local.targets.contains { $0.kind == "test" && testTargetNames.contains($0.name) }
        }
        return [(package.root, scratchPath)]
            + locals.map {
                ($0.root, scratchPath.appendingPathComponent("local/\($0.root.lastPathComponent)", isDirectory: true))
            }
    }

    private func arguments(for build: (root: URL, scratchPath: URL), buildSystem: [String]) -> [String] {
        var arguments = ["build", "--build-tests", "--scratch-path", build.scratchPath.path] + buildSystem
        if !resolutionAllowed {
            arguments.append("--disable-automatic-resolution")
        }
        return arguments
    }

    /* The package roots and every path dependency, which together hold every file a build of them reads. */
    private var inputRoots: [URL] {
        let roots = package.allPackages.flatMap { [$0.root] + $0.pathDependencyRoots }
        var seen = Set<String>()
        return roots.filter { root in
            !roots.contains { other in PackageModel.isInside(root, other) }
                && seen.insert(root.resolvingSymlinksInPath().path).inserted
        }
    }

    func run() throws -> Result {
        let start = Date()
        let snapshot =
            toolchain.hasPrefix("unknown")
            ? nil
            : BuildInputSnapshot.take(
                roots: inputRoots,
                resolved: package.root.appendingPathComponent("Package.resolved"),
                toolchain: toolchain,
                arguments: builds.flatMap { arguments(for: $0, buildSystem: Self.swiftBuildArguments) },
            )

        /*
         Nothing a build reads changed since the last build that succeeded, so its records are this build's
         records. Read them, and keep them only if every file of ours has one newer than its source: anything
         less (a scratch someone cleaned, a record gone) means building after all.
         */
        let stored = BuildInputSnapshot.stored(scratchPath: scratchPath)
        if let snapshot, let stored, stored != snapshot,
            let checked = checkBodiesOnly(snapshot: snapshot, stored: stored, start: start)
        {
            return checked
        }
        if let snapshot, stored == snapshot {
            do {
                let directories = try builds.flatMap { try buildDirectories(under: $0.scratchPath) }
                if !directories.isEmpty {
                    let reused = try read(
                        targetDirectories: directories,
                        failedBuilds: [],
                        start: start,
                        build:
                            "no input changed since the last successful build, so its compiler records were read without building",
                    )
                    if reused.record.filesWithoutRecord.isEmpty {
                        return reused
                    }
                }
            }
            catch {
                /* The records could not be read as they stood: build, which rewrites them. */
            }
        }

        /* Each file's interface as this build compiles it, so the next run can tell whether an edit stayed inside bodies. */
        let interfaces = snapshot == nil ? [:] : interfaceFingerprints()

        let buildSystem = try buildSystemArguments()
        var failedBuilds: [ProcessRunner.Result] = []
        var targetDirectories: [(String, URL)] = []
        for build in builds {
            let result = try runner.run("swift", arguments(for: build, buildSystem: buildSystem), in: build.root)
            if build.scratchPath != scratchPath {
                ScratchPrune.markUsed(build.scratchPath)
            }
            if !result.succeeded {
                failedBuilds.append(result)
            }
            let directories = try buildDirectories(under: build.scratchPath)
            /*
             A build that failed and left nothing failed before the compiler ran (dependency resolution, the
             manifest), which is a different fault from a build that succeeded into a layout we cannot read.
             Measured: `--no-fix` on a copy of ahraos-macos without its gitignored Package.resolved.
             */
            guard !directories.isEmpty else {
                if !result.succeeded {
                    throw TypesFailure(
                        description:
                            "swift build of \(build.root.path) failed before compiling anything, so the compiler never ran: \(Self.tail(of: result))"
                    )
                }
                throw TypesFailure(
                    description:
                        "the build of \(build.root.path) left no target directories under \(intermediates(of: build.scratchPath).path), so this engine does not recognise the build system's layout and cannot read what the compiler said (toolchain: \(toolchain))\n\(Self.tail(of: result))"
                )
            }
            targetDirectories.append(contentsOf: directories)
        }
        if var snapshot, failedBuilds.isEmpty {
            snapshot.interfaces = interfaces
            snapshot.store(scratchPath: scratchPath)
        }
        else {
            BuildInputSnapshot.forget(scratchPath: scratchPath)
        }
        var build = "swift build --build-tests of \(builds.count) packages, scratch \(scratchPath.path)"
        /* Only after every build succeeded, so no build of this run is still writing into the scratch. */
        if failedBuilds.isEmpty {
            let pruned = pruneScratch()
            if !pruned.sentence.isEmpty {
                build += "; \(pruned.sentence)"
            }
        }
        return try read(targetDirectories: targetDirectories, failedBuilds: failedBuilds, start: start, build: build)
    }

    /* The scratch prune over every scratch this run built into (`ScratchPrune`). A store that cannot be opened is left as it is. */
    private func pruneScratch() -> ScratchPrune.Outcome {
        guard let library = try? IndexStore.toolchainLibraryPath(runner: runner) else { return ScratchPrune.Outcome() }
        let stores = builds.compactMap { build -> (store: IndexStore, storePath: URL)? in
            let storePath = IndexStore.storePath(scratchPath: build.scratchPath)
            return (try? IndexStore(libraryPath: library, storePath: storePath)).map { ($0, storePath) }
        }
        return ScratchPrune.run(
            stores: stores,
            scratchPath: scratchPath,
            listedLocalPackages: Set(package.localPackages.map(\.root.lastPathComponent)),
        )
    }

    /* The words the types record uses for a run that checked its edited files in process, so a reader can tell it from a build. */
    static let checkedInProcessWords = "changed inside function bodies only, so sourcekitd checked them in this process"

    /*
     The edited files checked in sourcekitd, and every other file read from the last build's records, when the
     edits cannot have changed what the compiler says about any other file. Nil whenever that is not certain,
     which means building: anything but owned Swift files changed, a file's interface changed, the last build
     left no interface for it, its compile arguments are not in the build's record, or sourcekitd could not
     answer. Measured on Presence before this path: a one-file edit cost a 19 to 33s incremental build.
     */
    private func checkBodiesOnly(snapshot: BuildInputSnapshot, stored: BuildInputSnapshot, start: Date) -> Result? {
        guard let changed = snapshot.filesChanged(since: stored), !changed.isEmpty else { return nil }
        let owned = Dictionary(
            files.map { ($0.url.resolvingSymlinksInPath().path, $0) },
            uniquingKeysWith: { first, _ in first },
        )
        var edited: [(path: String, file: FileSet.OwnedFile)] = []
        for path in changed {
            let resolved = URL(fileURLWithPath: path).resolvingSymlinksInPath().path
            guard let file = owned[resolved], let before = stored.interfaces[resolved], let tree = tree(of: file),
                InterfaceFingerprint.of(tree) == before
            else { return nil }
            edited.append((resolved, file))
        }
        do {
            var tables = [try CompileCommands.read(scratchPath: scratchPath)]
            for build in builds.dropFirst() where FileManager.default.fileExists(atPath: build.scratchPath.path) {
                tables.append(try CompileCommands.read(scratchPath: build.scratchPath))
            }
            let session = try Sourcekitd.shared(runner: runner)
            var answered: [String: [FindingRecord]] = [:]
            for (path, file) in edited.sorted(by: { $0.path < $1.path }) {
                guard let command = tables.lazy.compactMap({ $0.command(for: file.url) }).first else { return nil }
                answered[path] = try session.diagnostics(file: path, arguments: command.arguments).compactMap {
                    diagnostic in
                    guard let severity = diagnostic.findingSeverity,
                        URL(fileURLWithPath: diagnostic.file).resolvingSymlinksInPath().path == path
                    else { return nil }
                    return FindingRecord(
                        source: .compiler,
                        file: path,
                        line: max(diagnostic.line, 1),
                        column: max(diagnostic.column, 1),
                        severity: severity,
                        rule: TypesOracle.group(of: diagnostic),
                        messageId: "",
                        message: diagnostic.message,
                    )
                }
            }
            let directories = try builds.flatMap { try buildDirectories(under: $0.scratchPath) }
            guard !directories.isEmpty else { return nil }
            let result = try read(
                targetDirectories: directories,
                failedBuilds: [],
                start: start,
                build:
                    "\(edited.count == 1 ? "1 file" : "\(edited.count) files") \(Self.checkedInProcessWords), and the rest were read from the last successful build's compiler records",
                answeredInProcess: answered,
            )
            return result.record.filesWithoutRecord.isEmpty ? result : nil
        }
        catch {
            /* Anything this path could not do, a build does. */
            return nil
        }
    }

    /* Every owned file's interface, by resolved path, fingerprinted side by side. A file that cannot be read has none, so an edit to it always builds. */
    private func interfaceFingerprints() -> [String: String] {
        let owned = files
        let byPath = Mutex<[String: String]>([:])
        DispatchQueue.concurrentPerform(iterations: owned.count) { index in
            guard let fingerprint = tree(of: owned[index]).map(InterfaceFingerprint.of) else { return }
            let path = owned[index].url.resolvingSymlinksInPath().path
            byPath.withLock { $0[path] = fingerprint }
        }
        return byPath.withLock { $0 }
    }

    private func tree(of file: FileSet.OwnedFile) -> SourceFileSyntax? {
        if let tree = parsed[file.url.path] {
            return tree
        }
        guard let source = try? String(contentsOf: file.url, encoding: .utf8) else { return nil }
        return Parser.parse(source: source)
    }

    /* Every owned file's diagnostics from the records under these target directories, and which files have none. */
    private func read(
        targetDirectories: [(String, URL)],
        failedBuilds: [ProcessRunner.Result],
        start: Date,
        build: String,
        answeredInProcess: [String: [FindingRecord]] = [:],
    ) throws -> Result {
        let reader = try SerializedDiagnosticsReader(
            libraryPath: SerializedDiagnosticsReader.toolchainLibraryPath(runner: runner)
        )

        let ours = Dictionary(
            files.map { ($0.url.resolvingSymlinksInPath().path, $0) },
            uniquingKeysWith: { first, _ in first },
        )
        var seen = Set<String>()
        var findings: [FindingRecord] = []
        var recordFor: [String: Date] = [:]

        for (targetName, directory) in targetDirectories {
            for record in try diaFiles(in: directory) {
                let modified =
                    try record.resourceValues(forKeys: [.contentModificationDateKey]).contentModificationDate
                    ?? .distantPast
                /* Keyed by target as well as name: two targets may each hold a `Utilities.swift`, and one must not vouch for the other. */
                let key = "\(targetName)/\(record.deletingPathExtension().lastPathComponent)"
                /*
                 A record libclang cannot read (seen once on Presence, a `.dia` left half-written by a build that
                 was interrupted) vouches for nothing: its file is left without a record, which marks the run
                 incomplete and makes a reused build rebuild, rather than one bad file ending the whole run. The
                 record is removed, since it sits in the engine's own scratch: SwiftPM recompiles a file whose
                 output is missing, so the build that follows writes a good one.
                 */
                let diagnostics: [SerializedDiagnosticsReader.Diagnostic]
                do {
                    diagnostics = try reader.read(record)
                }
                catch {
                    do {
                        try FileManager.default.removeItem(at: record)
                    }
                    catch {
                        /* Left in place, the file stays without a record and the run says so; only the repair is lost. */
                    }
                    continue
                }
                recordFor[key] = max(recordFor[key] ?? .distantPast, modified)
                for diagnostic in diagnostics {
                    guard let severity = diagnostic.severity else { continue }
                    let file = URL(fileURLWithPath: diagnostic.file).resolvingSymlinksInPath().path
                    /* A file sourcekitd just checked is answered by that check; its record describes text that has changed since. */
                    guard ours[file] != nil, answeredInProcess[file] == nil else { continue }
                    let identity = "\(file):\(diagnostic.line):\(diagnostic.column):\(diagnostic.message)"
                    guard seen.insert(identity).inserted else { continue }
                    findings.append(
                        FindingRecord(
                            source: .compiler,
                            file: file,
                            line: max(diagnostic.line, 1),
                            column: max(diagnostic.column, 1),
                            severity: severity,
                            rule: diagnostic.group,
                            messageId: "",
                            message: diagnostic.message,
                        )
                    )
                }
            }
        }

        findings.append(contentsOf: answeredInProcess.values.joined())

        /* Coverage: each file's own record, by stem, newer than the source as it stands, or a check in this process. */
        var withoutRecord: [String] = []
        for file in files where answeredInProcess[file.url.resolvingSymlinksInPath().path] == nil {
            let key = "\(file.targetName)/\(file.url.deletingPathExtension().lastPathComponent)"
            let sourceModified =
                try file.url.resourceValues(forKeys: [.contentModificationDateKey]).contentModificationDate
                ?? .distantFuture
            guard let recorded = recordFor[key], recorded >= sourceModified else {
                withoutRecord.append(file.url.path)
                continue
            }
        }

        /* A failed build with no error in any file of ours failed somewhere no file carries: linking, a dependency, the manifest. Said, never swallowed. */
        if let failed = failedBuilds.first, !findings.contains(where: { $0.severity == .error }) {
            findings.append(
                FindingRecord(
                    source: .compiler,
                    file: package.root.appendingPathComponent("Package.swift").path,
                    line: 1,
                    column: 1,
                    severity: .error,
                    rule: "",
                    messageId: "buildFailed",
                    message:
                        "swift build failed outside any file this run checks, so the package does not build: \(Self.tail(of: failed))",
                )
            )
        }

        findings.sort { ($0.file, $0.line, $0.column) < ($1.file, $1.line, $1.column) }
        let record = TypesRecord(
            diagnostics: findings.count,
            files: files.count,
            elapsedMilliseconds: Pipeline.milliseconds(since: start),
            filesWithoutRecord: withoutRecord.sorted(),
            build: build,
        )
        return Result(findings: findings, record: record, hasErrors: findings.contains { $0.severity == .error })
    }

    /*
     The build system is named rather than left to the toolchain's default, because every record this phase
     reads sits where swiftbuild puts it. Swift 6.4 builds with swiftbuild by default; Swift 6.3 (Xcode 26) still
     defaults to the native build system, whose layout differs, while offering swiftbuild as an option. A
     toolchain that does not offer it fails here by name, never as a types phase that quietly read nothing.
     Asked only when the phase builds; a run that reuses the last build's records never pays for it.
     */
    private func buildSystemArguments() throws -> [String] {
        let help = try runner.run("swift", ["build", "--help"], in: package.root)
        guard
            let arguments = Self.buildSystemArguments(
                help: String(decoding: help.standardOutput, as: UTF8.self) + help.standardError
            )
        else {
            throw TypesFailure(
                description:
                    "this toolchain's swift build offers no swiftbuild build system, and the types phase reads only the records swiftbuild writes (toolchain: \(toolchain)). Use a toolchain that offers `swift build --build-system swiftbuild`."
            )
        }
        return arguments
    }

    static let swiftBuildArguments = ["--build-system", "swiftbuild"]

    /*
     `--build-system swiftbuild` when `swift build --help` offers swiftbuild, else nil. Help lists the build
     systems either one per line (6.4) or inline as `(values: ...)`, so the word is looked for, not a layout.
     */
    static func buildSystemArguments(help: String) -> [String]? {
        help.contains("--build-system") && help.contains("swiftbuild") ? swiftBuildArguments : nil
    }

    /* The build could not be read at all. */
    struct TypesFailure: Error, CustomStringConvertible {
        var description: String
    }

    private func intermediates(of scratchPath: URL) -> URL {
        scratchPath.appendingPathComponent("out/Intermediates.noindex", isDirectory: true)
    }

    /* Every build directory belonging to a target we own files in, with that target's name: `<Target>-p.build`, `<Target>-t.build`, and the testable variants. */
    private func buildDirectories(under scratchPath: URL) throws -> [(String, URL)] {
        let targetNames = Set(files.map(\.targetName))
        let manager = FileManager.default
        let intermediatesDirectory = intermediates(of: scratchPath)
        guard manager.fileExists(atPath: intermediatesDirectory.path) else { return [] }
        var directories: [(String, URL)] = []
        for packageDirectory in try manager.contentsOfDirectory(
            at: intermediatesDirectory,
            includingPropertiesForKeys: nil,
        ) {
            let configuration = packageDirectory.appendingPathComponent("Debug", isDirectory: true)
            guard manager.fileExists(atPath: configuration.path) else { continue }
            for targetDirectory in try manager.contentsOfDirectory(at: configuration, includingPropertiesForKeys: nil) {
                let name = targetDirectory.lastPathComponent
                guard name.hasSuffix(".build"), !name.contains("-product-") else { continue }
                /*
                 Matched by the directory's exact stem, never by prefix: `cohere-swift-` is a prefix of
                 `cohere-swift-parity-p.build`, and `AhraOsServices-` of nothing it should not be only by luck.
                 */
                let stem = Self.directoryStem(name)
                if let target = targetNames.contains(stem) ? stem : productTargets[stem], targetNames.contains(target) {
                    directories.append((target, targetDirectory))
                }
            }
        }
        return directories
    }

    /* Every single-target product of every package this phase builds, to its target. */
    private var productTargets: [String: String] {
        package.allPackages.reduce(into: [:]) { merged, member in
            merged.merge(member.productTargets) { first, _ in first }
        }
    }

    /*
     The target or product a build directory belongs to: `AhraOs-p.build` and `CohereSwift-t.build` drop their
     last dash-separated part, and a testable variant, `AhraOsServices--36C56AD10F55DF70-testable-t.build`,
     keeps what comes before its double dash.
     */
    static func directoryStem(_ name: String) -> String {
        let base = String(name.dropLast(".build".count))
        if let variant = base.range(of: "--") {
            return String(base[..<variant.lowerBound])
        }
        guard let lastDash = base.lastIndex(of: "-") else { return base }
        return String(base[..<lastDash])
    }

    private func diaFiles(in targetDirectory: URL) throws -> [URL] {
        let objects = targetDirectory.appendingPathComponent("Objects-normal", isDirectory: true)
        guard FileManager.default.fileExists(atPath: objects.path) else { return [] }
        var records: [URL] = []
        for architecture in try FileManager.default.contentsOfDirectory(at: objects, includingPropertiesForKeys: nil) {
            records.append(
                contentsOf: try FileManager.default.contentsOfDirectory(
                    at: architecture,
                    includingPropertiesForKeys: nil,
                ).filter { $0.pathExtension == "dia" }
            )
        }
        return records
    }

    /* The last lines of the build's own output, colours stripped, for a failure no file carries. */
    static func tail(of build: ProcessRunner.Result) -> String {
        let text = String(decoding: build.standardOutput, as: UTF8.self) + build.standardError
        let plain = text.replacingOccurrences(
            of: #"\u{1B}\[[0-9;]*m|\u{1B}\]8;;[^\u{1B}]*\u{1B}\\"#,
            with: "",
            options: .regularExpression,
        )
        let errors = plain.split(separator: "\n").filter { $0.contains("error:") }
        return (errors.isEmpty ? plain.split(separator: "\n").suffix(5) : errors.prefix(5)).joined(separator: " / ")
    }
}
