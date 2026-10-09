import Foundation
import SwiftOperators
import SwiftParser
import SwiftSyntax

/*
 No value written between the single quotes of a SQL string: `"SELECT count(*) FROM items WHERE kind = '\(kind)'"`.
 A quote in the value ends the string early, so the query breaks on ordinary input like `O'Brien`, and a crafted
 value runs as SQL. The repair is a bound parameter (`WHERE kind = ?` with `[.text(kind)]` beside the query), which
 moves the value out of the statement's text.

 It ports our TypeScript rule `nexus/security-no-interpolated-sql-string`, whose doc comment is the definition
 this follows, and shares its name. A TypeScript template literal is a Swift string literal with interpolations,
 a `${value}` is a `\(value)`, and the template's texts are the literal's segments as the program sees them
 (escapes resolved, a multi-line literal's lines joined), read by the same reader, character for character:

   - A statement: the text opens, after whitespace and `(`, with an uppercase `SELECT`, `INSERT`, `UPDATE`,
     `DELETE`, `REPLACE`, `WITH`, `SHOW`, `EXPLAIN`, `PRAGMA`, `WHERE`, `AND`, `OR`, `SET`, `VALUES` or `HAVING`,
     and, unless that is `WHERE` or `HAVING`, a second uppercase SQL keyword stands outside its strings and
     comments. The whole literal is read as SQL (single-quoted strings with `''` as one quote, double-quoted and
     backquoted names, `--` and `/* */` comments), and an interpolation is a finding when it sits inside a
     single-quoted string that closes in the same literal and opens where a value goes: after a comparison,
     `||`, `+`, `(`, `,`, or `LIKE`, `GLOB`, `REGEXP`, `RLIKE`, `BETWEEN`, `AND`, `SELECT`, `WHEN`, `THEN`,
     `ELSE`, `ESCAPE`, `DEFAULT` or `INTERVAL`.
   - A `LIKE` pattern in a fragment that is not a statement: an uppercase `LIKE` and a single quote, in a literal
     that also carries an uppercase `AND`, `OR`, `NOT`, `WHERE`, `SELECT`, `FROM` or `ESCAPE` and has an even
     number of quotes before the `LIKE`.

 The reader gives up on the whole literal, rather than guess, on a backslash inside quoted text, a `#` outside a
 string, `--` not followed by whitespace, a dollar quote, and a literal that ends inside a string, a name or a
 block comment. Keywords count in uppercase only: lowercase SQL is missed so that English is never read as SQL.

 The shape of the text decides, not the call it is handed to, as in TypeScript. In Presence the statement reaches
 SQLite through `SQLiteConnection` (`execute`, `rows`, `row`, `value`, each taking the statement as one string),
 and the store's `verify` passes it through a local closure first (`count("SELECT ... '\(kind)' ...")`), which a
 gate on the wrapper's calls would miss. The reader's gates are what keep a sentence that merely contains SQL words
 out. A literal whose type is a SQL library's own (one whose interpolation binds each value, the analog of a
 tagged template, which TypeScript skips) cannot be told from a `String` here and is read the same; a value
 quoted inside one is broken anyway, since it binds as the text `'?'`.

 Which values can hold a quote is the TypeScript rule's type gate, so this is a typed rule. The build's index names
 the declaration behind every name but a local, and the toolchain's demangler prints it with its type
 (`AhraOSPresence.SQLValue.text : Swift.String?`); a local's type is read from its own binding in this file. A
 value is a finding only when its type is shown to be `String`, `Substring`, `Character` or `Any`, or an optional
 of one, the members of TypeScript's `string`, `any` and `unknown`, read through `try`, `await`, `!`, `?.`, `??`,
 `+` and `as`:

   - A name bound in this file: a parameter annotated with one of those types (the annotation's name resolved to
     the standard library's); a `let` or `var` annotated with one, or initialized with a value shown to be one; an
     `if let`, `guard let` or `while let` of an optional one; a `for` loop variable over an array literal of string
     literals (`for location in ["Store", "Dataset"]`) or over a collection whose element is shown, through the
     standard library's `sorted`, `reversed`, `filter`, `shuffled`, `prefix`, `suffix`, `dropFirst`, `dropLast`,
     `lazy`, `enumerated`, `keys`, `values`, `Array(...)`, `Set(...)` and `+`, a tuple pattern taking a
     dictionary's key and value or a tuple's members in order.
   - A declaration the index names: a function's or initializer's result, from any module (`String(describing:)`,
     `store.create(...)`); a property declared outside this package (`url.path`); a property of ours that is
     optional or a collection; a non-optional property of ours only when it is declared in this file, read the way
     a local is.
   - A string literal that interpolates, or whose own text holds `'` or `\`.

 Safe, and silent:

   - An integer, a float, a boolean, and every type not listed above. A struct, class or array is not reported,
     as TypeScript does not report an object type.
   - A closed enum's `rawValue` (an enum's own, as its symbol says), and an enum case interpolated whole, which
     prints its case name: TypeScript's literal union.
   - A `let` initialized with a string literal that holds no quote, local or a property in this file
     (`let path = "Genshin/Motions/Clip.vrma"`, `static let table = "items"`): a compile-time name, TypeScript's
     `const` with a literal type. A `var` initialized the same way is a `String`, as a TypeScript `let` is.
   - A value inside double quotes or backquotes (an identifier), and one outside any quotes (a table name, a clause
     of `?` placeholders, `LIMIT \(count)`): the TypeScript rule reports neither.
   - The one escape TypeScript accepts, written where the reader can see it: the last two steps of a
     `replacingOccurrences(of:with:)` or `replacing(_:with:)` chain (Foundation's or the standard library's) double
     every backslash and every single quote, in either order, with string literal arguments (or a regex literal
     for `replacing`), on the value itself or in the initializer of the local `let` it names. With both doubled no
     input closes the string on any engine. Doubling quotes alone is reported, as in TypeScript: correct on
     SQLite, open on MySQL, and the text does not say which engine runs it.

 Known misses, each a place this rule cannot be sure and so says nothing:

   - A value whose type the file and the index do not show: an unannotated closure parameter, a generic parameter
     (TypeScript reports an unconstrained one), a binding from a `switch`, `catch` or `if case` pattern, a local
     closure's result, a value chosen by `?:`, `if` or `switch`, a subscript other than a dictionary's, an element
     of a collection whose type is not shown, and a member read from any of those.
   - A non-optional `String` property of ours declared in another file, which could be a compile-time constant
     (`static let table = "items"`) or an open value, and nothing the index carries says which. So is a struct's
     `rawValue` declared elsewhere, which TypeScript would report as a branded string.
   - An enum whose raw value itself holds a quote, which TypeScript would report as a literal holding one.
   - An interpolation with a label or more than one argument (`\(value, format: ...)`), which a custom
     interpolation decides.
   - Everything the TypeScript rule misses: SQL assembled with `+` from separate literals, a quote that opens in
     one literal and closes in another, a value inside a double-quoted string, lowercase SQL, and a statement
     opening with an interpolation. `ATTACH '\(path)'` is one too: `ATTACH` is not among the TypeScript rule's
     value keywords, so Presence's attach of its payload database (whose path doubles its quotes, correct for
     SQLite) is left alone.

 A typed rule reports all its shapes together, so a file no source of symbols describes is reported as unchecked,
 never as clean. No fix: the repair spans the statement and the parameters beside it.
 */
public struct SecurityNoInterpolatedSqlString: TypedFileRule {
    public let name = "cohere-swift/security-no-interpolated-sql-string"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /*
     Every finding is an interpolation (`\(` or `\#(`, so a backslash) inside single quotes, in a literal that
     opens with a leading keyword or carries `LIKE`.
     */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("'") && file.source.contains("\\")
            && (file.source.contains("LIKE") || SqlReading.leadingKeywords.contains { file.source.contains($0) })
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        /* Folded so `a ?? b` and `a + b` read as one expression each. Folding moves no token, so positions in the copy are positions in the file. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        guard !visitor.found.isEmpty else { return [] }
        /* A toolchain without its demangler leaves every declaration unread, a miss and never a finding. */
        let values = ValueTypes(file: file, symbols: symbols, demangler: try? SwiftDemangler.shared())
        return visitor.found.compactMap { value in
            guard values.canHoldQuote(value), !values.isEscaped(value) else { return nil }
            return file.finding(
                at: value,
                rule: name,
                message: RuleMessages.SecurityNoInterpolatedSqlString.interpolatedSqlString(),
            )
        }
    }

    /* Every interpolated value the SQL reading places inside a single-quoted string. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [ExprSyntax] = []

        override func visit(_ node: StringLiteralExprSyntax) -> SyntaxVisitorContinueKind {
            guard [.stringQuote, .multilineStringQuote].contains(node.openingQuote.tokenKind),
                let pieces = Self.pieces(of: node)
            else {
                return .visitChildren
            }
            for index in SqlReading.quotedSpans(pieces.texts) {
                /* `\(value)` alone; a label or a second argument is a custom interpolation's to decide. */
                let arguments = pieces.interpolations[index].expressions
                guard arguments.count == 1, let argument = arguments.first, argument.label == nil else { continue }
                found.append(argument.expression)
            }
            return .visitChildren
        }

        /*
         The literal's texts as the program sees them, one more than its interpolations, and the interpolations.
         Each segment is resolved on its own through a literal holding only it, so its escapes read as the
         compiler reads them. Nil when there is no interpolation or a segment does not resolve.
         */
        static func pieces(
            of literal: StringLiteralExprSyntax
        ) -> (texts: [[UInt8]], interpolations: [ExpressionSegmentSyntax])? {
            var texts: [[UInt8]] = [[]]
            var interpolations: [ExpressionSegmentSyntax] = []
            for segment in literal.segments {
                switch segment {
                    case .stringSegment(let text):
                        let alone = StringLiteralExprSyntax(
                            openingPounds: literal.openingPounds,
                            openingQuote: literal.openingQuote,
                            segments: StringLiteralSegmentListSyntax([.stringSegment(text)]),
                            closingQuote: literal.closingQuote,
                            closingPounds: literal.closingPounds,
                        )
                        guard let value = alone.representedLiteralValue else { return nil }
                        texts[texts.count - 1].append(contentsOf: Array(value.utf8))
                    case .expressionSegment(let interpolation):
                        interpolations.append(interpolation)
                        texts.append([])
                }
            }
            return interpolations.isEmpty ? nil : (texts, interpolations)
        }
    }
}

