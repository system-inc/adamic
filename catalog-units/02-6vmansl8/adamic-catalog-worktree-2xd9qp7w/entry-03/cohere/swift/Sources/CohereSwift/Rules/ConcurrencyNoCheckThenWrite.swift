import SwiftOperators
import SwiftSyntax

/*
 No claiming a file name in two steps: a loop that looks for a free name by asking whether each candidate exists,
 then a write after the loop that creates the chosen name without refusing a file already there.

     var url = folder.appendingPathComponent("\(base).png")
     var number = 2
     while FileManager.default.fileExists(atPath: url.path) {
         url = folder.appendingPathComponent("\(base) \(number).png")
         number += 1
     }
     try data.write(to: url)

 The check and the write are two steps, so two saves running at once (two tasks, or two processes) can both find
 the same name free, and the second write silently replaces the first's file. The repair claims the name in the
 step that creates it: `data.write(to: url, options: .withoutOverwriting)` fails with
 `CocoaError.fileWriteFileExists` when the name is taken (so does `FileManager.moveItem` from a temporary file),
 and the loop tries the next name on that error instead of checking first.

 It ports `nexus/concurrency-no-check-then-write`, condition for condition, each one required:
 - A loop that stops on a free name. The TypeScript rule's first shape, a loop whose condition is exactly the
   check: a `while` whose one condition is the check, or a `repeat { } while` whose condition is. Swift has no
   C-style `for`. The check is `fileExists(atPath:)` with its one argument, resolved through the index the build
   wrote to Foundation's `FileManager` (`c:objc(cs)NSFileManager(im)fileExistsAtPath:`), whatever the receiver is
   spelled; a method of ours with that name never matches. The TypeScript rule's second shape, a `try` whose
   check throws on a missing name and whose `catch` is only `break`, has no Swift idiom anyone writes and is not
   ported (a known miss, below).
 - The checked path is built from a counter. Some binding the path reads is changed inside the loop by `+=` or
   `-=` (Swift has no `++`), directly or through a binding assigned or declared in the loop from one that is. That
   is what makes the loop a search for a free name rather than a wait for a lock file to go away, a different bug
   the rule leaves alone.
 - A write after the loop creates that same path without exclusive create. The write is in a statement that runs
   after the loop in the same function: climbing out through blocks, branches, `do` and `switch` cases, never
   past an enclosing loop (where the next round searches again), a `guard`'s `else` (which leaves the scope), a
   `defer`, or a closure or nested function (whose code runs on its own schedule), and never inside one. It
   resolves to one of Foundation's writes that replace a file: `Data.write(to:options:)`, `NSData`'s
   `write(to:options:)`, `write(toFile:options:)`, `write(to:atomically:)` and `write(toFile:atomically:)`,
   `StringProtocol`'s `write(to:atomically:encoding:)` and `write(toFile:atomically:encoding:)`, and
   `FileManager.createFile(atPath:contents:attributes:)`, which overwrites a file already there. It is not
   exclusive only when the source says so, as the TypeScript rule reads `flag`: no `options:`, or options written
   as `.atomic`, `[]` or a list of `.atomic` (which writes a temporary file and renames it over the name, so it
   replaces too). Options the source does not show (a variable, a qualified name, any other member) leave the
   write alone. Node's `copyFile` and `rename` replace, so the TypeScript rule counts them; `FileManager.copyItem`
   and `moveItem` refuse an existing destination, so they claim the name, and are the repair rather than the bug.
 - Same path, proven, by the TypeScript rule's definition. The two paths are the same expression once `let`
   locals are read through their initializers: names resolving to the same local binding, string literals spelled
   alike with the same interpolated expressions, integer literals, `+`, and Foundation's pure path builders, which
   take the place of `path.join`: `URL.appendingPathComponent(_:)`, `appendingPathComponent(_:isDirectory:)`,
   `appendingPathExtension(_:)`, `appending(path:)`, `appending(component:)`, and `NSString`'s
   `appendingPathComponent(_:)` and `appendingPathExtension(_:)`, each resolved through the index and compared
   receiver, labels and arguments. One mapping Swift needs and Node does not: the check takes a path string and
   most writes take a URL, so a write to `url` matches a check of `url.path`, and `.path` resolved to Foundation's
   `URL.path` is read as pure wherever it appears (a property of a value type, unlike the object property the
   TypeScript rule refuses). `appendingPathComponent(_:)` may look at the disk, but only to decide a trailing
   slash, which never changes which file the path names. A `let` is read through only when it was declared after
   the loop in the loop's function; one declared earlier is compared as itself, because its value was fixed
   before the counter moved. A name resolves by Swift's own scope rules, read from the tree: the nearest local
   `let` or `var`, parameter, `if let`, `guard let`, `for`, `case` or `catch` binding, closure parameter or
   capture. A name that is not one of those (a property, a global, a local function or type) is never proven.
 - The path did not move in between. A `let`, a non-`inout` parameter and every pattern or condition binding
   cannot change. A `var`, an `inout` parameter or an untyped closure parameter is written only in the function
   holding the loop (never from a nested closure or function, which could run during an `await`), and never
   between the end of the loop and the write. A write is an assignment or compound assignment to the binding or
   through its members, an `&` passing it `inout`, or a method called on it or on a member of it (which may be
   `mutating`) other than the pure path builders.

 It reports on the check, as the TypeScript rule does, with a message naming the write, so the reader can find the
 second half of the race. No fix: the repair changes the shape of the loop, and the write is often far from it.

 What it accepts as safe, and why: a write with `.withoutOverwriting` or with options it cannot read (either may
 claim the name), `copyItem` and `moveItem` (they refuse an existing destination), a loop with no counter (a
 wait, not a search), a loop that stops on a name that exists (`while !fileExists`), and every path it cannot
 prove equal to the checked one.

 Known misses, every one a finding not made and never one invented:
 - A check with no loop (`if !fileExists(atPath:) { write }`), the TypeScript rule's measured decline: nearly all
   create-if-missing and cache files, where a second writer writes the same bytes.
 - A check that is not the loop's whole condition. `Stage/StagePane+Summoning.swift:309` in Presence loops
   `while fileExists(atPath: url.path), (try? CharacterCard.read(from: url))?.bodyId != bodyId`, which also stops
   on a name it means to replace, so the condition is not exactly the check. It also writes through
   `CharacterCard.write`, ours (below). Its loop and write run on the main actor with no suspension between them,
   so only two processes can race it.
 - A write through a function of ours, even one that calls `Data.write` (`CharacterCard.write`), and a free name
   returned to a caller that writes it later: `Studio/FaceSets.swift:126` (`unusedUrl(for:)`, written by
   `CharacterCard.write` in `StudioModel`) and `Studio/Bodies/BodyImport.swift:302` (`free(_:)`, whose caller
   lands the file with `moveItem`, which refuses an existing destination, so that one is not a race at all).
   One function is all the rule reads.
 - The throwing shape (`do { _ = try url.checkResourceIsReachable(); number += 1 } catch { break }`), a check
   with a second argument (`fileExists(atPath:isDirectory:)`), a check reached through optional chaining, and
   a write reached through it (`data?.write(to:)`).
 - A path read through a property other than `URL.path` (`self.folder`, `CharacterCard.folder`, `input.directory`),
   a global, a call that is not a pure path builder (a function of ours, `URL(fileURLWithPath:)`, which resolves
   a relative path against the working directory as `path.resolve` does), a path builder given a `directoryHint`,
   a cast, and literals spelled differently.
 - A `var` the rule cannot prove still: written from a closure, or with any method called on it between the loop
   and the write, even one that does not mutate.
 - A loop at file scope in `main.swift`, whose bindings are globals.
 - `FileManager.createDirectory(at:withIntermediateDirectories: true)`, which succeeds on a directory already
   there, `FileHandle`, `fopen`, `open` without `O_EXCL`, image destinations and every other writer: the
   TypeScript rule's list has no counterpart for them.
 */
