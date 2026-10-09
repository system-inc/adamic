import SwiftOperators
import SwiftSyntax

/*
 No write-only collection: a local array, set or dictionary that is only ever added to, and never read.

     var ranks: [Int: Int] = [:]
     for (index, row) in rows.enumerated() {
         ranks[row.id] = index
     }

 Every use puts something in or takes something out, and nothing reads what it holds, so the work that fills it is
 thrown away. Either the code that was meant to read it is missing (a value kept for the result and never copied
 in), or the collection is dead and should be deleted along with its writes. Which one is the author's call, so
 there is no fix. swiftc says nothing here: measured with swiftc 6.4 on 2026-10-03, `append`, `insert`, `+=`,
 `_ = popLast()` and a subscript assignment all count as uses, and only a binding that is reassigned and never
 read is warned about ("variable 'xs' was written to, but never read").

 It ports `nexus/correctness-no-write-only-collection`, condition for condition, each one required:
 - A local `var` declared inside a function, initializer, accessor or closure, one identifier, with an
   initializer, no attribute (a property wrapper), no modifier (`lazy`) and no accessor block (`didSet` reads). A
   binding at file scope is a global any file can read, as TypeScript declines a module-level one; so is one inside
   `#if`, which the code after the `#endif` can name.
 - Its type is the standard library's `Array`, `Set` or `Dictionary`, read from the source. `[T]` and `[K: V]`,
   as an annotation or a constructor (`[Int]()`), and an array or dictionary literal with no annotation, are those
   types by the language's own rules, whatever else is declared. `Array`, `Set` and `Dictionary` spelled by name,
   in an annotation or as a constructor (`Set<Int>()`, `Set(rows)`), are taken only when the build's index resolves
   that name to the standard library's (`s:Sa`, `s:Sh`, `s:SD`), as TypeScript ties `new Map` to the default
   library's: a type of ours named `Set` is not one.
 - Every reference to it is a write, and there is at least one. A collection never referenced is swiftc's ("never
   used") as it is `no-unused-vars`'s there. References are found by Swift's scope rules, read from the tree with
   the scope reading `concurrency-no-check-then-write` uses, since the index does not record locals: a binding of
   the same name in an inner scope (a `let`, a loop's or a closure's parameter, an `if let`) is another variable,
   and its uses are not this one's.
 - A write is one of three things, and anything else is a read:
   1. A call of one of the collection's mutating methods on the binding itself, whose result is dropped, each
      resolved by the index to the standard library's declaration, so an overload a module of ours adds in an
      `extension Array` is not one. The methods are TypeScript's mutators where Swift has them (`push` and
      `unshift` are `append` and `insert`, `pop` and `shift` are `popLast`, `removeLast` and `removeFirst`,
      `splice` is `remove(at:)`, `removeSubrange` and `replaceSubrange`; `Map.set`, `delete` and `clear` are
      `updateValue`, `removeValue(forKey:)` and `removeAll`; `Set.add`, `delete` and `clear` are `insert`,
      `remove` and `removeAll`), and the ones Swift adds that read nothing: `append(contentsOf:)`,
      `insert(contentsOf:at:)`, `reserveCapacity`, `removeAll(keepingCapacity:)`, `Set`'s `update(with:)`,
      `popFirst`, `formUnion`, `formIntersection`, `formSymmetricDifference` and `subtract`.
   2. A plain `=` through a subscript on the binding: `ranks[id] = rank`, `ranks[id, default: 0] = rank`,
      `slots[index] = value`.
   3. `+=` on an array, which TypeScript has no counterpart for: it is `append(contentsOf:)`, and resolved by the
      index to the standard library's operator.

 Where Swift needs a different exactness than TypeScript:
 - `var`, where TypeScript reads only `const`. A Swift collection is a value, so only a `var` can be written at all.
   TypeScript declines `let` because a reassigned binding is not one collection; here a reassignment of the whole
   binding (`ranks = [:]`) is a read, so the binding is declined the same way.
 - Any initializer, where TypeScript reads only a literal or a fresh `Map` or `Set`. TypeScript needs a fresh one
   because `const copy = original` aliases it, and a push to `copy` is a push to `original`. A Swift array, set or
   dictionary is copied on assignment, so `var copy: [Int] = original` is a collection of its own, and writes to it
   reach nothing else.
 - A dropped result. TypeScript counts a call only as a statement of its own, and declines every arrow function
   with an expression body, because the arrow returns the call's result. In Swift a single-expression closure,
   function or `if` branch returns its value the same way, so the rule asks what the method returns. A method that
   returns nothing (`append`, `removeAll`, `formUnion`) is a write wherever it stands as a statement, including
   `rows.forEach { ids.append($0.id) }`, the most common Swift spelling, since there is no value to hand on. A
   method that returns one (`insert` on a set, `remove`, `updateValue`, `popLast`) is a write only where its value
   cannot be handed on: a statement in a list of several, the single statement of a loop body, of an `if` or
   `switch` that is itself such a statement, or `_ = ...`. `try` and `await` in front of a call change neither.
 - `seen.insert(x).inserted`, Swift's dedupe idiom, reads the insert's answer, which is a read, as TypeScript reads
   `const length = list.push(x)`. So does every `if` or `guard` on it.
 - A method that hands the elements to code of ours reads them: `removeAll(where:)`, `merge(_:uniquingKeysWith:)`,
   `sort(by:)`. They are reads here, and none is in TypeScript's list either.
 - A write can trap: `slots[9] = value` past the end, `removeFirst()` on an empty array. A trap is a bug, not a
   use of what the collection holds, so these are writes, as their TypeScript counterparts (which never trap) are.

 What it accepts as safe, and why: anything that reads the collection, its elements, its count or its identity
 (iterating it, `count`, `contains`, `[key]` read, a compound `+=` through a subscript, passing it, returning it,
 interpolating it, capturing it in a capture list, passing it `inout`); a mutating call whose result is used; a
 local type's member of the same name; any reference the scope reading cannot place.

 The one assumption that could report wrongly: an element type's `hash(into:)` and `==`, which `insert`,
 `remove` and a dictionary's subscript call, are taken to have no effect beyond their answer. TypeScript's `Map`
 and `Set` call no code of ours at all.

 Known misses, every one a finding not made and never one invented:
 - A write through an element: `groups[key, default: []].append(item)`, `counts[key, default: 0] += 1`,
   `rows[index].done = true`. TypeScript reads `groups.get(key).push(item)` and `counts[0] += 1` as reads, and so
   does this port, though each only writes.
 - A value-returning write as the single statement of a closure, a function, a `do` or a `catch`
   (`rows.forEach { seen.insert($0) }`), where Swift may hand the value on. TypeScript's arrow-body decline.
 - A collection whose type the source does not spell: `var ids = makeIds()`, `var ids = rows.map(\.id)`, and
   `Swift.Array` written with its module.
 - `ContiguousArray`, `ArraySlice`, Foundation's and the collections package's types, and every mutating method
   not listed above (`sort()`, `reverse()`, `shuffle()`, `swapAt`).
 - A collection in top-level code (`main.swift`), declared inside `#if`, `lazy`, observed or wrapped.
 */
