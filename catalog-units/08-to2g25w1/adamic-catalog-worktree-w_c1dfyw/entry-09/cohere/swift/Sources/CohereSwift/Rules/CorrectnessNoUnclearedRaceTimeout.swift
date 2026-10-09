import SwiftSyntax

/*
 No timeout raced in a task group that nothing cancels: a `withTaskGroup` or `withThrowingTaskGroup` that adds a
 child doing the work and a child that only sleeps and then throws or returns, takes the first result with
 `group.next()`, and never calls `group.cancelAll()`. A task group always waits for every child before it returns,
 so when the work wins, the caller still sits out the whole timeout before it sees the result. When the timeout
 child returns a value instead of throwing and wins, the group waits for the work as well, so the timeout bounds
 nothing at all. Measured with swiftc 6.4 on 2026-10-03: work of 100 milliseconds raced against a 2 second sleep
 returned after 2.1 seconds without `cancelAll` and after 0.1 seconds with it; a 100 millisecond timeout that
 returns `nil` raced against 2 seconds of work returned after 2.07 seconds.

 It ports `nexus/correctness-no-uncleared-race-timeout`, which reports a `setTimeout` armed in a `Promise.race`
 executor whose handle nothing can reach, so it is never cleared when the other side wins. The race is the task
 group, the timer is the sleeping child, and the handle that clears it is the group's `cancelAll()`. Swift needs a
 different exactness in these places:
 - Returning is not clearing. A `Promise.race` that settles lets its caller go on while the timer lingers; a task
   group's body that returns, from any depth (`if let first = try await group.next() { return first }`), still
   waits for the sleeping child, which was measured above. So the Swift bug holds the caller, not only the
   process, and only `cancelAll()` clears it.
 - Throwing is clearing. A throwing group whose body throws cancels every child that is left (measured: a body
   that throws after the work wins returns in 0.1 seconds). So a body that cannot end normally once the result is
   read is not flagged: it has no `return` of its own after the read, and its last statement throws (or calls
   `fatalError` or `preconditionFailure`) on every branch of an `if`/`else`, `switch` or `do`/`catch`. TypeScript
   has no counterpart: a rejected race leaves its timer armed all the same.
 - The race is the shape that takes one first result: exactly one `next()` (or `nextResult()`) call on the group,
   not inside a loop of the body, with at least two children added before it in source order. TypeScript's race is
   `Promise.race` over an array literal. A loop over `next()` or `for await` over the group collects or drains,
   and a second `next()` reads past the first result; neither is read.
 - The timer is a child, added with `addTask` or `addTaskUnlessCancelled` (by trailing closure or `operation:`),
   whose body is exactly two statements: an awaited `Task.sleep` (any arguments, under `try`, `try?` or neither),
   then a `throw` or a `return` (with or without a value) holding no `await`. TypeScript reads any `setTimeout`
   in the executor whatever its callback does; a Swift child that sleeps and then works (a hedged second request)
   is not a timeout, and a child that only sleeps is also how a minimum duration is written, so neither is read.
 - The handle is lost when nothing but `addTask`, `addTaskUnlessCancelled` and `next` or `nextResult` touches the
   group in the body. Every other use is a way the group could be cancelled and stays silent without the rule
   following it, as every other use of a TypeScript handle does: `cancelAll()` anywhere (a `defer` at the top is
   the cleanest repair), `waitForAll()`, `isEmpty`, the group handed to a function, or named inside a nested
   closure or function. A nested closure or function that declares its own parameter of the group's name (a child
   running a task group of its own, also called `group`) is another binding, and is skipped.
 - TypeScript resolves `Promise.race` and `setTimeout` with the checker. This rule reads names: `withTaskGroup`,
   `withThrowingTaskGroup` (also through `_Concurrency.`) and `Task.sleep` are read as Swift's. A function of ours
   with one of those names would be read as the library's; neither proving ground declares one.

 The finding is the timeout child's `addTask` call, as the TypeScript finding is the `setTimeout` call. Two
 sleeping children racing each other are both reported, as two racing TypeScript timers are.

 What it accepts as safe, and why: a group that calls `cancelAll()` anywhere in its body; a group whose body throws
 on every path after the read; `for await` over the group and `next()` in a loop (they collect, they do not race);
 `withDiscardingTaskGroup`, which has no `next()`; a group with one child before the read; and every use of the
 group the rule does not follow.

 Known misses, every one a finding not made and never one invented:
 - `cancelAll()` on only some paths after the read (`if let first { group.cancelAll(); return first }; return
   nil`, where the `nil` path waits): any `cancelAll` silences the group, as any read of a TypeScript handle does.
 - A race read through `for try await result in group { return result }`, or through `while let`.
 - A timeout from a helper (`group.addTask { try await timeout(after: limit) }`, or `addTask(operation: timeout)`),
   which is not followed, as a TypeScript `timeoutAfter(ms)` is not.
 - A sleep on a clock (`try await ContinuousClock().sleep(for: limit)`, `clock.sleep(until:)`), a sleep written
   `_ = try? await Task.sleep(...)`, and a timeout child that does anything else besides.
 - Children added in a nested closure or function (`func add() { group.addTask { ... } }`), which names the group
   inside a nested function and so silences it.
 - A race on a group of our own type, on a group from a function of ours (`withCancellingTaskGroup`, which may
   cancel for itself), or on a group passed into a function as a parameter rather than created at the site.
 - A body that binds the group's name again anywhere (`let group = spare`, `if let group`), which is read as
   another group and silences the race.

 Beside this rule, not part of it: `cancelAll()` only asks. A child that ignores cancellation, such as
 `withCheckedContinuation` around a callback, still holds the group until it finishes, so a timeout raced against
 it bounds nothing even with `cancelAll()` (measured: 2.1 seconds for a 0.1 second timeout against a callback
 answering in 2). That is a different shape, and the rule does not read the work child.
 */
