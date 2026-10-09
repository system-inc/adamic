import SwiftSyntax

/*
 Every escape hatch carries a why-comment directly above the declaration that uses it.

 The ruling (@system_cohere, on Kirk's suppression ruling): `@unchecked Sendable` and
 `nonisolated(unsafe)` are allowed only with a why-comment directly above each one. An escape hatch the
 compiler is told to trust is a suppression, and every suppression states its reason. `@preconcurrency
 import` is the same kind of thing (the compiler is told to stop checking a module), so it is held to the
 same rule.

 "Directly above" means a comment in the declaration's leading trivia with no blank line between it and the
 declaration. A comment two paragraphs up is about something else.
 */
public struct ConcurrencyRequireEscapeHatchReason: FileRule {
    public let name = "cohere-swift/concurrency-require-escape-hatch-reason"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.compactMap { hatch in
            guard let declaration = Self.enclosingDeclaration(of: hatch.node) else { return nil }
            if Self.hasCommentDirectlyAbove(declaration) {
                return nil
            }
            return file.finding(
                at: hatch.node,
                rule: name,
                message: RuleMessages.ConcurrencyRequireEscapeHatchReason.escapeHatchWithoutReason(
                    spelling: hatch.spelling
                ),
            )
        }
    }

    /* The nearest enclosing declaration: the one whose leading comment would explain the hatch. */
    static func enclosingDeclaration(of node: Syntax) -> Syntax? {
        var current: Syntax? = node
        while let candidate = current {
            if candidate.is(DeclSyntax.self) {
                return candidate
            }
            current = candidate.parent
        }
        return nil
    }

    /* A comment in the declaration's leading trivia, followed by at most one newline before the declaration. */
    static func hasCommentDirectlyAbove(_ declaration: Syntax) -> Bool {
        var newlinesSinceComment: Int?
        for piece in declaration.leadingTrivia {
            switch piece {
                case .lineComment, .blockComment, .docLineComment, .docBlockComment:
                    newlinesSinceComment = 0
                case let .newlines(count):
                    newlinesSinceComment = newlinesSinceComment.map { $0 + count }
                case let .carriageReturnLineFeeds(count):
                    newlinesSinceComment = newlinesSinceComment.map { $0 + count }
                default:
                    break
            }
        }
        guard let newlines = newlinesSinceComment else { return false }
        return newlines <= 1
    }

    /* One escape hatch: the node to report at and how it is spelled in the message. */
    struct Hatch {
        var node: Syntax
        var spelling: String
    }

    /* Collects `@unchecked`, `nonisolated(unsafe)` and `@preconcurrency import`. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Hatch] = []

        override func visit(_ node: AttributeSyntax) -> SyntaxVisitorContinueKind {
            let name = node.attributeName.trimmedDescription
            if name == "unchecked" {
                found.append(Hatch(node: Syntax(node), spelling: "@unchecked Sendable"))
            }
            else if name == "preconcurrency", node.parent?.parent?.is(ImportDeclSyntax.self) == true {
                found.append(Hatch(node: Syntax(node), spelling: "@preconcurrency import"))
            }
            return .visitChildren
        }

        override func visit(_ node: DeclModifierSyntax) -> SyntaxVisitorContinueKind {
            if node.name.tokenKind == .keyword(.nonisolated), node.detail?.detail.text == "unsafe" {
                found.append(Hatch(node: Syntax(node), spelling: "nonisolated(unsafe)"))
            }
            return .visitChildren
        }
    }
}
