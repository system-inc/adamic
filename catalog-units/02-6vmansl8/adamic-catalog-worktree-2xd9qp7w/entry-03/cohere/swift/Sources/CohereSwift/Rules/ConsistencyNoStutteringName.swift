import Foundation
import SwiftSyntax

/*
 No placeholder name that stutters against its own member: `result.result`, `value.value`, `outcome?.outcome`.
 The Swift form of `nexus/consistency-no-stuttering-name`, with its word list and its message.

 Why the stutter, and not every generic name: a name is read where it is used, so the question is whether the
 word still carries anything there. `let response = try await session.data(for: request)` needs no better
 word. `value.value.intValue` is a value whose own name had nothing to say, so it borrowed its member's word
 and said it twice. The repair is to rename the value for what it holds (`entry.value.intValue`), the type it
 came back as, or whatever tells it apart from the other values in scope. No fix: the right noun is the
 judgment the rule exists to ask for.

 What the TypeScript rule judges, and what that is in Swift. The Go rule fires on a property access whose
 object is a bare identifier spelled the same as the property, when the word is on the generic list. In
 TypeScript a member of the enclosing class is always reached through `this.`, so a bare identifier is a
 binding the code chose: a local, a parameter, an import or a module constant. The Swift node is the same,
 `MemberAccessExprSyntax` whose base is a plain `DeclReferenceExprSyntax`, but Swift's implicit `self` breaks
 the TypeScript guarantee: a bare `state` inside a method may be a stored property, and a property's name may
 be a protocol requirement or an override, chosen in another file or another module. So the base is resolved
 lexically first, and the stutter is reported only when the base is a binding whose name was chosen right
 here:

 - a local `let` or `var`, a closure parameter, a `[name = expression]` capture, a function parameter's name,
   an accessor's `set(name)`, and every name a pattern binds (`if let name = …`, `guard let`, `while let`,
   `if case let`, `for name in`, `case .some(let name)`, `catch let name`);
 - a variable declared at the top level of this file, when the use is not inside a type, where a member of
   the same name would shadow it.

 Skipped, each because the name belongs to someone else or the rule cannot see whose it is:

 - a base that does not resolve to such a binding: a property reached through implicit `self`, a global from
   another file, a module. The property may satisfy a protocol requirement or override a superclass, its
   conformance may be declared in an extension elsewhere, and syntax cannot tell; the decided exemption is to
   leave those alone, and the price is a miss, never a wrong finding.
 - an `override`'s single-name parameter, which is the superclass's argument label. A parameter with two names
   is judged by its second, which is always ours. A non-overriding single-name parameter is judged, as
   `consistency-no-abbreviated-identifier` judges it, because the repair keeps the label and adds a name
   (`result outcome: Outcome`), so a protocol's label survives the rename.
 - `if let value`, `guard let value` and a `[value]` capture with no initializer, which re-bind the outer
   `value` under its own name: resolution continues outward to the declaration that chose the word.
 - `self.value.value`, `model.value.value` and any other base that is not a bare name: the TypeScript rule
   skips `this.value.value` for the same reason, the outer word is a member, not a placeholder.
 - a word that is not on the generic list. `user.user` or `camera.camera` may be exactly right.

 Optional chaining and force unwrapping (`value?.value`, `value!.value`) are the same stutter written another
 way and are reported, the first as the TypeScript rule reports `outcome?.outcome`. TypeScript's computed
 access `outcome['outcome']` has no Swift spelling to carry over; a subscript with a string key is a lookup,
 not a member.
 */