/*
 The TypeScript rule's SQL reading, ported line for line: it reads a literal's texts with each interpolation an
 opaque value between two of them, and answers which interpolations sit inside a single-quoted string in a value
 position. Bytes, as the Go reads them.
 */
enum SqlReading {
    static let leadingKeywords: Set<String> = [
        "SELECT", "INSERT", "UPDATE", "DELETE", "REPLACE", "WITH", "SHOW", "EXPLAIN", "PRAGMA", "WHERE", "AND", "OR",
        "SET", "VALUES", "HAVING",
    ]

    /* An uppercase `WHERE` or `HAVING` opening a literal is a clause, where `SELECT` or `AND` could open a shouted sentence. */
    static let selfConfirmingKeywords: Set<String> = ["WHERE", "HAVING"]

    static let keywords: Set<String> = [
        "SELECT", "INSERT", "UPDATE", "DELETE", "REPLACE", "WITH", "SHOW", "EXPLAIN", "PRAGMA", "WHERE", "AND", "OR",
        "SET", "VALUES", "HAVING",
        "FROM", "INTO", "LIKE", "GLOB", "BETWEEN", "JOIN", "ON", "IN", "IS", "NOT", "NULL", "ORDER", "GROUP", "BY",
        "LIMIT", "OFFSET", "AS",
        "DISTINCT", "COUNT", "TABLE", "STATUS", "CASE", "WHEN", "THEN", "ELSE", "END", "UNION", "ESCAPE", "ASC", "DESC",
    ]

    /* The tokens after which an opening quote begins a value. */
    static let valueTokens: Set<String> = [
        "=", "==", "<", ">", "<=", ">=", "<>", "!=", "||", "+", "(", ",",
        "LIKE", "GLOB", "REGEXP", "RLIKE", "BETWEEN", "AND", "SELECT", "WHEN", "THEN", "ELSE", "ESCAPE", "DEFAULT",
        "INTERVAL",
    ]

    static let fragmentKeywords: Set<String> = ["AND", "OR", "NOT", "WHERE", "SELECT", "FROM", "ESCAPE"]

    static let quote = UInt8(ascii: "'")
    static let backslash = UInt8(ascii: "\\")

    static func quotedSpans(_ texts: [[UInt8]]) -> [Int] {
        leadingKeywords.contains(leadingWord(texts[0])) ? readStatement(texts) : readLikePatterns(texts)
    }

    /* The first word after whitespace and `(`. */
    static func leadingWord(_ head: [UInt8]) -> String {
        var start = 0
        while start < head.count, isWhitespace(head[start]) || head[start] == UInt8(ascii: "(") {
            start += 1
        }
        return String(decoding: head[start..<start + wordLength(head, start)], as: UTF8.self)
    }

