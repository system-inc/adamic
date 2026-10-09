import SwiftSyntax

/*
 No condition tested twice in one branch instruction: the same conditions on two branches of one `if` /
 `else if` chain, or the same pattern on two cases of one `switch`. The chain stops at the first branch
 that matches, so when the same test comes round a second time it has already failed, and the second
 branch can never run. The usual cause is a copied branch whose condition was never edited. The repair is
 to write the condition that branch was meant to test, or to delete the branch if it was never needed.

 What counts as the same, matching SwiftLint's `duplicate_conditions`:
 - An `if` branch's condition list is compared as a set of clauses, so `if a, b` and `else if b, a` are
   the same, and `if a, b` and `else if a` are not. Every branch of the chain is compared with every other,
   and each copy is reported, the first included.
 - A `switch` case item is its pattern plus its `where` clause, compared item by item across every case of
   one case list, so `case "a", "b":` and `case "c", "a":` share `"a"`.
 - Two clauses are the same when their tokens are the same, ignoring whitespace and comments. SwiftLint
   compares the source text, so `x < 5` and `x<5` are a difference to it; they are one test, and the rule
   finds them. That is the only place it finds more than SwiftLint.

 Where it differs from SwiftLint on purpose, toward finding less:
 - A clause that calls a function, expands a macro, or awaits is skipped. Each evaluation is a new call
   and may return a new answer, so `if reader.next() == "{" ... else if reader.next() == "{"` reads two
   characters and both branches can run. SwiftLint flags it; a finding there would be wrong, so the miss
   is accepted. `#available` and `#unavailable` are availability conditions, not macros, and are compared.
 - A branch whose clauses call, expand or await is also a wall between the branches before and after it.
   Its call runs between the two tests and may change what the later copy reads, so `if queue.isEmpty
   ... else if queue.refill() ... else if queue.isEmpty` can reach its third branch. Copies on the same
   side of every such branch are still found.
 - An enum case pattern is spelled like a call, `case .failure(let error)`, and is not one: matching it
   runs no code. In a pattern, a call whose callee is a member (`.failure(...)`, `Result.failure(...)`) is
   read as an enum case and only its arguments are searched for calls. A bare call in a pattern,
   `case makeLimit():`, is a call.
 - `case` items inside an `#if` clause are compared only with the items of the same clause, as SwiftLint
   does, since the items of two clauses are never compiled together.

 Property reads and subscripts are trusted to answer the same way twice, as SwiftLint trusts them. A
 getter that changes what it returns between two tests in one chain is not one the rule should protect.
 */
