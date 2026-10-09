import SwiftParser
import SwiftSyntax

/* Finds the `name: "<target>"` argument that declares a target in a manifest, so a package finding points at the line someone would edit. */
final class TargetNameFinder: SyntaxVisitor {
    private let targetName: String
    private(set) var found: LabeledExprSyntax?

    private init(targetName: String) {
        self.targetName = targetName
        super.init(viewMode: .sourceAccurate)
    }

    static func find(_ targetName: String, in tree: SourceFileSyntax) -> LabeledExprSyntax? {
        let finder = TargetNameFinder(targetName: targetName)
        finder.walk(tree)
        return finder.found
    }

    override func visit(_ node: LabeledExprSyntax) -> SyntaxVisitorContinueKind {
        guard found == nil else { return .skipChildren }
        if node.label?.text == "name",
            let literal = node.expression.as(StringLiteralExprSyntax.self),
            literal.representedLiteralValue == targetName,
            isTargetDeclaration(node)
        {
            found = node
            return .skipChildren
        }
        return .visitChildren
    }

    /* Only `.target(name:)`, `.executableTarget(name:)`, `.testTarget(name:)` and kin: a product or a dependency can carry the same name. */
    private func isTargetDeclaration(_ node: LabeledExprSyntax) -> Bool {
        guard let call = node.parent?.parent?.as(FunctionCallExprSyntax.self),
            let member = call.calledExpression.as(MemberAccessExprSyntax.self)
        else {
            return false
        }
        return Self.targetDeclarations.contains(member.declName.baseName.text)
    }

    /* PackageDescription's target constructors, spelled out rather than matched by suffix, so a new API with a similar name is a decision rather than an accident. */
    private static let targetDeclarations: Set<String> = [
        "target", "executableTarget", "testTarget", "systemLibrary", "binaryTarget", "plugin", "macro",
    ]
}
