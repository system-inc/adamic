import SwiftOperators
import SwiftSyntax

/*
 No conditional whose every branch does the same thing: `highQuality ? "pro" : "pro"`, or `if ready { start() }
 else { start() }`. The condition chooses nothing, so the code runs the same way whether it is true or false. It is
 nearly always a copied branch that was never edited, and the reader trusts a decision that is not being made. The
 repair is to write the branch that was meant, or, if both really are the same, to drop the conditional and keep one
 copy. There is no fix, because which branch was meant is the author's call.

 It ports `nexus/correctness-no-identical-branches`, whose first site was ahra's Kling request builder, where a
 quality flag picked `'pro'` either way. The shapes map one to one:
 - A ternary is a ternary. The engine's tree is not operator-folded, so the rule folds its own copy with the standard
   operator table, as identical-operands does; folding moves no token, so positions are the file's. Parentheses
   around a whole branch are dropped, so `ready ? (count + 1) : count + 1` matches.
 - An `if` with an `else`, whether a statement or an `if` expression. An `if` with no `else` has no second branch.
 - An `else if` chain is judged from its last `if`, the one with the final `else`: when its two branches match, that
   condition is ignored, and the finding widens up the chain over every earlier `if` whose branch is the same too.
   `if a { x() } else if b { x() } else { x() }` is one finding on the whole chain, and `if a { y() } else if b { x() }
   else { x() }` is one finding on `if b`. A ternary chain widens the same way through its false branch. Two equal
   branches with a different one after them, `if a { x() } else if b { x() } else { y() }`, are not reported: that is
   `if a || b` written long, and both conditions still decide. A chain with no final `else` is never reported.
 - Empty branches (`{}`, or braces holding only comments) are not reported: a commented empty branch carries its
   intent in the comments, the one thing the comparison ignores. Branches that only `return`, `break`, `continue` or
   `throw` are reported.
 - `switch` is not covered, as the original declines it: a token-equal pair of cases can still differ in what falls
   through to them, and the default is usually the case written differently.
 - The original's `else if` against an `if` held in the first branch, `if a { if c { p() } else { q() } } else if c
   { p() } else { q() }`, is kept: the one statement of the first branch is compared with the whole `else if`.

 Same means the same tree written with the same tokens, as the original compares it, and as identical-operands and
 duplicate-conditions compare token kinds: node kinds, each child in the same slot, and every token's kind and text,
 with whitespace and comments ignored. The tree matters as well as the tokens because Swift reads a line break:
 `load\n(next)` is a reference and a tuple, two statements, where `load(next)` is one call, and the two have the
 same tokens. `"pro"` and `#"pro"#` are different tokens and stay different. A multi-line string's segments carry
 their indentation past the closing delimiter in the token, so two literals that differ only there differ.

 Where Swift needs a different exactness than TypeScript, each toward finding less:
 - The original asks the type checker whether a narrowing condition makes the same tokens resolve differently in
   each branch (`typeof value === 'string' ? format(value) : format(value)`). Swift narrows nothing through a Boolean
   condition: `x is String` and `x != nil` change no type, and a ternary's condition can only be a Boolean. So a
   ternary, and an `if` whose conditions are all Boolean expressions, mean the same thing in both branches, and the
   comparison needs no types. What Swift has instead are conditions that bind a name, `if let`, `if var` and `if
   case`, whose first branch sees a new declaration that may shadow one the second branch sees under the same
   spelling (`if let name { show(name) } else { show(name) }` passes a `String` and then a `String?`). An `if`
   whose binding or pattern conditions spell any identifier its first branch also spells is left alone; one whose
   branches use none of those names is still judged, since the binding cannot reach them.
 - `if #available` and `if #unavailable` are left alone: the same call can resolve to a newer overload in the first
   branch and an older one in the second, which is the reason to write such a pair.
 - A statement `if` that a result builder may transform is left alone. Under a builder each branch becomes its own
   `buildEither` arm, so SwiftUI gives `if flag { Editor() } else { Editor() }` two identities, and flipping the flag
   rebuilds the view and resets its state: the same tokens are different code. A file cannot see which builder
   applies, so the rule asks where the `if` sits. A closure, or a function, getter or subscript with a result type
   other than `Void`, may be transformed unless its own body holds a `return`, which turns the transform off (SE-0289,
   for a closure, an attributed declaration and one inferred from a protocol like `View.body`). An `if` there is
   skipped. An `if` in expression position (`let mode = if a { ... } else { ... }`), in an initializer, a setter, an
   observer, a `Void` function, top-level code, or a body with a `return`, is judged. A ternary is one value of one
   type under any builder, and is judged everywhere.
 - Branches that expand a freestanding macro are left alone. `#line` and `#column` are the position they are written
   at, so `if verbose { log(#line) } else { log(#line) }` tells the two apart, and any macro may read its position.
   A default argument that captures the caller's position (`fatalError`'s `line: UInt = #line`) is not treated this
   way: two copies of a crash in two branches still choose nothing, as two copies of a `throw` do in TypeScript.
 - A `#if` is a compile-time choice, not a conditional the code makes, and is not read.

 Known misses, each a finding not made and never one invented:
 - An `if` in a closure with no `return`, which is most callbacks (`Button` actions, `Task { }`, completion
   handlers), and in a value-returning function or getter with no `return`, which includes a body that is one `if`
   expression. Without the compiler's word on which closure parameters carry a builder, these cannot be told from
   `View.body`.
 - An `if let` or `if case` whose first branch spells any identifier its pattern spells, even one the pattern does
   not bind: `if case .loaded(expected) = state { show(expected) } else { show(expected) }` compares with an existing
   `expected` rather than binding one, and `if case .loaded = state { log(.loaded) } else { log(.loaded) }` only names
   the case. Both are left alone. A name spelled only in the value matched (`cache[key]`) does not count.
 - A statement `if` in a function or getter whose `Void` result is spelled `Swift.Void`, which is read as a result.
 - Branches that differ only in a trailing semicolon or in backticks on a name, which are different tokens; and
   branches that expand any freestanding macro, `#expect` and `#selector` included.
 - A ternary whose operator sequence holds an infix operator the standard table does not know, whose grouping the
   fold can only guess, as identical-operands skips the same.
 */
