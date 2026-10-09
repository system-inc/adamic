import SwiftSyntax

/*
 No dictionary literal that spells the same key twice: `[.one: 1, .two: 2, .one: 3]`.

 A `Dictionary` built from a literal with a repeated key stops the program at run time ("Fatal error:
 Dictionary literal contains duplicate keys"), and a custom `ExpressibleByDictionaryLiteral` type usually
 keeps one of the two values without a word. The compiler sees neither. The repair is to delete the stale
 entry, or to fix the key that was meant to be different (the usual cause is a copied line whose key was
 never edited).

 Two keys are the same when they are spelled with the same tokens, comments and whitespace aside. This is
 SwiftLint's `duplicated_key_in_dictionary_literal` notion of "same", narrowed to keys whose spelling
 decides their value: literals (numbers, strings, booleans, `nil`, a negated number), names and member
 chains (`foo`, `.one`, `Kind.one`, `self.base.name`, `owner?.name`, `owner!.name`), key paths made of
 properties (`\Item.name`), a parenthesized key, and strings that interpolate only such keys. No two
 different spellings are ever claimed equal (`1` and `0x1`, `"A"` and `"\u{41}"`), since that needs
 evaluation.

 Exemptions, each a place where a repeated spelling is not a repeated value or not a mistake:
 - A call (`UUID()`, `make(1)`, or one interpolated into a string), a subscript, a macro (`#line`), a
   closure, an operator expression, or a cast as a key: evaluating it twice may give two values, so a
   repeat proves nothing. SwiftLint exempts calls and macros for the same reason, but still flags a call
   inside an interpolation (`"\(make())"`); we do not.
 - A key whose last name is `nan` or `signalingNaN`: NaN is never equal to itself, so two of them are two
   entries and nothing crashes. SwiftLint flags these, wrongly.
 - A literal whose contextual type is visibly `KeyValuePairs`, which keeps repeated keys on purpose: a
   declaration whose annotation or signature mentions `KeyValuePairs`, an `as KeyValuePairs` cast, or an
   argument to `KeyValuePairs(...)` or `Mirror(...)`. SwiftLint flags these, wrongly. A literal passed to
   some other function's `KeyValuePairs` parameter cannot be seen without types and is a known false
   positive; none exists in either proving ground.

 Where we flag and SwiftLint 0.65.1 does not, measured on a probe file, the key really is repeated and
 the literal really crashes: `true` twice, a key path twice, and `Kind.one` against `Kind . one`.

 The finding is on the second and every later occurrence of a key, the entry to delete or correct.
 Report only.
 */
