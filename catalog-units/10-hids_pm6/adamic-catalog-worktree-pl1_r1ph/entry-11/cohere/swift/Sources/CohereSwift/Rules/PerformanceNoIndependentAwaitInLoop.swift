import Foundation
import SwiftOperators
import SwiftSyntax

/*
 No loop that awaits each item's own work one item at a time when nothing carries from one iteration to the
 next: `for entry in lanes { results.append(await entry.lane.snapshot()) }`. Every lane waits for the one
 before it to answer, so the loop takes the sum of their times where it could take the longest. The repair
 starts them together: `withTaskGroup`, collecting each result with its index when the order matters (a task
 group hands results back as they finish, not in the order they were started), or `async let` for a fixed few,
 or a bounded group when the collection can be large or the far side rate-limits.

 It ports `nexus/performance-no-independent-await-in-loop`, whose question is the same: report a loop over a
 collection only when its iterations can be shown not to depend on each other, so the finding is nearly always
 a real `Promise.all`. Its conditions map to Swift like this:
 - The loop walks a collection. A `for`-`in` that is not `for await` or `for try await`, over anything but a
   range or `stride`, or the indexed form `for index in items.indices` or `for index in start..<items.count`
   (the TypeScript `i < xs.length`). Any other range (`0..<60`, `1...3`, `0...items.count`, `0..<attempts`), a
   `stride`, `while` and `repeat` never fire: those are probes, retries and polling. A wildcard item (`for _
   in`) has no item to do work on, so it never fires either. A `for` binding is a `let`, so TypeScript's "the
   body writes its index" has no Swift shape.
 - The loop itself waits. An `await` in the body, not inside a closure, a nested declaration, or a nested
   loop's body or condition (that loop is judged on its own; a nested `for`'s sequence runs once per pass and
   is this loop's). An await on the item itself (`await task.value` over tasks started before the loop, or
   `await tasks[index].value`) serializes nothing and does not count.
 - No loop-carried state. Names are resolved by Swift's own scoping, read from the tree: a name declared in
   the loop (the item, a `let` or `var` in the body, an `if let`, a `catch`, a `case` pattern, a closure
   parameter) is the iteration's own, and since Swift has no function-scoped `var`, TypeScript's "a `var` in
   the body" has no Swift shape; one declared in the enclosing function before the loop, or a parameter of it,
   is a local of the function; anything else (a property reached through implicit `self`, a global, a capture,
   an `inout` parameter, `self` itself) is shared state. Then, as in TypeScript, a write to a local of the
   function (`total += n`, `cursor = next`, `&buffer`) is silence, and so is `pop`, `popFirst`, `popLast`,
   `removeFirst`, `removeLast`, `removeAll`, `next`, `read`, `sort`, `reverse`, `shuffle`, `swapAt` or
   `partition` on a local or shared receiver. Collecting results is allowed on a local of the function:
   `append`, `insert`, `updateValue`, `removeValue`, `merge`, `formUnion`, `add`, `set`, any `add`, `append`
   or `insert` followed by a capitalized noun, `local[key] = value` and `local.member = value`, provided the
   body, the loop's `where` clause and its sequence never read the filled path back (`results.count`,
   `queue` walked while it grows). Paths compare as member chains, as in TypeScript. Stricter than
   TypeScript, because an actor's or an object's properties are what the next pass (and every other caller)
   reads, and Swift's value types change in place: any write to shared state (`self.appearance = x`,
   `cache.append(x)`), any method called on `self` (`self.note(x)`), any bare lower-case call (`note(x)`, a
   method on implicit `self` or a global function, either of which may write shared state), and any method
   other than a collecting one on a local `var` (which may be `mutating`) is silence.
 - No exit on an awaited value. A `break` or labeled `continue` that leaves this loop, a `return`, or a `throw`
   that no exhaustive `catch` in the body handles, is silence when what decides it reads an awaited value: an
   enclosing `if`, `guard`, `switch`, `case`, `while`, `repeat` or nested `for`; a `catch` whose `do` awaits; a
   `do` that awaits before the exit and has a `catch`; an earlier statement that jumps on such a value; or the
   exit's own expression. "Awaited" is a fixpoint over the body's declarations, `if let`/`guard let` bindings,
   assignments, `case` patterns and catch bindings, as in TypeScript. Stricter than TypeScript: a `try` (not
   `try?` or `try!`) that no exhaustive `catch` in the body handles is an exit too, because Swift spells the
   failure path where TypeScript leaves it implicit in the await, and a loop that stops at the first failure
   never starts the rest, which a task group would already have started. A plain `continue`, and a
   `do`/`catch` that records the failure and moves on, are fine.
 - No ordered side effect, by name, anywhere in the body including closures: `print`, `debugPrint`, `dump`,
   `NSLog`, `os_log`, `fputs`, `fputc`, `puts`, `putchar`, `fflush`, `write`, `usleep`, `nanosleep`,
   `asyncAfter`, `yield` (`Task.yield()`, a continuation's `yield`), a bare `log(...)` or any function whose
   name ends in `Log` (`presenceLog`); names starting `sleep` (`Task.sleep`, `Thread.sleep`), `delay`, `wait`,
   `pause`, `throttle`, `backoff`, `rateLimit` or `print`; a name or receiver containing `progress` or
   `spinner`; and `log`, `info`, `notice`, `warning`, `warn`, `error`, `debug`, `trace`, `critical` or `fault`
   on a receiver whose last segment contains `log` (`logger.info`, `poolLog.info`). Inside a `catch` in the
   body, and one level into every function of the called name declared in this file, the list narrows as
   TypeScript's does: a log receiver's `error`, `warning`, `warn`, `debug`, `trace`, `critical` and `fault`, a
   write to `standardError` or `stderr`, and `asyncAfter` report a failure or time a request, so they do not
   count, and a callee's own `catch` blocks are not read. TypeScript follows the callee through the checker,
   imports included; this rule reads the file it is given, so a callee elsewhere is not read (see below).

 Stricter than TypeScript, for actor isolation: every await the loop owns must be the item's own work. An
 `await` covers its whole expression, so the syntax cannot say which call in it suspends: every call, member
 chain and subscript in the operand (outside closures) must reach the item binding (`entry.lane.snapshot()`),
 `items[index]` in the indexed form, or a body `let` bound to such a chain, and at least one must; a bare name
 read as a value and a literal may stand beside them (`await entry.lane.snapshot(resolveSessionId: resolver) <
 0`). An await on `self`, on an implicit-`self` method, on a free function, on a type's static member or on any
 outer value is the same actor (or, for all the syntax can say, the same actor) on every pass, so the
 iterations would queue on it anyway, and the loop is left alone. That is what keeps AhraOS Presence's probes
 silent (`await self.save()` after setting `self.appearance`, `await settled()`, `await fromOld.nearest(to:)`,
 `withCheckedContinuation` around a callback), along with their `presenceLog` lines, `Task.sleep` pacing and
 outer counters.

 The finding spans the loop header, from `for` to the end of its sequence or `where` clause. Measured on
 2026-10-03: `ProviderProxy.swift:575` and `:620` in ahraos-macos, and nothing in ahraos-presence's 1,196 files.

 What it accepts as safe, and why: everything above that silences, because it is a reason the code shows for
 going one at a time; a nested `for await` in the body (the iteration consumes a stream, which is not a plain
 map); an `async let` (its work started already, so reading it later, by name or through a member, is not the
 item's work being started); and an await in the loop's own `where` clause.

 Known misses, every one a finding not made:
 - An await on a free function or an implicit-`self` method that is in fact nonisolated (`await
   fetchItem(item)`, TypeScript's commonest shape), a `withCheckedContinuation` around per-item callback work,
   and an await whose operand also calls something not rooted at the item (`await results.append(item.load())`,
   `await item.send(makePayload())`). The syntax cannot tell a nonisolated function from one isolated to the
   caller's actor.
 - A loop that also calls a pure global function (`min`, `zip`) or a non-mutating method on a local `var`,
   which the stricter state rules cannot tell from one that writes.
 - A `try` that leaves the loop on the first failure, even where starting every item is harmless.
 - An indexed walk bounded by anything but `.count` (`0..<items.endIndex`), or one whose awaited receiver is
   another collection subscripted by the index (`others[index]`).

 What it cannot see, and so could report where it should not (none did on either proving ground):
 - A callee that prints or sleeps, declared in another file or reached through an initializer; the callee scan
   reads this file's functions by name. Same-named functions in this file are all read, so that scan can
   silence more than it should, never less.
 - Shared state behind the item: two items that are the same object or actor, or an item's method isolated to
   one global actor (`@MainActor`), where starting them together gains nothing. TypeScript's rule has the same
   floor for a shared browser page or a filesystem race two calls down.
 - An ordered effect spelled some other way than the names above, as in TypeScript.
 */