public struct CorrectnessNoIdenticalBranches: FileRule {
    public let name = "cohere-swift/correctness-no-identical-branches"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* A ternary needs its `?` and an `if` needs its `else`: nothing else can hold this shape. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("?") || file.source.contains("else")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        /* Unknown operators are reported to this handler and folded at a guess; a ternary among them is skipped below. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(original: file.tree)
        visitor.walk(folded)
        return visitor.found.map { conditional in
            file.finding(
                at: conditional,
                rule: name,
                message: RuleMessages.CorrectnessNoIdenticalBranches.identicalBranches(),
            )
        }
    }

    /* Collects each conditional whose branches match, widened up its chain. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Syntax] = []

        /* The file's own unfolded tree, where a ternary's flat operator sequence can still be read whole. */
        let original: SourceFileSyntax

        init(original: SourceFileSyntax) {
            self.original = original
            super.init(viewMode: .sourceAccurate)
        }

        override func visit(_ node: TernaryExprSyntax) -> SyntaxVisitorContinueKind {
            let thenBranch = Syntax(Self.withoutParentheses(node.thenExpression))
            guard foldedWithKnownOperators(node), !Self.expandsMacro(thenBranch),
                Self.same(thenBranch, Syntax(Self.withoutParentheses(node.elseExpression)))
            else {
                return .visitChildren
            }
            var anchor = node
            while let parent = Self.parentOutsideParentheses(Syntax(anchor))?.as(TernaryExprSyntax.self),
                Self.withoutParentheses(parent.elseExpression).id == anchor.id,
                foldedWithKnownOperators(parent),
                Self.same(Syntax(Self.withoutParentheses(parent.thenExpression)), thenBranch)
            {
                anchor = parent
            }
            found.append(Syntax(anchor))
            return .visitChildren
        }

        override func visit(_ node: IfExprSyntax) -> SyntaxVisitorContinueKind {
            guard let elseBody = node.elseBody, !node.body.statements.isEmpty, Self.judgesItsBranches(node),
                !Self.expandsMacro(Syntax(node.body.statements)), !Self.mayBeTransformedByResultBuilder(node)
            else {
                return .visitChildren
            }
            switch elseBody {
                case .codeBlock(let block):
                    guard !block.statements.isEmpty, Self.same(Syntax(node.body.statements), Syntax(block.statements))
                    else {
                        return .visitChildren
                    }
                case .ifExpr(let elseIf):
                    /* An `else if` against the first branch's one statement: the rare `if a { if c {...} else {...} } else if c {...} else {...}`. */
                    guard node.body.statements.count == 1, let only = node.body.statements.first, only.semicolon == nil,
                        Self.same(Self.unwrappedStatement(only.item), Syntax(elseIf))
                    else {
                        return .visitChildren
                    }
            }
            var anchor = node
            while let parent = Syntax(anchor).parent?.as(IfExprSyntax.self), case .ifExpr(let elseIf) = parent.elseBody,
                elseIf.id == anchor.id,
                Self.judgesItsBranches(parent), Self.same(Syntax(parent.body.statements), Syntax(node.body.statements))
            {
                anchor = parent
            }
            found.append(Syntax(anchor))
            return .visitChildren
        }

