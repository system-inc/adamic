import SwiftParser
import SwiftSyntax

/*
 No calendar day rendered by ISO 8601 formatting without a time zone: `date.formatted(.iso8601.year().month().day())`,
 or a timestamp from `ISO8601DateFormatter().string(from: date)` cut down with `.prefix(10)`. ISO 8601 formatting in
 Foundation is UTC unless it is given a zone, so the result is the UTC day whether or not that was the day meant,
 and in Utah the UTC day turns over at 5 pm in winter and 6 pm in summer: an evening run gets tomorrow. The repair
 says the zone on the line: `Date.ISO8601FormatStyle(timeZone: .current).year().month().day()` (or the formatter's
 `timeZone = .current`) when the local day is meant, `timeZone: .gmt` when the UTC day is. At a site that already
 meant UTC the repair only writes down what the default left unsaid.

 It ports `nexus/consistency-no-iso-string-date-cut`, which reports `toISOString()` cut to its date part. That rule
 bans the shape rather than guessing intent, because nothing in the code tells an intended UTC day from a mistaken
 one, and this one does the same. Swift has two routes to the TypeScript shape and one it lacks:
 - The cut itself (`isoStringCutToDate`). A full ISO timestamp, cut by `.prefix(n)` with an integer literal from 1
   to 10, so the result lies inside `YYYY-MM-DD` (`slice(0, 10)`, `substring`, `substr` all map here; Swift has
   no index-pair cut that reads as plainly), or by `.split(separator: "T")` or `.components(separatedBy: "T")`
   with that one argument and then `.first` or `[0]` (`split('T')[0]`). The timestamp is `ISO8601Format()` with
   no argument, `formatted(.iso8601)`, or an `ISO8601DateFormatter`'s `string(from:)` whose options start with the
   full dashed date and go on to a time or a zone (the default, `.withInternetDateTime`). The split forms also
   need a time after a `T`, so `.withSpaceBetweenDateAndTime` leaves them alone. The string may be held in a
   `let` with an initializer declared in this file, as the TypeScript rule follows a `const`: a local, found by
   Swift's scope rules read from the tree, or a property, found through the index.
 - A date-only rendering (`isoDateWithoutTimeZone`), which TypeScript has no single call for. A
   `Date.ISO8601FormatStyle` whose fields are only date fields (`year()`, `month()`, `day()`, `weekOfYear()`, in
   any order, with any separator calls) and whose zone is the default: built from `.iso8601`,
   `Date.ISO8601FormatStyle.iso8601`, or `Date.ISO8601FormatStyle(...)` given only separator and fraction
   arguments. It is reported where it formats: as the argument of `formatted(_:)` or `ISO8601Format(_:)`, or as
   the receiver of `format(_:)`. Or an `ISO8601DateFormatter` whose `formatOptions` hold only date options
   (`.withFullDate`, `.withYear`, `.withMonth`, `.withDay`, `.withWeekOfYear`, `.withDashSeparatorInDate`),
   reported at its `string(from:)`.

 Each of those names is resolved through the index the build wrote, which is why this is a typed rule:
 `DateComponents.ISO8601FormatStyle` is spelled `.iso8601.year().month().day()` too and formats components that
 are already calendar fields, an `iso8601` of our own could build a style with any zone, and the formatter could
 be a type of ours. The style's calls must resolve to Foundation's `Date.ISO8601FormatStyle`, its base to
 Foundation's `iso8601` or initializer, the formatter to Foundation's `NSISO8601DateFormatter` (its initializer and
 `string(from:)`), and the cuts to the standard library's `prefix`, `split` and `first` and Foundation's
 `components(separatedBy:)`.

 A formatter is a class, and anyone holding it can set its `timeZone`, so a formatter is read only where the file
 shows every touch: `ISO8601DateFormatter()` called inline, or a local `let` initialized that way whose every use
 in its block is one `formatOptions` assignment (a statement of the declaring block, an array literal of options
 or one option) followed by `string(from:)` calls. A `timeZone` assignment, a second assignment, an argument, a
 return, a capture or any other use, and the formatter is not read.

 What it accepts as safe, and why: a style or formatter given a zone (`Date.ISO8601FormatStyle(timeZone: .current)`,
 `.iso8601Date(timeZone:)`, `formatter.timeZone = ...`, `ISO8601DateFormatter.string(from:timeZone:formatOptions:)`),
 because the zone is said, UTC included; a full timestamp left whole (`.iso8601`, `ISO8601Format()`, a style with
 `time(includingFractionalSeconds:)`), because it carries its time and its `Z`; a rendering with a zone field
 (`.timeZone(separator:)`, `.withTimeZone`), which names its zone; a cut past the date part (`prefix(16)`) or of
 the time half (`[1]`, `.last`); a string touched before the cut (`.lowercased().prefix(10)`); a `var`; a style
 held anywhere (see below); `DateFormatter` with a `yyyy-MM-dd` format, which is local time by default.

 Known misses, every one a finding not made and never one invented:
 - A style held in a binding before it formats (`let style = Date.ISO8601FormatStyle().year().month().day()`,
   then `date.formatted(style)`), and a date-only style handed to anything but the three formatting calls above,
   SwiftUI's `Text(_:format:)` and `TextField(value:format:)` among them, or called with no receiver inside an
   extension of `Date` (`formatted(.iso8601.year())` with no `date.` before it).
 - A formatter that is a property, a static, or a local that escapes its block, including the common
   closure-initialized `static let dayFormatter: ISO8601DateFormatter = { ... }()`: another file may set its
   `timeZone`. A formatter whose options are built any other way (`insert`, a variable, a qualified name the rule
   does not read).
 - Cuts the rule does not read: `dropFirst(n).prefix(m)`, `prefix(upTo:)`, `prefix(while:)`, index arithmetic,
   `split(separator: "T", maxSplits: 1)` or any split with more arguments (the TypeScript rule leaves a split
   with a limit alone too), and a timestamp from a style chain with explicit fields
   (`.iso8601.year().month().day().time(includingFractionalSeconds: false)`) or from `format(_:)`.
 - A string held in a `let` of another file, a `var`, or any binding but a `let` with an initializer (`if let`,
   `guard let`, a tuple pattern).
 - Parsing a date-only ISO string (`Date(text, strategy: .iso8601.year().month().day())`) reads the day as UTC
   midnight, the same trap the other way round, and is not this shape; the TypeScript rule judges no parsing either.
 */