public struct ConcurrencyNoCheckThenWrite: TypedFileRule {
    public let name = "cohere-swift/concurrency-no-check-then-write"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* `FileManager.fileExists(atPath:)`, as the index names the Objective-C method it imports. */
    static let fileExists = "c:objc(cs)NSFileManager(im)fileExistsAtPath:"

    /* `URL.path`, the string a check of a URL's file is given. */
    static let urlPath = "s:10Foundation3URLV4pathSSvp"

    /* Whether a write's destination is a URL (matched against a check of its `.path`) or a path string. */
    enum Destination {
        case url
        case path
    }

    /* One write the rule can read: the label of its destination argument, what that argument is, and the label of its options, when it takes any. */
    struct Write {
        let destinationLabel: String
        let destination: Destination
        let optionsLabel: String?
    }

    /* Foundation's writes that replace a file already at the destination, by the symbol the index gives each. */
    static let writes: [String: Write] = [
        "s:10Foundation4DataV5write2to7optionsyAA3URLV_So20NSDataWritingOptionsVtKF": Write(
            destinationLabel: "to",
            destination: .url,
            optionsLabel: "options",
        ),
        "c:objc(cs)NSData(im)writeToURL:options:error:": Write(
            destinationLabel: "to",
            destination: .url,
            optionsLabel: "options",
        ),
        "c:objc(cs)NSData(im)writeToFile:options:error:": Write(
            destinationLabel: "toFile",
            destination: .path,
            optionsLabel: "options",
        ),
        "c:objc(cs)NSData(im)writeToURL:atomically:": Write(
            destinationLabel: "to",
            destination: .url,
            optionsLabel: nil,
        ),
        "c:objc(cs)NSData(im)writeToFile:atomically:": Write(
            destinationLabel: "toFile",
            destination: .path,
            optionsLabel: nil,
        ),
        "s:Sy10FoundationE5write2to10atomically8encodingyAA3URLV_SbSSAAE8EncodingVtKF": Write(
            destinationLabel: "to",
            destination: .url,
            optionsLabel: nil,
        ),
        "s:Sy10FoundationE5write6toFile10atomically8encodingyqd___SbSSAAE8EncodingVtKSyRd__lF": Write(
            destinationLabel: "toFile",
            destination: .path,
            optionsLabel: nil,
        ),
        "c:objc(cs)NSFileManager(im)createFileAtPath:contents:attributes:": Write(
            destinationLabel: "atPath",
            destination: .path,
            optionsLabel: nil,
        ),
    ]

    /* Foundation's pure path builders, the counterpart of `path.join`: the same receiver and arguments give the same path. */
    static let pathBuilders: Set<String> = [
        "s:10Foundation3URLV22appendingPathComponentyACSSF",
        "s:10Foundation3URLV22appendingPathComponent_11isDirectoryACSS_SbtF",
        "s:10Foundation3URLV22appendingPathExtensionyACSSF",
        "s:10Foundation3URLV9appending4path13directoryHintACx_AC09DirectoryF0OtSyRzlF",
        "s:10Foundation3URLV9appending9component13directoryHintACx_AC09DirectoryF0OtSyRzlF",
        "c:objc(cs)NSString(im)stringByAppendingPathComponent:",
        "c:objc(cs)NSString(im)stringByAppendingPathExtension:",
    ]

