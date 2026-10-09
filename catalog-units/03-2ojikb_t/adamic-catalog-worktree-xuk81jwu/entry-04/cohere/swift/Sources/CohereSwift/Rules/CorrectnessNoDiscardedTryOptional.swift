import Foundation
import SwiftSyntax

/*
 No `try?` whose result is thrown away: `try? remove(file)` as a statement, or `_ = try? remove(file)`.

 The Swift form of "errors are not swallowed" (our TypeScript `no-floating-promises`). A `try?` whose
 optional is used is a legitimate way to say "absent if it fails". A `try?` whose result is discarded
 throws the error away without a word, and a failure nobody can see is the kind that ships. The repair is
 `do`/`catch` with a reason in the catch, even when the reason is "ignored, because …".

 Only the discarded form is matched. Of the 1,428 `try?` the design doc counted, most use their value,
 and those are not the rule's business.

 A discarded `try? await Task.sleep(for:)` (or `nanoseconds:`) is consistency-no-hand-rolled-delay's to
 report, so one site gets one finding with the right repair: that rule names the house's
 `Task.sleepUnlessCancelled`, where this rule's do/catch would be the same pause written out by hand.

 A `try?` that is the only statement of a body is often that body's value, not a discard, and syntax says
 so in most places: a function or getter with a return type, a computed property, an `if` or `switch` used
 as an expression. Those are skipped.

 A closure's sole `try?` is where syntax stops and the types decide, because `compactMap { try? fit($0) }`
 (a value) and `forEach { try? save($0) }` (a discard) are spelled alike. The closure's own signature
 settles it when written (`{ () -> Void in try? save() }`), and so does the annotation of the variable it
 initializes (`let work: () -> Void = { try? save() }`). Otherwise the compiler's answer is read: the
 declaration the called name resolves to, from the index, demangled into its signature by the toolchain's
 own demangler, and the parameter the closure is passed to, labeled, positional or trailing. The `try?` is
 flagged when that parameter's function type returns `()`: `forEach`, `DispatchQueue.async`, a SwiftUI
 `Button` action, `onAppear`, `task`, a `() -> Void` parameter of ours. A generic result is the value, not a
 discard (`map`, `compactMap`, `flatMap`, `DispatchQueue.sync`, `withAnimation`), and is not flagged.

 One generic result is a discard all the same: a concurrency `Task`'s operation returns the task's value,
 which only the task's handle can read, so a `Task { try? await send() }` (or `Task.detached`) whose handle is
 dropped, as a statement or `_ =`, drops the `try?` with it, and is flagged. The handle kept (`let task =`,
 `.value`) keeps the value, and is not. A call whose value is dropped because it is the sole statement of
 another closure is judged by that closure, so `onAppear { Task { try? await send() } }` is found and
 `items.map { _ in Task { try? await send() } }` is not.

 Read as text, the demangled signature is matched to the call carefully, and anything unsure is not flagged:
 a callee not recorded, or recorded more than once, or not Swift's (an Objective-C method such as
 `NotificationCenter.addObserver(forName:object:queue:using:)`); an unlabeled closure beside a variadic
 parameter; a trailing closure that could bind to more than one parameter whose results disagree; a
 parameter whose type is generic, `Any`, or an autoclosure; a closure that is a result builder's body. Those,
 a closure assigned to a property (`onDone = { try? save() }`) or annotated with a typealias,
 `DispatchQueue.sync { try? write() }` as a statement (its value is the call's, dropped by the call, not by
 the closure), and a toolchain without its demangler are the misses. Raised by
 @system_cohere_swift_ahraos_presence, who measured about 30 of Presence's 515 findings here.

 The statement shapes need no types, but a typed rule reports all its shapes together, so a file that no
 source of symbols describes is reported as unchecked for them too, never as clean.
 */
