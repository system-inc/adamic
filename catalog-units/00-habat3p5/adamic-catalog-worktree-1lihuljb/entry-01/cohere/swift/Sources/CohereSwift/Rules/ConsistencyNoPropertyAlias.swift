import SwiftSyntax

/*
 No property alias: a `let` inside a function that only renames a property onto a local of the same name.

     let tooltip = filterMode.tooltip
     IconButton(systemName: symbol, tooltip: tooltip)

 The reach `filterMode.tooltip` carries where the value came from and reads the same in any window; the local makes
 a reader scroll up to learn it. A naked local should mean this scope made the value. The repair writes the reach
 where the local is used and deletes the declaration. No fix: the rewrite is the author's to read.

 It ports `nexus/consistency-no-property-alias` (Kirk's "reach over alias"), shape for shape:
 - A `let` binding one name to a member read whose last name is that name: `let value = state.inner.value`. The
   chain is names and member reads at any depth (`self.model.review.takes.count`), through parentheses. TypeScript's
   `const` and `let` both map to Swift's `let`; a Swift `var` is a copy made to be changed, and is never judged.
 - A different name is not an alias (`let value = options.other`), in both languages.
 - A call anywhere in the chain is caching, not aliasing, as in TypeScript (`format().timeZone`,
   `self.plainTerminalView.getTerminal().cols`). A subscript is a call in Swift (an index can be computed by any
   code), so `items[index].name` is never judged either, where TypeScript judges `record[key].value`.
 - An optional chain (`window?.screen`) is exempt in both, wherever its `?` sits. Swift adds the force unwrap
   (`a!.b`), which traps where TypeScript's `!` is only a type assertion, and casts (`(a as T).b`), which convert.
 - File scope is out of scope in both (`main.swift`'s top-level code is TypeScript's module scope), and so is a
   type's own member (`let shared = Store.shared` beside the methods).
 - TypeScript's one hatch, a local read in a React hook's dependency array, has no Swift idiom and nothing to map to.

 Where Swift needs a different exactness, and why. In Swift any property may be computed: `clock.now` is a new
 instant on every read, `EffectLibrary.shared.generation` takes a lock, `ProcessInfo.processInfo.environment` builds
 a dictionary, and a class's stored property can change under any call. Neither the syntax nor the index can tell a
 stored property from a computed one (measured: sourcekitd records a getter call for a stored `let`, a stored `var`
 and a computed `var` alike), so the TypeScript rule's own exemption, "a chain that computes is caching", cannot be
 decided per property. What can be decided is when the property's nature cannot matter: the local is read exactly
 once, and nothing that could run code happens between the declaration and that read. Then the reach is evaluated
 once either way, at a moment nothing could have changed, and the rewrite changes nothing the program does. That is
 the whole of what is flagged. Each item below is a place where TypeScript reports and this rule is silent on purpose:
 - Read more than once. The local may be a snapshot: of a clock (macOS `ProviderProxyRequestTiming`: `let now =
   clock.now`, stored into three fields), of state another thread writes (Presence `OnsetStrip`: `let history =
   Ear.shared.history`, read field by field), of a counter compared and then stored (Presence `EffectRenderer`: `let
   generation = EffectLibrary.shared.generation`), or of a computed value worth computing once (`let angle =
   short.angle`, a quaternion's arc tangent). Each reach would read again, and could disagree or cost.
 - Read after a suspension point. Kirk's ruling: a `let x = object.x` taken before an `await` and used after it is
   the fix `concurrency-no-lost-update` asks for, a deliberate snapshot, so it stays silent. An `await` or a
   `for await` anywhere from the declaration to the end of the read's statement silences the line, so `library.removal
   = await perform()` (Presence `BodyProperties`), which writes through the local after the suspension, is silent too.
 - Read after code that could change it. Any call (a method, a function, an initializer, a subscript, a macro), any
   assignment or compound assignment, and any `&` argument between the declaration and the read, wherever it points:
   a class's property can change under a call to anything. A take-and-clear (Presence `ScreenRecorder`: `let conform =
   self.conform; self.conform = nil`) is this shape. When the read is the object written through (`layer.contents =
   make()`), the rest of its statement counts too, since the write lands after the right side runs.
 - Read inside a closure, a nested function, a `defer` or an `async let`, or named in a capture list. The closure
   runs on its own schedule, and capturing the value rather than the object is often the point: it keeps `self` out
   of a `Task.detached`, it crosses into a `@Sendable` closure, it reads SwiftUI state where the body observes it
   (Presence `DesignerBodies`: "Read here rather than in each tile, because a lazy grid builds its cells outside this
   body's observation").
 - Read inside a loop the declaration is outside of: the reach would run once per iteration. Hoisting a class's
   array before a hot loop is documented practice here (Presence EarDSP `Texture.frame`: "each read of a class's
   array property is an exclusivity check").
 - Never read. The compiler already warns, and there is no read to write the reach at.
 - A redeclaration of the chain's first name between the declaration and the read, after which the reach would name
   another object.
 - `if let`, `guard let` and `while let` are never judged: they narrow an optional, the Swift analogue of TypeScript's
   narrowing, and the reach cannot carry the unwrap. A shorthand `guard let conform` after a declaration counts as a
   read of it.
 - A type annotation (`let width: CGFloat = view.width`) may convert, so the line is not judged.

 Judged like any other chain: a value copy of a struct's property and a read through `self.` inside a type. With one
 read and nothing running between, there is no change for the copy to have been kept from, and `self.x` is the
 reach the message names.

 Taken as safe, as the TypeScript rule takes it: a property read, an operator and a string interpolation do not
 change the property being aliased, so they may sit between the declaration and the read.

 Known misses, every one a finding not made and never one invented:
 - Every alias read more than once, which is most of what TypeScript reports. Measured on 2026-10-03: of the 517
   alias-shaped `let`s inside functions in Presence's package, 420 are read more than once; of macOS's 16, 14. Many
   are true aliases (macOS `SessionDetail`: `let username = position.username`, read three times with nothing
   between), but nothing in the tree tells them from a snapshot.
 - A single read after harmless code (a `print`, a pure function, a value type's initializer), since a call is a
   call, and a single read inside a closure that is not escaping (`map`), which the tree cannot tell from one that is.
 */
