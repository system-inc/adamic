import SwiftSyntax

/* Counts every node in a tree, for the coverage line's `nodes visited`: a population, so a reader can tell a rule set that walked a file from one that never reached it. */
final class NodeCounter: SyntaxAnyVisitor {
    private(set) var count = 0

    override func visitAny(_ node: Syntax) -> SyntaxVisitorContinueKind {
        count += 1
        return .visitChildren
    }
}