public struct CorrectnessNoWriteOnlyCollection: TypedFileRule {
    public let name = "cohere-swift/correctness-no-write-only-collection"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Every flagged binding is a `var`, and its type is spelled with a bracket or one of the three names. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("var")
            && (file.source.contains("[") || file.source.contains("Set") || file.source.contains("Array")
                || file.source.contains("Dictionary"))
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        /* Folded, so `ranks[id] = rank` is one assignment. Folding moves no token, so positions in the copy are positions in the file. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let collector = Collector(viewMode: .sourceAccurate)
        collector.walk(folded)
        let reader = Reader(file: file, symbols: symbols)
        return collector.bindings.compactMap { variable, binding, token in
            guard let kind = reader.kind(of: binding), reader.isOnlyWritten(token, declaredBy: variable, kind: kind)
            else { return nil }
            return file.finding(
                at: token,
                rule: name,
                message: RuleMessages.CorrectnessNoWriteOnlyCollection.writeOnlyCollection(),
            )
        }
    }

    /* Every plain local `var` binding: one identifier, with a value, and nothing that runs code on a read or a write. */
    final class Collector: SyntaxVisitor {
        private(set) var bindings: [(VariableDeclSyntax, PatternBindingSyntax, TokenSyntax)] = []

        override func visit(_ node: VariableDeclSyntax) -> SyntaxVisitorContinueKind {
            guard node.bindingSpecifier.tokenKind == .keyword(.var), node.attributes.isEmpty, node.modifiers.isEmpty,
                let list = node.parent?.parent?.as(CodeBlockItemListSyntax.self), let scope = list.parent,
                !scope.is(SourceFileSyntax.self), !scope.is(IfConfigClauseSyntax.self),
                ConcurrencyNoCheckThenWrite.enclosingFunction(Syntax(node)) != nil,
                Self.functionBeforeType(Syntax(node))
            else {
                return .visitChildren
            }
            for binding in node.bindings where binding.accessorBlock == nil && binding.initializer != nil {
                if let identifier = binding.pattern.as(IdentifierPatternSyntax.self) {
                    bindings.append((node, binding, identifier.identifier))
                }
            }
            return .visitChildren
        }

