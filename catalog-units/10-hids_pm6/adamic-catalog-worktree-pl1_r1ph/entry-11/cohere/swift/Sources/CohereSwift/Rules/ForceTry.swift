import SwiftSyntax

/*
 No force try: `try! work()`. A thrown error becomes a crash, and the error's own explanation is lost with
 it. The repair is `do`/`catch` that handles the failure, or `try` that passes it up, so the failure is said
 rather than detonated.

 Kirk's ruling on force unwraps covers this, since it is the same unchecked assertion: forbidden
 everywhere, report only. Matches SwiftLint's `force_try` and swift-format's `NeverUseForceTry`.
 */
public struct ForceTry: FileRule {
    public let name = "cohere-swift/force-try"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "force_try"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { mark in
            file.finding(
                at: mark,
                rule: name,
                messageId: "forceTry",
                message:
                    "try! turns a thrown error into a crash and throws away the error's explanation. Handle it with do/catch, or pass it up with try.",
            )
        }
    }

    /* Collects the `!` of every `try!`. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [TokenSyntax] = []

        override func visit(_ node: TryExprSyntax) -> SyntaxVisitorContinueKind {
            if let mark = node.questionOrExclamationMark, mark.tokenKind == .exclamationMark {
                found.append(mark)
            }
            return .visitChildren
        }
    }
}