    /* Operators ending in `=` that compare rather than assign. Every other one is read as a write to its left side. */
    static let comparisons: Set<String> = ["==", "!=", "<=", ">=", "===", "!==", "~="]

    /* How deep a comparison reads through `let`s and nested expressions. A chain this long is never written by hand; past it the paths are not proven equal. */
    static let depthLimit = 16

    /* Every flagged loop holds a `fileExists` check, and every write the rule knows is a `write` or a `createFile`. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("fileExists") && (file.source.contains("write") || file.source.contains("createFile"))
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        /* Folded, so `number += 1` and `url = ...` are one assignment each and `a + b` one sum. Folding moves no token. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        let analysis = Analysis(file: file, symbols: symbols)
        return visitor.loops.compactMap { loop, condition in
            guard let race = analysis.race(loop: loop, condition: condition) else { return nil }
            return file.finding(
                at: race.check,
                rule: name,
                message: RuleMessages.ConcurrencyNoCheckThenWrite.checkThenWrite(
                    write: Self.spelledName(of: race.write)
                ),
            )
        }
    }

    /* The write as the message names it, with the labels it was called with: `write(to:)`, `createFile(atPath:contents:)`. */
    static func spelledName(of call: FunctionCallExprSyntax) -> String {
        let callee =
            call.calledExpression.as(MemberAccessExprSyntax.self)?.declName.baseName.text
            ?? call.calledExpression.trimmedDescription
        return callee + "(" + call.arguments.map { ($0.label?.text ?? "_") + ":" }.joined() + ")"
    }

    /* Every loop whose whole condition is one expression: a `while` with one condition, and every `repeat { } while`. */
    final class Visitor: SyntaxVisitor {
        private(set) var loops: [(loop: Syntax, condition: ExprSyntax)] = []

        override func visit(_ node: WhileStmtSyntax) -> SyntaxVisitorContinueKind {
            if node.conditions.count == 1, let only = node.conditions.first,
                case .expression(let expression) = only.condition
            {
                loops.append((Syntax(node), expression))
            }
            return .visitChildren
        }

        override func visit(_ node: RepeatStmtSyntax) -> SyntaxVisitorContinueKind {
            loops.append((Syntax(node), node.condition))
            return .visitChildren
        }
    }

    /* A local binding a name resolves to. Two names are the same binding when their declarations' ids are equal. */
    struct Declaration {
        /* The declaring identifier's token, or the clause or accessor that binds an implicit `error` or `newValue`. */
        let id: SyntaxIdentifier
        let name: String
        /* A `var`, an `inout` parameter or an untyped closure parameter: its writes are checked. Everything else cannot change. */
        let isVariable: Bool
        /* `let name = value`, the value a comparison may read through. */
        let initializer: ExprSyntax?
        /* Where it is declared, for its position and the function it lives in. */
        let node: Syntax
    }

    /* What one scope says about a name: bound here, or bound by something the rule does not read (which stops the search, since an outer binding would be the wrong one). */
    enum Lookup {
        case declared(Declaration)
        case unknown
    }

    /* The reading of one loop at a time: name resolution, the counter, the writes after it, and the path proof. */
    struct Analysis {
        let file: ParsedFile
        let symbols: FileSymbols

        /* The declaration a name token resolves to in the index, when exactly one written reference was recorded there. */
        func resolved(_ token: TokenSyntax) -> String? {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            return symbols.reference(line: location.line, column: location.column)?.symbol
        }

        /* The check and the write, when the loop is a free-name search followed by a write that claims the name without refusing a file there. */
        func race(
            loop: Syntax,
            condition: ExprSyntax,
        ) -> (check: FunctionCallExprSyntax, write: FunctionCallExprSyntax)? {
            guard let check = checkedPath(condition) else { return nil }
            let function = ConcurrencyNoCheckThenWrite.enclosingFunction(loop)
            guard reads(check.path, anyOf: counters(in: loop)) else { return nil }
            for statement in ConcurrencyNoCheckThenWrite.followingStatements(loop) {
                if let write = firstWrite(in: statement, checked: check.path, loop: loop, function: function) {
                    return (check.call, write)
                }
            }
            return nil
        }

        // MARK: The check and the counter

        /* `receiver.fileExists(atPath: path)` resolved to FileManager's, with that one argument: the call and the path it checks. */
        func checkedPath(_ condition: ExprSyntax) -> (call: FunctionCallExprSyntax, path: ExprSyntax)? {
            guard let call = ConcurrencyNoCheckThenWrite.withoutParentheses(condition).as(FunctionCallExprSyntax.self),
                call.trailingClosure == nil,
                let member = call.calledExpression.as(MemberAccessExprSyntax.self),
                member.declName.baseName.text == "fileExists", member.declName.argumentNames == nil,
                call.arguments.count == 1, let argument = call.arguments.first, argument.label?.text == "atPath",
                resolved(member.declName.baseName) == ConcurrencyNoCheckThenWrite.fileExists
            else {
                return nil
            }
            return (call, argument.expression)
        }

