import SwiftSyntax

/*
 No filtering a collection only to take its first element: `items.filter { $0.isActive }.first` is
 `items.first { $0.isActive }`. `first(where:)` says the question being asked and stops at the first match,
 where `filter` runs the predicate on every element and builds a whole new array (or dictionary, set or
 string) only to keep one element of it and throw the rest away.

 The shape is SwiftLint's `first_where`: `first` read from a member `filter` call, with a trailing closure
 (`items.filter { }.first`), an argument (`filter(predicate)`, `filter({ })`), or wrapped in one pair of
 parentheses (`(items.filter { }).first`), anywhere in a chain (`items.map { }.filter { }.first`), followed by
 anything (`.first?.name`, `.first!`), and on a line break (`filter { }\n.first`). The finding starts where
 SwiftLint's does, at the start of the `filter` call. SwiftLint also reads `first(where:)` called on a
 `filter`'s result as the shape, and so does this rule: `items.filter { a }.first { b }` still builds the
 array, and the repair is one `first(where:)` with both predicates.

 SwiftLint judges the spelling, so it skips by spelling too: a `filter` whose argument is a string literal or
 an `NSPredicate(...)` call, which is Realm's query `filter`, not a collection's. This rule asks the types
 instead, and flags only a `filter` the compiler resolved to one of the standard library declarations that
 build a new collection (`Array`'s, `Sequence`'s, `RangeReplaceableCollection`'s, `Substring`'s,
 `Dictionary`'s, `Set`'s), read from the index the build wrote, with a `first` that resolved to the standard
 library too. Realm's `filter`, and any `filter` of ours, resolves to its own module and is never flagged,
 whatever its argument, even when it is added to `Sequence` by an extension (`s:ST7ControlE6filter`, which no
 eager declaration's name starts with); a `first` an extension of ours adds is not the standard library's
 either; a standard library `filter` cannot take a string or a predicate object at all.

 A lazy `filter` (`items.lazy.filter { }.first`) is the standard library's and is not flagged: it builds
 nothing, and its `first` already stops at the first match, so the reason does not hold. Nor is an
 asynchronous sequence's `filter`, which is lazy in the same way. Both are left out by naming the eager
 declarations rather than excluding the lazy ones, so a declaration this rule has not seen is never flagged.

 On a dictionary or a set either spelling gives some matching element, not a particular one; when several
 match, the repair may give a different one, which neither spelling promised. Misses, accepted: a `filter`
 declared on a type of ours resolves to ours and is not flagged, even when it builds an array; a `filter`
 inside parentheses with `try` or `await` (`(try items.filter { }).first`) is not read, as SwiftLint does not
 read it. Where SwiftLint stops short and this rule does not: a bare `filter` inside a collection's own
 extension (`filter { }.first`) is found as well as a member one.

 Shared with the sibling rules: `items.filter { }.first != nil` is found here, and its repair,
 `first(where:) != nil`, is then `contains-over-first-not-nil`'s shape, whose repair is `contains(where:)`;
 going straight there is the better fix. `last-where` is this rule's twin for `last`, and reads the shape
 through this rule's `filterCalls(reading:filters:in:symbols:)`, so the two cannot drift apart.
 */
public struct FirstWhere: TypedFileRule {
    public let name = "cohere-swift/first-where"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "first_where"

    public init() {}

    /*
     The standard library `filter` declarations that build a new collection, by the start of their symbol
     names: `Array`'s (and `ArraySlice`'s and `ContiguousArray`'s), `Sequence`'s, `RangeReplaceableCollection`'s
     (`String`'s, its views'), `Substring`'s, `Dictionary`'s and `Set`'s. Every receiver of one is a sequence,
     so every one has `first(where:)`.
     */
    static let eagerFilters = [
        "s:s14_ArrayProtocolPsE6filter",
        "s:STsE6filter",
        "s:SmsE6filter",
        "s:Ss6filter",
        "s:SD6filter",
        "s:Sh6filter",
    ]

    /* A `filter` and a `first` anywhere in the file: nothing else can hold this shape. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("filter") && file.source.contains("first")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        Self.filterCalls(reading: "first", filters: Self.eagerFilters, in: file, symbols: symbols).map { call in
            file.finding(
                at: call,
                rule: name,
                messageId: "firstWhere",
                message:
                    "Taking the first element of a filtered collection builds a whole new collection of every match only to keep one. Use first(where:) with the filter's predicate (joined by && with first's own, when it has one): it says what is meant, and it stops at the first match.",
            )
        }
    }

    /*
     Every `filter` call whose `member` (`first`, or `last` for `last-where`) is read, where the `filter`
     resolved to a declaration named in `filters` and the member to the standard library. The pair's one
     matcher, so the two rules read the same shape.
     */
    static func filterCalls(
        reading member: String,
        filters: [String],
        in file: ParsedFile,
        symbols: FileSymbols,
    ) -> [FunctionCallExprSyntax] {
        let visitor = Visitor(member: member)
        visitor.walk(file.tree)
        return visitor.found.filter { candidate in
            guard let resolvedFilter = Self.reference(at: candidate.filter, in: file, symbols: symbols),
                filters.contains(where: { resolvedFilter.symbol.hasPrefix($0) })
            else {
                return false
            }
            return Self.reference(at: candidate.member, in: file, symbols: symbols)?.isStandardLibrary == true
        }.map(\.call)
    }

    /* The declaration the name at this token is a written reference to, if the compiler recorded exactly one. */
    static func reference(at token: TokenSyntax, in file: ParsedFile, symbols: FileSymbols) -> FileSymbols.Occurrence? {
        let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
        return symbols.reference(line: location.line, column: location.column)
    }

    /* One shape found: the `filter` call a finding starts at, its `filter` name token, and the member's name token. */
    struct Candidate {
        var call: FunctionCallExprSyntax
        var filter: TokenSyntax
        var member: TokenSyntax
    }

    /* Every read of the member from a `filter` call, with the name tokens the index places. */
    final class Visitor: SyntaxVisitor {
        let member: String
        private(set) var found: [Candidate] = []

        init(member: String) {
            self.member = member
            super.init(viewMode: .sourceAccurate)
        }

        override func visit(_ node: MemberAccessExprSyntax) -> SyntaxVisitorContinueKind {
            if node.declName.baseName.text == member, node.declName.argumentNames == nil, let base = node.base,
                let filterCall = Self.filterCall(base)
            {
                found.append(
                    Candidate(call: filterCall.call, filter: filterCall.filter, member: node.declName.baseName)
                )
            }
            return .visitChildren
        }

        /* The call and `filter` name token of `items.filter { }`, `filter(predicate)`, or either inside one pair of parentheses. */
        static func filterCall(_ expression: ExprSyntax) -> (call: FunctionCallExprSyntax, filter: TokenSyntax)? {
            var callExpression = expression
            if let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                callExpression = only.expression
            }
            guard let call = callExpression.as(FunctionCallExprSyntax.self) else { return nil }
            if let calledMember = call.calledExpression.as(MemberAccessExprSyntax.self),
                calledMember.declName.baseName.text == "filter"
            {
                return (call, calledMember.declName.baseName)
            }
            if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self),
                reference.baseName.text == "filter"
            {
                return (call, reference.baseName)
            }
            return nil
        }
    }
}