    static func isWordCharacter(_ character: UInt8) -> Bool {
        character == UInt8(ascii: "_") || (character >= UInt8(ascii: "a") && character <= UInt8(ascii: "z"))
            || (character >= UInt8(ascii: "A") && character <= UInt8(ascii: "Z"))
            || (character >= UInt8(ascii: "0") && character <= UInt8(ascii: "9"))
    }

    static func wordLength(_ text: [UInt8], _ start: Int) -> Int {
        var end = start
        while end < text.count, isWordCharacter(text[end]) {
            end += 1
        }
        return end - start
    }

    static func isWhitespace(_ character: UInt8) -> Bool {
        character == UInt8(ascii: " ") || character == UInt8(ascii: "\t") || character == UInt8(ascii: "\r")
            || character == UInt8(ascii: "\n")
    }

    static func isOperatorCharacter(_ character: UInt8) -> Bool {
        Array("=<>!|".utf8).contains(character)
    }

    enum State {
        case code
        case string
        case doubleQuotes
        case backquotes
        case lineComment
        case blockComment
    }

    /* The literal read as one statement. Empty when it is not confirmed as SQL or its reading is uncertain anywhere. */
    static func readStatement(_ texts: [[UInt8]]) -> [Int] {
        var state = State.code
        var previousToken = ""
        var keywordCount = 0
        var openedAtValue = false
        var quotedSpans: [Int] = []
        for (textIndex, text) in texts.enumerated() {
            var index = 0
            while index < text.count {
                let character = text[index]
                switch state {
                    case .code:
                        if isWhitespace(character) {
                            index += 1
                        }
                        else if character == quote {
                            state = .string
                            openedAtValue = valueTokens.contains(previousToken)
                            index += 1
                        }
                        else if character == UInt8(ascii: "\"") {
                            state = .doubleQuotes
                            index += 1
                        }
                        else if character == UInt8(ascii: "`") {
                            state = .backquotes
                            index += 1
                        }
                        else if character == UInt8(ascii: "-"), index + 1 < text.count,
                            text[index + 1] == UInt8(ascii: "-")
                        {
                            /* MySQL reads `a--1` as arithmetic, so only `-- ` is a comment everywhere. */
                            guard index + 2 < text.count, isWhitespace(text[index + 2]) else { return [] }
                            state = .lineComment
                            index += 3
                        }
                        else if character == UInt8(ascii: "/"), index + 1 < text.count,
                            text[index + 1] == UInt8(ascii: "*")
                        {
                            state = .blockComment
                            index += 2
                        }
                        else if character == UInt8(ascii: "#") {
                            /* A comment in MySQL only. */
                            return []
                        }
                        else if character == UInt8(ascii: "$") {
                            /* `$1` and `$name` are placeholders; `$$` and `$tag$` open a dollar-quoted string. */
                            let length = wordLength(text, index + 1)
                            if index + 1 + length < text.count, text[index + 1 + length] == UInt8(ascii: "$") {
                                return []
                            }
                            previousToken = "$"
                            index += 1 + length
                        }
                        else if isWordCharacter(character) {
                            let length = wordLength(text, index)
                            let word = String(decoding: text[index..<index + length], as: UTF8.self)
                            if keywords.contains(word) {
                                keywordCount += 1
                            }
                            previousToken = word.uppercased()
                            index += length
                        }
                        else if isOperatorCharacter(character) {
                            var end = index
                            while end < text.count, isOperatorCharacter(text[end]) {
                                end += 1
                            }
                            previousToken = String(decoding: text[index..<end], as: UTF8.self)
                            index = end
                        }
                        else {
                            previousToken = String(decoding: [character], as: UTF8.self)
                            index += 1
                        }
                    case .string:
                        if character == backslash {
                            /* An escape in MySQL, an ordinary character elsewhere: where the string ends depends on the engine. */
                            return []
                        }
                        if character == quote {
                            if index + 1 < text.count, text[index + 1] == quote {
                                index += 2
                                continue
                            }
                            state = .code
                            previousToken = "'"
                        }
                        index += 1
                    case .doubleQuotes, .backquotes:
                        let closing = state == .backquotes ? UInt8(ascii: "`") : UInt8(ascii: "\"")
                        if character == backslash {
                            return []
                        }
                        if character == closing {
                            if index + 1 < text.count, text[index + 1] == closing {
                                index += 2
                                continue
                            }
                            state = .code
                            previousToken = String(decoding: [closing], as: UTF8.self)
                        }
                        index += 1
                    case .lineComment:
                        if character == UInt8(ascii: "\n") {
                            state = .code
                        }
                        index += 1
                    case .blockComment:
                        if character == UInt8(ascii: "*"), index + 1 < text.count, text[index + 1] == UInt8(ascii: "/")
                        {
                            state = .code
                            index += 2
                            continue
                        }
                        index += 1
                }
            }
            if textIndex == texts.count - 1 {
                break
            }
            /* The interpolation after this text. */
            switch state {
                case .string:
                    if openedAtValue {
                        quotedSpans.append(textIndex)
                    }
                case .code:
                    previousToken = "\\()"
                case .doubleQuotes, .backquotes, .lineComment, .blockComment:
                    break
            }
        }
        guard state == .code || state == .lineComment else { return [] }
        guard keywordCount >= 2 || selfConfirmingKeywords.contains(leadingWord(texts[0])) else { return [] }
        return quotedSpans
    }

    /* In a literal that is not a statement, the interpolations inside the string after each uppercase `LIKE`. */
    static func readLikePatterns(_ texts: [[UInt8]]) -> [Int] {
        guard hasFragmentKeyword(texts) else { return [] }
        let like = Array("LIKE".utf8)
        var quotedSpans: [Int] = []
        var quotesBefore = 0
        var textIndex = 0
        while textIndex < texts.count {
            var text = texts[textIndex]
            var index = 0
            while index < text.count {
                defer { index += 1 }
                if text[index] == quote {
                    quotesBefore += 1
                    continue
                }
                guard text[index...].starts(with: like), index == 0 || !isWordCharacter(text[index - 1]) else {
                    continue
                }
                var opening = index + like.count
                guard opening < text.count, isWhitespace(text[opening]) else { continue }
                while opening < text.count, isWhitespace(text[opening]) {
                    opening += 1
                }
                guard opening < text.count, text[opening] == quote, quotesBefore % 2 == 0 else { continue }
                guard let string = readString(texts, textIndex: textIndex, start: opening + 1) else {
                    return quotedSpans
                }
                quotedSpans.append(contentsOf: string.spans)
                /* Resume after the closing quote; the two quotes are a pair, so the count before what follows stays even. */
                textIndex = string.closingText
                text = texts[textIndex]
                index = string.closingIndex
            }
            textIndex += 1
        }
        return quotedSpans
    }

