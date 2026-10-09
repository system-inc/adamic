import SwiftSyntax

/*
 No C-era spellings where Swift has its own: `CGPointMake(x, y)`, `CGRectGetWidth(rect)`, `NSWidth(rect)`,
 `arc4random_uniform(n)`. The old spellings predate Swift's own API; the modern form says the same thing and
 reads as Swift. One rule for four SwiftLint incumbents, each family under its own message id so findings
 compare per incumbent:
 - `legacyConstructor` (`legacy_constructor`): `CGPointMake(x, y)` becomes `CGPoint(x: x, y: y)`, and the same
   for CGSize, CGRect, CGVector, NSPoint, NSSize, NSRect, NSRange, UIEdgeInsets, NSEdgeInsets and UIOffset.
   Fixed: the callee becomes the type and each argument gets its label, the argument text left as written.
   Each of these C functions is a field-by-field initializer, so the rewrite cannot change meaning.
 - `legacyCGGeometryFunction` (`legacy_cggeometry_functions`): `CGRectGetWidth(rect)` becomes `rect.width`,
   `CGRectInset(rect, 10, 5)` becomes `rect.insetBy(dx: 10, dy: 5)`. Fixed: Swift imports these CoreGraphics
   functions under exactly those member names, so the member is the same function called the Swift way.
   The fix applies when the receiver carries its own type; otherwise the rewrite is a suggestion (below).
 - `legacyNSGeometryFunction` (`legacy_nsgeometry_functions`): `NSWidth(rect)` and its family. Fixed only for
   `NSEqualPoints` and `NSEqualSizes`, whose `==` compares the same two fields. Every other member is reported
   with the rewrite offered as a suggestion and never applied, because it is not the same function: the
   NSGeometry functions read the stored fields, while CGRect's members work on the standardized rectangle and
   treat empty and null rectangles their own way. Measured on macOS for a rect of width -4: `NSWidth` is -4 and
   `.width` is 4, `NSMinX` and `.minX` differ, `NSIntersectionRect` of disjoint rects is the zero rect where
   `.intersection` is the null rect, and `NSEqualRects` says false where `==` says true. SwiftLint applies all
   of these as fixes; this rule leaves that judgment to the reader.
 - `legacyRandom` (`legacy_random`): `arc4random()`, `arc4random_uniform(n)`, `drand48()`. Report only: the
   modern `Int.random(in:)` returns a different type than `UInt32`, so the repair is a decision about types.

 The rewrites move argument text as written and add parentheses only where moving it out of the call would
 change how it binds: `CGRectGetWidth(a ?? b)` offers `(a ?? b).width`, `!NSEqualPoints(a, b)` becomes
 `!(a == b)`, and a chain ending in a trailing closure is parenthesized so an `if` condition cannot misread it.

 A fix is applied without anyone looking, so it must never be able to break the build. An argument that
 leaves its parameter position (the receiver of a property or method, either side of `==`) also leaves the
 parameter's type behind, and anything whose type the compiler could have inferred from that parameter may
 no longer type-check: `CGRectGetWidth(decode())` with a generic `decode<T>() -> T` becomes `decode().width`,
 which does not compile. So the fix is applied only when every moved argument carries its own type: a plain
 name, `self` or `super`, a property chain on one (`view.frame`, `CGRect.zero`, `frames!.first`), or an
 explicit `as` cast (`value as CGRect`). Anything else (a function, initializer or subscript call anywhere in
 it, a generic member, an operator, a literal, a closure) keeps the finding and offers the rewrite as a
 suggestion instead, the way the NSGeometry rewrites are offered. Arguments that stay arguments (an
 initializer's, or a method's after the receiver) keep their parameter and are never a reason to withhold.
 No rewrite is offered at all when the edit would delete a comment, or when the moved receiver is typed only
 by the parameter (`CGRectGetWidth(.zero)` or `nil`), since `.zero.width` cannot compile however it is spelled.

 Matched by name, as SwiftLint matches: a call whose callee is the bare name. Not matched, each on purpose:
 - A call whose arity, labels or trailing closures do not fit the C function. The C function takes exactly
   that many unlabeled arguments, so any other shape is someone's own function. SwiftLint checks arity for
   the geometry functions only, and flags `arc4random(5)` or `CGPointMake(1, 2, 3)`.
 - Any name the same file declares: a function, a variable, a parameter or a closure parameter. A local
   `func CGPointMake(...)` or `func arc4random() -> Int` is the false positive matching by name invites, so
   once a file declares the name, no call to it in that file is reported, wherever the declaration sits.
   SwiftLint does not look. A shadow declared in another file of the module is invisible to one file's
   syntax and is the remaining risk; it would take a declaration whose name copies an Apple C function.
 - `NSEdgeInsetsEqual`, which SwiftLint rewrites to `==`. NSEdgeInsets is not Equatable, so that rewrite does
   not compile and there is no modern spelling to point to.
 - A qualified call (`Darwin.arc4random()`) and a bare reference (`.reduce(.null, CGRectUnion)`), which
   SwiftLint skips too: neither is a call of the bare name with arguments to move.
 */
