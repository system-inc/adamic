import SwiftSyntax

/*
 No optional binding that binds nothing: `if let _ = value`, `guard var _ = value`, `while let _ = next()`,
 or a tuple pattern made only of `_`, such as `if let (_, _) = pair`.

 The binding throws its value away the moment it makes it, so the condition is a nil check wearing the
 clothes of an unwrap. A reader has to work that out; `value != nil` says it. The repair depends on what the
 right side is, and the message names the one that fits:

 - plain: `value != nil`.
 - `try? work()`: the question is whether `work()` succeeded. `(try? work()) != nil` asks it, with the
   parentheses, because `try? work() != nil` covers the comparison and yields a `Bool?`. If the failure
   deserves a reason, `do`/`catch` says it. This is the only report such a line gets: cohere's
   `correctness-no-discarded-try-optional` reports a `try?` only when it is a whole statement or the right side of
   `_ =`, never inside a condition, so the two rules cannot both report one `try?`.
 - `value as? Type`: the question is a type test, and `value is Type` is its name.

 Matches SwiftLint's `unused_optional_binding` with its default `ignore_optional_try: false`: `let` and
 `var`, in `if`, `guard` and `while` conditions, the wildcard pattern and tuple patterns made entirely of
 wildcards at any depth (the empty tuple `()` among them, since it binds nothing either), with or
 without a type annotation or tuple labels. The finding sits where SwiftLint puts it, on the pattern.
 A type annotation does not exempt the line: `if let _: Model = decode()` is still a nil check, spelled
 `(decode() as Model?) != nil` when the annotation is what picks the overload.

 Not reported, each because the line is not the shape this rule is about:
 - a tuple pattern that binds any name, such as `let (_, second) = pair`. Something is used.
 - `if case let _ = value` and `if case _? = value`. These are pattern matches, not optional binding, and
   SwiftLint does not report them either.
 - `let _ = work()` as a statement. That is a discard, not a condition, and other rules own it.
 - a backticked `` `_` ``, which the parser reads as a name rather than a wildcard. SwiftLint agrees.
 */
public struct UnusedOptionalBinding: FileRule {
    public let name = "cohere-swift/unused-optional-binding"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "unused_optional_binding"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { found in
            switch found.repair {
                case .nilCheck:
                    file.finding(
                        at: found.pattern,
                        rule: name,
                        messageId: "unusedOptionalBinding",
                        message:
                            "This binding unwraps a value and throws it away, so it is only a nil check in disguise. Write value != nil instead.",
                    )
                case .successCheck:
                    file.finding(
                        at: found.pattern,
                        rule: name,
                        messageId: "unusedOptionalBindingTry",
                        message:
                            "This binding keeps nothing from try?, so it only asks whether the call succeeded. Write (try? call()) != nil, with the parentheses, or use do/catch if the failure needs a reason.",
                    )
                case .typeCheck:
                    file.finding(
                        at: found.pattern,
                        rule: name,
                        messageId: "unusedOptionalBindingCast",
                        message:
                            "This binding keeps nothing from as?, so it only asks what type the value is. Write value is Type instead.",
                    )
            }
        }
    }

    /* Which honest condition the binding stands in for, read from the right side of its `=`. */
    enum Repair {
        case nilCheck
        case successCheck
        case typeCheck
    }

    struct Found {
        var pattern: PatternSyntax
        var repair: Repair
    }

    /* Collects the pattern of every optional-binding condition whose pattern binds no name. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Found] = []

        override func visit(_ node: OptionalBindingConditionSyntax) -> SyntaxVisitorContinueKind {
            if Self.bindsNothing(node.pattern) {
                found.append(Found(pattern: node.pattern, repair: Self.repair(for: node.initializer?.value)))
            }
            return .visitChildren
        }

        /*
         `_`, or a tuple whose every element is itself nothing-binding, `()` included. The expression spellings
         are the shape an older parser gave the same source, kept so the rule reads either tree.
         */
        static func bindsNothing(_ pattern: PatternSyntax) -> Bool {
            if pattern.is(WildcardPatternSyntax.self) {
                return true
            }
            if let tuple = pattern.as(TuplePatternSyntax.self) {
                return tuple.elements.allSatisfy { bindsNothing($0.pattern) }
            }
            if let expressionPattern = pattern.as(ExpressionPatternSyntax.self) {
                return isDiscardExpression(expressionPattern.expression)
            }
            return false
        }

        static func isDiscardExpression(_ expression: ExprSyntax) -> Bool {
            if expression.is(DiscardAssignmentExprSyntax.self) {
                return true
            }
            if let tuple = expression.as(TupleExprSyntax.self) {
                return tuple.elements.allSatisfy { isDiscardExpression($0.expression) }
            }
            return false
        }

        /*
         `try? …` asks for success and `… as? Type` asks for a type; anything else, including plain `try`
         and `await`, is a nil check. The tree is unfolded, so a lone `as?` is a three-part sequence.
         */
        static func repair(for value: ExprSyntax?) -> Repair {
            guard let value else { return .nilCheck }
            if let tryExpression = value.as(TryExprSyntax.self),
                tryExpression.questionOrExclamationMark?.tokenKind == .postfixQuestionMark
            {
                return .successCheck
            }
            if let sequence = value.as(SequenceExprSyntax.self), sequence.elements.count == 3 {
                let parts = Array(sequence.elements)
                if let cast = parts[1].as(UnresolvedAsExprSyntax.self),
                    cast.questionOrExclamationMark?.tokenKind == .postfixQuestionMark
                {
                    return .typeCheck
                }
            }
            return .nilCheck
        }
    }
}
