import SwiftSyntax

/*
 No starting a task that can throw and then dropping it: `Task { try await save() }` as a statement. A task
 keeps the error its closure throws as its result, and only a reader of `value` or `result` ever sees it. A
 task nothing holds has no reader, so the error is gone without a trace, and the work fails silently. The
 repair is to catch the error inside the task with `do` and `catch`, or to keep the task and read its value
 with `try await`.

 The shapes are SwiftLint's `unhandled_throwing_task`: a call of `Task`, plain or specialised with `_` as its
 failure type (`Task<_, _>`, `Task<Void, _>`), whose closure holds a `try` (not `try?` or `try!`) or a `throw`
 that no catch-all `catch` (a bare `catch`, or `catch let error`) around it handles, including one in a
 catch-all's own body; and whose result is neither assigned (`let task =`, `self.task =`), nor read
 (`.value`, `.result`), nor returned (`return Task { }`, or the only statement of a function with a return
 type). A `Result { }` inside the closure handles what it holds. The finding starts where SwiftLint's does,
 at the `Task` the call names.

 SwiftLint judges whether the closure can throw by its spelling, and gets it wrong both ways. A `try` inside a
 nested closure or function counts for it, though that error never reaches the task, and a type of ours named
 `Task` is flagged: both false. A `catch` whose pattern is not a binding (`catch is CancellationError`, `catch
 Failure.timeout`) or that has a `where` clause counts as catching everything: both missed. This rule asks the
 compiler instead, reading which declaration the call resolved to from the index the build wrote. The
 concurrency `Task` (`s:ScT`) declares its initializer and its static `detached`, `immediate`,
 `immediateDetached` and `runDetached` twice, once where the failure type is `Never` and once where it is
 `any Error`, and the compiler picks the `any Error` one exactly when the closure can throw. That one answer
 covers all SwiftLint reads from the syntax (`try` inside an exhaustive `do`, `try?`, `try!`, `Result { }`,
 nested closures) and more it cannot: a typed `throws(Failure)` closure, `for try await`, a narrow `catch`, an
 operation passed by name (`Task(operation: work)`). A type of ours named `Task` resolves to its own module
 and is never flagged.

 Where we differ on purpose: `Task.detached`, `Task.immediate`, `Task.immediateDetached`, `Task.init` and the
 module-qualified `_Concurrency.Task` drop the error the same way and are found too; SwiftLint reads only
 `Task`. And this rule flags a task only where its value provably goes nowhere, as a statement. SwiftLint
 flags every task it does not see assigned, read or returned, so a task passed as an argument
 (`tasks.append(Task { })`), put in an array, or returned from a closure (`urls.map { url in Task { } }`) is
 flagged though something holds it; this rule does not. A statement drops its value except as the only
 statement of a body that returns it: a function with a return type other than `Void`, a getter, a closure,
 or a branch of an `if` or `switch` whose value is itself used. A statement of a result builder's body is
 collected, not dropped, and the builder calls the compiler records there say so; it is not flagged, where
 SwiftLint flags it. `_ = Task { }` is a written choice to drop the task, and is not flagged, as SwiftLint
 does not flag it. Nor is a task whose failure type is written as anything but `_` (`Task<Void, any Error>`),
 as SwiftLint does not: the writer chose the failure type by hand, and the declaration the compiler picked
 then says nothing about whether the closure throws.

 Misses, accepted: a task that is the only statement of a closure without a written `-> Void` is not flagged,
 though in a closure its caller types as returning nothing (`Button { Task { try await save() } }`,
 `.onAppear { }`) its value is dropped. The index does not say what the closure's caller does with its value,
 and `urls.map { url in Task { } }` reads the same and keeps every task. A task whose result is chained into
 anything but an assignment (`Task { }.cancel()`) is not read, and a typealias of `Task` is not followed. No
 sibling typed rule shares this shape.
 */
public struct UnhandledThrowingTask: TypedFileRule {
    public let name = "cohere-swift/unhandled-throwing-task"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "unhandled_throwing_task"

    public init() {}

    /* The concurrency `Task`'s extension where `Failure == any Error`: every way it creates a task that can throw. */
    static let throwingExtension = "s:ScT12_Concurrencys5Error_pRs_rlE"

    /* Its static members that start a task, as the symbol spells each name after the extension; its initializer is told by its `fc` suffix. */
    static let throwingStarters = ["8detached", "9immediate", "17immediateDetached", "11runDetached"]

    /* The member names a task is started by, `init` among them. */
    static let starterNames: Set<String> = ["init", "detached", "immediate", "immediateDetached", "runDetached"]