        /* Whether the nearest enclosing function comes before the nearest enclosing type: a local, not a member of a local type. */
        static func functionBeforeType(_ node: Syntax) -> Bool {
            var current = node.parent
            while let candidate = current {
                if ConcurrencyNoCheckThenWrite.isFunctionLike(candidate) {
                    return true
                }
                if ConcurrencyNoCheckThenWrite.isTypeDeclaration(candidate) {
                    return false
                }
                current = candidate.parent
            }
            return false
        }
    }

    // MARK: The collections

    enum Kind {
        case array
        case set
        case dictionary
    }

    /* A mutating method by its base name and argument labels (nil for an unlabeled argument), and whether it returns a value its caller could read. */
    struct Writer: Hashable {
        let name: String
        let labels: [String?]
        let returnsValue: Bool
    }

    static let arrayWriters: Set<Writer> = [
        Writer(name: "append", labels: [nil], returnsValue: false),
        Writer(name: "append", labels: ["contentsOf"], returnsValue: false),
        Writer(name: "insert", labels: [nil, "at"], returnsValue: false),
        Writer(name: "insert", labels: ["contentsOf", "at"], returnsValue: false),
        Writer(name: "remove", labels: ["at"], returnsValue: true),
        Writer(name: "removeFirst", labels: [], returnsValue: true),
        Writer(name: "removeFirst", labels: [nil], returnsValue: false),
        Writer(name: "removeLast", labels: [], returnsValue: true),
        Writer(name: "removeLast", labels: [nil], returnsValue: false),
        Writer(name: "popLast", labels: [], returnsValue: true),
        Writer(name: "removeAll", labels: [], returnsValue: false),
        Writer(name: "removeAll", labels: ["keepingCapacity"], returnsValue: false),
        Writer(name: "removeSubrange", labels: [nil], returnsValue: false),
        Writer(name: "replaceSubrange", labels: [nil, "with"], returnsValue: false),
        Writer(name: "reserveCapacity", labels: [nil], returnsValue: false),
    ]

    static let setWriters: Set<Writer> = [
        Writer(name: "insert", labels: [nil], returnsValue: true),
        Writer(name: "update", labels: ["with"], returnsValue: true),
        Writer(name: "remove", labels: [nil], returnsValue: true),
        Writer(name: "removeFirst", labels: [], returnsValue: true),
        Writer(name: "popFirst", labels: [], returnsValue: true),
        Writer(name: "removeAll", labels: [], returnsValue: false),
        Writer(name: "removeAll", labels: ["keepingCapacity"], returnsValue: false),
        Writer(name: "formUnion", labels: [nil], returnsValue: false),
        Writer(name: "formIntersection", labels: [nil], returnsValue: false),
        Writer(name: "formSymmetricDifference", labels: [nil], returnsValue: false),
        Writer(name: "subtract", labels: [nil], returnsValue: false),
        Writer(name: "reserveCapacity", labels: [nil], returnsValue: false),
    ]

