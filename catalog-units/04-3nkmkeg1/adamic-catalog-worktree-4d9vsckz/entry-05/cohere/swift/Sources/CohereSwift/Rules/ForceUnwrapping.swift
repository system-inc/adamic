import SwiftSyntax

/*
 No force unwrap: `value!`. The Swift form of our TypeScript `no-non-null-assertion`.

 Kirk's ruling: forbidden everywhere, report only, no automatic fix. A force unwrap is a crash the author
 decided could not happen, written so the reader cannot see the reasoning. The repair is a `guard let` or
 `if let` that says what happens on nil, and choosing that behavior is the point, so no fixer chooses it.

 Matches SwiftLint's `force_unwrapping` and swift-format's `NeverForceUnwrap` (off in our `.swift-format`
 because this rule replaces it). Implicitly unwrapped optional types (`T!`) are their own rule.
 */
public struct ForceUnwrapping: FileRule {
    public let name = "cohere-swift/force-unwrapping"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "force_unwrapping"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { node in
            file.finding(
                at: node.exclamationMark,
                rule: name,
                messageId: "forceUnwrap",
                message:
                    "This force unwrap crashes the process when the value is nil, and says nothing about why that cannot happen. Say what happens on nil with guard let or if let instead.",
            )
        }
    }

    /* Collects every force unwrap expression. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [ForceUnwrapExprSyntax] = []

        override func visit(_ node: ForceUnwrapExprSyntax) -> SyntaxVisitorContinueKind {
            found.append(node)
            return .visitChildren
        }
    }
}