public struct PerformanceNoIndependentAwaitInLoop: FileRule {
    public let name = "cohere-swift/performance-no-independent-await-in-loop"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* A flagged loop holds both words; nothing else can. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("for") && file.source.contains("await")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        /* Folded, so `0..<items.count` is one range and `total += n` one assignment. Folding moves no token. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        guard !visitor.loops.isEmpty else { return [] }
        let callees = Callees(functions: visitor.functions)
        return visitor.loops.compactMap { loop in
            guard let analysis = Analysis(loop: loop, callees: callees), analysis.isIndependent() else { return nil }
            let headerEnd = loop.whereClause.map { Syntax($0) } ?? Syntax(loop.sequence)
            let start = file.locations.location(for: loop.forKeyword.positionAfterSkippingLeadingTrivia)
            let end = file.locations.location(for: headerEnd.endPositionBeforeTrailingTrivia)
            let message = RuleMessages.PerformanceNoIndependentAwaitInLoop.independentAwaitInLoop()
            return FindingRecord(
                source: .rule,
                file: file.url.path,
                line: start.line,
                column: start.column,
                endLine: end.line,
                endColumn: end.column,
                severity: .error,
                rule: name,
                messageId: message.id,
                message: message.text,
            )
        }
    }

    /* Every `for`-`in`, and every function declared in the file by name for the one-level callee scan. */
    final class Visitor: SyntaxVisitor {
        private(set) var loops: [ForStmtSyntax] = []
        private(set) var functions: [String: [FunctionDeclSyntax]] = [:]

        override func visit(_ node: ForStmtSyntax) -> SyntaxVisitorContinueKind {
            loops.append(node)
            return .visitChildren
        }

        override func visit(_ node: FunctionDeclSyntax) -> SyntaxVisitorContinueKind {
            functions[PerformanceNoIndependentAwaitInLoop.plainName(node.name), default: []].append(node)
            return .visitChildren
        }
    }

    // MARK: Names

    /* What a name in the loop refers to, by Swift's own lexical scoping. */
    enum Binding {
        /* The loop's own item (or index), bound by its pattern. */
        case item
        /* Declared inside the loop, with a plain `let`'s initializer, kept so a body local can carry the item's root. */
        case inner(letInitializer: ExprSyntax?)
        /* Declared in the enclosing function before the loop: a parameter or a local, and whether it is a `var`. */
        case functionLocal(isVariable: Bool)
        /* `self`, a property reached through implicit `self`, a global, an `inout` parameter, or a capture. */
        case shared
    }

    /* A token's name without backticks, so `` `default` `` and `default` compare equal. */
    static func plainName(_ token: TokenSyntax) -> String {
        String(token.text.filter { $0 != "`" })
    }

    /* Every name a pattern binds, including `let` inside an expression pattern (`case let .done(value)`). */
    static func boundNames(_ node: some SyntaxProtocol) -> Set<String> {
        let collector = BoundNameCollector(viewMode: .sourceAccurate)
        collector.walk(node)
        return collector.names
    }

    final class BoundNameCollector: SyntaxVisitor {
        private(set) var names: Set<String> = []
        private var bindingDepth = 0

        override func visit(_ node: IdentifierPatternSyntax) -> SyntaxVisitorContinueKind {
            names.insert(PerformanceNoIndependentAwaitInLoop.plainName(node.identifier))
            return .skipChildren
        }

        override func visit(_ node: ValueBindingPatternSyntax) -> SyntaxVisitorContinueKind {
            bindingDepth += 1
            return .visitChildren
        }

        override func visitPost(_ node: ValueBindingPatternSyntax) {
            bindingDepth -= 1
        }

        /* Under `let`, a bare name in an expression pattern binds it rather than comparing against it. */
        override func visit(_ node: DeclReferenceExprSyntax) -> SyntaxVisitorContinueKind {
            if bindingDepth > 0 {
                names.insert(PerformanceNoIndependentAwaitInLoop.plainName(node.baseName))
            }
            return .skipChildren
        }
    }

    /* A function, accessor, initializer, subscript or closure: code that runs when called, not when written. */
    static func isFunctionBoundary(_ node: Syntax) -> Bool {
        node.is(FunctionDeclSyntax.self) || node.is(InitializerDeclSyntax.self) || node.is(DeinitializerDeclSyntax.self)
            || node.is(AccessorDeclSyntax.self) || node.is(SubscriptDeclSyntax.self) || node.is(ClosureExprSyntax.self)
    }

    /*
     A boundary for awaits and jumps: any function-like node, a nested type, or a computed property's accessors.
     Not `#if`, whose active and inactive clauses both stay in view: an effect under `#if DEBUG` is still one.
     */
    static func isWalkBoundary(_ node: Syntax) -> Bool {
        isFunctionBoundary(node) || node.is(AccessorBlockSyntax.self) || node.is(StructDeclSyntax.self)
            || node.is(ClassDeclSyntax.self)
            || node.is(EnumDeclSyntax.self) || node.is(ActorDeclSyntax.self) || node.is(ProtocolDeclSyntax.self)
            || node.is(ExtensionDeclSyntax.self)
    }

    /* What a scope node declares under the name, seen from `child`, or nil. `isShared` marks an `inout` parameter. */
    struct Declared {
        var letInitializer: ExprSyntax?
        var isVariable = false
        var isShared = false
    }

