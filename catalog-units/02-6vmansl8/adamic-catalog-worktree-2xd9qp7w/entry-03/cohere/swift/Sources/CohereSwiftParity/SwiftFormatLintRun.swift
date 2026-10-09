import CohereSwift
import Foundation

/*
 swift-format's linter (the toolchain's, `xcrun swift-format lint`) over the engine's files, with every
 rule off except the mapped ones. Its findings are printed as text, one per line:
 `<path>:<line>:<column>: warning: [<Rule>] <message>`.
 */
struct SwiftFormatLintRun {
    let root: URL
    let files: [URL]
    let runner: ProcessRunner

    struct Failure: Error, CustomStringConvertible {
        var description: String
    }

    func findings() throws -> [RuleFinding] {
        let mapped = Set(RuleMapping.incumbentRules(of: .swiftFormat))

        /* The default configuration with its rule table rewritten, so a rule added to swift-format later arrives off rather than unmapped. */
        let dump = try runner.run("xcrun", ["swift-format", "dump-configuration"], in: root)
        guard dump.succeeded,
            var configuration = try JSONSerialization.jsonObject(with: dump.standardOutput) as? [String: Any],
            let rules = configuration["rules"] as? [String: Any]
        else {
            throw Failure(
                description: "xcrun swift-format dump-configuration gave no rule table: \(dump.standardError)"
            )
        }
        configuration["rules"] = Dictionary(uniqueKeysWithValues: rules.keys.map { ($0, mapped.contains($0)) })
        let configurationFile = TemporaryFile.url(extension: "swift-format")
        try JSONSerialization.data(withJSONObject: configuration).write(to: configurationFile)
        defer { TemporaryFile.remove(configurationFile) }

        let result = try runner.run(
            "xcrun",
            ["swift-format", "lint", "--configuration", configurationFile.path] + files.map(\.path),
            in: root,
        )
        var findings: [RuleFinding] = []
        for line in result.standardError.split(separator: "\n") {
            guard
                let match = line.firstMatch(of: #/^(.+?):(\d+):\d+: \w+: \[(\w+)\]/#),
                mapped.contains(String(match.3)),
                let lineNumber = Int(match.2)
            else { continue }
            findings.append(
                RuleFinding(
                    file: PackagePath.relative(String(match.1), to: root),
                    line: lineNumber,
                    rule: String(match.3),
                )
            )
        }
        /* Exit 1 means it found something. A nonzero exit with nothing parsed means it said something this reader does not understand. */
        if !result.succeeded && findings.isEmpty {
            throw Failure(
                description:
                    "xcrun swift-format lint exited \(result.exitCode) with no finding this harness could read: \(result.standardError.prefix(500))"
            )
        }
        return findings
    }
}
