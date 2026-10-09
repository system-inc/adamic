import CohereSwift
import Foundation

/*
 SwiftLint over exactly the files the engine owns, with only the mapped rules on, so every finding it
 prints is one a cohere rule claims to match.

 The binary is the pinned 0.65.1 under cohere's `.cache`, never one from the PATH: a parity result is a
 statement about one version, recorded beside the comparison.
 */
struct SwiftLintRun {
    let executable: URL
    let root: URL
    let files: [URL]
    let runner: ProcessRunner

    /* SwiftLint's JSON reporter, the fields read. */
    private struct Violation: Decodable {
        var file: String?
        var line: Int?
        var ruleIdentifier: String

        enum CodingKeys: String, CodingKey {
            case file
            case line
            case ruleIdentifier = "rule_id"
        }
    }

    struct Failure: Error, CustomStringConvertible {
        var description: String
    }

    func version() throws -> String {
        String(decoding: try runner.run(executable.path, ["version"], in: root).standardOutput, as: UTF8.self)
            .trimmingCharacters(in: .whitespacesAndNewlines)
    }

    func findings() throws -> [RuleFinding] {
        let rules = Set(RuleMapping.incumbentRules(of: .swiftLint))
        let configuration = TemporaryFile.url(extension: "yml")
        /* `file_length` runs at cohere's threshold, not SwiftLint's 400, so the two rules answer the same question. */
        let thresholds =
            rules.contains("file_length")
            ? "file_length:\n  warning: \(FileLength.maximumLines)\n  error: \(FileLength.maximumLines)\n" : ""
        try ("only_rules:\n" + rules.sorted().map { "  - \($0)\n" }.joined() + thresholds).write(
            to: configuration,
            atomically: true,
            encoding: .utf8,
        )
        defer { TemporaryFile.remove(configuration) }

        let result = try runner.run(
            executable.path,
            ["lint", "--config", configuration.path, "--reporter", "json", "--quiet", "--no-cache"] + files.map(\.path),
            in: root,
        )
        /* SwiftLint exits 2 when it found violations at error severity, which is an answer; anything else nonzero is a failure. */
        guard result.exitCode == 0 || result.exitCode == 2 else {
            throw Failure(description: "swiftlint exited \(result.exitCode): \(result.standardError)")
        }
        return try JSONDecoder().decode([Violation].self, from: result.standardOutput).compactMap { violation in
            guard rules.contains(violation.ruleIdentifier), let file = violation.file else { return nil }
            return RuleFinding(
                file: PackagePath.relative(file, to: root),
                line: violation.line ?? 1,
                rule: violation.ruleIdentifier,
            )
        }
    }
}
