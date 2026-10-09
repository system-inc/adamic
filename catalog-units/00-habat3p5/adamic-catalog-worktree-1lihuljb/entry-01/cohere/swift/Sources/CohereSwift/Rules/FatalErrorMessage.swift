import SwiftParser
import SwiftSyntax

/*
 `fatalError`, `preconditionFailure` and `assertionFailure` say why.

 A crash with no message is a crash someone has to reproduce under a debugger to understand. Matches
 SwiftLint's `fatal_error_message`, extended to the two other unconditional failures, which deserve the
 same sentence for the same reason. An empty string literal counts as no message.
 */
public struct FatalErrorMessage: FileRule {
    public let name = "cohere-swift/fatal-error-message"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "fatal_error_message"

    static let failures: Set<String> = ["fatalError", "preconditionFailure", "assertionFailure"]

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { callee in
            file.finding(
                at: callee,
                rule: name,
                messageId: "failureWithoutMessage",
                message:
                    "\(callee.baseName.text)() with no message crashes without saying why. Pass the reason, so the crash report explains itself.",
            )
        }
    }

    /* Collects unconditional failures called with no message or an empty one. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [DeclReferenceExprSyntax] = []

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            guard let callee = node.calledExpression.as(DeclReferenceExprSyntax.self),
                FatalErrorMessage.failures.contains(callee.baseName.text)
            else {
                return .visitChildren
            }
            guard let message = node.arguments.first?.expression else {
                found.append(callee)
                return .visitChildren
            }
            if let literal = message.as(StringLiteralExprSyntax.self), literal.representedLiteralValue?.isEmpty == true
            {
                found.append(callee)
            }
            return .visitChildren
        }
    }
}