public struct DuplicatedKeyInDictionaryLiteral: FileRule {
    public let name = "cohere-swift/duplicated-key-in-dictionary-literal"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "duplicated_key_in_dictionary_literal"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { key in
            file.finding(
                at: key,
                rule: name,
                messageId: "duplicatedDictionaryKey",
                message:
                    "This key already appears earlier in the same dictionary literal, which crashes at run time or silently drops one of the values. Delete this entry, or correct the key that was meant to be different.",
            )
        }
    }

    /* Collects every key that repeats an earlier key of the same literal. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [ExprSyntax] = []

        override func visit(_ node: DictionaryElementListSyntax) -> SyntaxVisitorContinueKind {
            if let dictionary = node.parent?.as(DictionaryExprSyntax.self), Self.isKeyValuePairs(dictionary) {
                return .visitChildren
            }
            var seen: Set<String> = []
            for element in node {
                guard Self.isSpellingDecided(element.key), !Self.isNotANumber(element.key) else { continue }
                let spelling = Self.spelling(of: element.key)
                if seen.contains(spelling) {
                    found.append(element.key)
                }
                else {
                    seen.insert(spelling)
                }
            }
            return .visitChildren
        }

        /* The key's tokens without their trivia, so `foo.bar` and `foo . bar` read the same. */
        static func spelling(of expression: ExprSyntax) -> String {
            expression.tokens(viewMode: .sourceAccurate).map(\.text).joined(separator: " ")
        }

        /* Whether two keys with this spelling are certain to be the same value. */
        static func isSpellingDecided(_ expression: ExprSyntax) -> Bool {
            if expression.is(IntegerLiteralExprSyntax.self) || expression.is(FloatLiteralExprSyntax.self)
                || expression.is(BooleanLiteralExprSyntax.self) || expression.is(NilLiteralExprSyntax.self)
            {
                return true
            }
            if let string = expression.as(StringLiteralExprSyntax.self) {
                return string.segments.allSatisfy { segment in
                    guard case .expressionSegment(let interpolation) = segment else { return true }
                    guard interpolation.expressions.count == 1, let only = interpolation.expressions.first,
                        only.label == nil
                    else {
                        return false
                    }
                    return isSpellingDecided(only.expression)
                }
            }
            if let negated = expression.as(PrefixOperatorExprSyntax.self) {
                return negated.operator.text == "-"
                    && (negated.expression.is(IntegerLiteralExprSyntax.self)
                        || negated.expression.is(FloatLiteralExprSyntax.self))
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self) {
                return reference.argumentNames == nil
            }
            if let member = expression.as(MemberAccessExprSyntax.self) {
                guard member.declName.argumentNames == nil else { return false }
                return member.base.map(isSpellingDecided) ?? true
            }
            if let keyPath = expression.as(KeyPathExprSyntax.self) {
                return keyPath.components.allSatisfy { component in
                    guard case .property(let property) = component.component else { return false }
                    return property.declName.argumentNames == nil && property.genericArgumentClause == nil
                }
            }
            if let chained = expression.as(OptionalChainingExprSyntax.self) {
                return isSpellingDecided(chained.expression)
            }
            if let unwrapped = expression.as(ForceUnwrapExprSyntax.self) {
                return isSpellingDecided(unwrapped.expression)
            }
            if let parenthesized = expression.as(TupleExprSyntax.self), parenthesized.elements.count == 1,
                let only = parenthesized.elements.first, only.label == nil
            {
                return isSpellingDecided(only.expression)
            }
            return false
        }

        /* `.nan`, `Double.nan`, `.signalingNaN`: never equal to themselves, so never a duplicate. */
        static func isNotANumber(_ expression: ExprSyntax) -> Bool {
            var inner = expression
            while let parenthesized = inner.as(TupleExprSyntax.self), parenthesized.elements.count == 1,
                let only = parenthesized.elements.first
            {
                inner = only.expression
            }
            let lastName =
                inner.as(MemberAccessExprSyntax.self)?.declName.baseName.text
                ?? inner.as(DeclReferenceExprSyntax.self)?.baseName.text
            return lastName == "nan" || lastName == "signalingNaN"
        }

        /* Whether the literal's contextual type is visibly `KeyValuePairs`, which allows repeated keys. */
        static func isKeyValuePairs(_ dictionary: DictionaryExprSyntax) -> Bool {
            if let sequence = dictionary.parent?.as(ExprListSyntax.self) {
                let parts = Array(sequence)
                if let index = parts.firstIndex(where: { $0.id == dictionary.id }), index + 2 < parts.count,
                    parts[index + 1].is(UnresolvedAsExprSyntax.self), mentionsKeyValuePairs(parts[index + 2])
                {
                    return true
                }
            }
            if let call = dictionary.parent?.parent?.parent?.as(FunctionCallExprSyntax.self),
                dictionary.parent?.is(LabeledExprSyntax.self) == true
            {
                if call.calledExpression.trimmedDescription == "Mirror" || mentionsKeyValuePairs(call.calledExpression)
                {
                    return true
                }
            }
            var ancestor = dictionary.parent
            while let current = ancestor {
                if let binding = current.as(PatternBindingSyntax.self) {
                    return binding.typeAnnotation.map { mentionsKeyValuePairs($0) } ?? false
                }
                if let function = current.as(FunctionDeclSyntax.self) {
                    return mentionsKeyValuePairs(function.signature)
                }
                if let subscriptDeclaration = current.as(SubscriptDeclSyntax.self) {
                    return mentionsKeyValuePairs(subscriptDeclaration.returnClause)
                }
                if let closure = current.as(ClosureExprSyntax.self), let signature = closure.signature {
                    return mentionsKeyValuePairs(signature)
                }
                if current.is(DeclSyntax.self) {
                    return false
                }
                ancestor = current.parent
            }
            return false
        }

        static func mentionsKeyValuePairs(_ node: some SyntaxProtocol) -> Bool {
            node.tokens(viewMode: .sourceAccurate).contains { $0.text == "KeyValuePairs" }
        }
    }
}