        /* The loop's counters, the bindings it changes with `+=` or `-=`, and every binding assigned or declared in the loop from one of them. */
        func counters(in loop: Syntax) -> Set<SyntaxIdentifier> {
            var varying: Set<SyntaxIdentifier> = []
            var assignments: [(target: SyntaxIdentifier, value: ExprSyntax)] = []
            ConcurrencyNoCheckThenWrite.forEachOwnNode(loop) { node in
                if let infix = node.as(InfixOperatorExprSyntax.self),
                    let target = ConcurrencyNoCheckThenWrite.withoutParentheses(infix.leftOperand).as(
                        DeclReferenceExprSyntax.self
                    ),
                    let declared = declaration(of: target)
                {
                    if let operation = infix.operator.as(BinaryOperatorExprSyntax.self),
                        ["+=", "-="].contains(operation.operator.text)
                    {
                        varying.insert(declared.id)
                    }
                    else if infix.operator.is(AssignmentExprSyntax.self) {
                        assignments.append((declared.id, infix.rightOperand))
                    }
                }
                if let binding = node.as(PatternBindingSyntax.self),
                    let identifier = binding.pattern.as(IdentifierPatternSyntax.self),
                    let value = binding.initializer?.value
                {
                    assignments.append((identifier.identifier.id, value))
                }
            }
            var changed = true
            while changed {
                changed = false
                for assignment in assignments
                where !varying.contains(assignment.target) && reads(assignment.value, anyOf: varying) {
                    varying.insert(assignment.target)
                    changed = true
                }
            }
            return varying
        }

        /* Whether an expression reads, by name, any binding in a set, not counting names inside closures. */
        func reads(_ expression: ExprSyntax, anyOf bindings: Set<SyntaxIdentifier>) -> Bool {
            guard !bindings.isEmpty else { return false }
            var found = false
            ConcurrencyNoCheckThenWrite.forEachOwnNode(Syntax(expression)) { node in
                if !found, let reference = node.as(DeclReferenceExprSyntax.self),
                    let declared = declaration(of: reference), bindings.contains(declared.id)
                {
                    found = true
                }
            }
            return found
        }

        // MARK: The write

        /* The first write in a statement that creates the checked path without refusing a file there. */
        func firstWrite(
            in statement: CodeBlockItemSyntax,
            checked: ExprSyntax,
            loop: Syntax,
            function: Syntax?,
        ) -> FunctionCallExprSyntax? {
            var found: FunctionCallExprSyntax?
            ConcurrencyNoCheckThenWrite.forEachOwnNode(Syntax(statement)) { node in
                guard found == nil, let call = node.as(FunctionCallExprSyntax.self),
                    let destination = nonExclusiveDestination(call)
                else { return }
                if samePath(
                    checked: checked,
                    destination: destination.expression,
                    kind: destination.kind,
                    loop: loop,
                    function: function,
                    write: call,
                ) {
                    found = call
                }
            }
            return found
        }

        /* The destination of a call that is one of Foundation's replacing writes, when its source shows it does not refuse an existing file. */
        func nonExclusiveDestination(_ call: FunctionCallExprSyntax) -> (expression: ExprSyntax, kind: Destination)? {
            guard call.trailingClosure == nil, call.additionalTrailingClosures.isEmpty else { return nil }
            let nameToken: TokenSyntax
            if let member = call.calledExpression.as(MemberAccessExprSyntax.self) {
                guard !(member.base.map(ConcurrencyNoCheckThenWrite.isOptionallyChained) ?? false) else { return nil }
                nameToken = member.declName.baseName
            }
            else if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self) {
                nameToken = reference.baseName
            }
            else {
                return nil
            }
            guard let symbol = resolved(nameToken), let write = ConcurrencyNoCheckThenWrite.writes[symbol],
                let destination = call.arguments.first(where: { $0.label?.text == write.destinationLabel })
            else {
                return nil
            }
            if let optionsLabel = write.optionsLabel,
                let options = call.arguments.first(where: { $0.label?.text == optionsLabel }),
                !ConcurrencyNoCheckThenWrite.optionsReplace(options.expression)
            {
                return nil
            }
            return (destination.expression, write.destination)
        }

        // MARK: The path proof

        /* Whether a write's destination is proven to be the checked path when the write runs. */
        func samePath(
            checked: ExprSyntax,
            destination: ExprSyntax,
            kind: Destination,
            loop: Syntax,
            function: Syntax?,
            write: FunctionCallExprSyntax,
        ) -> Bool {
            var bindings: [Declaration] = []
            let matched: Bool
            switch kind {
                case .path:
                    matched = same(checked, destination, loop: loop, function: function, bindings: &bindings, depth: 0)
                case .url:
                    /* The check is given the URL's `.path`, the write the URL itself. */
                    guard let access = expand(checked, loop: loop, function: function).as(MemberAccessExprSyntax.self),
                        isUrlPath(access), let base = access.base
                    else { return false }
                    matched = same(base, destination, loop: loop, function: function, bindings: &bindings, depth: 0)
            }
            return matched && bindings.allSatisfy { holdsStill($0, loop: loop, function: function, write: write) }
        }

        /* An expression with parentheses removed and any `let` declared after the loop, in the loop's function, read through to its initializer. */
        func expand(_ expression: ExprSyntax, loop: Syntax, function: Syntax?) -> ExprSyntax {
            var current = ConcurrencyNoCheckThenWrite.withoutParentheses(expression)
            for _ in 0..<ConcurrencyNoCheckThenWrite.depthLimit {
                guard let reference = current.as(DeclReferenceExprSyntax.self),
                    let declared = declaration(of: reference), let initializer = declared.initializer,
                    declared.node.positionAfterSkippingLeadingTrivia >= loop.endPositionBeforeTrailingTrivia,
                    ConcurrencyNoCheckThenWrite.enclosingFunction(declared.node)?.id == function?.id
                else {
                    return current
                }
                current = ConcurrencyNoCheckThenWrite.withoutParentheses(initializer)
            }
            return current
        }

