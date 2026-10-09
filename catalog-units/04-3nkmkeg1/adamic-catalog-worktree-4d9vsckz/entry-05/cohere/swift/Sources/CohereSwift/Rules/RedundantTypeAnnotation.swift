import Foundation
import SwiftSyntax

/*
 No naming a declaration's type twice: `let url: URL = URL(fileURLWithPath: path)` says `URL` in the annotation
 and again in the initializer, and the annotation adds nothing the initializer did not already say. Written
 once, the line reads as what it does, and the two spellings cannot drift apart when one of them changes.

 The shapes are SwiftLint's `redundant_type_annotation` in its default configuration: a `let` or `var` binding
 (stored or local, `lazy` or not, one binding of several in one declaration) or an `if let`, `if var`,
 `guard let` or `guard var` binding, whose annotation's type is the type the initializer names. SwiftLint reads
 the initializer as a call of the type (`URL()`, `URL(string: text)`), a call of its `init` (`URL.init()`), a
 generic spelling (`Set<Int>([])`, with an annotation of `Set<Int>` or a bare `Set`), the sugar
 (`[Int]()`), a qualified or nested name (`A.B()`), any of those force unwrapped (`URL(string: text)!`), and
 any member read from the type or chain hanging off it (`CharacterSet.alphanumerics`, `Int.random(in:)`,
 `Direction.up`, `A.b.c.d`, `A.f().b`). Its defaults skip a declaration marked `@IBInspectable`, whose
 annotation Interface Builder reads, check properties as well as locals, and leave a literal with its default
 type alone (`let count: Int = 0`). An annotation SwiftLint spells differently is never the shape: an optional
 one (`let url: URL? = URL(string: text)`, which SwiftLint never matches), an existential (`any Shape`), and an
 implicit member (`let url: URL = .init(fileURLWithPath: path)`), whose annotation is what names the type.
 The finding starts where SwiftLint's does, at the annotation's colon.

 SwiftLint compares spellings. This rule compares declarations, read from the index the build wrote: the
 annotation's type name resolves to a type, the initializer's type name resolves to a type, and they must be
 the same one. So `let value: Base = Derived()` and `let shape: any Shape = Circle()` are never flagged, and nor
 is a type alias over a different spelling (`let circle: Circle = Round()`), which this rule cannot see through;
 the same alias written on both sides is the same declaration and is flagged. Where the annotation spells
 generic arguments, the initializer must spell the same ones, as SwiftLint asks, because
 `let ids: Set<Int> = Set([])` needs its annotation to choose the element type.

 Removing the annotation must leave the type as it was, which an initializer that can return nil would break.
 An `init!` called without `!` binds to a non-optional annotation and to an optional without one, and nothing
 in the call shows which it is. So outside an optional binding and a force unwrap, both of which bind the
 initializer's type whatever it is, this rule flags only an initializer it can show never returns nil: a Swift
 initializer whose symbol has no optional in it (a failable one's names its optional result), or a literal the
 compiler coerces to the type (`Int(5)` is `5 as Int` and records no initializer at all). `try` and `await` in
 front of the initializer change no type and are looked through, as is `try?` in an optional binding, which
 unwraps the optional it adds; SwiftLint reads none of them, and this is a difference in the direction of finding more of the same thing.

 A member read from the type, or a chain hanging off the type or off an initializer of it, is SwiftLint's too
 (`CharacterSet.alphanumerics`, `Int.random(in:)`, `Direction.up`, `A.b.c.d`, `A.f().b`,
 `URL(fileURLWithPath: path).deletingLastPathComponent()`): SwiftLint flags it when any link, read from the
 left through member reads, calls and generic arguments, spells the annotation. This rule asks the same of the
 chain, a link that names the annotation's type, and then asks what SwiftLint does not, the chain's type, which
 is its last member's. An enum case of the annotation's own enum (`Direction.up`, `Direction.moved(by: 1)`) is
 always that enum. Any other last member is read from its declaration, demangled into the compiler's own words
 by the toolchain's demangler (`Foundation.URL.deletingLastPathComponent() -> Foundation.URL`), with the
 signature reading `correctness-no-discarded-try-optional` uses, and its result must be the annotation's own declaration as
 the demangler prints it, or `Self` on it. `Self` is printed two ways: `Self` for a class's dynamic `Self`, and
 `A` for a protocol's (`static (extension in Swift):Swift.FixedWidthInteger.random(in: Swift.ClosedRange<A>) -> A`),
 where `A` is `Self` only because the member is a protocol's; on a type, `A` is the member's own generic
 parameter (`static func decode<T>() -> T`), which the annotation chooses. So an `A` counts only when the
 demangler confirms the member's context is a protocol, by printing a prefix of the member's symbol as both
 that type and a protocol descriptor for it. `Self` is the type of the link the member is read from, so that link must be the
 annotated type, named (`UInt64.random(in:)`) or shown to have it (`Circle().again()`). A static member of a
 class that holds a subclass is read by its declared type, which is the type the binding takes without the
 annotation (`static let shared: Base = Derived()` makes `Base.shared` a `Base`). An optional result counts
 only where a force unwrap or an optional binding takes it unwrapped, so an implicitly unwrapped member
 (`static var current: Session!`), whose annotation is what unwraps it, is never flagged. An annotation that is
 a protocol is never read this way, since the demangler prints an existential as its protocol's bare name.

 What SwiftLint stops at, this rule stops at: an optional chain, a force unwrap or a subscript inside the chain
 (`A.b?.c`, `A.b!.c`, `A.b[0]`), parentheses, and a chain whose link spells another name, even one the
 types could judge (`items.count`). Misses, accepted: a result spelled with generic arguments (`Set<A>`,
 `Control.Box<A>`), an alias's member or a member returning an alias, which print a different declaration, `Self`
 from a constrained protocol extension, whose context prints its constraints, a property of function type that
 is called (`Circle.factory()`), a member read off the sugar (`[Int]().reversed()`), and an Objective-C member,
 whose symbol does not demangle. Misses too: an Objective-C initializer (`NSView()`, and the initializers of
 our own `@objc` classes), whose symbol does not say whether it can return nil, flagged only when force
 unwrapped or bound; a Swift initializer that takes or names an optional anywhere (`DispatchQueue(label:)`,
 whose `target` is one), which its symbol cannot tell from a failable one; and a generic spelled differently on
 the two sides (`Set<Int>` against `Set<Swift.Int>`). A toolchain without its demangler leaves the members
 other than enum cases unjudged.

 The one thing this rule trusts and does not check: that the annotation does not choose between two
 declarations of the same name that the same arguments fit, an initializer failable and not, or a member
 returning the annotated type and another returning something else. Only the one that fits the annotation
 could bind to it, and without it the other might win, or the call be ambiguous. No type we have read declares
 such a pair.

 No sibling rule reads this shape.
 */