    /* Whether a fragment carries an uppercase `AND`, `OR`, `NOT`, `WHERE`, `SELECT`, `FROM` or `ESCAPE` as a whole word. */
    static func hasFragmentKeyword(_ texts: [[UInt8]]) -> Bool {
        for text in texts {
            var index = 0
            while index < text.count {
                let length = wordLength(text, index)
                if length == 0 {
                    index += 1
                    continue
                }
                if fragmentKeywords.contains(String(decoding: text[index..<index + length], as: UTF8.self)) {
                    return true
                }
                index += length
            }
        }
        return false
    }

    /*
     A single-quoted string read from just after its opening quote, across interpolations: the interpolations
     inside it, and where its closing quote is. Nil when it never closes or holds a backslash.
     */
    static func readString(
        _ texts: [[UInt8]],
        textIndex start: Int,
        start position: Int,
    ) -> (spans: [Int], closingText: Int, closingIndex: Int)? {
        var spans: [Int] = []
        var index = position
        for textIndex in start..<texts.count {
            let text = texts[textIndex]
            while index < text.count {
                if text[index] == backslash {
                    return nil
                }
                if text[index] == quote {
                    if index + 1 < text.count, text[index + 1] == quote {
                        index += 2
                        continue
                    }
                    return (spans, textIndex, index)
                }
                index += 1
            }
            if textIndex < texts.count - 1 {
                spans.append(textIndex)
            }
            index = 0
        }
        return nil
    }
}

/*
 What a value's type is, as far as this file and the index can show it. The shapes are the few the type gate and
 the `for` loops need; anything else is nil, which reads as "not shown" and is never a finding.
 */
indirect enum ValueShape: Equatable {
    /* A type as the demangler prints it, `Swift.String`. */
    case named(String)
    /* A string literal's own value, nil when it interpolates. A `let` keeps it, as TypeScript's `const` keeps a literal type. */
    case literal(String?)
    case optional(ValueShape)
    /* Something a `for` loop walks, by its element. */
    case sequence(ValueShape)
    case dictionary(key: ValueShape, value: ValueShape)
    case tuple([ValueShape])

    static let textTypes: Set<String> = ["Swift.String", "Swift.Substring", "Swift.Character", "Any"]

    /* TypeScript's `string`, `any` and `unknown`, and a literal holding a quote or a backslash. */
    var canHoldQuote: Bool {
        switch self {
            case .named(let name): Self.textTypes.contains(name)
            case .literal(let text): text.map { $0.contains("'") || $0.contains("\\") } ?? true
            case .optional(let wrapped): wrapped.canHoldQuote
            case .sequence, .dictionary, .tuple: false
        }
    }

    /* A `var`, a function's result, or a collection's member: a literal's type widens to `String`. */
    var widened: ValueShape {
        if case .literal = self {
            return .named("Swift.String")
        }
        return self
    }

    var unwrapped: ValueShape {
        if case .optional(let wrapped) = self {
            return wrapped
        }
        return self
    }

    var element: ValueShape? {
        switch self {
            case .sequence(let element): element
            case .dictionary(let key, let value): .tuple([key, value])
            case .named, .literal, .optional, .tuple: nil
        }
    }

    /* A type as the demangler prints it: `[Swift.String : Swift.String]`, `Swift.String?`, `(key: Swift.String, value: Swift.Int)`. */
    static func parse(_ spelled: Substring) -> ValueShape? {
        var text = spelled.trimmingCharacters(in: .whitespaces)[...]
        for prefix in ["inout ", "__owned ", "__shared ", "consuming ", "borrowing "] where text.hasPrefix(prefix) {
            text = text.dropFirst(prefix.count)
        }
        guard !text.isEmpty, topLevel(" -> ", in: text) == nil else { return nil }
        if text.hasSuffix("?") || text.hasSuffix("!") {
            return parse(text.dropLast()).map(ValueShape.optional)
        }
        if text.first == "[", text.last == "]", closing(of: text) == text.index(before: text.endIndex) {
            let inner = text.dropFirst().dropLast()
            if let colon = topLevel(" : ", in: inner) {
                guard let key = parse(inner[..<colon.lowerBound]), let value = parse(inner[colon.upperBound...]) else {
                    return nil
                }
                return .dictionary(key: key, value: value)
            }
            return parse(inner).map(ValueShape.sequence)
        }
        if text.first == "(", text.last == ")", closing(of: text) == text.index(before: text.endIndex) {
            let members = split(text.dropFirst().dropLast()).map(dropLabel)
            guard !members.isEmpty else { return nil }
            let shapes = members.compactMap(parse)
            guard shapes.count == members.count else { return nil }
            return shapes.count == 1 ? shapes[0] : .tuple(shapes)
        }
        if let open = text.firstIndex(of: "<"), text.last == ">" {
            let arguments = split(text[text.index(after: open)..<text.index(before: text.endIndex)]).map(parse)
            switch (text[..<open], arguments.count) {
                case ("Swift.Optional", 1): return arguments[0].map(ValueShape.optional)
                case ("Swift.Array", 1), ("Swift.Set", 1), ("Swift.ArraySlice", 1), ("Swift.ContiguousArray", 1):
                    return arguments[0].map(ValueShape.sequence)
                case ("Swift.Dictionary", 2):
                    guard let key = arguments[0], let value = arguments[1] else { return nil }
                    return .dictionary(key: key, value: value)
                default: return .named(String(text))
            }
        }
        return .named(String(text))
    }

    /* The range of the first `separator` outside brackets, parentheses and angle brackets (an arrow's `>` is not one). */
    static func topLevel(_ separator: String, in text: Substring) -> Range<Substring.Index>? {
        var depth = 0
        var index = text.startIndex
        while index < text.endIndex {
            if depth == 0, text[index...].hasPrefix(separator) {
                return index..<text.index(index, offsetBy: separator.count)
            }
            if text[index...].hasPrefix("->") {
                index = text.index(index, offsetBy: 2)
                continue
            }
            switch text[index] {
                case "(", "[", "<": depth += 1
                case ")", "]", ">": depth -= 1
                default: break
            }
            index = text.index(after: index)
        }
        return nil
    }

    /* Where the bracket opening `text` closes. */
    static func closing(of text: Substring) -> Substring.Index? {
        var depth = 0
        var index = text.startIndex
        while index < text.endIndex {
            if text[index...].hasPrefix("->") {
                index = text.index(index, offsetBy: 2)
                continue
            }
            switch text[index] {
                case "(", "[", "<": depth += 1
                case ")", "]", ">":
                    depth -= 1
                    if depth == 0 {
                        return index
                    }
                default: break
            }
            index = text.index(after: index)
        }
        return nil
    }

    static func split(_ text: Substring) -> [Substring] {
        var parts: [Substring] = []
        var rest = text
        while let comma = topLevel(", ", in: rest) {
            parts.append(rest[..<comma.lowerBound])
            rest = rest[comma.upperBound...]
        }
        if !rest.trimmingCharacters(in: .whitespaces).isEmpty {
            parts.append(rest)
        }
        return parts
    }

    /* `key: Swift.String` is `Swift.String`. */
    static func dropLabel(_ member: Substring) -> Substring {
        guard let colon = member.range(of: ": "),
            member[..<colon.lowerBound].allSatisfy({ $0.isLetter || $0.isNumber || $0 == "_" })
        else {
            return member
        }
        return member[colon.upperBound...]
    }
}