    static func declared(_ name: String, in parent: Syntax, below child: Syntax) -> Declared? {
        if let list = parent.as(CodeBlockItemListSyntax.self) {
            for statement in list where statement.position < child.position {
                if let variable = statement.item.as(VariableDeclSyntax.self) {
                    for binding in variable.bindings where boundNames(binding.pattern).contains(name) {
                        /* An `async let` started its work already; reading it later is not the item's own work being started. */
                        let isPlainLet =
                            variable.bindingSpecifier.tokenKind == .keyword(.let)
                            && !variable.modifiers.contains { $0.name.tokenKind == .keyword(.async) }
                        return Declared(
                            letInitializer: isPlainLet ? binding.initializer?.value : nil,
                            isVariable: variable.bindingSpecifier.tokenKind == .keyword(.var),
                        )
                    }
                }
                else if let guardStatement = statement.item.as(GuardStmtSyntax.self),
                    conditionNames(guardStatement.conditions).contains(name)
                {
                    return Declared(
                        isVariable: guardStatement.conditions.contains { element in
                            element.condition.as(OptionalBindingConditionSyntax.self)?.bindingSpecifier.tokenKind
                                == .keyword(.var)
                        }
                    )
                }
                else if let function = statement.item.as(FunctionDeclSyntax.self), plainName(function.name) == name {
                    return Declared()
                }
            }
            return nil
        }
        if let conditions = parent.as(ConditionElementListSyntax.self) {
            for element in conditions
            where element.position < child.position && conditionNames([element]).contains(name) {
                return Declared()
            }
            return nil
        }
        if let ifExpression = parent.as(IfExprSyntax.self), child.id == ifExpression.body.id {
            return conditionNames(ifExpression.conditions).contains(name) ? Declared() : nil
        }
        if let whileStatement = parent.as(WhileStmtSyntax.self), child.id == whileStatement.body.id {
            return conditionNames(whileStatement.conditions).contains(name) ? Declared() : nil
        }
        if let forStatement = parent.as(ForStmtSyntax.self),
            child.id == forStatement.body.id || child.id == forStatement.whereClause?.id
        {
            return boundNames(forStatement.pattern).contains(name) ? Declared() : nil
        }
        if let catchClause = parent.as(CatchClauseSyntax.self), child.id == catchClause.body.id {
            if catchClause.catchItems.isEmpty {
                return name == "error" ? Declared() : nil
            }
            return catchClause.catchItems.contains { item in item.pattern.map { boundNames($0).contains(name) } ?? false
            } ? Declared() : nil
        }
        if let switchCase = parent.as(SwitchCaseSyntax.self), child.id == switchCase.statements.id,
            case let .case(label) = switchCase.label
        {
            return label.caseItems.contains { boundNames($0.pattern).contains(name) } ? Declared() : nil
        }
        if let closure = parent.as(ClosureExprSyntax.self) {
            return closureNames(closure).contains(name) ? Declared() : nil
        }
        if let function = parent.as(FunctionDeclSyntax.self) {
            return parameter(name, in: function.signature.parameterClause.parameters)
        }
        if let initializer = parent.as(InitializerDeclSyntax.self) {
            return parameter(name, in: initializer.signature.parameterClause.parameters)
        }
        if let subscriptDeclaration = parent.as(SubscriptDeclSyntax.self) {
            return parameter(name, in: subscriptDeclaration.parameterClause.parameters)
        }
        if let accessor = parent.as(AccessorDeclSyntax.self) {
            let names: Set<String> = accessor.parameters.map { [plainName($0.name)] } ?? ["newValue", "oldValue"]
            return names.contains(name) ? Declared() : nil
        }
        return nil
    }

    static func conditionNames(_ conditions: some Sequence<ConditionElementSyntax>) -> Set<String> {
        var names: Set<String> = []
        for element in conditions {
            switch element.condition {
                case let .optionalBinding(binding): names.formUnion(boundNames(binding.pattern))
                case let .matchingPattern(matching): names.formUnion(boundNames(matching.pattern))
                default: break
            }
        }
        return names
    }

    static func closureNames(_ closure: ClosureExprSyntax) -> Set<String> {
        var names: Set<String> = []
        if let signature = closure.signature {
            for capture in signature.capture?.items ?? [] {
                names.insert(plainName(capture.name))
            }
            switch signature.parameterClause {
                case let .simpleInput(parameters):
                    for parameter in parameters {
                        names.insert(plainName(parameter.name))
                    }
                case let .parameterClause(clause):
                    for parameter in clause.parameters {
                        names.insert(plainName(parameter.secondName ?? parameter.firstName))
                    }
                case nil:
                    break
            }
        }
        return names
    }

    static func parameter(_ name: String, in parameters: FunctionParameterListSyntax) -> Declared? {
        for parameter in parameters where plainName(parameter.secondName ?? parameter.firstName) == name {
            let isInout =
                parameter.type.as(AttributedTypeSyntax.self)?.specifiers.contains { specifier in
                    specifier.as(SimpleTypeSpecifierSyntax.self)?.specifier.tokenKind == .keyword(.inout)
                } ?? false
            return Declared(isShared: isInout)
        }
        return nil
    }

    // MARK: Expressions

    /* Removes parentheses, `try`, `!` and `?`, none of which change which value an expression names. */
    static func stripped(_ expression: ExprSyntax) -> ExprSyntax {
        var current = expression
        while true {
            if let tuple = current.as(TupleExprSyntax.self), tuple.elements.count == 1, let only = tuple.elements.first,
                only.label == nil
            {
                current = only.expression
            }
            else if let tryExpression = current.as(TryExprSyntax.self) {
                current = tryExpression.expression
            }
            else if let unwrap = current.as(ForceUnwrapExprSyntax.self) {
                current = unwrap.expression
            }
            else if let chaining = current.as(OptionalChainingExprSyntax.self) {
                current = chaining.expression
            }
            else {
                return current
            }
        }
    }

    /* A member chain spelled as names (`self.cache.entries` as `self`, `cache`, `entries`) with its root reference, or nil. */
    static func chain(_ expression: ExprSyntax) -> (root: DeclReferenceExprSyntax, segments: [String])? {
        let expression = stripped(expression)
        if let reference = expression.as(DeclReferenceExprSyntax.self) {
            return (reference, [plainName(reference.baseName)])
        }
        if let member = expression.as(MemberAccessExprSyntax.self), let base = member.base, let inner = chain(base) {
            return (inner.root, inner.segments + [plainName(member.declName.baseName)])
        }
        return nil
    }

    /* The reference a chain starts from, through calls, members and subscripts (`entry` in `entry.lane[0].snapshot()`), or nil. */
    static func baseReference(_ expression: ExprSyntax) -> DeclReferenceExprSyntax? {
        let expression = stripped(expression)
        if let reference = expression.as(DeclReferenceExprSyntax.self) {
            return reference
        }
        if let call = expression.as(FunctionCallExprSyntax.self) {
            return baseReference(call.calledExpression)
        }
        if let member = expression.as(MemberAccessExprSyntax.self), let base = member.base {
            return baseReference(base)
        }
        if let subscriptCall = expression.as(SubscriptCallExprSyntax.self) {
            return baseReference(subscriptCall.calledExpression)
        }
        return nil
    }

    /* Whether a node is the callee, base or subscripted value of a chain that goes on above it, through `?` and `!`. */
    static func continuesChain(_ node: Syntax) -> Bool {
        var current = node
        while let parent = current.parent,
            parent.is(ForceUnwrapExprSyntax.self) || parent.is(OptionalChainingExprSyntax.self)
        {
            current = parent
        }
        guard let parent = current.parent else { return false }
        if let call = parent.as(FunctionCallExprSyntax.self) {
            return call.calledExpression.id == current.id
        }
        if let member = parent.as(MemberAccessExprSyntax.self) {
            return member.base?.id == current.id
        }
        if let subscriptCall = parent.as(SubscriptCallExprSyntax.self) {
            return subscriptCall.calledExpression.id == current.id
        }
        return false
    }