    static let dictionaryWriters: Set<Writer> = [
        Writer(name: "updateValue", labels: [nil, "forKey"], returnsValue: true),
        Writer(name: "removeValue", labels: ["forKey"], returnsValue: true),
        Writer(name: "removeAll", labels: [], returnsValue: false),
        Writer(name: "removeAll", labels: ["keepingCapacity"], returnsValue: false),
        Writer(name: "reserveCapacity", labels: [nil], returnsValue: false),
    ]

    static func writers(of kind: Kind) -> Set<Writer> {
        switch kind {
            case .array:
                return arrayWriters
            case .set:
                return setWriters
            case .dictionary:
                return dictionaryWriters
        }
    }

    /* The standard library's types, by their index symbols: `Array`, `Set`, `Dictionary`. */
    static let typeSymbols: [String: (kind: Kind, symbol: String)] = [
        "Array": (.array, "s:Sa"),
        "Set": (.set, "s:Sh"),
        "Dictionary": (.dictionary, "s:SD"),
    ]

    // MARK: Reading one binding

    struct Reader {
        let file: ParsedFile
        let symbols: FileSymbols

        /* The collection a binding holds, by its annotation or, without one, by its initializer. Nil for anything the source does not spell. */
        func kind(of binding: PatternBindingSyntax) -> Kind? {
            if let annotation = binding.typeAnnotation {
                return kind(ofType: annotation.type)
            }
            guard let value = binding.initializer?.value else { return nil }
            let expression = ConcurrencyNoCheckThenWrite.withoutParentheses(value)
            /* A literal with no annotation takes its default type, which nothing can redeclare. An empty one does not compile without one. */
            if let array = expression.as(ArrayExprSyntax.self) {
                return array.elements.isEmpty ? nil : .array
            }
            if let dictionary = expression.as(DictionaryExprSyntax.self) {
                guard case .elements = dictionary.content else { return nil }
                return .dictionary
            }
            guard let call = expression.as(FunctionCallExprSyntax.self) else { return nil }
            let callee = ConcurrencyNoCheckThenWrite.withoutParentheses(call.calledExpression)
            /* `[Int]()` and `[String: Int]()`: the brackets are the type. */
            if callee.is(ArrayExprSyntax.self) {
                return .array
            }
            if callee.is(DictionaryExprSyntax.self) {
                return .dictionary
            }
            if let reference = callee.as(DeclReferenceExprSyntax.self) {
                return named(reference.baseName)
            }
            if let specialized = callee.as(GenericSpecializationExprSyntax.self),
                let reference = specialized.expression.as(DeclReferenceExprSyntax.self)
            {
                return named(reference.baseName)
            }
            return nil
        }

        func kind(ofType type: TypeSyntax) -> Kind? {
            if type.is(ArrayTypeSyntax.self) {
                return .array
            }
            if type.is(DictionaryTypeSyntax.self) {
                return .dictionary
            }
            if let identifier = type.as(IdentifierTypeSyntax.self) {
                return named(identifier.name)
            }
            return nil
        }

        /* `Array`, `Set` or `Dictionary` by name, when the index resolves this token to the standard library's type. */
        func named(_ token: TokenSyntax) -> Kind? {
            guard let entry = CorrectnessNoWriteOnlyCollection.typeSymbols[token.text] else { return nil }
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            let resolved = symbols.occurrences(line: location.line, column: location.column).contains {
                $0.isReference && !$0.isImplicit && $0.symbol == entry.symbol
            }
            return resolved ? entry.kind : nil
        }

        /* Whether the token at this place resolves to a standard library declaration, by exactly one written reference. */
        func isStandardLibrary(_ token: TokenSyntax) -> Bool {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            return symbols.reference(line: location.line, column: location.column)?.isStandardLibrary ?? false
        }

