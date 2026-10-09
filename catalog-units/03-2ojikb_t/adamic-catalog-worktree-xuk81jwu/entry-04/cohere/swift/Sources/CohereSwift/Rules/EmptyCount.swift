import SwiftOperators
import SwiftSyntax

/*
 No comparing a collection's `count` to zero: `items.count == 0` is `items.isEmpty`, and `items.count > 0` is
 `!items.isEmpty`. `isEmpty` says the question being asked, and it is constant time on every collection,
 where `count` walks a `String` or a lazy collection to the end to answer it.

 The shapes are SwiftLint's `empty_count`: `count` compared with the literal `0` by `==`, `!=`, `<`, `<=`,
 `>` or `>=`, on either side, written as a member (`items.count`) or bare (`count`, inside a collection's own
 extension). SwiftLint judges the spelling, so it also flags a `count` of ours that is not a collection's (a
 tally, a retry counter) where `isEmpty` does not exist. This rule flags only a `count` the compiler resolved
 to a standard library declaration (`Array.count`, `Collection.count`, `String.count`, `Set.count`), read from
 the index the build wrote. That is a difference in our favour on every non-collection `count`, and one miss
 accepted: a collection type of our own that declares its own `count` resolves to ours, not the standard
 library's, and is not flagged.

 The comparison is read from an operator-folded copy of the tree, so `total + items.count == 0` is judged as
 the sum compared with zero, which is not this shape, rather than as `items.count == 0`. Folding moves no
 token, so positions in the copy are positions in the file.
 */
public struct EmptyCount: TypedFileRule {
    public let name = "cohere-swift/empty-count"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "empty_count"

    public init() {}

    static let comparisons: Set<String> = ["==", "!=", "<", "<=", ">", ">="]

    /* A `count` and a `0` anywhere in the file: nothing else can hold this shape. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("count") && file.source.contains("0")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        return visitor.found.compactMap { comparison, count in
            let location = file.locations.location(for: count.positionAfterSkippingLeadingTrivia)
            guard let resolved = symbols.reference(line: location.line, column: location.column),
                resolved.isStandardLibrary
            else { return nil }
            return file.finding(
                at: comparison,
                rule: name,
                messageId: "emptyCount",
                message:
                    "Comparing a collection's count to zero asks whether it is empty in a roundabout way. Use isEmpty (or !isEmpty): it says what is meant, and it is constant time where count may walk the whole collection.",
            )
        }
    }

    /* Every comparison of a `count` with the literal zero, and the `count` token the compiler resolved. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [(InfixOperatorExprSyntax, TokenSyntax)] = []

        override func visit(_ node: InfixOperatorExprSyntax) -> SyntaxVisitorContinueKind {
            guard let operation = node.operator.as(BinaryOperatorExprSyntax.self),
                EmptyCount.comparisons.contains(operation.operator.text)
            else {
                return .visitChildren
            }
            if Self.isZero(node.rightOperand), let count = Self.countToken(node.leftOperand) {
                found.append((node, count))
            }
            else if Self.isZero(node.leftOperand), let count = Self.countToken(node.rightOperand) {
                found.append((node, count))
            }
            return .visitChildren
        }

        static func isZero(_ expression: ExprSyntax) -> Bool {
            expression.as(IntegerLiteralExprSyntax.self)?.literal.text == "0"
        }

        /* `items.count`, `self.items.count` or a bare `count`: the name token, which the index places. */
        static func countToken(_ expression: ExprSyntax) -> TokenSyntax? {
            if let member = expression.as(MemberAccessExprSyntax.self), member.declName.baseName.text == "count",
                member.declName.argumentNames == nil
            {
                return member.declName.baseName
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self), reference.baseName.text == "count",
                reference.argumentNames == nil
            {
                return reference.baseName
            }
            return nil
        }
    }
}
