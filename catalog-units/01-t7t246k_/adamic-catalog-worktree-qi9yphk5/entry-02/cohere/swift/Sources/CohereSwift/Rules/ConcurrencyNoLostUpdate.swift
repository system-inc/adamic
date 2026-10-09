import SwiftOperators
import SwiftSyntax

/*
 No read, await, write: a write whose new value is computed from shared state as it was before an `await`.

     let count = self.count
     await save()
     self.count = count + 1

 While the function is suspended, other work runs on the same actor: another call into this actor, another task on
 the main actor, a callback. Swift's actor isolation stops two of them from touching `count` at the same instant,
 not from taking turns at every `await`, so a change made during the suspension is overwritten by a value computed
 before it. The repair reads after the suspension (`await save(); self.count += 1`), or keeps the whole read,
 compute and write on one side of it. No fix: which repair is right is the author's call.

 It ports `nexus/concurrency-no-lost-update`, condition for condition, each one required:
 - The function can suspend: it owns an `await` or a `for await` loop, never one inside a closure or a nested
   function, which runs on its own schedule. A closure is a function of its own here, as an arrow function is in
   TypeScript: `Task { ... }` is judged on its own awaits.
 - The write is a plain `=`. Compound assignments (`+=` and the rest) are never judged, and this is where Swift
   differs from JavaScript. Measured with swiftc 6.4 on 2026-10-03, an actor whose property another task bumps
   during the suspension: `total += await pause()` keeps the bump, because `+=` takes its left side `inout` and the
   access begins when the operator runs, after the right side has been evaluated and the await is over. So does
   `table["k", default: 0] += await pause()`, the same on a `@MainActor` class and on a global. TypeScript's
   `x += await y` reads `x` first and loses the update; Swift's does not. The same run showed what does lose it:
   `total = total + (await pause())`, `total = max(total, await pause())`, `items = items + [await pause()]`,
   `items = items.appending(await pause())` (a method's receiver is read before its arguments), `table["k"] =
   (table["k"] ?? 0) + (await pause())`, and a `let` read before the await and used after it.
 - The target is a member path: `self.count`, a bare `count` that no local binds (a property through implicit
   `self`, or a global, or a static member as `Type.count`), or a member of a parameter (`entry.retries`). `self.`
   and the bare spelling are the same path, since a bare name that is not a local resolves to a member of `self`
   before a global whenever `self` has one. Optional chaining and force unwraps (`self?.count`, `self!.count`) are
   transparent, and a subscript with one literal key (`self.table["main"]`) is a member, as TypeScript reads a
   quoted key. Names resolve by Swift's scope rules, read from the tree with the scope reading
   `concurrency-no-check-then-write` uses (`ConcurrencyNoCheckThenWrite.lookup`).
 - Someone else can write the target during the suspension. A member of `self`, a global and a static are shared:
   `self` is an actor, a class, or a struct whose property has a `nonmutating` setter, since nothing else compiles.
   The exception is a `mutating` or `consuming` method or an initializer (or a closure inside one): there `self` is
   held exclusively for the whole call, so nothing else can write its members while it is suspended, and a bare
   name there could be either a member or a global, so neither is judged. A member of a parameter that is not
   `inout` is shared for the same reason, since writing through it compiles only for a reference: the caller holds
   the object. An `inout` parameter is held exclusively, and an untyped closure parameter may be `inout`, so
   neither is judged.
 - The new value is computed from the target's prior value, read before a suspension that runs before the write on
   some path. The read is either in the written expression itself (`self.total = self.total + (await price())`),
   or reached through a `let` of this function whose initializer read it, through any chain of such `let`s.
   Expressions are evaluated in Swift's order along each path: left to right, a method's receiver before its
   arguments, `&&`, `||`, `??` and `?:` fork, and a suspension stales every read already taken on its path. Across
   statements, through the function's control flow: from the end of the `let` to the write, is there a path
   through an `await` (or a `for await` step) that does not declare the `let` again? Loops, `if`, `guard`,
   `switch`, `do`/`catch`, `defer`, labeled `break` and `continue`, `fallthrough`, `return`, `throw` and calls to
   `fatalError`, `preconditionFailure`, `exit` and `abort` are followed. A `try` reaches its `catch` before anything
   it covers ran, and also after, when what it covers is one call (which then threw from inside, past its own
   awaits); a `try` over several calls may have thrown from an early one, so only the first edge is drawn. An
   `await` whose expression holds the write runs after it, as in TypeScript.
 - The value is modified rather than copied, and computed here, which is the TypeScript rule's definition of a
   deliberate restore. Writing back exactly what was saved (`self.appearance = kept`) is a restore and never
   judged; any operation on the value (arithmetic, a call, a member read, a collection literal) makes it a
   modification. Two things are not: an awaited call's result, which is the callee's answer rather than this
   function's computation (`self.clip = await self.skeletons.clip(self.clip)`), and a write after this function
   itself wrote the target between the read and the write, which is the save-and-restore protocol (`let kept =
   self.appearance; for preset in presets { self.appearance = preset; await self.save() }; self.appearance =
   kept`). Stricter than TypeScript in the safe direction: a write to a path that contains the target counts as
   that write (`self.state = State()` replaces `self.state.count`), and so does a method called on the target or
   on a path containing it (`self.items.removeAll()`, which may be `mutating`) and passing it `inout`.

 The finding spans the whole assignment and its message names the target as written.

 What it accepts as safe, and why: a compound assignment (measured above), a value read after the suspension, a
 verbatim restore, a restore after this function's own replacement, an awaited call's result, a local of this or
 an enclosing function (see below), and any path it cannot resolve.

 Known misses, every one a finding not made and never one invented:
 - A local captured by a closure. TypeScript judges an outer variable written from a nested function; Swift 6
   rejects mutating a captured `var` in any closure that can run concurrently (`Task`, a task group's child), and
   a non-escaping closure (a task group's body, `forEach`) cannot run during the suspension, which leaves only a
   non-`Sendable` callback run on the same actor. The tree cannot tell an escaping closure from one that is not, so
   no local is judged, and neither is a member of a local (`let model = self.model; model.count = ...`).
 - An awaited operand that is not a call. `await` covers its whole expression and the syntax cannot say where in
   it the suspension falls, so reads inside it are taken as after it: `self.total = await self.total + pause()`
   loses the update (measured) and is not reported.
 - The TypeScript rule's store half (`getAccount` and `updateAccount` called with the same key across an await),
   which has no Swift idiom with that shape on the proving grounds; `UserDefaults` and dictionary subscripts with
   a variable key are not read.
 - A computed subscript key (`self.table[key] = ...`), as in TypeScript, since nothing here can show `key` held
   still.
 - A value copy of a path that contains the target (`let snapshot = self.state; await save(); self.state.count =
   snapshot.count + 1`), which is stale when `State` is a struct and live when it is a class: TypeScript reads it
   as an alias and stays silent, and so does this rule, since only types could tell the two apart.
 - `if let`, `guard let`, `var` and `async let` bindings as carriers of the old value, and tuple destructuring.
 - A read outside a closure written inside it (`let old = self.count; await save(); await MainActor.run {
   self.count = old + 1 }`), and a `let` of an enclosing function read in a `Task` that starts after it.
 - A function holding `#if`, or an expression the standard operators cannot fold (a custom operator), where the
   control flow or the writes are not all readable.
 - The value an `if` or `switch` expression chooses (`self.count = if flag { self.count + 1 } else { 0 }`); only
   whether one of its branches can suspend is read.
 - Top-level code in `main.swift`, and the implicit await at the end of an `async let`'s scope.

 The one assumption that could report wrongly: every `await` is taken to suspend, as TypeScript takes `await 5`.
 An `await` over nothing asynchronous never does, and the compiler warns about it ("no 'async' operations occur
 within 'await' expression"), so a finding there arrives beside that warning.
 */