public struct ConsistencyNoStutteringName: FileRule {
    public let name = "cohere-swift/consistency-no-stuttering-name"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    /*
     The words that count as saying nothing when they stutter: the Go rule's `defaultGenericNames`, kept short
     on purpose because these already read as placeholders everywhere.
     */
    static let genericNames: Set<String> = [
        "outcome", "result", "data", "value", "output", "response", "state", "item", "thing",
    ]

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { member in
            file.finding(
                at: member,
                rule: name,
                message: RuleMessages.ConsistencyNoStutteringName.stutteringName(name: Self.spelling(member)),
            )
        }
    }

    /* A name as written, without the backticks that let it reuse a keyword. */
    static func spelling(_ token: TokenSyntax) -> String {
        token.text.trimmingCharacters(in: ["`"])
    }

    /* Collects the member name token of every `word.word` whose base is a binding chosen in this file. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [TokenSyntax] = []

        override func visit(_ node: MemberAccessExprSyntax) -> SyntaxVisitorContinueKind {
            guard
                let base = node.base.flatMap(Self.bareName),
                base.argumentNames == nil,
                case .identifier = base.baseName.tokenKind,
                case .identifier = node.declName.baseName.tokenKind
            else { return .visitChildren }
            let word = ConsistencyNoStutteringName.spelling(base.baseName)
            guard word == ConsistencyNoStutteringName.spelling(node.declName.baseName),
                ConsistencyNoStutteringName.genericNames.contains(word)
            else {
                return .visitChildren
            }
            if Resolution.isChosenHere(word, from: Syntax(base)) {
                found.append(node.declName.baseName)
            }
            return .visitChildren
        }

        /* The bare name under `name`, `name?` or `name!`; anything longer is a member chain, not a placeholder. */
        static func bareName(_ expression: ExprSyntax) -> DeclReferenceExprSyntax? {
            if let optional = expression.as(OptionalChainingExprSyntax.self) {
                return optional.expression.as(DeclReferenceExprSyntax.self)
            }
            if let forced = expression.as(ForceUnwrapExprSyntax.self) {
                return forced.expression.as(DeclReferenceExprSyntax.self)
            }
            return expression.as(DeclReferenceExprSyntax.self)
        }
    }

    /*
     Lexical lookup of a bare name, innermost scope first, the way Swift itself looks it up within a function.
     It answers one question: did the code at this point choose the word, or did it arrive from somewhere the
     rule cannot see? Reaching a type, or the file without a match, answers "cannot see".
     */
    enum Resolution {
        /* What a scope says about the name: it declares it here, it declares it under someone else's spelling, or nothing. */
        enum Verdict {
            case chosenHere
            case dictated
            case notDeclared
        }

        static func isChosenHere(_ word: String, from reference: Syntax) -> Bool {
            var child = reference
            while let scope = child.parent {
                if scope.is(StructDeclSyntax.self) || scope.is(ClassDeclSyntax.self) || scope.is(EnumDeclSyntax.self)
                    || scope.is(ActorDeclSyntax.self) || scope.is(ExtensionDeclSyntax.self)
                    || scope.is(ProtocolDeclSyntax.self)
                {
                    return false
                }
                switch verdict(of: scope, for: word, reachedThrough: child) {
                    case .chosenHere:
                        return true
                    case .dictated:
                        return false
                    case .notDeclared:
                        child = scope
                }
            }
            return false
        }

        static func verdict(of scope: Syntax, for word: String, reachedThrough child: Syntax) -> Verdict {
            if let list = scope.as(CodeBlockItemListSyntax.self) {
                return statements(list, declare: word, before: child)
            }
            /*
             Parameters and captures are seen by the body only. A capture's initializer or a default argument is
             evaluated outside, so `[value = value.value]` reads the outer `value`, not the capture.
             */
            if let closure = scope.as(ClosureExprSyntax.self), child.id == closure.statements.id {
                return closureDeclares(closure, word)
            }
            if let function = scope.as(FunctionDeclSyntax.self), child.id == function.body?.id {
                return parameters(
                    function.signature.parameterClause.parameters,
                    declare: word,
                    overriding: isOverride(function.modifiers),
                )
            }
            if let initializer = scope.as(InitializerDeclSyntax.self), child.id == initializer.body?.id {
                return parameters(
                    initializer.signature.parameterClause.parameters,
                    declare: word,
                    overriding: isOverride(initializer.modifiers),
                )
            }
            if let subscriptDeclaration = scope.as(SubscriptDeclSyntax.self),
                child.id == subscriptDeclaration.accessorBlock?.id
            {
                return parameters(
                    subscriptDeclaration.parameterClause.parameters,
                    declare: word,
                    overriding: isOverride(subscriptDeclaration.modifiers),
                )
            }
            if let accessor = scope.as(AccessorDeclSyntax.self), child.id == accessor.body?.id {
                return accessor.parameters.map { ConsistencyNoStutteringName.spelling($0.name) == word } == true
                    ? .chosenHere : .notDeclared
            }
            if let conditions = scope.as(ConditionElementListSyntax.self) {
                /* A condition sees the bindings of the conditions before it. */
                return conditionsDeclare(conditions.prefix { $0.id != child.id }, word)
            }
            if let ifExpression = scope.as(IfExprSyntax.self), child.id == ifExpression.body.id {
                return conditionsDeclare(Array(ifExpression.conditions), word)
            }
            if let whileStatement = scope.as(WhileStmtSyntax.self), child.id == whileStatement.body.id {
                return conditionsDeclare(Array(whileStatement.conditions), word)
            }
            if let forStatement = scope.as(ForStmtSyntax.self),
                child.id == forStatement.body.id || child.id == forStatement.whereClause?.id
            {
                return binds(Syntax(forStatement.pattern), word) ? .chosenHere : .notDeclared
            }
            if let switchCase = scope.as(SwitchCaseSyntax.self), child.id == switchCase.statements.id,
                case .case(let label) = switchCase.label
            {
                return label.caseItems.contains { binds(Syntax($0.pattern), word) } ? .chosenHere : .notDeclared
            }
            if let catchClause = scope.as(CatchClauseSyntax.self), child.id == catchClause.body.id {
                return catchClause.catchItems.contains { item in item.pattern.map { binds(Syntax($0), word) } ?? false }
                    ? .chosenHere : .notDeclared
            }
            return .notDeclared
        }

        /*
         Statements before the one holding the reference: `let`/`var` and `guard`. At the top level of a file,
         every variable counts wherever it is written, since a global is visible before its line.
         */
        static func statements(_ list: CodeBlockItemListSyntax, declare word: String, before child: Syntax) -> Verdict {
            let topLevel = list.parent?.is(SourceFileSyntax.self) == true
            for statement in list {
                if statement.id == child.id && !topLevel {
                    break
                }
                if let variable = statement.item.as(VariableDeclSyntax.self),
                    variable.bindings.contains(where: { binds(Syntax($0.pattern), word) })
                {
                    return .chosenHere
                }
                if !topLevel, let guardStatement = statement.item.as(GuardStmtSyntax.self) {
                    let verdict = conditionsDeclare(Array(guardStatement.conditions), word)
                    if verdict != .notDeclared {
                        return verdict
                    }
                }
            }
            return .notDeclared
        }

        /* `if let word = …` and `case let … word … = …` choose the word; `if let word` re-binds the outer one. */
        static func conditionsDeclare(_ conditions: [ConditionElementSyntax], _ word: String) -> Verdict {
            for element in conditions.reversed() {
                if let binding = element.condition.as(OptionalBindingConditionSyntax.self), binding.initializer != nil,
                    binds(Syntax(binding.pattern), word)
                {
                    return .chosenHere
                }
                if let match = element.condition.as(MatchingPatternConditionSyntax.self),
                    binds(Syntax(match.pattern), word)
                {
                    return .chosenHere
                }
            }
            return .notDeclared
        }

        /* Closure parameters, and captures that give a new name; `[word]` alone re-binds the outer `word`. */
        static func closureDeclares(_ closure: ClosureExprSyntax, _ word: String) -> Verdict {
            guard let signature = closure.signature else { return .notDeclared }
            if let captures = signature.capture?.items,
                captures.contains(where: {
                    ConsistencyNoStutteringName.spelling($0.name) == word && $0.initializer != nil
                })
            {
                return .chosenHere
            }
            switch signature.parameterClause {
                case .simpleInput(let shorthand):
                    return shorthand.contains { ConsistencyNoStutteringName.spelling($0.name) == word }
                        ? .chosenHere : .notDeclared
                case .parameterClause(let clause):
                    return clause.parameters.contains {
                        ConsistencyNoStutteringName.spelling($0.secondName ?? $0.firstName) == word
                    } ? .chosenHere : .notDeclared
                case nil:
                    return .notDeclared
            }
        }

        /*
         A parameter with two names is ours by its second. A single name is label and name at once: ours unless
         the declaration overrides, when the superclass chose it.
         */
        static func parameters(_ list: FunctionParameterListSyntax, declare word: String, overriding: Bool) -> Verdict {
            for parameter in list {
                if let secondName = parameter.secondName {
                    if ConsistencyNoStutteringName.spelling(secondName) == word {
                        return .chosenHere
                    }
                }
                else if ConsistencyNoStutteringName.spelling(parameter.firstName) == word {
                    return overriding ? .dictated : .chosenHere
                }
            }
            return .notDeclared
        }

        static func isOverride(_ modifiers: DeclModifierListSyntax) -> Bool {
            modifiers.contains { $0.name.tokenKind == .keyword(.override) }
        }

        /* Whether a pattern introduces the word as a name: any identifier pattern inside it, outside a closure's own scope. */
        static func binds(_ pattern: Syntax, _ word: String) -> Bool {
            if let identifier = pattern.as(IdentifierPatternSyntax.self) {
                return ConsistencyNoStutteringName.spelling(identifier.identifier) == word
            }
            if pattern.is(ClosureExprSyntax.self) {
                return false
            }
            return pattern.children(viewMode: .sourceAccurate).contains { binds($0, word) }
        }
    }
}