    /* A reference that is a member's name (`.lane` in `entry.lane`) rather than a read of a binding. */
    static func isMemberName(_ reference: DeclReferenceExprSyntax) -> Bool {
        guard let member = reference.parent?.as(MemberAccessExprSyntax.self) else { return false }
        return member.declName.id == reference.id
    }

    /* The longest member chain a read starts, so `results.count` is one read of `results.count`. */
    static func readSegments(_ reference: DeclReferenceExprSyntax) -> [String] {
        var segments = [plainName(reference.baseName)]
        var current = Syntax(reference)
        while let parent = current.parent {
            if parent.is(ForceUnwrapExprSyntax.self) || parent.is(OptionalChainingExprSyntax.self) {
                current = parent
                continue
            }
            guard let member = parent.as(MemberAccessExprSyntax.self), member.base?.id == current.id else { break }
            segments.append(plainName(member.declName.baseName))
            current = parent
        }
        return segments
    }

    static func pathsOverlap(_ first: [String], _ second: [String]) -> Bool {
        guard !first.isEmpty, !second.isEmpty else { return false }
        return zip(first, second).allSatisfy { $0 == $1 }
    }

    // MARK: Ordered effects

    static let orderedNames: Set<String> = [
        "print", "debugPrint", "dump", "NSLog", "os_log", "fputs", "fputc", "puts", "putchar", "fflush", "write",
        "usleep", "nanosleep", "asyncAfter", "yield", "log",
    ]
    static let pacingPrefixes = ["sleep", "delay", "wait", "pause", "throttle", "backoff", "ratelimit", "print"]
    static let logMethods: Set<String> = [
        "log", "info", "notice", "warning", "warn", "error", "debug", "trace", "critical", "fault",
    ]
    /* The methods a log receiver uses to report a problem rather than to show progress. */
    static let quietLogMethods: Set<String> = ["warning", "warn", "error", "debug", "trace", "critical", "fault"]

    /* The called name and its receiver's segments, through `?` and `!`. */
    static func calledName(_ call: FunctionCallExprSyntax) -> (name: String, receiver: [String]?)? {
        let callee = stripped(call.calledExpression)
        if let reference = callee.as(DeclReferenceExprSyntax.self) {
            return (plainName(reference.baseName), nil)
        }
        if let member = callee.as(MemberAccessExprSyntax.self) {
            return (plainName(member.declName.baseName), member.base.flatMap { chain($0)?.segments })
        }
        return nil
    }

    /*
     Whether a call is an effect whose order the reader can see. `narrow` is the question asked inside a `catch`
     and inside a callee: a failure report and a request timer are not output a reader follows.
     */
    static func isOrderedCall(_ call: FunctionCallExprSyntax, narrow: Bool) -> Bool {
        guard let called = calledName(call) else { return false }
        let lastReceiver = called.receiver?.last?.lowercased() ?? ""
        if narrow {
            if quietLogMethods.contains(called.name) && lastReceiver.contains("log") {
                return false
            }
            if called.receiver?.contains("standardError") == true || called.name == "asyncAfter" {
                return false
            }
            if called.name == "fputs" || called.name == "fputc",
                call.arguments.contains(where: {
                    stripped($0.expression).as(DeclReferenceExprSyntax.self)?.baseName.text == "stderr"
                })
            {
                return false
            }
        }
        if orderedNames.contains(called.name) || called.name.hasSuffix("Log") {
            return true
        }
        let lowerName = called.name.lowercased()
        if pacingPrefixes.contains(where: { lowerName.hasPrefix($0) }) {
            return true
        }
        if lowerName.contains("progress") || lowerName.contains("spinner") || lastReceiver.contains("progress")
            || lastReceiver.contains("spinner")
        {
            return true
        }
        return logMethods.contains(called.name) && lastReceiver.contains("log")
    }

    /* The file's functions by name, and which of their bodies make an ordered call, memoized across loops. */
    final class Callees {
        let functions: [String: [FunctionDeclSyntax]]
        private var answers: [SyntaxIdentifier: Bool] = [:]

        init(functions: [String: [FunctionDeclSyntax]]) {
            self.functions = functions
        }

        /*
         One level, deliberately, and every same-named function in the file, since the syntax cannot say which
         overload is called. Reading too many can only silence a loop. A callee's `catch` is its failure path.
         */
        func makesOrderedCall(_ name: String) -> Bool {
            for function in functions[name] ?? [] {
                guard let body = function.body else { continue }
                if let known = answers[function.id] {
                    if known { return true }
                    continue
                }
                let scanner = OrderedCallScanner(viewMode: .sourceAccurate)
                scanner.walk(body)
                answers[function.id] = scanner.found
                if scanner.found { return true }
            }
            return false
        }
    }

    final class OrderedCallScanner: SyntaxVisitor {
        private(set) var found = false