public struct ConcurrencyNoLostUpdate: FileRule {
    public let name = "cohere-swift/concurrency-no-lost-update"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Every flagged function owns an `await`, and a `for await` is spelled with one too. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("await")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        /* Folded, so `a = b + c` is one assignment of one sum. Folding moves no token, so positions in the copy are positions in the file. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let collector = Collector(viewMode: .sourceAccurate)
        collector.walk(folded)
        var findings: [FindingRecord] = []
        for (node, statements) in collector.roots {
            let analysis = Analysis(root: node, statements: statements)
            for write in analysis.losingWrites() {
                findings.append(
                    file.finding(
                        at: write,
                        rule: name,
                        message: RuleMessages.ConcurrencyNoLostUpdate.lostUpdate(
                            suspendedRead: .variable(target: write.leftOperand.trimmedDescription)
                        ),
                    )
                )
            }
        }
        return findings
    }

    /* The statements of a function, an initializer, an accessor with a body, or a closure: the code one suspending call runs. */
    static func body(of node: Syntax) -> CodeBlockItemListSyntax? {
        if let function = node.as(FunctionDeclSyntax.self) {
            return function.body?.statements
        }
        if let initializer = node.as(InitializerDeclSyntax.self) {
            return initializer.body?.statements
        }
        if let accessor = node.as(AccessorDeclSyntax.self) {
            return accessor.body?.statements
        }
        if let closure = node.as(ClosureExprSyntax.self) {
            return closure.statements
        }
        return nil
    }

    /* Every function, initializer, accessor and closure with a body, each judged on its own. */
    final class Collector: SyntaxAnyVisitor {
        private(set) var roots: [(Syntax, CodeBlockItemListSyntax)] = []

        override func visitAny(_ node: Syntax) -> SyntaxVisitorContinueKind {
            if let statements = ConcurrencyNoLostUpdate.body(of: node) {
                roots.append((node, statements))
            }
            return .visitChildren
        }
    }

    // MARK: Paths

    /*
     Where a path starts: `self` or a name no local binds (one place, however it is spelled), or one local binding,
     marked when it is a parameter that is not `inout`, whose members live in an object the caller holds.
     */
    enum Origin: Hashable {
        case shared
        case binding(SyntaxIdentifier, callerHeld: Bool)
    }

    /* A member path: an origin and the members read through it, a literal subscript key written as `["key"]`. */
    struct Path: Hashable {
        var origin: Origin
        var names: [String]

        /* Whether this path is the other one or contains it: a write here replaces what the other names. */
        func covers(_ other: Path) -> Bool {
            origin == other.origin && names.count <= other.names.count
                && Array(other.names.prefix(names.count)) == names
        }
    }

    /* What a bare name resolves to: a local binding, a member or global, or something the rule does not read. */
    enum Resolution {
        case local(ConcurrencyNoCheckThenWrite.Declaration)
        case shared
        case unknown
    }

    /* A bare name resolved by Swift's scope rules, with the scope reading that `concurrency-no-check-then-write` uses. */
    static func resolve(_ reference: DeclReferenceExprSyntax) -> Resolution {
        guard reference.argumentNames == nil, case .identifier = reference.baseName.tokenKind else { return .unknown }
        let name = reference.baseName.text
        var child = Syntax(reference)
        while let node = child.parent {
            /* Past every function: a member of the enclosing type, or a global. */
            if node.is(MemberBlockItemListSyntax.self) {
                return .shared
            }
            if let list = node.as(CodeBlockItemListSyntax.self), list.parent?.is(SourceFileSyntax.self) == true {
                return .shared
            }
            switch ConcurrencyNoCheckThenWrite.lookup(name, in: node, below: child) {
                case .declared(let declared)?:
                    return .local(declared)
                case .unknown?:
                    return .unknown
                case nil:
                    break
            }
            child = node
        }
        return .unknown
    }

    /* An expression read as a member path, through parentheses, `?` and `!`. */
    static func path(of expression: ExprSyntax) -> Path? {
        let expression = transparent(expression)
        if let reference = expression.as(DeclReferenceExprSyntax.self) {
            if reference.baseName.tokenKind == .keyword(.self) {
                return Path(origin: .shared, names: [])
            }
            switch resolve(reference) {
                case .local(let declared):
                    let isParameter =
                        declared.node.parent?.is(FunctionParameterSyntax.self) == true
                        || declared.node.parent?.is(ClosureParameterSyntax.self) == true
                    return Path(
                        origin: .binding(declared.id, callerHeld: isParameter && !declared.isVariable),
                        names: [],
                    )
                case .shared:
                    return Path(origin: .shared, names: [reference.baseName.text])
                case .unknown:
                    return nil
            }
        }
        if let member = expression.as(MemberAccessExprSyntax.self) {
            guard let base = member.base, member.declName.argumentNames == nil, var path = path(of: base) else {
                return nil
            }
            path.names.append(member.declName.baseName.text)
            return path
        }
        if let subscriptCall = expression.as(SubscriptCallExprSyntax.self) {
            guard subscriptCall.trailingClosure == nil, subscriptCall.additionalTrailingClosures.isEmpty,
                subscriptCall.arguments.count == 1,
                let only = subscriptCall.arguments.first, only.label == nil, let key = literalKey(only.expression),
                var path = path(of: subscriptCall.calledExpression)
            else {
                return nil
            }
            path.names.append("[" + key + "]")
            return path
        }
        return nil
    }