public struct ConsistencyNoIsoStringDateCut: TypedFileRule {
    public let name = "cohere-swift/consistency-no-iso-string-date-cut"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Foundation's declarations, by the symbol the index gives each, read from a real build's index. */
    static let styleMemberPrefix = "s:10Foundation4DateV18ISO8601FormatStyleV"
    static let dateFormatted = "s:10Foundation4DateV9formattedy12FormatOutputQzxAA0D5StyleRzAC0D5InputRtzlF"
    static let dateIso8601Format = "s:10Foundation4DateV13ISO8601FormatySSAC0cD5StyleVF"
    static let styleFormat = "s:10Foundation4DateV18ISO8601FormatStyleV6formatySSACF"
    static let formatterInitializer = "c:objc(cs)NSISO8601DateFormatter(im)init"
    static let formatterString = "c:objc(cs)NSISO8601DateFormatter(im)stringFromDate:"
    static let formatterOptions = "c:objc(cs)NSISO8601DateFormatter(py)formatOptions"
    static let componentsSeparatedByPrefix = "s:Sy10FoundationE10components11separatedBy"

    /* What a rendering holds, as fields: what decides whether it is a date alone and where its date part ends. */
    static let dateFields: Set<String> = ["year", "month", "day", "weekOfYear"]
    static let internetDateTime: Set<String> = ["year", "month", "day", "dashSeparatorInDate", "time", "timeZone"]

    /* The style's chain calls and the fields each adds; separators add none. */
    static let styleCalls: [String: Set<String>] = [
        "year": ["year"],
        "month": ["month"],
        "day": ["day"],
        "weekOfYear": ["weekOfYear"],
        "time": ["time"],
        "timeZone": ["timeZone"],
        "dateSeparator": [],
        "dateTimeSeparator": [],
        "timeSeparator": [],
        "timeZoneSeparator": [],
    ]

