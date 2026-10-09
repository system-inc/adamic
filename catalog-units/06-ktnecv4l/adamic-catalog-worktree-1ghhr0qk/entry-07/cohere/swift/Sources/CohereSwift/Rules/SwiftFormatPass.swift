import Foundation
import SwiftFormat
import SwiftOperators
import SwiftSyntax

/*
 One swift-format lint pass per file, shared by every `SwiftFormatRule`.

 Each wrapped rule used to fold the tree and run swift-format's lint pipeline on its own, so two naming
 rules cost two passes over every file: on ahraos-presence they took lint from 0.6s to 3.1s. Now the first
 rule to ask runs one pass with every wrapped incumbent rule enabled, and the rest read its result.

 The result is keyed by the file's path and its exact text, so a file the fixer rewrote is a new key and is
 never answered from an older text. Files are linted in parallel, so the store is locked; a pass runs
 outside the lock, and two rules asking at once may both run it, which costs time and never correctness.
 */
final class SwiftFormatPass: @unchecked Sendable {
    /* What the pass found, reduced to what a rule reports. */
    struct Finding: Sendable {
        var rule: String
        var line: Int
        var column: Int
        var text: String
    }

    typealias Outcome = [Finding]

    static let shared = SwiftFormatPass()

    private let lock = NSLock()
    private var outcomes: [String: Outcome] = [:]

    func findings(in file: ParsedFile) throws -> Outcome {
        let key = file.url.path + "\u{0}" + file.source
        if let known = lock.withLock({ outcomes[key] }) {
            return known
        }
        let found = try Self.run(file)
        lock.withLock { outcomes[key] = found }
        return found
    }

    private static func run(_ file: ParsedFile) throws -> Outcome {
        /* The house format with its rule table rewritten: the layout settings are the house's, and only the wrapped rules run. */
        var configuration = HouseSwiftFormat.configuration
        for rule in configuration.rules.keys {
            configuration.rules[rule] = false
        }
        for rule in SwiftFormatRule.incumbentRules {
            configuration.rules[rule] = true
        }
        var reported: [Finding] = []
        let linter = SwiftLinter(configuration: configuration) { finding in
            guard let location = finding.location else { return }
            reported.append(
                Finding(
                    rule: String(describing: finding.category),
                    line: location.line,
                    column: location.column,
                    text: finding.message.text,
                )
            )
        }
        linter.debugOptions = [.disablePrettyPrint]
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        guard let tree = folded.as(SourceFileSyntax.self) else { return [] }
        try linter.lint(
            syntax: tree,
            source: file.source,
            operatorTable: .standardOperators,
            assumingFileURL: file.url,
        )
        return reported
    }
}
