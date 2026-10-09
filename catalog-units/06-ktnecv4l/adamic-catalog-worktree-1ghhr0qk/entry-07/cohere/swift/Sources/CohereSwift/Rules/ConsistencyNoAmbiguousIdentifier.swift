import Foundation
import SwiftSyntax

/*
 No single-letter names: `let n = items.count`, `for i in rows`, `{ e in report(e) }`. The Swift form of
 `nexus/consistency-no-ambiguous-identifier`, with its judgment and its messages carried over.

 A name is read everywhere it is used and declared only once, so the keystrokes a single letter saves are
 saved at the declaration and paid for at every read. The repair is the word the letter stands for:
 `count`, `rowIndex`, `error`.

 Which names: declared names only, exactly the set `cohere-swift/consistency-no-abbreviated-identifier` judges, found by
 its own visitor (`ConsistencyNoAbbreviatedIdentifier.Visitor`). A reference is spelled by whoever declared it, and our
 own single-letter declaration is reported once, where it is written. That visitor already skips an
 `override`'s name and its single-name labels (the superclass chose them), a label that is not also the
 parameter's name (API another type may dictate), and `if let value` with no initializer (it re-binds a name
 declared elsewhere, where the finding belongs).

 Only a single lowercase ASCII letter is in scope, which is the Go rule's line. An uppercase single letter is
 a type parameter by convention (`<T>`), and the Go rule leaves it alone, so this rule does too, even though
 the house spells type parameters as words (`<Element>`). A lowercase generic parameter, `<t>`, is judged
 like any other name. One place the shared visitor is wider than the Go rule: a protocol's property
 requirement is judged here, where the Go rule skips a TypeScript interface's property signature as a name
 somebody else may have chosen. A protocol is declared by this file, so its names are ours to rename.

 Exemptions, each the Go rule's:
 - `x`, `y` and `z`, the coordinate names, where the single letter is the conventional spelling and a longer
   one reads worse.
 - `a` and `b` in a sort comparator, because that is the shape everyone reads and naming them is noise. The
   Go rule exempts an `a` or `b` whose nearest enclosing function is the first argument of a `.sort(...)`
   call on some object. The Swift comparators are `sort(by:)`, `sorted(by:)`, `min(by:)` and `max(by:)`, so
   here the nearest enclosing closure must be the only argument of a call to one of those four names on some
   base (`items.sorted { a, b in ... }` or `items.sorted(by: { a, b in ... })`): a trailing closure with no
   other argument, or the single argument labeled `by`. A bare `sorted { ... }` with no base is not a
   comparator the Go rule would recognize, and a nested closure inside the comparator is judged on its own,
   both as the Go rule does.

 `_` is not judged, which is where Swift parts from the Go rule's `noUnderscore`. In Swift `_` is the discard
 pattern and the empty label, not a name anyone has to track.

 One exemption is ours, Kirk's ruling (2026-10-03): math notation in numeric kernels. A formula from a paper
 is checked against the paper, and there the letters are the names: Möller and Trumbore's `u`, `v` and `t`,
 the law of cosines' `a`, `b` and `c`, `lerp(a, b, t)`. Spelled out (`alongEdge1` beside an exempt `x`), the
 formula no longer reads as the one on the page. A numeric kernel is read from the signature, since this rule
 sees syntax and no types: the nearest enclosing `func`, or closure that writes every parameter's type, takes
 at least one parameter, every parameter is a number, the result is a number, an optional number or nothing,
 and a real number appears somewhere among them. A number is `Float`, `Double`, `CGFloat`, `Float16`, `Float80`,
 an `Int` of any size, a `SIMD2` to `SIMD64` of those, a simd type (`simd_float3`, `simd_quatd`,
 `simd_double3x3`, `matrix_float4x4`), an array or tuple of those, or any of them `inout`. A signature of
 integers alone is index and byte arithmetic, not a formula, so it is not a kernel. Everything declared inside
 the kernel counts, its parameters and locals and the parameters of a closure that leaves its types to
 inference; a nested `func`, a closure with its own written signature, a type, an initializer, a subscript or
 an accessor decides for itself. Inside a kernel only `kernelLetters` are allowed, a short set, each with the
 reason a paper uses it; `e` is never among them. Measured on the two AhraOS repositories before their renames
 (the cohere-zero branch), the kernel rule allows 274 of the 1,571 single letters this rule reported, all of them in the presence repository.

 `e` gets its own message, because it is the ambiguous one: it could be an error or an event, and which one
 decides the right name. The rule infers from context and says what it inferred, the Go rule's inference
 carried to Swift's shapes:
 - error: `catch let e`, or `catch let e as SomeError`, binds the thrown error itself. A payload bound in a
   catch pattern, `catch Failure.denied(let e)`, is not the error and falls through to the checks below.
 - event: an argument or trailing closure labeled `onSomething:` (the Go rule's `{ onClick: (e) => ... }` and
   `onClick={(e) => ...}`), a binding named like a handler whose initializer holds the `e`
   (`let handleClick = { e in ... }`), or a function named like a handler (`func handleKey(_ e: Key)`).
   "Named like a handler" is the Go rule's case-insensitive `handle|on[A-Z]|event`, which under case folding
   matches "on" followed by any letter.
 - otherwise "event", marked "context unclear", as the Go rule does.

 No fix, for the Go rule's reason: renaming a binding without following its references through scope leaves
 every other use pointing at a name that no longer exists. The suggested name travels in the message.
 */
