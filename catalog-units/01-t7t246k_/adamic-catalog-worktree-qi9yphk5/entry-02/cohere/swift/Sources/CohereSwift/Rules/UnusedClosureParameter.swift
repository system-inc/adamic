import SwiftSyntax

/*
 No named closure parameter that the body never reads: `items.map { item in 3 }`. Name it `_`.

 A parameter name is a promise that the value matters to the closure. When the body never reads it, the
 name sends the reader looking for a use that is not there, and hides the fact that the closure ignores
 its argument. The repair is `_`, which says the argument is ignored on purpose. The rule ships that
 rename as a fix: the parameter's name token alone becomes `_`, so the parameter count stays the same and
 any type annotation stays where it was (`(number: TypeA)` becomes `(_: TypeA)`), which leaves the
 closure's type, and so every overload choice and inference that depends on it, untouched. The compiler
 agrees the two spellings mean the same: a parameter named `self` that the body never spells out is
 replaceable too, since implicit member references in a closure resolve through the enclosing method's
 `self`, never through a parameter that happens to be named `self` (checked against Swift 6.4).

 Matches SwiftLint's `unused_closure_parameter`, finding for finding on every example it documents, reported
 at the parameter's name. A parameter counts as read when the body names it anywhere SwiftLint looks: a
 plain reference, string interpolation, a backticked spelling of either side (`` `class` `` and `class`
 are the same name), or a `$name` projection for a `$name` parameter. A name written only in a comment,
 a string, or after a dot (`.foo` or `other.foo`) is not a read. On top of SwiftLint, these count as reads
 because they are, and without them the rename to `_` would not compile:
 - A shorthand optional binding, `if let name` or `guard let name`, which reads the outer `name`.
 - A capture written without an initializer, `[name]` or `[weak name]`, in a nested closure's capture list.
 - `_name` for a `$name` parameter, the property wrapper's backing storage.
 Where we are stricter than SwiftLint, each one a parameter the body provably never reads:
 - A key path component, `\.name`, names a property, never the parameter, so it is not a read. SwiftLint
   counts it as one and misses the finding.
 - A nested closure that binds the same name, as a parameter or a capture (`{ name in ... }` or
   `[name = other] in`), shadows the outer parameter for its whole body, so references inside that body are
   the inner name's. Its capture list is still read in the outer scope. SwiftLint matches by name alone and
   misses the finding.
 Not flagged, each on purpose:
 - A parameter shadowed by a local declaration in the body: `{ value in let value = 3; return value }`.
   Whether every later reference means the inner `value` depends on statement order, nested scopes, local
   functions that can be called before their declaration, and patterns that bind in conditions, which is a
   full scope resolver's job. Every reference to the name counts as a read, so this case is a known miss,
   the same miss SwiftLint has, never a wrong finding. A nested function whose parameter reuses the name is
   treated the same way.
 - When the outer parameter is a `$name` projection, nested closures are not treated as shadowing it,
   because an inner plain `name` hides `name` but not `$name`; every spelling counts as a read.
 - A parameter with both an argument label and a name, `(label name: Int)`. Closures reject labels, so the
   compiler already reports it.
 No fix for a `$name` parameter. `$name` asks the property wrapper to wrap the argument, and `_` drops that
 request, which can change which overload a SwiftUI initializer resolves to. The finding stays; the rename
 is the reader's call.
 */
