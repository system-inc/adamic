import SwiftSyntax

/*
 No `print` or `debugPrint` in libraries or apps; they log through `os.Logger`.

 Kirk's ruling: decided by target kind, inside the rule, never by a path list in config. A command-line
 tool's stdout is its interface, so printing there is the job. An app's stdout goes nowhere anyone reads,
 and a library cannot know where its caller's stdout goes. Both log, so the message lands in the unified
 log with a subsystem, a category and a level.

 - Library targets: every print is a finding.
 - Application targets: every print is a finding. SwiftPM spells an app and a command-line tool the same
   way (`executableTarget`), so the file set relabels an executable whose files import SwiftUI, AppKit or
   UIKit as `application` (see FileSet.isApplication).
 - Executable targets that are not applications are command-line tools, and may print.
 - Test targets may print. Test output is read by the person running the tests, and our TypeScript
   configuration has no console rule at all, so forbidding it here would be stricter than the house.

 Only the free functions are matched: a method named `print` on some type (`printer.print(page)`) is
 that type's business.
 */
public struct ConsistencyNoPrint: FileRule {
    public let name = "cohere-swift/consistency-no-print"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    public func applies(to file: ParsedFile) -> Bool {
        switch file.targetKind {
            case "test", "manifest", "plugin", "macro":
                return false
            case "executable":
                return false
            default:
                return true
        }
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        let kind = file.targetKind == "application" ? "an app" : "a library"
        return visitor.found.map { call in
            file.finding(
                at: call,
                rule: name,
                message: RuleMessages.ConsistencyNoPrint.consistencyNoConsole(kind: kind),
            )
        }
    }

    /* Collects calls to the free functions `print` and `debugPrint`. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [DeclReferenceExprSyntax] = []

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            if let callee = node.calledExpression.as(DeclReferenceExprSyntax.self),
                callee.baseName.text == "print" || callee.baseName.text == "debugPrint"
            {
                found.append(callee)
            }
            return .visitChildren
        }
    }
}