    /* The initializer labels that leave the zone at its default. `from` (a decoder) can carry any zone. */
    static let styleInitializerLabels: Set<String> = [
        "dateSeparator", "dateTimeSeparator", "timeSeparator", "timeZoneSeparator", "includingFractionalSeconds",
    ]

    /* `ISO8601DateFormatter.Options` and the fields each adds. */
    static let formatterOptionFields: [String: Set<String>] = [
        "withInternetDateTime": internetDateTime,
        "withFullDate": ["year", "month", "day", "dashSeparatorInDate"],
        "withFullTime": ["time", "timeZone"],
        "withYear": ["year"],
        "withMonth": ["month"],
        "withWeekOfYear": ["weekOfYear"],
        "withDay": ["day"],
        "withTime": ["time"],
        "withTimeZone": ["timeZone"],
        "withSpaceBetweenDateAndTime": ["spaceBetweenDateAndTime"],
        "withDashSeparatorInDate": ["dashSeparatorInDate"],
        "withColonSeparatorInTime": [],
        "withColonSeparatorInTimeZone": [],
        "withFractionalSeconds": [],
    ]

    /* Names Swift binds without a pattern (`catch`'s `error`, an observer's `newValue` and `oldValue`); a local of one of these names is not read. */
    static let implicitNames: Set<String> = ["error", "newValue", "oldValue"]

    enum Shape {
        case cut
        case dateOnlyStyle
        case dateOnlyFormatter

        var message: RuleMessages.Message {
            switch self {
                case .cut:
                    RuleMessages.ConsistencyNoIsoStringDateCut.isoStringCutToDate()
                case .dateOnlyStyle:
                    RuleMessages.ConsistencyNoIsoStringDateCut.isoDateWithoutTimeZone(renderer: .formatStyle)
                case .dateOnlyFormatter:
                    RuleMessages.ConsistencyNoIsoStringDateCut.isoDateWithoutTimeZone(renderer: .dateFormatter)
            }
        }
    }

    /* Every spelling the rule reads names ISO 8601, in one casing or the other. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("ISO8601") || file.source.contains("iso8601")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        let visitor = Visitor(reader: Reader(file: file, symbols: symbols))
        visitor.walk(file.tree)
        return visitor.found.map { node, shape in
            file.finding(at: node, rule: name, message: shape.message)
        }
    }

    /* A rendering's fields, and what they make of its text. */
    struct Rendering {
        var fields: Set<String>

        /* A date and nothing after it: no time, and no zone that would name itself. */
        var isDateOnly: Bool {
            !fields.isDisjoint(with: ConsistencyNoIsoStringDateCut.dateFields) && !fields.contains("time")
                && !fields.contains("timeZone")
        }

        /* The text opens with `YYYY-MM-DD` and goes on past it. */
        var opensWithFullDate: Bool {
            fields.isSuperset(of: ["year", "month", "day", "dashSeparatorInDate"]) && !fields.contains("weekOfYear")
                && !isDateOnly
        }

        /* `YYYY-MM-DDT...`, so a split on `T` puts the date first. */
        var separatesTimeWithT: Bool {
            opensWithFullDate && fields.contains("time") && !fields.contains("spaceBetweenDateAndTime")
        }
    }

    /* Reads names through the index and bindings through the tree. */
    final class Reader {
        let file: ParsedFile
        let symbols: FileSymbols
        /* This file's `let` bindings with an initializer, by the symbol the index gives their name, built on first use. */
        private var declaredHere: [String: PatternBindingSyntax]?

        init(file: ParsedFile, symbols: FileSymbols) {
            self.file = file
            self.symbols = symbols
        }

        /* Whether a name written here resolves to a declaration the predicate accepts. More than one may be recorded at a place (a type and its initializer), and any may answer. */
        func resolves(_ token: TokenSyntax, _ predicate: (FileSymbols.Occurrence) -> Bool) -> Bool {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            return symbols.occurrences(line: location.line, column: location.column).contains {
                $0.isReference && !$0.isImplicit && predicate($0)
            }
        }

        func resolves(_ token: TokenSyntax, to symbol: String) -> Bool {
            resolves(token) { $0.symbol == symbol }
        }