        /*
         Whether every infix operator in the ternary's flat sequence is one the standard table knows, so the fold grouped
         its branches the way the compiler does. Found by the `?`'s position in the unfolded tree.
         */
        func foldedWithKnownOperators(_ ternary: TernaryExprSyntax) -> Bool {
            guard let token = original.token(at: ternary.questionMark.positionAfterSkippingLeadingTrivia),
                let sequence = token.parent?.parent?.parent?.as(SequenceExprSyntax.self)
            else {
                return false
            }
            return sequence.elements.allSatisfy { element in
                guard let binaryOperator = element.as(BinaryOperatorExprSyntax.self) else { return true }
                return OperatorTable.standardOperators.infixOperator(named: binaryOperator.operator.text) != nil
            }
        }

        /*
         The same tree written with the same tokens: the same node kinds, each child in the same slot, and every token
         the same kind and text. Trivia is never read, so whitespace and comments drop out.
         */
        static func same(_ left: Syntax, _ right: Syntax) -> Bool {
            guard left.kind == right.kind else { return false }
            if let leftToken = left.as(TokenSyntax.self), let rightToken = right.as(TokenSyntax.self) {
                return leftToken.tokenKind == rightToken.tokenKind
            }
            let leftChildren = Array(left.children(viewMode: .sourceAccurate))
            let rightChildren = Array(right.children(viewMode: .sourceAccurate))
            guard leftChildren.count == rightChildren.count else { return false }
            return zip(leftChildren, rightChildren).allSatisfy { leftChild, rightChild in
                leftChild.keyPathInParent == rightChild.keyPathInParent && same(leftChild, rightChild)
            }
        }