public struct UnusedClosureParameter: FileRule {
    public let name = "cohere-swift/unused-closure-parameter"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "unused_closure_parameter"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.found.map { parameter in
            file.finding(
                at: parameter.token,
                rule: name,
                messageId: "unusedClosureParameter",
                message:
                    "The closure never reads this parameter, so its name tells the reader it matters when it does not. Name it _ to say the closure ignores this argument.",
                fixes: parameter.isProjection
                    ? []
                    : [
                        FindingRecord.Edit(
                            start: parameter.token.positionAfterSkippingLeadingTrivia.utf8Offset,
                            end: parameter.token.endPositionBeforeTrailingTrivia.utf8Offset,
                            text: "_",
                        )
                    ],
            )
        }
    }

    /* One named closure parameter: its name token, the name with `$` and backticks removed, and whether it was spelled `$name`. */
    struct Parameter {
        let token: TokenSyntax
        let bareName: String
        let isProjection: Bool

        init?(token: TokenSyntax) {
            guard token.tokenKind != .wildcard else { return nil }
            self.token = token
            self.bareName = UnusedClosureParameter.bareName(token.text)
            self.isProjection = token.text.hasPrefix("$")
        }
    }

    /* `$name` and `` `name` `` are both spellings of `name`. */
    static func bareName(_ text: String) -> String {
        String(text.filter { $0 != "$" && $0 != "`" })
    }

    /* The named parameters a closure declares, shorthand `{ a, b in }` or clause `{ (a: A, b) in }`. */
    static func parameters(of closure: ClosureExprSyntax) -> [Parameter] {
        switch closure.signature?.parameterClause {
            case .simpleInput(let list):
                return list.compactMap { Parameter(token: $0.name) }
            case .parameterClause(let clause):
                /* A labelled parameter binds its second name; closures reject labels anyway, so it is left to the compiler. */
                return clause.parameters.compactMap { $0.secondName == nil ? Parameter(token: $0.firstName) : nil }
            case nil:
                return []
        }
    }

    /* The names a nested closure binds for its own body: its parameters and its capture list. */
    static func boundNames(of closure: ClosureExprSyntax) -> Set<String> {
        var names = Set(parameters(of: closure).map(\.bareName))
        if case .parameterClause(let clause) = closure.signature?.parameterClause {
            for parameter in clause.parameters {
                if let secondName = parameter.secondName {
                    names.insert(bareName(secondName.text))
                }
            }
        }
        for capture in closure.signature?.capture?.items ?? [] {
            names.insert(bareName(capture.name.text))
        }
        return names
    }

    /* Collects every named parameter of every closure whose body never reads it. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [Parameter] = []

        override func visit(_ node: ClosureExprSyntax) -> SyntaxVisitorContinueKind {
            for parameter in UnusedClosureParameter.parameters(of: node) {
                let reads = ReadFinder(parameter: parameter, viewMode: .sourceAccurate)
                reads.walk(node.statements)
                if !reads.isRead {
                    found.append(parameter)
                }
            }
            return .visitChildren
        }
    }

    /* Walks a closure body looking for anything that reads one parameter, and stops looking once it finds one. */
    final class ReadFinder: SyntaxVisitor {
        let parameter: Parameter
        private(set) var isRead = false

        init(parameter: Parameter, viewMode: SyntaxTreeViewMode) {
            self.parameter = parameter
            super.init(viewMode: viewMode)
        }

        /* Whether a spelling names this parameter: `name`, `` `name` ``, `$name`, and `_name` when the parameter is `$name`. */
        func names(_ text: String) -> Bool {
            if UnusedClosureParameter.bareName(text) == parameter.bareName {
                return true
            }
            return parameter.isProjection && text == "_" + parameter.bareName
        }

        override func visit(_ node: DeclReferenceExprSyntax) -> SyntaxVisitorContinueKind {
            if isRead {
                return .skipChildren
            }
            /* The `foo` of `.foo`, `other.foo` and `\.foo` names a member, never a local. */
            let keyPath = node.keyPathInParent
            if keyPath == \MemberAccessExprSyntax.declName || keyPath == \KeyPathPropertyComponentSyntax.declName {
                return .skipChildren
            }
            if names(node.baseName.text) {
                isRead = true
            }
            return .skipChildren
        }

        /* `if let name` and `guard let name` read the outer `name` without spelling a reference. */
        override func visit(_ node: OptionalBindingConditionSyntax) -> SyntaxVisitorContinueKind {
            if isRead {
                return .skipChildren
            }
            if node.initializer == nil, let pattern = node.pattern.as(IdentifierPatternSyntax.self),
                names(pattern.identifier.text)
            {
                isRead = true
            }
            return .visitChildren
        }

        /* `[name]` and `[weak name]` capture the outer `name`, which is a read. */
        override func visit(_ node: ClosureCaptureSyntax) -> SyntaxVisitorContinueKind {
            if isRead {
                return .skipChildren
            }
            if node.initializer == nil, names(node.name.text) {
                isRead = true
            }
            return .visitChildren
        }

        /*
         A nested closure that binds the same name hides the parameter for its whole body; only its capture
         list is still in the outer scope. A `$name` parameter is never treated as hidden, see the rule's comment.
         */
        override func visit(_ node: ClosureExprSyntax) -> SyntaxVisitorContinueKind {
            if isRead {
                return .skipChildren
            }
            guard !parameter.isProjection, UnusedClosureParameter.boundNames(of: node).contains(parameter.bareName)
            else {
                return .visitChildren
            }
            if let capture = node.signature?.capture {
                let captureReads = ReadFinder(parameter: parameter, viewMode: viewMode)
                captureReads.walk(capture)
                isRead = captureReads.isRead
            }
            return .skipChildren
        }
    }
}