        /*
         Whether every use of the binding after its declaration writes it, and one does. Every identifier token of its
         name is judged: a reference that resolves to it is a write or a read, a member name (`row.ids`), an argument
         label and the declaration of another binding are not uses, and anything else is taken as a read.
         */
        func isOnlyWritten(_ declared: TokenSyntax, declaredBy variable: VariableDeclSyntax, kind: Kind) -> Bool {
            guard let item = variable.parent?.as(CodeBlockItemSyntax.self),
                let list = item.parent?.as(CodeBlockItemListSyntax.self)
            else { return false }
            let name = CorrectnessNoWriteOnlyCollection.bareName(declared)
            var writes = 0
            var reachedDeclaration = false
            for statement in list {
                if statement.id == item.id {
                    reachedDeclaration = true
                }
                guard reachedDeclaration else { continue }
                for token in statement.tokens(viewMode: .sourceAccurate)
                where token.id != declared.id && CorrectnessNoWriteOnlyCollection.bareName(token) == name {
                    switch use(of: token, declared: declared, kind: kind) {
                        case .none:
                            continue
                        case .write:
                            writes += 1
                        case .read:
                            return false
                    }
                }
            }
            return writes > 0
        }

        enum Use {
            case none
            case write
            case read
        }

        func use(of token: TokenSyntax, declared: TokenSyntax, kind: Kind) -> Use {
            guard case .identifier = token.tokenKind, let parent = token.parent else {
                return .read
            }
            if let reference = parent.as(DeclReferenceExprSyntax.self), reference.baseName.id == token.id {
                /* `row.ids` and `.ids` name a member of something else, not this local. */
                if let member = reference.parent?.as(MemberAccessExprSyntax.self), member.declName.id == reference.id {
                    return .none
                }
                guard resolvesTo(declared, reference) else { return .none }
                return isWrite(reference, kind: kind) ? .write : .read
            }
            if let label = parent.as(LabeledExprSyntax.self), label.label?.id == token.id {
                return .none
            }
            /* Another binding of the same name: a `let`, a loop variable, a parameter. `if let ids` alone also reads the outer one. */
            if let pattern = parent.as(IdentifierPatternSyntax.self) {
                if let optional = pattern.parent?.as(OptionalBindingConditionSyntax.self), optional.initializer == nil {
                    return .read
                }
                return .none
            }
            if parent.is(ClosureShorthandParameterSyntax.self) || parent.is(ClosureParameterSyntax.self)
                || parent.is(FunctionParameterSyntax.self)
            {
                return .none
            }
            /*
             Anything else is a read, which keeps the binding silent: a capture list's `[ids]` (the name token of a
             capture, which copies the collection and binds the name the closure's uses then resolve to), a local
             function or type of the same name, a key path component.
             */
            return .read
        }

        /*
         Whether a reference names this binding: no other binding of its name between them. A reference no scope places
         (a local function or type of that name, a binding the scope reading does not know) is taken as this binding's.
         */
        func resolvesTo(_ declared: TokenSyntax, _ reference: DeclReferenceExprSyntax) -> Bool {
            let name = reference.baseName.text
            var child = Syntax(reference)
            while let node = child.parent {
                if node.is(MemberBlockItemListSyntax.self) || node.is(SourceFileSyntax.self) {
                    return true
                }
                switch ConcurrencyNoCheckThenWrite.lookup(name, in: node, below: child) {
                    case .declared(let found)?:
                        return found.id == declared.id
                    case .unknown?:
                        return true
                    case nil:
                        break
                }
                child = node
            }
            return true
        }