    /* A string literal with no interpolation, or an integer literal: a key that names one entry every time. */
    static func literalKey(_ expression: ExprSyntax) -> String? {
        if let integer = expression.as(IntegerLiteralExprSyntax.self) {
            return integer.literal.text
        }
        if let string = expression.as(StringLiteralExprSyntax.self),
            string.segments.allSatisfy({ $0.is(StringSegmentSyntax.self) })
        {
            return string.trimmedDescription
        }
        return nil
    }

    /* Parentheses, `?` and `!` removed: none changes which value or which storage is meant. */
    static func transparent(_ expression: ExprSyntax) -> ExprSyntax {
        var current = ConcurrencyNoCheckThenWrite.withoutParentheses(expression)
        while true {
            if let chained = current.as(OptionalChainingExprSyntax.self) {
                current = ConcurrencyNoCheckThenWrite.withoutParentheses(chained.expression)
            }
            else if let unwrapped = current.as(ForceUnwrapExprSyntax.self) {
                current = ConcurrencyNoCheckThenWrite.withoutParentheses(unwrapped.expression)
            }
            else {
                return current
            }
        }
    }

    // MARK: Structure

    /* Visits a subtree in source order without entering closures, nested functions or local types. */
    static func forEachOwnNode(_ node: Syntax, _ visit: (Syntax) -> Void) {
        visit(node)
        for child in node.children(viewMode: .sourceAccurate)
        where !ConcurrencyNoCheckThenWrite.isFunctionLike(child)
            && !ConcurrencyNoCheckThenWrite.isTypeDeclaration(child)
        {
            forEachOwnNode(child, visit)
        }
    }

    /* Whether `self` is held exclusively by an enclosing `mutating` or `consuming` method, or an initializer, up to the type. */
    static func holdsSelfExclusively(_ root: Syntax) -> Bool {
        var current: Syntax? = root
        while let node = current, !ConcurrencyNoCheckThenWrite.isTypeDeclaration(node) {
            if node.is(InitializerDeclSyntax.self) {
                return true
            }
            let modifiers = node.as(FunctionDeclSyntax.self)?.modifiers ?? node.as(AccessorDeclSyntax.self)?.modifiers
            if let modifiers, modifiers.contains(where: { Self.exclusiveModifiers.contains($0.name.tokenKind) }) {
                return true
            }
            current = node.parent
        }
        return false
    }

    static let exclusiveModifiers: [TokenKind] = [.keyword(.mutating), .keyword(.consuming)]

    /* `=` alone: the write this rule judges. */
    static func isPlainAssignment(_ infix: InfixOperatorExprSyntax) -> Bool {
        infix.operator.is(AssignmentExprSyntax.self)
    }

    /* `=` or a compound assignment (`+=`, `<<=`): any write to its left side, for the save-and-restore check. */
    static func isAssignment(_ infix: InfixOperatorExprSyntax) -> Bool {
        if isPlainAssignment(infix) {
            return true
        }
        guard let operation = infix.operator.as(BinaryOperatorExprSyntax.self)?.operator.text else { return false }
        return operation.hasSuffix("=") && !ConcurrencyNoCheckThenWrite.comparisons.contains(operation)
    }

    // MARK: Values

    /* What an evaluated value carries of the target's prior value: read before a suspension yet or not, and the value itself or something computed from it. */
    struct Dependencies: OptionSet, Hashable {
        let rawValue: UInt8
        static let freshVerbatim = Dependencies(rawValue: 1)
        static let freshModified = Dependencies(rawValue: 2)
        static let staleVerbatim = Dependencies(rawValue: 4)
        static let staleModified = Dependencies(rawValue: 8)

        /* What a suspension does to a value read before it. */
        var staled: Dependencies {
            var result = subtracting([.freshVerbatim, .freshModified])
            if contains(.freshVerbatim) {
                result.insert(.staleVerbatim)
            }
            if contains(.freshModified) {
                result.insert(.staleModified)
            }
            return result
        }

        /* What any operation on a value does to it. */
        var modified: Dependencies {
            var result = subtracting([.freshVerbatim, .staleVerbatim])
            if contains(.freshVerbatim) {
                result.insert(.freshModified)
            }
            if contains(.staleVerbatim) {
                result.insert(.staleModified)
            }
            return result
        }
    }

    /* One evaluation path through an expression. */
    struct Outcome: Hashable {
        var dependencies: Dependencies
        /* Whether a suspension has happened on this path, counting one before the expression. */
        var suspended: Bool
        /* Whether a suspension happened inside this expression, which stales what its earlier siblings read. */
        var during: Bool
    }

    /* Where a value is judged: at a write, or at the declaration of a `let` whose initializer carries the target. */
    enum Anchor: Hashable {
        case write(SyntaxIdentifier)
        case declaration(SyntaxIdentifier)
    }

    /* A `let` of the function, named by one identifier and given a value: the only binding that can carry a read across statements. */
    struct Local {
        let token: TokenSyntax
        let initializer: ExprSyntax
    }

    // MARK: The analysis of one function

    final class Analysis {
        let root: Syntax
        let statements: CodeBlockItemListSyntax
        var locals: [SyntaxIdentifier: Local] = [:]
        lazy var graph = FlowGraph(statements: statements, locals: Set(locals.keys))
        var localMemo: [LocalKey: Dependencies] = [:]
        var localInProgress: Set<LocalKey> = []
        var betweenMemo: [BetweenKey: Bool] = [:]

        struct LocalKey: Hashable {
            let local: SyntaxIdentifier
            let target: Path
        }

        struct BetweenKey: Hashable {
            let local: SyntaxIdentifier
            let anchor: Anchor
            let target: Path
        }

        init(root: Syntax, statements: CodeBlockItemListSyntax) {
            self.root = root
            self.statements = statements
        }