/* The type gate and the escape, read from this file's bindings and the declarations the index names. */
final class ValueTypes {
    let file: ParsedFile
    let symbols: FileSymbols
    let demangler: SwiftDemangler?
    /* The bindings this file declares, by the symbol the index gives their name, built on first use. */
    private var declaredHere: [String: PatternBindingSyntax]?

    init(file: ParsedFile, symbols: FileSymbols, demangler: SwiftDemangler?) {
        self.file = file
        self.symbols = symbols
        self.demangler = demangler
    }

    /* The stdlib members that keep a collection's element. */
    static let elementPreserving: Set<String> = [
        "sorted", "reversed", "filter", "shuffled", "prefix", "suffix", "dropFirst", "dropLast",
    ]

    func canHoldQuote(_ value: ExprSyntax) -> Bool {
        shape(of: value)?.canHoldQuote ?? false
    }

    func shape(of expression: ExprSyntax) -> ValueShape? {
        if let literal = expression.as(StringLiteralExprSyntax.self) {
            return literal.segments.contains { $0.is(ExpressionSegmentSyntax.self) }
                ? .literal(nil) : literal.representedLiteralValue.map { .literal($0) }
        }
        if let tried = expression.as(TryExprSyntax.self) {
            return shape(of: tried.expression)
        }
        if let awaited = expression.as(AwaitExprSyntax.self) {
            return shape(of: awaited.expression)
        }
        if let tuple = expression.as(TupleExprSyntax.self) {
            guard tuple.elements.count == 1, let only = tuple.elements.first, only.label == nil else { return nil }
            return shape(of: only.expression)
        }
        if let unwrapped = expression.as(ForceUnwrapExprSyntax.self) {
            return shape(of: unwrapped.expression)?.unwrapped
        }
        if let chained = expression.as(OptionalChainingExprSyntax.self) {
            return shape(of: chained.expression)
        }
        if let cast = expression.as(AsExprSyntax.self) {
            let target = shape(of: cast.type)
            return cast.questionOrExclamationMark?.tokenKind == .postfixQuestionMark
                ? target.map(ValueShape.optional) : target
        }
        if let array = expression.as(ArrayExprSyntax.self) {
            /* An array literal of string literals is `[String]`, as TypeScript widens `['a', 'b']` to `string[]`. */
            guard !array.elements.isEmpty, array.elements.allSatisfy({ $0.expression.is(StringLiteralExprSyntax.self) })
            else { return nil }
            return .sequence(.named("Swift.String"))
        }
        if let infix = expression.as(InfixOperatorExprSyntax.self) {
            return shape(ofInfix: infix)
        }
        if let reference = expression.as(DeclReferenceExprSyntax.self) {
            guard reference.argumentNames == nil else { return nil }
            if let binding = localBinding(named: reference.baseName.text, from: Syntax(reference)) {
                return binding.shape
            }
            return declared(at: reference.baseName)?.shape
        }
        if let member = expression.as(MemberAccessExprSyntax.self) {
            return shape(ofMember: member)
        }
        if let call = expression.as(FunctionCallExprSyntax.self) {
            return shape(ofCall: call)
        }
        if let subscripted = expression.as(SubscriptCallExprSyntax.self) {
            /* A dictionary's subscript by key answers an optional value; anything else is not read. */
            guard case .dictionary(_, let value)? = shape(of: subscripted.calledExpression) else { return nil }
            return .optional(value)
        }
        return nil
    }

    func shape(ofInfix infix: InfixOperatorExprSyntax) -> ValueShape? {
        guard let operation = infix.operator.as(BinaryOperatorExprSyntax.self) else { return nil }
        /* An operator of ours could return anything. */
        if let declaration = reference(at: operation.operator), symbols.isOwned(declaration) {
            return nil
        }
        switch operation.operator.text {
            case "??":
                return shape(of: infix.leftOperand)?.unwrapped.widened
            case "+":
                let left = shape(of: infix.leftOperand)
                let right = shape(of: infix.rightOperand)
                /* Strings concatenate to a string; collections keep their element, which either side may show. */
                if left?.widened == .named("Swift.String") || right?.widened == .named("Swift.String") {
                    return .named("Swift.String")
                }
                if case .sequence(let element)? = left {
                    return .sequence(element)
                }
                if case .sequence(let element)? = right {
                    return .sequence(element)
                }
                return nil
            default:
                return nil
        }
    }

    func shape(ofMember member: MemberAccessExprSyntax) -> ValueShape? {
        guard member.declName.argumentNames == nil, let declaration = reference(at: member.declName.baseName) else {
            return nil
        }
        if declaration.isStandardLibrary, let base = member.base {
            switch (member.declName.baseName.text, shape(of: base)) {
                case ("keys", .dictionary(let key, _)?): return .sequence(key)
                case ("values", .dictionary(_, let value)?): return .sequence(value)
                case ("lazy", let collection?): return collection.element.map(ValueShape.sequence)
                default: break
            }
        }
        return declared(declaration)?.shape
    }