public struct ConsistencyNoPropertyAlias: FileRule {
    public let name = "cohere-swift/consistency-no-property-alias"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Every flagged declaration is a `let`. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("let")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        Self.candidates(in: Syntax(file.tree)).compactMap { candidate in
            guard candidate.silences.isEmpty else { return nil }
            var finding = file.finding(
                at: candidate.binding,
                rule: name,
                message: RuleMessages.ConsistencyNoPropertyAlias.noPropertyAlias(
                    name: candidate.local.text,
                    reach: candidate.reach.trimmedDescription,
                ),
            )
            /* The binding ends at its value, before the comma that joins it to the next binding of the same `let`. */
            let end = candidate.reach.endLocation(converter: file.locations)
            finding.endLine = end.line
            finding.endColumn = end.column
            return finding
        }
    }

    /* Why an alias-shaped declaration is not reported. */
    enum Silence: String, Sendable {
        /* A reference to the name that the scope reading cannot place, such as one behind a `#if` that declares it again. */
        case unreadable
        /* Never read: the compiler already says so, and there is no read to write the reach at. */
        case unread
        /* Read more than once: possibly a snapshot. */
        case repeated
        /* Read inside a closure, a nested function, a `defer` or an `async let`, or captured by name. */
        case captured
        /* Read inside a loop the declaration is outside of. */
        case hoisted
        /* An `await` between the declaration and the read: Kirk's lost-update snapshot. */
        case suspension
        /* Code that could change the property runs between the declaration and the read. */
        case runs
        /* The chain's first name is declared again before the read. */
        case shadowed
    }

    /* A same-named `let` of a plain member chain inside a function, with every reason it is not reported. */
    struct Candidate {
        let binding: PatternBindingSyntax
        let local: TokenSyntax
        let reach: MemberAccessExprSyntax
        let silences: [Silence]
    }

    static func candidates(in tree: Syntax) -> [Candidate] {
        let collector = Collector(viewMode: .sourceAccurate)
        collector.walk(tree)
        return collector.declarations.flatMap { judge($0) }
    }

    final class Collector: SyntaxVisitor {
        private(set) var declarations: [VariableDeclSyntax] = []

        override func visit(_ node: VariableDeclSyntax) -> SyntaxVisitorContinueKind {
            declarations.append(node)
            return .visitChildren
        }
    }

    /* The alias-shaped bindings of one `let` statement inside a function. */
    static func judge(_ variable: VariableDeclSyntax) -> [Candidate] {
        guard variable.bindingSpecifier.tokenKind == .keyword(.let), variable.attributes.isEmpty,
            variable.modifiers.isEmpty,
            let item = variable.parent?.as(CodeBlockItemSyntax.self),
            let list = item.parent?.as(CodeBlockItemListSyntax.self), isInsideFunction(list)
        else { return [] }
        var candidates: [Candidate] = []
        for binding in variable.bindings {
            guard let identifier = binding.pattern.as(IdentifierPatternSyntax.self), binding.typeAnnotation == nil,
                binding.accessorBlock == nil,
                let reach = binding.initializer?.value.as(MemberAccessExprSyntax.self), let base = reach.base,
                reach.declName.argumentNames == nil,
                case .identifier = reach.declName.baseName.tokenKind,
                reach.declName.baseName.text == identifier.identifier.text,
                let root = rootName(of: base)
            else { continue }
            candidates.append(
                classify(binding: binding, local: identifier.identifier, reach: reach, root: root, list: list)
            )
        }
        return candidates
    }

    /* Whether a block's statements run inside a function, closure or accessor, rather than at file scope or among a type's members. */
    static func isInsideFunction(_ list: CodeBlockItemListSyntax) -> Bool {
        var current = list.parent
        while let node = current {
            if node.is(SourceFileSyntax.self) || ConcurrencyNoCheckThenWrite.isTypeDeclaration(node) {
                return false
            }
            if ConcurrencyNoCheckThenWrite.isFunctionLike(node) {
                return true
            }
            current = node.parent
        }
        return false
    }

    /*
     The first name of a chain of names and member reads, through parentheses, or nil when anything in it calls,
     subscripts, unwraps, casts, awaits or starts from an implicit member.
     */
    static func rootName(of expression: ExprSyntax) -> TokenSyntax? {
        if let reference = expression.as(DeclReferenceExprSyntax.self) {
            return reference.argumentNames == nil ? reference.baseName : nil
        }
        if let member = expression.as(MemberAccessExprSyntax.self), let base = member.base,
            member.declName.argumentNames == nil
        {
            return rootName(of: base)
        }
        if let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1, let only = tuple.elements.first,
            only.label == nil
        {
            return rootName(of: only.expression)
        }
        return nil
    }

    /* Every place a name is read: references (never a member's own name or a key path's), shorthand optional bindings, and capture list entries. */
    final class Reads: SyntaxVisitor {
        let name: String
        private(set) var references: [DeclReferenceExprSyntax] = []
        private(set) var shorthands: [Syntax] = []

        init(name: String) {
            self.name = name
            super.init(viewMode: .sourceAccurate)
        }

        override func visit(_ node: DeclReferenceExprSyntax) -> SyntaxVisitorContinueKind {
            if node.baseName.text == name, node.argumentNames == nil,
                node.parent?.as(MemberAccessExprSyntax.self)?.declName.id != node.id,
                node.parent?.is(KeyPathPropertyComponentSyntax.self) != true
            {
                references.append(node)
            }
            return .visitChildren
        }

        /* `if let name` with no value reads the outer `name`. */
        override func visit(_ node: OptionalBindingConditionSyntax) -> SyntaxVisitorContinueKind {
            if node.initializer == nil, node.pattern.as(IdentifierPatternSyntax.self)?.identifier.text == name {
                shorthands.append(Syntax(node))
            }
            return .visitChildren
        }

        /* `[name]` captures the outer `name`. */
        override func visit(_ node: ClosureCaptureSyntax) -> SyntaxVisitorContinueKind {
            if node.initializer == nil, node.name.text == name {
                shorthands.append(Syntax(node))
            }
            return .visitChildren
        }
    }

    static func classify(
        binding: PatternBindingSyntax,
        local: TokenSyntax,
        reach: MemberAccessExprSyntax,
        root: TokenSyntax,
        list: CodeBlockItemListSyntax,
    ) -> Candidate {
        var silences: [Silence] = []
        let reads = Reads(name: local.text)
        reads.walk(list)
        /* A shorthand naming the local after it is taken as a read of it: the scope reading resolves references, and a wrong guess here can only add silence. */
        var uses: [Syntax] = reads.shorthands.filter { $0.position >= binding.endPosition }
        for reference in reads.references where reference.position >= binding.endPosition {
            switch ConcurrencyNoLostUpdate.resolve(reference) {
                case .local(let declared):
                    if declared.id == local.id {
                        uses.append(Syntax(reference))
                    }
                case .shared:
                    break
                case .unknown:
                    silences.append(.unreadable)
            }
        }
        if uses.isEmpty {
            silences.append(.unread)
        }
        if uses.count > 1 {
            silences.append(.repeated)
        }
        for use in uses {
            if use.is(ClosureCaptureSyntax.self) || isNested(use, below: list) {
                silences.append(.captured)
            }
            if isRepeated(use, below: list) {
                silences.append(.hoisted)
            }
        }
        if let use = uses.min(by: { $0.position < $1.position }) {
            /*
             The read's own statement counts too when the value is consumed after the rest of it: an `await` anywhere in
             it (`library.removal = await perform()` writes through the local after the suspension), and anything that
             runs after a write target is read (the write lands when the right side is done).
             */
            let statementEnd = statement(of: use)?.endPosition ?? use.endPosition
            var between = Between(
                span: binding.endPosition..<use.position,
                consumedBy: isWriteTarget(use) ? statementEnd : use.position,
                suspensionEnd: statementEnd,
                use: use,
                root: root.text,
            )
            between.visit(Syntax(list))
            if between.suspends {
                silences.append(.suspension)
            }
            if between.runs {
                silences.append(.runs)
            }
            if between.shadowed {
                silences.append(.shadowed)
            }
        }
        var unique: [Silence] = []
        for silence in silences where !unique.contains(silence) {
            unique.append(silence)
        }
        return Candidate(binding: binding, local: local, reach: reach, silences: unique)
    }

    /* Whether a read runs on another schedule than the declaration: inside a closure, a nested function, a `defer` or an `async let`. */
    static func isNested(_ use: Syntax, below list: CodeBlockItemListSyntax) -> Bool {
        var current = use.parent
        while let node = current, node.id != list.id {
            if ConcurrencyNoCheckThenWrite.isFunctionLike(node) || node.is(DeferStmtSyntax.self) {
                return true
            }
            if let variable = node.as(VariableDeclSyntax.self),
                variable.modifiers.contains(where: { $0.name.tokenKind == .keyword(.async) })
            {
                return true
            }
            current = node.parent
        }
        return false
    }

    /* Whether a read sits in the repeated part of a loop the declaration is outside of: a `for` loop's body or `where`, or anywhere in a `while` or `repeat`. */
    static func isRepeated(_ use: Syntax, below list: CodeBlockItemListSyntax) -> Bool {
        var child = use
        var current = use.parent
        while let node = current, node.id != list.id {
            if let forStatement = node.as(ForStmtSyntax.self), forStatement.sequence.id != child.id {
                return true
            }
            if node.is(WhileStmtSyntax.self) || node.is(RepeatStmtSyntax.self) {
                return true
            }
            child = node
            current = node.parent
        }
        return false
    }

    /* The statement a read belongs to: its nearest enclosing block item. */
    static func statement(of use: Syntax) -> CodeBlockItemSyntax? {
        var current = use.parent
        while let node = current {
            if let item = node.as(CodeBlockItemSyntax.self) {
                return item
            }
            current = node.parent
        }
        return nil
    }

    /* Whether a read is the object written through: on the left of an assignment, or under an `&`. */
    static func isWriteTarget(_ use: Syntax) -> Bool {
        var current = use.parent
        while let node = current, !node.is(CodeBlockItemSyntax.self), !ConcurrencyNoCheckThenWrite.isFunctionLike(node)
        {
            if node.is(InOutExprSyntax.self) {
                return true
            }
            if let sequence = node.as(SequenceExprSyntax.self) {
                let elements = Array(sequence.elements)
                if let operatorIndex = elements.firstIndex(where: { isAssignment($0) }),
                    elements[operatorIndex].position > use.position
                {
                    return true
                }
            }
            current = node.parent
        }
        return false
    }

    /*
     What runs between the end of a declaration and its one read, in source order. Nodes that contain the read are
     skipped, since a call whose argument is the read runs after it; their earlier parts, a receiver's call or an
     argument before the read, are visited like anything else. Suspensions are looked for up to the end of the read's
     statement, and so is other code when the read is a write target.
     */
    struct Between {
        let span: Range<AbsolutePosition>
        let consumedBy: AbsolutePosition
        let suspensionEnd: AbsolutePosition
        let use: Syntax
        let root: String
        var suspends = false
        var runs = false
        var shadowed = false

        mutating func visit(_ node: Syntax) {
            guard node.endPosition > span.lowerBound, node.position < max(suspensionEnd, consumedBy) else { return }
            let containsUse = node.position <= use.position && use.endPosition <= node.endPosition
            if !containsUse, node.position >= span.lowerBound {
                inspect(node)
            }
            for child in node.children(viewMode: .sourceAccurate) {
                visit(child)
            }
        }

        mutating func inspect(_ node: Syntax) {
            if node.position < suspensionEnd,
                node.is(AwaitExprSyntax.self) || node.as(ForStmtSyntax.self)?.awaitKeyword != nil
            {
                suspends = true
            }
            guard node.position < span.upperBound || node.position >= use.endPosition && node.position < consumedBy
            else { return }
            if node.is(FunctionCallExprSyntax.self) || node.is(SubscriptCallExprSyntax.self)
                || node.is(MacroExpansionExprSyntax.self) || node.is(InOutExprSyntax.self)
            {
                runs = true
            }
            if let sequence = node.as(SequenceExprSyntax.self),
                sequence.elements.contains(where: { ConsistencyNoPropertyAlias.isAssignment($0) })
            {
                runs = true
            }
            if node.as(IdentifierPatternSyntax.self)?.identifier.text == root {
                shadowed = true
            }
        }
    }

    static let comparisons: Set<String> = ["==", "!=", "<=", ">=", "===", "!=="]

    /* `=`, or an operator that ends in `=` and is not a comparison: `+=`, `<<=`, `&+=`. */
    static func isAssignment(_ element: ExprSyntax) -> Bool {
        if element.is(AssignmentExprSyntax.self) {
            return true
        }
        guard let operation = element.as(BinaryOperatorExprSyntax.self)?.operator.text else { return false }
        return operation.hasSuffix("=") && !comparisons.contains(operation)
    }
}