        // MARK: The shapes

        /* `date.formatted(<style>)` or `date.ISO8601Format(<style>)` with a date-only UTC style, or `<style>.format(date)`. */
        func isDateOnlyStyleFormatting(_ call: FunctionCallExprSyntax) -> Bool {
            guard let member = call.calledExpression.as(MemberAccessExprSyntax.self), let base = member.base,
                call.trailingClosure == nil,
                member.declName.argumentNames == nil, call.arguments.count == 1, let argument = call.arguments.first,
                argument.label == nil
            else { return false }
            switch member.declName.baseName.text {
                case "formatted":
                    return resolves(member.declName.baseName, to: ConsistencyNoIsoStringDateCut.dateFormatted)
                        && styleRendering(argument.expression)?.isDateOnly == true
                case "ISO8601Format":
                    return resolves(member.declName.baseName, to: ConsistencyNoIsoStringDateCut.dateIso8601Format)
                        && styleRendering(argument.expression)?.isDateOnly == true
                case "format":
                    return resolves(member.declName.baseName, to: ConsistencyNoIsoStringDateCut.styleFormat)
                        && styleRendering(base)?.isDateOnly == true
                default:
                    return false
            }
        }

        /* `<formatter>.string(from: date)` with a formatter whose options are a date alone. */
        func isDateOnlyFormatterString(_ call: FunctionCallExprSyntax) -> Bool {
            guard let formatter = formatterOfString(call) else { return false }
            return formatterRendering(formatter)?.isDateOnly == true
        }

        /* `<timestamp>.prefix(n)` with `n` from 1 to 10. */
        func isPrefixCut(_ call: FunctionCallExprSyntax) -> Bool {
            guard let member = call.calledExpression.as(MemberAccessExprSyntax.self), let base = member.base,
                member.declName.baseName.text == "prefix",
                member.declName.argumentNames == nil, call.trailingClosure == nil, call.arguments.count == 1,
                let argument = call.arguments.first,
                argument.label == nil, let literal = argument.expression.as(IntegerLiteralExprSyntax.self),
                let length = Int(literal.literal.text),
                (1...10).contains(length), resolves(member.declName.baseName, { $0.isStandardLibrary })
            else { return false }
            return timestampRendering(base, depth: 0)?.opensWithFullDate == true
        }

        /* `<timestamp>.split(separator: "T").first` or `.components(separatedBy: "T").first`. */
        func isFirstOfSplit(_ member: MemberAccessExprSyntax) -> Bool {
            guard member.declName.baseName.text == "first", member.declName.argumentNames == nil,
                let base = member.base,
                resolves(member.declName.baseName, { $0.isStandardLibrary })
            else { return false }
            return isSplitOnT(base)
        }

        /* `<timestamp>.split(separator: "T")[0]` or `.components(separatedBy: "T")[0]`. */
        func isFirstOfSplit(_ subscripted: SubscriptCallExprSyntax) -> Bool {
            guard subscripted.trailingClosure == nil, subscripted.arguments.count == 1,
                let argument = subscripted.arguments.first, argument.label == nil,
                argument.expression.as(IntegerLiteralExprSyntax.self)?.literal.text == "0"
            else { return false }
            return isSplitOnT(subscripted.calledExpression)
        }

        func isSplitOnT(_ expression: ExprSyntax) -> Bool {
            guard let call = expression.as(FunctionCallExprSyntax.self),
                let member = call.calledExpression.as(MemberAccessExprSyntax.self), let base = member.base,
                call.trailingClosure == nil, call.arguments.count == 1, let argument = call.arguments.first,
                let literal = argument.expression.as(StringLiteralExprSyntax.self),
                literal.representedLiteralValue == "T"
            else { return false }
            switch (member.declName.baseName.text, argument.label?.text) {
                case ("split", "separator"):
                    guard resolves(member.declName.baseName, { $0.isStandardLibrary }) else { return false }
                case ("components", "separatedBy"):
                    guard
                        resolves(
                            member.declName.baseName,
                            { $0.symbol.hasPrefix(ConsistencyNoIsoStringDateCut.componentsSeparatedByPrefix) },
                        )
                    else { return false }
                default:
                    return false
            }
            return timestampRendering(base, depth: 0)?.separatesTimeWithT == true
        }

