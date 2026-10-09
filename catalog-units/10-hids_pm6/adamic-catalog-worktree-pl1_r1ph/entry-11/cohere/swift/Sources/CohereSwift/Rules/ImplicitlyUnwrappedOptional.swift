import SwiftSyntax

/*
 No implicitly unwrapped optional: `var value: Type!`. Every later read is a force unwrap nobody wrote,
 which is worse than a written one: the crash site has no `!` to find. The repair is an optional the
 readers unwrap, or a non-optional set in the initializer.

 Kirk's ruling on force unwraps covers this. There is no exception for `@IBOutlet`: none of our code uses
 Interface Builder, and a rule decision belongs in the rule, never in config. Matches SwiftLint's
 `implicitly_unwrapped_optional` and swift-format's `NeverUseImplicitlyUnwrappedOptionals`.
 */
public struct ImplicitlyUnwrappedOptional: FileRule {
    public let name = "cohere-swift/implicitly-unwrapped-optional"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "implicitly_unwrapped_optional"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { type in
            file.finding(
                at: type,
                rule: name,
                messageId: "implicitlyUnwrappedOptional",
                message:
                    "An implicitly unwrapped optional makes every read a hidden force unwrap, so the crash site has no ! to find. Declare it optional and unwrap at the reads, or make it non-optional and set it in the initializer.",
            )
        }
    }

    /* Collects every `Type!`. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [ImplicitlyUnwrappedOptionalTypeSyntax] = []

        override func visit(_ node: ImplicitlyUnwrappedOptionalTypeSyntax) -> SyntaxVisitorContinueKind {
            found.append(node)
            return .visitChildren
        }
    }
}