        /* Every plain `=` in this function that writes shared state from a stale read of it. */
        func losingWrites() -> [InfixOperatorExprSyntax] {
            var suspends = false
            var unreadable = false
            var assignments: [InfixOperatorExprSyntax] = []
            ConcurrencyNoLostUpdate.forEachOwnNode(Syntax(statements)) { node in
                if node.is(AwaitExprSyntax.self) || node.as(ForStmtSyntax.self)?.awaitKeyword != nil {
                    suspends = true
                }
                /* `#if` hides which branch is compiled, and an unfolded sequence hides its writes. */
                if node.is(IfConfigDeclSyntax.self) || node.is(SequenceExprSyntax.self) {
                    unreadable = true
                }
                if let infix = node.as(InfixOperatorExprSyntax.self), ConcurrencyNoLostUpdate.isPlainAssignment(infix) {
                    assignments.append(infix)
                }
                if let variable = node.as(VariableDeclSyntax.self),
                    variable.bindingSpecifier.tokenKind == .keyword(.let), variable.attributes.isEmpty,
                    !variable.modifiers.contains(where: { $0.name.tokenKind == .keyword(.async) })
                {
                    for binding in variable.bindings where binding.accessorBlock == nil {
                        if let identifier = binding.pattern.as(IdentifierPatternSyntax.self),
                            let value = binding.initializer?.value
                        {
                            locals[identifier.identifier.id] = Local(token: identifier.identifier, initializer: value)
                        }
                    }
                }
            }
            /* A function that never suspends cannot hold a stale read. */
            guard suspends, !unreadable else { return [] }
            let exclusive = ConcurrencyNoLostUpdate.holdsSelfExclusively(root)
            return assignments.filter { assignment in
                guard let target = ConcurrencyNoLostUpdate.path(of: assignment.leftOperand), !target.names.isEmpty,
                    isShared(target, exclusive: exclusive)
                else { return false }
                let evaluator = Evaluator(
                    analysis: self,
                    target: target,
                    anchor: .write(assignment.id),
                    anchorNode: Syntax(assignment),
                )
                return evaluator.evaluate(assignment.rightOperand, suspended: false).contains {
                    $0.dependencies.contains(.staleModified)
                }
            }
        }

        /* Whether something other than this line can write the target while the function is suspended. */
        func isShared(_ target: Path, exclusive: Bool) -> Bool {
            switch target.origin {
                case .shared:
                    return !exclusive
                case .binding(_, let callerHeld):
                    return callerHeld
            }
        }

        /* What a `let` carries of the target at the end of its declaration. */
        func dependencies(of local: Local, target: Path) -> Dependencies {
            let key = LocalKey(local: local.token.id, target: target)
            if let memoized = localMemo[key] {
                return memoized
            }
            if localInProgress.contains(key) {
                return []
            }
            localInProgress.insert(key)
            let evaluator = Evaluator(
                analysis: self,
                target: target,
                anchor: .declaration(local.token.id),
                anchorNode: Syntax(local.token),
            )
            var dependencies: Dependencies = []
            for outcome in evaluator.evaluate(local.initializer, suspended: false) {
                dependencies.formUnion(outcome.dependencies)
            }
            localInProgress.remove(key)
            localMemo[key] = dependencies
            return dependencies
        }

        /*
         Whether some path runs from the end of a `let` through a suspension to the anchor, without declaring the
         `let` again and without this function writing the target in between.

         That last clause is the save-and-restore protocol: the function itself overwrote the target after reading
         it, so the later write puts back the state from before its own change rather than building on a value
         someone else may have changed.
         */
        func suspensionBetween(_ local: SyntaxIdentifier, anchor: Anchor, anchorNode: Syntax, target: Path) -> Bool {
            let key = BetweenKey(local: local, anchor: anchor, target: target)
            if let memoized = betweenMemo[key] {
                return memoized
            }
            let graph = self.graph
            struct Position {
                let block: Int
                let start: Int
                let suspended: Bool
            }
            var queue: [Position] = []
            for (index, block) in graph.blocks.enumerated() where graph.reachable[index] {
                for (eventIndex, event) in block.events.enumerated() {
                    if case .declarationEnd(let id) = event, id == local {
                        queue.append(Position(block: index, start: eventIndex + 1, suspended: false))
                    }
                }
            }
            var visited: Set<[Int]> = []
            var answer = false
            search: while !queue.isEmpty {
                let position = queue.removeFirst()
                var suspended = position.suspended
                var stopped = false
                let events = graph.blocks[position.block].events
                for event in events[position.start...] {
                    switch event {
                        case .suspend(let node):
                            /* A suspension whose expression contains the anchor runs after it. */
                            if !(node.map {
                                $0.position <= anchorNode.position && anchorNode.endPosition <= $0.endPosition
                            } ?? false) {
                                suspended = true
                            }
                        case .declarationEnd(let id):
                            /* Declaring the `let` again binds a fresh one, so this path says nothing about the value the anchor reads. */
                            if id == local {
                                stopped = true
                            }
                        case .declarationStart(let id):
                            if anchor == .declaration(id), suspended {
                                answer = true
                                break search
                            }
                        case .write(let node):
                            if anchor == .write(node.id) {
                                if suspended {
                                    answer = true
                                    break search
                                }
                            }
                            else if writes(node, target: target) {
                                stopped = true
                            }
                    }
                    if stopped {
                        break
                    }
                }
                if stopped {
                    continue
                }
                for successor in graph.blocks[position.block].successors where graph.reachable[successor] {
                    let key = [successor, suspended ? 1 : 0]
                    if visited.insert(key).inserted {
                        queue.append(Position(block: successor, start: 0, suspended: suspended))
                    }
                }
            }
            betweenMemo[key] = answer
            return answer
        }

        /* Whether a write replaces the target: an assignment to it or to a path containing it, a method called on such a path, or passing it `inout`. */
        func writes(_ node: Syntax, target: Path) -> Bool {
            let written: Path?
            if let infix = node.as(InfixOperatorExprSyntax.self) {
                written = ConcurrencyNoLostUpdate.path(of: infix.leftOperand)
            }
            else if let call = node.as(FunctionCallExprSyntax.self),
                let member = call.calledExpression.as(MemberAccessExprSyntax.self), let base = member.base
            {
                written = ConcurrencyNoLostUpdate.path(of: base)
            }
            else if let passed = node.as(InOutExprSyntax.self) {
                written = ConcurrencyNoLostUpdate.path(of: passed.expression)
            }
            else {
                written = nil
            }
            /* An object alone (`self.save()`, `model.wear(body)`) is the receiver of every method called on it, not a write of every member, as in TypeScript. */
            guard let written, !written.names.isEmpty else { return false }
            return written.covers(target)
        }
    }