public struct CorrectnessNoDiscardedTryOptional: TypedFileRule {
    public let name = "cohere-swift/correctness-no-discarded-try-optional"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Every shape this rule flags is a `try?`, written as one token pair with nothing between. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("try?")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        /*
         Loaded only when a closure needs it. A toolchain without the demangler leaves the typed shapes
         unjudged, a miss and never a finding, which is the rule's answer whenever it cannot tell.
         */
        let demangler = visitor.closureValues.isEmpty ? nil : try? SwiftDemangler.shared()
        let closures = ClosureResults(file: file, symbols: symbols, demangler: demangler)
        let discardedByClosure = visitor.closureValues.filter { closures.isDiscarded($0.closure) }.map(\.mark)
        let statements = visitor.found.map { (mark: $0, inClosure: false) }
        let marks = (statements + discardedByClosure.map { (mark: $0, inClosure: true) }).sorted {
            $0.mark.position < $1.mark.position
        }
        return marks.map { mark in
            mark.inClosure
                ? file.finding(
                    at: mark.mark,
                    rule: name,
                    message: RuleMessages.CorrectnessNoDiscardedTryOptional.discardedTryOptionalInClosure(),
                )
                : file.finding(
                    at: mark.mark,
                    rule: name,
                    message: RuleMessages.CorrectnessNoDiscardedTryOptional.discardedTryOptional(),
                )
        }
    }

    /*
     Collects the `?` of every `try?` that is a whole statement, or the right side of `_ =`, and apart from
     them each `try?` that is the sole statement of a closure, for the types to judge.
     */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [TokenSyntax] = []
        private(set) var closureValues: [(mark: TokenSyntax, closure: ClosureExprSyntax)] = []

        override func visit(_ node: TryExprSyntax) -> SyntaxVisitorContinueKind {
            guard let mark = node.questionOrExclamationMark, mark.tokenKind == .postfixQuestionMark else {
                return .visitChildren
            }
            if Self.isStatement(node) || Self.isDiscardAssignment(node) {
                if ConsistencyNoHandRolledDelay.Visitor.sleepCall(awaitedBy: node.expression) != nil {
                    return .visitChildren
                }
                found.append(mark)
            }
            else if let item = node.parent?.as(CodeBlockItemSyntax.self), let closure = Self.soleClosure(of: item) {
                closureValues.append((mark, closure))
            }
            return .visitChildren
        }

        /* `try? work()` on its own line: the expression is the whole code-block item, and not the value of its body. */
        static func isStatement(_ node: TryExprSyntax) -> Bool {
            guard let item = node.parent?.as(CodeBlockItemSyntax.self) else { return false }
            return !isImplicitValue(item)
        }

        /* The closure whose whole body is this one item, if it is one. */
        static func soleClosure(of item: CodeBlockItemSyntax) -> ClosureExprSyntax? {
            guard let list = item.parent?.as(CodeBlockItemListSyntax.self), list.count == 1 else { return nil }
            return list.parent?.as(ClosureExprSyntax.self)
        }

        /* Whether the item is the sole statement of a body that returns it: the implicit-return positions. */
        static func isImplicitValue(_ item: CodeBlockItemSyntax) -> Bool {
            guard let list = item.parent?.as(CodeBlockItemListSyntax.self), list.count == 1, let owner = list.parent
            else { return false }
            if owner.is(ClosureExprSyntax.self) || owner.is(AccessorBlockSyntax.self) {
                return true
            }
            if let switchCase = owner.as(SwitchCaseSyntax.self) {
                return switchCase.parent?.parent.map(isUsedAsValue) ?? false
            }
            guard let block = owner.as(CodeBlockSyntax.self), let body = block.parent else { return false }
            if let function = body.as(FunctionDeclSyntax.self) {
                return function.signature.returnClause.map { !isVoid($0.type) } ?? false
            }
            if let accessor = body.as(AccessorDeclSyntax.self) {
                return accessor.accessorSpecifier.tokenKind == .keyword(.get)
            }
            if body.is(IfExprSyntax.self) {
                return isUsedAsValue(body)
            }
            return false
        }

        /* An `if` or `switch` is a value unless it stands as a statement; an `else if` answers for its whole chain. */
        static func isUsedAsValue(_ expression: Syntax) -> Bool {
            var top = expression
            while let parent = top.parent, parent.is(IfExprSyntax.self) {
                top = parent
            }
            guard let parent = top.parent else { return false }
            return !(parent.is(CodeBlockItemSyntax.self) || parent.is(ExpressionStmtSyntax.self))
        }

        static func isVoid(_ type: TypeSyntax) -> Bool {
            if let tuple = type.as(TupleTypeSyntax.self) {
                return tuple.elements.isEmpty
            }
            return type.as(IdentifierTypeSyntax.self)?.name.text == "Void"
        }

        /* `_ = try? work()`, which the unfolded tree spells as the sequence `_`, `=`, `try? work()`. */
        static func isDiscardAssignment(_ node: some SyntaxProtocol) -> Bool {
            guard let elements = node.parent?.as(ExprListSyntax.self), elements.count == 3 else { return false }
            let parts = Array(elements)
            return parts[0].is(DiscardAssignmentExprSyntax.self) && parts[1].is(AssignmentExprSyntax.self)
                && parts[2].id == node.id
        }
    }

    /* Whether a closure's result goes nowhere, decided by its written signature or annotation, or by the parameter the compiler bound it to. */
    struct ClosureResults {
        let file: ParsedFile
        let symbols: FileSymbols
        let demangler: SwiftDemangler?

        func isDiscarded(_ closure: ClosureExprSyntax) -> Bool {
            if let returnClause = closure.signature?.returnClause {
                return Visitor.isVoid(returnClause.type)
            }
            /* A result builder's body collects its statements, and the compiler records the builder at the brace it transforms. */
            let brace = file.locations.location(for: closure.leftBrace.positionAfterSkippingLeadingTrivia)
            guard symbols.occurrences(line: brace.line, column: brace.column).isEmpty else { return false }
            if let binding = closure.parent?.as(InitializerClauseSyntax.self)?.parent?.as(PatternBindingSyntax.self) {
                return binding.typeAnnotation.map { Self.returnsVoid($0.type) } ?? false
            }
            guard let call = Self.call(passing: closure), let callee = callee(of: call),
                let types = Self.boundParameterTypes(of: closure, in: call, signature: callee.signature)
            else {
                return false
            }
            let kinds = types.map(Signature.kind(of:))
            if kinds.allSatisfy({ $0.returnsVoid }) {
                return true
            }
            /* The operation of a concurrency `Task`, whose value only the task's handle can read. */
            guard callee.declaration.isStandardLibrary, callee.declaration.symbol.hasPrefix("s:ScT"),
                let success = Signature.taskSuccess(callee.signature.result)
            else { return false }
            return kinds.allSatisfy { $0 == .function(returning: success) } && isDiscarded(call)
        }

        /* A call whose value is dropped: a statement that is no body's value, `_ =`, or the sole statement of a closure whose own result goes nowhere. */
        func isDiscarded(_ call: FunctionCallExprSyntax) -> Bool {
            guard let item = call.parent?.as(CodeBlockItemSyntax.self) else {
                return Visitor.isDiscardAssignment(call)
            }
            if let closure = Visitor.soleClosure(of: item) {
                return isDiscarded(closure)
            }
            return !Visitor.isImplicitValue(item)
        }

        /* The call the closure is an argument of: in parentheses, trailing, or an additional labeled trailing closure. */
        static func call(passing closure: ClosureExprSyntax) -> FunctionCallExprSyntax? {
            if let call = closure.parent?.as(FunctionCallExprSyntax.self), call.trailingClosure?.id == closure.id {
                return call
            }
            if let argument = closure.parent?.as(LabeledExprSyntax.self) {
                return argument.parent?.parent?.as(FunctionCallExprSyntax.self)
            }
            if let element = closure.parent?.as(MultipleTrailingClosureElementSyntax.self) {
                return element.parent?.parent?.as(FunctionCallExprSyntax.self)
            }
            return nil
        }

        /* The one function declaration the called name resolves to, and its signature. `Task { }` also records the type `Task` at its name, which is not a call. */
        func callee(of call: FunctionCallExprSyntax) -> (declaration: FileSymbols.Occurrence, signature: Signature)? {
            guard let demangler, let token = Self.nameToken(call.calledExpression) else { return nil }
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            let functions = symbols.occurrences(line: location.line, column: location.column).filter {
                $0.isReference && !$0.isImplicit && $0.name.contains("(")
            }
            guard let declaration = functions.first,
                functions.allSatisfy({ $0.symbol == declaration.symbol && $0.name == declaration.name })
            else { return nil }
            guard let demangled = demangler.declaration(ofSymbol: declaration.symbol),
                let signature = Signature(demangled: demangled, name: declaration.name)
            else { return nil }
            return (declaration, signature)
        }

        /* The token the index places a call's declaration at: `forEach` in `items.forEach`, `Task` in `Task { }`. */
        static func nameToken(_ expression: ExprSyntax) -> TokenSyntax? {
            if let reference = expression.as(DeclReferenceExprSyntax.self) {
                return reference.baseName
            }
            if let member = expression.as(MemberAccessExprSyntax.self) {
                return member.declName.baseName
            }
            if let specialized = expression.as(GenericSpecializationExprSyntax.self) {
                return nameToken(specialized.expression)
            }
            return nil
        }

        /*
         The type of every parameter the closure could be bound to; nil when unsure. A labeled closure binds to
         the one parameter of its label. A closure in parentheses binds as the compiler binds arguments, each to
         the next parameter of its label (`_` for none). A trailing closure binds to a parameter after those,
         and before the label of the first additional trailing closure, that can take a closure; when that is
         more than one, the compiler's choice turns on defaults the signature does not show, so all are returned.
         A variadic parameter takes any number of arguments, so with one in the signature only a label is trusted.
         */
        static func boundParameterTypes(
            of closure: ClosureExprSyntax,
            in call: FunctionCallExprSyntax,
            signature: Signature,
        ) -> [String]? {
            let labels = signature.labels
            let isVariadic = signature.parameters.contains { $0.hasSuffix("...") }
            func labeled(_ label: String) -> [String]? {
                let matches = labels.indices.filter { labels[$0] == label }
                return matches.count == 1 ? matches.map { signature.parameters[$0] } : nil
            }
            if let element = closure.parent?.as(MultipleTrailingClosureElementSyntax.self) {
                return labeled(element.label.text)
            }
            if let argument = closure.parent?.as(LabeledExprSyntax.self), let label = argument.label {
                return labeled(label.text)
            }
            guard !isVariadic else { return nil }
            var bound: [Int] = []
            for argument in call.arguments {
                let label = argument.label?.text ?? "_"
                guard let index = labels.indices.first(where: { $0 > (bound.last ?? -1) && labels[$0] == label }) else {
                    return nil
                }
                bound.append(index)
            }
            if let position = Array(call.arguments).firstIndex(where: { $0.expression.id == closure.id }) {
                return [signature.parameters[bound[position]]]
            }
            let next = (bound.last ?? -1) + 1
            var end = labels.count
            if let first = call.additionalTrailingClosures.first {
                guard let index = labels.indices.first(where: { $0 >= next && labels[$0] == first.label.text }) else {
                    return nil
                }
                end = index
            }
            guard next < end else { return nil }
            let candidates = signature.parameters[next..<end].filter { Signature.kind(of: $0) != .other }
            return candidates.isEmpty ? nil : Array(candidates)
        }

        /* `() -> Void`, `@escaping () -> Void`, `(() -> Void)?`: an annotation that says the closure returns nothing. */
        static func returnsVoid(_ type: TypeSyntax) -> Bool {
            if let attributed = type.as(AttributedTypeSyntax.self) {
                return returnsVoid(attributed.baseType)
            }
            if let optional = type.as(OptionalTypeSyntax.self) {
                return returnsVoid(optional.wrappedType)
            }
            if let tuple = type.as(TupleTypeSyntax.self), tuple.elements.count == 1, let element = tuple.elements.first,
                element.firstName == nil
            {
                return returnsVoid(element.type)
            }
            return type.as(FunctionTypeSyntax.self).map { Visitor.isVoid($0.returnClause.type) } ?? false
        }
    }

    /*
     A function's signature as the demangler prints it, read carefully: `(extension in Dispatch):__C.OS_dispatch_queue.async(group: __C.OS_dispatch_group?, qos: Dispatch.DispatchQoS, flags: Dispatch.DispatchWorkItemFlags, execute: @escaping @convention(block) () -> ()) -> ()`.
     Parentheses, brackets and angle brackets nest; the `>` of an arrow closes nothing. The labels come from
     the declaration's name in the index (`async(group:qos:flags:execute:)`), and a signature whose parameters
     do not match them one for one, or whose function cannot be found by name, is not read at all.
     */
    struct Signature {
        let labels: [String]
        let parameters: [String]
        let result: String

        /* What a parameter's type can take: a function, with its result as printed; a type that no closure converts to; or one that cannot be told. */
        enum Kind: Equatable {
            case function(returning: String)
            case other
            case unknown

            var returnsVoid: Bool {
                self == .function(returning: "()") || self == .function(returning: "Swift.Void")
            }
        }

        static let specifiers = [
            "__owned", "__shared", "sending", "inout", "borrowing", "consuming", "isolated", "nonisolated(nonsending)",
        ]

        init?(demangled: String, name: String) {
            guard let open = name.firstIndex(of: "("), name.hasSuffix(")") else { return nil }
            let baseName = Array(name[..<open])
            let labels = name[name.index(after: open)..<name.index(before: name.endIndex)].split(
                separator: ":",
                omittingEmptySubsequences: false,
            ).dropLast().map(String.init)
            let characters = Array(demangled)
            guard !baseName.isEmpty, let depths = Self.depths(characters) else { return nil }
            let needle = ["."] + baseName
            let starts = characters.indices.filter { start in
                let after = start + needle.count
                return depths[start] == 0 && after < characters.count && Array(characters[start..<after]) == needle
                    && (characters[after] == "(" || characters[after] == "<")
            }
            guard starts.count == 1, var parametersOpen = starts.first.map({ $0 + needle.count }) else { return nil }
            if characters[parametersOpen] == "<" {
                guard let close = Self.closing(characters, depths, from: parametersOpen) else { return nil }
                parametersOpen = close + 1
            }
            guard parametersOpen < characters.count, characters[parametersOpen] == "(",
                let parametersClose = Self.closing(characters, depths, from: parametersOpen)
            else { return nil }
            var entries: [String] = []
            var entryStart = parametersOpen + 1
            for index in entryStart...parametersClose
            where index == parametersClose || (depths[index] == 1 && characters[index] == ",") {
                entries.append(String(characters[entryStart..<index]).trimmingCharacters(in: .whitespaces))
                entryStart = index + 1
            }
            if entries == [""] {
                entries = []
            }
            guard entries.count == labels.count else { return nil }
            var parameters: [String] = []
            for (entry, label) in zip(entries, labels) {
                let printed = Self.printedLabel(entry)
                guard (printed?.label ?? "_") == label else { return nil }
                parameters.append(printed?.type ?? entry)
            }
            guard let arrow = Self.arrow(characters, depths, after: parametersClose, depth: 0) else { return nil }
            self.labels = labels
            self.parameters = parameters
            result = String(characters[(arrow + 2)...]).trimmingCharacters(in: .whitespaces)
        }

        /* `group: __C.OS_dispatch_group?` is labeled; `(A.Element) throws -> ()` is not. */
        static func printedLabel(_ entry: String) -> (label: String, type: String)? {
            guard let colon = entry.range(of: ": ") else { return nil }
            let label = entry[..<colon.lowerBound]
            guard !label.isEmpty, label.allSatisfy({ $0.isLetter || $0.isNumber || $0 == "_" }) else { return nil }
            return (String(label), String(entry[colon.upperBound...]))
        }

        /* How deep each character sits inside brackets, the brackets themselves at the depth outside them. Nil for text that does not balance. */
        static func depths(_ characters: [Character]) -> [Int]? {
            var depths: [Int] = []
            var depth = 0
            for index in characters.indices {
                let character = characters[index]
                if character == ")" || character == "]"
                    || (character == ">" && (index == 0 || characters[index - 1] != "-"))
                {
                    depth -= 1
                }
                guard depth >= 0 else { return nil }
                depths.append(depth)
                if character == "(" || character == "[" || character == "<" {
                    depth += 1
                }
            }
            return depth == 0 ? depths : nil
        }

        /* The bracket closing the one opened at `open`. */
        static func closing(_ characters: [Character], _ depths: [Int], from open: Int) -> Int? {
            characters.indices.first { $0 > open && depths[$0] == depths[open] && ")]>".contains(characters[$0]) }
        }

        /* The first `->` after a place, at a depth. */
        static func arrow(_ characters: [Character], _ depths: [Int], after start: Int, depth: Int) -> Int? {
            characters.indices.first {
                $0 > start && $0 + 1 < characters.count && depths[$0] == depth && characters[$0] == "-"
                    && characters[$0 + 1] == ">"
            }
        }

        /*
         What a parameter of this printed type can take. Attributes and ownership come off first; an optional
         function is a function. A module-qualified type (`Swift.String?`, `Dispatch.DispatchQoS`) takes no
         closure, as the compiler skips it when binding a trailing closure; a generic parameter (`A`,
         `A.Element`), `Any`, an existential or an autoclosure might, and is unknown.
         */
        static func kind(of type: String) -> Kind {
            var text = Substring(type.trimmingCharacters(in: .whitespaces))
            if text.hasSuffix("...") {
                return .unknown
            }
            while true {
                if text.hasPrefix("@autoclosure") {
                    return .unknown
                }
                if text.hasPrefix("@") {
                    text = text.dropFirst().drop { $0.isLetter || $0.isNumber || $0 == "_" || $0 == "." }
                    if text.first == "(" {
                        let characters = Array(text)
                        guard let depths = depths(characters), let close = closing(characters, depths, from: 0) else {
                            return .unknown
                        }
                        text = text.dropFirst(close + 1)
                    }
                    text = text.drop { $0 == " " }
                    continue
                }
                if let specifier = specifiers.first(where: { text.hasPrefix($0 + " ") }) {
                    text = text.dropFirst(specifier.count + 1)
                    continue
                }
                break
            }
            let characters = Array(text)
            guard let first = characters.first, let depths = depths(characters) else { return .unknown }
            if first == "(" {
                guard let close = closing(characters, depths, from: 0) else { return .unknown }
                let rest = String(characters[(close + 1)...])
                if rest.isEmpty {
                    return String(characters).contains("->") ? .unknown : .other
                }
                if rest == "?" {
                    return kind(of: String(characters[1..<close]))
                }
                guard let arrow = arrow(characters, depths, after: close, depth: 0) else { return .unknown }
                return .function(returning: String(characters[(arrow + 2)...]).trimmingCharacters(in: .whitespaces))
            }
            if first == "[" {
                return .other
            }
            /* A keyword and a space (`any P`, `some P`, `each T`): not a plain type. */
            if let space = characters.firstIndex(of: " "),
                characters[..<space].allSatisfy({ $0.isLowercase || $0 == "_" })
            {
                return .unknown
            }
            let head = characters.prefix { !".<?![".contains($0) }
            let isQualified = characters.dropFirst(head.count).first == "."
            let isGenericParameter =
                head.first.map { $0.isUppercase && $0.isASCII } == true && head.dropFirst().allSatisfy(\.isNumber)
            return isQualified && !isGenericParameter && head.first != "τ" ? .other : .unknown
        }

        /* `A` in `Swift.Task<A, Swift.Never>`: the task's value type, which its operation returns. */
        static func taskSuccess(_ result: String) -> String? {
            guard result.hasPrefix("Swift.Task<"), result.hasSuffix(">") else { return nil }
            let arguments = result.dropFirst("Swift.Task<".count).dropLast()
            guard let comma = arguments.firstIndex(of: ","), !arguments[..<comma].contains("<") else { return nil }
            return String(arguments[..<comma]).trimmingCharacters(in: .whitespaces)
        }
    }
}