    func shape(ofCall call: FunctionCallExprSyntax) -> ValueShape? {
        let callee = call.calledExpression
        if let member = callee.as(MemberAccessExprSyntax.self), let base = member.base,
            let declaration = reference(at: member.declName.baseName), declaration.isStandardLibrary
        {
            let name = member.declName.baseName.text
            if Self.elementPreserving.contains(name), let element = shape(of: base)?.element {
                return .sequence(element)
            }
            if name == "enumerated", call.arguments.isEmpty, let element = shape(of: base)?.element {
                return .sequence(.tuple([.named("Swift.Int"), element]))
            }
        }
        if let type = callee.as(DeclReferenceExprSyntax.self),
            ["Array", "Set", "ContiguousArray"].contains(type.baseName.text), call.trailingClosure == nil,
            let argument = call.arguments.first, call.arguments.count == 1, argument.label == nil,
            isStandardLibrary(at: type.baseName)
        {
            return shape(of: argument.expression)?.element.map(ValueShape.sequence)
        }
        let name: TokenSyntax
        if let member = callee.as(MemberAccessExprSyntax.self) {
            name = member.declName.baseName
        }
        else if let reference = callee.as(DeclReferenceExprSyntax.self) {
            /* A local closure's result is not read. */
            guard localBinding(named: reference.baseName.text, from: Syntax(reference)) == nil else { return nil }
            name = reference.baseName
        }
        else {
            return nil
        }
        /* A type's name called as a function records the type and its initializer at one place; the one function among them answers. */
        let location = file.locations.location(for: name.positionAfterSkippingLeadingTrivia)
        let functions = symbols.occurrences(line: location.line, column: location.column).filter {
            $0.isReference && !$0.isImplicit
        }.compactMap(declared).filter(\.isFunction)
        guard functions.count == 1, let function = functions.first else { return nil }
        return function.shape
    }

    /* The one written reference the index recorded at this name. */
    func reference(at token: TokenSyntax) -> FileSymbols.Occurrence? {
        let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
        return symbols.reference(line: location.line, column: location.column)
    }

    /* The index records both the type and its initializer at `Array(...)`; every written reference there must be the standard library's. */
    func isStandardLibrary(at token: TokenSyntax) -> Bool {
        let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
        let references = symbols.occurrences(line: location.line, column: location.column).filter {
            $0.isReference && !$0.isImplicit
        }
        return !references.isEmpty && references.allSatisfy(\.isStandardLibrary)
    }

    func declared(at token: TokenSyntax) -> (shape: ValueShape?, isFunction: Bool)? {
        reference(at: token).flatMap(declared)
    }

    /*
     A declaration's type, or its result for a function or initializer, from its demangled form:
     `AhraOSPresence.SQLValue.text : Swift.String?`, `AhraOSPresence.PresenceStore.create(...) throws -> Swift.String`.
     */
    func declared(_ occurrence: FileSymbols.Occurrence) -> (shape: ValueShape?, isFunction: Bool)? {
        /* An enum's own `rawValue` is one of its raw values, a closed set. */
        if occurrence.symbol.contains("O8rawValue") {
            return nil
        }
        guard let demangler, var demangled = demangler.declaration(ofSymbol: occurrence.symbol)?[...] else {
            return nil
        }
        if demangled.hasPrefix("static ") {
            demangled = demangled.dropFirst("static ".count)
        }
        if let colon = ValueShape.topLevel(" : ", in: demangled) {
            /* A parameter the index recorded prints as `id #1 : Swift.String in <its function>`, and is never a constant. */
            let isLocal = demangled[..<colon.lowerBound].contains(" #")
            var spelled = demangled[colon.upperBound...]
            if isLocal, let context = ValueShape.topLevel(" in ", in: spelled) {
                spelled = spelled[..<context.lowerBound]
            }
            guard let type = ValueShape.parse(spelled) else { return (nil, false) }
            /*
             A non-optional text property of ours could be a compile-time constant (`static let table = "items"`),
             which TypeScript types as a literal. Its initializer is readable only when it is declared here.
             */
            if !isLocal, symbols.isOwned(occurrence), case .named = type, type.canHoldQuote {
                return (declaredHere(occurrence.symbol).flatMap { self.shape(of: $0, isLet: Self.isLet($0)) }, false)
            }
            return (type, false)
        }
        if let arrow = ValueShape.topLevel(" -> ", in: demangled),
            [")", " throws", " rethrows", " async"].contains(where: { demangled[..<arrow.lowerBound].hasSuffix($0) })
        {
            return (ValueShape.parse(demangled[arrow.upperBound...])?.widened, true)
        }
        return (nil, false)
    }

    /* The binding in this file whose name the index records as declaring `symbol`. */
    func declaredHere(_ symbol: String) -> PatternBindingSyntax? {
        if declaredHere == nil {
            var found: [String: PatternBindingSyntax] = [:]
            for binding in file.tree.tokens(viewMode: .sourceAccurate).compactMap({
                $0.parent?.as(IdentifierPatternSyntax.self)?.parent?.as(PatternBindingSyntax.self)
            }) {
                guard let name = binding.pattern.as(IdentifierPatternSyntax.self)?.identifier else { continue }
                let location = file.locations.location(for: name.positionAfterSkippingLeadingTrivia)
                for occurrence in symbols.occurrences(line: location.line, column: location.column)
                where !occurrence.isReference {
                    found[occurrence.symbol] = binding
                }
            }
            declaredHere = found
        }
        return declaredHere?[symbol]
    }

    static func isLet(_ binding: PatternBindingSyntax) -> Bool {
        binding.parent?.parent?.as(VariableDeclSyntax.self)?.bindingSpecifier.tokenKind == .keyword(.let)
    }

    /* A binding of a name: its shape when shown, and a `let`'s initializer, which the escape follows. */
    struct Binding {
        var shape: ValueShape?
        var letInitializer: ExprSyntax?
    }

    /* A variable declaration's binding: its annotation, or its initializer, a literal kept only by a `let`. */
    func shape(of binding: PatternBindingSyntax, isLet: Bool) -> ValueShape? {
        if let annotation = binding.typeAnnotation {
            return shape(of: annotation.type)
        }
        /* A computed property with no annotation does not compile, and an accessor block is not read. */
        guard binding.accessorBlock == nil, let initializer = binding.initializer?.value else { return nil }
        let value = shape(of: initializer)
        return isLet ? value : value?.widened
    }

