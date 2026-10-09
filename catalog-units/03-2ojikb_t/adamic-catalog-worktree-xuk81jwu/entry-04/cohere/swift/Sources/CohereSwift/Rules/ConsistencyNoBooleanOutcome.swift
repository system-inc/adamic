import SwiftSyntax

/*
 No result shape that says how an operation turned out with a `Bool` beside an error or a value:
 `struct LookupResult { let success: Bool; let error: String? }`, or `typealias FetchResult = (ok: Bool, value: String)`.
 One bit collapses every way the operation can turn out, at the moment the distinction is cheapest to keep, and it
 leaves the reader to know which of the other fields hold for which value of the flag: is `value` the answer when
 `ok` is false, a placeholder, or nothing? "Not shaped like a code" and "shaped like one, but no such code exists"
 are different facts a caller usually wants to act on, and a lookup that finds nothing is an answer, not a fault. The
 repair is a sum type, which says by construction what each outcome carries: an enum with a case for each way the
 operation can turn out and the payload on the case that has it, or `Result`, or a throw for the failure.

 It ports `nexus/consistency-no-boolean-outcome`, which reads an `interface` and a `type` alias whose type is a bare
 object literal, and nothing else. The shapes map one to one:
 - An `interface` is a named field set with no behavior, and Swift's is a `struct`: its fields are its stored instance
   properties (a `let` or `var` with no accessor block, or with only `willSet`/`didSet`, which still stores). A
   computed property is derived, not a field a reader must know the validity of, and a `static` one is not the
   value's, so neither is read. `var ok, failed: Bool` gives both names the trailing type, as the compiler does.
 - A `type` alias of an object literal is a `typealias` of a tuple type with labels, each label a field. As in the
   original, only the bare literal counts: a parenthesized, optional or generic-wrapped tuple is not read.
 - A flag is one of the original's names exactly (`success`, `succeeded`, `ok`, `isSuccess`, `isOk`, `failed`,
   `isError`, `isFailure`) declared as `Bool` or `Swift.Bool`, the one type its `boolean` keyword names. A companion
   is a field named `error`, `errors`, `message`, `value`, `result`, `reason` or `data`, of any type. The first flag
   in declaration order is the one reported, on its binding (`ok: Bool`) or tuple element, as the original reports
   the property.
 - The original's exemptions are kept as they are, so the two languages decide the same field set the same way: a
   flag with no companion is ordinary state; a field set carrying any of TanStack Query's own fields (`isLoading`,
   `isPending`, `isFetching`, `isRefetching`, `status`) or every field of Cloudflare API v4's envelope (`success`,
   `errors`, `messages`, `result`) mirrors a third party's result; `isSuccess` beside `isError` is one settled state
   seen from two sides; a name ending `Properties` is display state. TanStack has no Swift counterpart, and the
   exemption is kept anyway, since a `status` beside `ok` mirrors an HTTP reply here as it does there.
 - The original exempts React properties, by their name, because a component's inputs are what to render, handed
   down by the parent. SwiftUI's components are the structs that conform to `View`, `ViewModifier` or one of the
   representable protocols, so a struct that declares one of those, on itself or in an extension in this file, is
   left alone the same way.
 - Generated code is held to the rule like hand-written code, as the original holds it, and so is a test target.
 - The original's `allowedTypeNames` option has no counterpart: this engine gives a rule no options.

 What it accepts as safe, and why: an enum with payloads, `Result`, and a throwing function, which are the repair; a
 literal or optional flag (`Bool?`), as the original skips `success: true` and a union with `undefined`, since neither
 is the bare two-valued flag; a flag named anything else (`isDone`, `matched`); a companion named anything else
 (`said`, `errorCode`, `daemonValue`); a field set with no companion, like Presence's `BodyRemoval.Outcome`, whose
 `isFailure` sits beside `said`.

 Known misses, every one a finding not made and never one invented:
 - An inline tuple, which is where Swift most often writes this shape: a property's type (macOS's
   `DeveloperSettingsView`, `[String: (ok: Bool, value: String)]`), a function's return type, a closure's. The
   original reads only declarations and leaves `Record<string, { ok: boolean; value: string }>` alone the same way,
   but TypeScript reaches for a named interface far more often than Swift reaches for a named tuple, so this port
   finds less of the idea than the original does.
 - A `class`, `actor`, `protocol` or enum case with labeled associated values: the original listens to no class,
   and a protocol's property requirements are a contract on conformers, not a result. A struct's stored properties
   inside `#if`, and a flag whose type is inferred (`var failed = false`), which an interface cannot spell.
 - Swift's spelling of an acronym, `isOK`, which the original's name list does not hold, and a type named `Bool` of
   our own, which a file rule cannot tell from the standard library's.
 - A struct made a `View` by an extension in another file, which is read and may report.
 */
