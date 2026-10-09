import Foundation

/*
 The command line, as the front door passes it: `cohere-swift --contract <version> --root <dir> [flags] [paths]`.

 Parsing refuses rather than guesses. An unknown flag, a flag with no Swift meaning yet, or a contract
 version this engine does not speak stops the run before any record is written. A flag accepted and
 ignored produces a run that reads as though it did what was asked.

 Flags are accepted with one dash or two, because Go's flag package (the front door's) treats them
 the same and a user may have typed either.
 */
public struct CommandOptions: Equatable, Sendable {
    /* The command line asked for something this engine cannot do. Exit 2: nothing was checked. */
    public struct UsageFailure: Error, CustomStringConvertible {
        public var description: String
    }

    public var contract: Int?
    public var root: URL?
    public var noFix = false
    public var fixOnly = false
    public var typesOnly = false
    public var lintOnly = false
    public var formatAll = false
    public var lintConfiguration: URL?
    public var fixPasses = 10
    public var singleThreaded = false
    public var listRules = false
    public var listRulesEnabled = false
    public var showVersion = false
    public var unused = false
    public var paths: [String] = []

    public var runFix: Bool { fixOnly || !anyPhaseNamed }
    public var runTypes: Bool { typesOnly || !anyPhaseNamed }
    public var runLint: Bool { lintOnly || !anyPhaseNamed }
    public var mutate: Bool { runFix && !noFix }
    private var anyPhaseNamed: Bool { fixOnly || typesOnly || lintOnly }

    public init() {}

    public static func parse(_ arguments: [String], workingDirectory: URL) throws -> CommandOptions {
        var options = CommandOptions()
        var remaining = arguments[...]

        func value(for flag: String) throws -> String {
            guard let next = remaining.popFirst() else {
                throw UsageFailure(description: "\(flag) needs a value")
            }
            return next
        }

        while let argument = remaining.popFirst() {
            guard argument.hasPrefix("-"), argument != "-" else {
                options.paths.append(argument)
                continue
            }
            let flag = "--" + argument.drop { $0 == "-" }
            switch flag {
                case "--contract":
                    let spelled = try value(for: flag)
                    guard let version = Int(spelled) else {
                        throw UsageFailure(description: "--contract takes a number, not \(spelled)")
                    }
                    options.contract = version
                case "--root":
                    options.root =
                        URL(fileURLWithPath: try value(for: flag), relativeTo: workingDirectory).standardizedFileURL
                case "--no-fix": options.noFix = true
                case "--fix": options.fixOnly = true
                case "--types": options.typesOnly = true
                case "--lint": options.lintOnly = true
                /* Formatting is on by default for Swift (the contract says why), so the flag is accepted and changes nothing. */
                case "--format": break
                case "--format-all": options.formatAll = true
                case "--lint-config":
                    options.lintConfiguration =
                        URL(fileURLWithPath: try value(for: flag), relativeTo: workingDirectory).standardizedFileURL
                case "--fix-passes":
                    let spelled = try value(for: flag)
                    guard let passes = Int(spelled), passes > 0 else {
                        throw UsageFailure(description: "--fix-passes takes a positive number, not \(spelled)")
                    }
                    options.fixPasses = passes
                case "--single-threaded": options.singleThreaded = true
                case "--rules": options.listRules = true
                case "--rules-enabled": options.listRulesEnabled = true
                case "--version": options.showVersion = true
                case "--unused", "--unused-all", "--unused-deep": options.unused = true
                case "--timing", "--explain":
                    throw UsageFailure(
                        description:
                            "\(flag) is not implemented for Swift yet, so the run was refused rather than run without it"
                    )
                case "--tsconfig":
                    throw UsageFailure(
                        description: "--tsconfig names a TypeScript program, and this is a Swift package"
                    )
                default:
                    throw UsageFailure(description: "unknown flag \(argument)")
            }
        }

        if options.noFix && options.fixOnly {
            throw UsageFailure(
                description:
                    "--fix and --no-fix contradict each other: --fix runs only the mutating phase, --no-fix mutates nothing"
            )
        }
        if let contract = options.contract, contract != EngineVersion.contract {
            throw UsageFailure(
                description:
                    "the front door speaks contract \(contract) and this engine speaks contract \(EngineVersion.contract), so nothing was checked"
            )
        }
        return options
    }
}