    /*
     The binding a name refers to at this place, found by walking out through the scopes that enclose it, or nil
     when none in this file binds it before this place (a member or a global, which the index names). A binding
     whose type this does not read (a `switch` or `catch` pattern, a closure parameter with no annotation) is
     found with no shape, so the name is never read past it.
     */
    func localBinding(named name: String, from start: Syntax) -> Binding? {
        var child = start
        while let scope = child.parent {
            defer { child = scope }
            if scope.is(StructDeclSyntax.self) || scope.is(ClassDeclSyntax.self) || scope.is(EnumDeclSyntax.self)
                || scope.is(ActorDeclSyntax.self)
                || scope.is(ExtensionDeclSyntax.self) || scope.is(ProtocolDeclSyntax.self)
            {
                return nil
            }
            if let list = scope.as(CodeBlockItemListSyntax.self) {
                for statement in list.reversed() where statement.endPosition <= child.position {
                    if let binding = binding(named: name, inStatement: statement.item) {
                        return binding
                    }
                }
            }
            else if let conditions = scope.as(ConditionElementListSyntax.self) {
                for condition in conditions.reversed() where condition.endPosition <= child.position {
                    if let binding = binding(named: name, in: condition, at: scope) {
                        return binding
                    }
                }
            }
            else if let ifExpression = scope.as(IfExprSyntax.self), child.id == ifExpression.body.id {
                if let binding = binding(named: name, in: ifExpression.conditions, at: scope) {
                    return binding
                }
            }
            else if let whileStatement = scope.as(WhileStmtSyntax.self), child.id == whileStatement.body.id {
                if let binding = binding(named: name, in: whileStatement.conditions, at: scope) {
                    return binding
                }
            }
            else if let forStatement = scope.as(ForStmtSyntax.self),
                child.id == forStatement.body.id || child.id == forStatement.whereClause?.id
            {
                if Self.binds(name, forStatement.pattern) {
                    guard forStatement.caseKeyword == nil, forStatement.typeAnnotation == nil,
                        let element = shape(of: forStatement.sequence)?.element
                    else {
                        return Binding()
                    }
                    return Binding(shape: Self.destructure(forStatement.pattern, element, name: name))
                }
            }
            else if let closure = scope.as(ClosureExprSyntax.self), child.id == closure.statements.id {
                if let binding = binding(named: name, in: closure) {
                    return binding
                }
            }
            else if let parameters = Self.parameters(of: scope) {
                for parameter in parameters where (parameter.secondName ?? parameter.firstName).text == name {
                    let type = shape(of: parameter.type)
                    return Binding(shape: parameter.ellipsis == nil ? type : type.map(ValueShape.sequence))
                }
                if let accessor = scope.as(AccessorDeclSyntax.self),
                    accessor.parameters?.name.text == name
                        || (accessor.parameters == nil && ["newValue", "oldValue"].contains(name))
                {
                    return Binding()
                }
            }
            else if let switchCase = scope.as(SwitchCaseSyntax.self), child.id == switchCase.statements.id {
                /* A case pattern's bindings are not read. */
                if Self.binds(name, switchCase.label) {
                    return Binding()
                }
            }
            else if let catchClause = scope.as(CatchClauseSyntax.self), child.id == catchClause.body.id {
                /* Nor a catch pattern's, nor its implicit `error`. */
                if Self.binds(name, catchClause.catchItems) || (catchClause.catchItems.isEmpty && name == "error") {
                    return Binding()
                }
            }
        }
        return nil
    }

    /* A function's, initializer's or subscript's parameters, or an accessor's named one. */
    static func parameters(of scope: Syntax) -> [FunctionParameterSyntax]? {
        if let function = scope.as(FunctionDeclSyntax.self) {
            return Array(function.signature.parameterClause.parameters)
        }
        if let initializer = scope.as(InitializerDeclSyntax.self) {
            return Array(initializer.signature.parameterClause.parameters)
        }
        if let subscriptDeclaration = scope.as(SubscriptDeclSyntax.self) {
            return Array(subscriptDeclaration.parameterClause.parameters)
        }
        if scope.is(AccessorDeclSyntax.self) {
            return []
        }
        return nil
    }

    /* A statement before the name's own that binds it for what follows: a declaration, or a `guard`'s conditions. */
    func binding(named name: String, inStatement statement: CodeBlockItemSyntax.Item) -> Binding? {
        if let declaration = statement.as(VariableDeclSyntax.self) {
            let isLet = declaration.bindingSpecifier.tokenKind == .keyword(.let)
            for binding in declaration.bindings where Self.binds(name, binding.pattern) {
                guard binding.pattern.is(IdentifierPatternSyntax.self) else {
                    return Binding(
                        shape: shape(of: binding, isLet: isLet).flatMap {
                            Self.destructure(binding.pattern, $0, name: name)
                        }
                    )
                }
                return Binding(
                    shape: shape(of: binding, isLet: isLet),
                    letInitializer: isLet ? binding.initializer?.value : nil,
                )
            }
            return nil
        }
        if let guardStatement = statement.as(GuardStmtSyntax.self) {
            return binding(named: name, in: guardStatement.conditions, at: Syntax(guardStatement))
        }
        if let function = statement.as(FunctionDeclSyntax.self), function.name.text == name {
            return Binding()
        }
        return nil
    }

    /* The last condition in a list that binds the name. */
    func binding(named name: String, in conditions: ConditionElementListSyntax, at statement: Syntax) -> Binding? {
        for condition in conditions.reversed() {
            if let binding = binding(named: name, in: condition, at: statement) {
                return binding
            }
        }
        return nil
    }

    func binding(named name: String, in condition: ConditionElementSyntax, at statement: Syntax) -> Binding? {
        guard Self.binds(name, condition) else { return nil }
        guard let optional = condition.condition.as(OptionalBindingConditionSyntax.self),
            let pattern = optional.pattern.as(IdentifierPatternSyntax.self)
        else {
            /* `if case let`, and anything else that binds, is not read. */
            return Binding()
        }
        if let annotation = optional.typeAnnotation {
            return Binding(shape: shape(of: annotation.type))
        }
        if let initializer = optional.initializer?.value {
            return Binding(shape: shape(of: initializer)?.unwrapped.widened)
        }
        /* `if let name`, the outer name unwrapped, read from before the statement. */
        if let outer = localBinding(named: name, from: statement) {
            return Binding(shape: outer.shape?.unwrapped.widened)
        }
        return Binding(shape: declared(at: pattern.identifier)?.shape?.unwrapped.widened)
    }

    /* A closure's parameters, annotated or not, and its capture list. */
    func binding(named name: String, in closure: ClosureExprSyntax) -> Binding? {
        guard let signature = closure.signature else {
            return nil
        }
        if let captures = signature.capture?.items {
            for capture in captures where capture.name.text == name {
                return Binding(shape: capture.initializer.flatMap { shape(of: $0.value) }?.widened)
            }
        }
        switch signature.parameterClause {
            case .simpleInput(let names):
                if names.contains(where: { $0.name.text == name }) {
                    return Binding()
                }
            case .parameterClause(let clause):
                for parameter in clause.parameters where (parameter.secondName ?? parameter.firstName).text == name {
                    return Binding(shape: parameter.type.flatMap { shape(of: $0) })
                }
            case nil:
                break
        }
        return nil
    }