        /* Two path expressions compared as the TypeScript rule compares them, with Foundation's path builders in place of `path.join`. Every changeable binding read is collected for the stillness check. */
        func same(
            _ left: ExprSyntax,
            _ right: ExprSyntax,
            loop: Syntax,
            function: Syntax?,
            bindings: inout [Declaration],
            depth: Int,
        ) -> Bool {
            guard depth <= ConcurrencyNoCheckThenWrite.depthLimit else { return false }
            let left = expand(left, loop: loop, function: function)
            let right = expand(right, loop: loop, function: function)
            if let leftReference = left.as(DeclReferenceExprSyntax.self),
                let rightReference = right.as(DeclReferenceExprSyntax.self)
            {
                guard let leftDeclared = declaration(of: leftReference),
                    let rightDeclared = declaration(of: rightReference), leftDeclared.id == rightDeclared.id
                else { return false }
                if leftDeclared.isVariable {
                    bindings.append(leftDeclared)
                }
                return true
            }
            if let leftLiteral = left.as(StringLiteralExprSyntax.self),
                let rightLiteral = right.as(StringLiteralExprSyntax.self)
            {
                guard leftLiteral.openingPounds?.text == rightLiteral.openingPounds?.text,
                    leftLiteral.openingQuote.text == rightLiteral.openingQuote.text,
                    leftLiteral.segments.count == rightLiteral.segments.count
                else {
                    return false
                }
                for (leftSegment, rightSegment) in zip(leftLiteral.segments, rightLiteral.segments) {
                    switch (leftSegment, rightSegment) {
                        case (.stringSegment(let leftText), .stringSegment(let rightText)):
                            guard leftText.content.text == rightText.content.text else { return false }
                        case (.expressionSegment(let leftInterpolation), .expressionSegment(let rightInterpolation)):
                            guard leftInterpolation.pounds?.text == rightInterpolation.pounds?.text,
                                leftInterpolation.expressions.count == 1, rightInterpolation.expressions.count == 1,
                                let leftOnly = leftInterpolation.expressions.first,
                                let rightOnly = rightInterpolation.expressions.first, leftOnly.label == nil,
                                rightOnly.label == nil,
                                same(
                                    leftOnly.expression,
                                    rightOnly.expression,
                                    loop: loop,
                                    function: function,
                                    bindings: &bindings,
                                    depth: depth + 1,
                                )
                            else {
                                return false
                            }
                        default:
                            return false
                    }
                }
                return true
            }
            if let leftInteger = left.as(IntegerLiteralExprSyntax.self),
                let rightInteger = right.as(IntegerLiteralExprSyntax.self)
            {
                return leftInteger.literal.text == rightInteger.literal.text
            }
            if let leftSum = left.as(InfixOperatorExprSyntax.self),
                let rightSum = right.as(InfixOperatorExprSyntax.self)
            {
                guard leftSum.operator.as(BinaryOperatorExprSyntax.self)?.operator.text == "+",
                    rightSum.operator.as(BinaryOperatorExprSyntax.self)?.operator.text == "+"
                else { return false }
                return same(
                    leftSum.leftOperand,
                    rightSum.leftOperand,
                    loop: loop,
                    function: function,
                    bindings: &bindings,
                    depth: depth + 1,
                )
                    && same(
                        leftSum.rightOperand,
                        rightSum.rightOperand,
                        loop: loop,
                        function: function,
                        bindings: &bindings,
                        depth: depth + 1,
                    )
            }
            if let leftAccess = left.as(MemberAccessExprSyntax.self),
                let rightAccess = right.as(MemberAccessExprSyntax.self)
            {
                guard isUrlPath(leftAccess), isUrlPath(rightAccess), let leftBase = leftAccess.base,
                    let rightBase = rightAccess.base
                else { return false }
                return same(leftBase, rightBase, loop: loop, function: function, bindings: &bindings, depth: depth + 1)
            }
            if let leftCall = left.as(FunctionCallExprSyntax.self),
                let rightCall = right.as(FunctionCallExprSyntax.self)
            {
                guard leftCall.trailingClosure == nil, rightCall.trailingClosure == nil,
                    leftCall.additionalTrailingClosures.isEmpty, rightCall.additionalTrailingClosures.isEmpty,
                    let leftMember = leftCall.calledExpression.as(MemberAccessExprSyntax.self),
                    let rightMember = rightCall.calledExpression.as(MemberAccessExprSyntax.self),
                    let leftBase = leftMember.base, let rightBase = rightMember.base,
                    let builder = resolved(leftMember.declName.baseName),
                    ConcurrencyNoCheckThenWrite.pathBuilders.contains(builder),
                    resolved(rightMember.declName.baseName) == builder,
                    leftCall.arguments.count == rightCall.arguments.count,
                    same(leftBase, rightBase, loop: loop, function: function, bindings: &bindings, depth: depth + 1)
                else {
                    return false
                }
                for (leftArgument, rightArgument) in zip(leftCall.arguments, rightCall.arguments) {
                    guard leftArgument.label?.text == rightArgument.label?.text,
                        same(
                            leftArgument.expression,
                            rightArgument.expression,
                            loop: loop,
                            function: function,
                            bindings: &bindings,
                            depth: depth + 1,
                        )
                    else {
                        return false
                    }
                }
                return true
            }
            return false
        }

        /* `url.path`, read as a property and resolved to Foundation's `URL.path`. */
        func isUrlPath(_ access: MemberAccessExprSyntax) -> Bool {
            access.base != nil && access.declName.baseName.text == "path" && access.declName.argumentNames == nil
                && resolved(access.declName.baseName) == ConcurrencyNoCheckThenWrite.urlPath
        }

        // MARK: Stillness

