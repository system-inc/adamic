import SwiftSyntax

/*
 No building an intersection only to ask whether it is empty: `left.intersection(right).isEmpty` is
 `left.isDisjoint(with: right)`. `isDisjoint(with:)` says the question being asked, where the intersection spells
 out one way of computing the answer, and on a `Set` it answers without building anything, stopping at the first
 common element, where `intersection` builds a whole new set of every common element only to throw it away.

 The shape is SwiftLint's `is_disjoint`: `isEmpty` read from a call of a member named `intersection`, written
 directly (`left.intersection(right).isEmpty`) or inside one pair of parentheses
 (`(left.intersection(right)).isEmpty`), with or without a `!` in front. The finding starts where SwiftLint's
 does, at the `intersection` name.

 SwiftLint judges the spelling, so it also flags an `intersection` of ours, on a type that may have no
 `isDisjoint(with:)` at all, and CoreGraphics' `CGRect.intersection`, whose rectangle has none. This rule flags
 only an `intersection` the compiler resolved, in the index the build wrote, to one whose type has the repair:
 `Set`'s (with a set or any sequence, both of which have an `isDisjoint(with:)` taking the same argument), the
 `SetAlgebra` requirement on a generic, `OptionSet`'s, `RangeSet`'s, and Foundation's `IndexSet` and
 `CharacterSet`, which are `SetAlgebra` types. The symbols are matched whole, type and member, so a member our
 own module adds to `Set` is not taken for the standard library's. `RangeSet`'s repair is spelled
 `isDisjoint(_:)`, and the message names that spelling for it. On an option set, a generic, an `IndexSet` or a
 `CharacterSet`, `isDisjoint(with:)` is `SetAlgebra`'s default, which is `intersection(other).isEmpty`: the
 rewrite runs the same and reads as what is meant, and the message claims no more than that. Where we differ on
 purpose, in the direction of finding more of the same thing: a bare `intersection(other).isEmpty` inside a
 set's own extension is found as well as a member one.

 Out, decided: `NSSet` has no `intersection` to flag. Its question is `intersects(_:)`, already a yes or no, and
 `NSMutableSet.intersect(_:)` mutates and returns nothing. `DateInterval.intersection(with:)` and
 `NSRange.intersection(_:)` return an optional, which has no `isEmpty`.

 Misses, accepted: an `intersection` declared on a type of ours resolves to ours and is not flagged, even when the
 type is a `SetAlgebra` with an `isDisjoint(with:)`; `left.intersection(right).count == 0` asks the same question
 and is not SwiftLint's shape here, though `empty-count` reports it with `isEmpty` as its repair; and
 `left.intersection(right) == []` is left alone.
 */
public struct IsDisjoint: TypedFileRule {
    public let name = "cohere-swift/is-disjoint"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "is_disjoint"

    public init() {}

    /*
     Every `intersection` whose type has the repair, as its symbol name begins: `Set`'s, the `SetAlgebra`
     requirement, `OptionSet`'s, `RangeSet`'s, and Foundation's `IndexSet`'s and `CharacterSet`'s.
     */
    static let intersections = [
        "s:Sh12intersection",
        "s:s10SetAlgebraP12intersection",
        "s:s9OptionSetPsE12intersection",
        "s:s8RangeSetV12intersection",
        "s:10Foundation8IndexSetV12intersection",
        "s:10Foundation12CharacterSetV12intersection",
    ]

    /* An `intersection` and an `isEmpty` anywhere in the file: nothing else can hold this shape. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("intersection") && file.source.contains("isEmpty")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.compactMap { intersection in
            let location = file.locations.location(for: intersection.positionAfterSkippingLeadingTrivia)
            guard let resolved = symbols.reference(line: location.line, column: location.column),
                Self.intersections.contains(where: { resolved.symbol.hasPrefix($0) })
            else {
                return nil
            }
            let repair = resolved.symbol.hasPrefix("s:s8RangeSetV") ? "isDisjoint(_:)" : "isDisjoint(with:)"
            return file.finding(
                at: intersection,
                rule: name,
                messageId: "isDisjoint",
                message:
                    "Asking whether an intersection is empty spells out how to compute the answer rather than the question. Use \(repair): it says what is meant, and on a Set it stops at the first common element instead of building the whole intersection.",
            )
        }
    }

    /* The `intersection` name token of every `isEmpty` read from an `intersection` call. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [TokenSyntax] = []

        override func visit(_ node: MemberAccessExprSyntax) -> SyntaxVisitorContinueKind {
            if node.declName.baseName.text == "isEmpty", node.declName.argumentNames == nil, let base = node.base,
                let intersection = Self.intersectionToken(base)
            {
                found.append(intersection)
            }
            return .visitChildren
        }

        /* The `intersection` name token of `left.intersection(right)`, `intersection(right)`, or either inside one pair of parentheses. */
        static func intersectionToken(_ expression: ExprSyntax) -> TokenSyntax? {
            var callExpression = expression
            if let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                callExpression = only.expression
            }
            guard let call = callExpression.as(FunctionCallExprSyntax.self) else { return nil }
            if let member = call.calledExpression.as(MemberAccessExprSyntax.self),
                member.declName.baseName.text == "intersection"
            {
                return member.declName.baseName
            }
            if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self),
                reference.baseName.text == "intersection"
            {
                return reference.baseName
            }
            return nil
        }
    }
}