public struct RedundantTypeAnnotation: TypedFileRule {
    public let name = "cohere-swift/redundant-type-annotation"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "redundant_type_annotation"

    public init() {}

    /* SwiftLint's default `ignore_attributes`: Interface Builder reads an inspectable property's annotation. */
    static let ignoredAttributes: Set<String> = ["IBInspectable"]

    /* An annotation's colon and an initializer's equals sign anywhere in the file: nothing else can hold this shape. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains(":") && file.source.contains("=")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        /* Loaded only when a binding needs it. A toolchain without the demangler leaves the members unjudged, a miss and never a finding. */
        let demangler = visitor.found.isEmpty ? nil : try? SwiftDemangler.shared()
        return visitor.found.compactMap { candidate in
            guard Self.isRedundant(candidate, in: file, symbols: symbols, demangler: demangler) else { return nil }
            return file.finding(
                at: candidate.annotation,
                rule: name,
                messageId: "redundantTypeAnnotation",
                message:
                    "This annotation names \(candidate.annotation.type.trimmedDescription), which the initializer already names. Remove the annotation and let the type be inferred: the declaration then says its type once.",
            )
        }
    }

    /* One binding with both an annotation and an initializer, and whether it is an optional binding, which unwraps whatever the initializer gives. */
    struct Candidate {
        var annotation: TypeAnnotationSyntax
        var initializer: ExprSyntax
        var isOptionalBinding: Bool
    }

    static func isRedundant(
        _ candidate: Candidate,
        in file: ParsedFile,
        symbols: FileSymbols,
        demangler: SwiftDemangler?,
    ) -> Bool {
        var expression = candidate.initializer
        var isForced = false
        while true {
            if let tryExpression = expression.as(TryExprSyntax.self),
                tryExpression.questionOrExclamationMark?.tokenKind != .postfixQuestionMark
                    || candidate.isOptionalBinding
            {
                expression = tryExpression.expression
            }
            else if let awaitExpression = expression.as(AwaitExprSyntax.self) {
                expression = awaitExpression.expression
            }
            else if let forceUnwrap = expression.as(ForceUnwrapExprSyntax.self) {
                expression = forceUnwrap.expression
                isForced = true
            }
            else {
                break
            }
        }
        /* Either way the binding takes the initializer's type unwrapped, so whether it could return nil does not matter. */
        let mayReturnNil = isForced || candidate.isOptionalBinding
        let annotationType = candidate.annotation.type

        /* The sugar, `[Int]` and `[String: Int]`: no name to resolve, so the two spellings must match token for token. */
        if annotationType.is(ArrayTypeSyntax.self) || annotationType.is(DictionaryTypeSyntax.self) {
            guard let construction = Self.construction(expression),
                construction.type.is(ArrayExprSyntax.self) || construction.type.is(DictionaryExprSyntax.self)
            else {
                return false
            }
            return Self.spelling(construction.type) == Self.spelling(annotationType)
                && Self.buildsWithoutNil(construction, in: file, symbols: symbols, mayReturnNil: mayReturnNil)
        }

        guard let annotationName = Self.typeName(annotationType),
            let annotationResolved = Self.reference(at: annotationName, in: file, symbols: symbols)
        else {
            return false
        }
        /* Generic arguments the annotation spells choose the type, so the initializer must spell the same ones. */
        let spellsGenericArguments = annotationType.tokens(viewMode: .sourceAccurate).contains {
            $0.tokenKind == .leftAngle
        }
        func namesTheAnnotatedType(_ typeExpression: ExprSyntax) -> Bool {
            guard let typeName = Self.typeName(typeExpression),
                Self.names(typeName, annotationResolved.symbol, in: file, symbols: symbols)
            else { return false }
            return !spellsGenericArguments || Self.spelling(typeExpression) == Self.spelling(annotationType)
        }

        if let construction = Self.construction(expression), namesTheAnnotatedType(construction.type) {
            return Self.buildsWithoutNil(construction, in: file, symbols: symbols, mayReturnNil: mayReturnNil)
        }
        if let enumCase = Self.enumCase(expression), namesTheAnnotatedType(enumCase.type),
            let resolvedCase = Self.reference(at: enumCase.member, in: file, symbols: symbols),
            Self.isCase(resolvedCase.symbol, named: enumCase.member, of: annotationResolved.symbol)
        {
            return true
        }

        /*
         A member or a chain, judged by its last member's result. The annotation must be a Swift type that is not a
         protocol: the demangler prints a protocol descriptor only for a protocol's symbol.
         */
        guard let demangler, Self.isRooted(expression, namesTheAnnotatedType),
            let annotatedType = demangler.declaration(ofSymbol: annotationResolved.symbol),
            demangler.demangle("$s\(annotationResolved.symbol.dropFirst(2))Mp") == nil
        else {
            return false
        }
        let chain = Chain(
            file: file,
            symbols: symbols,
            demangler: demangler,
            annotatedType: annotatedType,
            namesTheAnnotatedType: namesTheAnnotatedType,
        )
        return chain.hasAnnotatedType(expression, isUnwrapped: mayReturnNil)
    }

    /*
     Whether a link of the chain, read from the left as SwiftLint reads it (through member reads, calls and
     generic arguments, and nothing else), names the annotation's type: `URL` in `URL(fileURLWithPath: path).standardized`,
     `UInt64` in `UInt64.random(in: range)`.
     */
    static func isRooted(_ expression: ExprSyntax, _ namesTheAnnotatedType: (ExprSyntax) -> Bool) -> Bool {
        var link: ExprSyntax? = expression
        while let current = link {
            if namesTheAnnotatedType(current) {
                return true
            }
            if let call = current.as(FunctionCallExprSyntax.self) {
                link = call.calledExpression
            }
            else if let member = current.as(MemberAccessExprSyntax.self) {
                link = member.base
            }
            else if let specialization = current.as(GenericSpecializationExprSyntax.self) {
                link = specialization.expression
            }
            else {
                link = nil
            }
        }
        return false
    }

    /* The type of a chain's links, read from each member's demangled declaration and compared with the annotated type as the demangler prints it (`Foundation.URL`). */
    struct Chain {
        let file: ParsedFile
        let symbols: FileSymbols
        let demangler: SwiftDemangler
        let annotatedType: String
        let namesTheAnnotatedType: (ExprSyntax) -> Bool

        /*
         Whether the expression's value has the annotated type: a construction of it, or a member whose result is
         it, or is `Self` read from a link that is the type. Where the binding unwraps the value, an optional of it
         counts too.
         */
        func hasAnnotatedType(_ expression: ExprSyntax, isUnwrapped: Bool) -> Bool {
            if let construction = RedundantTypeAnnotation.construction(expression),
                namesTheAnnotatedType(construction.type)
            {
                return true
            }
            guard let link = Self.link(expression),
                let member = RedundantTypeAnnotation.reference(at: link.member, in: file, symbols: symbols),
                let declaration = demangler.declaration(ofSymbol: member.symbol),
                let reading = Member(demangled: declaration, name: member.name, isCalled: link.isCalled)
            else {
                return false
            }
            var result = reading.result
            if isUnwrapped, result.hasSuffix("?") {
                result.removeLast()
            }
            if result == annotatedType {
                return true
            }
            let isSelf = result == "Self" || (result == "A" && isProtocol(reading.context, declaring: member.symbol))
            return isSelf && (namesTheAnnotatedType(link.base) || hasAnnotatedType(link.base, isUnwrapped: false))
        }

        /* The last link: a member read from a base, `A.b`, or called, `A.f()`. Never `init`, which is a construction, nor an implicit member, which has no base. */
        static func link(_ expression: ExprSyntax) -> (base: ExprSyntax, member: TokenSyntax, isCalled: Bool)? {
            let call = expression.as(FunctionCallExprSyntax.self)
            guard let access = (call?.calledExpression ?? expression).as(MemberAccessExprSyntax.self),
                let base = access.base, access.declName.argumentNames == nil,
                access.declName.baseName.tokenKind != .keyword(.`init`)
            else {
                return nil
            }
            return (base, access.declName.baseName, call != nil)
        }

        /*
         Whether a member's context is a protocol: some prefix of the member's symbol is the protocol's type, which
         the demangler prints both as a type (`$sSzD`, `Swift.BinaryInteger`) and as a protocol descriptor (`$sSzMp`,
         `protocol descriptor for Swift.BinaryInteger`), each naming the context exactly. Both, because each alone
         is lenient: a protocol descriptor also accepts a bare name with no kind after it (`$s7Control6CircleMp` is
         `protocol descriptor for Control.Circle`, found by this rule's own test), which is no type, and a struct's
         type (`$s7Control6CircleVD`) is no protocol.
         */
        func isProtocol(_ context: String, declaring symbol: String) -> Bool {
            guard symbol.hasPrefix("s:") else { return false }
            let mangled = Array(symbol.dropFirst(2))
            let descriptor = "protocol descriptor for \(context)"
            return (1..<mangled.count).contains { end in
                let prefix = String(mangled[..<end])
                return demangler.demangle("$s\(prefix)D") == context
                    && demangler.demangle("$s\(prefix)Mp") == descriptor
            }
        }
    }

    /*
     A member's declaration as the demangler prints it, read for its context and its result: a function's result
     after its arrow, read by `correctness-no-discarded-try-optional`'s signature, which checks the labels against the index's
     name (`random(in:)`), or a property's type after its colon (`static Foundation.CharacterSet.alphanumerics :
     Foundation.CharacterSet`). A call must be of a function and a read must be of a property: a called property
     of function type, or a function read without a call, is not read.
     */
    struct Member {
        let context: String
        let result: String

        init?(demangled: String, name: String, isCalled: Bool) {
            let characters = Array(demangled)
            guard let depths = CorrectnessNoDiscardedTryOptional.Signature.depths(characters) else { return nil }
            let baseName = name.firstIndex(of: "(").map { String(name[..<$0]) } ?? name
            guard isCalled == (baseName != name) else { return nil }
            let needle = Array(".\(baseName)")
            let followers: [Character] = isCalled ? ["(", "<"] : [" "]
            let starts = characters.indices.filter { start in
                let after = start + needle.count
                return depths[start] == 0 && after < characters.count && Array(characters[start..<after]) == needle
                    && followers.contains(characters[after])
            }
            guard starts.count == 1, let start = starts.first else { return nil }
            if isCalled {
                guard let signature = CorrectnessNoDiscardedTryOptional.Signature(demangled: demangled, name: name)
                else { return nil }
                result = signature.result
            }
            else {
                let colon = Array(" : ")
                let after = start + needle.count
                guard after + colon.count < characters.count, Array(characters[after..<(after + colon.count)]) == colon
                else { return nil }
                result = String(characters[(after + colon.count)...])
            }
            var context = Substring(String(characters[..<start]))
            if context.hasPrefix("static ") {
                context = context.dropFirst("static ".count)
            }
            if context.hasPrefix("(extension in "), let close = context.range(of: "):") {
                context = context[close.upperBound...]
            }
            self.context = String(context)
        }
    }

    /*
     A call that builds a type, with the type it names and the token the index places the initializer at:
     `URL(...)` and `Set<Int>(...)` at the type's name, `URL.init(...)` at `init`, and `[Int](...)` at its bracket.
     An implicit `.init(...)` names no type and is not one.
     */
    static func construction(
        _ expression: ExprSyntax
    ) -> (type: ExprSyntax, initializer: TokenSyntax, call: FunctionCallExprSyntax)? {
        guard let call = expression.as(FunctionCallExprSyntax.self) else { return nil }
        let called = call.calledExpression
        if let member = called.as(MemberAccessExprSyntax.self), member.declName.baseName.tokenKind == .keyword(.`init`)
        {
            guard let base = member.base else { return nil }
            return (base, member.declName.baseName, call)
        }
        if let typeName = Self.typeName(called) {
            return (called, typeName, call)
        }
        if let array = called.as(ArrayExprSyntax.self) {
            return (called, array.leftSquare, call)
        }
        if let dictionary = called.as(DictionaryExprSyntax.self) {
            return (called, dictionary.leftSquare, call)
        }
        return nil
    }

    /* A member read from a named type, `Direction.up` or `Direction.moved(by: 1)`: the type, and the member's name token. */
    static func enumCase(_ expression: ExprSyntax) -> (type: ExprSyntax, member: TokenSyntax)? {
        let member =
            expression.as(MemberAccessExprSyntax.self)
            ?? expression.as(FunctionCallExprSyntax.self)?.calledExpression.as(MemberAccessExprSyntax.self)
        guard let member, let base = member.base, member.declName.argumentNames == nil,
            member.declName.baseName.tokenKind != .keyword(.`init`)
        else {
            return nil
        }
        return (base, member.declName.baseName)
    }

    /*
     Whether a member is a case declared in the enum: its symbol is the enum's followed by the case's
     length-prefixed name, and it ends as an enum case's does, `F` (`s:7Control9DirectionO2upyA2CmF`). A static
     member ends `Z` and a nested type with its kind letter, and an instance member read from the type is a
     function, which no annotation of the type would accept.
     */
    static func isCase(_ symbol: String, named member: TokenSyntax, of enumSymbol: String) -> Bool {
        let name = String(member.text.filter { $0 != "`" })
        return enumSymbol.hasPrefix("s:") && symbol.hasPrefix("\(enumSymbol)\(name.utf8.count)\(name)")
            && symbol.hasSuffix("F")
    }

    /*
     Whether a construction gives the type itself, never an optional of it. Where the binding unwraps it anyway,
     any construction does. Otherwise the initializer must be a Swift one whose symbol holds no optional (a
     failable one's spells its result `Sg`), or a literal coerced to the type, which records no initializer.
     */
    static func buildsWithoutNil(
        _ construction: (type: ExprSyntax, initializer: TokenSyntax, call: FunctionCallExprSyntax),
        in file: ParsedFile,
        symbols: FileSymbols,
        mayReturnNil: Bool,
    ) -> Bool {
        if mayReturnNil {
            return true
        }
        let location = file.locations.location(for: construction.initializer.positionAfterSkippingLeadingTrivia)
        let initializers = symbols.occurrences(line: location.line, column: location.column).filter {
            $0.isReference && !$0.isImplicit && $0.name.hasPrefix("init(")
        }
        guard let initializer = initializers.first else {
            /* Only `Int(5)` is coerced; `Int.init(5)` and `[Int](...)` are real calls, which record their initializer. */
            return Self.typeName(construction.call.calledExpression) != nil && Self.isCoercedLiteral(construction.call)
        }
        return initializers.count == 1 && initializer.symbol.hasPrefix("s:") && initializer.symbol.hasSuffix("fc")
            && !Self.namesOptional(initializer.symbol)
    }

    /*
     Whether a symbol may name an optional: the sugar `Sg`, `Optional` itself `Sq`, or either repeated (`S2g`).
     Read over the whole symbol, so an optional parameter, or an identifier that happens to hold the letters,
     reads as one too; that only ever leaves a finding out.
     */
    static func namesOptional(_ symbol: String) -> Bool {
        let bytes = Array(symbol.utf8)
        for (index, byte) in bytes.enumerated() where byte == UInt8(ascii: "S") {
            var next = index + 1
            while next < bytes.count, bytes[next] >= UInt8(ascii: "0"), bytes[next] <= UInt8(ascii: "9") {
                next += 1
            }
            if next < bytes.count, bytes[next] == UInt8(ascii: "g") || bytes[next] == UInt8(ascii: "q") {
                return true
            }
        }
        return false
    }

    /* `Int(5)`, `String("text")`: one unlabeled literal and no closure, which the compiler reads as the literal of that type. */
    static func isCoercedLiteral(_ call: FunctionCallExprSyntax) -> Bool {
        guard call.arguments.count == 1, let argument = call.arguments.first, argument.label == nil,
            call.trailingClosure == nil, call.additionalTrailingClosures.isEmpty
        else {
            return false
        }
        let value = argument.expression
        return value.is(IntegerLiteralExprSyntax.self) || value.is(FloatLiteralExprSyntax.self)
            || value.is(StringLiteralExprSyntax.self) || value.is(BooleanLiteralExprSyntax.self)
    }

    /* The name token a type is resolved at: `URL`, the `URL` of `Foundation.URL`, the `Set` of `Set<Int>`, in a type or an expression. */
    static func typeName(_ type: TypeSyntax) -> TokenSyntax? {
        if let identifier = type.as(IdentifierTypeSyntax.self) {
            return identifier.name
        }
        if let member = type.as(MemberTypeSyntax.self) {
            return member.name
        }
        return nil
    }

    static func typeName(_ expression: ExprSyntax) -> TokenSyntax? {
        if let reference = expression.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil {
            return reference.baseName
        }
        if let member = expression.as(MemberAccessExprSyntax.self), member.base != nil,
            member.declName.argumentNames == nil, member.declName.baseName.tokenKind != .keyword(.`init`)
        {
            return member.declName.baseName
        }
        if let specialization = expression.as(GenericSpecializationExprSyntax.self) {
            return Self.typeName(specialization.expression)
        }
        return nil
    }

    /* A type's or an expression's tokens, without trivia: `Set<Int>` the type and `Set<Int>` the expression spell the same. */
    static func spelling(_ node: some SyntaxProtocol) -> [String] {
        node.tokens(viewMode: .sourceAccurate).map(\.text)
    }

    /* The declaration the name at this token is a written reference to, if the compiler recorded exactly one. */
    static func reference(at token: TokenSyntax, in file: ParsedFile, symbols: FileSymbols) -> FileSymbols.Occurrence? {
        let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
        return symbols.reference(line: location.line, column: location.column)
    }

    /* Whether the name at this token refers to the declaration. A type called as an initializer also records the initializer there, so this asks among every reference. */
    static func names(_ token: TokenSyntax, _ symbol: String, in file: ParsedFile, symbols: FileSymbols) -> Bool {
        let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
        return symbols.occurrences(line: location.line, column: location.column).contains {
            $0.isReference && !$0.isImplicit && $0.symbol == symbol
        }
    }

    /* Every annotated binding with an initializer: SwiftLint's two places, a variable declaration's bindings and an optional binding. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Candidate] = []

        override func visit(_ node: PatternBindingSyntax) -> SyntaxVisitorContinueKind {
            if let declaration = node.parent?.parent?.as(VariableDeclSyntax.self), !Self.isIgnored(declaration),
                let annotation = node.typeAnnotation, let initializer = node.initializer?.value
            {
                found.append(Candidate(annotation: annotation, initializer: initializer, isOptionalBinding: false))
            }
            return .visitChildren
        }

        override func visit(_ node: OptionalBindingConditionSyntax) -> SyntaxVisitorContinueKind {
            if let annotation = node.typeAnnotation, let initializer = node.initializer?.value {
                found.append(Candidate(annotation: annotation, initializer: initializer, isOptionalBinding: true))
            }
            return .visitChildren
        }

        static func isIgnored(_ declaration: VariableDeclSyntax) -> Bool {
            declaration.attributes.contains { element in
                guard let attribute = element.as(AttributeSyntax.self) else { return false }
                return RedundantTypeAnnotation.ignoredAttributes.contains(attribute.attributeName.trimmedDescription)
            }
        }
    }
}
