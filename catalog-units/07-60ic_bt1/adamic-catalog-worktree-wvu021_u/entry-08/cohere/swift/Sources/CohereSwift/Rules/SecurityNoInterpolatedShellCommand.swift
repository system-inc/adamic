import SwiftOperators
import SwiftParser
import SwiftSyntax

/*
 No value built into a shell's command string: `process.executableURL = URL(fileURLWithPath: "/bin/zsh")` with
 `process.arguments = ["-l", "-i", "-c", "cd \(projectRoot) && ahra os sleep"]`. The shell parses the whole string
 as code, the value included, so a space or a quote in `projectRoot` breaks the command and a `$(...)` or a
 backtick in it runs. Quoting the value by hand is one escape away from the same bug. The repair keeps the script
 literal and hands the value over as its own argument, after the script, where the shell reads it as `$1` and
 never parses it (`["-c", "cd \"$1\" && make", "zsh", path]`), or drops the shell and runs the program directly
 (`currentDirectoryURL` in place of the `cd`).

 It ports `nexus/security-no-interpolated-shell-command`, which judges a command string handed to node's
 `child_process` functions that run it through a shell. Swift has no function that adds a shell on its own, so
 the shell is always the program, chosen by name, and its `-c` operand is the command string. That is the shape
 the TypeScript rule leaves alone (`execFile('sh', ['-c', script])`, "a list rather than a type"); here it is the
 only shape there is, so the list is written down, and node's `shell: true` over an argument list has no Swift
 counterpart. The program is one of `sh`, `bash`, `zsh`, `dash`, `ksh` or `mksh` by the last component of a
 literal path, given as a string or `URL(fileURLWithPath:)` or `URL(filePath:)`, directly or through a `let`. It
 is given with its arguments in one of two ways:
 - `receiver.arguments = [...]`, where `receiver.executableURL` or `receiver.launchPath` is set to the shell in
   the same statement list, and every such assignment to that receiver in the enclosing body names a shell (a
   `Process`, or anything spelled like one);
 - a call with an argument labeled `arguments` and another argument that is the shell
   (`Process.run(URL(fileURLWithPath: "/bin/zsh"), arguments: [...])`,
   `PaneProcessSpec(executable: "/bin/zsh", arguments: [...])`, a spawn request carrying both).
 The arguments are an array literal, directly or through a `let`. They are read the way the shell reads its
 own: literal options (`-l`, `-i`, `-lc`, `+o`, `-o NAME`, `--login`), and a `c` in a short cluster makes the
 first operand after them the command string. A non-literal element before the `c`, a non-literal one after it
 whose text starts like an option, or a long option the rule does not know, stops the reading and nothing is
 judged. Operands after the command string are the shell's `$0`, `$1` and on, data the shell never parses, so
 they are never judged.

 From there it is the TypeScript rule's judgement, piece for piece:
 - The command is built here: a string literal with an interpolation, a `+` concatenation, or a ternary with
   such a branch, through parentheses and `let` bindings. An opaque command (a parameter, a `var`, a call's
   result) is not judged: it may be a whole script its caller wrote as a literal, and nothing at this call says
   a value was spliced into it. So `PaneProcessSpec.interactiveClaude(command:)`, which passes its `command`
   straight to `-c`, is not flagged, as `execSync(command)` is not.
 - The command is flattened into the text the author wrote and the values spliced into it, through
   interpolations, `+` operands, and `let` bindings whose initializer is a literal, an interpolation, a
   concatenation, a ternary or another name. A quoted heredoc's body is data (`cat <<'EOSQL'` through
   `EOSQL`), so a value there is skipped, with the TypeScript rule's grammar and its one-way errors: a `<<` it
   cannot read declines the command, as does a ternary branch that may open one.
 - A value is unsafe when it is text that is not a literal, which is what the TypeScript rule's `string` is.
   Swift has no literal types, so a value is text only when the rule can show it: a non-literal operand of a
   `+` with a string literal among its operands (Swift adds nothing but text to a `String`); a name declared
   `String`, `Substring`, `Character` or `NSString` (or an optional of one), a `var` initialized with text, or a
   `let`, `if let` or `guard let` bound to such a value; `String(...)`; and a call or property whose
   declaration, read from the index the build wrote and demangled by the toolchain, gives one of those types,
   or that is one of `NSString`'s `stringBy` methods and properties, every one of which the SDK declares as
   returning `NSString`. A `let` bound to a literal is the literal's text, as a `const` is. A number, a `Bool`,
   or a name declared with one, is safe, as the TypeScript rule's `number` and `boolean` are.

 What it accepts as safe, and why: literal text, numbers and booleans (no shell syntax the author did not
 write), a value in a quoted heredoc's body (the shell expands nothing there), an opaque command (see above),
 and the operands after the command string (positional parameters, never parsed). A quoting helper is not one
 of them. The TypeScript rule reports a hand-escaped value on purpose, because the escape is a convention every
 call must repeat exactly and the fix makes it unnecessary, and a helper adds a reason of its own: whether an
 escape is right depends on where the result lands, not only on the helper's body. Single-quote escaping
 (`'` + `'\''` + `'`) is right outside quotes and wrong inside double quotes, where `'` is an ordinary
 character and `$(...)` still runs, so no helper's name or body proves the splice safe.
 `Studio/Described.swift`'s `quoted(_:)` escapes correctly where it is used, and is reported anyway.

 Known misses, every one a finding not made and never one invented:
 - A value interpolated whose type the rule cannot read is not judged: a subscript
   (`ProcessInfo.processInfo.environment[...]`, so the env-gated bench hooks in `AhraOsWindowModel.swift` are
   missed), a generic result, an Objective-C declaration outside `NSString`'s `stringBy` family, a closure's or
   a loop's untyped parameter, a pattern binding, an enum or a `URL`, and a name declared in another file of our
   own, which may be a `let` bound to a literal there (that last one is not judged as a `+` operand either).
   Such a value could also open a quoted heredoc the rule cannot see, so a value after it in the command is not
   judged either.
 - Arguments not given as an array literal or a `let` bound to one (a `let arguments: [String]` assigned in
   branches, which is the second bench hook), a spread of arrays, an executable set in another statement list
   than its arguments or held in a `var`, a program run through `/usr/bin/env`, and an argument vector for
   `posix_spawn` with the shell as its first element.
 - Interpreters that take a script by flag but are not POSIX shells (`fish`, `csh`, `python3 -c`,
   `osascript -e`): each has its own grammar for its script.
 - A value that itself holds a quoted heredoc's delimiter line, ending the body early: the TypeScript rule's
   accepted miss.
 */