        /* `(value)` is `value`: a parenthesised single unlabeled element, however deep. */
        static func withoutParentheses(_ expression: ExprSyntax) -> ExprSyntax {
            var current = expression
            while let tuple = current.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                current = only.expression
            }
            return current
        }

        /* The node a branch belongs to, looking out through any parentheses around it. */
        static func parentOutsideParentheses(_ node: Syntax) -> Syntax? {
            var current = node
            while let element = current.parent?.as(LabeledExprSyntax.self), element.label == nil,
                let list = element.parent?.as(LabeledExprListSyntax.self), list.count == 1,
                let tuple = list.parent?.as(TupleExprSyntax.self)
            {
                current = Syntax(tuple)
            }
            return current.parent
        }

        /* A statement as the tree holds it: an `if` in statement position is wrapped in an expression statement. */
        static func unwrappedStatement(_ item: CodeBlockItemSyntax.Item) -> Syntax {
            if case .stmt(let statement) = item, let wrapped = statement.as(ExpressionStmtSyntax.self) {
                return Syntax(wrapped.expression)
            }
            return Syntax(item)
        }

        /*
         Whether this `if`'s conditions mean the same in both branches: no availability condition, and no name a
         binding or pattern condition spells is spelled in the first branch, which is the only one that can see it.
         */
        static func judgesItsBranches(_ node: IfExprSyntax) -> Bool {
            var conditionNames: Set<String> = []
            for element in node.conditions {
                switch element.condition {
                    case .expression:
                        continue
                    case .availability:
                        return false
                    case .optionalBinding(let binding):
                        conditionNames.formUnion(identifierNames(Syntax(binding.pattern)))
                    case .matchingPattern(let matching):
                        conditionNames.formUnion(identifierNames(Syntax(matching.pattern)))
                }
            }
            return conditionNames.isDisjoint(with: identifierNames(Syntax(node.body.statements)))
        }

        /*
         Every name spelled under a node, without backticks, so `` `default` `` and `default` are one name. `self` counts
         too, since `if let self` rebinds it.
         */
        static func identifierNames(_ node: Syntax) -> Set<String> {
            Set(
                node.tokens(viewMode: .sourceAccurate).compactMap { token in
                    switch token.tokenKind {
                        case .identifier(let text):
                            return text.count > 1 && text.hasPrefix("`") && text.hasSuffix("`")
                                ? String(text.dropFirst().dropLast()) : text
                        case .keyword(.self):
                            return "self"
                        default:
                            return nil
                    }
                }
            )
        }

        static func expandsMacro(_ node: Syntax) -> Bool {
            let finder = MacroFinder(viewMode: .sourceAccurate)
            finder.walk(node)
            return finder.foundMacro
        }

        /*
         Whether a result builder may transform this `if`: it heads its chain in statement position, and the body it
         belongs to is a closure, or a function, getter or subscript with a result type other than `Void`, with no
         `return` of its own. See the comment at the top for why each owner is read as it is.
         */
        static func mayBeTransformedByResultBuilder(_ node: IfExprSyntax) -> Bool {
            var head = Syntax(node)
            while let parent = head.parent?.as(IfExprSyntax.self), case .ifExpr(let elseIf) = parent.elseBody,
                elseIf.id == head.id
            {
                head = Syntax(parent)
            }
            guard
                head.parent?.is(ExpressionStmtSyntax.self) == true || head.parent?.is(CodeBlockItemSyntax.self) == true
            else {
                /* An `if` expression is a value, and a builder transforms statements only. */
                return false
            }
            var current = head.parent
            while let step = current {
                if let closure = step.as(ClosureExprSyntax.self) {
                    return !containsReturn(Syntax(closure.statements))
                }
                if let function = step.as(FunctionDeclSyntax.self) {
                    guard let body = function.body, !isVoid(function.signature.returnClause?.type) else { return false }
                    return !containsReturn(Syntax(body))
                }
                if let accessor = step.as(AccessorDeclSyntax.self) {
                    guard accessor.accessorSpecifier.tokenKind == .keyword(.get), let body = accessor.body,
                        let accessorBlock = accessor.parent?.parent?.as(AccessorBlockSyntax.self)
                    else {
                        return false
                    }
                    return !isVoid(resultType(of: accessorBlock)) && !containsReturn(Syntax(body))
                }
                if let accessorBlock = step.as(AccessorBlockSyntax.self) {
                    guard case .getter(let body) = accessorBlock.accessors else { return false }
                    return !isVoid(resultType(of: accessorBlock)) && !containsReturn(Syntax(body))
                }
                if step.is(InitializerDeclSyntax.self) || step.is(DeinitializerDeclSyntax.self)
                    || step.is(SourceFileSyntax.self)
                {
                    return false
                }
                current = step.parent
            }
            return false
        }

        /* The type a getter produces: its property's annotation, or its subscript's result. */
        static func resultType(of accessorBlock: AccessorBlockSyntax) -> TypeSyntax? {
            if let binding = accessorBlock.parent?.as(PatternBindingSyntax.self) {
                return binding.typeAnnotation?.type
            }
            return accessorBlock.parent?.as(SubscriptDeclSyntax.self)?.returnClause.type
        }

        /* No result, `Void` or `()`: nothing for a builder to build. */
        static func isVoid(_ type: TypeSyntax?) -> Bool {
            guard let type else { return true }
            if let identifier = type.as(IdentifierTypeSyntax.self) {
                return identifier.name.text == "Void" && identifier.genericArgumentClause == nil
            }
            if let tuple = type.as(TupleTypeSyntax.self) {
                return tuple.elements.isEmpty
            }
            return false
        }

        static func containsReturn(_ body: Syntax) -> Bool {
            let finder = ReturnFinder(viewMode: .sourceAccurate)
            finder.walk(body)
            return finder.foundReturn
        }
    }

    /* Finds a `return` that belongs to the body walked, not to a closure, function or type declared inside it. */
    final class ReturnFinder: SyntaxVisitor {
        private(set) var foundReturn = false

        override func visit(_ node: ReturnStmtSyntax) -> SyntaxVisitorContinueKind {
            foundReturn = true
            return .skipChildren
        }

        override func visit(_ node: ClosureExprSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
        override func visit(_ node: FunctionDeclSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
        override func visit(_ node: InitializerDeclSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
        override func visit(_ node: AccessorBlockSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
        override func visit(_ node: StructDeclSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
        override func visit(_ node: ClassDeclSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
        override func visit(_ node: EnumDeclSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
        override func visit(_ node: ActorDeclSyntax) -> SyntaxVisitorContinueKind { .skipChildren }
    }

    /* Finds a freestanding macro expansion, an expression (`#line`) or a declaration (`#warning`). */
    final class MacroFinder: SyntaxVisitor {
        private(set) var foundMacro = false

        override func visit(_ node: MacroExpansionExprSyntax) -> SyntaxVisitorContinueKind {
            foundMacro = true
            return .skipChildren
        }

        override func visit(_ node: MacroExpansionDeclSyntax) -> SyntaxVisitorContinueKind {
            foundMacro = true
            return .skipChildren
        }
    }
}