    // MARK: Evaluation order

    /* Walks one expression in Swift's evaluation order for one target, path by path. */
    struct Evaluator {
        let analysis: Analysis
        let target: Path
        let anchor: Anchor
        let anchorNode: Syntax

        func leaf(_ dependencies: Dependencies, _ suspended: Bool) -> [Outcome] {
            [Outcome(dependencies: dependencies, suspended: suspended, during: false)]
        }

        func unique(_ outcomes: [Outcome]) -> [Outcome] {
            var seen: Set<Outcome> = []
            return outcomes.filter { seen.insert($0).inserted }
        }

        func modifiedAll(_ outcomes: [Outcome]) -> [Outcome] {
            unique(
                outcomes.map {
                    Outcome(dependencies: $0.dependencies.modified, suspended: $0.suspended, during: $0.during)
                }
            )
        }

        /* Expressions evaluated left to right with their values unioned. A suspension in a later one stales what earlier ones read. */
        func sequence(_ expressions: [ExprSyntax], suspended: Bool) -> [Outcome] {
            var states = leaf([], suspended)
            for expression in expressions {
                var next: [Outcome] = []
                for state in states {
                    for result in evaluate(expression, suspended: state.suspended) {
                        let carried = result.during ? state.dependencies.staled : state.dependencies
                        next.append(
                            Outcome(
                                dependencies: carried.union(result.dependencies),
                                suspended: result.suspended,
                                during: state.during || result.during,
                            )
                        )
                    }
                }
                states = unique(next)
            }
            return states
        }

        /* The second expression evaluated after each outcome of the first, keeping only the second's value. */
        func then(_ first: [Outcome], _ second: ExprSyntax) -> [Outcome] {
            var outcomes: [Outcome] = []
            for state in first {
                for var result in evaluate(second, suspended: state.suspended) {
                    result.during = result.during || state.during
                    outcomes.append(result)
                }
            }
            return outcomes
        }

        /* The expressions directly inside a node, in source order, which is Swift's evaluation order, never entering a closure or a declaration. */
        static func childExpressions(of node: Syntax) -> [ExprSyntax] {
            var found: [ExprSyntax] = []
            for child in node.children(viewMode: .sourceAccurate) {
                if let expression = child.as(ExprSyntax.self) {
                    found.append(expression)
                }
                else if !ConcurrencyNoCheckThenWrite.isFunctionLike(child),
                    !ConcurrencyNoCheckThenWrite.isTypeDeclaration(child), !child.is(CodeBlockSyntax.self)
                {
                    found += childExpressions(of: child)
                }
            }
            return found
        }

        /* Every path through an expression, given whether a suspension already happened on the path. */
        func evaluate(_ expression: ExprSyntax, suspended: Bool) -> [Outcome] {
            if expression.is(ClosureExprSyntax.self) {
                /* A closure's body runs later, if at all; its value carries nothing read now. */
                return leaf([], suspended)
            }
            if let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                return evaluate(only.expression, suspended: suspended)
            }
            if let tryExpression = expression.as(TryExprSyntax.self) {
                return evaluate(tryExpression.expression, suspended: suspended)
            }
            if let chained = expression.as(OptionalChainingExprSyntax.self) {
                return evaluate(chained.expression, suspended: suspended)
            }
            if let unwrapped = expression.as(ForceUnwrapExprSyntax.self) {
                return evaluate(unwrapped.expression, suspended: suspended)
            }
            if let cast = expression.as(AsExprSyntax.self) {
                /* A cast names the same value, so a saved value cast back is still a restore. */
                return evaluate(cast.expression, suspended: suspended)
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self) {
                if let path = ConcurrencyNoLostUpdate.path(of: expression) {
                    let matched = match(path)
                    if !matched.isEmpty {
                        return leaf(matched, suspended)
                    }
                }
                return leaf(identifier(reference, suspended: suspended), suspended)
            }
            if expression.is(MemberAccessExprSyntax.self) || expression.is(SubscriptCallExprSyntax.self),
                let path = ConcurrencyNoLostUpdate.path(of: expression)
            {
                let matched = match(path)
                if !matched.isEmpty {
                    return leaf(matched, suspended)
                }
                /* A path that is not the target can still be a member read of a `let` carrying it. Nothing in a member path can suspend. */
                var rootExpression = ConcurrencyNoLostUpdate.transparent(expression)
                while true {
                    if let member = rootExpression.as(MemberAccessExprSyntax.self), let base = member.base {
                        rootExpression = ConcurrencyNoLostUpdate.transparent(base)
                    }
                    else if let subscriptCall = rootExpression.as(SubscriptCallExprSyntax.self) {
                        rootExpression = ConcurrencyNoLostUpdate.transparent(subscriptCall.calledExpression)
                    }
                    else {
                        break
                    }
                }
                guard let rootReference = rootExpression.as(DeclReferenceExprSyntax.self) else {
                    return leaf([], suspended)
                }
                return leaf(identifier(rootReference, suspended: suspended).modified, suspended)
            }
            if let awaitExpression = expression.as(AwaitExprSyntax.self) {
                /*
                 An awaited call's value is the callee's answer, produced across the suspension, and not a computation
                 made here from what this function handed it. Reads inside the operand are not staled: the syntax
                 cannot say where in the operand the suspension falls, so they are taken as after it.
                 */
                var operand = ConcurrencyNoCheckThenWrite.withoutParentheses(awaitExpression.expression)
                if let tryExpression = operand.as(TryExprSyntax.self) {
                    operand = ConcurrencyNoCheckThenWrite.withoutParentheses(tryExpression.expression)
                }
                let producedByCallee =
                    operand.is(FunctionCallExprSyntax.self) || operand.is(SubscriptCallExprSyntax.self)
                    || operand.is(MacroExpansionExprSyntax.self)
                return unique(
                    evaluate(awaitExpression.expression, suspended: suspended).map { outcome in
                        Outcome(
                            dependencies: producedByCallee ? [] : outcome.dependencies,
                            suspended: true,
                            during: true,
                        )
                    }
                )
            }
            if let infix = expression.as(InfixOperatorExprSyntax.self) {
                if ConcurrencyNoLostUpdate.isAssignment(infix) {
                    /* An assignment's value is `()`. */
                    return unique(
                        sequence([infix.leftOperand, infix.rightOperand], suspended: suspended).map {
                            Outcome(dependencies: [], suspended: $0.suspended, during: $0.during)
                        }
                    )
                }
                if let operation = infix.operator.as(BinaryOperatorExprSyntax.self)?.operator.text,
                    ["&&", "||", "??"].contains(operation)
                {
                    let left = evaluate(infix.leftOperand, suspended: suspended)
                    return unique(left + then(left, infix.rightOperand))
                }
                return modifiedAll(sequence([infix.leftOperand, infix.rightOperand], suspended: suspended))
            }
            if let ternary = expression.as(TernaryExprSyntax.self) {
                let condition = evaluate(ternary.condition, suspended: suspended)
                return unique(then(condition, ternary.thenExpression) + then(condition, ternary.elseExpression))
            }
            if expression.is(IfExprSyntax.self) || expression.is(SwitchExprSyntax.self) {
                /* Branches of statements: their values are not read, only whether one of them can suspend. */
                var awaits = false
                ConcurrencyNoLostUpdate.forEachOwnNode(Syntax(expression)) { node in
                    if node.is(AwaitExprSyntax.self) || node.as(ForStmtSyntax.self)?.awaitKeyword != nil {
                        awaits = true
                    }
                }
                return [Outcome(dependencies: [], suspended: suspended || awaits, during: awaits)]
            }
            if let call = expression.as(FunctionCallExprSyntax.self) {
                /* A call reads its receiver, then its arguments. A bare callee is a function's name, not a read of a property spelled alike. */
                var parts: [ExprSyntax] = []
                if let member = call.calledExpression.as(MemberAccessExprSyntax.self) {
                    if let base = member.base {
                        parts.append(base)
                    }
                }
                else if !call.calledExpression.is(DeclReferenceExprSyntax.self) {
                    parts.append(call.calledExpression)
                }
                parts += call.arguments.map(\.expression)
                return modifiedAll(sequence(parts, suspended: suspended))
            }
            return modifiedAll(sequence(Self.childExpressions(of: Syntax(expression)), suspended: suspended))
        }