public struct SecurityNoInterpolatedShellCommand: TypedFileRule {
    public let name = "cohere-swift/security-no-interpolated-shell-command"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* The POSIX shells whose `-c` takes a command string, by the last component of the program's path. */
    static let shells: Set<String> = ["sh", "bash", "zsh", "dash", "ksh", "mksh"]

    /* Long options that take no argument, so the reading can step over them. Any other long option stops it. */
    static let longOptions: Set<String> = [
        "--login", "--interactive", "--norc", "--noprofile", "--noediting", "--posix", "--restricted", "--verbose",
    ]

    /* Long options whose argument is the next element. */
    static let longOptionsWithArgument: Set<String> = ["--rcfile", "--init-file"]

    /*
     How deep the walk through interpolations, concatenations and bindings goes. A chain this long is never
     written by hand; past it the command is declined, a finding missed and never one invented.
     */
    static let depthLimit = 16

    /* Every argument list here is labeled or assigned `arguments`, and every shell's name holds `sh`. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("arguments") && file.source.contains("sh")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        /* Folded, so `process.arguments = [...]` is one assignment and `"a" + b + "c"` one concatenation. Folding moves no token. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        guard !visitor.assignments.isEmpty || !visitor.calls.isEmpty else { return [] }
        /* Loaded only when there is something to judge. Without it, a call's or a property's type is unread: a miss, never a finding. */
        let judge = Judge(file: file, symbols: symbols, demangler: try? SwiftDemangler.shared())
        var lists: [ExprSyntax] = []
        for (assignment, member) in visitor.assignments where judge.runsShell(assignment, receiverOf: member) {
            lists.append(assignment.rightOperand)
        }
        for call in visitor.calls {
            /* Filtered as an array of the call's own nodes: filtering the syntax collection builds a detached copy, whose positions start at zero. */
            let arguments = Array(call.arguments).filter { $0.label?.text == "arguments" }
            guard arguments.count == 1, let list = arguments.first,
                call.arguments.contains(where: {
                    $0.label?.text != "arguments" && judge.isShell($0.expression, depth: 0)
                })
            else {
                continue
            }
            lists.append(list.expression)
        }
        return lists.compactMap { list in
            guard let elements = judge.arrayElements(list, depth: 0),
                let script = Self.script(in: elements, judge: judge),
                let built = judge.built(script, depth: 0), judge.judge(built, depth: 0) == .unsafe
            else {
                return nil
            }
            return file.finding(
                at: script,
                rule: name,
                message: RuleMessages.SecurityNoInterpolatedShellCommand.interpolatedShellCommand(),
            )
        }
    }

    /* Every `receiver.arguments = ...` and every call with an `arguments:` argument, the two places a shell is given its arguments. */
    final class Visitor: SyntaxVisitor {
        private(set) var assignments: [(InfixOperatorExprSyntax, MemberAccessExprSyntax)] = []
        private(set) var calls: [FunctionCallExprSyntax] = []

        override func visit(_ node: InfixOperatorExprSyntax) -> SyntaxVisitorContinueKind {
            if node.operator.is(AssignmentExprSyntax.self),
                let member = node.leftOperand.as(MemberAccessExprSyntax.self), member.base != nil,
                member.declName.baseName.text == "arguments"
            {
                assignments.append((node, member))
            }
            return .visitChildren
        }

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            if node.arguments.contains(where: { $0.label?.text == "arguments" }) {
                calls.append(node)
            }
            return .visitChildren
        }
    }

    /* What the rule concludes about a command, a piece of one, or a value spliced into one. */
    enum Verdict: Equatable {
        /* Every piece the shell parses is literal text, a number or a boolean. */
        case safe
        /* Some piece the shell parses is text that is not a literal. */
        case unsafe
        /* Some piece's type cannot be read, and no unsafe piece comes before it. */
        case unknown
        /* The rule cannot tell which pieces the shell parses (a heredoc it cannot read), so it says nothing. */
        case declined
    }

    /* What a value's type says about its text: text, no shell syntax at all (a number, a boolean), or unread. */
    enum Kind: Equatable {
        case text
        case inert
        case unknown
    }

    /*
     One stretch of a flattened command: text the author wrote, a value whose text is not known here (and whether it
     is an operand of a `+` that concatenates text), or an interpolation `String` does not define, judged unknown.
     */
    enum Piece {
        case text(String)
        case value(ExprSyntax, isConcatenated: Bool)
        case opaque

        var text: String? {
            if case .text(let text) = self {
                return text
            }
            return nil
        }
    }

    /* What a name is bound to, as far as this file shows. */
    enum Binding {
        /* `let name = value`, `if let name = value`, `guard let name = value`: the name only ever holds the value. */
        case constant(ExprSyntax)
        /* `var name = value`: the value's type, not its text, since it may be reassigned. */
        case variable(ExprSyntax)
        /* A declared type: a parameter, `let name: T` assigned later, `var name: T`. */
        case typed(TypeSyntax)
        /* Bound by a closure, a loop, a pattern or the language (`newValue`, `error`): nothing is known. */
        case opaque
    }

    static let textTypes: Set<String> = ["String", "Substring", "Character", "NSString"]
    static let inertTypes: Set<String> = [
        "Int", "Int8", "Int16", "Int32", "Int64", "UInt", "UInt8", "UInt16", "UInt32", "UInt64", "Double", "Float",
        "Float16", "CGFloat", "Bool",
    ]
    static let textDemangled: Set<String> = ["Swift.String", "Swift.Substring", "Swift.Character", "__C.NSString"]
    static let inertDemangled: Set<String> = Set(
        [
            "Int", "Int8", "Int16", "Int32", "Int64", "UInt", "UInt8", "UInt16", "UInt32", "UInt64", "Double", "Float",
            "Float16", "Bool",
        ].map { "Swift.\($0)" }
            + ["CoreFoundation.CGFloat", "CoreGraphics.CGFloat"]
    )

    /* The shell's arguments read as the shell reads them: the command string after the options, when a `c` among them asks for one. */
    static func script(in elements: [ExprSyntax], judge: Judge) -> ExprSyntax? {
        var index = 0
        var takesCommand = false
        while index < elements.count {
            guard let option = judge.literalText(elements[index], depth: 0) else {
                /* Not literal: the command string when a `c` came before it, unless it starts like an option, which would make the command a later element. */
                let first = judge.flatten(elements[index], depth: 0)?.first?.text
                guard takesCommand, first?.hasPrefix("-") != true, first?.hasPrefix("+") != true else { return nil }
                break
            }
            if option == "--" || option == "-" {
                index += 1
                break
            }
            guard option.count > 1, option.hasPrefix("-") || option.hasPrefix("+") else { break }
            if option.hasPrefix("--") {
                if Self.longOptionsWithArgument.contains(option) {
                    index += 2
                }
                else if Self.longOptions.contains(option) {
                    index += 1
                }
                else {
                    return nil
                }
                continue
            }
            let letters = option.dropFirst()
            if option.hasPrefix("-"), letters.contains("c") {
                takesCommand = true
            }
            /* `-o NAME` and `-O NAME` (and their `+` forms) take the next element as the option's name. */
            index += letters.contains("o") || letters.contains("O") ? 2 : 1
        }
        guard takesCommand, index < elements.count else { return nil }
        return elements[index]
    }

    /* The reading of one file: name resolution, flattening and judging, with the symbols and the demangler it needs. */
    struct Judge {
        let file: ParsedFile
        let symbols: FileSymbols
        let demangler: SwiftDemangler?

        // MARK: The program and its arguments

        /*
         Whether `receiver.arguments = ...` gives a shell its arguments: the same receiver's `executableURL` or
         `launchPath` is set to a shell in the same statement list, and to nothing but a shell anywhere in the
         enclosing body.
         */
        func runsShell(_ assignment: InfixOperatorExprSyntax, receiverOf member: MemberAccessExprSyntax) -> Bool {
            guard let receiver = member.base?.trimmedDescription,
                let list = assignment.parent?.as(CodeBlockItemSyntax.self)?.parent?.as(CodeBlockItemListSyntax.self)
            else {
                return false
            }
            var body = Syntax(list)
            while let parent = body.parent, !Self.isBody(body) {
                body = parent
            }
            let programs = body.tokens(viewMode: .sourceAccurate).compactMap { token -> InfixOperatorExprSyntax? in
                guard token.text == "executableURL" || token.text == "launchPath",
                    let access = token.parent?.parent?.as(MemberAccessExprSyntax.self),
                    access.declName.baseName.id == token.id, access.base?.trimmedDescription == receiver,
                    let setting = access.parent?.as(InfixOperatorExprSyntax.self),
                    setting.operator.is(AssignmentExprSyntax.self), setting.leftOperand.id == access.id
                else {
                    return nil
                }
                return setting
            }
            return programs.contains { $0.parent?.parent?.id == list.id }
                && programs.allSatisfy { isShell($0.rightOperand, depth: 0) }
        }

        /* A function's, closure's or accessor's body, or the file: the scope a program's settings are searched in. */
        static func isBody(_ node: Syntax) -> Bool {
            if node.is(SourceFileSyntax.self) {
                return true
            }
            guard let block = node.as(CodeBlockSyntax.self) else {
                return node.as(CodeBlockItemListSyntax.self)?.parent?.is(ClosureExprSyntax.self) == true
            }
            return block.parent?.is(FunctionDeclSyntax.self) == true
                || block.parent?.is(InitializerDeclSyntax.self) == true
                || block.parent?.is(AccessorDeclSyntax.self) == true
                || block.parent?.is(DeinitializerDeclSyntax.self) == true
                || block.parent?.is(AccessorBlockSyntax.self) == true
        }

        /* A shell named by a literal path, as a string or `URL(fileURLWithPath:)` or `URL(filePath:)`, directly or through a `let`. */
        func isShell(_ expression: ExprSyntax, depth: Int) -> Bool {
            guard depth <= SecurityNoInterpolatedShellCommand.depthLimit else { return false }
            let expression = Self.withoutParentheses(expression)
            if let call = expression.as(FunctionCallExprSyntax.self) {
                guard ["URL", "Foundation.URL"].contains(call.calledExpression.trimmedDescription),
                    let first = call.arguments.first,
                    first.label?.text == "fileURLWithPath" || first.label?.text == "filePath"
                else {
                    return false
                }
                return isShell(first.expression, depth: depth + 1)
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil,
                case .constant(let value) = binding(of: reference)
            {
                return isShell(value, depth: depth + 1)
            }
            guard let path = literalText(expression, depth: depth) else { return false }
            return SecurityNoInterpolatedShellCommand.shells.contains(
                String(path.split(separator: "/", omittingEmptySubsequences: false).last ?? "")
            )
        }

        /* An array literal's elements, directly or through a `let`. */
        func arrayElements(_ expression: ExprSyntax, depth: Int) -> [ExprSyntax]? {
            guard depth <= SecurityNoInterpolatedShellCommand.depthLimit else { return nil }
            let expression = Self.withoutParentheses(expression)
            if let array = expression.as(ArrayExprSyntax.self) {
                return array.elements.map(\.expression)
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil,
                case .constant(let value) = binding(of: reference)
            {
                return arrayElements(value, depth: depth + 1)
            }
            return nil
        }

        // MARK: Building and flattening

        /* The expression a command is built by: an interpolation, a `+` concatenation, or a ternary with such a branch, through parentheses and `let` bindings. Nil for an opaque command. */
        func built(_ expression: ExprSyntax, depth: Int) -> ExprSyntax? {
            guard depth <= SecurityNoInterpolatedShellCommand.depthLimit else { return nil }
            let expression = Self.withoutParentheses(expression)
            if let literal = expression.as(StringLiteralExprSyntax.self) {
                return literal.segments.contains(where: { $0.is(ExpressionSegmentSyntax.self) }) ? expression : nil
            }
            if Self.isConcatenation(expression) {
                return expression
            }
            if let ternary = expression.as(TernaryExprSyntax.self) {
                return built(ternary.thenExpression, depth: depth + 1) != nil
                    || built(ternary.elseExpression, depth: depth + 1) != nil ? expression : nil
            }
            if let reference = expression.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil,
                case .constant(let value) = binding(of: reference)
            {
                return built(value, depth: depth + 1)
            }
            return nil
        }

        /* The text of an expression made only of literal text, or nil. */
        func literalText(_ expression: ExprSyntax, depth: Int) -> String? {
            guard let pieces = flatten(expression, depth: depth) else { return nil }
            var text = ""
            for piece in pieces {
                guard let part = piece.text else { return nil }
                text += part
            }
            return text
        }

        /*
         An expression's pieces in order: the text of string literals, and the pieces of every interpolation, `+`
         operand and `let` binding it is built from. Anything else is one value. Nil when the walk runs too deep or
         a literal does not parse.
         */
        func flatten(_ expression: ExprSyntax, depth: Int, isConcatenated: Bool = false) -> [Piece]? {
            guard depth <= SecurityNoInterpolatedShellCommand.depthLimit else { return nil }
            let expression = Self.withoutParentheses(expression)
            if let literal = expression.as(StringLiteralExprSyntax.self) {
                var pieces: [Piece] = []
                for segment in literal.segments {
                    switch segment {
                        case .stringSegment(let text):
                            /* A literal of this one segment, so the parser resolves its escapes and line endings as the compiler does. */
                            let alone = StringLiteralExprSyntax(
                                openingPounds: literal.openingPounds,
                                openingQuote: literal.openingQuote,
                                segments: StringLiteralSegmentListSyntax([.stringSegment(text)]),
                                closingQuote: literal.closingQuote,
                                closingPounds: literal.closingPounds,
                            )
                            guard let value = alone.representedLiteralValue else { return nil }
                            pieces.append(.text(value))
                        case .expressionSegment(let interpolation):
                            guard interpolation.expressions.count == 1, let only = interpolation.expressions.first,
                                only.label == nil
                            else {
                                pieces.append(.opaque)
                                continue
                            }
                            guard let inner = flatten(only.expression, depth: depth + 1) else { return nil }
                            pieces += inner
                    }
                }
                return pieces
            }
            if let infix = expression.as(InfixOperatorExprSyntax.self), Self.isConcatenation(expression) {
                /* A `+` with a string literal among its operands concatenates text, so every operand is text: Swift adds no number to a `String`. */
                let concatenatesText = isConcatenated || Self.concatenatesText(infix)
                guard let left = flatten(infix.leftOperand, depth: depth + 1, isConcatenated: concatenatesText),
                    let right = flatten(infix.rightOperand, depth: depth + 1, isConcatenated: concatenatesText)
                else {
                    return nil
                }
                return left + right
            }
            if case .constant(let value)? = constantBinding(expression) {
                let value = Self.withoutParentheses(value)
                if value.is(StringLiteralExprSyntax.self) || value.is(InfixOperatorExprSyntax.self)
                    || value.is(TernaryExprSyntax.self) || value.is(DeclReferenceExprSyntax.self)
                    || value.is(MemberAccessExprSyntax.self)
                {
                    return flatten(value, depth: depth + 1, isConcatenated: isConcatenated)
                }
            }
            return [.value(expression, isConcatenated: isConcatenated)]
        }

        /* The binding of a bare name, or of a member of the enclosing type read through `self`, `Self` or the type's name, as far as this file shows. */
        func constantBinding(_ expression: ExprSyntax) -> Binding? {
            if let reference = expression.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil {
                return binding(of: reference)
            }
            if let member = expression.as(MemberAccessExprSyntax.self), member.declName.argumentNames == nil {
                return memberBinding(member)
            }
            return nil
        }

        // MARK: Judging

        /*
         Whether the shell parses a value in a command that is not a literal. It flattens the command, finds the
         values the shell reads as data (inside a quoted heredoc's body), and judges the rest in order. A value
         whose type cannot be read could open a heredoc the scan never saw, so no value after it is judged.
         */
        func judge(_ expression: ExprSyntax, depth: Int) -> Verdict {
            guard depth <= SecurityNoInterpolatedShellCommand.depthLimit,
                let pieces = flatten(expression, depth: depth), let quoted = Self.quotedHeredocValues(pieces)
            else {
                return .declined
            }
            var verdict = Verdict.safe
            var sawUnknown = false
            for (index, piece) in pieces.enumerated() where !quoted.contains(index) {
                let pieceVerdict: Verdict
                switch piece {
                    case .text:
                        continue
                    case .opaque:
                        pieceVerdict = .unknown
                    case .value(let value, let isConcatenated):
                        pieceVerdict = judgeValue(value, depth: depth, isConcatenated: isConcatenated)
                }
                switch pieceVerdict {
                    case .declined:
                        return .declined
                    case .unsafe:
                        if !sawUnknown {
                            verdict = .unsafe
                        }
                    case .unknown:
                        sawUnknown = true
                    case .safe:
                        break
                }
            }
            return verdict == .safe && sawUnknown ? .unknown : verdict
        }

        /*
         One value the shell parses. A ternary or a `??` is judged alternative by alternative, each as a command of
         its own, and one that may open a heredoc is declined, since its body would run on past it into text judged
         without it. A `let` is judged by what it is bound to. Anything else by its type: text is unsafe, a number
         or a boolean safe. An operand of a `+` that concatenates text is text whatever its type reads as, unless it
         may be a `let` of ours declared in another file, bound to a literal there.
         */
        func judgeValue(_ value: ExprSyntax, depth: Int, isConcatenated: Bool = false) -> Verdict {
            guard depth <= SecurityNoInterpolatedShellCommand.depthLimit else { return .declined }
            let value = Self.withoutWrappers(value)
            if let alternatives = Self.alternatives(value) {
                var verdict = Verdict.safe
                for alternative in alternatives {
                    guard let pieces = flatten(alternative, depth: depth + 1) else { return .declined }
                    if pieces.contains(where: { $0.text?.contains("<<") == true }) {
                        return .declined
                    }
                    switch judge(alternative, depth: depth + 1) {
                        case .declined:
                            return .declined
                        case .unsafe:
                            verdict = .unsafe
                        case .unknown:
                            if verdict == .safe {
                                verdict = .unknown
                            }
                        case .safe:
                            break
                    }
                }
                return verdict
            }
            if value.is(StringLiteralExprSyntax.self) || Self.isConcatenation(value) {
                return judge(value, depth: depth + 1)
            }
            let verdict: Verdict
            switch constantBinding(value) {
                case .constant(let bound)?:
                    return judgeValue(bound, depth: depth + 1, isConcatenated: isConcatenated)
                case .variable(let initializer)?:
                    verdict = Self.verdict(kind(initializer, depth: depth + 1))
                case .typed(let type)?:
                    verdict = Self.verdict(Self.kind(of: type))
                case .opaque?:
                    verdict = .unknown
                case nil:
                    /* A property of our own declared in another file may be a `let` bound to a literal there, which a `const` would make safe. */
                    verdict = Self.verdict(kind(value, depth: depth + 1, ownedPropertiesAreUnknown: true))
                    if verdict == .unknown, isConcatenated, mayBeOwnedConstant(value) {
                        return .unknown
                    }
            }
            return verdict == .unknown && isConcatenated ? .unsafe : verdict
        }

        /* A name or a property read that resolves to a declaration of ours, or to nothing the index recorded: it may be a `let` bound to a literal somewhere this file does not show. */
        func mayBeOwnedConstant(_ value: ExprSyntax) -> Bool {
            let value = Self.withoutWrappers(value)
            guard value.is(DeclReferenceExprSyntax.self) || value.is(MemberAccessExprSyntax.self),
                let token = Self.nameToken(value)
            else { return false }
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            guard let resolved = symbols.reference(line: location.line, column: location.column) else { return true }
            return symbols.isOwned(resolved)
        }

        /* What a value's type says about its text. */
        func kind(_ value: ExprSyntax, depth: Int, ownedPropertiesAreUnknown: Bool = false) -> Kind {
            guard depth <= SecurityNoInterpolatedShellCommand.depthLimit else { return .unknown }
            let value = Self.withoutWrappers(value)
            if value.is(StringLiteralExprSyntax.self) {
                return .text
            }
            if value.is(IntegerLiteralExprSyntax.self) || value.is(FloatLiteralExprSyntax.self)
                || value.is(BooleanLiteralExprSyntax.self)
            {
                return .inert
            }
            if let alternatives = Self.alternatives(value) ?? Self.operands(value) {
                let kinds = alternatives.map { kind($0, depth: depth + 1) }
                if kinds.contains(.text) {
                    return .text
                }
                return kinds.allSatisfy { $0 == .inert } ? .inert : .unknown
            }
            switch constantBinding(value) {
                case .constant(let bound)?, .variable(let bound)?:
                    return kind(bound, depth: depth + 1)
                case .typed(let type)?:
                    return Self.kind(of: type)
                case .opaque?:
                    return .unknown
                case nil:
                    break
            }
            if let call = value.as(FunctionCallExprSyntax.self) {
                let called = call.calledExpression
                if called.trimmedDescription == "String" || called.trimmedDescription == "String.init"
                    || called.trimmedDescription == "Swift.String"
                {
                    return .text
                }
                guard let token = Self.nameToken(called) else { return .unknown }
                return declaredKind(at: token, isCalled: true, ownedPropertiesAreUnknown: ownedPropertiesAreUnknown)
            }
            guard let token = Self.nameToken(value) else { return .unknown }
            return declaredKind(at: token, isCalled: false, ownedPropertiesAreUnknown: ownedPropertiesAreUnknown)
        }

        /* The type of the declaration a name resolves to, read from the index and demangled. */
        func declaredKind(at token: TokenSyntax, isCalled: Bool, ownedPropertiesAreUnknown: Bool) -> Kind {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            guard let resolved = symbols.reference(line: location.line, column: location.column) else {
                return .unknown
            }
            /* Every `stringBy` method and property `NSString.h` declares returns `NSString`; the rest of Objective-C is unread. */
            if resolved.symbol.hasPrefix("c:objc(cs)NSString(im)stringBy")
                || resolved.symbol.hasPrefix("c:objc(cs)NSString(py)stringBy")
            {
                return .text
            }
            guard let demangler, let declaration = demangler.declaration(ofSymbol: resolved.symbol),
                let member = RedundantTypeAnnotation.Member(
                    demangled: declaration,
                    name: resolved.name,
                    isCalled: isCalled,
                )
            else {
                return .unknown
            }
            var result = member.result
            while result.hasSuffix("?") || result.hasSuffix("!") {
                result.removeLast()
            }
            if SecurityNoInterpolatedShellCommand.inertDemangled.contains(result) {
                return .inert
            }
            guard SecurityNoInterpolatedShellCommand.textDemangled.contains(result) else { return .unknown }
            return !isCalled && ownedPropertiesAreUnknown && symbols.isOwned(resolved) ? .unknown : .text
        }

        static func verdict(_ kind: Kind) -> Verdict {
            switch kind {
                case .text:
                    return .unsafe
                case .inert:
                    return .safe
                case .unknown:
                    return .unknown
            }
        }

        /* A declared type's kind: `String`, `Substring`, `Character` or `NSString` is text, a number or `Bool` is inert, through optionals. */
        static func kind(of type: TypeSyntax) -> Kind {
            var type = type
            while true {
                if let optional = type.as(OptionalTypeSyntax.self) {
                    type = optional.wrappedType
                }
                else if let unwrapped = type.as(ImplicitlyUnwrappedOptionalTypeSyntax.self) {
                    type = unwrapped.wrappedType
                }
                else if let identifier = type.as(IdentifierTypeSyntax.self), identifier.name.text == "Optional",
                    let argument = identifier.genericArgumentClause?.arguments.first,
                    identifier.genericArgumentClause?.arguments.count == 1, case .type(let wrapped) = argument.argument
                {
                    type = wrapped
                }
                else {
                    break
                }
            }
            let name: String
            if let identifier = type.as(IdentifierTypeSyntax.self), identifier.genericArgumentClause == nil {
                name = identifier.name.text
            }
            else if let member = type.as(MemberTypeSyntax.self),
                ["Swift", "Foundation", "CoreGraphics"].contains(member.baseType.trimmedDescription),
                member.genericArgumentClause == nil
            {
                name = member.name.text
            }
            else {
                return .unknown
            }
            if SecurityNoInterpolatedShellCommand.textTypes.contains(name) {
                return .text
            }
            return SecurityNoInterpolatedShellCommand.inertTypes.contains(name) ? .inert : .unknown
        }

        // MARK: Shapes

        static func withoutParentheses(_ expression: ExprSyntax) -> ExprSyntax {
            var expression = expression
            while let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                expression = only.expression
            }
            return expression
        }

        /* Parentheses, `try`, `await` and a force unwrap change no text. */
        static func withoutWrappers(_ expression: ExprSyntax) -> ExprSyntax {
            var expression = withoutParentheses(expression)
            while true {
                if let tryExpression = expression.as(TryExprSyntax.self) {
                    expression = withoutParentheses(tryExpression.expression)
                }
                else if let awaitExpression = expression.as(AwaitExprSyntax.self) {
                    expression = withoutParentheses(awaitExpression.expression)
                }
                else if let forced = expression.as(ForceUnwrapExprSyntax.self) {
                    expression = withoutParentheses(forced.expression)
                }
                else {
                    return expression
                }
            }
        }

        static func isConcatenation(_ expression: ExprSyntax) -> Bool {
            expression.as(InfixOperatorExprSyntax.self)?.operator.as(BinaryOperatorExprSyntax.self)?.operator.text
                == "+"
        }

        /* Whether the chain of `+` a node belongs to, up to its parentheses, has a string literal among its operands. */
        static func concatenatesText(_ infix: InfixOperatorExprSyntax) -> Bool {
            var root = ExprSyntax(infix)
            while let parent = root.parent?.as(InfixOperatorExprSyntax.self), isConcatenation(ExprSyntax(parent)) {
                root = ExprSyntax(parent)
            }
            var operands = [root]
            while let operand = operands.popLast() {
                if let inner = operand.as(InfixOperatorExprSyntax.self), isConcatenation(operand) {
                    operands += [inner.leftOperand, inner.rightOperand]
                }
                else if operand.is(StringLiteralExprSyntax.self) {
                    return true
                }
            }
            return false
        }

        /* A ternary's branches, or a `??`'s two sides: the values one of which is the result. */
        static func alternatives(_ expression: ExprSyntax) -> [ExprSyntax]? {
            if let ternary = expression.as(TernaryExprSyntax.self) {
                return [ternary.thenExpression, ternary.elseExpression]
            }
            if let infix = expression.as(InfixOperatorExprSyntax.self),
                infix.operator.as(BinaryOperatorExprSyntax.self)?.operator.text == "??"
            {
                return [infix.leftOperand, infix.rightOperand]
            }
            return nil
        }

        /* A `+`'s two operands: text if either is, a number if both are. */
        static func operands(_ expression: ExprSyntax) -> [ExprSyntax]? {
            guard let infix = expression.as(InfixOperatorExprSyntax.self), isConcatenation(expression) else {
                return nil
            }
            return [infix.leftOperand, infix.rightOperand]
        }

        /* The token the index places a name at: `name`, `base.name`, or the name a call is made through. */
        static func nameToken(_ expression: ExprSyntax) -> TokenSyntax? {
            if let reference = expression.as(DeclReferenceExprSyntax.self) {
                return reference.baseName
            }
            if let member = expression.as(MemberAccessExprSyntax.self), member.base != nil {
                return member.declName.baseName
            }
            return nil
        }

        // MARK: Names

        /*
         The declaration a bare name refers to, searched outward from the use the way Swift scopes it: earlier
         statements of each enclosing block (every statement at the top of a file), the conditions of an enclosing
         `if`, `while` or earlier `guard`, a closure's, function's, initializer's or accessor's parameters, the
         members of an enclosing type. Nil when nothing in this file declares it.
         */
        func binding(of reference: DeclReferenceExprSyntax) -> Binding? {
            let name = reference.baseName.text
            var child = Syntax(reference)
            var current = child.parent
            while let node = current {
                if let found = Self.binding(of: name, in: node, below: child) {
                    return found
                }
                child = node
                current = node.parent
            }
            return nil
        }

        /* A member of the enclosing type read as `self.name`, `Self.name` or `TypeName.name`. */
        func memberBinding(_ member: MemberAccessExprSyntax) -> Binding? {
            guard let base = member.base?.trimmedDescription else { return nil }
            var current = member.parent
            while let node = current {
                if let members = Self.members(of: node) {
                    guard base == "self" || base == "Self" || base == Self.typeName(of: node) else { return nil }
                    return Self.memberBinding(member.declName.baseName.text, in: members)
                }
                current = node.parent
            }
            return nil
        }

        static func members(of node: Syntax) -> MemberBlockItemListSyntax? {
            node.asProtocol((any DeclGroupSyntax).self)?.memberBlock.members
        }

        static func typeName(of node: Syntax) -> String? {
            if let extensionDecl = node.as(ExtensionDeclSyntax.self) {
                return extensionDecl.extendedType.trimmedDescription
            }
            return node.asProtocol((any NamedDeclSyntax).self)?.name.text
        }

        static func memberBinding(_ name: String, in members: MemberBlockItemListSyntax) -> Binding? {
            for member in members {
                if let variable = member.decl.as(VariableDeclSyntax.self),
                    let found = Self.binding(of: name, in: variable)
                {
                    return found
                }
            }
            return nil
        }

        /* What one scope node binds `name` to, for a use inside its child `child`. */
        static func binding(of name: String, in node: Syntax, below child: Syntax) -> Binding? {
            if let list = node.as(CodeBlockItemListSyntax.self) {
                let statements = Array(list)
                let isFileScope = list.parent?.is(SourceFileSyntax.self) == true
                let position = statements.firstIndex { $0.id == child.id } ?? statements.count
                let visible =
                    isFileScope ? statements.filter { $0.id != child.id } : Array(statements[..<position]).reversed()
                for statement in visible {
                    if let variable = statement.item.as(VariableDeclSyntax.self),
                        let found = Self.binding(of: name, in: variable)
                    {
                        return found
                    }
                    if let guardStatement = statement.item.as(GuardStmtSyntax.self),
                        let found = Self.binding(of: name, in: Array(guardStatement.conditions))
                    {
                        return found
                    }
                }
                return nil
            }
            if let members = node.as(MemberBlockItemListSyntax.self) {
                return Self.memberBinding(name, in: members)
            }
            if let conditions = node.as(ConditionElementListSyntax.self) {
                let position = conditions.firstIndex { $0.id == child.id }
                return position.flatMap { Self.binding(of: name, in: Array(conditions[..<$0])) }
            }
            if let ifExpression = node.as(IfExprSyntax.self), ifExpression.body.id == child.id {
                return Self.binding(of: name, in: Array(ifExpression.conditions))
            }
            if let whileStatement = node.as(WhileStmtSyntax.self), whileStatement.body.id == child.id {
                return Self.binding(of: name, in: Array(whileStatement.conditions))
            }
            if let forStatement = node.as(ForStmtSyntax.self), forStatement.body.id == child.id {
                return Self.binds(forStatement.pattern, name) ? .opaque : nil
            }
            if let switchCase = node.as(SwitchCaseSyntax.self) {
                return Self.binds(switchCase.label, name) ? .opaque : nil
            }
            if let catchClause = node.as(CatchClauseSyntax.self) {
                return (catchClause.catchItems.isEmpty && name == "error") || Self.binds(catchClause.catchItems, name)
                    ? .opaque : nil
            }
            if let closure = node.as(ClosureExprSyntax.self) {
                return Self.binding(of: name, in: closure)
            }
            if let accessor = node.as(AccessorDeclSyntax.self) {
                return accessor.parameters?.name.text == name || name == "newValue" || name == "oldValue"
                    ? .opaque : nil
            }
            let parameters: FunctionParameterListSyntax?
            if let function = node.as(FunctionDeclSyntax.self) {
                parameters = function.signature.parameterClause.parameters
            }
            else if let initializer = node.as(InitializerDeclSyntax.self) {
                parameters = initializer.signature.parameterClause.parameters
            }
            else if let subscriptDecl = node.as(SubscriptDeclSyntax.self) {
                parameters = subscriptDecl.parameterClause.parameters
            }
            else {
                parameters = nil
            }
            guard let parameter = parameters?.first(where: { ($0.secondName ?? $0.firstName).text == name }) else {
                return nil
            }
            return .typed(parameter.type)
        }

        static func binding(of name: String, in closure: ClosureExprSyntax) -> Binding? {
            guard let signature = closure.signature else { return nil }
            if signature.capture?.items.contains(where: { $0.name.text == name }) == true {
                return .opaque
            }
            switch signature.parameterClause {
                case .simpleInput(let parameters)?:
                    return parameters.contains { $0.name.text == name } ? .opaque : nil
                case .parameterClause(let clause)?:
                    guard
                        let parameter = clause.parameters.first(where: { ($0.secondName ?? $0.firstName).text == name })
                    else { return nil }
                    return parameter.type.map { .typed($0) } ?? .opaque
                case nil:
                    return nil
            }
        }

        /* `let name = value` and its kin, in one declaration that may bind several names. */
        static func binding(of name: String, in variable: VariableDeclSyntax) -> Binding? {
            let isConstant = variable.bindingSpecifier.tokenKind == .keyword(.let)
            for binding in variable.bindings {
                guard let identifier = binding.pattern.as(IdentifierPatternSyntax.self) else {
                    if Self.binds(binding.pattern, name) {
                        return .opaque
                    }
                    continue
                }
                guard identifier.identifier.text == name else { continue }
                if isConstant, let value = binding.initializer?.value {
                    return .constant(value)
                }
                if let type = binding.typeAnnotation?.type {
                    return .typed(type)
                }
                if let value = binding.initializer?.value {
                    return .variable(value)
                }
                return .opaque
            }
            return nil
        }

        /* `if let name = value`, `guard var name = value`, `if case let ...`: the conditions searched last first. A shorthand `if let name` rebinds the outer name, so it is passed over. */
        static func binding(of name: String, in conditions: [ConditionElementSyntax]) -> Binding? {
            for condition in conditions.reversed() {
                if let optional = condition.condition.as(OptionalBindingConditionSyntax.self) {
                    guard let identifier = optional.pattern.as(IdentifierPatternSyntax.self) else {
                        if Self.binds(optional.pattern, name) {
                            return .opaque
                        }
                        continue
                    }
                    guard identifier.identifier.text == name, let value = optional.initializer?.value else { continue }
                    if optional.bindingSpecifier.tokenKind == .keyword(.let) {
                        return .constant(value)
                    }
                    return optional.typeAnnotation.map { .typed($0.type) } ?? .variable(value)
                }
                if condition.condition.is(MatchingPatternConditionSyntax.self), Self.binds(condition.condition, name) {
                    return .opaque
                }
            }
            return nil
        }

        /* Whether a pattern anywhere under a node names `name`. */
        static func binds(_ node: some SyntaxProtocol, _ name: String) -> Bool {
            node.tokens(viewMode: .sourceAccurate).contains { token in
                token.text == name && token.parent?.is(IdentifierPatternSyntax.self) == true
            }
        }

        // MARK: Heredocs

        /*
         The values that sit in the body of a heredoc whose delimiter is quoted (`<<'EOSQL'`, `<<"EOSQL"`,
         `<<\EOSQL`, each with an optional `-`). The shell expands nothing in such a body, so a value there is
         data. A body runs from the line after its opener to a line holding only the delimiter (after leading tabs,
         for `<<-`), and several openers on one line are read in order. An unquoted heredoc's body is expanded, so
         its values stay judged, but it is tracked so the bodies after it line up. A value stands as a NUL, which no
         command line can carry. Nil when a `<<` is followed by something that is not a delimiter, and the command
         is then declined. Reading a `<<` the shell would not (inside quotes, a comment, another language's script)
         can only mark values safe that are not: a miss, never a finding.
         */
        static func quotedHeredocValues(_ pieces: [Piece]) -> Set<Int>? {
            var bytes: [UInt8] = []
            var valueAt: [Int: Int] = [:]
            for (index, piece) in pieces.enumerated() {
                if let text = piece.text {
                    bytes += Array(text.utf8)
                }
                else {
                    valueAt[bytes.count] = index
                    bytes.append(0)
                }
            }
            struct Heredoc {
                var delimiter: [UInt8] = []
                var stripsTabs = false
                var quotedOpening = false
            }
            let newline = UInt8(ascii: "\n")
            let tab = UInt8(ascii: "\t")
            let less = UInt8(ascii: "<")
            var quoted = Set<Int>()
            var pending: [Heredoc] = []
            var body: Heredoc?
            var lineStart = 0
            while lineStart <= bytes.count {
                let lineEnd = bytes[lineStart...].firstIndex(of: newline) ?? bytes.count
                let line = Array(bytes[lineStart..<lineEnd])
                if let open = body {
                    let candidate = open.stripsTabs ? Array(line.drop { $0 == tab }) : line
                    if candidate == open.delimiter {
                        body = nil
                    }
                    else if open.quotedOpening {
                        for offset in lineStart..<lineEnd {
                            if let index = valueAt[offset] {
                                quoted.insert(index)
                            }
                        }
                    }
                }
                else {
                    var offset = 0
                    while offset + 1 < line.count {
                        guard
                            let opener = (offset..<(line.count - 1)).first(where: {
                                line[$0] == less && line[$0 + 1] == less
                            })
                        else { break }
                        var cursor = opener + 2
                        if cursor < line.count, line[cursor] == less {
                            /* `<<<` is a here-string, not a heredoc. */
                            offset = cursor + 1
                            continue
                        }
                        var opened = Heredoc()
                        if cursor < line.count, line[cursor] == UInt8(ascii: "-") {
                            opened.stripsTabs = true
                            cursor += 1
                        }
                        while cursor < line.count, line[cursor] == UInt8(ascii: " ") || line[cursor] == tab {
                            cursor += 1
                        }
                        guard let delimiter = Self.heredocDelimiter(Array(line[cursor...])) else { return nil }
                        opened.delimiter = delimiter.word
                        opened.quotedOpening = delimiter.isQuoted
                        pending.append(opened)
                        offset = cursor + delimiter.length
                    }
                }
                if body == nil, !pending.isEmpty, lineEnd < bytes.count {
                    body = pending.removeFirst()
                }
                lineStart = lineEnd + 1
            }
            return quoted
        }

        /* The delimiter word at the start of the text after `<<`: `'word'`, `"word"`, `\word` or a bare word, how many bytes it spans, and whether it was quoted. */
        static func heredocDelimiter(_ text: [UInt8]) -> (word: [UInt8], length: Int, isQuoted: Bool)? {
            guard let first = text.first else { return nil }
            if first == UInt8(ascii: "'") || first == UInt8(ascii: "\"") {
                guard let closing = text.dropFirst().firstIndex(of: first), closing > 1 else { return nil }
                let word = Array(text[1..<closing])
                guard !word.contains(0), !word.contains(UInt8(ascii: "\n")) else { return nil }
                return (word, closing + 1, true)
            }
            if first == UInt8(ascii: "\\") {
                let length = Self.wordLength(Array(text.dropFirst()))
                guard length > 0 else { return nil }
                return (Array(text[1...length]), length + 1, true)
            }
            let length = Self.wordLength(text)
            guard length > 0 else { return nil }
            return (Array(text[..<length]), length, false)
        }

        /* How many letters, digits and underscores a text starts with. */
        static func wordLength(_ text: [UInt8]) -> Int {
            text.prefix { character in
                character == UInt8(ascii: "_") || (UInt8(ascii: "0")...UInt8(ascii: "9")).contains(character)
                    || (UInt8(ascii: "a")...UInt8(ascii: "z")).contains(character)
                    || (UInt8(ascii: "A")...UInt8(ascii: "Z")).contains(character)
            }.count
        }
    }
}
