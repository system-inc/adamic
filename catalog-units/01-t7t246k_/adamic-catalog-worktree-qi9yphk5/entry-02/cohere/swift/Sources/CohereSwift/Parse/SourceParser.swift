import Foundation
import SwiftDiagnostics
import SwiftParser
import SwiftParserDiagnostics
import SwiftSyntax

/*
 Parses every file in parallel, once.

 A file that does not parse is not dropped. It comes back as findings, because a file the run silently
 skipped and a file with nothing wrong in it would otherwise look identical, and the fix phase bails on
 them: rewriting a tree the parser could not read writes nonsense.
 */
public struct SourceParser: Sendable {
    /* The outcome of parsing every file in scope. */
    public struct Result: Sendable {
        public var files: [ParsedFile]
        public var parseErrors: [FindingRecord]
        public var unreadable: [LintRecord.Crash]
    }

    public init() {}

    public func parse(_ owned: [FileSet.OwnedFile]) async -> Result {
        await withTaskGroup(of: (Int, Outcome).self) { group in
            for (index, file) in owned.enumerated() {
                group.addTask { (index, Self.parseOne(file)) }
            }
            var outcomes = [Outcome?](repeating: nil, count: owned.count)
            for await (index, outcome) in group {
                outcomes[index] = outcome
            }
            /* Kept in input order, so two runs over the same tree write their findings in the same order. */
            var result = Result(files: [], parseErrors: [], unreadable: [])
            for outcome in outcomes.compactMap({ $0 }) {
                switch outcome {
                    case let .parsed(file, errors):
                        result.files.append(file)
                        result.parseErrors.append(contentsOf: errors)
                    case let .unreadable(crash):
                        result.unreadable.append(crash)
                }
            }
            return result
        }
    }

    /* One file's outcome. Cases rather than optionals, because "could not read it" must never collapse into "nothing to say". */
    enum Outcome: Sendable {
        case parsed(ParsedFile, [FindingRecord])
        case unreadable(LintRecord.Crash)
    }

    static func parseOne(_ owned: FileSet.OwnedFile) -> Outcome {
        let source: String
        do {
            source = try String(contentsOf: owned.url, encoding: .utf8)
        }
        catch {
            return .unreadable(
                LintRecord.Crash(
                    file: owned.url.path,
                    error: "it is not valid UTF-8 or could not be opened (\(error))",
                )
            )
        }
        let tree = Parser.parse(source: source)
        let counter = NodeCounter(viewMode: .sourceAccurate)
        counter.walk(tree)
        let file = ParsedFile(
            url: owned.url,
            targetName: owned.targetName,
            targetKind: owned.targetKind,
            source: source,
            tree: tree,
            nodeCount: counter.count,
            packageRoot: owned.packageRoot,
        )
        guard tree.hasError else {
            return .parsed(file, [])
        }
        let errors = ParseDiagnosticsGenerator.diagnostics(for: tree).map { diagnostic in
            let location = diagnostic.location(converter: file.locations)
            return FindingRecord(
                source: .compiler,
                file: owned.url.path,
                line: location.line,
                column: location.column,
                severity: diagnostic.diagMessage.severity == .warning ? .warning : .error,
                rule: "",
                messageId: "parseError",
                message: diagnostic.message,
            )
        }
        return .parsed(file, errors)
    }
}
