import Foundation

/*
 The gate before the types phase may take its diagnostics from sourcekitd instead of the build: for every
 file the engine owns, what sourcekitd says against what the build's `.dia` records say, finding for
 finding.

 The build is the oracle because it is the compiler's own verdict on the package as `swift build` compiles
 it. sourcekitd is the same compiler, but it is fed the arguments read back from the build's task store,
 and a flag lost on the way (a define, a search path, an upcoming feature) changes what it reports. Only
 agreement on whole real packages shows the arguments arrived intact.

 A comparison that could not fail proves nothing, so the oracle also asks one file again with an injected
 type error and requires that error back at its line. Coverage comes first in the report: files asked,
 files no compile command listed, and files sourcekitd could not answer for are all counted, so a run that
 compared nothing never reads as agreement.
 */
public struct TypesOracle {
    /* One compiler diagnostic, compared by where it is, how severe, and what it says. */
    public struct Diagnostic: Hashable, Comparable, Sendable {
        public var file: String
        public var line: Int
        public var column: Int
        public var severity: String
        public var message: String

        public static func < (left: Diagnostic, right: Diagnostic) -> Bool {
            (left.file, left.line, left.column, left.message) < (right.file, right.line, right.column, right.message)
        }
    }

    public struct Report: Sendable {
        public var filesOwned: Int
        public var filesAsked: Int
        public var filesWithoutCommand: [String]
        public var filesUnanswered: [String: String]
        public var build: [Diagnostic]
        public var sourcekitd: [Diagnostic]
        /* Diagnostics both sides report whose warning group differs: the build's name, then the one sourcekitd's documentation link gives. */
        public var groupDifferences: [(Diagnostic, String, String)]
        public var controlFile: String
        public var controlFound: Bool
        public var buildMilliseconds: Int
        public var sourcekitdMilliseconds: Int
        public var slowestFiles: [(String, Int)]

        public var onlyBuild: [Diagnostic] { Set(build).subtracting(sourcekitd).sorted() }
        public var onlySourcekitd: [Diagnostic] { Set(sourcekitd).subtracting(build).sorted() }

        /* Agreement means every owned file was asked and answered, the two sides match, and the control was caught. */
        public var agrees: Bool {
            filesWithoutCommand.isEmpty && filesUnanswered.isEmpty && onlyBuild.isEmpty && onlySourcekitd.isEmpty
                && groupDifferences.isEmpty && controlFound && filesAsked == filesOwned
        }
    }

    let root: URL
    let runner: ProcessRunner

    public init(root: URL, runner: ProcessRunner = ProcessRunner()) {
        self.root = root
        self.runner = runner
    }

    /* The control: a line no package of ours holds, and an error the compiler cannot not report. */
    static let controlLine = "let cohereOracleControl: Int = \"not a number\""

