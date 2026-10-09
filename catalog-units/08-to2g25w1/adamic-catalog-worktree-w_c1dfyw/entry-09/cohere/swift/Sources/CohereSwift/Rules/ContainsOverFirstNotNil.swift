import SwiftOperators
import SwiftSyntax

/*
 No searching a collection for an element only to ask whether the search found one:
 `items.first(where: { $0.isActive }) != nil` is `items.contains(where: { $0.isActive })`, and
 `items.firstIndex(of: target) == nil` is `!items.contains(target)`. `contains` says the question being asked, a yes or no, where `first` and `firstIndex` hand back an element or
 an index that the comparison then throws away, so the reader has to work out that nothing else wanted it.

 The shape is SwiftLint's `contains_over_first_not_nil`: a call named `first` or `firstIndex`, compared with
 the `nil` literal by `==` or `!=`. SwiftLint reads the name and not the arguments, so its shape holds every
 standard library call by those names: `first(where:)`, `firstIndex(where:)` and `firstIndex(of:)`, with an
 argument (`first(where: { })`, `first(where: predicate)`) or a trailing closure (`first { }`), on any base
 (`items.map { }.first { }`), or wrapped in one pair of parentheses (`(items.first { }) != nil`). The message
 names the repair for the call it found: `contains(_:)` for `firstIndex(of:)`, `contains(where:)` for the rest,
 each negated for `== nil`. A finding starts where SwiftLint's does, at the start of the call.

 SwiftLint judges the spelling, so it also flags a `first` of ours, on a type that is no collection and may
 have no `contains` at all. This rule flags only a `first` or `firstIndex` the compiler resolved to a standard
 library declaration (`Sequence.first(where:)`, `Collection.firstIndex(where:)`, `Collection.firstIndex(of:)`,
 `Set.firstIndex(of:)`, `AsyncSequence.first(where:)`), read from the index the build wrote. A lazy sequence's
 `first(where:)` is the standard library's too, and is flagged: the reason is what the code says, not what it
 costs, and `contains(where:)` on a lazy sequence stops at the first match just as `first(where:)` does.

 Where we differ on purpose, in the direction of finding more of the same thing: `nil` may stand on either
 side (`nil != items.first { }`), and a bare `first(where:)` inside a collection's own extension is found as
 well as a member one. The comparison is read from an operator-folded copy of the tree, as SwiftLint reads
 it, so `ready && items.first { } != nil` is found. Folding moves no token, so positions in the copy are
 positions in the file.

 Misses, accepted: a `first` declared on a type of ours resolves to ours and is not flagged. A call reached
 through optional chaining (`items?.first { } != nil`, `owner?.items.firstIndex(of: target) == nil`) is not
 flagged, where SwiftLint flags it: the chain makes the result optional, so the repair is
 `items?.contains { } == true`, which trades one comparison for another and is no clearer. `=== nil` is not
 the shape. Overlap with a sibling: `items.filter { }.first != nil` reads the `first` property, not a call, so
 it is not this shape; `first-where` reports it, and its repair, `first(where:) != nil`, is then this one.
 `items.filter { a }.first { b } != nil` is both rules' shape, and both report it: `first-where` folds the two
 predicates into one search, and this rule turns that search into `contains(where:)`.
 */
public struct ContainsOverFirstNotNil: TypedFileRule {
    public let name = "cohere-swift/contains-over-first-not-nil"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "contains_over_first_not_nil"

    public init() {}

    static let comparisons: Set<String> = ["==", "!="]

    static let methods: Set<String> = ["first", "firstIndex"]

    /* A `first` and a `nil` anywhere in the file: `firstIndex` starts with `first`, and nothing else can hold this shape. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("first") && file.source.contains("nil")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        return visitor.found.compactMap { candidate in
            let location = file.locations.location(for: candidate.method.positionAfterSkippingLeadingTrivia)
            guard let resolved = symbols.reference(line: location.line, column: location.column),
                resolved.isStandardLibrary
            else { return nil }
            let searched =
                candidate.call.arguments.first?.label?.text == "of"
                ? "firstIndex(of:)" : "\(candidate.method.text)(where:)"
            let repair = searched == "firstIndex(of:)" ? "contains(_:)" : "contains(where:)"
            return file.finding(
                at: candidate.call,
                rule: name,
                messageId: "containsOverFirstNotNil",
                message:
                    "Comparing \(searched) with nil searches for something only to throw it away and ask whether it was found. Use \(repair) (or !\(repair) for == nil): it asks the yes or no question that is meant.",
            )
        }
    }

    /* One shape found: the `first` or `firstIndex` call, and its name token, which the index places. */
    struct Candidate {
        var call: FunctionCallExprSyntax
        var method: TokenSyntax
    }

    /* Every `first` or `firstIndex` call compared with `nil` by `==` or `!=`, on either side. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Candidate] = []

        override func visit(_ node: InfixOperatorExprSyntax) -> SyntaxVisitorContinueKind {
            guard let operation = node.operator.as(BinaryOperatorExprSyntax.self),
                ContainsOverFirstNotNil.comparisons.contains(operation.operator.text)
            else {
                return .visitChildren
            }
            if node.rightOperand.is(NilLiteralExprSyntax.self), let candidate = Self.searchCall(node.leftOperand) {
                found.append(candidate)
            }
            else if node.leftOperand.is(NilLiteralExprSyntax.self), let candidate = Self.searchCall(node.rightOperand) {
                found.append(candidate)
            }
            return .visitChildren
        }

        /* The call of `items.first { }`, `firstIndex(of: target)`, or either inside one pair of parentheses, when no optional chain leads to it. */
        static func searchCall(_ expression: ExprSyntax) -> Candidate? {
            var callExpression = expression
            if let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                callExpression = only.expression
            }
            guard let call = callExpression.as(FunctionCallExprSyntax.self) else { return nil }
            if let member = call.calledExpression.as(MemberAccessExprSyntax.self),
                ContainsOverFirstNotNil.methods.contains(member.declName.baseName.text)
            {
                if let base = member.base, isOptionalChain(base) {
                    return nil
                }
                return Candidate(call: call, method: member.declName.baseName)
            }
            if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self),
                ContainsOverFirstNotNil.methods.contains(reference.baseName.text)
            {
                return Candidate(call: call, method: reference.baseName)
            }
            return nil
        }

        /* Whether a `?` sits anywhere along this postfix chain, which makes everything read from it optional. */
        static func isOptionalChain(_ expression: ExprSyntax) -> Bool {
            if expression.is(OptionalChainingExprSyntax.self) {
                return true
            }
            if let member = expression.as(MemberAccessExprSyntax.self), let base = member.base {
                return isOptionalChain(base)
            }
            if let call = expression.as(FunctionCallExprSyntax.self) {
                return isOptionalChain(call.calledExpression)
            }
            if let subscriptCall = expression.as(SubscriptCallExprSyntax.self) {
                return isOptionalChain(subscriptCall.calledExpression)
            }
            if let forceUnwrap = expression.as(ForceUnwrapExprSyntax.self) {
                return isOptionalChain(forceUnwrap.expression)
            }
            return false
        }
    }
}
