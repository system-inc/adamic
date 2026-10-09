import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `security-no-interpolated-shell-command` both ways. The unit cases parse a source string and hand the rule symbols
 built by position, one occurrence at every token a case names, so each case says what the compiler would have
 resolved; a name no case resolves is a name the index did not record. The four real sites the survey found are
 modeled first, in AhraOS's macOS app and in Presence, each before and after its repair, then every other shape
 the rule flags and every shape it leaves alone, the TypeScript rule's cases carried over where Swift has them.
 The end-to-end case runs a real package, so the symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct SecurityNoInterpolatedShellCommandTests {
    /* The words a finding carries, pinned here so a generator that drops or mangles them fails. */
    static let message =
        "This command string is run by a shell, and a value built into it is not a literal, so the shell parses that value as code: a `\"`, `'` or space in it breaks the command, and a `$(...)` or backtick in it runs. Quoting it by hand is one escape away from the same bug. Keep the script literal and pass the value after it as its own argument, which the shell reads as `$1` and never parses (`[\"-c\", \"cd \\\"$1\\\" && make\", \"zsh\", path]`), or run the program directly with no shell (`currentDirectoryURL` in place of a `cd`)."

    /* `NSString.appendingPathComponent(_:)`, as the index names the Objective-C method it imports. */
    static let appendingPathComponent: (symbol: String, name: String) = (
        "c:objc(cs)NSString(im)stringByAppendingPathComponent:", "appendingPathComponent(_:)",
    )
    static let quoted: (symbol: String, name: String) = ("s:7Control9DescribedO6quotedyS2SFZ", "quoted(_:)")
    static let urlPath: (symbol: String, name: String) = ("s:10Foundation3URLV4pathSSvp", "path")
    static let processIdentifier: (symbol: String, name: String) = (
        "c:objc(cs)NSTask(py)processIdentifier", "processIdentifier",
    )
    /* `static let root: String` of ours, declared in another file. */
    static let configurationRoot: (symbol: String, name: String) = ("s:7Control6ConfigO4rootSSvpZ", "root")
    static let boxCount: (symbol: String, name: String) = ("s:7Control3BoxV5countSivp", "count")

    /* The text of every finding's span, with every token named in `resolving` resolved to its declaration. */
    static func spans(_ source: String, resolving: [String: (symbol: String, name: String)] = [:]) -> [String] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        var occurrences: [FileSymbols.Occurrence] = []
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            guard let resolved = resolving[token.text] else { continue }
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            occurrences.append(
                FileSymbols.Occurrence(
                    line: location.line,
                    column: location.column,
                    symbol: resolved.symbol,
                    name: resolved.name,
                    isReference: true,
                )
            )
        }
        let rule = SecurityNoInterpolatedShellCommand()
        let findings = rule.findings(in: file, symbols: FileSymbols(occurrences, ownedModules: ["Control"]))
        #expect(findings.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        for finding in findings {
            #expect(finding.messageId == "interpolatedShellCommand")
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
            return String(decoding: bytes, as: UTF8.self).trimmingCharacters(in: .whitespaces)
        }
    }

    // MARK: The real sites

    /*
     `ahraos-macos` `Sources/AhraOsServices/MemberLifecycle.swift:34` on 2026-10-03: the home directory is spliced
     unquoted after `cd`, so a space or a quote in it breaks the sweep. `projectRoot` is NSString's
     `appendingPathComponent`, which the index names by its Objective-C selector.
     */
    static func memberLifecycle(_ argumentLines: String, setup: String = "") -> String {
        #"""
        import Foundation

        enum MemberLifecycle {
            static func sweepDeadWatcherRows() {
                let projectRoot = (NSHomeDirectory() as NSString).appendingPathComponent("Projects/ahra")
                let process = Process()
                process.executableURL = URL(fileURLWithPath: "/bin/zsh")\#(setup)
                process.arguments = [
        \#(argumentLines)
                ]
                process.standardOutput = nil
                process.standardError = nil
                do {
                    try process.run()
                    process.waitUntilExit()
                }
                catch {
                    TraceLog.write("resetAll: dead-watcher sweep failed to launch: \(error.localizedDescription)")
                }
            }
        }

        """#
    }

    static let memberLifecycleArguments = #"""
                    "-l", "-i", "-c",
                    "cd \(projectRoot) && ahra os sleep --force --dead-watchers --yes",
        """#

    @Test func memberLifecyclesHomePathIsFlagged() {
        let source = Self.memberLifecycle(Self.memberLifecycleArguments)
        #expect(
            Self.spans(source, resolving: ["appendingPathComponent": Self.appendingPathComponent]) == [
                #""cd \(projectRoot) && ahra os sleep --force --dead-watchers --yes""#
            ]
        )
    }

    /* Without the index, `appendingPathComponent`'s result is unread, and nothing is said rather than guessed. */
    @Test func memberLifecycleWithoutTheIndexIsNotGuessed() {
        #expect(Self.spans(Self.memberLifecycle(Self.memberLifecycleArguments)).isEmpty)
    }

    /* The repairs: the path as the shell's `$1`, never parsed, or as the working directory with no `cd` at all. */
    @Test func memberLifecyclesRepairsAreClean() {
        let positional = Self.memberLifecycle(
            #"""
                        "-l", "-i", "-c",
                        "cd \"$1\" && ahra os sleep --force --dead-watchers --yes",
                        "zsh", projectRoot,
            """#
        )
        #expect(Self.spans(positional, resolving: ["appendingPathComponent": Self.appendingPathComponent]).isEmpty)
        let workingDirectory = Self.memberLifecycle(
            #"""
                        "-l", "-i", "-c",
                        "ahra os sleep --force --dead-watchers --yes",
            """#,
            setup: "\n        process.currentDirectoryURL = URL(fileURLWithPath: projectRoot)",
        )
        #expect(
            Self.spans(workingDirectory, resolving: ["appendingPathComponent": Self.appendingPathComponent]).isEmpty
        )
    }

    /*
     `ahraos-presence` `Sources/AhraOSPresence/Studio/Kingdom.swift:47` on 2026-10-03: a `String` parameter is
     concatenated after `cd`. Its callers pass literals today, and the parameter's type says nothing of that, as a
     `string` parameter does not in TypeScript.
     */
    static func kingdom(_ argumentsLine: String, parameter: String = "_ command: String") -> String {
        #"""
        import Foundation

        @MainActor
        enum Kingdom {
            static func awake() async -> [String] {
                await Self.read("ahra os kingdom --awake 2>/dev/null")
            }

            private static func read(\#(parameter)) async -> [String] {
                await withCheckedContinuation { continuation in
                    DispatchQueue.global(qos: .userInitiated).async {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                        \#(argumentsLine)

                        let output = Pipe()
                        process.standardOutput = output
                        guard (try? process.run()) != nil else {
                            continuation.resume(returning: [])
                            return
                        }
                        let data = output.fileHandleForReading.readDataToEndOfFile()
                        process.waitUntilExit()
                        continuation.resume(returning: [String(decoding: data, as: UTF8.self)])
                    }
                }
            }
        }

        """#
    }

    @Test func kingdomsConcatenatedParameterIsFlagged() {
        let source = Self.kingdom(#"process.arguments = ["-lc", "cd ~/Projects/ahra && " + command]"#)
        #expect(Self.spans(source) == [#""cd ~/Projects/ahra && " + command"#])
    }

    /* The repair: the command as `$1`, run by the shell as a command but never spliced into the script's text. */
    @Test func kingdomsRepairIsClean() {
        let source = Self.kingdom(#"process.arguments = ["-lc", "cd ~/Projects/ahra && eval \"$1\"", "zsh", command]"#)
        #expect(Self.spans(source).isEmpty)
    }

    /* A parameter whose type the rule does not read (an enum of the commands the author wrote) is not guessed at. */
    @Test func kingdomWithAnEnumParameterIsNotGuessed() {
        let source = Self.kingdom(
            #"process.arguments = ["-lc", "cd ~/Projects/ahra && " + command.rawValue]"#,
            parameter: "_ command: Listing",
        )
        #expect(Self.spans(source, resolving: ["rawValue": ("s:7Control7ListingO8rawValueSSvp", "rawValue")]).isEmpty)
    }

    /*
     `ahraos-presence` `Sources/AhraOSPresence/Studio/Described.swift:121` on 2026-10-03: a description and a prompt
     wrapped by a single-quote escaping helper. The escape is right where it is used, and it is reported anyway, as
     the TypeScript rule reports `python3 -c '${script.replace(...)}'`: the doc comment says why.
     */
    static func described(_ commandLines: String) -> String {
        #"""
        import Foundation

        enum Described {
            static func instruction(for capabilities: Int) -> String {
                "You are choosing values for a character."
            }

            static func ask(_ description: String, capabilities: Int) async -> String? {
                let process = Process()
                process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                process.arguments = [
                    "-lc",
        \#(commandLines)
                ]
                let output = Pipe()
                process.standardOutput = output
                do {
                    try process.run()
                }
                catch {
                    return nil
                }
                let data = output.fileHandleForReading.readDataToEndOfFile()
                process.waitUntilExit()
                return String(data: data, encoding: .utf8)
            }

            /// Shell quoting, because a description is a sentence and sentences contain quotes.
            private static func quoted(_ text: String) -> String {
                "'" + text.replacingOccurrences(of: "'", with: "'\\''") + "'"
            }
        }

        """#
    }

    @Test func describedsQuotingHelperIsStillFlagged() {
        let source = Self.described(
            #"""
                        "cd ~/Projects/ahra && ahra claude chat "
                            + Self.quoted(description)
                            + " --system-prompt "
                            + Self.quoted(Self.instruction(for: capabilities)),
            """#
        )
        let expected = [
            #"""
            "cd ~/Projects/ahra && ahra claude chat "
                            + Self.quoted(description)
                            + " --system-prompt "
                            + Self.quoted(Self.instruction(for: capabilities))
            """#
        ]
        #expect(Self.spans(source, resolving: ["quoted": Self.quoted]) == expected)
        /* An operand of a `+` with a literal is text whatever the index says, so the finding stands without it. */
        #expect(Self.spans(source) == expected)
    }

    @Test func describedsRepairIsClean() {
        let source = Self.described(
            #"""
                        "cd ~/Projects/ahra && ahra claude chat \"$1\" --system-prompt \"$2\"",
                        "zsh", description, Self.instruction(for: capabilities),
            """#
        )
        #expect(Self.spans(source, resolving: ["quoted": Self.quoted]).isEmpty)
    }

    /*
     `ahraos-macos` `Sources/AhraOs/PaneProcessSpec.swift:29`: a whole command handed to `-c` by design. An opaque
     command is not judged, as `execSync(command)` is not: nothing at this call splices a value into it.
     */
    @Test func paneProcessSpecsWholeCommandIsNotFlagged() {
        let source = #"""
            struct PaneProcessSpec {
                var executable: String = "/bin/zsh"
                var arguments: [String]
                var workingDirectory: String?

                static func interactiveClaude(command: String, workingDirectory: String? = nil) -> PaneProcessSpec {
                    PaneProcessSpec(
                        executable: "/bin/zsh",
                        arguments: ["-l", "-i", "-c", command],
                        workingDirectory: workingDirectory
                    )
                }
            }

            """#
        #expect(Self.spans(source).isEmpty)
    }

    /*
     `ahraos-macos` `Sources/AhraOs/AhraOsWindowModel.swift:1068` and `:1151`, the env-gated bench hooks. The command
     comes from a dictionary's subscript, whose type the rule does not read, and the second hook's arguments are a
     `let` assigned in branches: both are known misses, written down in the rule.
     */
    @Test func benchHooksAreKnownMisses() {
        let source = #"""
            import Foundation

            final class WindowModel {
                func spawnTerminal() -> PaneProcessSpec? {
                    if let benchCommand = ProcessInfo.processInfo.environment["AHRAOS_FILL_BENCH_CMD"], !benchCommand.isEmpty {
                        return PaneProcessSpec(
                            executable: "/bin/zsh",
                            arguments: ["-l", "-i", "-c", "\(benchCommand); exec zsh -l -i"],
                            workingDirectory: "/Users/kirkouimet/Projects/ahra"
                        )
                    }
                    return nil
                }

                func spawnLocalTerminal(benchCommand: String) -> PaneProcessSpec {
                    let arguments: [String]
                    if ProcessInfo.processInfo.environment["AHRAOS_FILL_BENCH_LOCAL"] == "1" {
                        arguments = ["-l", "-i", "-c", "\(benchCommand); exec zsh -l -i"]
                    }
                    else {
                        arguments = ["-l", "-i", "-o", "INC_APPEND_HISTORY"]
                    }
                    return PaneProcessSpec(executable: "/bin/zsh", arguments: arguments, workingDirectory: nil)
                }
            }

            """#
        #expect(Self.spans(source).isEmpty)
    }

    /* The same hook with the command as a `String` parameter: spliced into a larger command, it is flagged, as `execSync(`${command}; exec zsh`)` is. */
    @Test func aWholeCommandSplicedIntoAnotherIsFlagged() {
        let source = #"""
            func benchSpec(benchCommand: String) -> PaneProcessSpec {
                PaneProcessSpec(
                    executable: "/bin/zsh",
                    arguments: ["-l", "-i", "-c", "\(benchCommand); exec zsh -l -i"],
                    workingDirectory: nil
                )
            }

            """#
        #expect(Self.spans(source) == [#""\(benchCommand); exec zsh -l -i""#])
    }

    // MARK: Flagged shapes

    @Test func everyFlaggedShapeIsFound() {
        let cases: [(name: String, source: String, resolving: [String: (symbol: String, name: String)], span: String)] =
            [
                (
                    "Process.run with a URL and arguments",
                    #"""
                    func remove(path: String) throws {
                        _ = try Process.run(URL(fileURLWithPath: "/bin/bash"), arguments: ["-c", "rm -rf \(path)"])
                    }
                    """#,
                    [:], #""rm -rf \(path)""#,
                ),
                (
                    "launchPath and a cluster with c",
                    #"""
                    func open(url: String) {
                        let task = Process()
                        task.launchPath = "/bin/sh"
                        task.arguments = ["-ec", "open \(url)"]
                        task.launch()
                    }
                    """#,
                    [:], #""open \(url)""#,
                ),
                (
                    "an option with its argument before -c",
                    #"""
                    func run(path: String) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/usr/local/bin/bash")
                        process.arguments = ["-o", "pipefail", "--login", "-c", "ls \(path) | wc -l"]
                    }
                    """#,
                    [:], #""ls \(path) | wc -l""#,
                ),
                (
                    "a command built into a let",
                    #"""
                    func build(directory: String) {
                        let script = "cd \(directory) && make"
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                        process.arguments = ["-lc", script]
                    }
                    """#,
                    [:], "script",
                ),
                (
                    "arguments and the shell through lets",
                    #"""
                    func build(directory: String) {
                        let shell = URL(fileURLWithPath: "/bin/zsh")
                        let arguments = ["-lc", "cd \(directory) && make"]
                        let process = Process()
                        process.executableURL = shell
                        process.arguments = arguments
                    }
                    """#,
                    [:], #""cd \(directory) && make""#,
                ),
                (
                    "a URL's path, read from the index",
                    #"""
                    import Foundation
                    func build(directory: URL) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                        process.arguments = ["-lc", "cd \(directory.path) && make"]
                    }
                    """#,
                    ["path": Self.urlPath], #""cd \(directory.path) && make""#,
                ),
                (
                    "a var bound to a literal, which may be reassigned",
                    #"""
                    func build(clean: Bool) {
                        var target = "all"
                        if clean {
                            target = "clean"
                        }
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "make \(target)"]
                    }
                    """#,
                    [:], #""make \(target)""#,
                ),
                (
                    "a guard let bound to an optional String",
                    #"""
                    func open(file: String?) {
                        guard let target = file else { return }
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "open '\(target)'"]
                    }
                    """#,
                    [:], #""open '\(target)'""#,
                ),
                (
                    "an optional with a fallback",
                    #"""
                    func open(file: String?) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "open \(file ?? "/tmp")"]
                    }
                    """#,
                    [:], #""open \(file ?? "/tmp")""#,
                ),
                (
                    "a text branch of a ternary",
                    #"""
                    func list(path: String, all: Bool) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "ls \(all ? "-a \(path)" : "-1")"]
                    }
                    """#,
                    [:], #""ls \(all ? "-a \(path)" : "-1")""#,
                ),
                (
                    "a ternary command",
                    #"""
                    func open(url: String, background: Bool) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", background ? "open -g \(url)" : "true"]
                    }
                    """#,
                    [:], #"background ? "open -g \(url)" : "true""#,
                ),
                (
                    "String(...), which is text whatever it was built from",
                    #"""
                    func stop(pid: Int32) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "kill " + String(pid)]
                    }
                    """#,
                    [:], #""kill " + String(pid)"#,
                ),
                (
                    "a subscript concatenated with a literal",
                    #"""
                    import Foundation
                    func home() {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "ls " + ProcessInfo.processInfo.environment["HOME", default: "/"]]
                    }
                    """#,
                    [:], #""ls " + ProcessInfo.processInfo.environment["HOME", default: "/"]"#,
                ),
                (
                    "a Substring and a Character",
                    #"""
                    func run(name: Substring, separator: Character) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/dash")
                        process.arguments = ["-c", "echo \(name)\(separator)"]
                    }
                    """#,
                    [:], #""echo \(name)\(separator)""#,
                ),
                (
                    "a property of this type declared as String",
                    #"""
                    struct Builder {
                        let directory: String

                        func build() {
                            let process = Process()
                            process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                            process.arguments = ["-lc", "cd \(self.directory) && make"]
                        }
                    }
                    """#,
                    [:], #""cd \(self.directory) && make""#,
                ),
                /* The SQL sits in a quoted heredoc and is data, but the output path follows the closing line and is code. */
                (
                    "a value after a quoted heredoc has closed",
                    #"""
                    func query(sql: String, outputPath: String) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/bash")
                        process.arguments = ["-c", "cat <<'EOSQL' > /tmp/query.sql\n\(sql)\nEOSQL\nmv /tmp/query.sql \(outputPath)"]
                    }
                    """#,
                    [:], #""cat <<'EOSQL' > /tmp/query.sql\n\(sql)\nEOSQL\nmv /tmp/query.sql \(outputPath)""#,
                ),
                /* An unquoted delimiter leaves the body expanded: a `$(...)` in the SQL runs. */
                (
                    "a value in an unquoted heredoc's body",
                    #"""
                    func query(sql: String) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/bash")
                        process.arguments = ["-c", """
                            cat <<EOSQL | pscale shell phi main
                            \(sql)
                            EOSQL
                            """]
                    }
                    """#,
                    [:],
                    "\"\"\"\n        cat <<EOSQL | pscale shell phi main\n        \\(sql)\n        EOSQL\n        \"\"\"",
                ),
                (
                    "values on a quoted heredoc's opener line",
                    #"""
                    func pscaleShell(database: String, branch: String, sql: String) {
                        let command = "cat <<'EOSQL' | pscale shell \(database) \(branch)\n\(sql)\nEOSQL"
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/bash")
                        process.arguments = ["-c", command]
                    }
                    """#,
                    [:], "command",
                ),
                (
                    "a here-string's word",
                    #"""
                    func count(text: String) {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/bash")
                        process.arguments = ["-c", "wc -c <<< \"\(text)\""]
                    }
                    """#,
                    [:], #""wc -c <<< \"\(text)\"""#,
                ),
                (
                    "a closure parameter declared as String",
                    #"""
                    let open: (String) -> Void = { (url: String) in
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "open \(url)"]
                    }
                    """#,
                    [:], #""open \(url)""#,
                ),
                (
                    "a value before one whose type is unread",
                    #"""
                    func stop(name: String, process: Process) {
                        let shell = Process()
                        shell.executableURL = URL(fileURLWithPath: "/bin/sh")
                        shell.arguments = ["-c", "echo \(name); kill \(process.processIdentifier)"]
                    }
                    """#,
                    ["processIdentifier": Self.processIdentifier],
                    #""echo \(name); kill \(process.processIdentifier)""#,
                ),
            ]
        for testCase in cases {
            #expect(Self.spans(testCase.source, resolving: testCase.resolving) == [testCase.span], "\(testCase.name)")
        }
    }

    // MARK: Shapes left alone

    @Test func everySafeShapeIsLeftAlone() {
        let cases: [(name: String, source: String, resolving: [String: (symbol: String, name: String)])] = [
            (
                "an all-literal command",
                #"""
                let spec = SpawnRequest(shell: "/bin/zsh", arguments: ["-l", "-i", "-c", "cd /Users/kirkouimet/Projects/ahra && claude 'hi'"], cwd: "/tmp")
                """#,
                [:],
            ),
            (
                "numbers and booleans",
                #"""
                func stop(pid: Int32, force: Bool) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/sh")
                    process.arguments = ["-c", "kill -\(force ? 9 : 15) \(pid); sleep \(2)"]
                }
                """#,
                [:],
            ),
            (
                "a number read from the index",
                #"""
                func show(box: Box) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/sh")
                    process.arguments = ["-c", "seq \(box.count)"]
                }
                """#,
                ["count": Self.boxCount],
            ),
            (
                "literal constants, local and of the type",
                #"""
                enum Paths {
                    static let projects = "/Users/kirkouimet/Projects"

                    static func build() {
                        let project = "ahra"
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                        process.arguments = ["-lc", "cd \(Self.projects)/\(project) && " + "make " + project]
                    }
                }
                """#,
                [:],
            ),
            (
                "a let of ours in another file, which may be a literal there",
                #"""
                func build() {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                    process.arguments = ["-lc", "cd \(Config.root) && " + Config.root]
                }
                """#,
                ["root": Self.configurationRoot],
            ),
            (
                "the value as a positional parameter",
                #"""
                func remove(path: String) throws {
                    _ = try Process.run(URL(fileURLWithPath: "/bin/sh"), arguments: ["-c", "rm -rf -- \"$1\"", "sh", path])
                }
                """#,
                [:],
            ),
            (
                "no shell: the arguments reach the program as they are",
                #"""
                func log(path: String) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/usr/bin/git")
                    process.arguments = ["-C", path, "log", "-c", "--format=\(path)"]
                }
                """#,
                [:],
            ),
            (
                "an interpreter that is not a POSIX shell",
                #"""
                func run(script: String) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/usr/bin/python3")
                    process.arguments = ["-c", "print('\(script)')"]
                    let fish = Process()
                    fish.executableURL = URL(fileURLWithPath: "/opt/homebrew/bin/fish")
                    fish.arguments = ["-c", "echo \(script)"]
                }
                """#,
                [:],
            ),
            (
                "a shell with no -c",
                #"""
                func open(history: String) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                    process.arguments = ["-l", "-i", "-o", "INC_APPEND_HISTORY", "\(history)"]
                }
                """#,
                [:],
            ),
            (
                "an option the rule cannot read before the command",
                #"""
                func run(flag: String, path: String) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                    process.arguments = [flag, "-c", "ls \(path)"]
                    let other = Process()
                    other.executableURL = URL(fileURLWithPath: "/bin/zsh")
                    other.arguments = ["--command", "ls \(path)"]
                }
                """#,
                [:],
            ),
            (
                "a program that may not be a shell",
                #"""
                func run(path: String, useGit: Bool) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/sh")
                    if useGit {
                        process.executableURL = URL(fileURLWithPath: "/usr/bin/git")
                    }
                    process.arguments = ["-c", "ls \(path)"]
                }
                """#,
                [:],
            ),
            (
                "another receiver's arguments",
                #"""
                func run(path: String) {
                    let shell = Process()
                    shell.executableURL = URL(fileURLWithPath: "/bin/sh")
                    let git = Process()
                    git.executableURL = URL(fileURLWithPath: "/usr/bin/git")
                    git.arguments = ["-c", "core.pager=\(path)", "log"]
                }
                """#,
                [:],
            ),
            (
                "an opaque command held in a var",
                #"""
                func run(path: String) {
                    var command = "ls"
                    command = "ls \(path)"
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/sh")
                    process.arguments = ["-c", command]
                }
                """#,
                [:],
            ),
            /* `i18n-conversion.ts:243`'s shape: the SQL is the body of a heredoc whose delimiter is quoted. */
            (
                "a value in a quoted heredoc's body",
                #"""
                func pscaleQuery(sql: String) {
                    let database = "connected"
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/bash")
                    process.arguments = ["-c", "cat <<'EOSQL' | pscale shell \(database) main --replica\n\(sql)\nEOSQL"]
                }
                """#,
                [:],
            ),
            (
                "values in a double-quoted, a backslashed and a tab-stripped heredoc's body",
                #"""
                func feed(first: String, second: String, third: String) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/bash")
                    process.arguments = ["-c", "cat <<\"ONE\" <<\\TWO <<-'THREE'\n\(first)\nONE\n\(second)\nTWO\n\t\(third)\n\tTHREE"]
                }
                """#,
                [:],
            ),
            (
                "a heredoc delimiter it cannot read",
                #"""
                func feed(delimiter: String, sql: String) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/bash")
                    process.arguments = ["-c", "cat <<'\(delimiter)'\n\(sql)\n\(delimiter)"]
                }
                """#,
                [:],
            ),
            (
                "a ternary branch that opens a heredoc",
                #"""
                func feed(strip: Bool, sql: String) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/bash")
                    process.arguments = ["-c", "pscale shell phi main \(strip ? "<<-'EOSQL'" : "<<'EOSQL'")\n\(sql)\nEOSQL"]
                }
                """#,
                [:],
            ),
            (
                "a value after one whose type is unread",
                #"""
                func stop(name: String, process: Process) {
                    let shell = Process()
                    shell.executableURL = URL(fileURLWithPath: "/bin/sh")
                    shell.arguments = ["-c", "kill \(process.processIdentifier); echo \(name)"]
                }
                """#,
                ["processIdentifier": Self.processIdentifier],
            ),
            (
                "a closure's untyped parameter shadowing a String",
                #"""
                func open(url: String, urls: [String]) {
                    for url in urls {
                        let process = Process()
                        process.executableURL = URL(fileURLWithPath: "/bin/sh")
                        process.arguments = ["-c", "open \(url)"]
                    }
                    _ = urls.map { url in
                        Process.run(URL(fileURLWithPath: "/bin/sh"), arguments: ["-c", "open \(url)"])
                    }
                }
                """#,
                [:],
            ),
            (
                "a value of a type the rule does not read",
                #"""
                func open(target: URL, mode: Mode) {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: "/bin/sh")
                    process.arguments = ["-c", "open \(target) --mode \(mode)"]
                }
                """#,
                [:],
            ),
        ]
        for testCase in cases {
            #expect(Self.spans(testCase.source, resolving: testCase.resolving).isEmpty, "\(testCase.name)")
        }
    }

    /* A file with no argument list and no shell is not read at all. */
    @Test func thePrefilterDeclinesAFileWithNeitherWord() {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: "let value = 1\n",
            tree: Parser.parse(source: "let value = 1\n"),
            nodeCount: 0,
        )
        #expect(!SecurityNoInterpolatedShellCommand().applies(to: file))
    }

    // MARK: End to end

    static let packageSource = #"""
        import Foundation

        enum Described {
            static func ask(_ description: String) -> [String] {
                ["-lc", "ahra claude chat " + Self.quoted(description)]
            }

            static func quoted(_ text: String) -> String {
                "'" + text.replacingOccurrences(of: "'", with: "'\\''") + "'"
            }
        }

        enum Lifecycle {
            static func sweep() throws {
                let projectRoot = (NSHomeDirectory() as NSString).appendingPathComponent("Projects/ahra")
                let process = Process()
                process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                process.arguments = ["-l", "-i", "-c", "cd \(projectRoot) && ahra os sleep"]
                try process.run()
            }

            static func build(directory: URL) throws {
                let process = Process()
                process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                process.arguments = ["-lc", "cd \(directory.path) && make"]
                try process.run()
            }

            static func stop(running: Process) throws {
                let process = Process()
                process.executableURL = URL(fileURLWithPath: "/bin/sh")
                process.arguments = ["-c", "kill \(running.processIdentifier)"]
                try process.run()
            }

            static func ask(_ description: String) throws {
                let process = Process()
                process.executableURL = URL(fileURLWithPath: "/bin/zsh")
                process.arguments = Described.ask(description)
                try process.run()
            }
        }

        """#

    @Test func theIndexResolvesNSStringsAndFoundationsTextAndLeavesAProcessIdentifierAlone() async throws {
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

        /* One run builds the package and writes its index. The rule is run by hand on what the run left. */
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
        let rule = SecurityNoInterpolatedShellCommand()
        let candidates = parsed.files.filter { rule.applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            rule.findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.line):\($0.column)"
            }
        }
        /* The home path through NSString and the URL's path are flagged; the process identifier is not; `Described.ask`'s list is no shell's until a caller makes it one. */
        #expect(found == ["18:48", "25:37"], "\(found)")
    }
}