    /* Whether any identifier pattern under this node binds the name. */
    static func binds(_ name: String, _ node: some SyntaxProtocol) -> Bool {
        node.tokens(viewMode: .sourceAccurate).contains {
            $0.text == name && $0.parent?.is(IdentifierPatternSyntax.self) == true
        }
    }

    /* The part of a shown type a pattern gives a name: the whole for a name, a member by position for a tuple. */
    static func destructure(_ pattern: PatternSyntax, _ shape: ValueShape, name: String) -> ValueShape? {
        if let identifier = pattern.as(IdentifierPatternSyntax.self) {
            return identifier.identifier.text == name ? shape.widened : nil
        }
        guard let tuple = pattern.as(TuplePatternSyntax.self), case .tuple(let members) = shape,
            members.count == tuple.elements.count
        else { return nil }
        for (element, member) in zip(tuple.elements, members) where binds(name, element.pattern) {
            return destructure(element.pattern, member, name: name)
        }
        return nil
    }

    /* An annotation's type, its names resolved through the index to the standard library's. */
    func shape(of type: TypeSyntax) -> ValueShape? {
        if let optional = type.as(OptionalTypeSyntax.self) {
            return shape(of: optional.wrappedType).map(ValueShape.optional)
        }
        if let unwrapped = type.as(ImplicitlyUnwrappedOptionalTypeSyntax.self) {
            return shape(of: unwrapped.wrappedType).map(ValueShape.optional)
        }
        if let array = type.as(ArrayTypeSyntax.self) {
            return shape(of: array.element).map(ValueShape.sequence)
        }
        if let dictionary = type.as(DictionaryTypeSyntax.self) {
            guard let key = shape(of: dictionary.key), let value = shape(of: dictionary.value) else { return nil }
            return .dictionary(key: key, value: value)
        }
        if let tuple = type.as(TupleTypeSyntax.self) {
            let members = tuple.elements.compactMap { shape(of: $0.type) }
            guard members.count == tuple.elements.count, !members.isEmpty else { return nil }
            return members.count == 1 ? members[0] : .tuple(members)
        }
        if let attributed = type.as(AttributedTypeSyntax.self) {
            return shape(of: attributed.baseType)
        }
        let name: TokenSyntax
        let arguments: GenericArgumentListSyntax?
        if let identifier = type.as(IdentifierTypeSyntax.self) {
            if identifier.name.text == "Any", identifier.genericArgumentClause == nil {
                return .named("Any")
            }
            name = identifier.name
            arguments = identifier.genericArgumentClause?.arguments
        }
        else if let member = type.as(MemberTypeSyntax.self) {
            name = member.name
            arguments = member.genericArgumentClause?.arguments
        }
        else {
            return nil
        }
        guard let declaration = reference(at: name), declaration.isStandardLibrary,
            let demangled = demangler?.declaration(ofSymbol: declaration.symbol)
        else { return nil }
        guard let arguments else { return .named(demangled) }
        let shapes = arguments.compactMap { $0.argument.as(TypeSyntax.self).flatMap { shape(of: $0) } }
        guard shapes.count == arguments.count else { return nil }
        switch (demangled, shapes.count) {
            case ("Swift.Optional", 1): return .optional(shapes[0])
            case ("Swift.Array", 1), ("Swift.Set", 1), ("Swift.ArraySlice", 1), ("Swift.ContiguousArray", 1):
                return .sequence(shapes[0])
            case ("Swift.Dictionary", 2): return .dictionary(key: shapes[0], value: shapes[1])
            default: return nil
        }
    }

    /* The value, or the initializer of the local `let` it names, ends in the escape the TypeScript rule accepts. */
    func isEscaped(_ value: ExprSyntax) -> Bool {
        var expression = value
        while let inner = expression.as(TryExprSyntax.self)?.expression
            ?? expression.as(AwaitExprSyntax.self)?.expression
        {
            expression = inner
        }
        if isEscapeChain(expression) {
            return true
        }
        guard let reference = expression.as(DeclReferenceExprSyntax.self),
            let initializer = localBinding(named: reference.baseName.text, from: Syntax(reference))?.letInitializer
        else {
            return false
        }
        return isEscapeChain(initializer)
    }

    /* The last two calls double every backslash and every single quote, in either order. */
    func isEscapeChain(_ expression: ExprSyntax) -> Bool {
        guard let outer = replaceStep(expression), let inner = replaceStep(outer.receiver) else { return false }
        return Set([outer.doubled, inner.doubled]) == ["'", "\\"]
    }

    /*
     `replacingOccurrences(of: "'", with: "''")`, or `replacing("'", with: "''")` or `replacing(/'/, with: "''")`, and
     the backslash twins, called on a value and resolved to Foundation's or the standard library's: which character
     it doubles, and what it was called on.
     */
    func replaceStep(_ expression: ExprSyntax) -> (doubled: String, receiver: ExprSyntax)? {
        guard let call = expression.as(FunctionCallExprSyntax.self), call.trailingClosure == nil,
            call.arguments.count == 2,
            let member = call.calledExpression.as(MemberAccessExprSyntax.self), var receiver = member.base,
            !receiver.is(OptionalChainingExprSyntax.self),
            let declaration = reference(at: member.declName.baseName), !symbols.isOwned(declaration)
        else { return nil }
        let arguments = Array(call.arguments)
        let labels = arguments.map { $0.label?.text }
        let doubled: String?
        switch (member.declName.baseName.text, labels) {
            case ("replacingOccurrences", ["of", "with"]):
                doubled = arguments[0].expression.as(StringLiteralExprSyntax.self)?.representedLiteralValue
            case ("replacing", [nil, "with"]):
                doubled =
                    arguments[0].expression.as(StringLiteralExprSyntax.self)?.representedLiteralValue
                    ?? arguments[0].expression.as(RegexLiteralExprSyntax.self).flatMap {
                        ["'": "'", "\\\\": "\\"][$0.regex.text]
                    }
            default:
                doubled = nil
        }
        guard let doubled, ["'", "\\"].contains(doubled),
            arguments[1].expression.as(StringLiteralExprSyntax.self)?.representedLiteralValue == doubled + doubled
        else {
            return nil
        }
        while let inner = receiver.as(TupleExprSyntax.self), inner.elements.count == 1, let only = inner.elements.first,
            only.label == nil
        {
            receiver = only.expression
        }
        return (doubled, receiver)
    }
}
