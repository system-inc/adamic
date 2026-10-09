import SwiftSyntax

/*
 No `unowned` references: `[unowned self]`, `[unowned(safe) delegate]`, `[unowned(unsafe) self]` in a capture
 list, and `unowned let owner: Owner` or `unowned(unsafe) var owner: Owner` on a stored property or a local.

 `unowned` is a force unwrap of a reference. It asserts the object outlives every use, and when that is
 wrong the process crashes (`unowned`, `unowned(safe)`) or reads freed memory (`unowned(unsafe)`), the same
 unchecked assertion the house forbids as `!`. The repair is `weak`, with a `guard let` that says what
 happens when the object is gone. Report only, no fix, because choosing that behavior is the point.

 Matches SwiftLint's `unowned_variable_capture` on capture lists, including `unowned(unsafe)`, which it also
 flags under its default `allow_explicit_unsafe_unowned: false`. Findings point at the `unowned` keyword,
 the column SwiftLint marks.

 Stored properties and locals are cohere-only by design. SwiftLint checks captures alone and lists
 `unowned var value: First` among its non-triggering examples, but a property is the same crash with a
 longer fuse: it holds the reference for the object's whole life instead of one closure's. The repair is
 the same `weak var` and a guard at each use.

 Nothing is exempt. The keyword is matched as the parser reads it (a closure capture specifier or a
 declaration modifier), so an identifier that happens to be spelled `unowned` is never confused with it.
 */
public struct UnownedVariableCapture: FileRule {
    public let name = "cohere-swift/unowned-variable-capture"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "unowned_variable_capture"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        let captures = visitor.captures.map { keyword in
            file.finding(
                at: keyword,
                rule: name,
                messageId: "unownedCapture",
                message:
                    "An unowned capture crashes the process if the object is gone when the closure runs, like a force unwrap of the reference. Capture it weak and guard let it, saying what happens when it is gone.",
            )
        }
        let properties = visitor.properties.map { keyword in
            file.finding(
                at: keyword,
                rule: name,
                messageId: "unownedProperty",
                message:
                    "An unowned reference crashes the process if the object is gone when it is read, like a force unwrap of the reference. Declare it weak and guard let it where it is used, saying what happens when it is gone.",
            )
        }
        return (captures + properties).sorted { ($0.line, $0.column) < ($1.line, $1.column) }
    }

    /* Collects the `unowned` keyword of every capture specifier and every declaration modifier. */
    final class Visitor: SyntaxVisitor {
        private(set) var captures: [TokenSyntax] = []
        private(set) var properties: [TokenSyntax] = []

        override func visit(_ node: ClosureCaptureSpecifierSyntax) -> SyntaxVisitorContinueKind {
            if node.specifier.tokenKind == .keyword(.unowned) {
                captures.append(node.specifier)
            }
            return .visitChildren
        }

        override func visit(_ node: DeclModifierSyntax) -> SyntaxVisitorContinueKind {
            if node.name.tokenKind == .keyword(.unowned) {
                properties.append(node.name)
            }
            return .visitChildren
        }
    }
}
