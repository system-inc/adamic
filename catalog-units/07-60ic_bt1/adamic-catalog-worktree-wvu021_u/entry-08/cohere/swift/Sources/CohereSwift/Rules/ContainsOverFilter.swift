import SwiftOperators
import SwiftSyntax

/*
 No filtering a collection only to ask whether anything passed: `items.filter { $0.isActive }.isEmpty` is
 `!items.contains { $0.isActive }`, and `items.filter { $0.isActive }.count > 0` is `items.contains { $0.isActive }`.
 `contains(where:)` says the question being asked and stops at the first match, where `filter` runs the
 predicate on every element and builds a whole new array (or dictionary, set or string) only to count it and
 throw it away.

 The shapes are SwiftLint's `contains_over_filter_count` and `contains_over_filter_is_empty`. The first is a
 `filter` call's `count` compared with a zero literal by `==`, `!=` or `>`; the second is `isEmpty` read from
 a `filter` call. Both take the call with a trailing closure (`filter { }`), an argument (`filter(predicate)`,
 `filter({ })`), or wrapped in one pair of parentheses (`(items.filter { }).isEmpty`). A zero is read as
 SwiftLint reads it, so `0`, `00`, `0x0` and `0_0` are zero and `01` is not. The two keep their own message
 ids, `filterCount` and `filterIsEmpty`, so each finding maps to its SwiftLint rule.

 SwiftLint judges the spelling, so it also flags a `filter` of ours, on a type that is no collection and may
 have no `contains(where:)` at all. This rule flags only a `filter` the compiler resolved to a standard
 library declaration (`Sequence.filter`, `Array.filter`, `Dictionary.filter`, `Set.filter`, `String.filter`),
 and for the count shape only a `count` that resolved there too, read from the index the build wrote. A lazy
 `filter` (`items.lazy.filter { }`) is not flagged although it is the standard library's: it builds nothing,
 and its `isEmpty` already stops at the first match, so the reason does not hold. The rule tells it by
 `Lazy` in the symbol name, which every lazy `filter` declaration carries.

 Where we differ on purpose, in the direction of finding more of the same thing: the comparison is read from
 an operator-folded copy of the tree, so `items.filter { }.count > 0 && ready` is found, where SwiftLint reads
 only a comparison that is the whole expression. Zero may stand on either side (`0 < items.filter { }.count`,
 `0 == ...`, `0 != ...`), as it may in `empty_count`. A bare `filter` inside a collection's own extension
 (`filter { }.isEmpty`) is found as well as a member one. Folding moves no token, so positions in the copy are
 positions in the file, and a finding starts where SwiftLint's does: at the comparison, or at the `isEmpty`
 read.

 Misses, accepted: a `filter` declared on a type of ours resolves to ours and is not flagged, even when it
 builds an array; `count >= 1`, `count < 1` and `count <= 0` ask the same question and are not SwiftLint's
 shapes, so they are left alone; a lazy `filter`'s `count` compared with zero walks the whole sequence where
 `contains(where:)` would stop, and is skipped with the rest of the lazy case. A `filter { }.count == 0` on an
 array is also `empty_count`'s shape, and that rule reports it too: its repair is `isEmpty` on the filtered
 array, this one's is no array at all.
 */
public struct ContainsOverFilter: TypedFileRule {
    public let name = "cohere-swift/contains-over-filter"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "contains_over_filter_count"

    public init() {}

    static let comparisons: Set<String> = ["==", "!=", ">"]

    /* The comparison read with zero on the left, so `0 < count` is `count > 0`. */
    static let mirroredComparisons: Set<String> = ["==", "!=", "<"]

