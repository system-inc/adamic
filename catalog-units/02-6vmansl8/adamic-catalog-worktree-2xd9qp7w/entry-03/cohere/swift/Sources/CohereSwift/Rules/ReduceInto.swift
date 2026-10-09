import SwiftParser
import SwiftSyntax

/*
 No folding into a copy-on-write value with `reduce(_:_:)`: `values.reduce([]) { $0 + [$1] }` is
 `values.reduce(into: []) { $0.append($1) }`. `reduce(_:_:)` hands the closure its accumulator as a constant and
 takes back a new one, so a closure that grows an array, dictionary, set or string copies the whole of it on every
 element, quadratic in the length. `reduce(into:_:)` hands the closure the accumulator as `inout`, and one value
 grows in place.

 The shape is SwiftLint's `reduce_into`: a call of `reduce`, as a member (`values.reduce`) or bare (`reduce`,
 inside a sequence's own extension), with two arguments or one and a trailing closure, whose first argument is
 unlabelled (so not `reduce(into:)`) and is one of: a string literal; an array or dictionary literal; a call of
 `Array`, `Dictionary` or `Set`, plain or specialised (`Array<Int>()`, `Set<Int>(minimumCapacity: 4)`), or of
 their `init` (`Dictionary<String, Int>.init()`); or a call of the array or dictionary sugar
 (`[Int](repeating: 0, count: 10)`, `[String: Int]()`). The finding starts where SwiftLint's does, at the
 `reduce` name.

 SwiftLint judges the spelling, so it also flags a `reduce` of ours, and a literal that became a type which is
 not copied at all. This rule flags only a `reduce` the compiler resolved to the standard library's
 `Sequence.reduce(_:_:)`, read from the index the build wrote, and then asks the index what the initial value
 is. An array or dictionary literal is flagged only where the compiler recorded the literal as building an
 `Array`, `ContiguousArray`, `Set` or `Dictionary`: an `OptionSet`'s `[]`, the common `flags.reduce([]) {
 $0.union($1) }`, is a bit field that nothing copies, and records no initializer at all, and a literal type of
 ours records its own. An interpolated string literal is flagged when it records `String`'s interpolation. A plain string
 literal that became a `String` records nothing, the same nothing a `Character` or a `StaticString` records,
 so it is read by what it holds: one character could be a `Character` (or a `Unicode.Scalar`) and is not
 flagged, anything else (`""`, `"abc"`) is a `String` or a `StaticString`, and a `StaticString` cannot be
 combined into anything, so it is flagged as a `String`. A literal that records another type's initializer (a
 `Substring`, a type of ours) is not flagged. A named type is confirmed by what its name resolved to,
 so a `Set` of our own declaring is not the standard library's. The sugar `[Int]` and `[String: Int]` can only
 be an array and a dictionary, so its calls are taken by their syntax.

 Where we differ on purpose, in the direction of finding more of the same thing: `String(...)` and
 `ContiguousArray(...)` are copy-on-write too and are found as initial values, and `[Int].init()` is found as
 `[Int]()` is. A lazy sequence's `reduce` is the same `Sequence.reduce(_:_:)` and copies its accumulator the same
 way, so it is flagged. No sibling typed rule shares this shape.

 Misses, accepted: a `reduce` declared on a type of ours resolves to ours and is not flagged; a one-character
 string literal (`reduce(",")`) cannot be told from a `Character` and is not flagged; an initial value that is a
 variable, a property, a function's result, `.init()` with its type inferred, `Swift.Array<Int>()`, or a
 typealias of an array resolves to nothing this rule can read as a copy-on-write type, and is not flagged.
 */
public struct ReduceInto: TypedFileRule {
    public let name = "cohere-swift/reduce-into"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "reduce_into"

    public init() {}

    /* `Sequence.reduce`, as its symbol name begins, matched whole so a `reduce` another module adds to `Sequence` is not taken for it. */
    static let sequenceReduce = "s:STsE6reduce"

    /* The copy-on-write types an initial value may name, as written: SwiftLint's three, and `String` and `ContiguousArray`. */
    static let copyOnWriteTypeNames: Set<String> = ["Array", "ContiguousArray", "Dictionary", "Set", "String"]

    /* The same types as the compiler names them: `Array`, `ContiguousArray`, `Dictionary`, `Set` and `String`. */
    static let copyOnWriteTypes: Set<String> = ["s:Sa", "s:s15ContiguousArrayV", "s:SD", "s:Sh", "s:SS"]