public struct DuplicateConditions: FileRule {
    public let name = "cohere-swift/duplicate-conditions"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "duplicate_conditions"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { mark in
            file.finding(
                at: mark,
                rule: name,
                messageId: "duplicateConditions",
                message:
                    "This condition is tested again in the same if/else if chain or switch, so whichever copy comes second can never run. Change the copy to the condition that branch was meant to test, or delete the branch.",
            )
        }
    }

    /* One comparable test: a clause's token kinds, or a case item's pattern and `where` clause. */
    typealias Key = [TokenKind]

    /* Collects the first token of every condition list or case item that repeats one earlier or later in its chain. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Syntax] = []

        override func visit(_ node: IfExprSyntax) -> SyntaxVisitorContinueKind {
            /* An `else if` is read as part of the chain its first `if` heads. */
            if node.parent?.is(IfExprSyntax.self) == true {
                return .visitChildren
            }
            var chain: [IfExprSyntax] = []
            var link: IfExprSyntax? = node
            while let current = link {
                chain.append(current)
                link = current.elseBody?.as(IfExprSyntax.self)
            }
            let branches = chain.map { branch in
                Branch(
                    mark: Syntax(branch.conditions),
                    key: branch.conditions.contains { Self.mayVary(Syntax($0.condition)) }
                        ? nil
                        : Set(branch.conditions.map { Self.key($0.condition) }),
                )
            }
            report(branches)
            return .visitChildren
        }

        override func visit(_ node: SwitchCaseListSyntax) -> SyntaxVisitorContinueKind {
            var items: [Item] = []
            for element in node {
                guard case .switchCase(let switchCase) = element, case .case(let label) = switchCase.label else {
                    continue
                }
                for caseItem in label.caseItems {
                    let varies =
                        Self.mayVary(Syntax(caseItem.pattern))
                        || (caseItem.whereClause.map { Self.mayVary(Syntax($0)) } ?? false)
                    let whereKey = caseItem.whereClause.map { Self.key($0) } ?? []
                    items.append(
                        Item(
                            mark: Syntax(caseItem),
                            key: varies ? nil : Self.key(caseItem.pattern) + [.semicolon] + whereKey,
                        )
                    )
                }
            }
            report(items)
            return .visitChildren
        }

        /*
         Reports every test that appears more than once between two walls. A test without a key is a wall:
         it may vary, and it may change what the tests after it read.
         */
        func report<Compared: Hashable>(_ tests: [Test<Compared>]) {
            var segment: [Test<Compared>] = []
            for test in tests + [Test(mark: nil, key: nil)] {
                if test.key != nil {
                    segment.append(test)
                    continue
                }
                var counts: [Compared: Int] = [:]
                for member in segment {
                    if let key = member.key {
                        counts[key, default: 0] += 1
                    }
                }
                for member in segment {
                    if let key = member.key, let mark = member.mark, counts[key, default: 0] > 1 {
                        found.append(mark)
                    }
                }
                segment = []
            }
        }

        /* Token for token, kinds and text, with trivia ignored. */
        static func key(_ node: some SyntaxProtocol) -> Key {
            node.tokens(viewMode: .sourceAccurate).map(\.tokenKind)
        }

        static func mayVary(_ node: Syntax) -> Bool {
            let finder = EvaluationFinder(viewMode: .sourceAccurate)
            finder.walk(node)
            return finder.foundEvaluation
        }
    }

    struct Test<Compared: Hashable> {
        /* Where the finding points; `nil` only for the closing wall `report` appends. */
        var mark: Syntax?
        /* What is compared; `nil` when the test may vary, which makes it a wall. */
        var key: Compared?
    }

    typealias Branch = Test<Set<Key>>
    typealias Item = Test<Key>

    /*
     Finds a call, a macro expansion or an `await`. Inside a pattern, a call on a member is an enum case
     pattern (`.some(let value)`), so only its arguments are searched.
     */
    final class EvaluationFinder: SyntaxVisitor {
        private(set) var foundEvaluation = false

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            if Self.isInPattern(node), node.calledExpression.is(MemberAccessExprSyntax.self) {
                return .visitChildren
            }
            foundEvaluation = true
            return .skipChildren
        }

        override func visit(_ node: MacroExpansionExprSyntax) -> SyntaxVisitorContinueKind {
            foundEvaluation = true
            return .skipChildren
        }

        override func visit(_ node: AwaitExprSyntax) -> SyntaxVisitorContinueKind {
            foundEvaluation = true
            return .skipChildren
        }

        /*
         Whether the call is the pattern of a `case` (a switch item, or `if case`), or an argument nested in
         one, rather than an expression the pattern is matched against.
         */
        static func isInPattern(_ node: FunctionCallExprSyntax) -> Bool {
            var current = Syntax(node)
            while let parent = current.parent {
                if parent.is(ExpressionPatternSyntax.self) {
                    return true
                }
                let isPatternArgumentStep =
                    parent.is(LabeledExprSyntax.self) || parent.is(LabeledExprListSyntax.self)
                    || parent.as(FunctionCallExprSyntax.self).map { call in
                        call.calledExpression.is(MemberAccessExprSyntax.self) && call.calledExpression.id != current.id
                    } ?? false
                    || parent.is(TupleExprSyntax.self)
                guard isPatternArgumentStep else { return false }
                current = parent
            }
            return false
        }
    }
}