        // MARK: What a value renders

        /*
         A `Date.ISO8601FormatStyle` built inline with the default zone: Foundation's chain calls, read outward from the
         base, over Foundation's `iso8601` or an initializer given only separator and fraction arguments.
         */
        func styleRendering(_ expression: ExprSyntax) -> Rendering? {
            var fields: Set<String> = []
            var current = Reader.unwrapped(expression)
            while let call = current.as(FunctionCallExprSyntax.self),
                let member = call.calledExpression.as(MemberAccessExprSyntax.self), let base = member.base,
                let added = ConsistencyNoIsoStringDateCut.styleCalls[member.declName.baseName.text]
            {
                let callName = member.declName.baseName.text
                guard call.trailingClosure == nil,
                    resolves(
                        member.declName.baseName,
                        {
                            $0.symbol.hasPrefix(
                                ConsistencyNoIsoStringDateCut.styleMemberPrefix + "\(callName.utf8.count)\(callName)"
                            )
                        },
                    )
                else { return nil }
                fields.formUnion(added)
                current = Reader.unwrapped(base)
            }
            return isDefaultZoneStyle(current) ? Rendering(fields: fields) : nil
        }

        func isDefaultZoneStyle(_ expression: ExprSyntax) -> Bool {
            if let member = expression.as(MemberAccessExprSyntax.self) {
                return member.declName.baseName.text == "iso8601" && member.declName.argumentNames == nil
                    && resolves(member.declName.baseName, { $0.symbol.hasPrefix("s:10Foundation") })
            }
            guard let call = expression.as(FunctionCallExprSyntax.self), call.trailingClosure == nil,
                call.arguments.allSatisfy({
                    $0.label.map { ConsistencyNoIsoStringDateCut.styleInitializerLabels.contains($0.text) } ?? false
                })
            else { return false }
            let typeName: TokenSyntax
            if let member = call.calledExpression.as(MemberAccessExprSyntax.self) {
                typeName = member.declName.baseName
            }
            else if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self) {
                typeName = reference.baseName
            }
            else {
                return false
            }
            return typeName.text == "ISO8601FormatStyle"
                && resolves(
                    typeName,
                    {
                        $0.symbol.hasPrefix(ConsistencyNoIsoStringDateCut.styleMemberPrefix)
                            && $0.symbol.hasSuffix("cfc")
                    },
                )
        }

        /*
         A full ISO timestamp as text: `ISO8601Format()`, `formatted(.iso8601)`, an `ISO8601DateFormatter`'s
         `string(from:)`, or a `let` this file declares with one of those as its initializer.
         */
        func timestampRendering(_ expression: ExprSyntax, depth: Int) -> Rendering? {
            let expression = Reader.unwrapped(expression)
            if let reference = expression.as(DeclReferenceExprSyntax.self) {
                guard depth < 4, reference.argumentNames == nil, let initializer = heldValue(reference) else {
                    return nil
                }
                return timestampRendering(initializer, depth: depth + 1)
            }
            guard let call = expression.as(FunctionCallExprSyntax.self),
                let member = call.calledExpression.as(MemberAccessExprSyntax.self), member.base != nil,
                call.trailingClosure == nil, member.declName.argumentNames == nil
            else { return nil }
            switch member.declName.baseName.text {
                case "ISO8601Format":
                    guard call.arguments.isEmpty,
                        resolves(member.declName.baseName, to: ConsistencyNoIsoStringDateCut.dateIso8601Format)
                    else { return nil }
                    return Rendering(fields: ConsistencyNoIsoStringDateCut.internetDateTime)
                case "formatted":
                    guard call.arguments.count == 1, let argument = call.arguments.first, argument.label == nil,
                        let style = argument.expression.as(MemberAccessExprSyntax.self), style.base == nil,
                        style.declName.baseName.text == "iso8601",
                        style.declName.argumentNames == nil,
                        resolves(style.declName.baseName, { $0.symbol.hasPrefix("s:10Foundation") }),
                        resolves(member.declName.baseName, to: ConsistencyNoIsoStringDateCut.dateFormatted)
                    else { return nil }
                    return Rendering(fields: ConsistencyNoIsoStringDateCut.internetDateTime)
                case "string":
                    return formatterOfString(call).flatMap(formatterRendering)
                default:
                    return nil
            }
        }