        /*
         Whether a changeable binding keeps the value the check saw until the write runs: every write to it is in the
         loop's own function, none from a nested one, and none between the end of the loop and the write.
         */
        func holdsStill(
            _ declared: Declaration,
            loop: Syntax,
            function: Syntax?,
            write: FunctionCallExprSyntax,
        ) -> Bool {
            let scope = ConcurrencyNoCheckThenWrite.enclosingFunction(declared.node) ?? declared.node.root
            for token in scope.tokens(viewMode: .sourceAccurate) where token.text == declared.name {
                guard let reference = token.parent?.as(DeclReferenceExprSyntax.self), reference.baseName.id == token.id,
                    declaration(of: reference)?.id == declared.id,
                    writes(reference)
                else {
                    continue
                }
                if ConcurrencyNoCheckThenWrite.enclosingFunction(Syntax(reference))?.id != function?.id {
                    return false
                }
                let position = reference.positionAfterSkippingLeadingTrivia
                if position >= loop.endPositionBeforeTrailingTrivia
                    && position < write.positionAfterSkippingLeadingTrivia
                {
                    return false
                }
            }
            return true
        }

        /* Whether a name is written where it stands: assigned, passed `inout`, or the receiver of a method that may be `mutating`, directly or through its members. */
        func writes(_ reference: DeclReferenceExprSyntax) -> Bool {
            var current = Syntax(reference)
            while let parent = current.parent {
                if let access = parent.as(MemberAccessExprSyntax.self), access.base?.id == current.id {
                    if let call = access.parent?.as(FunctionCallExprSyntax.self), call.calledExpression.id == access.id,
                        !(resolved(access.declName.baseName).map(ConcurrencyNoCheckThenWrite.pathBuilders.contains)
                            ?? false)
                    {
                        return true
                    }
                    current = parent
                }
                else if let subscriptCall = parent.as(SubscriptCallExprSyntax.self),
                    subscriptCall.calledExpression.id == current.id
                {
                    current = parent
                }
                else if parent.is(OptionalChainingExprSyntax.self) || parent.is(ForceUnwrapExprSyntax.self)
                    || parent.is(TupleExprSyntax.self)
                    || parent.is(LabeledExprSyntax.self) && parent.parent?.parent?.is(TupleExprSyntax.self) == true
                {
                    current = parent
                }
                else {
                    break
                }
            }
            if current.parent?.is(InOutExprSyntax.self) == true {
                return true
            }
            guard let infix = current.parent?.as(InfixOperatorExprSyntax.self), infix.leftOperand.id == current.id
            else { return false }
            if infix.operator.is(AssignmentExprSyntax.self) {
                return true
            }
            guard let operation = infix.operator.as(BinaryOperatorExprSyntax.self)?.operator.text else { return false }
            return operation.hasSuffix("=") && !ConcurrencyNoCheckThenWrite.comparisons.contains(operation)
        }

        // MARK: Names