    /* A `Task` anywhere in the file: every shape names it. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("Task")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.filter { candidate in
            let starts = Self.references(at: candidate.starter, in: file, symbols: symbols).contains {
                Self.startsThrowingTask($0.symbol)
            }
            return starts && !Self.isBuilderComponent(candidate.call, in: file, symbols: symbols)
        }.map { candidate in
            file.finding(
                at: candidate.call.calledExpression,
                rule: name,
                messageId: "unhandledThrowingTask",
                message:
                    "This task's closure can throw, and nothing reads the task's result, so any error it throws is dropped without a trace. Catch the error inside the task with do and catch, or keep the task and read its value with try await.",
            )
        }
    }

    /* The concurrency `Task`'s initializer or static starter where the failure type is `any Error`. */
    static func startsThrowingTask(_ symbol: String) -> Bool {
        guard symbol.hasPrefix(throwingExtension) else { return false }
        let member = symbol.dropFirst(throwingExtension.count)
        return symbol.hasSuffix("fc") || throwingStarters.contains { member.hasPrefix($0) }
    }

    /* The written references the compiler recorded at this token. */
    static func references(
        at token: TokenSyntax,
        in file: ParsedFile,
        symbols: FileSymbols,
    ) -> [FileSymbols.Occurrence] {
        let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
        return symbols.occurrences(line: location.line, column: location.column).filter {
            $0.isReference && !$0.isImplicit
        }
    }

    /*
     Whether the statement is a component of a result builder's body, where a statement's value is not dropped
     but collected. The compiler records a builder's transform where it applies it: `buildExpression` at the
     statement's first token, `buildBlock` at the opening brace of each body it builds, and nothing else is
     ever recorded at a brace. So a reference at the call's first token that is not the `Task`'s own (or the
     `_Concurrency` module's, in `_Concurrency.Task`), or any reference at a brace between the call and its
     closure or function, marks a builder's body.
     */
    static func isBuilderComponent(_ call: FunctionCallExprSyntax, in file: ParsedFile, symbols: FileSymbols) -> Bool {
        if let first = call.firstToken(viewMode: .sourceAccurate),
            references(at: first, in: file, symbols: symbols).contains(where: {
                !$0.symbol.hasPrefix("s:ScT") && !$0.symbol.hasPrefix("c:@M@")
            })
        {
            return true
        }
        var node = call.parent
        while let current = node {
            var brace: TokenSyntax?
            var isBoundary = false
            if let closure = current.as(ClosureExprSyntax.self) {
                brace = closure.leftBrace
                isBoundary = true
            }
            else if let accessors = current.as(AccessorBlockSyntax.self) {
                brace = accessors.leftBrace
                isBoundary = true
            }
            else if let block = current.as(CodeBlockSyntax.self) {
                brace = block.leftBrace
                isBoundary =
                    block.parent.map {
                        $0.is(FunctionDeclSyntax.self) || $0.is(AccessorDeclSyntax.self)
                            || $0.is(InitializerDeclSyntax.self) || $0.is(DeinitializerDeclSyntax.self)
                    } ?? true
            }
            else if let switchExpression = current.as(SwitchExprSyntax.self) {
                brace = switchExpression.leftBrace
            }
            if let brace, !references(at: brace, in: file, symbols: symbols).isEmpty {
                return true
            }
            if isBoundary {
                return false
            }
            node = current.parent
        }
        return false
    }

    /* One task started as a statement: the call, and the name token the index places its declaration at. */
    struct Candidate {
        var call: FunctionCallExprSyntax
        var starter: TokenSyntax
    }

    /* Every call that starts a `Task` with an inferred failure type and drops its value. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Candidate] = []

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            if let starter = Self.starter(node.calledExpression), let item = node.parent?.as(CodeBlockItemSyntax.self),
                Self.isDropped(item)
            {
                found.append(Candidate(call: node, starter: starter))
            }
            return .visitChildren
        }

        /*
         The name token of `Task`, `_Concurrency.Task`, `Task.detached` (or `immediate`, `init`, any starter),
         each plain or specialised: the token the index records the initializer or the static member at. Nil
         for a specialisation whose failure type is written as anything but `_`.
         */
        static func starter(_ called: ExprSyntax) -> TokenSyntax? {
            if let member = called.as(MemberAccessExprSyntax.self),
                UnhandledThrowingTask.starterNames.contains(member.declName.baseName.text), let base = member.base
            {
                return taskName(base) == nil ? nil : member.declName.baseName
            }
            return taskName(called)
        }

