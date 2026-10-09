import SwiftSyntax

/*
 No pause written out by hand where the house has one by name: `try? await Task.sleep(for: .seconds(1))`, or
 `do { try await Task.sleep(for: interval) } catch {}`. Each says "wait this long, and stop waiting if the task is
 cancelled", which is exactly `Task.sleepUnlessCancelled`, the primitive both AhraOS projects declare once in
 `Task+SleepUnlessCancelled.swift`. `Task.sleep` throws only `CancellationError`, so the `try?` or the empty `catch`
 does nothing but end the pause early on cancellation, and spelled out at every call it makes the reader stop and
 check each one for the variant it might be (a value read from the `try?`, a `return` in the catch) before reading
 past it. One named pause says it is a plain pause, and the primitive is the one place that says why the
 cancellation is let go.

 It ports `nexus/consistency-no-hand-rolled-delay`, which reports `new Promise((resolve) => setTimeout(resolve, ms))`
 in favour of `await delay(ms)`. Each of its conditions maps to a Swift one, and Swift needs a different exactness in
 these places:
 - The act being compared is not the same act. TypeScript's `delay` is a timer nothing can cut short, so its
   hand-rolled form is a timer nothing can cut short. The Swift primitive ends early when the task is cancelled, so
   its hand-rolled form is a `Task.sleep` whose cancellation is swallowed: a discarded `try?`, or a `do` holding only
   the sleep with a bare `catch` holding no statement (a comment is not one). `try await Task.sleep` passes the
   cancellation up, a different act, and is not flagged, nor is a timer nothing can cancel
   (`withCheckedContinuation` around `DispatchQueue.main.asyncAfter`), which is the TypeScript shape's own
   semantics but would change behavior if rewritten to the primitive.
 - "Resolves with nothing" is the `try?`'s `()?` going nowhere: the `try?` is a whole statement that is no body's
   value, or the right side of `_ =`, read with `correctness-no-discarded-try-optional`'s own tests for those positions. The
   TypeScript rule also flags a delay promise held in a variable, because a promise resolved with `undefined`
   carries nothing; a `try?`'s `()?` carries whether the pause was cut short, so one that is read
   (`if (try? await Task.sleep(for: x)) == nil { return }`) is a signal, not a pause, and is not flagged.
 - The timer, called with exactly its duration (`setTimeout(resolve, ms)`), is `Task.sleep` (or
   `_Concurrency.Task.sleep`, as `globalThis.setTimeout` is the same timer reached through its owner) called with
   exactly one argument, labeled `for:` or `nanoseconds:`. The primitive passes nothing else, so a `tolerance:` or a
   `clock:` is a pause the primitive does not take, as a third `setTimeout` argument is a value `delay` does not
   pass. `until:` takes a deadline, not a duration, and has no duration to hand the primitive, as a one-argument
   `setTimeout` has none to hand `delay`. Parentheses around the awaited call, or around the `await`, are read
   through.
 - The definition of the primitive is not a use of it, decided the TypeScript rule's way, with no path or name in
   the rule: the shape is the whole body of a function or closure (its only statement) and its duration is a bare
   reference to one of that function's own parameters, by the name the body uses (`for duration: Duration` is
   `duration`, `nanoseconds: UInt64` is `nanoseconds`, a closure's `$0` when it declares none). Both proving
   grounds' `sleepUnlessCancelled` are this shape. A fixed duration, a computed one (`.seconds(seconds)`), another
   statement in the body, or the outer function's parameter makes it a use, and it is flagged.

 Where the message cannot be exact: the primitive is the project's own, not a library's, so its argument differs by
 project (`nanoseconds:` in macOS, `for:` in Presence), and the message names both rather than guess, leaving a
 conversion to the author where the site wrote the other. A target that cannot see its project's primitive (an
 executable or iOS target beside the one that declares it, as Presence's `CapturePhone`) is flagged all the same:
 the pause is still the primitive written by hand, and the repair starts by making the one primitive reachable.

 Beside `correctness-no-discarded-try-optional`: that rule flags the `try?` form too, as an error thrown away, and its repair,
 a `do`/`catch` with the reason in the catch, applied to a sleep writes this rule's second shape, which it does not
 flag. This rule names the house's pause for both, and its message says the `do`/`catch` is not the repair.

 What it accepts as safe, and why: `try await Task.sleep` (cancellation passed up, as Presence's `WindowClimb`
 does); a `do`/`catch` whose catch does something (macOS's `DelayWidthUntilIdle` returns, so a cancelled debounce
 does not commit); a `do` holding more than the sleep, whose catch also catches the rest; a `try?` whose value is
 read; `try! await Task.sleep`, which `force-try` owns; the primitive itself; and any other clock or timer.

 Known misses, every one a finding not made and never one invented:
 - A `try?` sleep that is the only statement of a closure (`Task { try? await Task.sleep(for: x) }`) is the
   closure's value, which only types could show to be discarded, and is not read. A pause alone in a closure does
   nothing to wait for, and neither proving ground has one.
 - `catch is CancellationError {}`, `catch _ {}` and `catch let error {}`: the same pause with a pattern, which
   the rule leaves to the reader rather than judge every pattern.
 - `Task<Never, Never>.sleep(...)`, a `Task.sleep` reached any other way, `ContinuousClock().sleep(for:)` and a
   `Clock`'s `sleep`, and `Task.sleep(for: x, clock: .continuous)`, the default written out.
 - A second definition of the primitive elsewhere (`func pause(for duration: Duration) async { try? await
   Task.sleep(for: duration) }`), which is a duplicate helper rather than a hand-rolled pause, as the TypeScript
   rule leaves `function sleep(ms)` alone.
 - `Task` is read as Swift's: a type of ours named `Task` with a throwing static `sleep` would be flagged. Neither
   proving ground declares one, and the TypeScript rule reads `Promise` the same way.
 */