        /* The receiver of Foundation's `ISO8601DateFormatter.string(from:)`. */
        func formatterOfString(_ call: FunctionCallExprSyntax) -> ExprSyntax? {
            guard let member = call.calledExpression.as(MemberAccessExprSyntax.self), let base = member.base,
                member.declName.baseName.text == "string",
                member.declName.argumentNames == nil, call.trailingClosure == nil, call.arguments.count == 1,
                call.arguments.first?.label?.text == "from",
                resolves(member.declName.baseName, to: ConsistencyNoIsoStringDateCut.formatterString)
            else { return nil }
            return base
        }

        /* A formatter whose every touch this file shows: `ISO8601DateFormatter()` inline, or a confined local. */
        func formatterRendering(_ expression: ExprSyntax) -> Rendering? {
            let expression = Reader.unwrapped(expression)
            if isNewFormatter(expression) {
                return Rendering(fields: ConsistencyNoIsoStringDateCut.internetDateTime)
            }
            guard let reference = expression.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil else {
                return nil
            }
            return confinedFormatter(reference)
        }

        /* `ISO8601DateFormatter()`, resolved to Foundation's initializer. */
        func isNewFormatter(_ expression: ExprSyntax) -> Bool {
            guard let call = Reader.unwrapped(expression).as(FunctionCallExprSyntax.self), call.arguments.isEmpty,
                call.trailingClosure == nil
            else { return false }
            let typeName: TokenSyntax
            if let member = call.calledExpression.as(MemberAccessExprSyntax.self) {
                typeName = member.declName.baseName
            }
            else if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self) {
                typeName = reference.baseName
            }
            else {
                return false
            }
            return typeName.text == "ISO8601DateFormatter"
                && resolves(typeName, to: ConsistencyNoIsoStringDateCut.formatterInitializer)
        }

        /*
         A local `let name = ISO8601DateFormatter()` whose every use in its block, after it, is one `name.formatOptions =`
         statement of that block and then `name.string(from:)` calls. Its options are the assignment's, or the default.
         */
        func confinedFormatter(_ reference: DeclReferenceExprSyntax) -> Rendering? {
            guard let local = localLet(reference), local.isInFunctionBody,
                let initializer = local.binding.initializer?.value, isNewFormatter(initializer)
            else {
                return nil
            }
            let name = reference.baseName.text
            var fields = ConsistencyNoIsoStringDateCut.internetDateTime
            var assignmentEnd: AbsolutePosition?
            var uses: [AbsolutePosition] = []
            for use in Reader.references(named: name, in: Syntax(local.block))
            where use.position >= local.binding.endPosition {
                guard let member = use.parent?.as(MemberAccessExprSyntax.self), member.base?.id == use.id else {
                    return nil
                }
                if member.declName.baseName.text == "formatOptions", member.declName.argumentNames == nil,
                    let elements = member.parent?.as(ExprListSyntax.self),
                    let sequence = elements.parent?.as(SequenceExprSyntax.self),
                    sequence.parent?.as(CodeBlockItemSyntax.self)?.parent?.id == local.block.id, elements.count == 3,
                    elements.first?.id == member.id, elements.dropFirst().first?.is(AssignmentExprSyntax.self) == true,
                    let value = elements.last,
                    assignmentEnd == nil,
                    resolves(member.declName.baseName, to: ConsistencyNoIsoStringDateCut.formatterOptions),
                    let options = Reader.optionFields(value)
                {
                    fields = options
                    assignmentEnd = sequence.endPosition
                }
                else if member.declName.baseName.text == "string",
                    let call = member.parent?.as(FunctionCallExprSyntax.self),
                    call.calledExpression.id == member.id, formatterOfString(call) != nil
                {
                    uses.append(call.position)
                }
                else {
                    return nil
                }
            }
            if let assignmentEnd, uses.contains(where: { $0 < assignmentEnd }) {
                return nil
            }
            return Rendering(fields: fields)
        }

        /* An array literal of `ISO8601DateFormatter.Options` members, or one member, as fields. */
        static func optionFields(_ value: ExprSyntax) -> Set<String>? {
            let members: [ExprSyntax]
            if let array = value.as(ArrayExprSyntax.self) {
                members = array.elements.map(\.expression)
            }
            else {
                members = [value]
            }
            var fields: Set<String> = []
            for member in members {
                guard let access = member.as(MemberAccessExprSyntax.self), access.base == nil,
                    access.declName.argumentNames == nil,
                    let added = ConsistencyNoIsoStringDateCut.formatterOptionFields[access.declName.baseName.text]
                else { return nil }
                fields.formUnion(added)
            }
            return fields
        }

        // MARK: Bindings

        /* The initializer of the `let` a name reads: a local found by scope, or a binding of this file the index names. */
        func heldValue(_ reference: DeclReferenceExprSyntax) -> ExprSyntax? {
            if let local = localLet(reference) {
                return local.binding.initializer?.value
            }
            if Reader.hasLocalBinder(named: reference.baseName.text, above: Syntax(reference)) {
                return nil
            }
            let location = file.locations.location(for: reference.baseName.positionAfterSkippingLeadingTrivia)
            guard let resolved = symbols.reference(line: location.line, column: location.column),
                let binding = bindingsDeclaredHere()[resolved.symbol]
            else {
                return nil
            }
            return binding.initializer?.value
        }

        /* `let` bindings with an initializer and a plain name, keyed by the symbol the index records at the name. */
        func bindingsDeclaredHere() -> [String: PatternBindingSyntax] {
            if let declaredHere {
                return declaredHere
            }
            var found: [String: PatternBindingSyntax] = [:]
            for token in file.tree.tokens(viewMode: .sourceAccurate) {
                guard let pattern = token.parent?.as(IdentifierPatternSyntax.self),
                    let binding = pattern.parent?.as(PatternBindingSyntax.self),
                    binding.initializer != nil, binding.accessorBlock == nil, Reader.isLet(binding)
                else { continue }
                let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
                for occurrence in symbols.occurrences(line: location.line, column: location.column)
                where !occurrence.isReference {
                    found[occurrence.symbol] = binding
                }
            }
            declaredHere = found
            return found
        }

        struct LocalLet {
            var binding: PatternBindingSyntax
            var block: CodeBlockItemListSyntax
            /* Declared in a function, closure, accessor or case body, not at a file's top level, where it is a global. */
            var isInFunctionBody: Bool
        }

        /*
         The nearest `let name = ...` before this use in an enclosing statement list, when nothing else in that list
         binds the same name: no other pattern, parameter, closure parameter or capture, so no binding between the
         declaration and the use can shadow it.
         */
        func localLet(_ reference: DeclReferenceExprSyntax) -> LocalLet? {
            let name = reference.baseName.text
            guard reference.argumentNames == nil, !ConsistencyNoIsoStringDateCut.implicitNames.contains(name) else {
                return nil
            }
            var child = Syntax(reference)
            while let scope = child.parent {
                defer { child = scope }
                if Reader.isTypeScope(scope) {
                    return nil
                }
                guard let list = scope.as(CodeBlockItemListSyntax.self) else { continue }
                for statement in list.reversed() where statement.endPosition <= child.position {
                    guard let variable = statement.item.as(VariableDeclSyntax.self) else { continue }
                    for binding in variable.bindings
                    where binding.pattern.as(IdentifierPatternSyntax.self)?.identifier.text == name {
                        guard variable.bindingSpecifier.tokenKind == .keyword(.let), binding.initializer != nil,
                            Reader.binders(named: name, in: Syntax(list)) == 1
                        else { return nil }
                        let isInFunctionBody = !(list.parent?.is(SourceFileSyntax.self) ?? true)
                        return LocalLet(binding: binding, block: list, isInFunctionBody: isInFunctionBody)
                    }
                }
            }
            return nil
        }

        /* Whether a statement list or parameter list between this use and its type or file binds the name, so the index's answer, which skips locals, would not be the binding the use reads. */
        static func hasLocalBinder(named name: String, above start: Syntax) -> Bool {
            var child = start
            while let scope = child.parent {
                defer { child = scope }
                if isTypeScope(scope) || scope.is(SourceFileSyntax.self) {
                    return false
                }
                if scope.is(FunctionDeclSyntax.self) || scope.is(InitializerDeclSyntax.self)
                    || scope.is(ClosureExprSyntax.self) || scope.is(AccessorDeclSyntax.self)
                    || scope.is(SubscriptDeclSyntax.self) || scope.is(CodeBlockItemListSyntax.self)
                {
                    if binders(named: name, in: scope) > 0 {
                        return true
                    }
                }
            }
            return false
        }

        static func isTypeScope(_ node: Syntax) -> Bool {
            node.is(StructDeclSyntax.self) || node.is(ClassDeclSyntax.self) || node.is(EnumDeclSyntax.self)
                || node.is(ActorDeclSyntax.self)
                || node.is(ExtensionDeclSyntax.self) || node.is(ProtocolDeclSyntax.self)
        }

        /* How many places under a node bind a name: patterns, parameters of every kind, and closure captures. */
        static func binders(named name: String, in node: Syntax) -> Int {
            var count = 0
            for token in node.tokens(viewMode: .sourceAccurate) where token.text == name {
                guard let parent = token.parent else { continue }
                if parent.is(IdentifierPatternSyntax.self) || parent.is(ClosureShorthandParameterSyntax.self)
                    || parent.is(ClosureCaptureSyntax.self)
                    || parent.is(AccessorParametersSyntax.self)
                {
                    count += 1
                }
                else if let parameter = parent.as(FunctionParameterSyntax.self),
                    (parameter.secondName ?? parameter.firstName).id == token.id
                {
                    count += 1
                }
                else if let parameter = parent.as(ClosureParameterSyntax.self),
                    (parameter.secondName ?? parameter.firstName).id == token.id
                {
                    count += 1
                }
            }
            return count
        }

        /* Every bare use of a name under a node: a reference, not a member's name after a dot. */
        static func references(named name: String, in node: Syntax) -> [DeclReferenceExprSyntax] {
            node.tokens(viewMode: .sourceAccurate).compactMap { token in
                guard token.text == name, let reference = token.parent?.as(DeclReferenceExprSyntax.self),
                    reference.argumentNames == nil,
                    reference.parent?.as(MemberAccessExprSyntax.self)?.declName.id != reference.id
                else { return nil }
                return reference
            }
        }

        static func isLet(_ binding: PatternBindingSyntax) -> Bool {
            binding.parent?.parent?.as(VariableDeclSyntax.self)?.bindingSpecifier.tokenKind == .keyword(.let)
        }

        /* Through parentheses around one unlabeled value. */
        static func unwrapped(_ expression: ExprSyntax) -> ExprSyntax {
            var current = expression
            while let tuple = current.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                current = only.expression
            }
            return current
        }
    }

    /* Each formatting call, cut and `[0]` the rule reports, with its shape. */
    final class Visitor: SyntaxVisitor {
        let reader: Reader
        private(set) var found: [(Syntax, Shape)] = []

        init(reader: Reader) {
            self.reader = reader
            super.init(viewMode: .sourceAccurate)
        }

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            if reader.isDateOnlyStyleFormatting(node) {
                found.append((Syntax(node), .dateOnlyStyle))
            }
            else if reader.isDateOnlyFormatterString(node) {
                found.append((Syntax(node), .dateOnlyFormatter))
            }
            else if reader.isPrefixCut(node) {
                found.append((Syntax(node), .cut))
            }
            return .visitChildren
        }

        override func visit(_ node: MemberAccessExprSyntax) -> SyntaxVisitorContinueKind {
            if reader.isFirstOfSplit(node) {
                found.append((Syntax(node), .cut))
            }
            return .visitChildren
        }

        override func visit(_ node: SubscriptCallExprSyntax) -> SyntaxVisitorContinueKind {
            if reader.isFirstOfSplit(node) {
                found.append((Syntax(node), .cut))
            }
            return .visitChildren
        }
    }
}