        /* The `Task` token of `Task`, `_Concurrency.Task`, or either specialised with `_` as its failure type. */
        static func taskName(_ expression: ExprSyntax) -> TokenSyntax? {
            var named = expression
            if let specialized = expression.as(GenericSpecializationExprSyntax.self) {
                guard
                    specialized.genericArgumentClause.arguments.last?.argument.as(IdentifierTypeSyntax.self)?.name.text
                        == "_"
                else { return nil }
                named = specialized.expression
            }
            if let reference = named.as(DeclReferenceExprSyntax.self), reference.baseName.text == "Task",
                reference.argumentNames == nil
            {
                return reference.baseName
            }
            if let member = named.as(MemberAccessExprSyntax.self), member.declName.baseName.text == "Task",
                member.base?.as(DeclReferenceExprSyntax.self)?.baseName.text == "_Concurrency"
            {
                return member.declName.baseName
            }
            return nil
        }

        /*
         Whether a statement's value goes nowhere. One of several statements always does: Swift returns only a
         body's lone expression. A lone statement does where its body returns nothing: a function without a
         return type or returning `Void`, an initializer, a setter or observer, a loop, a `do`, a `catch`, a
         `guard`, a `defer`, a file's top level, and a closure written `-> Void`. A branch of an `if` or a
         `switch` is dropped when the `if` or `switch` is. Anywhere else, a getter, a closure of unwritten type,
         a function returning a value, the value may be returned, and this is not judged dropped.
         */
        static func isDropped(_ item: CodeBlockItemSyntax) -> Bool {
            guard let list = item.parent?.as(CodeBlockItemListSyntax.self), let owner = list.parent else {
                return false
            }
            if list.count > 1 || owner.is(SourceFileSyntax.self) {
                return true
            }
            if let closure = owner.as(ClosureExprSyntax.self) {
                return closure.signature?.returnClause.map { isVoid($0.type) } ?? false
            }
            if let switchCase = owner.as(SwitchCaseSyntax.self) {
                return switchCase.parent?.parent?.as(SwitchExprSyntax.self).map {
                    isDropped(expression: ExprSyntax($0))
                } ?? false
            }
            if let clause = owner.as(IfConfigClauseSyntax.self) {
                return clause.parent?.parent?.parent?.as(CodeBlockItemSyntax.self).map(isDropped) ?? false
            }
            guard let block = owner.as(CodeBlockSyntax.self), let blockOwner = block.parent else { return false }
            if let function = blockOwner.as(FunctionDeclSyntax.self) {
                return function.signature.returnClause.map { isVoid($0.type) } ?? true
            }
            if let accessor = blockOwner.as(AccessorDeclSyntax.self) {
                return accessor.accessorSpecifier.tokenKind != .keyword(.get)
            }
            if let conditional = blockOwner.as(IfExprSyntax.self) {
                return isDropped(expression: ExprSyntax(conditional))
            }
            return blockOwner.is(InitializerDeclSyntax.self) || blockOwner.is(DeinitializerDeclSyntax.self)
                || blockOwner.is(DoStmtSyntax.self)
                || blockOwner.is(CatchClauseSyntax.self) || blockOwner.is(ForStmtSyntax.self)
                || blockOwner.is(WhileStmtSyntax.self)
                || blockOwner.is(RepeatStmtSyntax.self) || blockOwner.is(GuardStmtSyntax.self)
                || blockOwner.is(DeferStmtSyntax.self)
        }

        /* An `if` or `switch` (through any `else if` above it) whose own value goes nowhere: a statement that is dropped. */
        static func isDropped(expression: ExprSyntax) -> Bool {
            var outermost = Syntax(expression)
            /* An `if` whose parent is an `if` is that `if`'s `else if`: its conditions hold expressions, never an `if` directly. */
            while let parent = outermost.parent, parent.is(IfExprSyntax.self), outermost.is(IfExprSyntax.self) {
                outermost = parent
            }
            if let statement = outermost.parent?.as(ExpressionStmtSyntax.self) {
                outermost = Syntax(statement)
            }
            return outermost.parent?.as(CodeBlockItemSyntax.self).map(isDropped) ?? false
        }

        /* `Void` or `()`, as a return type is written. */
        static func isVoid(_ type: TypeSyntax) -> Bool {
            if let identifier = type.as(IdentifierTypeSyntax.self) {
                return identifier.name.text == "Void" && identifier.genericArgumentClause == nil
            }
            return type.as(TupleTypeSyntax.self)?.elements.isEmpty ?? false
        }
    }
}
