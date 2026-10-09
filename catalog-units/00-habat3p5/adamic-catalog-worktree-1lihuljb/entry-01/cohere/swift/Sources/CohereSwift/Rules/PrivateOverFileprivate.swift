import SwiftSyntax

/*
 No `fileprivate` on a top-level declaration: `fileprivate func helper()` at file scope. Write `private`.

 Least access. At file scope `private` and `fileprivate` mean the same thing, the whole file, so the
 wider-sounding keyword says nothing the narrower one does not and teaches the reader to reach for
 `fileprivate` by habit. Since Swift 4 `private` also reaches same-file extensions of the type, which removed
 the old reason to widen. The repair is the keyword swap, and the rule ships it as a fix: at file scope the
 swap cannot change meaning, and `fileprivate(set)` becomes `private(set)` with its `(set)` kept.

 Matches SwiftLint's `private_over_fileprivate` in its default configuration, finding for finding: the same
 declaration kinds (actor, class, enum, func, protocol, struct, typealias, var and let), at the top level of
 the file or of a top-level `#if`, reported at the `fileprivate` keyword. Not flagged, each on purpose:
 - A `fileprivate` member nested inside a type or extension. Whether it could be `private` depends on every
   use in the file (an extension of a different type, or a sibling type, may read it), which needs more than
   one declaration's syntax to prove. SwiftLint draws the same line.
 - `fileprivate extension`. SwiftLint checks extensions only with `validate_extensions`, off by default, so
   the house default leaves the spelling of an extension's access to its author.
 Where we differ from SwiftLint is the fix only. SwiftLint replaces the whole modifier, so its fix turns a
 top-level `fileprivate(set) var` into `private var`, which also narrows the getter and breaks readers in
 other files. This rule replaces the keyword token alone.
 */
public struct PrivateOverFileprivate: FileRule {
    public let name = "cohere-swift/private-over-fileprivate"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "private_over_fileprivate"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { keyword in
            file.finding(
                at: keyword,
                rule: name,
                messageId: "privateOverFileprivate",
                message:
                    "At the top of a file, fileprivate and private both mean the whole file, so fileprivate only sounds wider than it is. Write private, the least access that says the same thing.",
                fixes: [
                    FindingRecord.Edit(
                        start: keyword.positionAfterSkippingLeadingTrivia.utf8Offset,
                        end: keyword.endPositionBeforeTrailingTrivia.utf8Offset,
                        text: "private",
                    )
                ],
            )
        }
    }

    /*
     Collects the `fileprivate` keyword of every top-level declaration. Every code-block item the walk reaches
     is top level: the walk descends only through the file's statements and the clauses of a top-level `#if`,
     and skips the inside of everything else, so a member, a local, or a closure body is never seen.
     */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [TokenSyntax] = []

        override func visit(_ node: CodeBlockItemSyntax) -> SyntaxVisitorContinueKind {
            if node.item.is(IfConfigDeclSyntax.self) {
                return .visitChildren
            }
            if let modifiers = Self.modifiers(of: node.item),
                let keyword = modifiers.first(where: { $0.name.tokenKind == .keyword(.fileprivate) })?.name
            {
                found.append(keyword)
            }
            return .skipChildren
        }

        /* The declaration kinds SwiftLint checks by default; an extension is left out, as `validate_extensions: false` leaves it. */
        static func modifiers(of item: CodeBlockItemSyntax.Item) -> DeclModifierListSyntax? {
            if let declaration = item.as(ActorDeclSyntax.self) { return declaration.modifiers }
            if let declaration = item.as(ClassDeclSyntax.self) { return declaration.modifiers }
            if let declaration = item.as(EnumDeclSyntax.self) { return declaration.modifiers }
            if let declaration = item.as(FunctionDeclSyntax.self) { return declaration.modifiers }
            if let declaration = item.as(ProtocolDeclSyntax.self) { return declaration.modifiers }
            if let declaration = item.as(StructDeclSyntax.self) { return declaration.modifiers }
            if let declaration = item.as(TypeAliasDeclSyntax.self) { return declaration.modifiers }
            if let declaration = item.as(VariableDeclSyntax.self) { return declaration.modifiers }
            return nil
        }
    }
}