public struct LegacyConstructors: FileRule {
    public let name = "cohere-swift/legacy-constructors"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "legacy_constructor"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let declared = DeclaredNames(viewMode: .sourceAccurate)
        declared.walk(file.tree)
        let visitor = Visitor(viewMode: .sourceAccurate, shadowed: declared.names)
        visitor.walk(file.tree)
        return visitor.found.map { match in
            let edits = Self.edits(for: match, in: file.source) ?? []
            let applies = match.legacy.family.rewritesSafely(match.legacy) && Self.movedArgumentsKeepTheirTypes(match)
            return file.finding(
                at: match.call,
                rule: name,
                messageId: match.legacy.family.messageId,
                message: Self.message(for: match),
                fixes: applies ? edits : [],
                suggestions: applies || edits.isEmpty
                    ? [] : [FindingRecord.Suggestion(message: "Write \(match.legacy.modern)", fixes: edits)],
            )
        }
    }

    static func message(for match: Match) -> String {
        let legacyName = match.name
        let modern = match.legacy.modern
        switch match.legacy.family {
            case .constructor:
                return
                    "\(legacyName) is the C spelling from before Swift had its own initializers. Write \(modern), which builds the same value and reads as Swift."
            case .coreGraphicsGeometry:
                return
                    "\(legacyName) is the C spelling of a CGRect member. Write \(modern), the same function called the Swift way."
            case .appKitGeometry where match.legacy.exact:
                return
                    "\(legacyName) is the C spelling from before Swift had its own API. Write \(modern), which compares the same fields and reads as Swift."
            case .appKitGeometry:
                return
                    "\(legacyName) is the C spelling from before Swift had its own API. Swift's form is \(modern), but it works on the standardized rectangle, so it differs when a width or height is negative or the rectangle is empty. Switch where that cannot happen."
            case .random:
                return
                    "\(legacyName) is the C random API from before Swift had its own. Write \(modern). The result type differs, so choose the type the code needs."
        }
    }

    /* A call the rule reports: the node, the bare name it calls, and what that name means. */
    struct Match {
        var call: FunctionCallExprSyntax
        var name: String
        var legacy: Legacy
    }

    /* The incumbent each name belongs to, which sets the message id. */
    enum Family {
        case constructor
        case coreGraphicsGeometry
        case appKitGeometry
        case random

        var messageId: String {
            switch self {
                case .constructor: "legacyConstructor"
                case .coreGraphicsGeometry: "legacyCGGeometryFunction"
                case .appKitGeometry: "legacyNSGeometryFunction"
                case .random: "legacyRandom"
            }
        }

        /* Whether the rewrite is applied as a fix rather than offered: only where it is the same function. */
        func rewritesSafely(_ legacy: Legacy) -> Bool {
            switch self {
                case .constructor, .coreGraphicsGeometry: true
                case .appKitGeometry: legacy.exact
                case .random: false
            }
        }
    }

    /* How the modern form is spelled from the call's own arguments. */
    enum Rewrite {
        /* `Type(label: first, ...)`: the callee becomes the type and each argument gets its label. */
        case initializer(type: String, labels: [String])
        /* `first.name`. */
        case property(String)
        /* `first.name(label: second, ...)`, or `second.name(first)` when reversed. An empty label is unlabeled. */
        case method(String, labels: [String], reversed: Bool)
        /* `first == second`. */
        case equality
        /* No mechanical rewrite; reported only. */
        case none
    }

    struct Legacy {
        var family: Family
        var arity: Int
        var rewrite: Rewrite
        /* The modern form as the message spells it. */
        var modern: String
        /* For the NSGeometry family: whether the modern form computes exactly what the C function does. */
        var exact = false
    }

    static let legacies: [String: Legacy] = {
        var table: [String: Legacy] = [:]
        let initializers: [(String, String, [String])] = [
            ("CGPointMake", "CGPoint", ["x", "y"]),
            ("CGSizeMake", "CGSize", ["width", "height"]),
            ("CGRectMake", "CGRect", ["x", "y", "width", "height"]),
            ("CGVectorMake", "CGVector", ["dx", "dy"]),
            ("NSMakePoint", "NSPoint", ["x", "y"]),
            ("NSMakeSize", "NSSize", ["width", "height"]),
            ("NSMakeRect", "NSRect", ["x", "y", "width", "height"]),
            ("NSMakeRange", "NSRange", ["location", "length"]),
            ("UIEdgeInsetsMake", "UIEdgeInsets", ["top", "left", "bottom", "right"]),
            ("NSEdgeInsetsMake", "NSEdgeInsets", ["top", "left", "bottom", "right"]),
            ("UIOffsetMake", "UIOffset", ["horizontal", "vertical"]),
        ]
        for (legacyName, type, labels) in initializers {
            table[legacyName] = Legacy(
                family: .constructor,
                arity: labels.count,
                rewrite: .initializer(type: type, labels: labels),
                modern: "\(type)(\(labels.map { "\($0):" }.joined()))",
            )
        }
        let properties: [(Family, String, String)] = [
            (.coreGraphicsGeometry, "CGRectGetWidth", "width"),
            (.coreGraphicsGeometry, "CGRectGetHeight", "height"),
            (.coreGraphicsGeometry, "CGRectGetMinX", "minX"),
            (.coreGraphicsGeometry, "CGRectGetMidX", "midX"),
            (.coreGraphicsGeometry, "CGRectGetMaxX", "maxX"),
            (.coreGraphicsGeometry, "CGRectGetMinY", "minY"),
            (.coreGraphicsGeometry, "CGRectGetMidY", "midY"),
            (.coreGraphicsGeometry, "CGRectGetMaxY", "maxY"),
            (.coreGraphicsGeometry, "CGRectIsNull", "isNull"),
            (.coreGraphicsGeometry, "CGRectIsEmpty", "isEmpty"),
            (.coreGraphicsGeometry, "CGRectIsInfinite", "isInfinite"),
            (.coreGraphicsGeometry, "CGRectStandardize", "standardized"),
            (.coreGraphicsGeometry, "CGRectIntegral", "integral"),
            (.appKitGeometry, "NSWidth", "width"),
            (.appKitGeometry, "NSHeight", "height"),
            (.appKitGeometry, "NSMinX", "minX"),
            (.appKitGeometry, "NSMidX", "midX"),
            (.appKitGeometry, "NSMaxX", "maxX"),
            (.appKitGeometry, "NSMinY", "minY"),
            (.appKitGeometry, "NSMidY", "midY"),
            (.appKitGeometry, "NSMaxY", "maxY"),
            (.appKitGeometry, "NSIsEmptyRect", "isEmpty"),
            (.appKitGeometry, "NSIntegralRect", "integral"),
        ]
        for (family, legacyName, property) in properties {
            table[legacyName] = Legacy(
                family: family,
                arity: 1,
                rewrite: .property(property),
                modern: "rect.\(property)",
            )
        }
        let methods: [(Family, String, String, [String], Bool, String)] = [
            (.coreGraphicsGeometry, "CGRectInset", "insetBy", ["dx", "dy"], false, "rect.insetBy(dx:dy:)"),
            (.coreGraphicsGeometry, "CGRectOffset", "offsetBy", ["dx", "dy"], false, "rect.offsetBy(dx:dy:)"),
            (.coreGraphicsGeometry, "CGRectUnion", "union", [""], false, "rect.union(other)"),
            (.coreGraphicsGeometry, "CGRectIntersection", "intersection", [""], false, "rect.intersection(other)"),
            (.coreGraphicsGeometry, "CGRectContainsRect", "contains", [""], false, "rect.contains(other)"),
            (.coreGraphicsGeometry, "CGRectContainsPoint", "contains", [""], false, "rect.contains(point)"),
            (.coreGraphicsGeometry, "CGRectIntersectsRect", "intersects", [""], false, "rect.intersects(other)"),
            (.appKitGeometry, "NSInsetRect", "insetBy", ["dx", "dy"], false, "rect.insetBy(dx:dy:)"),
            (.appKitGeometry, "NSOffsetRect", "offsetBy", ["dx", "dy"], false, "rect.offsetBy(dx:dy:)"),
            (.appKitGeometry, "NSUnionRect", "union", [""], false, "rect.union(other)"),
            (.appKitGeometry, "NSIntersectionRect", "intersection", [""], false, "rect.intersection(other)"),
            (.appKitGeometry, "NSContainsRect", "contains", [""], false, "rect.contains(other)"),
            (.appKitGeometry, "NSIntersectsRect", "intersects", [""], false, "rect.intersects(other)"),
            (.appKitGeometry, "NSPointInRect", "contains", [""], true, "rect.contains(point)"),
        ]
        for (family, legacyName, method, labels, reversed, modern) in methods {
            table[legacyName] = Legacy(
                family: family,
                arity: labels.count + 1,
                rewrite: .method(method, labels: labels, reversed: reversed),
                modern: modern,
            )
        }
        table["NSEqualPoints"] = Legacy(
            family: .appKitGeometry,
            arity: 2,
            rewrite: .equality,
            modern: "point == other",
            exact: true,
        )
        table["NSEqualSizes"] = Legacy(
            family: .appKitGeometry,
            arity: 2,
            rewrite: .equality,
            modern: "size == other",
            exact: true,
        )
        table["NSEqualRects"] = Legacy(family: .appKitGeometry, arity: 2, rewrite: .equality, modern: "rect == other")
        table["arc4random"] = Legacy(
            family: .random,
            arity: 0,
            rewrite: .none,
            modern: "UInt32.random(in: .min ... .max), or random(in:) over the range the code needs",
        )
        table["arc4random_uniform"] = Legacy(
            family: .random,
            arity: 1,
            rewrite: .none,
            modern: "Int.random(in: 0..<upperBound), or UInt32.random(in:) where the UInt32 result matters",
        )
        table["drand48"] = Legacy(family: .random, arity: 0, rewrite: .none, modern: "Double.random(in: 0..<1)")
        return table
    }()

    /* Collects every call of a legacy name whose shape fits the C function, unless the file declares that name. */
    final class Visitor: SyntaxVisitor {
        let shadowed: Set<String>
        private(set) var found: [Match] = []

        init(viewMode: SyntaxTreeViewMode, shadowed: Set<String>) {
            self.shadowed = shadowed
            super.init(viewMode: viewMode)
        }

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            if let legacyName = LegacyConstructors.legacyName(of: node),
                !shadowed.contains(legacyName),
                let legacy = LegacyConstructors.legacies[legacyName],
                LegacyConstructors.fitsCFunction(node, arity: legacy.arity)
            {
                found.append(Match(call: node, name: legacyName, legacy: legacy))
            }
            return .visitChildren
        }
    }

    /* The bare name a call calls, without backticks: `CGPointMake(...)`, never `Module.CGPointMake(...)` or `CGPointMake(_:_:)(...)`. */
    static func legacyName(of call: FunctionCallExprSyntax) -> String? {
        guard let reference = call.calledExpression.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil
        else { return nil }
        return unquoted(reference.baseName.text)
    }

    static func unquoted(_ text: String) -> String {
        text.hasPrefix("`") && text.hasSuffix("`") && text.count > 1 ? String(text.dropFirst().dropLast()) : text
    }

    /* Exactly the C function's arity, in parentheses, with no labels and no trailing closure. */
    static func fitsCFunction(_ call: FunctionCallExprSyntax, arity: Int) -> Bool {
        call.leftParen != nil && call.rightParen != nil && call.trailingClosure == nil
            && call.additionalTrailingClosures.isEmpty
            && call.arguments.count == arity && call.arguments.allSatisfy { $0.label == nil && $0.colon == nil }
    }

    /*
     Every name the file declares that a call could resolve to: functions, variables and constants, function
     and closure parameters. Scope is ignored on purpose, so a shadow anywhere in the file silences the name.
     */
    final class DeclaredNames: SyntaxVisitor {
        private(set) var names: Set<String> = []

        override func visit(_ node: FunctionDeclSyntax) -> SyntaxVisitorContinueKind {
            names.insert(LegacyConstructors.unquoted(node.name.text))
            return .visitChildren
        }

        override func visit(_ node: IdentifierPatternSyntax) -> SyntaxVisitorContinueKind {
            names.insert(LegacyConstructors.unquoted(node.identifier.text))
            return .visitChildren
        }

        override func visit(_ node: FunctionParameterSyntax) -> SyntaxVisitorContinueKind {
            names.insert(LegacyConstructors.unquoted((node.secondName ?? node.firstName).text))
            return .visitChildren
        }

        override func visit(_ node: ClosureParameterSyntax) -> SyntaxVisitorContinueKind {
            names.insert(LegacyConstructors.unquoted((node.secondName ?? node.firstName).text))
            return .visitChildren
        }

        override func visit(_ node: ClosureShorthandParameterSyntax) -> SyntaxVisitorContinueKind {
            names.insert(LegacyConstructors.unquoted(node.name.text))
            return .visitChildren
        }
    }

    /*
     The edits that spell the call the modern way, or nil when there is no rewrite or it could not be made
     safely. Edits touch only the callee, the parentheses and commas, and label positions, never the argument
     text, so a legacy call nested in another's arguments carries its own edits and both apply in one pass.
     */
    static func edits(for match: Match, in source: String) -> [FindingRecord.Edit]? {
        let call = match.call
        guard let rightParen = call.rightParen else { return nil }
        let arguments = Array(call.arguments)
        let calleeStart = call.calledExpression.positionAfterSkippingLeadingTrivia.utf8Offset
        let closeEnd = rightParen.endPositionBeforeTrailingTrivia.utf8Offset
        func start(_ index: Int) -> Int { arguments[index].expression.positionAfterSkippingLeadingTrivia.utf8Offset }
        func end(_ index: Int) -> Int { arguments[index].expression.endPositionBeforeTrailingTrivia.utf8Offset }
        func labeled(_ label: String) -> String { label.isEmpty ? "" : "\(label): " }

        var edits: [FindingRecord.Edit]
        switch match.legacy.rewrite {
            case .none:
                return nil
            case let .initializer(type, labels):
                edits = [
                    FindingRecord.Edit(
                        start: calleeStart,
                        end: call.calledExpression.endPositionBeforeTrailingTrivia.utf8Offset,
                        text: type,
                    )
                ]
                for (index, label) in labels.enumerated() {
                    edits.append(FindingRecord.Edit(start: start(index), end: start(index), text: labeled(label)))
                }
                return edits
            case let .property(property):
                guard let shape = receiverShape(arguments[0].expression) else { return nil }
                edits = [
                    FindingRecord.Edit(start: calleeStart, end: start(0), text: shape.open),
                    FindingRecord.Edit(start: end(0), end: closeEnd, text: "\(shape.close).\(property)"),
                ]
            case let .method(method, labels, reversed: false):
                guard let shape = receiverShape(arguments[0].expression) else { return nil }
                edits = [
                    FindingRecord.Edit(start: calleeStart, end: start(0), text: shape.open),
                    FindingRecord.Edit(
                        start: end(0),
                        end: start(1),
                        text: shape.close + ".\(method)(" + labeled(labels[0]),
                    ),
                ]
                for index in labels.indices.dropFirst() {
                    edits.append(
                        FindingRecord.Edit(
                            start: start(index + 1),
                            end: start(index + 1),
                            text: labeled(labels[index]),
                        )
                    )
                }
            case let .method(method, _, reversed: true):
                /* Swapping the arguments also swaps their evaluation order, so this form is only ever a suggestion. */
                guard let shape = receiverShape(arguments[1].expression) else { return nil }
                let receiver = arguments[1].expression.trimmedDescription
                let argument = arguments[0].expression.trimmedDescription
                return [
                    FindingRecord.Edit(
                        start: calleeStart,
                        end: closeEnd,
                        text: "\(shape.open)\(receiver)\(shape.close).\(method)(\(argument))",
                    )
                ]
            case .equality:
                let left = operandShape(arguments[0].expression)
                let right = operandShape(arguments[1].expression)
                let outer = equalityNeedsParentheses(call) ? Shape.parenthesized : Shape.bare
                edits = [
                    FindingRecord.Edit(start: calleeStart, end: start(0), text: outer.open + left.open),
                    FindingRecord.Edit(start: end(0), end: start(1), text: "\(left.close) == \(right.open)"),
                    FindingRecord.Edit(start: end(1), end: closeEnd, text: right.close + outer.close),
                ]
        }
        /* The replaced spans hold only the callee, parentheses, commas and trivia, so a slash there is a comment the edit would delete. */
        let bytes = Array(source.utf8)
        let deletesComment = edits.contains { edit in
            edit.start < edit.end && edit.end <= bytes.count && bytes[edit.start..<edit.end].contains(UInt8(ascii: "/"))
        }
        return deletesComment ? nil : edits
    }

    /* Whether moved text needs parentheses to keep binding as it did inside the call's own. */
    enum Shape {
        case bare
        case parenthesized

        var open: String { self == .bare ? "" : "(" }
        var close: String { self == .bare ? "" : ")" }
    }

    /*
     How an argument reads once a member is put after it, or nil when it cannot be moved out of the call at
     all because its type comes from the parameter it was passed to. A postfix chain rooted at a name (`rect`,
     `view.frame`, `layout(for: item)`, `frames[index]!`) takes the member bare; anything else is parenthesized,
     and so is a chain with a trailing closure, which an `if` condition would misread once the parentheses go.
     */
    static func receiverShape(_ expression: ExprSyntax) -> Shape? {
        let contextual = ContextualTyping(viewMode: .sourceAccurate)
        contextual.walk(expression)
        if contextual.found {
            return nil
        }
        return operandShape(expression)
    }

    static func operandShape(_ expression: ExprSyntax) -> Shape {
        var current = expression
        while true {
            if current.is(DeclReferenceExprSyntax.self) || current.is(SuperExprSyntax.self)
                || current.is(TupleExprSyntax.self)
            {
                return .bare
            }
            if let member = current.as(MemberAccessExprSyntax.self) {
                guard let base = member.base else { return .bare }
                current = base
            }
            else if let call = current.as(FunctionCallExprSyntax.self) {
                guard call.trailingClosure == nil, call.additionalTrailingClosures.isEmpty else {
                    return .parenthesized
                }
                current = call.calledExpression
            }
            else if let subscriptCall = current.as(SubscriptCallExprSyntax.self) {
                guard subscriptCall.trailingClosure == nil, subscriptCall.additionalTrailingClosures.isEmpty else {
                    return .parenthesized
                }
                current = subscriptCall.calledExpression
            }
            else if let unwrap = current.as(ForceUnwrapExprSyntax.self) {
                current = unwrap.expression
            }
            else if let specialization = current.as(GenericSpecializationExprSyntax.self) {
                current = specialization.expression
            }
            else {
                return .parenthesized
            }
        }
    }

    /* The arguments a rewrite moves out of their parameter position: the receiver, or both sides of `==`. */
    static func movedArguments(of match: Match) -> [ExprSyntax] {
        let arguments = match.call.arguments.map(\.expression)
        switch match.legacy.rewrite {
            case .initializer, .none: return []
            case .property, .method(_, _, reversed: false): return Array(arguments.prefix(1))
            case .method(_, _, reversed: true): return Array(arguments.dropFirst().prefix(1))
            case .equality: return arguments
        }
    }

    static func movedArgumentsKeepTheirTypes(_ match: Match) -> Bool {
        movedArguments(of: match).allSatisfy(carriesItsOwnType)
    }

    /*
     Whether an expression's type is settled by its own text, with nothing for the compiler to infer from the
     parameter it was passed to: a name, `self` or `super`, a property chain or force unwrap on one, a single
     parenthesized one, or an explicit `as` cast. A call of any kind, a generic or implicit member, an operator,
     a literal or a closure is not, because its type may have come from the parameter.
     */
    static func carriesItsOwnType(_ expression: ExprSyntax) -> Bool {
        if let reference = expression.as(DeclReferenceExprSyntax.self) {
            return reference.argumentNames == nil
        }
        if expression.is(SuperExprSyntax.self) {
            return true
        }
        if let member = expression.as(MemberAccessExprSyntax.self) {
            guard let base = member.base, member.declName.argumentNames == nil else { return false }
            return carriesItsOwnType(base)
        }
        if let unwrap = expression.as(ForceUnwrapExprSyntax.self) {
            return carriesItsOwnType(unwrap.expression)
        }
        if let tuple = expression.as(TupleExprSyntax.self) {
            guard tuple.elements.count == 1, let only = tuple.elements.first, only.label == nil else { return false }
            return carriesItsOwnType(only.expression)
        }
        if let cast = expression.as(AsExprSyntax.self) {
            return cast.questionOrExclamationMark == nil
        }
        /* The unfolded tree spells `value as CGRect` as the sequence `value`, `as`, `CGRect`. */
        if let sequence = expression.as(SequenceExprSyntax.self) {
            let parts = Array(sequence.elements)
            guard parts.count == 3, let cast = parts[1].as(UnresolvedAsExprSyntax.self) else { return false }
            return cast.questionOrExclamationMark == nil && parts[2].is(TypeExprSyntax.self)
        }
        return false
    }

    /*
     Finds what takes its type from context: an implicit member (`.zero`, `.init(...)`) or `nil`, outside the
     arguments of a call inside the receiver, which keep their own parameter's context, and outside closures.
     */
    final class ContextualTyping: SyntaxVisitor {
        private(set) var found = false

        override func visit(_ node: MemberAccessExprSyntax) -> SyntaxVisitorContinueKind {
            if node.base == nil {
                found = true
            }
            return .visitChildren
        }

        override func visit(_ node: NilLiteralExprSyntax) -> SyntaxVisitorContinueKind {
            found = true
            return .skipChildren
        }

        override func visit(_ node: LabeledExprListSyntax) -> SyntaxVisitorContinueKind {
            node.parent?.is(TupleExprSyntax.self) == true ? .visitChildren : .skipChildren
        }

        override func visit(_ node: ClosureExprSyntax) -> SyntaxVisitorContinueKind {
            .skipChildren
        }
    }

    /*
     `first == second` replaces a postfix call, so it needs parentheses unless it stands where nothing binds
     tighter: a statement, an initializer, a `return`, a condition, or an argument of a call that stays a call.
     */
    static func equalityNeedsParentheses(_ call: FunctionCallExprSyntax) -> Bool {
        guard let parent = call.parent else { return true }
        if parent.is(CodeBlockItemSyntax.self) || parent.is(InitializerClauseSyntax.self)
            || parent.is(ReturnStmtSyntax.self)
            || parent.is(ConditionElementSyntax.self)
        {
            return false
        }
        if parent.is(LabeledExprSyntax.self), let owner = parent.parent?.parent {
            if let enclosing = owner.as(FunctionCallExprSyntax.self), let enclosingName = legacyName(of: enclosing) {
                return legacies[enclosingName] != nil
            }
            return false
        }
        return true
    }
}