public struct ConsistencyNoAmbiguousIdentifier: FileRule {
    public let name = "cohere-swift/consistency-no-ambiguous-identifier"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* The coordinate and math names, where the single letter is the conventional spelling. */
    static let alwaysAllowedSingleLetters: Set<String> = ["x", "y", "z"]

    /*
     The letters a numeric kernel may keep, each the one a paper writes: `a`, `b`, `c` for the operands, corners or
     sides (`lerp(a, b, t)`, the law of cosines); `d` for a difference or a distance; `p`, `q`, `r` for points,
     quaternions and a rotation or residual (`[r | t]`); `s` and `t` for the parameters along a path
     (`s = 1 - t`); `u`, `v`, `w` for barycentric weights and an SVD's factors. Indices (`i`, `j`, `k`), counts
     (`n`) and matrices (`m`, `h`) read as well or better spelled out, so they stay out.
     */
    static let kernelLetters: Set<String> = ["a", "b", "c", "d", "p", "q", "r", "s", "t", "u", "v", "w"]

    /* The Swift spellings of the Go rule's `.sort(...)`: every standard library method that takes a comparator closure. */
    static let comparatorMethodNames: Set<String> = ["sort", "sorted", "min", "max"]

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = ConsistencyNoAbbreviatedIdentifier.Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.declared.compactMap { token in
            let spelled = token.text.trimmingCharacters(in: CharacterSet(charactersIn: "`"))
            guard Self.isSingleLowercaseLetter(spelled), !Self.alwaysAllowedSingleLetters.contains(spelled) else {
                return nil
            }
            if spelled == "e" {
                let inference = Self.inferEventOrError(token)
                return file.finding(
                    at: token,
                    rule: name,
                    message: RuleMessages.ConsistencyNoAmbiguousIdentifier.noAmbiguousE(
                        context: inference.context,
                        suggestedName: inference.suggestedName,
                    ),
                )
            }
            /* A sort comparator is one place a and b read correctly. */
            if spelled == "a" || spelled == "b", Self.isInsideSortComparator(token) {
                return nil
            }
            /* A numeric kernel is the other: there the letter is the paper's notation, and a word for it hides the formula. */
            if Self.kernelLetters.contains(spelled), Self.isInsideNumericKernel(token) {
                return nil
            }
            return file.finding(
                at: token,
                rule: name,
                message: RuleMessages.ConsistencyNoAmbiguousIdentifier.noSingleLetter(name: spelled),
            )
        }
    }

    /* The Go rule's `^[a-z]$`: one ASCII lowercase letter, nothing else. */
    static func isSingleLowercaseLetter(_ name: String) -> Bool {
        guard name.utf8.count == 1, let byte = name.utf8.first else { return false }
        return byte >= UInt8(ascii: "a") && byte <= UInt8(ascii: "z")
    }

    /*
     The Go rule's case-insensitive `handle|on[A-Z]|event`. Case folding turns `[A-Z]` into any ASCII letter,
     so `on` followed by a letter anywhere matches, and so does `Handle` or `EVENT`. Written out by hand rather
     than as a regular expression so the folding is visible.
     */
    static func isNamedLikeHandler(_ name: String) -> Bool {
        let loweredName = name.lowercased()
        if loweredName.contains("handle") || loweredName.contains("event") {
            return true
        }
        let lowered = Array(loweredName.utf8)
        guard lowered.count >= 3 else { return false }
        for start in 0..<(lowered.count - 2)
        where lowered[start] == UInt8(ascii: "o") && lowered[start + 1] == UInt8(ascii: "n") {
            let following = lowered[start + 2]
            if following >= UInt8(ascii: "a") && following <= UInt8(ascii: "z") {
                return true
            }
        }
        return false
    }

    /* The Go rule's `^on[A-Z]`, case-sensitive: `onClick`, never `online`. */
    static func isEventHandlerKey(_ label: String) -> Bool {
        let bytes = Array(label.utf8)
        guard bytes.count >= 3, bytes[0] == UInt8(ascii: "o"), bytes[1] == UInt8(ascii: "n") else { return false }
        return bytes[2] >= UInt8(ascii: "A") && bytes[2] <= UInt8(ascii: "Z")
    }

    /*
     Whether the nearest closure around the token is the comparator handed to `sort`, `sorted`, `min` or `max`
     on some base: the call's trailing closure with no other argument, or its single argument, labeled `by`.
     The nearest closure decides, as the nearest function does in the Go rule, so an `a` in a closure nested
     inside a comparator is judged.
     */
    static func isInsideSortComparator(_ token: TokenSyntax) -> Bool {
        guard let closure = token.parent?.ancestorOrSelf(mapping: { $0.as(ClosureExprSyntax.self) }) else {
            return false
        }
        let call: FunctionCallExprSyntax
        if let argument = closure.parent?.as(LabeledExprSyntax.self) {
            guard argument.label?.text == "by", let list = argument.parent?.as(LabeledExprListSyntax.self),
                list.count == 1,
                let owner = list.parent?.as(FunctionCallExprSyntax.self), owner.trailingClosure == nil
            else { return false }
            call = owner
        }
        else if let owner = closure.parent?.as(FunctionCallExprSyntax.self), owner.trailingClosure?.id == closure.id {
            guard owner.arguments.isEmpty, owner.additionalTrailingClosures.isEmpty else { return false }
            call = owner
        }
        else {
            return false
        }
        guard let callee = call.calledExpression.as(MemberAccessExprSyntax.self), callee.base != nil else {
            return false
        }
        return comparatorMethodNames.contains(callee.declName.baseName.text)
    }

    /*
     Whether the token is declared in a numeric kernel: its nearest enclosing function, or closure whose signature
     writes every type, takes only numbers and returns a number or nothing. A closure whose types are left to
     inference is part of the function around it. A type declared in between ends the walk, and so does an
     initializer, a subscript or an accessor, none of which is a kernel.
     */
    static func isInsideNumericKernel(_ token: TokenSyntax) -> Bool {
        var current = token.parent
        while let node = current {
            if let function = node.as(FunctionDeclSyntax.self) {
                return isNumericKernel(
                    parameters: function.signature.parameterClause.parameters.map(\.type),
                    result: function.signature.returnClause?.type,
                )
            }
            if let closure = node.as(ClosureExprSyntax.self), let signature = closure.signature,
                let types = writtenParameterTypes(signature)
            {
                return isNumericKernel(parameters: types, result: signature.returnClause?.type)
            }
            if node.is(InitializerDeclSyntax.self) || node.is(SubscriptDeclSyntax.self)
                || node.is(AccessorDeclSyntax.self) || node.is(AccessorBlockSyntax.self)
                || node.asProtocol((any DeclGroupSyntax).self) != nil
            {
                return false
            }
            current = node.parent
        }
        return false
    }

    /*
     At least one parameter, every one a number, the result a number, an optional number (a hit or none), or
     nothing, and a real number somewhere among them. A signature of integers alone is index and byte arithmetic
     (a terminal's column count, a hex nibble), not a formula, and its letters are judged as anywhere else.
     */
    static func isNumericKernel(parameters: [TypeSyntax], result: TypeSyntax?) -> Bool {
        guard !parameters.isEmpty, parameters.allSatisfy(isNumeric) else { return false }
        let returned = result.map { $0.as(OptionalTypeSyntax.self)?.wrappedType ?? $0 }
        if let returned, !isNumeric(returned), returned.as(IdentifierTypeSyntax.self)?.name.text != "Void",
            returned.as(TupleTypeSyntax.self)?.elements.isEmpty != true
        {
            return false
        }
        return (parameters + [returned].compactMap { $0 }).contains(where: isReal)
    }

    /* A closure's parameter types when it writes every one, `{ (a: Float, b: Float) -> Float in ... }`; nil for `{ a, b in ... }`. */
    static func writtenParameterTypes(_ signature: ClosureSignatureSyntax) -> [TypeSyntax]? {
        guard case .parameterClause(let clause) = signature.parameterClause else { return nil }
        let types = clause.parameters.compactMap(\.type)
        return types.count == clause.parameters.count ? types : nil
    }

    /* The real scalar types, and with the integers every scalar a kernel computes with. */
    static let realScalars: Set<String> = ["Float", "Double", "Float16", "Float80", "CGFloat"]
    static let numericScalars: Set<String> = realScalars.union([
        "Int", "Int8", "Int16", "Int32", "Int64", "UInt", "UInt8", "UInt16", "UInt32", "UInt64",
    ])

    /* The vector types that take a scalar as their generic argument. */
    static let numericVectors: Set<String> = ["SIMD2", "SIMD3", "SIMD4", "SIMD8", "SIMD16", "SIMD32", "SIMD64"]

    /*
     A number as a kernel holds one: a scalar, a SIMD vector of them, one of simd's own types (`simd_float3`,
     `simd_quatf`, `simd_float4x4`, `matrix_float4x4`), an array or a tuple of those, or any of them `inout`.
     */
    static func isNumeric(_ type: TypeSyntax) -> Bool {
        if let identifier = type.as(IdentifierTypeSyntax.self) {
            let name = identifier.name.text
            guard let generic = identifier.genericArgumentClause else {
                return numericScalars.contains(name) || name.hasPrefix("simd_") || name.hasPrefix("matrix_")
            }
            return numericVectors.contains(name)
                && generic.arguments.allSatisfy { argument in
                    guard case .type(let scalar) = argument.argument else { return false }
                    return isNumeric(scalar)
                }
        }
        if let array = type.as(ArrayTypeSyntax.self) {
            return isNumeric(array.element)
        }
        if let tuple = type.as(TupleTypeSyntax.self) {
            return !tuple.elements.isEmpty && tuple.elements.allSatisfy { isNumeric($0.type) }
        }
        if let attributed = type.as(AttributedTypeSyntax.self) {
            return isNumeric(attributed.baseType)
        }
        return false
    }

    /* Whether a numeric type holds a real number anywhere: a real scalar, a vector of one, a simd float, double, half or quaternion type, or an array or tuple holding one. */
    static func isReal(_ type: TypeSyntax) -> Bool {
        if let identifier = type.as(IdentifierTypeSyntax.self) {
            let name = identifier.name.text
            guard let generic = identifier.genericArgumentClause else {
                let isSimd = name.hasPrefix("simd_") || name.hasPrefix("matrix_")
                return realScalars.contains(name)
                    || (isSimd && ["float", "double", "half", "quat"].contains { name.contains($0) })
            }
            return generic.arguments.contains { argument in
                guard case .type(let scalar) = argument.argument else { return false }
                return isReal(scalar)
            }
        }
        if let array = type.as(ArrayTypeSyntax.self) {
            return isReal(array.element)
        }
        if let tuple = type.as(TupleTypeSyntax.self) {
            return tuple.elements.contains { isReal($0.type) }
        }
        if let attributed = type.as(AttributedTypeSyntax.self) {
            return isReal(attributed.baseType)
        }
        return false
    }

    /* The suggestion and the reason for it, said in the message so a reader can disagree with the reasoning and not only the verdict. */
    struct Inference {
        let suggestedName: String
        let context: RuleMessages.ConsistencyNoAmbiguousIdentifier.Context
    }

    static func inferEventOrError(_ token: TokenSyntax) -> Inference {
        if isCaughtError(token) {
            return Inference(suggestedName: "error", context: .error)
        }
        let event = Inference(suggestedName: "event", context: .event)
        var current = token.parent
        while let node = current {
            if let parent = node.parent {
                /* A handler passed as `Button(onTap: { e in ... })`. */
                if let argument = parent.as(LabeledExprSyntax.self), argument.expression.id == node.id,
                    let label = argument.label, isEventHandlerKey(label.text)
                {
                    return event
                }
                /* A handler passed as a labeled trailing closure, `} onChange: { e in ... }`. */
                if let element = parent.as(MultipleTrailingClosureElementSyntax.self), element.closure.id == node.id,
                    isEventHandlerKey(element.label.text)
                {
                    return event
                }
                /* Assigned to something named like a handler, `let handleClick = { e in ... }`. */
                if let binding = parent.as(PatternBindingSyntax.self), binding.initializer?.id == node.id,
                    let identifier = binding.pattern.as(IdentifierPatternSyntax.self),
                    isNamedLikeHandler(identifier.identifier.text)
                {
                    return event
                }
            }
            /* A function named like a handler, `func handleKey(_ e: Key)`. */
            if let function = node.as(FunctionDeclSyntax.self), isNamedLikeHandler(function.name.text) {
                return event
            }
            current = node.parent
        }
        return Inference(suggestedName: "event", context: .unclear)
    }

    /*
     Whether the token is the thrown error a `catch` binds: `catch let e`, `catch var e`, or `catch let e as
     SomeError`. The walk from the name to the catch item may pass only through the binding keyword and the
     cast, so a payload bound inside a pattern, `catch Failure.denied(let e)`, is not the error.
     */
    static func isCaughtError(_ token: TokenSyntax) -> Bool {
        guard let identifier = token.parent?.as(IdentifierPatternSyntax.self) else { return false }
        var current = Syntax(identifier)
        while let parent = current.parent {
            if parent.is(CatchItemSyntax.self) {
                return true
            }
            if parent.is(ValueBindingPatternSyntax.self) || parent.is(PatternExprSyntax.self)
                || parent.is(ExpressionPatternSyntax.self)
                || parent.is(AsExprSyntax.self)
            {
                current = parent
                continue
            }
            /* The unfolded `e as SomeError` is the sequence `e`, `as SomeError`; the name must lead it. */
            if let elements = parent.as(ExprListSyntax.self), elements.first?.id == current.id,
                elements.dropFirst().first?.is(UnresolvedAsExprSyntax.self) == true, elements.count == 3,
                let sequence = elements.parent?.as(SequenceExprSyntax.self)
            {
                current = Syntax(sequence)
                continue
            }
            return false
        }
        return false
    }
}