    /* A `reduce` anywhere in the file: the shape starts with one. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("reduce")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.compactMap { reduce, initial in
            let location = file.locations.location(for: reduce.positionAfterSkippingLeadingTrivia)
            guard let resolved = symbols.reference(line: location.line, column: location.column),
                resolved.symbol.hasPrefix(Self.sequenceReduce),
                Self.isCopyOnWrite(initial, in: file, symbols: symbols)
            else {
                return nil
            }
            return file.finding(
                at: reduce,
                rule: name,
                messageId: "reduceInto",
                message:
                    "reduce(_:_:) with an array, dictionary, set or string as its accumulator copies the whole accumulator on every element. Use reduce(into:_:) and change the accumulator in place (append, insert, a subscript assignment, +=): it grows one value instead of building a new one per element.",
            )
        }
    }

    /* Whether the initial value is a copy-on-write value of the standard library's, by its syntax and what the compiler recorded for it. */
    static func isCopyOnWrite(_ initial: ExprSyntax, in file: ParsedFile, symbols: FileSymbols) -> Bool {
        if initial.is(ArrayExprSyntax.self) || initial.is(DictionaryExprSyntax.self) {
            return literalInitializers(of: initial, in: file, symbols: symbols).contains { isCopyOnWriteMember($0) }
        }
        if let literal = initial.as(StringLiteralExprSyntax.self) {
            let initializers = literalInitializers(of: initial, in: file, symbols: symbols)
            if !initializers.isEmpty {
                return initializers.contains { $0.isStandardLibrary && $0.symbol.hasPrefix("s:SS") }
            }
            /* Nothing recorded: a plain `String`, or a builtin literal type, which one character could be. */
            guard let value = literal.representedLiteralValue else { return false }
            return value.count != 1
        }
        guard let call = initial.as(FunctionCallExprSyntax.self) else { return false }
        var callee = call.calledExpression
        if let member = callee.as(MemberAccessExprSyntax.self), member.declName.baseName.text == "init",
            let base = member.base
        {
            callee = base
        }
        if callee.is(ArrayExprSyntax.self) || callee.is(DictionaryExprSyntax.self) {
            return true
        }
        if let specialization = callee.as(GenericSpecializationExprSyntax.self) {
            callee = specialization.expression
        }
        guard let typeName = callee.as(DeclReferenceExprSyntax.self),
            copyOnWriteTypeNames.contains(typeName.baseName.text)
        else { return false }
        let location = file.locations.location(for: typeName.baseName.positionAfterSkippingLeadingTrivia)
        return symbols.occurrences(line: location.line, column: location.column).contains {
            $0.isReference && !$0.isImplicit && copyOnWriteTypes.contains($0.symbol)
        }
    }

    /*
     The initializers the compiler recorded for a literal, at its first character (and, for a raw string, at its
     opening quote): implied by the source, so implicit, and the only record of the type a literal became.
     */
    static func literalInitializers(
        of literal: ExprSyntax,
        in file: ParsedFile,
        symbols: FileSymbols,
    ) -> [FileSymbols.Occurrence] {
        var positions = [literal.positionAfterSkippingLeadingTrivia]
        if let string = literal.as(StringLiteralExprSyntax.self) {
            positions.append(string.openingQuote.positionAfterSkippingLeadingTrivia)
        }
        return Set(positions).flatMap { position in
            let location = file.locations.location(for: position)
            return symbols.occurrences(line: location.line, column: location.column).filter(\.isReference)
        }
    }

    /* The standard library's own member of `Array`, `ContiguousArray`, `Dictionary`, `Set` or `String`, as its symbol name begins. */
    static func isCopyOnWriteMember(_ occurrence: FileSymbols.Occurrence) -> Bool {
        occurrence.isStandardLibrary && copyOnWriteTypes.contains { occurrence.symbol.hasPrefix($0) }
    }

    /* Every `reduce` call of SwiftLint's shape: the `reduce` token the index places, and the initial value. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [(TokenSyntax, ExprSyntax)] = []

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            guard let reduce = Self.reduceToken(node.calledExpression),
                node.arguments.count == 2 || (node.arguments.count == 1 && node.trailingClosure != nil),
                let initial = node.arguments.first, initial.label == nil
            else {
                return .visitChildren
            }
            found.append((reduce, initial.expression))
            return .visitChildren
        }

        /* The `reduce` name token of `values.reduce` or a bare `reduce`. */
        static func reduceToken(_ expression: ExprSyntax) -> TokenSyntax? {
            if let member = expression.as(MemberAccessExprSyntax.self), member.declName.baseName.text == "reduce" {
                return member.declName.baseName
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self), reference.baseName.text == "reduce" {
                return reference.baseName
            }
            return nil
        }
    }
}