public struct ConsistencyNoBooleanOutcome: FileRule {
    public let name = "cohere-swift/consistency-no-boolean-outcome"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Names that claim to say how an operation turned out. */
    static let outcomeFlagNames: Set<String> = [
        "success", "succeeded", "ok", "isSuccess", "isOk", "failed", "isError", "isFailure",
    ]

    /* Fields that turn a flag into a result envelope rather than ordinary state. */
    static let resultCompanionNames: Set<String> = ["error", "errors", "message", "value", "result", "reason", "data"]

    /* TanStack Query's own fields: a field set carrying one mirrors a third party's request result. */
    static let thirdPartyRequestFieldNames: Set<String> = [
        "isLoading", "isPending", "isFetching", "isRefetching", "status",
    ]

    /* Envelopes another system defines and sends over the wire, each its whole documented field set, so a shape of ours sharing a field or two is still read. */
    static let thirdPartyEnvelopes: [(name: String, fields: Set<String>)] = [
        /* Every Cloudflare API v4 response, success or failure. */
        ("Cloudflare API v4", ["success", "errors", "messages", "result"])
    ]

    /* Stripped from a declaration's name, repeatedly, so the suggestion reads right: `ClaudeCallResultInterface` suggests `ClaudeCallOutcome`. */
    static let roleSuffixes = ["Interface", "Type", "Result", "Response", "Properties"]

    /* SwiftUI's component protocols: a conforming struct's fields are display state its parent hands down, as React properties are. */
    static let componentProtocols: Set<String> = [
        "View", "ViewModifier", "NSViewRepresentable", "UIViewRepresentable", "NSViewControllerRepresentable",
        "UIViewControllerRepresentable",
    ]