        /* A member path read against the target: the target itself is a verbatim read, a longer path through it a read something is computed from. */
        func match(_ path: Path) -> Dependencies {
            guard target.covers(path) else { return [] }
            return path.names.count == target.names.count ? .freshVerbatim : .freshModified
        }

        /* A bare name read: a `let` of this function carrying the target, stale when a suspension lies between its declaration and the anchor. */
        func identifier(_ reference: DeclReferenceExprSyntax, suspended: Bool) -> Dependencies {
            guard case .local(let declared) = ConcurrencyNoLostUpdate.resolve(reference),
                let local = analysis.locals[declared.id], local.token.id != reference.baseName.id
            else {
                return []
            }
            let dependencies = analysis.dependencies(of: local, target: target)
            if dependencies.isEmpty {
                return []
            }
            if suspended
                || analysis.suspensionBetween(local.token.id, anchor: anchor, anchorNode: anchorNode, target: target)
            {
                return dependencies.staled
            }
            return dependencies
        }
    }

    // MARK: Control flow

    /* One event on a path through the function. */
    enum Event {
        /* An `await`, or a `for await` step (no node). */
        case suspend(Syntax?)
        case declarationStart(SyntaxIdentifier)
        case declarationEnd(SyntaxIdentifier)
        /* An assignment, a method call on a member path, or an `&` argument: the candidate writes, and the anchors. */
        case write(Syntax)
    }

    /* The function's control-flow graph: blocks of events and the edges between them. */
    final class FlowGraph {
        struct Block {
            var events: [Event] = []
            var successors: [Int] = []
        }

        /* A `break` or `continue` target, and how many defer scopes were open when it was entered. */
        struct Jump {
            let label: String?
            let isLoop: Bool
            let isSwitch: Bool
            let breakTarget: Int
            let continueTarget: Int?
            let depth: Int
        }

        private(set) var blocks: [Block] = [Block()]
        private(set) var reachable: [Bool] = []
        private let locals: Set<SyntaxIdentifier>
        private var current = 0
        private var scopes: [[CodeBlockSyntax]] = []
        private var jumps: [Jump] = []
        private var catches: [(target: Int, depth: Int)] = []
        private var fallthroughs: [Int?] = []
        private var pendingLabel: String?

        init(statements: CodeBlockItemListSyntax, locals: Set<SyntaxIdentifier>) {
            self.locals = locals
            walk(statements)
            reachable = Array(repeating: false, count: blocks.count)
            var queue = [0]
            reachable[0] = true
            while let block = queue.popLast() {
                for successor in blocks[block].successors where !reachable[successor] {
                    reachable[successor] = true
                    queue.append(successor)
                }
            }
        }

        private func newBlock() -> Int {
            blocks.append(Block())
            return blocks.count - 1
        }

        private func edge(_ from: Int, _ to: Int) {
            blocks[from].successors.append(to)
        }

        private func emit(_ event: Event) {
            blocks[current].events.append(event)
        }

        /* Leaves the current point for nowhere reachable: after a `return`, a `throw`, a jump. */
        private func cut() {
            current = newBlock()
        }

        /* An edge from here to a target (none for leaving the function), running the defers of every scope opened since depth on the way. */
        private func jump(to target: Int?, unwinding depth: Int) {
            let resume = current
            if scopes[depth...].contains(where: { !$0.isEmpty }) {
                let side = newBlock()
                edge(current, side)
                current = side
                for scope in scopes[depth...].reversed() {
                    for body in scope.reversed() {
                        walk(body.statements)
                    }
                }
            }
            if let target {
                edge(current, target)
            }
            /* What follows the jump starts a block of its own, so the jump's path does not run it. */
            let continuation = newBlock()
            edge(resume, continuation)
            current = continuation
        }

        /* A block of statements, its own scope for `defer`, whose defers run when it ends. */
        private func walk(_ statements: CodeBlockItemListSyntax) {
            scopes.append([])
            for entry in statements {
                switch entry.item {
                    case .decl(let declaration):
                        if let variable = declaration.as(VariableDeclSyntax.self) {
                            for binding in variable.bindings where binding.accessorBlock == nil {
                                let tracked = binding.pattern.as(IdentifierPatternSyntax.self).map(\.identifier.id)
                                    .flatMap { locals.contains($0) ? $0 : nil }
                                if let tracked {
                                    emit(.declarationStart(tracked))
                                }
                                if let value = binding.initializer?.value {
                                    expression(value)
                                }
                                if let tracked {
                                    emit(.declarationEnd(tracked))
                                }
                            }
                        }
                    case .stmt(let statement):
                        self.statement(statement)
                    case .expr(let value):
                        expression(value)
                        terminateIfNeverReturns(value)
                }
            }
            if let scope = scopes.popLast() {
                for body in scope.reversed() {
                    walk(body.statements)
                }
            }
        }

        /* `fatalError`, `preconditionFailure`, `exit` and `abort` never return, so nothing after them runs on this path. */
        private func terminateIfNeverReturns(_ value: ExprSyntax) {
            var called = ConcurrencyNoCheckThenWrite.withoutParentheses(value)
            if let tryExpression = called.as(TryExprSyntax.self) {
                called = tryExpression.expression
            }
            if let call = called.as(FunctionCallExprSyntax.self),
                let callee = call.calledExpression.as(DeclReferenceExprSyntax.self),
                ["fatalError", "preconditionFailure", "exit", "abort"].contains(callee.baseName.text)
            {
                cut()
            }
        }

        private func statement(_ node: StmtSyntax) {
            let label = pendingLabel
            pendingLabel = nil
            if let labeled = node.as(LabeledStmtSyntax.self) {
                pendingLabel = labeled.label.text
                statement(labeled.statement)
            }
            else if let expressionStatement = node.as(ExpressionStmtSyntax.self) {
                if let ifExpression = expressionStatement.expression.as(IfExprSyntax.self) {
                    self.ifExpression(ifExpression, label: label)
                }
                else if let switchExpression = expressionStatement.expression.as(SwitchExprSyntax.self) {
                    self.switchExpression(switchExpression, label: label)
                }
                else {
                    expression(expressionStatement.expression)
                    terminateIfNeverReturns(expressionStatement.expression)
                }
            }
            else if let guardStatement = node.as(GuardStmtSyntax.self) {
                let elseEntry = newBlock()
                conditions(guardStatement.conditions, failTo: elseEntry)
                let after = current
                current = elseEntry
                walk(guardStatement.body.statements)
                current = after
            }
            else if let whileStatement = node.as(WhileStmtSyntax.self) {
                let header = newBlock()
                edge(current, header)
                current = header
                let exit = newBlock()
                conditions(whileStatement.conditions, failTo: exit)
                jumps.append(
                    Jump(
                        label: label,
                        isLoop: true,
                        isSwitch: false,
                        breakTarget: exit,
                        continueTarget: header,
                        depth: scopes.count,
                    )
                )
                walk(whileStatement.body.statements)
                jumps.removeLast()
                edge(current, header)
                current = exit
            }
            else if let repeatStatement = node.as(RepeatStmtSyntax.self) {
                let bodyEntry = newBlock()
                let conditionBlock = newBlock()
                let exit = newBlock()
                edge(current, bodyEntry)
                current = bodyEntry
                jumps.append(
                    Jump(
                        label: label,
                        isLoop: true,
                        isSwitch: false,
                        breakTarget: exit,
                        continueTarget: conditionBlock,
                        depth: scopes.count,
                    )
                )
                walk(repeatStatement.body.statements)
                jumps.removeLast()
                edge(current, conditionBlock)
                current = conditionBlock
                expression(repeatStatement.condition)
                edge(current, bodyEntry)
                edge(current, exit)
                current = exit
            }
            else if let forStatement = node.as(ForStmtSyntax.self) {
                expression(forStatement.sequence)
                let header = newBlock()
                edge(current, header)
                current = header
                if forStatement.awaitKeyword != nil {
                    /* Each step of a `for await` suspends before the body runs, and before the loop ends. */
                    emit(.suspend(nil))
                    if forStatement.tryKeyword != nil, let handler = catches.last {
                        jump(to: handler.target, unwinding: handler.depth)
                    }
                }
                let exit = newBlock()
                edge(current, exit)
                let bodyEntry = newBlock()
                edge(current, bodyEntry)
                current = bodyEntry
                if let whereClause = forStatement.whereClause {
                    expression(whereClause.condition)
                    edge(current, header)
                    let matched = newBlock()
                    edge(current, matched)
                    current = matched
                }
                jumps.append(
                    Jump(
                        label: label,
                        isLoop: true,
                        isSwitch: false,
                        breakTarget: exit,
                        continueTarget: header,
                        depth: scopes.count,
                    )
                )
                walk(forStatement.body.statements)
                jumps.removeLast()
                edge(current, header)
                current = exit
            }
            else if let doStatement = node.as(DoStmtSyntax.self) {
                let join = newBlock()
                if label != nil {
                    jumps.append(
                        Jump(
                            label: label,
                            isLoop: false,
                            isSwitch: false,
                            breakTarget: join,
                            continueTarget: nil,
                            depth: scopes.count,
                        )
                    )
                }
                if doStatement.catchClauses.isEmpty {
                    walk(doStatement.body.statements)
                    edge(current, join)
                }
                else {
                    let dispatch = newBlock()
                    catches.append((dispatch, scopes.count))
                    walk(doStatement.body.statements)
                    catches.removeLast()
                    edge(current, join)
                    for clause in doStatement.catchClauses {
                        let entry = newBlock()
                        edge(dispatch, entry)
                        current = entry
                        for item in clause.catchItems {
                            if let whereClause = item.whereClause {
                                expression(whereClause.condition)
                            }
                        }
                        walk(clause.body.statements)
                        edge(current, join)
                    }
                }
                if label != nil {
                    jumps.removeLast()
                }
                current = join
            }
            else if let returnStatement = node.as(ReturnStmtSyntax.self) {
                if let value = returnStatement.expression {
                    expression(value)
                }
                jump(to: nil, unwinding: 0)
                cut()
            }
            else if let throwStatement = node.as(ThrowStmtSyntax.self) {
                expression(throwStatement.expression)
                if let handler = catches.last {
                    jump(to: handler.target, unwinding: handler.depth)
                }
                cut()
            }
            else if let breakStatement = node.as(BreakStmtSyntax.self) {
                let target =
                    breakStatement.label.map { name in jumps.last { $0.label == name.text } }
                    ?? jumps.last { $0.isLoop || $0.isSwitch }
                if let target {
                    jump(to: target.breakTarget, unwinding: target.depth)
                }
                cut()
            }
            else if let continueStatement = node.as(ContinueStmtSyntax.self) {
                let target =
                    continueStatement.label.map { name in jumps.last { $0.label == name.text } }
                    ?? jumps.last { $0.isLoop }
                if let target, let continueTarget = target.continueTarget {
                    jump(to: continueTarget, unwinding: target.depth)
                }
                cut()
            }
            else if node.is(FallThroughStmtSyntax.self) {
                if let next = fallthroughs.last, let next {
                    edge(current, next)
                }
                cut()
            }
            else if let deferStatement = node.as(DeferStmtSyntax.self) {
                if !scopes.isEmpty {
                    scopes[scopes.count - 1].append(deferStatement.body)
                }
            }
            else {
                for child in Evaluator.childExpressions(of: Syntax(node)) {
                    expression(child)
                }
            }
        }

        /* Each condition in turn, with an edge to the failure target after each one. */
        private func conditions(_ list: ConditionElementListSyntax, failTo failure: Int) {
            for element in list {
                switch element.condition {
                    case .expression(let value):
                        expression(value)
                    case .optionalBinding(let binding):
                        if let value = binding.initializer?.value {
                            expression(value)
                        }
                    case .matchingPattern(let matching):
                        expression(matching.initializer.value)
                    case .availability:
                        break
                }
                edge(current, failure)
                let next = newBlock()
                edge(current, next)
                current = next
            }
        }

        private func ifExpression(_ node: IfExprSyntax, label: String?) {
            let join = newBlock()
            let elseEntry = newBlock()
            if label != nil {
                jumps.append(
                    Jump(
                        label: label,
                        isLoop: false,
                        isSwitch: false,
                        breakTarget: join,
                        continueTarget: nil,
                        depth: scopes.count,
                    )
                )
            }
            conditions(node.conditions, failTo: elseEntry)
            walk(node.body.statements)
            edge(current, join)
            current = elseEntry
            switch node.elseBody {
                case .ifExpr(let elseIf)?:
                    ifExpression(elseIf, label: nil)
                case .codeBlock(let block)?:
                    walk(block.statements)
                case nil:
                    break
            }
            edge(current, join)
            if label != nil {
                jumps.removeLast()
            }
            current = join
        }

        private func switchExpression(_ node: SwitchExprSyntax, label: String?) {
            expression(node.subject)
            let dispatch = current
            let join = newBlock()
            let cases = node.cases.compactMap { element -> SwitchCaseSyntax? in
                if case .switchCase(let switchCase) = element {
                    return switchCase
                }
                return nil
            }
            let entries = cases.map { _ in newBlock() }
            for entry in entries {
                edge(dispatch, entry)
            }
            jumps.append(
                Jump(
                    label: label,
                    isLoop: false,
                    isSwitch: true,
                    breakTarget: join,
                    continueTarget: nil,
                    depth: scopes.count,
                )
            )
            for (index, switchCase) in cases.enumerated() {
                current = entries[index]
                if case .case(let caseLabel) = switchCase.label {
                    for item in caseLabel.caseItems {
                        if let whereClause = item.whereClause {
                            expression(whereClause.condition)
                        }
                    }
                }
                fallthroughs.append(index + 1 < entries.count ? entries[index + 1] : nil)
                walk(switchCase.statements)
                fallthroughs.removeLast()
                edge(current, join)
            }
            jumps.removeLast()
            current = join
        }

        /* An expression's events in evaluation order. Forks inside an expression are walked as one sequence: that only adds suspensions and writes to every path, never a path. */
        private func expression(_ node: ExprSyntax) {
            if node.is(ClosureExprSyntax.self) {
                return
            }
            if let ifExpression = node.as(IfExprSyntax.self) {
                self.ifExpression(ifExpression, label: nil)
                return
            }
            if let switchExpression = node.as(SwitchExprSyntax.self) {
                self.switchExpression(switchExpression, label: nil)
                return
            }
            if let infix = node.as(InfixOperatorExprSyntax.self), ConcurrencyNoLostUpdate.isAssignment(infix) {
                /* Before its operands, so a suspension in its own right side is the evaluator's to judge, path by path. */
                emit(.write(Syntax(infix)))
                expression(infix.leftOperand)
                expression(infix.rightOperand)
                return
            }
            if let awaitExpression = node.as(AwaitExprSyntax.self) {
                expression(awaitExpression.expression)
                emit(.suspend(Syntax(awaitExpression)))
                return
            }
            if let tryExpression = node.as(TryExprSyntax.self) {
                let throwsOut = tryExpression.questionOrExclamationMark == nil
                /* A throw before anything in it ran, always possible. */
                if throwsOut, let handler = catches.last {
                    jump(to: handler.target, unwinding: handler.depth)
                }
                expression(tryExpression.expression)
                /* A throw after everything in it ran, possible only when one call is the last thing it does and the only one that can throw. */
                if throwsOut, Self.isSingleCall(tryExpression.expression), let handler = catches.last {
                    jump(to: handler.target, unwinding: handler.depth)
                }
                return
            }
            for child in Evaluator.childExpressions(of: Syntax(node)) {
                expression(child)
            }
            if let call = node.as(FunctionCallExprSyntax.self),
                let member = call.calledExpression.as(MemberAccessExprSyntax.self), member.base != nil
            {
                emit(.write(Syntax(call)))
            }
            else if node.is(InOutExprSyntax.self) {
                emit(.write(Syntax(node)))
            }
        }

        /* A call, awaited or not, whose receiver and arguments hold no other call or subscript. */
        static func isSingleCall(_ expression: ExprSyntax) -> Bool {
            var operand = ConcurrencyNoCheckThenWrite.withoutParentheses(expression)
            if let awaitExpression = operand.as(AwaitExprSyntax.self) {
                operand = ConcurrencyNoCheckThenWrite.withoutParentheses(awaitExpression.expression)
            }
            guard let call = operand.as(FunctionCallExprSyntax.self) else { return false }
            var others = 0
            ConcurrencyNoLostUpdate.forEachOwnNode(Syntax(call)) { node in
                if node.id != call.id,
                    node.is(FunctionCallExprSyntax.self) || node.is(SubscriptCallExprSyntax.self)
                        || node.is(MacroExpansionExprSyntax.self)
                {
                    others += 1
                }
            }
            return others == 0
        }
    }
}