        /* One reference judged: a dropped mutating call on it, a plain `=` through its subscript, or `+=` on an array. */
        func isWrite(_ reference: DeclReferenceExprSyntax, kind: Kind) -> Bool {
            guard let parent = reference.parent else { return false }
            if let member = parent.as(MemberAccessExprSyntax.self), member.base?.id == reference.id,
                member.declName.argumentNames == nil,
                let call = member.parent?.as(FunctionCallExprSyntax.self), call.calledExpression.id == member.id,
                call.trailingClosure == nil, call.additionalTrailingClosures.isEmpty
            {
                let labels = call.arguments.map { $0.label?.text }
                guard
                    let writer = CorrectnessNoWriteOnlyCollection.writers(of: kind).first(where: {
                        $0.name == member.declName.baseName.text && $0.labels == labels
                    }),
                    isStandardLibrary(member.declName.baseName)
                else {
                    return false
                }
                return CorrectnessNoWriteOnlyCollection.dropsResult(
                    of: ExprSyntax(call),
                    returnsValue: writer.returnsValue,
                )
            }
            if let subscriptCall = parent.as(SubscriptCallExprSyntax.self),
                subscriptCall.calledExpression.id == reference.id,
                subscriptCall.trailingClosure == nil, subscriptCall.additionalTrailingClosures.isEmpty,
                let assignment = subscriptCall.parent?.as(InfixOperatorExprSyntax.self),
                assignment.operator.is(AssignmentExprSyntax.self),
                assignment.leftOperand.id == subscriptCall.id
            {
                let labels = subscriptCall.arguments.map { $0.label?.text }
                let keyed: Bool
                switch kind {
                    case .array:
                        keyed = labels == [nil]
                    case .dictionary:
                        keyed = labels == [nil] || labels == [nil, "default"]
                    case .set:
                        keyed = false
                }
                return keyed && isStandardLibrary(subscriptCall.leftSquare)
            }
            if kind == .array, let infix = parent.as(InfixOperatorExprSyntax.self),
                infix.leftOperand.id == reference.id,
                let operation = infix.operator.as(BinaryOperatorExprSyntax.self), operation.operator.text == "+="
            {
                return isStandardLibrary(operation.operator)
            }
            return false
        }
    }

    // MARK: Dropped results

    /*
     Whether a call's result goes nowhere: it is a statement, through `try` and `await`, or the right side of `_ =`. A
     call that returns a value must also stand where Swift cannot return it implicitly.
     */
    static func dropsResult(of call: ExprSyntax, returnsValue: Bool) -> Bool {
        var current = Syntax(call)
        while let wrapper = current.parent, wrapper.is(TryExprSyntax.self) || wrapper.is(AwaitExprSyntax.self) {
            current = wrapper
        }
        if let discard = current.parent?.as(InfixOperatorExprSyntax.self),
            discard.operator.is(AssignmentExprSyntax.self),
            discard.leftOperand.is(DiscardAssignmentExprSyntax.self), discard.rightOperand.id == current.id
        {
            return discard.parent?.is(CodeBlockItemSyntax.self) == true
        }
        guard let item = current.parent?.as(CodeBlockItemSyntax.self) else { return false }
        return !returnsValue || discardsValue(item)
    }

    /*
     Whether a statement's value is dropped: one of several statements (Swift returns implicitly only from a single
     expression), or the only statement of a loop body, a `defer`, or a branch of an `if` or `switch` that is itself
     such a statement. A closure, a function, an accessor, a `do` and a `catch` may hand it on.
     */
    static func discardsValue(_ item: CodeBlockItemSyntax) -> Bool {
        guard let list = item.parent?.as(CodeBlockItemListSyntax.self) else { return false }
        if list.count > 1 {
            return true
        }
        if let construct = list.parent?.as(CodeBlockSyntax.self)?.parent {
            if construct.is(ForStmtSyntax.self) || construct.is(WhileStmtSyntax.self)
                || construct.is(RepeatStmtSyntax.self) || construct.is(DeferStmtSyntax.self)
            {
                return true
            }
            if construct.is(IfExprSyntax.self) {
                return isStatement(construct)
            }
            return false
        }
        if let switchCase = list.parent?.as(SwitchCaseSyntax.self),
            let switchExpression = switchCase.parent?.parent?.as(SwitchExprSyntax.self)
        {
            return isStatement(Syntax(switchExpression))
        }
        return false
    }

    /* Whether an `if` (from any branch of an `else if` chain) or a `switch` stands as a statement whose value is dropped. */
    static func isStatement(_ node: Syntax) -> Bool {
        var current = node
        while let outer = current.parent?.as(IfExprSyntax.self) {
            current = Syntax(outer)
        }
        guard let statement = current.parent?.as(ExpressionStmtSyntax.self),
            let item = statement.parent?.as(CodeBlockItemSyntax.self)
        else { return false }
        return discardsValue(item)
    }

    /* An identifier as declared, without the backticks that let it reuse a keyword's spelling. */
    static func bareName(_ token: TokenSyntax) -> String {
        let text = token.text
        if text.count >= 2, text.hasPrefix("`"), text.hasSuffix("`") {
            return String(text.dropFirst().dropLast())
        }
        return text
    }
}
