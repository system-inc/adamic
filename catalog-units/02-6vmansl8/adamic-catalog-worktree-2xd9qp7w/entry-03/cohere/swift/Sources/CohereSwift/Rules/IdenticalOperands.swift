import SwiftOperators
import SwiftSyntax

/*
 No comparison of an expression with itself: `a == a`, `x.count < x.count`, `lhs.name != lhs.name`. Both
 sides read the same value, so the comparison can only ever give one answer, and the code reads as a
 condition while deciding nothing. It is nearly always a typo for the other operand (`lhs.name ==
 rhs.name`). The repair is to compare against the value that was meant.

 Covers the comparison operators SwiftLint's `identical_operands` covers: `==`, `!=`, `===`, `!==`, `<`,
 `<=`, `>`, `>=`. Two operands are the same when their tokens are the same, ignoring whitespace and
 comments, so `1    + 1 == 1 + 1` and a call split across lines are both found.

 The engine's shared tree is not operator-folded, so a comparison is one flat `SequenceExprSyntax` and its
 operands have no edges. The rule folds its own copy with the standard operator table before walking it.
 Folding moves no token, so every position in the folded copy is the position in the file. An operator
 the standard table does not know (a package's own `<=>`) still folds, but at a precedence swift-syntax
 guesses, so `a <=> b == a <=> b` could be grouped in a way the compiler does not. A comparison is
 therefore skipped when its flat sequence in the file holds any infix operator the table does not know.
 That is a miss and never a wrong finding; a parenthesised part is its own sequence and is still read.

 Floating point `x == x` and `x != x` are found, as SwiftLint finds them, although they are the old NaN
 test and do vary. That idiom spends the reader's attention on a trick. `x.isNaN` says the same thing in
 words, so the finding has a repair, and the message names it.

 A test that checks a type's `==` is reflexive (`#expect(value == value)`) is found too, as SwiftLint
 finds `XCTAssertTrue(s3 == s3)`. The repair is to compare against a copy, `let copy = value`, which tests
 the same law without looking like a typo.

 Exemption, where we differ from SwiftLint on purpose: an operand that calls a function, expands a macro,
 or awaits is skipped. Two calls are two evaluations and may return two values, so `iterator.next() ==
 iterator.next()` compares neighbours and `#expect(UUID() != UUID())` tests uniqueness; neither is a
 typo. SwiftLint flags `f(i: 2) == f(i: 2)`; we accept that miss rather than report code that is right.
 Property reads and subscripts are still found: a getter that answers differently on two reads in one
 expression is not one the rule should protect.
 */
public struct IdenticalOperands: FileRule {
    public let name = "cohere-swift/identical-operands"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "identical_operands"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        /* Unknown operators are reported to this handler and left unfolded; the rule skips them instead of failing the file. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(original: file.tree)
        visitor.walk(folded)
        return visitor.found.map { comparison in
            file.finding(
                at: comparison,
                rule: name,
                messageId: "identicalOperands",
                message:
                    "Both sides of this comparison are the same expression, so it can only ever give one answer and one side is probably a typo. Compare against the value you meant. If this checks for NaN, write value.isNaN instead.",
            )
        }
    }

    /* Collects every comparison whose two operands are the same tokens. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [InfixOperatorExprSyntax] = []

        /* The file's own unfolded tree, where each comparison's flat sequence can still be read whole. */
        let original: SourceFileSyntax

        static let comparisonOperators: Set<String> = ["==", "!=", "===", "!==", "<", "<=", ">", ">="]

        init(original: SourceFileSyntax) {
            self.original = original
            super.init(viewMode: .sourceAccurate)
        }

        override func visit(_ node: InfixOperatorExprSyntax) -> SyntaxVisitorContinueKind {
            guard let binaryOperator = node.operator.as(BinaryOperatorExprSyntax.self),
                Self.comparisonOperators.contains(binaryOperator.operator.text)
            else {
                return .visitChildren
            }
            if Self.sameTokens(node.leftOperand, node.rightOperand), !Self.mayVary(node.leftOperand),
                foldedWithKnownOperators(binaryOperator.operator)
            {
                found.append(node)
            }
            return .visitChildren
        }

        /*
         Whether every infix operator in the comparison's flat sequence is one the standard table knows, so
         the fold grouped it the way the compiler does. Found by the operator's position in the unfolded tree.
         */
        func foldedWithKnownOperators(_ comparison: TokenSyntax) -> Bool {
            guard let token = original.token(at: comparison.positionAfterSkippingLeadingTrivia),
                let sequence = token.parent?.parent?.parent?.as(SequenceExprSyntax.self)
            else {
                return false
            }
            return sequence.elements.allSatisfy { element in
                guard let binaryOperator = element.as(BinaryOperatorExprSyntax.self) else { return true }
                return OperatorTable.standardOperators.infixOperator(named: binaryOperator.operator.text) != nil
            }
        }

        /* Token for token, kinds and text, with trivia ignored. */
        static func sameTokens(_ left: ExprSyntax, _ right: ExprSyntax) -> Bool {
            let leftTokens = left.tokens(viewMode: .sourceAccurate).map(\.tokenKind)
            let rightTokens = right.tokens(viewMode: .sourceAccurate).map(\.tokenKind)
            return leftTokens == rightTokens
        }

        /* Whether evaluating the operand twice can give two values by design: a call, a macro, or a suspension. */
        static func mayVary(_ operand: ExprSyntax) -> Bool {
            let finder = EvaluationFinder(viewMode: .sourceAccurate)
            finder.walk(operand)
            return finder.foundEvaluation
        }
    }

    /* Finds a call, a macro expansion or an `await` anywhere inside one operand. */
    final class EvaluationFinder: SyntaxVisitor {
        private(set) var foundEvaluation = false

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            foundEvaluation = true
            return .skipChildren
        }

        override func visit(_ node: MacroExpansionExprSyntax) -> SyntaxVisitorContinueKind {
            foundEvaluation = true
            return .skipChildren
        }

        override func visit(_ node: AwaitExprSyntax) -> SyntaxVisitorContinueKind {
            foundEvaluation = true
            return .skipChildren
        }
    }
}