public struct ConsistencyNoHandRolledDelay: FileRule {
    public let name = "cohere-swift/consistency-no-hand-rolled-delay"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Every shape this rule flags names `Task` and `sleep`, however the call is spread across lines. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("Task") && file.source.contains("sleep")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { found in
            file.finding(
                at: found.node,
                rule: name,
                message: found.isDoCatch
                    ? RuleMessages.ConsistencyNoHandRolledDelay.handRolledDelayDoCatch()
                    : RuleMessages.ConsistencyNoHandRolledDelay.handRolledDelay(),
            )
        }
    }

    /* Collects each hand-rolled pause: the `try?` expression, or the `do` statement with its catch. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [(node: Syntax, isDoCatch: Bool)] = []

        /* The two labels that hand `Task.sleep` a duration, which the primitive takes one of. */
        static let durationLabels: Set<String> = ["for", "nanoseconds"]

        override func visit(_ node: TryExprSyntax) -> SyntaxVisitorContinueKind {
            guard node.questionOrExclamationMark?.tokenKind == .postfixQuestionMark,
                let call = Self.sleepCall(awaitedBy: node.expression)
            else {
                return .visitChildren
            }
            let item: CodeBlockItemSyntax?
            if CorrectnessNoDiscardedTryOptional.Visitor.isStatement(node) {
                item = node.parent?.as(CodeBlockItemSyntax.self)
            }
            else if CorrectnessNoDiscardedTryOptional.Visitor.isDiscardAssignment(node) {
                /* `_ = try? ...` is the sequence `_`, `=`, `try? ...`, whose list sits in a sequence expression that is the statement. */
                item = node.parent?.parent?.parent?.as(CodeBlockItemSyntax.self)
            }
            else {
                return .visitChildren
            }
            if !Self.isDefinition(item, duration: call) {
                found.append((Syntax(node), false))
            }
            return .visitChildren
        }

        override func visit(_ node: DoStmtSyntax) -> SyntaxVisitorContinueKind {
            guard node.throwsClause == nil, node.body.statements.count == 1,
                let attempt = node.body.statements.first?.item.as(TryExprSyntax.self),
                attempt.questionOrExclamationMark == nil,
                let call = Self.sleepCall(awaitedBy: attempt.expression)
            else {
                return .visitChildren
            }
            guard node.catchClauses.count == 1, let catchClause = node.catchClauses.first,
                catchClause.catchItems.isEmpty, catchClause.body.statements.isEmpty
            else {
                return .visitChildren
            }
            /* A labeled `do` is the label's statement. */
            let statement = node.parent?.as(LabeledStmtSyntax.self).map(Syntax.init) ?? Syntax(node)
            if !Self.isDefinition(statement.parent?.as(CodeBlockItemSyntax.self), duration: call) {
                found.append((Syntax(node), true))
            }
            return .visitChildren
        }

        /* `await Task.sleep(for: d)` or `await Task.sleep(nanoseconds: n)`, the call with its one duration argument. */
        static func sleepCall(awaitedBy expression: ExprSyntax) -> FunctionCallExprSyntax? {
            guard let awaited = unparenthesized(expression).as(AwaitExprSyntax.self),
                let call = unparenthesized(awaited.expression).as(FunctionCallExprSyntax.self)
            else {
                return nil
            }
            guard call.trailingClosure == nil, call.additionalTrailingClosures.isEmpty, call.arguments.count == 1,
                let label = call.arguments.first?.label?.text, durationLabels.contains(label)
            else {
                return nil
            }
            guard let callee = call.calledExpression.as(MemberAccessExprSyntax.self),
                callee.declName.baseName.text == "sleep", callee.declName.argumentNames == nil, let base = callee.base
            else {
                return nil
            }
            return isTask(base) ? call : nil
        }

        /* `Task`, or `_Concurrency.Task`: the concurrency task type by its name, with no generic arguments written. */
        static func isTask(_ expression: ExprSyntax) -> Bool {
            if let reference = expression.as(DeclReferenceExprSyntax.self) {
                return reference.baseName.text == "Task" && reference.argumentNames == nil
            }
            guard let member = expression.as(MemberAccessExprSyntax.self), member.declName.baseName.text == "Task",
                member.declName.argumentNames == nil
            else {
                return false
            }
            return member.base?.as(DeclReferenceExprSyntax.self)?.baseName.text == "_Concurrency"
        }

        /* An expression with any parentheses around it taken off: a one-element tuple with no label is a parenthesized expression. */
        static func unparenthesized(_ expression: ExprSyntax) -> ExprSyntax {
            guard let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let element = tuple.elements.first, element.label == nil
            else {
                return expression
            }
            return unparenthesized(element.expression)
        }

        /*
         Whether the pause is the definition of a pause primitive: the only statement of a function's or closure's
         body, with a duration that is a bare reference to one of that function's own parameters.
         */
        static func isDefinition(_ item: CodeBlockItemSyntax?, duration call: FunctionCallExprSyntax) -> Bool {
            guard let item, let list = item.parent?.as(CodeBlockItemListSyntax.self), list.count == 1 else {
                return false
            }
            guard let duration = call.arguments.first?.expression.as(DeclReferenceExprSyntax.self),
                duration.argumentNames == nil
            else { return false }
            let durationName = duration.baseName.text
            if let function = list.parent?.as(CodeBlockSyntax.self)?.parent?.as(FunctionDeclSyntax.self) {
                return function.signature.parameterClause.parameters.contains {
                    ($0.secondName ?? $0.firstName).text == durationName
                }
            }
            guard let closure = list.parent?.as(ClosureExprSyntax.self) else { return false }
            switch closure.signature?.parameterClause {
                case .simpleInput(let parameters)?:
                    return parameters.contains { $0.name.text == durationName }
                case .parameterClause(let clause)?:
                    return clause.parameters.contains { ($0.secondName ?? $0.firstName).text == durationName }
                case nil:
                    /* A closure that names no parameters reaches them as `$0`, `$1`. */
                    if case .dollarIdentifier = duration.baseName.tokenKind {
                        return true
                    }
                    return false
            }
        }
    }
}