public struct CorrectnessNoUnclearedRaceTimeout: FileRule {
    public let name = "cohere-swift/correctness-no-uncleared-race-timeout"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Every shape this rule flags names a `...TaskGroup` function and `Task.sleep`, however they are spread across lines. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("TaskGroup") && file.source.contains("sleep")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { found in
            file.finding(
                at: found.child,
                rule: name,
                message: RuleMessages.CorrectnessNoUnclearedRaceTimeout.unclearedRaceTimeout(group: found.group),
            )
        }
    }

    /* Collects every lost timeout child, with the name its group goes by. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [(child: FunctionCallExprSyntax, group: String)] = []

        static let groupFunctions: Set<String> = ["withTaskGroup", "withThrowingTaskGroup"]
        static let childMethods: Set<String> = ["addTask", "addTaskUnlessCancelled"]
        static let readMethods: Set<String> = ["next", "nextResult"]

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            guard let body = Self.groupBody(node), let group = Self.groupName(body) else {
                return .visitChildren
            }
            for child in Self.lostTimeouts(in: body, group: group) {
                found.append((child, group))
            }
            return .visitChildren
        }

        /* The body closure of `withTaskGroup` or `withThrowingTaskGroup`, written as the trailing closure or as `body:`. */
        static func groupBody(_ call: FunctionCallExprSyntax) -> ClosureExprSyntax? {
            guard isGroupFunction(call.calledExpression), call.additionalTrailingClosures.isEmpty else { return nil }
            if let trailing = call.trailingClosure {
                return trailing
            }
            return call.arguments.first { $0.label?.text == "body" }?.expression.as(ClosureExprSyntax.self)
        }

        /* `withTaskGroup`, `withThrowingTaskGroup`, or either through `_Concurrency.`. */
        static func isGroupFunction(_ expression: ExprSyntax) -> Bool {
            if let reference = expression.as(DeclReferenceExprSyntax.self) {
                return groupFunctions.contains(reference.baseName.text) && reference.argumentNames == nil
            }
            guard let member = expression.as(MemberAccessExprSyntax.self),
                groupFunctions.contains(member.declName.baseName.text), member.declName.argumentNames == nil
            else {
                return false
            }
            return member.base?.as(DeclReferenceExprSyntax.self)?.baseName.text == "_Concurrency"
        }

        /* The name the body calls its group by: its one parameter, or `$0` when it declares none. `_` names nothing the body can race on. */
        static func groupName(_ body: ClosureExprSyntax) -> String? {
            let name: String
            switch body.signature?.parameterClause {
                case .simpleInput(let parameters)?:
                    guard parameters.count == 1, let parameter = parameters.first else { return nil }
                    name = parameter.name.text
                case .parameterClause(let clause)?:
                    guard clause.parameters.count == 1, let parameter = clause.parameters.first else { return nil }
                    name = (parameter.secondName ?? parameter.firstName).text
                case nil:
                    name = "$0"
            }
            return name == "_" ? nil : name
        }

        /*
         The timeout children of one group whose cancellation is lost: the group is touched only to add children and
         read one first result outside any loop, at least two children were added before that read, and the body can
         still return normally after it.
         */
        static func lostTimeouts(in body: ClosureExprSyntax, group: String) -> [FunctionCallExprSyntax] {
            let uses = GroupUses(group: group)
            uses.walk(body.statements)
            guard !uses.rebinds else { return [] }
            var children: [FunctionCallExprSyntax] = []
            var reads: [FunctionCallExprSyntax] = []
            for reference in uses.references {
                guard isOwn(Syntax(reference), of: body),
                    let member = reference.parent?.as(MemberAccessExprSyntax.self), member.base?.id == reference.id,
                    member.declName.argumentNames == nil, let call = member.parent?.as(FunctionCallExprSyntax.self),
                    call.calledExpression.id == member.id
                else {
                    return []
                }
                let method = member.declName.baseName.text
                if childMethods.contains(method) {
                    children.append(call)
                }
                else if readMethods.contains(method) {
                    reads.append(call)
                }
                else {
                    return []
                }
            }
            guard reads.count == 1, let read = reads.first, !isInLoop(Syntax(read), of: body) else { return [] }
            let racing = children.filter { $0.endPosition <= read.position }
            guard racing.count >= 2, canReturnNormally(body, after: read) else { return [] }
            return racing.filter(isTimeout)
        }

        /* Whether a node belongs to the body itself rather than to a closure, function, accessor or type nested in it. */
        static func isOwn(_ node: Syntax, of body: ClosureExprSyntax) -> Bool {
            var current = node.parent
            while let ancestor = current, ancestor.id != body.id {
                if isBoundary(ancestor) {
                    return false
                }
                current = ancestor.parent
            }
            return current != nil
        }

        /* A node whose contents run apart from the code around it: another closure, a function or accessor, or a type. */
        static func isBoundary(_ node: Syntax) -> Bool {
            node.is(ClosureExprSyntax.self) || node.is(FunctionDeclSyntax.self) || node.is(AccessorBlockSyntax.self)
                || node.is(InitializerDeclSyntax.self)
                || node.is(DeinitializerDeclSyntax.self) || node.is(SubscriptDeclSyntax.self)
                || node.asProtocol((any DeclGroupSyntax).self) != nil
        }

        /* Whether the read sits in a loop of the body, where it collects results rather than taking the first. */
        static func isInLoop(_ node: Syntax, of body: ClosureExprSyntax) -> Bool {
            var current = node.parent
            while let ancestor = current, ancestor.id != body.id {
                if ancestor.is(ForStmtSyntax.self) || ancestor.is(WhileStmtSyntax.self)
                    || ancestor.is(RepeatStmtSyntax.self)
                {
                    return true
                }
                current = ancestor.parent
            }
            return false
        }

        /*
         Whether the body can end normally once the first result is read: a `return` of its own that ends after the
         read, or a last statement that does not always throw. A body that throws cancels what is left on its way out.
         */
        static func canReturnNormally(_ body: ClosureExprSyntax, after read: FunctionCallExprSyntax) -> Bool {
            let returns = Returns(viewMode: .sourceAccurate)
            returns.walk(body.statements)
            if returns.found.contains(where: { $0.endPosition > read.position && isOwn(Syntax($0), of: body) }) {
                return true
            }
            return !endsAbruptly(body.statements)
        }

        /* The calls that never return, which end a body as surely as a throw. */
        static let stoppingCalls: Set<String> = ["fatalError", "preconditionFailure"]

        /*
         Whether a statement list can never run off its end: its last statement is a `throw` or a call that stops the
         program, or an `if`, `switch` or `do` every branch of which ends that way.
         */
        static func endsAbruptly(_ statements: CodeBlockItemListSyntax) -> Bool {
            guard let last = statements.last?.item else { return false }
            if last.is(ThrowStmtSyntax.self) {
                return true
            }
            if let statement = last.as(ExpressionStmtSyntax.self) {
                return endsAbruptly(statement.expression)
            }
            if let expression = last.as(ExprSyntax.self) {
                return endsAbruptly(expression)
            }
            if let doStatement = last.as(DoStmtSyntax.self) {
                return endsAbruptly(doStatement.body.statements)
                    && doStatement.catchClauses.allSatisfy { endsAbruptly($0.body.statements) }
            }
            return false
        }

        static func endsAbruptly(_ expression: ExprSyntax) -> Bool {
            if let call = expression.as(FunctionCallExprSyntax.self),
                let callee = call.calledExpression.as(DeclReferenceExprSyntax.self)
            {
                return stoppingCalls.contains(callee.baseName.text)
            }
            if let conditional = expression.as(IfExprSyntax.self) {
                switch conditional.elseBody {
                    case .ifExpr(let elseIf)?:
                        return endsAbruptly(conditional.body.statements) && endsAbruptly(ExprSyntax(elseIf))
                    case .codeBlock(let block)?:
                        return endsAbruptly(conditional.body.statements) && endsAbruptly(block.statements)
                    case nil:
                        return false
                }
            }
            if let choice = expression.as(SwitchExprSyntax.self) {
                return !choice.cases.isEmpty
                    && choice.cases.allSatisfy {
                        $0.as(SwitchCaseSyntax.self).map { endsAbruptly($0.statements) } ?? false
                    }
            }
            return false
        }

        /*
         A child that is only a timer: its body is an awaited `Task.sleep`, then a `throw` or a `return` that awaits
         nothing more.
         */
        static func isTimeout(_ child: FunctionCallExprSyntax) -> Bool {
            guard child.additionalTrailingClosures.isEmpty,
                let operation = child.trailingClosure
                    ?? child.arguments.first(where: { $0.label?.text == "operation" })?.expression.as(
                        ClosureExprSyntax.self
                    )
            else {
                return false
            }
            let statements = Array(operation.statements)
            guard statements.count == 2, let sleep = statements[0].item.as(ExprSyntax.self), isSleep(sleep) else {
                return false
            }
            if let throwing = statements[1].item.as(ThrowStmtSyntax.self) {
                return !awaits(Syntax(throwing.expression))
            }
            if let returning = statements[1].item.as(ReturnStmtSyntax.self) {
                return returning.expression.map { !awaits(Syntax($0)) } ?? true
            }
            return false
        }

        /* `try await Task.sleep(...)`, `try? await Task.sleep(...)` or `await Task.sleep(...)`, with any arguments. */
        static func isSleep(_ expression: ExprSyntax) -> Bool {
            var inner = ConsistencyNoHandRolledDelay.Visitor.unparenthesized(expression)
            if let attempt = inner.as(TryExprSyntax.self) {
                inner = ConsistencyNoHandRolledDelay.Visitor.unparenthesized(attempt.expression)
            }
            guard let awaited = inner.as(AwaitExprSyntax.self),
                let call = ConsistencyNoHandRolledDelay.Visitor.unparenthesized(awaited.expression).as(
                    FunctionCallExprSyntax.self
                ),
                call.trailingClosure == nil, call.additionalTrailingClosures.isEmpty
            else {
                return false
            }
            guard let callee = call.calledExpression.as(MemberAccessExprSyntax.self),
                callee.declName.baseName.text == "sleep", callee.declName.argumentNames == nil, let base = callee.base
            else {
                return false
            }
            return ConsistencyNoHandRolledDelay.Visitor.isTask(base)
        }

        static func awaits(_ node: Syntax) -> Bool {
            node.tokens(viewMode: .sourceAccurate).contains { $0.tokenKind == .keyword(.await) }
        }
    }

    /*
     Every reference to the group's name in its body, skipping a nested closure or function that declares its own
     parameter of that name, and whether the body binds the name again.
     */
    final class GroupUses: SyntaxVisitor {
        let group: String
        private(set) var references: [DeclReferenceExprSyntax] = []
        private(set) var rebinds = false

        init(group: String) {
            self.group = group
            super.init(viewMode: .sourceAccurate)
        }

        override func visit(_ node: DeclReferenceExprSyntax) -> SyntaxVisitorContinueKind {
            /* `other.group` names a member, not the binding. */
            if node.baseName.text == group, node.argumentNames == nil,
                node.parent?.as(MemberAccessExprSyntax.self)?.declName.id != node.id
            {
                references.append(node)
            }
            return .visitChildren
        }

        override func visit(_ node: IdentifierPatternSyntax) -> SyntaxVisitorContinueKind {
            if node.identifier.text == group {
                rebinds = true
            }
            return .visitChildren
        }

        override func visit(_ node: ClosureExprSyntax) -> SyntaxVisitorContinueKind {
            /* A nested closure never sees the body's `$0`: it has its own, or names its parameters and may not use one. */
            if group == "$0" {
                return .skipChildren
            }
            switch node.signature?.parameterClause {
                case .simpleInput(let parameters)?:
                    return parameters.contains { $0.name.text == group } ? .skipChildren : .visitChildren
                case .parameterClause(let clause)?:
                    return clause.parameters.contains { ($0.secondName ?? $0.firstName).text == group }
                        ? .skipChildren : .visitChildren
                case nil:
                    return .visitChildren
            }
        }

        override func visit(_ node: FunctionDeclSyntax) -> SyntaxVisitorContinueKind {
            node.signature.parameterClause.parameters.contains { ($0.secondName ?? $0.firstName).text == group }
                ? .skipChildren : .visitChildren
        }
    }

    /* Every `return` statement in a body, its own and nested ones alike; the caller keeps its own. */
    final class Returns: SyntaxVisitor {
        private(set) var found: [ReturnStmtSyntax] = []

        override func visit(_ node: ReturnStmtSyntax) -> SyntaxVisitorContinueKind {
            found.append(node)
            return .visitChildren
        }
    }
}
