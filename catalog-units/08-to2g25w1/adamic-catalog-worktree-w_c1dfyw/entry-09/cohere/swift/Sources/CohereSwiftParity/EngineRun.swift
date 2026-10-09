import CohereSwift
import Foundation

/*
 cohere-swift's lint findings for a package, from the engine's own pipeline run in this process with
 `--no-fix --lint`: the same code `c` runs, nothing written, and no types phase, so a test target that
 does not compile cannot cut the comparison off.
 */
struct EngineRun {
    let root: URL

    func findings() async throws -> [RuleFinding] {
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix", "--lint"],
            workingDirectory: root,
        )
        var stream = Data()
        let writer = ContractWriter { stream.append($0) }
        _ = try await Pipeline(options: options, writer: writer, workingDirectory: root).run()

        let decoder = JSONDecoder()
        var findings: [RuleFinding] = []
        for line in stream.split(separator: UInt8(ascii: "\n")) {
            /* Only findings are decoded, and a finding that does not decode stops the run rather than shrinking the comparison. */
            let object = try JSONSerialization.jsonObject(with: Data(line)) as? [String: Any]
            guard object?["kind"] as? String == "finding" else { continue }
            let record = try decoder.decode(FindingRecord.self, from: Data(line))
            findings.append(
                RuleFinding(
                    file: PackagePath.relative(record.file, to: root),
                    line: record.line,
                    rule: record.rule,
                    messageId: record.messageId,
                )
            )
        }
        return findings
    }
}