    public func run() throws -> Report {
        let toolchain = EngineVersion.toolchain(runner: runner)
        let scratchPath = Pipeline.scratchPath(for: root)
        let package = try PackageModel.load(root: root, scratchPath: scratchPath, runner: runner, toolchain: toolchain)
        let files = try FileSet.build(package: package).owned

        let buildStart = Date()
        let types = try TypesPhase(
            package: package,
            files: files,
            scratchPath: scratchPath,
            resolutionAllowed: false,
            toolchain: toolchain,
            runner: runner,
        ).run()
        let buildMilliseconds = Pipeline.milliseconds(since: buildStart)
        var buildGroups: [Diagnostic: String] = [:]
        let build = types.findings.filter { $0.messageId != "buildFailed" }.map { finding in
            let diagnostic = Diagnostic(
                file: finding.file,
                line: finding.line,
                column: finding.column,
                severity: finding.severity.rawValue,
                message: finding.message,
            )
            buildGroups[diagnostic] = finding.rule
            return diagnostic
        }

        /* A local package's tests build into their own scratch, so their commands live in that scratch's task store. */
        var commandTables = [try CompileCommands.read(scratchPath: scratchPath)]
        for local in package.localPackages {
            let localScratch = scratchPath.appendingPathComponent(
                "local/\(local.root.lastPathComponent)",
                isDirectory: true,
            )
            if FileManager.default.fileExists(atPath: localScratch.path) {
                commandTables.append(try CompileCommands.read(scratchPath: localScratch))
            }
        }

        let session = try Sourcekitd.shared(runner: runner)
        let sourcekitdStart = Date()
        var sourcekitd: [Diagnostic] = []
        var sourcekitdGroups: [Diagnostic: String] = [:]
        var withoutCommand: [String] = []
        var unanswered: [String: String] = [:]
        var asked = 0
        var timings: [(String, Int)] = []
        var control: (file: String, arguments: [String])?
        for file in files {
            let path = file.url.resolvingSymlinksInPath().path
            guard let command = commandTables.lazy.compactMap({ $0.command(for: file.url) }).first else {
                withoutCommand.append(path)
                continue
            }
            asked += 1
            let fileStart = Date()
            do {
                for diagnostic in try session.diagnostics(file: path, arguments: command.arguments) {
                    guard let severity = diagnostic.findingSeverity else { continue }
                    let placed = URL(fileURLWithPath: diagnostic.file).resolvingSymlinksInPath().path
                    guard placed == path else { continue }
                    let compared = Diagnostic(
                        file: placed,
                        line: max(diagnostic.line, 1),
                        column: max(diagnostic.column, 1),
                        severity: severity.rawValue,
                        message: diagnostic.message,
                    )
                    sourcekitd.append(compared)
                    sourcekitdGroups[compared] = Self.group(of: diagnostic)
                }
            }
            catch {
                unanswered[path] = "\(error)"
            }
            timings.append((path, Pipeline.milliseconds(since: fileStart)))
            if control == nil {
                control = (path, command.arguments)
            }
        }
        let sourcekitdMilliseconds = Pipeline.milliseconds(since: sourcekitdStart)

        var controlFound = false
        if let control {
            let original = try String(contentsOfFile: control.file, encoding: .utf8)
            let injected = original + (original.hasSuffix("\n") ? "" : "\n") + Self.controlLine + "\n"
            let line = injected.split(separator: "\n", omittingEmptySubsequences: false).count - 1
            let found = try session.diagnostics(file: control.file, arguments: control.arguments, text: injected)
            controlFound = found.contains { $0.line == line && $0.findingSeverity == .error }
        }

        let both = Set(build).intersection(sourcekitd)
        let groupDifferences = both.sorted().compactMap { diagnostic -> (Diagnostic, String, String)? in
            let built = buildGroups[diagnostic] ?? ""
            let asked = sourcekitdGroups[diagnostic] ?? ""
            return built == asked ? nil : (diagnostic, built, asked)
        }

        return Report(
            filesOwned: files.count,
            filesAsked: asked,
            filesWithoutCommand: withoutCommand.sorted(),
            filesUnanswered: unanswered,
            build: build.sorted(),
            sourcekitd: sourcekitd.sorted(),
            groupDifferences: groupDifferences,
            controlFile: control?.file ?? "",
            controlFound: controlFound,
            buildMilliseconds: buildMilliseconds,
            sourcekitdMilliseconds: sourcekitdMilliseconds,
            slowestFiles: Array(timings.sorted { $0.1 > $1.1 }.prefix(5)),
        )
    }

    /*
     The warning group sourcekitd implies: the last segment of the diagnostic's documentation link, kebab case
     made Pascal case, the spelling `.dia` records give (`variable-never-mutated` is `VariableNeverMutated`).
     Empty when there is no link, as the build's is for a diagnostic in no group. The oracle checks the
     mapping on every diagnostic both sides report, so a group the link does not spell this way shows up as
     a difference rather than as a silent rename.
     */
    static func group(of diagnostic: Sourcekitd.Diagnostic) -> String {
        guard let link = diagnostic.educationalNotes?.first, let last = link.split(separator: "/").last else {
            return ""
        }
        return last.split(separator: "-").map { $0.prefix(1).uppercased() + $0.dropFirst() }.joined()
    }
}
