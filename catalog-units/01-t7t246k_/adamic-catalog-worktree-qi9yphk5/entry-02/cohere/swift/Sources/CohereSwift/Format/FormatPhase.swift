import Foundation
import SwiftFormat

/*
 The fix phase's formatting half: swift-format, as a library inside this process, applying the house
 format (`HouseSwiftFormat`) to every file it is handed.

 It formats from source text rather than from the tree the engine already parsed. The source-text entry
 point is the one measured byte-identical to `xcrun swift-format` (803 Presence files, 149 ahraos-macos
 files, 117 of them actually rewritten). The tree entry point needs the tree folded with swift-format's
 operator table first, and nothing has measured it. The cost is a second parse inside the formatter,
 about a millisecond per file across cores, and the parity is worth more than that.
 */
struct FormatPhase {
    /* What formatting did to each file. */
    enum Outcome: Sendable {
        case unchanged
        case changed(formatted: String)
        case failed(reason: String)
    }

    /* Per-file outcomes, in input order. */
    struct Result: Sendable {
        var outcomes: [(file: ParsedFile, outcome: Outcome)]
    }

    func run(_ files: [ParsedFile]) async -> Result {
        let outcomes = await withTaskGroup(of: (Int, Outcome).self) { group in
            for (index, file) in files.enumerated() {
                group.addTask { (index, Self.format(file)) }
            }
            var collected = [Outcome?](repeating: nil, count: files.count)
            for await (index, outcome) in group {
                collected[index] = outcome
            }
            return collected
        }
        return Result(
            outcomes: zip(files, outcomes).map { ($0, $1 ?? .failed(reason: "the formatter never answered")) }
        )
    }

    static func format(_ file: ParsedFile) -> Outcome {
        do {
            var output = ""
            try SwiftFormatter(configuration: HouseSwiftFormat.configuration).format(
                source: file.source,
                assumingFileURL: file.url,
                selection: .infinite,
                to: &output,
            )
            return output == file.source ? .unchanged : .changed(formatted: output)
        }
        catch {
            return .failed(reason: "\(error)")
        }
    }

    /* The first line where the formatted text differs, for the finding a `--no-fix` run reports instead of writing. */
    static func firstDifferingLine(_ original: String, _ formatted: String) -> Int {
        let originalLines = original.split(separator: "\n", omittingEmptySubsequences: false)
        let formattedLines = formatted.split(separator: "\n", omittingEmptySubsequences: false)
        for index in 0..<max(originalLines.count, formattedLines.count) {
            let before = index < originalLines.count ? originalLines[index] : nil
            let after = index < formattedLines.count ? formattedLines[index] : nil
            if before != after {
                return index + 1
            }
        }
        return 1
    }
}
