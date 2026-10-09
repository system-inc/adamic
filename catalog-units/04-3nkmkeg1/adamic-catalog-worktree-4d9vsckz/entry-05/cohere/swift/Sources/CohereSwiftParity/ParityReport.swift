import Foundation

/*
 The comparison for one rule against one incumbent, in three lists: both, cohere only, incumbent only.
 Each "cohere only" is a false positive until shown otherwise, and each "incumbent only" is a miss until
 shown to be the incumbent's false positive. The harness prints them; explaining them is the reader's job,
 and the explanation goes on the task with the example.
 */
struct ParityReport {
    var mapping: RuleMapping
    var cohere: [RuleFinding]
    var other: [RuleFinding]

    /* What the comparison keys on, per the mapping: a line, or only the file. */
    private func keys(_ findings: [RuleFinding]) -> Set<String> {
        switch mapping.comparison {
            case .line:
                return Set(findings.map { "\($0.file):\($0.line)" })
            case .file:
                return Set(findings.map(\.file))
        }
    }

    var both: Set<String> { keys(cohere).intersection(keys(other)) }
    var cohereOnly: [String] { keys(cohere).subtracting(keys(other)).sorted() }
    var incumbentOnly: [String] { keys(other).subtracting(keys(cohere)).sorted() }

    func summaryLine() -> String {
        let name = "\(mapping.incumbent.rawValue) \(mapping.incumbentRule)"
        return
            "\(name.padding(toLength: 50, withPad: " ", startingAt: 0)) cohere \(cohere.count)  incumbent \(other.count)  both \(both.count)  cohere-only \(cohereOnly.count)  incumbent-only \(incumbentOnly.count)"
    }

    /* Each difference with the source line it points at, so a reader can judge it without opening the file. */
    func differences(root: URL, examples: Int) -> String {
        var text = ""
        for (label, list) in [("cohere only", cohereOnly), ("\(mapping.incumbent.rawValue) only", incumbentOnly)]
        where !list.isEmpty {
            text += "  \(label):\n"
            for key in list.prefix(examples) {
                text += "    \(key)\(Self.sourceLine(for: key, root: root))\n"
            }
            if list.count > examples {
                text += "    and \(list.count - examples) more\n"
            }
        }
        return text
    }

    private static func sourceLine(for key: String, root: URL) -> String {
        let parts = key.split(separator: ":")
        guard parts.count == 2, let line = Int(parts[1]),
            let text = try? String(contentsOf: root.appendingPathComponent(String(parts[0])), encoding: .utf8)
        else { return "" }
        let lines = text.split(separator: "\n", omittingEmptySubsequences: false)
        guard line >= 1, line <= lines.count else { return "" }
        return "  " + lines[line - 1].trimmingCharacters(in: .whitespaces).prefix(120)
    }
}