        /* The local binding a bare name resolves to by Swift's scope rules, or nil when it is not a local the rule can read (a property, a global, a local function). */
        func declaration(of reference: DeclReferenceExprSyntax) -> Declaration? {
            guard reference.argumentNames == nil, case .identifier = reference.baseName.tokenKind else { return nil }
            let name = reference.baseName.text
            var child = Syntax(reference)
            while let node = child.parent {
                if node.is(MemberBlockItemListSyntax.self) || node.is(SourceFileSyntax.self) {
                    return nil
                }
                switch ConcurrencyNoCheckThenWrite.lookup(name, in: node, below: child) {
                    case .declared(let declared)?:
                        return declared
                    case .unknown?:
                        return nil
                    case nil:
                        break
                }
                child = node
            }
            return nil
        }
    }

    /* What one scope binds a name to, seen from a child of it: the statements before it, the conditions before it, a loop's or a clause's pattern, a closure's or a function's parameters. */
    static func lookup(_ name: String, in node: Syntax, below child: Syntax) -> Lookup? {
        if let list = node.as(CodeBlockItemListSyntax.self) {
            /* A file-scope binding is a global, which any other file can change. */
            if list.parent?.is(SourceFileSyntax.self) == true {
                return .unknown
            }
            let statements = Array(list)
            guard let index = statements.firstIndex(where: { $0.id == child.id }) else { return nil }
            for statement in statements[..<index].reversed() {
                if let found = lookup(name, declaredBy: Syntax(statement.item)) {
                    return found
                }
            }
            return nil
        }
        if let conditions = node.as(ConditionElementListSyntax.self) {
            let all = Array(conditions)
            guard let index = all.firstIndex(where: { $0.id == child.id }) else { return nil }
            return lookup(name, in: Array(all[..<index]))
        }
        if let ifExpression = node.as(IfExprSyntax.self), ifExpression.body.id == child.id {
            return lookup(name, in: Array(ifExpression.conditions))
        }
        if let whileStatement = node.as(WhileStmtSyntax.self), whileStatement.body.id == child.id {
            return lookup(name, in: Array(whileStatement.conditions))
        }
        if let forStatement = node.as(ForStmtSyntax.self),
            forStatement.body.id == child.id || forStatement.whereClause?.id == child.id
        {
            return patternBinding(
                name,
                in: Syntax(forStatement.pattern),
                isVariable: forStatement.pattern.tokens(viewMode: .sourceAccurate).contains {
                    $0.tokenKind == .keyword(.var)
                },
            )
        }
        if let switchCase = node.as(SwitchCaseSyntax.self) {
            return patternBinding(
                name,
                in: Syntax(switchCase.label),
                isVariable: switchCase.label.tokens(viewMode: .sourceAccurate).contains {
                    $0.tokenKind == .keyword(.var)
                },
            )
        }
        if let catchClause = node.as(CatchClauseSyntax.self), catchClause.body.id == child.id {
            if catchClause.catchItems.isEmpty {
                return name == "error"
                    ? .declared(
                        Declaration(
                            id: catchClause.id,
                            name: name,
                            isVariable: false,
                            initializer: nil,
                            node: Syntax(catchClause),
                        )
                    ) : nil
            }
            return patternBinding(
                name,
                in: Syntax(catchClause.catchItems),
                isVariable: catchClause.catchItems.tokens(viewMode: .sourceAccurate).contains {
                    $0.tokenKind == .keyword(.var)
                },
            )
        }
        if let closure = node.as(ClosureExprSyntax.self), let signature = closure.signature {
            if let capture = signature.capture?.items.first(where: { $0.name.text == name }) {
                return .declared(
                    Declaration(
                        id: capture.name.id,
                        name: name,
                        isVariable: false,
                        initializer: nil,
                        node: Syntax(capture.name),
                    )
                )
            }
            switch signature.parameterClause {
                case .simpleInput(let parameters)?:
                    guard let parameter = parameters.first(where: { $0.name.text == name }) else { return nil }
                    /* Untyped, so it may be `inout`: its writes are checked. */
                    return .declared(
                        Declaration(
                            id: parameter.name.id,
                            name: name,
                            isVariable: true,
                            initializer: nil,
                            node: Syntax(parameter.name),
                        )
                    )
                case .parameterClause(let clause)?:
                    guard
                        let parameter = clause.parameters.first(where: { ($0.secondName ?? $0.firstName).text == name })
                    else { return nil }
                    let token = parameter.secondName ?? parameter.firstName
                    return .declared(
                        Declaration(
                            id: token.id,
                            name: name,
                            isVariable: parameter.type.map(isInout) ?? true,
                            initializer: nil,
                            node: Syntax(token),
                        )
                    )
                case nil:
                    return nil
            }
        }
        if let accessor = node.as(AccessorDeclSyntax.self) {
            if let parameter = accessor.parameters?.name, parameter.text == name {
                return .declared(
                    Declaration(
                        id: parameter.id,
                        name: name,
                        isVariable: false,
                        initializer: nil,
                        node: Syntax(parameter),
                    )
                )
            }
            return name == "newValue" || name == "oldValue"
                ? .declared(
                    Declaration(
                        id: accessor.id,
                        name: name,
                        isVariable: false,
                        initializer: nil,
                        node: Syntax(accessor),
                    )
                ) : nil
        }
        let parameters: FunctionParameterListSyntax?
        if let function = node.as(FunctionDeclSyntax.self) {
            parameters = function.signature.parameterClause.parameters
        }
        else if let initializer = node.as(InitializerDeclSyntax.self) {
            parameters = initializer.signature.parameterClause.parameters
        }
        else if let subscriptDeclaration = node.as(SubscriptDeclSyntax.self) {
            parameters = subscriptDeclaration.parameterClause.parameters
        }
        else {
            parameters = nil
        }
        guard let parameter = parameters?.first(where: { ($0.secondName ?? $0.firstName).text == name }) else {
            return nil
        }
        let token = parameter.secondName ?? parameter.firstName
        return .declared(
            Declaration(
                id: token.id,
                name: name,
                isVariable: isInout(parameter.type),
                initializer: nil,
                node: Syntax(token),
            )
        )
    }

    /* What a statement before the name declares it as: a `let` or `var`, a `guard` binding, or something the rule does not read. */
    static func lookup(_ name: String, declaredBy statement: Syntax) -> Lookup? {
        if let variable = statement.as(VariableDeclSyntax.self) {
            let isVariable = variable.bindingSpecifier.tokenKind == .keyword(.var)
            for binding in variable.bindings {
                guard let token = boundToken(name, in: Syntax(binding.pattern)) else { continue }
                /* A computed local or a wrapped one is evaluated on every read, so it holds no one value. */
                guard variable.attributes.isEmpty, binding.accessorBlock == nil else { return .unknown }
                let initializer =
                    !isVariable && binding.pattern.is(IdentifierPatternSyntax.self) ? binding.initializer?.value : nil
                return .declared(
                    Declaration(
                        id: token.id,
                        name: name,
                        isVariable: isVariable,
                        initializer: initializer,
                        node: Syntax(token),
                    )
                )
            }
            return nil
        }
        if let guardStatement = statement.as(GuardStmtSyntax.self) {
            return lookup(name, in: Array(guardStatement.conditions))
        }
        if let function = statement.as(FunctionDeclSyntax.self) {
            return function.name.text == name ? .unknown : nil
        }
        if let type = statement.asProtocol((any NamedDeclSyntax).self) {
            return type.name.text == name ? .unknown : nil
        }
        if statement.is(IfConfigDeclSyntax.self) {
            return statement.tokens(viewMode: .sourceAccurate).contains { $0.text == name } ? .unknown : nil
        }
        return nil
    }

    /* What a list of conditions binds a name to: the last `if let`, `guard var` or `case let` naming it. */
    static func lookup(_ name: String, in conditions: [ConditionElementSyntax]) -> Lookup? {
        for condition in conditions.reversed() {
            if let optional = condition.condition.as(OptionalBindingConditionSyntax.self),
                let token = boundToken(name, in: Syntax(optional.pattern))
            {
                return .declared(
                    Declaration(
                        id: token.id,
                        name: name,
                        isVariable: optional.bindingSpecifier.tokenKind == .keyword(.var),
                        initializer: nil,
                        node: Syntax(token),
                    )
                )
            }
            if let matching = condition.condition.as(MatchingPatternConditionSyntax.self),
                boundToken(name, in: Syntax(matching.pattern)) != nil
            {
                return patternBinding(
                    name,
                    in: Syntax(matching.pattern),
                    isVariable: matching.pattern.tokens(viewMode: .sourceAccurate).contains {
                        $0.tokenKind == .keyword(.var)
                    },
                )
            }
        }
        return nil
    }

    /* A name bound by a pattern, constant unless the pattern binds with `var`. */
    static func patternBinding(_ name: String, in pattern: Syntax, isVariable: Bool) -> Lookup? {
        guard let token = boundToken(name, in: pattern) else { return nil }
        return .declared(
            Declaration(id: token.id, name: name, isVariable: isVariable, initializer: nil, node: Syntax(token))
        )
    }

    /* The identifier a pattern binds under this name, at any depth of a tuple or enum pattern. */
    static func boundToken(_ name: String, in pattern: Syntax) -> TokenSyntax? {
        if let identifier = pattern.as(IdentifierPatternSyntax.self) {
            return identifier.identifier.text == name ? identifier.identifier : nil
        }
        for child in pattern.children(viewMode: .sourceAccurate) {
            if child.is(ExprSyntax.self) && !child.is(PatternExprSyntax.self) {
                /* An enum case's associated value pattern sits in a call's arguments; anything else in an expression is matched, not bound. */
                if let call = child.as(FunctionCallExprSyntax.self) {
                    for argument in call.arguments {
                        if let found = boundToken(name, in: Syntax(argument.expression)) {
                            return found
                        }
                    }
                }
                continue
            }
            if let found = boundToken(name, in: child) {
                return found
            }
        }
        return nil
    }

    static func isInout(_ type: TypeSyntax) -> Bool {
        type.tokens(viewMode: .sourceAccurate).contains { $0.tokenKind == .keyword(.inout) }
    }

    // MARK: Structure

    /* A function, closure, accessor or initializer: code that runs on its own schedule relative to the code around it. */
    static func isFunctionLike(_ node: Syntax) -> Bool {
        node.is(FunctionDeclSyntax.self) || node.is(InitializerDeclSyntax.self) || node.is(DeinitializerDeclSyntax.self)
            || node.is(AccessorDeclSyntax.self)
            || node.is(AccessorBlockSyntax.self) || node.is(ClosureExprSyntax.self) || node.is(SubscriptDeclSyntax.self)
    }

    static func isTypeDeclaration(_ node: Syntax) -> Bool {
        node.is(StructDeclSyntax.self) || node.is(ClassDeclSyntax.self) || node.is(EnumDeclSyntax.self)
            || node.is(ActorDeclSyntax.self) || node.is(ProtocolDeclSyntax.self)
            || node.is(ExtensionDeclSyntax.self)
    }

    static func enclosingFunction(_ node: Syntax) -> Syntax? {
        var current = node.parent
        while let candidate = current {
            if isFunctionLike(candidate) {
                return candidate
            }
            current = candidate.parent
        }
        return nil
    }

    /* Visits a subtree in source order without entering closures, nested functions or local types. */
    static func forEachOwnNode(_ root: Syntax, _ visit: (Syntax) -> Void) {
        visit(root)
        for child in root.children(viewMode: .sourceAccurate) where !isFunctionLike(child) && !isTypeDeclaration(child)
        {
            forEachOwnNode(child, visit)
        }
    }

    /*
     The statements that run after the loop in the same function: climbing out through blocks, branches, `do` and
     `switch` cases, and stopping at an enclosing loop, a `guard`'s `else`, a `defer`, a function or a type.
     */
    static func followingStatements(_ loop: Syntax) -> [CodeBlockItemSyntax] {
        var following: [CodeBlockItemSyntax] = []
        var current = loop
        while let parent = current.parent {
            if isFunctionLike(parent) || isTypeDeclaration(parent) || parent.is(WhileStmtSyntax.self)
                || parent.is(RepeatStmtSyntax.self) || parent.is(ForStmtSyntax.self)
                || parent.is(GuardStmtSyntax.self) || parent.is(DeferStmtSyntax.self)
            {
                break
            }
            if let list = parent.as(CodeBlockItemListSyntax.self) {
                let items = Array(list)
                if let index = items.firstIndex(where: { $0.id == current.id }) {
                    following += items[(index + 1)...]
                }
            }
            current = parent
        }
        return following
    }

    /* Parentheses, as one unlabeled element in a tuple, removed. */
    static func withoutParentheses(_ expression: ExprSyntax) -> ExprSyntax {
        var current = expression
        while let tuple = current.as(TupleExprSyntax.self), tuple.elements.count == 1, let only = tuple.elements.first,
            only.label == nil
        {
            current = only.expression
        }
        return current
    }

    /* Whether a receiver chain passes through `?.`, so the call may not run. */
    static func isOptionallyChained(_ expression: ExprSyntax) -> Bool {
        Syntax(expression).tokens(viewMode: .sourceAccurate).contains { $0.tokenKind == .postfixQuestionMark }
    }

    /* Options the source shows replace a file there: `.atomic`, `[]`, or a list of `.atomic`. Anything else may hold `.withoutOverwriting`. */
    static func optionsReplace(_ expression: ExprSyntax) -> Bool {
        let expression = withoutParentheses(expression)
        let isAtomic = { (element: ExprSyntax) -> Bool in
            guard let member = withoutParentheses(element).as(MemberAccessExprSyntax.self) else { return false }
            return member.base == nil && member.declName.baseName.text == "atomic"
                && member.declName.argumentNames == nil
        }
        if let array = expression.as(ArrayExprSyntax.self) {
            return array.elements.allSatisfy { isAtomic($0.expression) }
        }
        return isAtomic(expression)
    }
}