    /* A `filter` anywhere in the file: both shapes start with one. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("filter")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        return visitor.found.compactMap { candidate in
            guard Self.resolvesToStandardLibrary(candidate.filter, in: file, symbols: symbols, excludingLazy: true)
            else { return nil }
            if let count = candidate.count {
                guard Self.resolvesToStandardLibrary(count, in: file, symbols: symbols, excludingLazy: false) else {
                    return nil
                }
                return file.finding(
                    at: candidate.node,
                    rule: name,
                    messageId: "filterCount",
                    message:
                        "Counting a filtered collection to compare with zero builds a whole new collection to ask whether anything matched. Use contains(where:) (or !contains(where:) for == 0): it says what is meant, and it stops at the first match.",
                )
            }
            return file.finding(
                at: candidate.node,
                rule: name,
                messageId: "filterIsEmpty",
                message:
                    "Asking whether a filtered collection is empty builds a whole new collection to ask whether anything matched. Use !contains(where:) (or contains(where:) for !isEmpty): it says what is meant, and it stops at the first match.",
            )
        }
    }

    /* Whether the name at this token is a written reference to a standard library declaration, and, when asked, not a lazy one. */
    static func resolvesToStandardLibrary(
        _ token: TokenSyntax,
        in file: ParsedFile,
        symbols: FileSymbols,
        excludingLazy: Bool,
    ) -> Bool {
        let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
        guard let resolved = symbols.reference(line: location.line, column: location.column), resolved.isStandardLibrary
        else { return false }
        return !(excludingLazy && resolved.symbol.contains("Lazy"))
    }

    /* One shape found: the node a finding starts at, the `filter` token, and the `count` token for the count shape (nil for `isEmpty`). */
    struct Candidate {
        var node: ExprSyntax
        var filter: TokenSyntax
        var count: TokenSyntax?
    }

    /* Every `filter` call's `count` compared with zero, and every `filter` call's `isEmpty`, with the name tokens the index places. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Candidate] = []

        override func visit(_ node: InfixOperatorExprSyntax) -> SyntaxVisitorContinueKind {
            guard let operation = node.operator.as(BinaryOperatorExprSyntax.self) else { return .visitChildren }
            let comparison = operation.operator.text
            if ContainsOverFilter.comparisons.contains(comparison), Self.isZero(node.rightOperand),
                let tokens = Self.filterCountTokens(node.leftOperand)
            {
                found.append(Candidate(node: ExprSyntax(node), filter: tokens.filter, count: tokens.count))
            }
            else if ContainsOverFilter.mirroredComparisons.contains(comparison), Self.isZero(node.leftOperand),
                let tokens = Self.filterCountTokens(node.rightOperand)
            {
                found.append(Candidate(node: ExprSyntax(node), filter: tokens.filter, count: tokens.count))
            }
            return .visitChildren
        }

        override func visit(_ node: MemberAccessExprSyntax) -> SyntaxVisitorContinueKind {
            if node.declName.baseName.text == "isEmpty", node.declName.argumentNames == nil, let base = node.base,
                let filter = Self.filterToken(base)
            {
                found.append(Candidate(node: ExprSyntax(node), filter: filter, count: nil))
            }
            return .visitChildren
        }

        /* A zero integer literal as SwiftLint reads one: any radix prefix and underscores dropped, then zero in value. */
        static func isZero(_ expression: ExprSyntax) -> Bool {
            guard let literal = expression.as(IntegerLiteralExprSyntax.self) else { return false }
            var digits = literal.literal.text.lowercased()
            for prefix in ["0x", "0o", "0b"] where digits.hasPrefix(prefix) {
                digits.removeFirst(prefix.count)
            }
            digits.removeAll { $0 == "_" }
            return !digits.isEmpty && digits.allSatisfy { $0 == "0" }
        }

        /* `<filter call>.count`: the `filter` token and the `count` token. */
        static func filterCountTokens(_ expression: ExprSyntax) -> (filter: TokenSyntax, count: TokenSyntax)? {
            guard let member = expression.as(MemberAccessExprSyntax.self), member.declName.baseName.text == "count",
                member.declName.argumentNames == nil,
                let base = member.base, let filter = filterToken(base)
            else {
                return nil
            }
            return (filter, member.declName.baseName)
        }

        /* The `filter` name token of `items.filter { }`, `filter(predicate)`, or either inside one pair of parentheses. */
        static func filterToken(_ expression: ExprSyntax) -> TokenSyntax? {
            var callExpression = expression
            if let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                callExpression = only.expression
            }
            guard let call = callExpression.as(FunctionCallExprSyntax.self) else { return nil }
            if let member = call.calledExpression.as(MemberAccessExprSyntax.self),
                member.declName.baseName.text == "filter"
            {
                return member.declName.baseName
            }
            if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self),
                reference.baseName.text == "filter"
            {
                return reference.baseName
            }
            return nil
        }
    }
}