        override func visit(_ node: CatchClauseSyntax) -> SyntaxVisitorContinueKind {
            .skipChildren
        }

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            if PerformanceNoIndependentAwaitInLoop.isOrderedCall(node, narrow: true) {
                found = true
                return .skipChildren
            }
            return .visitChildren
        }
    }

    // MARK: The analysis of one loop

    struct Analysis {
        let loop: ForStmtSyntax
        let body: CodeBlockSyntax
        let itemNames: Set<String>
        let label: String?
        /* What the loop walks: the sequence, or `items` in `items.indices` and `start..<items.count`. */
        let collection: ExprSyntax
        let isIndexed: Bool
        let callees: Callees

        init?(loop: ForStmtSyntax, callees: Callees) {
            /* `for await` and `for try await` take the next element when it is ready: the deliberate sequential case. */
            guard loop.awaitKeyword == nil, loop.tryKeyword == nil else { return nil }
            let itemNames = PerformanceNoIndependentAwaitInLoop.boundNames(loop.pattern)
            guard !itemNames.isEmpty else { return nil }
            let sequence = PerformanceNoIndependentAwaitInLoop.stripped(loop.sequence)
            if let range = sequence.as(InfixOperatorExprSyntax.self),
                let operation = range.operator.as(BinaryOperatorExprSyntax.self),
                operation.operator.text.hasPrefix("..")
            {
                /* `start..<items.count` is the indexed walk; every other range is a probe, a retry or a count. */
                guard operation.operator.text == "..<",
                    let bound = PerformanceNoIndependentAwaitInLoop.stripped(range.rightOperand).as(
                        MemberAccessExprSyntax.self
                    ),
                    bound.declName.baseName.text == "count", bound.declName.argumentNames == nil, let base = bound.base,
                    PerformanceNoIndependentAwaitInLoop.chain(base) != nil
                else { return nil }
                self.collection = base
                self.isIndexed = true
            }
            else if sequence.is(PrefixOperatorExprSyntax.self) || sequence.is(PostfixOperatorExprSyntax.self) {
                /* A one-sided range (`0...`) counts without end. */
                return nil
            }
            else if let member = sequence.as(MemberAccessExprSyntax.self), member.declName.baseName.text == "indices",
                let base = member.base,
                PerformanceNoIndependentAwaitInLoop.chain(base) != nil
            {
                self.collection = base
                self.isIndexed = true
            }
            else if let call = sequence.as(FunctionCallExprSyntax.self),
                call.calledExpression.as(DeclReferenceExprSyntax.self)?.baseName.text == "stride"
            {
                return nil
            }
            else {
                self.collection = sequence
                self.isIndexed = false
            }
            self.loop = loop
            self.body = loop.body
            self.itemNames = itemNames
            self.label = loop.parent?.as(LabeledStmtSyntax.self).map {
                PerformanceNoIndependentAwaitInLoop.plainName($0.label)
            }
            self.callees = callees
        }

        /* The judgments, cheapest first. */
        func isIndependent() -> Bool {
            guard holdsOnlyItsItemsAwaits() else { return false }
            if hasOrderedSideEffect() || carriesState() {
                return false
            }
            let tainted = taintedNames()
            return !exitsOnAwaitedValue(tainted: tainted)
        }

        // MARK: Resolving names

        func resolve(_ reference: DeclReferenceExprSyntax) -> Binding {
            let name = PerformanceNoIndependentAwaitInLoop.plainName(reference.baseName)
            if name == "self" || name == "Self" || name == "super" {
                return .shared
            }
            var insideLoop = true
            var child = Syntax(reference)
            var current = child.parent
            while let parent = current {
                if parent.id == loop.id {
                    if child.id == body.id || child.id == loop.whereClause?.id, itemNames.contains(name) {
                        return .item
                    }
                    insideLoop = false
                }
                else if let found = PerformanceNoIndependentAwaitInLoop.declared(name, in: parent, below: child) {
                    if !insideLoop {
                        return found.isShared ? .shared : .functionLocal(isVariable: found.isVariable)
                    }
                    return .inner(letInitializer: found.letInitializer)
                }
                else if name.hasPrefix("$"), let closure = parent.as(ClosureExprSyntax.self),
                    closure.signature?.parameterClause == nil
                {
                    return insideLoop ? .inner(letInitializer: nil) : .functionLocal(isVariable: false)
                }
                /* Past the function that holds the loop, a name is a property, a global or a capture. */
                if !insideLoop && PerformanceNoIndependentAwaitInLoop.isFunctionBoundary(parent) {
                    return .shared
                }
                /* Top-level code's locals are globals. */
                if !insideLoop, parent.is(CodeBlockItemListSyntax.self),
                    parent.parent?.is(SourceFileSyntax.self) == true
                {
                    return .shared
                }
                child = parent
                current = parent.parent
            }
            return .shared
        }

        // MARK: Its own awaits

        /*
         True when the loop owns at least one await and every await it owns is the item's own work. An await on
         the item itself (a task started before the loop) neither counts nor blocks.
         */
        func holdsOnlyItsItemsAwaits() -> Bool {
            if let whereClause = loop.whereClause, containsAwait(Syntax(whereClause), before: nil) {
                return false
            }
            var found = false
            var blocked = false
            func visit(_ node: Syntax) {
                if blocked || PerformanceNoIndependentAwaitInLoop.isWalkBoundary(node) {
                    return
                }
                if let awaitExpression = node.as(AwaitExprSyntax.self) {
                    if isItemItself(awaitExpression.expression) {
                        return
                    }
                    if isItemRooted(awaitExpression.expression, depth: 0) {
                        found = true
                    }
                    else {
                        blocked = true
                        return
                    }
                }
                /* A nested loop's body and condition belong to that loop; only a `for`'s sequence runs once per pass of this one. */
                if let nested = node.as(ForStmtSyntax.self) {
                    if nested.awaitKeyword != nil {
                        blocked = true
                        return
                    }
                    visit(Syntax(nested.sequence))
                    return
                }
                if node.is(WhileStmtSyntax.self) || node.is(RepeatStmtSyntax.self) {
                    return
                }
                for child in node.children(viewMode: .sourceAccurate) {
                    visit(child)
                }
            }
            visit(Syntax(body))
            return found && !blocked
        }

        /* `item`, `item.value` or `item.result`, or `items[index]` and its `.value` in the indexed form. */
        func isItemItself(_ operand: ExprSyntax) -> Bool {
            var expression = PerformanceNoIndependentAwaitInLoop.stripped(operand)
            if let member = expression.as(MemberAccessExprSyntax.self), let base = member.base,
                ["value", "result"].contains(member.declName.baseName.text)
            {
                expression = PerformanceNoIndependentAwaitInLoop.stripped(base)
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self), case .item = resolve(reference) {
                return true
            }
            return isIndexedElement(expression)
        }

        /* `items[index]`, where `items` is the collection the indexed loop counts and `index` its item. */
        func isIndexedElement(_ expression: ExprSyntax) -> Bool {
            guard isIndexed,
                let subscriptCall = PerformanceNoIndependentAwaitInLoop.stripped(expression).as(
                    SubscriptCallExprSyntax.self
                ),
                subscriptCall.arguments.count == 1, let argument = subscriptCall.arguments.first, argument.label == nil,
                let index = PerformanceNoIndependentAwaitInLoop.stripped(argument.expression).as(
                    DeclReferenceExprSyntax.self
                ),
                case .item = resolve(index),
                let walked = PerformanceNoIndependentAwaitInLoop.chain(subscriptCall.calledExpression),
                let counted = PerformanceNoIndependentAwaitInLoop.chain(collection)
            else { return false }
            return walked.segments == counted.segments
        }

        /*
         Whether an await's operand is work reached from the item. An `await` covers its whole expression, so the
         syntax cannot say which call in it suspends: every call, member chain and subscript in the operand
         (outside closures) must reach the item, `items[index]`, or a body `let` bound to such a chain, and at
         least one must. A bare name read as a value (`resolver`) and a literal are allowed beside them, so
         `await entry.lane.snapshot(resolveSessionId: resolver) < 0` is the item's work.
         */
        func isItemRooted(_ operand: ExprSyntax, depth: Int) -> Bool {
            guard depth < 8 else { return false }
            var rooted = false
            var allRooted = true
            func visit(_ node: Syntax) {
                if !allRooted || node.is(ClosureExprSyntax.self) {
                    return
                }
                if node.is(AwaitExprSyntax.self) || node.is(MacroExpansionExprSyntax.self) {
                    allRooted = false
                    return
                }
                let isChain =
                    node.is(FunctionCallExprSyntax.self) || node.is(SubscriptCallExprSyntax.self)
                    || node.as(MemberAccessExprSyntax.self)?.base != nil
                if isChain, !PerformanceNoIndependentAwaitInLoop.continuesChain(node),
                    let expression = node.as(ExprSyntax.self)
                {
                    if rootIsItem(expression, depth: depth) {
                        rooted = true
                    }
                    else {
                        allRooted = false
                        return
                    }
                }
                for child in node.children(viewMode: .sourceAccurate) {
                    visit(child)
                }
            }
            visit(Syntax(operand))
            return rooted && allRooted
        }

        /* Whether a chain, walked down through calls, members and subscripts, reaches the item, `items[index]`, or a body `let` bound to such a chain. */
        func rootIsItem(_ expression: ExprSyntax, depth: Int) -> Bool {
            var current = PerformanceNoIndependentAwaitInLoop.stripped(expression)
            while true {
                if isIndexedElement(current) {
                    return true
                }
                if let call = current.as(FunctionCallExprSyntax.self) {
                    current = PerformanceNoIndependentAwaitInLoop.stripped(call.calledExpression)
                }
                else if let member = current.as(MemberAccessExprSyntax.self), let base = member.base {
                    current = PerformanceNoIndependentAwaitInLoop.stripped(base)
                }
                else if let subscriptCall = current.as(SubscriptCallExprSyntax.self) {
                    current = PerformanceNoIndependentAwaitInLoop.stripped(subscriptCall.calledExpression)
                }
                else if let reference = current.as(DeclReferenceExprSyntax.self) {
                    switch resolve(reference) {
                        case .item:
                            return true
                        case let .inner(letInitializer):
                            guard let initializer = letInitializer else { return false }
                            return isItemRooted(initializer, depth: depth + 1)
                        case .functionLocal, .shared:
                            return false
                    }
                }
                else {
                    return false
                }
            }
        }

        func containsAwait(_ node: Syntax, before position: AbsolutePosition?) -> Bool {
            if PerformanceNoIndependentAwaitInLoop.isWalkBoundary(node) {
                return false
            }
            if node.is(AwaitExprSyntax.self), position.map({ node.position < $0 }) ?? true {
                return true
            }
            return node.children(viewMode: .sourceAccurate).contains { containsAwait($0, before: position) }
        }

        // MARK: Ordered effects

        func hasOrderedSideEffect() -> Bool {
            func visit(_ node: Syntax, insideCatch: Bool) -> Bool {
                if let call = node.as(FunctionCallExprSyntax.self) {
                    if PerformanceNoIndependentAwaitInLoop.isOrderedCall(call, narrow: insideCatch) {
                        return true
                    }
                    if let called = PerformanceNoIndependentAwaitInLoop.calledName(call),
                        callees.makesOrderedCall(called.name)
                    {
                        return true
                    }
                }
                let nextInsideCatch = insideCatch || node.is(CatchClauseSyntax.self)
                return node.children(viewMode: .sourceAccurate).contains { visit($0, insideCatch: nextInsideCatch) }
            }
            return visit(Syntax(body), insideCatch: false)
        }

        // MARK: Loop-carried state

        static let consumingMethods: Set<String> = [
            "pop", "popFirst", "popLast", "removeFirst", "removeLast", "removeAll", "next", "read", "sort", "reverse",
            "shuffle", "swapAt", "partition",
        ]
        static let sinkMethods: Set<String> = [
            "append", "insert", "updateValue", "removeValue", "merge", "formUnion", "add", "set",
        ]

        /* A method that fills its receiver: one of the names above, or `add`, `append` or `insert` and a capitalized noun. */
        static func isSinkMethod(_ method: String) -> Bool {
            if sinkMethods.contains(method) {
                return true
            }
            return ["add", "append", "insert"].contains { verb in
                guard method.count > verb.count, method.hasPrefix(verb) else { return false }
                return method[method.index(method.startIndex, offsetBy: verb.count)].isUppercase
            }
        }

        static let comparisons: Set<String> = ["==", "!=", "<=", ">=", "===", "!=="]

        func carriesState() -> Bool {
            var carried = false
            var sinks: [[String]] = []
            var sinkRoots: Set<SyntaxIdentifier> = []
            var reads: [(root: SyntaxIdentifier, segments: [String])] = []

            /* A write into a target: silence unless it fills a collection that is a local of the function. */
            func write(into target: ExprSyntax, isPlainAssignment: Bool) {
                let target = PerformanceNoIndependentAwaitInLoop.stripped(target)
                if let reference = target.as(DeclReferenceExprSyntax.self) {
                    switch resolve(reference) {
                        case .item, .inner: return
                        case .functionLocal, .shared: carried = true
                    }
                    return
                }
                let filled: ExprSyntax
                if let subscriptCall = target.as(SubscriptCallExprSyntax.self) {
                    filled = subscriptCall.calledExpression
                }
                else if target.is(MemberAccessExprSyntax.self) {
                    filled = target
                }
                else {
                    /* A tuple or anything else written to: read each part as a whole write. */
                    if let tuple = target.as(TupleExprSyntax.self) {
                        for element in tuple.elements {
                            write(into: element.expression, isPlainAssignment: isPlainAssignment)
                        }
                    }
                    else if !target.is(DiscardAssignmentExprSyntax.self) {
                        carried = true
                    }
                    return
                }
                /*
                 `make().value = x` or `items[index].field = x` has no plain path, so it is judged by what it starts
                 from. A value type's element written in place writes the collection, so an indexed element is not
                 the item here.
                 */
                let path = PerformanceNoIndependentAwaitInLoop.chain(filled)
                sinkInto(
                    PerformanceNoIndependentAwaitInLoop.baseReference(filled),
                    segments: path?.segments,
                    isPlain: isPlainAssignment,
                )
            }

            func sinkInto(_ reference: DeclReferenceExprSyntax?, segments: [String]?, isPlain: Bool) {
                guard let reference else {
                    carried = true
                    return
                }
                switch resolve(reference) {
                    case .item, .inner:
                        return
                    case .shared:
                        carried = true
                    case .functionLocal:
                        guard isPlain, let segments else {
                            carried = true
                            return
                        }
                        sinks.append(segments)
                        sinkRoots.insert(reference.id)
                }
            }

            /* A call that fills, consumes or may change what the next pass reads. */
            func judge(_ call: FunctionCallExprSyntax) {
                let callee = PerformanceNoIndependentAwaitInLoop.stripped(call.calledExpression)
                if let bare = callee.as(DeclReferenceExprSyntax.self) {
                    /*
                     A bare lower-case call is a method on implicit `self` or a global function, and either may write
                     state every pass shares. An upper-case one constructs a value.
                     */
                    if bare.baseName.text.first?.isLowercase == true, case .shared = resolve(bare) {
                        carried = true
                    }
                    return
                }
                guard let member = callee.as(MemberAccessExprSyntax.self), let base = member.base,
                    let reference = PerformanceNoIndependentAwaitInLoop.baseReference(base)
                else { return }
                let method = PerformanceNoIndependentAwaitInLoop.plainName(member.declName.baseName)
                let receiverName = PerformanceNoIndependentAwaitInLoop.plainName(reference.baseName)
                let binding = resolve(reference)
                switch binding {
                    case .item, .inner:
                        return
                    case .functionLocal, .shared:
                        break
                }
                if Self.consumingMethods.contains(method) {
                    carried = true
                }
                else if Self.isSinkMethod(method) {
                    sinkInto(
                        reference,
                        segments: PerformanceNoIndependentAwaitInLoop.chain(base)?.segments,
                        isPlain: true,
                    )
                }
                else if receiverName == "self" || receiverName == "super" {
                    /* Any method of `self` may write the state every pass shares. */
                    carried = true
                }
                else if case .functionLocal(isVariable: true) = binding {
                    /* A method on a local `var` may be `mutating`, and the syntax cannot say it is not. */
                    carried = true
                }
            }

            func visit(_ node: Syntax) {
                if carried {
                    return
                }
                if let infix = node.as(InfixOperatorExprSyntax.self) {
                    if infix.operator.is(AssignmentExprSyntax.self) {
                        write(into: infix.leftOperand, isPlainAssignment: true)
                    }
                    else if let operation = infix.operator.as(BinaryOperatorExprSyntax.self),
                        operation.operator.text.hasSuffix("="),
                        !Self.comparisons.contains(operation.operator.text)
                    {
                        write(into: infix.leftOperand, isPlainAssignment: false)
                    }
                }
                else if let inOut = node.as(InOutExprSyntax.self) {
                    write(into: inOut.expression, isPlainAssignment: false)
                }
                else if let call = node.as(FunctionCallExprSyntax.self) {
                    judge(call)
                    if carried {
                        return
                    }
                }
                else if let reference = node.as(DeclReferenceExprSyntax.self),
                    !PerformanceNoIndependentAwaitInLoop.isMemberName(reference)
                {
                    if case .functionLocal = resolve(reference) {
                        reads.append((reference.id, PerformanceNoIndependentAwaitInLoop.readSegments(reference)))
                    }
                }
                for child in node.children(viewMode: .sourceAccurate) {
                    visit(child)
                }
            }
            visit(Syntax(body))
            if carried {
                return true
            }
            guard !sinks.isEmpty else { return false }

            /* The loop reads its sequence and its `where` clause each pass, so filling either is a work queue. */
            var headerReads: [[String]] = []
            for header in [Syntax(loop.sequence), loop.whereClause.map { Syntax($0) }].compactMap({ $0 }) {
                for token in header.tokens(viewMode: .sourceAccurate) {
                    if let reference = token.parent?.as(DeclReferenceExprSyntax.self),
                        !PerformanceNoIndependentAwaitInLoop.isMemberName(reference)
                    {
                        headerReads.append(PerformanceNoIndependentAwaitInLoop.readSegments(reference))
                    }
                }
            }
            for sink in sinks {
                for read in reads
                where !sinkRoots.contains(read.root)
                    && PerformanceNoIndependentAwaitInLoop.pathsOverlap(sink, read.segments)
                {
                    return true
                }
                for read in headerReads where PerformanceNoIndependentAwaitInLoop.pathsOverlap(sink, read) {
                    return true
                }
            }
            return false
        }

        // MARK: Exits on an awaited value

        /* The names declared in the loop whose value derives from an await, to a fixpoint. */
        func taintedNames() -> Set<String> {
            var tainted: Set<String> = []
            for _ in 0..<8 {
                let before = tainted.count
                func visit(_ node: Syntax) {
                    if PerformanceNoIndependentAwaitInLoop.isWalkBoundary(node) {
                        return
                    }
                    if let binding = node.as(PatternBindingSyntax.self), let initializer = binding.initializer,
                        containsTaint(Syntax(initializer.value), tainted: tainted)
                    {
                        tainted.formUnion(PerformanceNoIndependentAwaitInLoop.boundNames(binding.pattern))
                    }
                    else if let infix = node.as(InfixOperatorExprSyntax.self),
                        infix.operator.is(AssignmentExprSyntax.self) || isCompoundAssignment(infix),
                        containsTaint(Syntax(infix.rightOperand), tainted: tainted)
                    {
                        for token in infix.leftOperand.tokens(viewMode: .sourceAccurate) {
                            if let reference = token.parent?.as(DeclReferenceExprSyntax.self),
                                !PerformanceNoIndependentAwaitInLoop.isMemberName(reference)
                            {
                                tainted.insert(PerformanceNoIndependentAwaitInLoop.plainName(reference.baseName))
                            }
                        }
                    }
                    else if let binding = node.as(OptionalBindingConditionSyntax.self) {
                        let source = binding.initializer.map { Syntax($0.value) } ?? Syntax(binding.pattern)
                        if containsTaint(source, tainted: tainted) {
                            tainted.formUnion(PerformanceNoIndependentAwaitInLoop.boundNames(binding.pattern))
                        }
                    }
                    else if let matching = node.as(MatchingPatternConditionSyntax.self),
                        containsTaint(Syntax(matching.initializer.value), tainted: tainted)
                    {
                        tainted.formUnion(PerformanceNoIndependentAwaitInLoop.boundNames(matching.pattern))
                    }
                    else if let catchClause = node.as(CatchClauseSyntax.self),
                        let doStatement = catchClause.parent?.parent?.as(DoStmtSyntax.self),
                        containsAwait(Syntax(doStatement.body), before: nil)
                    {
                        if catchClause.catchItems.isEmpty {
                            tainted.insert("error")
                        }
                        for item in catchClause.catchItems {
                            if let pattern = item.pattern {
                                tainted.formUnion(PerformanceNoIndependentAwaitInLoop.boundNames(pattern))
                            }
                        }
                    }
                    else if let nested = node.as(ForStmtSyntax.self),
                        containsTaint(Syntax(nested.sequence), tainted: tainted)
                    {
                        tainted.formUnion(PerformanceNoIndependentAwaitInLoop.boundNames(nested.pattern))
                    }
                    else if let switchExpression = node.as(SwitchExprSyntax.self),
                        containsTaint(Syntax(switchExpression.subject), tainted: tainted)
                    {
                        for element in switchExpression.cases {
                            if case let .switchCase(switchCase) = element, case let .case(label) = switchCase.label {
                                for item in label.caseItems {
                                    tainted.formUnion(PerformanceNoIndependentAwaitInLoop.boundNames(item.pattern))
                                }
                            }
                        }
                    }
                    for child in node.children(viewMode: .sourceAccurate) {
                        visit(child)
                    }
                }
                visit(Syntax(body))
                if tainted.count == before {
                    break
                }
            }
            return tainted
        }

        func isCompoundAssignment(_ infix: InfixOperatorExprSyntax) -> Bool {
            guard let operation = infix.operator.as(BinaryOperatorExprSyntax.self) else { return false }
            return operation.operator.text.hasSuffix("=") && !Self.comparisons.contains(operation.operator.text)
        }

        /* Whether an expression awaits, or reads a name derived from an await. */
        func containsTaint(_ node: Syntax, tainted: Set<String>) -> Bool {
            if PerformanceNoIndependentAwaitInLoop.isWalkBoundary(node) {
                return false
            }
            if node.is(AwaitExprSyntax.self) {
                return true
            }
            if let reference = node.as(DeclReferenceExprSyntax.self),
                !PerformanceNoIndependentAwaitInLoop.isMemberName(reference),
                tainted.contains(PerformanceNoIndependentAwaitInLoop.plainName(reference.baseName))
            {
                return true
            }
            return node.children(viewMode: .sourceAccurate).contains { containsTaint($0, tainted: tainted) }
        }

        /* A `break`, `continue`, `return`, `throw` or uncaught `try` that leaves the loop on an awaited value. */
        func exitsOnAwaitedValue(tainted: Set<String>) -> Bool {
            func visit(_ node: Syntax) -> Bool {
                if PerformanceNoIndependentAwaitInLoop.isWalkBoundary(node) {
                    return false
                }
                if isJump(node), exitsLoop(node),
                    isGuardedByTaint(node, limit: Syntax(body), withSiblings: true, tainted: tainted)
                {
                    return true
                }
                return node.children(viewMode: .sourceAccurate).contains { visit($0) }
            }
            return visit(Syntax(body))
        }

        func isJump(_ node: Syntax) -> Bool {
            if let tryExpression = node.as(TryExprSyntax.self) {
                return tryExpression.questionOrExclamationMark == nil
            }
            return node.is(BreakStmtSyntax.self) || node.is(ContinueStmtSyntax.self) || node.is(ReturnStmtSyntax.self)
                || node.is(ThrowStmtSyntax.self)
        }

        /* A `do` whose catches handle every error: a bare `catch`, or one that only binds it. */
        static func catchesEverything(_ doStatement: DoStmtSyntax) -> Bool {
            doStatement.catchClauses.contains { clause in
                if clause.catchItems.isEmpty {
                    return true
                }
                return clause.catchItems.contains { item in
                    guard item.whereClause == nil, let pattern = item.pattern else { return item.whereClause == nil }
                    return pattern.as(ValueBindingPatternSyntax.self)?.pattern.is(IdentifierPatternSyntax.self) == true
                        || pattern.is(IdentifierPatternSyntax.self)
                }
            }
        }

        /* Whether a jump leaves this loop rather than a construct inside it. */
        func exitsLoop(_ jump: Syntax) -> Bool {
            if jump.is(ReturnStmtSyntax.self) {
                return true
            }
            if jump.is(ThrowStmtSyntax.self) || jump.is(TryExprSyntax.self) {
                /* Handled inside the body, a throw is a jump to the catch, not out of the loop. */
                var child = jump
                var current = jump.parent
                while let parent = current, child.id != body.id {
                    if let doStatement = parent.as(DoStmtSyntax.self), child.id == doStatement.body.id,
                        Self.catchesEverything(doStatement)
                    {
                        return false
                    }
                    child = parent
                    current = parent.parent
                }
                return true
            }
            let isBreak = jump.is(BreakStmtSyntax.self)
            let jumpLabel = (jump.as(BreakStmtSyntax.self)?.label ?? jump.as(ContinueStmtSyntax.self)?.label).map {
                PerformanceNoIndependentAwaitInLoop.plainName($0)
            }
            guard let jumpLabel else {
                /* The nearest loop, or for `break` the nearest `switch`, is the target. */
                var current = jump.parent
                while let parent = current {
                    if parent.id == loop.id {
                        return isBreak
                    }
                    if parent.is(ForStmtSyntax.self) || parent.is(WhileStmtSyntax.self)
                        || parent.is(RepeatStmtSyntax.self)
                        || (isBreak && parent.is(SwitchExprSyntax.self))
                    {
                        return false
                    }
                    current = parent.parent
                }
                return false
            }
            var current = jump.parent
            while let parent = current, parent.id != loop.id {
                if let labeled = parent.as(LabeledStmtSyntax.self),
                    PerformanceNoIndependentAwaitInLoop.plainName(labeled.label) == jumpLabel
                {
                    return false
                }
                current = parent.parent
            }
            if label == jumpLabel {
                return isBreak
            }
            /* A label outside this loop: the jump leaves it for an outer one. */
            return true
        }

        /*
         Climbs from a jump to `limit`, asking at each step whether what decides that the jump runs reads an
         awaited value. `withSiblings` also asks whether an earlier statement jumps on such a value.
         */
        func isGuardedByTaint(_ jump: Syntax, limit: Syntax, withSiblings: Bool, tainted: Set<String>) -> Bool {
            if let returnStatement = jump.as(ReturnStmtSyntax.self), let expression = returnStatement.expression,
                containsTaint(Syntax(expression), tainted: tainted)
            {
                return true
            }
            if let throwStatement = jump.as(ThrowStmtSyntax.self),
                containsTaint(Syntax(throwStatement.expression), tainted: tainted)
            {
                return true
            }
            if let tryExpression = jump.as(TryExprSyntax.self),
                containsTaint(Syntax(tryExpression.expression), tainted: tainted)
            {
                return true
            }
            var child = jump
            var current = jump.parent
            while let parent = current, child.id != limit.id {
                if let ifExpression = parent.as(IfExprSyntax.self),
                    containsTaint(Syntax(ifExpression.conditions), tainted: tainted)
                {
                    return true
                }
                if let guardStatement = parent.as(GuardStmtSyntax.self), child.id == guardStatement.body.id,
                    containsTaint(Syntax(guardStatement.conditions), tainted: tainted)
                {
                    return true
                }
                if let whileStatement = parent.as(WhileStmtSyntax.self),
                    containsTaint(Syntax(whileStatement.conditions), tainted: tainted)
                {
                    return true
                }
                if let repeatStatement = parent.as(RepeatStmtSyntax.self),
                    containsTaint(Syntax(repeatStatement.condition), tainted: tainted)
                {
                    return true
                }
                if let nested = parent.as(ForStmtSyntax.self), parent.id != loop.id {
                    if containsTaint(Syntax(nested.sequence), tainted: tainted)
                        || (nested.whereClause.map { containsTaint(Syntax($0), tainted: tainted) } ?? false)
                    {
                        return true
                    }
                }
                if parent.is(SwitchCaseSyntax.self),
                    let switchExpression = parent.parent?.parent?.as(SwitchExprSyntax.self)
                {
                    if containsTaint(Syntax(switchExpression.subject), tainted: tainted) {
                        return true
                    }
                    for element in switchExpression.cases {
                        if case let .switchCase(switchCase) = element, case let .case(label) = switchCase.label,
                            containsTaint(Syntax(label), tainted: tainted)
                        {
                            return true
                        }
                    }
                }
                if parent.is(CatchClauseSyntax.self), let doStatement = parent.parent?.parent?.as(DoStmtSyntax.self),
                    containsAwait(Syntax(doStatement.body), before: nil)
                {
                    return true
                }
                /* `do { return try await probe() } catch {}` leaves only when the await succeeded. */
                if let doStatement = parent.as(DoStmtSyntax.self), child.id == doStatement.body.id,
                    !doStatement.catchClauses.isEmpty,
                    containsAwait(Syntax(doStatement.body), before: jump.position)
                {
                    return true
                }
                if withSiblings, let list = parent.as(CodeBlockItemListSyntax.self),
                    earlierSiblingJumpsOnTaint(list, before: child, tainted: tainted)
                {
                    return true
                }
                child = parent
                current = parent.parent
            }
            return false
        }

        /* Conservative: any earlier tainted jump counts, even one whose target lies inside that statement. */
        func earlierSiblingJumpsOnTaint(
            _ list: CodeBlockItemListSyntax,
            before child: Syntax,
            tainted: Set<String>,
        ) -> Bool {
            for sibling in list where sibling.position < child.position {
                func visit(_ node: Syntax) -> Bool {
                    if PerformanceNoIndependentAwaitInLoop.isWalkBoundary(node) {
                        return false
                    }
                    let isStatementJump =
                        node.is(BreakStmtSyntax.self) || node.is(ContinueStmtSyntax.self)
                        || node.is(ReturnStmtSyntax.self) || node.is(ThrowStmtSyntax.self)
                    if isStatementJump,
                        isGuardedByTaint(node, limit: Syntax(sibling), withSiblings: false, tainted: tainted)
                    {
                        return true
                    }
                    return node.children(viewMode: .sourceAccurate).contains { visit($0) }
                }
                if visit(Syntax(sibling)) {
                    return true
                }
            }
            return false
        }
    }
}
