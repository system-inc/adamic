import Foundation
import SwiftSyntax

/*
 Full words in every name this code declares: the Swift form of `nexus/consistency-no-abbreviated-identifier`,
 judging the same vocabulary (`AbbreviationVocabulary`, read from the file the Go rule embeds).

 Declared names only, never references. A reference is spelled by whoever declared it: an Apple API named
 `src` is not ours to rename, and our own abbreviated declaration is reported once, where it is written,
 rather than at every use. What counts as a declaration:

 - every type (struct, class, enum, actor, protocol, typealias, associatedtype) and generic parameter;
 - functions, their parameter names (and a label only where it is also the name; see `declareParameters`),
   closure parameters, enum cases and their associated-value labels;
 - every binding a pattern introduces: `let`, `var`, `for`, `case let`, `catch let`, `if let`.

 Skipped, each because the spelling belongs to someone else:

 - an `override`'s name and argument labels, which the superclass chose (its parameter names are ours and
   are judged), the same exemption swift-format's naming rules make;
 - `if let value` with no initializer, which unwraps the existing `value` under its own name: the
   declaration of that `value` is where the finding belongs;
 - `_`, which names nothing.

 Single letters are not abbreviations, and the vocabulary has none: `x`, `y` and `i` are judged as the
 TypeScript rule judges them, which is not at all. The TypeScript rule's framework exemptions (`params` in
 Next.js routes, React's `ref`, `...args`) are policy for frameworks Swift does not have, and none
 carry over. No fixer, for the reason the Go rule gives: a rename without scope analysis can break a
 caller, and the suggestion in the message is applied by a reader with the scope in front of them.
 */
public struct ConsistencyNoAbbreviatedIdentifier: FileRule {
    public static let ruleName = "cohere-swift/consistency-no-abbreviated-identifier"
    public let name = ruleName
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil
    let vocabulary: AbbreviationVocabulary

    /* The vocabulary is handed in, loaded by the pipeline before anything is checked, never looked up here. */
    public init(vocabulary: AbbreviationVocabulary) {
        self.vocabulary = vocabulary
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.declared.compactMap { token in
            let spelled = token.text.trimmingCharacters(in: CharacterSet(charactersIn: "`"))
            guard let finding = vocabulary.find(spelled) else { return nil }
            return file.finding(at: token, rule: name, messageId: finding.messageId, message: finding.message)
        }
    }

    /* Collects the name token of every declaration this file makes. */
    final class Visitor: SyntaxVisitor {
        private(set) var declared: [TokenSyntax] = []

        private func declare(_ token: TokenSyntax?) {
            guard let token, case .identifier = token.tokenKind, token.text != "_" else { return }
            declared.append(token)
        }

        private static func isOverride(_ modifiers: DeclModifierListSyntax) -> Bool {
            modifiers.contains { $0.name.tokenKind == .keyword(.override) }
        }

        override func visit(_ node: StructDeclSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: ClassDeclSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: EnumDeclSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: ActorDeclSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: ProtocolDeclSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: TypeAliasDeclSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: AssociatedTypeDeclSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: GenericParameterSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }
        override func visit(_ node: EnumCaseElementSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }

        override func visit(_ node: EnumCaseParameterSyntax) -> SyntaxVisitorContinueKind {
            declare(node.firstName)
            declare(node.secondName)
            return .visitChildren
        }

        override func visit(_ node: FunctionDeclSyntax) -> SyntaxVisitorContinueKind {
            let overriding = Self.isOverride(node.modifiers)
            if !overriding {
                declare(node.name)
            }
            declareParameters(node.signature.parameterClause.parameters, overriding: overriding)
            return .visitChildren
        }

        override func visit(_ node: InitializerDeclSyntax) -> SyntaxVisitorContinueKind {
            declareParameters(node.signature.parameterClause.parameters, overriding: Self.isOverride(node.modifiers))
            return .visitChildren
        }

        override func visit(_ node: SubscriptDeclSyntax) -> SyntaxVisitorContinueKind {
            declareParameters(node.parameterClause.parameters, overriding: Self.isOverride(node.modifiers))
            return .visitChildren
        }

        /*
         A parameter with two names is judged by its second, the name the body uses, which is always ours. The
         label before it is API, and may be dictated: `requestOpenLink(source:link:params:)` is SwiftTerm's
         `TerminalViewDelegate` requirement, and renaming its label breaks the conformance. A parameter with one
         name is label and name at once and is judged, because the fix keeps the label and adds a name
         (`params parameters:`); an override's single name is the superclass's and is skipped.
         */
        private func declareParameters(_ parameters: FunctionParameterListSyntax, overriding: Bool) {
            for parameter in parameters {
                if let secondName = parameter.secondName {
                    declare(secondName)
                }
                else if !overriding {
                    declare(parameter.firstName)
                }
            }
        }

        override func visit(_ node: ClosureParameterSyntax) -> SyntaxVisitorContinueKind {
            declare(node.secondName ?? node.firstName)
            return .visitChildren
        }

        override func visit(_ node: ClosureShorthandParameterSyntax) -> SyntaxVisitorContinueKind {
            declare(node.name)
            return .visitChildren
        }

        override func visit(_ node: VariableDeclSyntax) -> SyntaxVisitorContinueKind {
            /* An overriding property's name is the superclass's; its accessors' bodies are still visited. */
            Self.isOverride(node.modifiers) ? visitSkippingBindings(node) : .visitChildren
        }

        private func visitSkippingBindings(_ node: VariableDeclSyntax) -> SyntaxVisitorContinueKind {
            for binding in node.bindings {
                if let accessors = binding.accessorBlock {
                    walk(accessors)
                }
                if let initializer = binding.initializer {
                    walk(initializer)
                }
            }
            return .skipChildren
        }

        override func visit(_ node: OptionalBindingConditionSyntax) -> SyntaxVisitorContinueKind {
            /* `if let value` with no initializer re-binds the existing `value`; only the explicit form declares a new name. */
            if node.initializer == nil {
                return .skipChildren
            }
            return .visitChildren
        }

        override func visit(_ node: IdentifierPatternSyntax) -> SyntaxVisitorContinueKind {
            declare(node.identifier)
            return .visitChildren
        }
    }
}