    /* Every flagged shape spells `Bool` as the flag's type. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("Bool")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        return visitor.declarations.compactMap { declaration in
            if declaration.name.hasSuffix("Properties") {
                return nil
            }
            if declaration.kind == .structure,
                declaration.isComponent || visitor.componentExtensionNames.contains(declaration.name)
            {
                return nil
            }
            guard let flag = Self.booleanOutcomeFlag(declaration.fields) else { return nil }
            var finding = file.finding(
                at: flag.start,
                rule: name,
                message: RuleMessages.ConsistencyNoBooleanOutcome.booleanOutcome(
                    flagName: flag.name,
                    declaration: "\(declaration.kind.rawValue) \(declaration.name)",
                    suggested: Self.suggestedName(declaration.name),
                ),
            )
            /* The field's name through its type, short of a trailing comma or an observer's block. */
            let end = flag.end.endLocation(converter: file.locations)
            finding.endLine = end.line
            finding.endColumn = end.column
            return finding
        }
    }

    /* The flag when a field set is a result envelope: a `Bool` outcome flag and a companion, with no exemption applying. */
    static func booleanOutcomeFlag(_ fields: [Field]) -> Field? {
        guard let flag = fields.first(where: { outcomeFlagNames.contains($0.name) && $0.isBool }) else { return nil }
        let names = Set(fields.map(\.name))
        guard !names.subtracting([flag.name]).isDisjoint(with: resultCompanionNames) else { return nil }
        guard names.isDisjoint(with: thirdPartyRequestFieldNames) else { return nil }
        guard !thirdPartyEnvelopes.contains(where: { names.isSuperset(of: $0.fields) }) else { return nil }
        guard !(names.contains("isSuccess") && names.contains("isError")) else { return nil }
        return flag
    }

    static func suggestedName(_ declarationName: String) -> String {
        var stripped = declarationName
        while let suffix = roleSuffixes.first(where: { stripped.hasSuffix($0) }) {
            stripped.removeLast(suffix.count)
        }
        return stripped.isEmpty ? declarationName : stripped
    }

    /* One field of a declaration: its name, whether it is declared exactly `Bool`, and the first and last nodes a finding on it covers. */
    struct Field {
        let name: String
        let isBool: Bool
        let start: Syntax
        let end: Syntax
    }

    /* A struct or tuple typealias, with its fields in declaration order. */
    struct Declaration {
        /* The keyword the message names the declaration by. */
        enum Kind: String {
            case structure = "struct"
            case typeAlias = "typealias"
        }

        let kind: Kind
        let name: String
        let fields: [Field]
        let isComponent: Bool
    }

    /* Collects every struct and every typealias of a labeled tuple, and the names this file's extensions make SwiftUI components. */
    final class Visitor: SyntaxVisitor {
        private(set) var declarations: [Declaration] = []
        private(set) var componentExtensionNames: Set<String> = []

        override func visit(_ node: StructDeclSyntax) -> SyntaxVisitorContinueKind {
            let fields = node.memberBlock.members.flatMap { Self.storedFields($0.decl) }
            declarations.append(
                Declaration(
                    kind: .structure,
                    name: node.name.text,
                    fields: fields,
                    isComponent: Self.declaresComponent(node.inheritanceClause),
                )
            )
            return .visitChildren
        }

        override func visit(_ node: TypeAliasDeclSyntax) -> SyntaxVisitorContinueKind {
            guard let tuple = node.initializer.value.as(TupleTypeSyntax.self) else { return .visitChildren }
            let fields = tuple.elements.compactMap { element -> Field? in
                guard let label = element.firstName, element.secondName == nil, label.tokenKind != .wildcard else {
                    return nil
                }
                return Field(
                    name: label.text,
                    isBool: Self.isBool(element.type),
                    start: Syntax(label),
                    end: Syntax(element.type),
                )
            }
            declarations.append(
                Declaration(kind: .typeAlias, name: node.name.text, fields: fields, isComponent: false)
            )
            return .visitChildren
        }

        override func visit(_ node: ExtensionDeclSyntax) -> SyntaxVisitorContinueKind {
            if Self.declaresComponent(node.inheritanceClause), let extended = Self.lastName(node.extendedType) {
                componentExtensionNames.insert(extended)
            }
            return .visitChildren
        }

        /*
         A member's stored instance properties. A binding with neither a type nor an initializer takes the type written
         after it (`var ok, failed: Bool`), as the compiler gives it; one with an initializer and no type is inferred and
         never read as `Bool`.
         */
        static func storedFields(_ member: DeclSyntax) -> [Field] {
            guard let variable = member.as(VariableDeclSyntax.self),
                !variable.modifiers.contains(where: {
                    $0.name.tokenKind == .keyword(.static) || $0.name.tokenKind == .keyword(.class)
                })
            else { return [] }
            var fields: [Field] = []
            var trailingType: TypeSyntax?
            for binding in variable.bindings.reversed() {
                let type: TypeSyntax?
                if let annotation = binding.typeAnnotation {
                    type = annotation.type
                }
                else if binding.initializer == nil {
                    type = trailingType
                }
                else {
                    type = nil
                }
                trailingType = type
                guard let identifier = binding.pattern.as(IdentifierPatternSyntax.self), isStored(binding) else {
                    continue
                }
                fields.insert(
                    Field(
                        name: identifier.identifier.text,
                        isBool: type.map(isBool) ?? false,
                        start: Syntax(binding.pattern),
                        end: binding.typeAnnotation.map { Syntax($0.type) } ?? Syntax(binding.pattern),
                    ),
                    at: 0,
                )
            }
            return fields
        }

        /* No accessor block, or one holding only observers. */
        static func isStored(_ binding: PatternBindingSyntax) -> Bool {
            guard let accessorBlock = binding.accessorBlock else { return true }
            guard case .accessors(let accessors) = accessorBlock.accessors else { return false }
            return accessors.allSatisfy {
                $0.accessorSpecifier.tokenKind == .keyword(.willSet)
                    || $0.accessorSpecifier.tokenKind == .keyword(.didSet)
            }
        }

        /* `Bool` or `Swift.Bool`, with nothing around it. */
        static func isBool(_ type: TypeSyntax) -> Bool {
            if let identifier = type.as(IdentifierTypeSyntax.self) {
                return identifier.name.text == "Bool" && identifier.genericArgumentClause == nil
            }
            guard let member = type.as(MemberTypeSyntax.self), member.name.text == "Bool",
                member.genericArgumentClause == nil,
                let base = member.baseType.as(IdentifierTypeSyntax.self)
            else { return false }
            return base.name.text == "Swift" && base.genericArgumentClause == nil
        }

        static func declaresComponent(_ clause: InheritanceClauseSyntax?) -> Bool {
            guard let clause else { return false }
            return clause.inheritedTypes.contains { inherited in
                componentNames(inherited.type).contains { ConsistencyNoBooleanOutcome.componentProtocols.contains($0) }
            }
        }

        /* The last name of each type an inheritance entry names, through `SwiftUI.View` and `View & Equatable`. */
        static func componentNames(_ type: TypeSyntax) -> [String] {
            if let composition = type.as(CompositionTypeSyntax.self) {
                return composition.elements.flatMap { componentNames($0.type) }
            }
            return lastName(type).map { [$0] } ?? []
        }

        static func lastName(_ type: TypeSyntax) -> String? {
            if let identifier = type.as(IdentifierTypeSyntax.self) {
                return identifier.name.text
            }
            return type.as(MemberTypeSyntax.self)?.name.text
        }
    }
}
