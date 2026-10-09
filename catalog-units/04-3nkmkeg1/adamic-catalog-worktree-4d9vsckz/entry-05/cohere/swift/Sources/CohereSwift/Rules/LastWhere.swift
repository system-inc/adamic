import SwiftSyntax

/*
 No filtering a collection only to take its last element: `items.filter { $0.isActive }.last` is
 `items.last { $0.isActive }`. `last(where:)` says the question being asked, searches from the end and stops
 at the first match it meets, where `filter` runs the predicate on every element and builds a whole new array
 (or string) only to keep one element of it and throw the rest away.

 The shape is SwiftLint's `last_where`, read by the same matcher as `first-where`
 (`FirstWhere.filterCalls(reading:filters:in:symbols:)`): `last` read from a member `filter` call, with a
 trailing closure (`items.filter { }.last`), an argument (`filter(predicate)`, `filter({ })`), or wrapped in
 one pair of parentheses (`(items.filter { }).last`), anywhere in a chain (`items.map { }.filter { }.last`),
 followed by anything (`.last?.name`, `.last!`), and on a line break (`filter { }\n.last`). The finding starts
 where SwiftLint's does, at the start of the `filter` call. `last(where:)` called on a `filter`'s result is the
 shape too, as SwiftLint reads it: the array is still built, and the repair is one `last(where:)` with both
 predicates.

 SwiftLint skips a `filter` whose argument is a string literal, Realm's query `filter`, by its spelling. This
 rule asks the types instead, as `first-where` does: Realm's `filter`, and any `filter` of ours, resolves to
 its own module and is never flagged, even one an extension of ours adds to `Sequence`; nor is a `last` an
 extension of ours adds to `BidirectionalCollection`.

 Where this rule is narrower than `first-where`, because the repair must exist: `last(where:)` is declared on
 `BidirectionalCollection`, not on every sequence, and a `filter` call always returns something with a `last`
 even when its receiver has no `last(where:)`. `dictionary.values.filter { }.last` compiles, and
 `dictionary.values.last(where:)` does not. So this rule flags only a `filter` the compiler resolved to a
 declaration whose receiver is known to be bidirectional: `Array`'s (an array, an array slice or a contiguous
 array), `Substring`'s, and `RangeReplaceableCollection`'s, which returns the receiver's own type, so a `last`
 that resolved to the standard library on the result proves the receiver has `last(where:)` too. `Sequence`'s
 `filter` is left out, since its receiver may be any sequence; `Dictionary`'s and `Set`'s return a collection
 with no `last`, so they never reach this shape.

 A lazy `filter` (`items.lazy.filter { }.last`) is not flagged: it builds nothing, and its `last` already
 searches from the end. Nor is an asynchronous sequence's `filter`. Both are left out by naming the eager
 declarations rather than excluding the lazy ones.

 Misses, accepted: `Sequence`'s `filter` on a receiver that is bidirectional after all (a range's
 `(0..<count).filter { }.last`, an array's `items.indices.filter { }.last`) is not flagged, because the
 declaration cannot tell a range from a dictionary's values; a `filter` declared on a type of ours resolves to
 ours and is not flagged; a `filter` inside parentheses with `try` or `await` is not read. Where SwiftLint
 stops short and this rule does not: a bare `filter` inside a collection's own extension (`filter { }.last`) is
 found as well as a member one.
 */
public struct LastWhere: TypedFileRule {
    public let name = "cohere-swift/last-where"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "last_where"

    public init() {}

    /*
     The standard library `filter` declarations that build a new collection from a receiver known to be
     bidirectional, by the start of their symbol names: `Array`'s (and `ArraySlice`'s and `ContiguousArray`'s),
     `RangeReplaceableCollection`'s (`String`'s, its views'), and `Substring`'s.
     */
    static let bidirectionalFilters = [
        "s:s14_ArrayProtocolPsE6filter",
        "s:SmsE6filter",
        "s:Ss6filter",
    ]

    /* A `filter` and a `last` anywhere in the file: nothing else can hold this shape. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("filter") && file.source.contains("last")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        FirstWhere.filterCalls(reading: "last", filters: Self.bidirectionalFilters, in: file, symbols: symbols).map {
            call in
            file.finding(
                at: call,
                rule: name,
                messageId: "lastWhere",
                message:
                    "Taking the last element of a filtered collection builds a whole new collection of every match only to keep one. Use last(where:) with the filter's predicate (joined by && with last's own, when it has one): it says what is meant, and it searches from the end and stops at the first match it meets.",
            )
        }
    }
}
