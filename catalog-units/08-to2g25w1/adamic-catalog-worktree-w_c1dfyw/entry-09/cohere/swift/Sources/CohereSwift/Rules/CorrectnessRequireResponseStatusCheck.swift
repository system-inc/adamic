import SwiftOperators
import SwiftSyntax

/*
 No trusting what URLSession returned without asking whether the request succeeded:

     let (data, _) = try await session.data(for: request)
     return try JSONDecoder().decode(Profile.self, from: data)

 URLSession throws only when no response arrives at all. A 401, a 404 or a 500 is a response like any other, so
 its body (an error page, a JSON error envelope) comes back as the data, and code that decodes it, returns it or
 caches it treats the failure as the answer: an error object decoded into a model that happens to accept it, an
 image cache storing a 404 page, a POST reported as saved when the server refused it. The repair casts the
 response to `HTTPURLResponse` and reads `statusCode` before the body is trusted, and throws or returns the
 failure when it is not a success.

 It ports `nexus/correctness-require-response-status-check`, whose definitions it keeps one for one:
 - Which calls are tracked. The TypeScript rule tracks the platform's `fetch` and Structure's
   `networkService.request`, decided by the declaration the call resolves to. Here it is URLSession, decided the
   same way, through the index the build wrote: the async methods Foundation's Swift extension declares on
   `NSURLSession` (`data(for:)`, `data(from:)`, `upload(for:from:)`, `upload(for:fromFile:)`, `bytes(for:)`,
   `bytes(from:)`, `download(for:)`, `download(from:)`, `download(resumeFrom:)`, each with its `delegate:`, matched
   by the symbol up to the declaration's full name), and the Objective-C methods that take a completion handler
   (`dataTask`, `uploadTask` and `downloadTask` with `completionHandler:`). A method of ours with the same name
   resolves to ours and is never tracked, whatever its receiver is called.
 - What the TypeScript rule calls the Response is here a pair. `let (data, response) = try await <tracked call>`,
   a `let` whose pattern is a two-element tuple of names or `_`, plays the part of the TypeScript `const`, and the
   completion handler's first two parameters (named, or `$0` and `$1`) play the part of the `.then` callback's
   one parameter. Swift hands the body over already read, so where the TypeScript rule waits for `.json()` or
   `.text()`, here any use of the first element (the `Data`, the `AsyncBytes`, the downloaded file's `URL`) is the
   body read: decoding it, returning it, passing it on, iterating it, reading `.count`. Two uses are not, because
   they read none of it: binding it in a condition (`guard let data`, `if let data = data`), which makes the new
   name the body, and comparing it with `nil`. The completion handler's body arrives optional, and asking whether
   it arrived is not trusting it.
 - What counts as checking is the second element, read as the TypeScript rule reads the Response. A read of
   `statusCode` is a check, wherever it is (in a condition, a log line, a thrown error), because reading more as a
   check can only silence a site. The status lives on `HTTPURLResponse`, so a cast of the response to it
   (`as?`, `as!`, through parentheses and optional chaining) is followed: `(response as? HTTPURLResponse)?.statusCode`
   is a check, and a cast bound to a name (`if let http = response as? HTTPURLResponse`, `guard let`, a plain
   `let`) makes that name the response. `url`, `mimeType`, `expectedContentLength`, `textEncodingName`,
   `suggestedFilename`, `allHeaderFields` and a called `value(forHTTPHeaderField:)` are neutral, the counterpart of
   `.headers` and `.url`. Everything else is an escape and counts as a check, as it does there: passing the
   response to a function (`try validate(response)`), returning it, storing it, a cast to any other type, an `is`
   test, a comparison. The receiver may read the status, and proving that it does would mean following it, so the
   rule assumes it does.
 - On a path, through the enclosing function's control flow, exactly as the TypeScript rule decides it. A body read
   is reported when some path leaves the declaration, reaches the read, and goes on to a normal exit (falling off
   the end, or a `return`) or to the declaration again on the next turn of a loop, without a check or an escape
   anywhere along it. So checking after reading is checking (`let model = try decode(data)` then a `guard` on the
   status, then `return model`); a path that leaves by `throw` is not an exit, so reading the body to put it in a
   thrown error is not trusting it; and a check on one branch does not cover the other. The graph is laid out the
   way the TypeScript rule's is: the first `try` inside a `do` forks to its `catch` clauses (later ones reuse that
   fork, ESLint's approximation), a `throw` goes to the innermost `do` with `catch` clauses or leaves, a `guard`'s
   `else` never falls through, `fatalError` and `preconditionFailure` end their path, and a condition list fails
   at each element. `&&`, `||`, `??`, the ternary and optional chaining are laid out in source order rather than
   forked around the operand they may skip. That puts every event of the operand on the path, which can only add
   checks to it, so it can only silence a site; the misses it costs are listed below.
 - One rule Swift needs and TypeScript does not: the branch where a cast of the response to `HTTPURLResponse`
   bound in a condition fails counts as checked. URLSession answers every http and https request with an
   `HTTPURLResponse`, so that branch is a non-HTTP URL (a `file:` or `data:` one), which has no status to read.
   Without it, `if let http = response as? HTTPURLResponse, !(200..<300).contains(http.statusCode) { throw ... }`,
   the check AhraOs writes before its event stream, would read as unchecked on a branch that never runs.
 - The two other shapes it reports, as the TypeScript rule does. Discarded: `_ = try await session.data(for:)`, the
   call as a statement of its own, `let _ =` or `let (_, _) =`, a tracked pair neither of whose names is ever
   referenced, and a completion handler that names neither its body nor its response. Each throws the status away
   with the response. Read inline: `try await session.data(for: request).0` takes the body of a response nothing
   else can see.

 No fix: what a failure should do (throw, return nothing, retry, show the server's message) is the author's call.

 What it accepts as safe, and why: a status read on every path to the use (before it, or after it and before the
 function returns); the response handed to a helper, returned or stored (the escape above); `.1` read inline,
 which reads no body; a body read only on paths that throw; any call that is not URLSession's (a client of ours
 that already throws on a bad status is out by design, as `R2Api.signedS3Request` is for the TypeScript rule: a
 wrapper is a missed finding, never a false one). Every real site in AhraOs checks first and stays silent:
 `ProfileImageCache.swift:169` and `AhraOsHttpClient.swift:303` behind a `guard` or `if` on the cast and its
 status, and `AhraOsHttpClient.swift:455`, whose `bytes(for:)` stream is read inside a `Task`. ProviderProxy's
 `Task { try await session.bytes(for: urlRequest) }` hands the pair back through the task, so it is not tracked.

 Known misses, every one a finding not made and never one invented:
 - A pair any closure or `defer` refers to is dropped whole, as the TypeScript rule drops a binding a closure sees:
   the closure may run before the use or after it, and the graph of the enclosing function cannot say which. So
   AhraOs's event stream, consumed in a `Task`, would not be reported even without its check. Code the layout does
   not model (a statement inside `#if`, a switch with `#if` cases) is read the same way, and drops the pair too.
 - A pair bound any other way: `var (data, response)` (the TypeScript rule follows only `const`), a single name
   `let result = try await session.data(for:)` read as `result.0`, `async let`, a labeled tuple pattern, a pair
   laundered through a `Task`, a `TaskGroup` or a helper that returns it, and a completion handler that is not a
   closure literal (a method reference).
 - A producer of ours wrapping URLSession, Combine's `dataTaskPublisher`, the delegate-based tasks without a
   completion handler, and `URLSessionWebSocketTask`.
 - The call alone as the only statement of a body (`Task { try await session.data(for: request) }`), which Swift
   may read as the body's implicit return, so it is not counted as discarded.
 - The operand-order layout above: `if flag || http.statusCode == 200 { return data }` returns the body unchecked
   when `flag` is true, and reads as checked. So does a check inside a `case` pattern of an earlier `case` item.
 - Only the first `try` in a `do` forks to `catch`, so a decode that throws after an unchecked read and lands in a
   `catch` that returns normally is not reported. The TypeScript rule's ESLint layout misses it the same way, and
   on the paths it does not see the error body is never kept, only decoded and thrown away.
 - A reference the walk never reached drops its pair rather than being judged. That guard, like the TypeScript
   rule's, has no fixture that reaches it: every statement the layout does not model is walked as code it cannot
   place, which drops the pair already. It stays because an unplaced check is exactly what would make a silent site
   a false one.
 */
public struct CorrectnessRequireResponseStatusCheck: TypedFileRule {
    public let name = "cohere-swift/correctness-require-response-status-check"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /*
     The async methods Foundation's Swift extension declares on `NSURLSession`, by their symbol up to the declaration's
     full name. What follows is the signature's mangling, which never starts with a digit; a longer name (one more
     label) would, so a match must not be followed by one.
     */
    static let asyncProducers: [String] = [
        "s:So12NSURLSessionC10FoundationE4data3for8delegate",
        "s:So12NSURLSessionC10FoundationE4data4from8delegate",
        "s:So12NSURLSessionC10FoundationE6upload3for4from8delegate",
        "s:So12NSURLSessionC10FoundationE6upload3for8fromFile8delegate",
        "s:So12NSURLSessionC10FoundationE5bytes3for8delegate",
        "s:So12NSURLSessionC10FoundationE5bytes4from8delegate",
        "s:So12NSURLSessionC10FoundationE8download3for8delegate",
        "s:So12NSURLSessionC10FoundationE8download4from8delegate",
        "s:So12NSURLSessionC10FoundationE8download10resumeFrom8delegate",
    ]

    /* The Objective-C methods that hand body, response and error to a completion handler. */
    static let completionProducers: Set<String> = [
        "c:objc(cs)NSURLSession(im)dataTaskWithRequest:completionHandler:",
        "c:objc(cs)NSURLSession(im)dataTaskWithURL:completionHandler:",
        "c:objc(cs)NSURLSession(im)uploadTaskWithRequest:fromData:completionHandler:",
        "c:objc(cs)NSURLSession(im)uploadTaskWithRequest:fromFile:completionHandler:",
        "c:objc(cs)NSURLSession(im)downloadTaskWithRequest:completionHandler:",
        "c:objc(cs)NSURLSession(im)downloadTaskWithURL:completionHandler:",
        "c:objc(cs)NSURLSession(im)downloadTaskWithResumeData:completionHandler:",
    ]

    static let httpResponse = "c:objc(cs)NSHTTPURLResponse"

    /* The members of a response that neither read the status nor hand the response anywhere: the counterpart of `.headers` and `.url`. */
    static let neutralProperties: Set<String> = [
        "url", "mimeType", "expectedContentLength", "textEncodingName", "suggestedFilename", "allHeaderFields",
    ]

    /* Every tracked method's name holds one of these, and a call names its method. */
    public func applies(to file: ParsedFile) -> Bool {
        file.source.contains("data") || file.source.contains("upload") || file.source.contains("bytes")
            || file.source.contains("download")
    }

    public func findings(in file: ParsedFile, symbols: FileSymbols) -> [FindingRecord] {
        /* Folded, so `response as? HTTPURLResponse` is one cast and `_ = call` one assignment. Folding moves no token. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let producers = Producers(file: file, symbols: symbols)
        let finder = Finder(viewMode: .sourceAccurate)
        finder.walk(folded)

        var reports: [Report] = []
        var producerRoots: Set<SyntaxIdentifier> = []
        var handlers: [SyntaxIdentifier: FunctionCallExprSyntax] = [:]
        for call in finder.calls {
            switch producers.kind(of: call) {
                case .async:
                    if let report = Self.loneReport(call) {
                        reports.append(report)
                    }
                    if let root = Self.enclosingRoot(of: Syntax(call)) {
                        producerRoots.insert(root)
                    }
                case .completion:
                    if let handler = Self.handler(of: call) {
                        handlers[handler.id] = call
                    }
                case nil:
                    break
            }
        }
        for root in finder.roots {
            let call = handlers[root.owner]
            guard call != nil || producerRoots.contains(root.owner) else { continue }
            let builder = Builder(producers: producers)
            reports += builder.analyze(root, completion: call)
        }

        var seen: Set<SyntaxIdentifier> = []
        return
            reports
            .filter { seen.insert($0.node.id).inserted }
            .sorted { $0.node.positionAfterSkippingLeadingTrivia < $1.node.positionAfterSkippingLeadingTrivia }
            .map { report in
                file.finding(
                    at: report.node,
                    rule: name,
                    message: report.discarded
                        ? RuleMessages.CorrectnessRequireResponseStatusCheck.responseDiscarded()
                        : RuleMessages.CorrectnessRequireResponseStatusCheck.bodyReadWithoutStatusCheck(),
                )
            }
    }

    /* One finding before it is placed: the node it covers, and whether it is a discard or a body read. */
    struct Report {
        let node: Syntax
        let discarded: Bool
    }

    /* A body of code the graph is built over, and the declaration or closure that owns it. */
    struct Root {
        let items: CodeBlockItemListSyntax
        let owner: SyntaxIdentifier
    }

    /* Every root, and every call (each one may be a tracked one). */
    final class Finder: SyntaxVisitor {
        private(set) var roots: [Root] = []
        private(set) var calls: [FunctionCallExprSyntax] = []

        override func visit(_ node: SourceFileSyntax) -> SyntaxVisitorContinueKind {
            roots.append(Root(items: node.statements, owner: node.id))
            return .visitChildren
        }

        override func visit(_ node: FunctionDeclSyntax) -> SyntaxVisitorContinueKind {
            if let body = node.body {
                roots.append(Root(items: body.statements, owner: node.id))
            }
            return .visitChildren
        }

        override func visit(_ node: InitializerDeclSyntax) -> SyntaxVisitorContinueKind {
            if let body = node.body {
                roots.append(Root(items: body.statements, owner: node.id))
            }
            return .visitChildren
        }

        override func visit(_ node: DeinitializerDeclSyntax) -> SyntaxVisitorContinueKind {
            if let body = node.body {
                roots.append(Root(items: body.statements, owner: node.id))
            }
            return .visitChildren
        }

        override func visit(_ node: AccessorDeclSyntax) -> SyntaxVisitorContinueKind {
            if let body = node.body {
                roots.append(Root(items: body.statements, owner: node.id))
            }
            return .visitChildren
        }

        override func visit(_ node: AccessorBlockSyntax) -> SyntaxVisitorContinueKind {
            if case .getter(let items) = node.accessors {
                roots.append(Root(items: items, owner: node.id))
            }
            return .visitChildren
        }

        override func visit(_ node: ClosureExprSyntax) -> SyntaxVisitorContinueKind {
            roots.append(Root(items: node.statements, owner: node.id))
            return .visitChildren
        }

        override func visit(_ node: FunctionCallExprSyntax) -> SyntaxVisitorContinueKind {
            calls.append(node)
            return .visitChildren
        }
    }

    /* The root a node's code runs in: the nearest function, initializer, accessor, getter, closure or file around it. */
    static func enclosingRoot(of node: Syntax) -> SyntaxIdentifier? {
        var current = node.parent
        while let candidate = current {
            if candidate.is(ClosureExprSyntax.self) || candidate.is(FunctionDeclSyntax.self)
                || candidate.is(InitializerDeclSyntax.self)
                || candidate.is(DeinitializerDeclSyntax.self) || candidate.is(AccessorDeclSyntax.self)
                || candidate.is(SourceFileSyntax.self)
            {
                return candidate.id
            }
            if let block = candidate.as(AccessorBlockSyntax.self), case .getter = block.accessors {
                return candidate.id
            }
            current = candidate.parent
        }
        return nil
    }

    /* The closure a completion-handler call hands its result to: the trailing closure, or a closure literal labeled `completionHandler:`. */
    static func handler(of call: FunctionCallExprSyntax) -> ClosureExprSyntax? {
        if let trailing = call.trailingClosure {
            return trailing
        }
        return call.arguments.first { $0.label?.text == "completionHandler" }?.expression.as(ClosureExprSyntax.self)
    }

    /*
     What a tracked async call reports on its own, with no pair to follow: `.0` read inline, and the result thrown
     away by `_ =`, by `let _` or `let (_, _)`, or as a statement among others.
     */
    static func loneReport(_ call: FunctionCallExprSyntax) -> Report? {
        let node = outward(Syntax(call))
        guard let parent = node.parent else { return nil }
        if let access = parent.as(MemberAccessExprSyntax.self), access.base.map({ Syntax($0).id == node.id }) == true,
            access.declName.baseName.text == "0"
        {
            return Report(node: Syntax(access), discarded: false)
        }
        if let assignment = parent.as(InfixOperatorExprSyntax.self), Syntax(assignment.rightOperand).id == node.id,
            assignment.operator.is(AssignmentExprSyntax.self),
            assignment.leftOperand.is(DiscardAssignmentExprSyntax.self)
        {
            return Report(node: node, discarded: true)
        }
        if let item = parent.as(CodeBlockItemSyntax.self), let list = item.parent?.as(CodeBlockItemListSyntax.self),
            list.count > 1
        {
            return Report(node: node, discarded: true)
        }
        if let initializer = parent.as(InitializerClauseSyntax.self),
            let binding = initializer.parent?.as(PatternBindingSyntax.self),
            let declaration = binding.parent?.parent?.as(VariableDeclSyntax.self), !isDeferred(declaration)
        {
            if binding.pattern.is(WildcardPatternSyntax.self) {
                return Report(node: node, discarded: true)
            }
            if let tuple = binding.pattern.as(TuplePatternSyntax.self),
                tuple.elements.allSatisfy({ $0.pattern.is(WildcardPatternSyntax.self) })
            {
                return Report(node: node, discarded: true)
            }
        }
        return nil
    }

    /* `lazy` and `async let` run their initializer at another time than where they are written. */
    static func isDeferred(_ declaration: VariableDeclSyntax) -> Bool {
        declaration.modifiers.contains { $0.name.text == "lazy" || $0.name.text == "async" }
    }

    /* A name as it is meant, without the backticks that let a keyword be one. */
    static func plainName(_ token: TokenSyntax) -> String {
        String(token.text.filter { $0 != "`" })
    }

    /*
     Out from an expression through what does not change which value it is: parentheses, `try`, `await`, `!` and `?`.
     `stripped` goes the same way inward, so a reference climbed out of and an initializer stripped into always meet.
     */
    static func outward(_ node: Syntax) -> Syntax {
        var current = node
        while let parent = current.parent {
            if parent.is(TryExprSyntax.self) || parent.is(AwaitExprSyntax.self) || parent.is(ForceUnwrapExprSyntax.self)
                || parent.is(OptionalChainingExprSyntax.self)
            {
                current = parent
                continue
            }
            if let element = parent.as(LabeledExprSyntax.self), element.label == nil,
                let list = element.parent?.as(LabeledExprListSyntax.self), list.count == 1,
                let tuple = list.parent?.as(TupleExprSyntax.self)
            {
                current = Syntax(tuple)
                continue
            }
            break
        }
        return current
    }

    static func stripped(_ expression: ExprSyntax) -> ExprSyntax {
        var current = expression
        while true {
            if let wrapper = current.as(TryExprSyntax.self) {
                current = wrapper.expression
            }
            else if let wrapper = current.as(AwaitExprSyntax.self) {
                current = wrapper.expression
            }
            else if let wrapper = current.as(ForceUnwrapExprSyntax.self) {
                current = wrapper.expression
            }
            else if let wrapper = current.as(OptionalChainingExprSyntax.self) {
                current = wrapper.expression
            }
            else if let tuple = current.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                current = only.expression
            }
            else {
                return current
            }
        }
    }

    /* Which tracked call a call is, if any, by the declaration the index resolved its name to. */
    struct Producers {
        enum Kind {
            case async
            case completion
        }

        let file: ParsedFile
        let symbols: FileSymbols

        func resolved(_ token: TokenSyntax) -> String? {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            return symbols.reference(line: location.line, column: location.column)?.symbol
        }

        func kind(of call: FunctionCallExprSyntax) -> Kind? {
            let token: TokenSyntax
            if let member = call.calledExpression.as(MemberAccessExprSyntax.self) {
                token = member.declName.baseName
            }
            else if let reference = call.calledExpression.as(DeclReferenceExprSyntax.self) {
                token = reference.baseName
            }
            else {
                return nil
            }
            guard let symbol = resolved(token) else { return nil }
            if CorrectnessRequireResponseStatusCheck.completionProducers.contains(symbol) {
                return .completion
            }
            for prefix in CorrectnessRequireResponseStatusCheck.asyncProducers where symbol.hasPrefix(prefix) {
                let next = symbol.utf8.dropFirst(prefix.utf8.count).first
                if let next, next >= UInt8(ascii: "0"), next <= UInt8(ascii: "9") {
                    continue
                }
                return .async
            }
            return nil
        }

        /* `HTTPURLResponse` or `Foundation.HTTPURLResponse`, resolved to Foundation's class. */
        func isHttpResponse(_ type: TypeSyntax) -> Bool {
            let token: TokenSyntax
            if let identifier = type.as(IdentifierTypeSyntax.self), identifier.genericArgumentClause == nil {
                token = identifier.name
            }
            else if let member = type.as(MemberTypeSyntax.self),
                member.baseType.as(IdentifierTypeSyntax.self)?.name.text == "Foundation"
            {
                token = member.name
            }
            else {
                return false
            }
            return token.text == "HTTPURLResponse"
                && resolved(token) == CorrectnessRequireResponseStatusCheck.httpResponse
        }
    }

    /*
     The control-flow graph of one root, built in one walk that also resolves names, and the analysis over it. A port
     of the TypeScript rule's use of `control_flow_graph`: blocks of events in evaluation order, a declaration event
     where a pair is bound, a use event where one of its names is read, and the same two walks over the result.
     */
    final class Builder {
        enum Role {
            case body
            case response
        }

        enum Binding {
            case tracked(request: Int, role: Role)
            case other
        }

        enum Use {
            case declare
            case bodyRead
            case check
            case escape
        }

        struct Event {
            let request: Int
            let use: Use
            let report: Syntax?
        }

        final class Block {
            var events: [Event] = []
            var successors: [Int] = []
            var isFinal = false
        }

        /* One tracked pair: where it came from, the names it was bound to, and what the walk learned of it. */
        struct Request {
            let producer: Syntax
            let names: Set<String>
            var referenced = false
            var dropped = false
        }

        /* A `do` with `catch` clauses: where a throw inside it goes, and whether its first `try` has forked there yet. */
        final class CatchFrame {
            let dispatch: Int
            var forked = false

            init(dispatch: Int) {
                self.dispatch = dispatch
            }
        }

        /* Where `break` and `continue` go from inside a statement. */
        struct Jump {
            let label: String?
            let breakBlock: Int
            let continueBlock: Int?
            let takesUnlabeledBreak: Bool
        }

        let producers: Producers
        var blocks: [Block] = [Block()]
        var current = 0
        var scopes: [[String: Binding]] = [[:]]
        var requests: [Request] = []
        var catchFrames: [CatchFrame] = []
        var jumps: [Jump] = []
        var fallthroughTargets: [Int?] = []
        var visited: Set<SyntaxIdentifier> = []

        init(producers: Producers) {
            self.producers = producers
        }

        // MARK: Building

        func newBlock() -> Int {
            blocks.append(Block())
            return blocks.count - 1
        }

        func link(_ from: Int, _ to: Int) {
            blocks[from].successors.append(to)
        }

        /* After a jump: what follows is laid out in a block nothing leads to. */
        func startUnreachable() {
            current = newBlock()
        }

        func emit(_ request: Int, _ use: Use, report: Syntax? = nil) {
            blocks[current].events.append(Event(request: request, use: use, report: report))
        }

        func lookup(_ name: String) -> Binding? {
            for scope in scopes.reversed() {
                if let binding = scope[name] {
                    return binding
                }
            }
            return nil
        }

        func bind(_ name: String, _ binding: Binding) {
            scopes[scopes.count - 1][name] = binding
        }

        func track(producer: Syntax, body: String?, response: String?) {
            let request = requests.count
            requests.append(Request(producer: producer, names: Set([body, response].compactMap { $0 })))
            if let body {
                bind(body, .tracked(request: request, role: .body))
            }
            if let response {
                bind(response, .tracked(request: request, role: .response))
            }
            emit(request, .declare)
        }

        func lowerItems(_ items: CodeBlockItemListSyntax) {
            for statement in items {
                switch statement.item {
                    case .decl(let declaration):
                        if let variable = declaration.as(VariableDeclSyntax.self) {
                            lowerVariable(variable)
                        }
                        else {
                            opaque(Syntax(declaration))
                        }
                    case .stmt(let statement):
                        lowerStatement(statement, label: nil)
                    case .expr(let expression):
                        walk(Syntax(expression))
                }
            }
        }

        func lowerBlock(_ block: CodeBlockSyntax) {
            scopes.append([:])
            lowerItems(block.statements)
            scopes.removeLast()
        }

        func lowerStatement(_ statement: StmtSyntax, label: String?) {
            if let expression = statement.as(ExpressionStmtSyntax.self) {
                if let conditional = expression.expression.as(IfExprSyntax.self) {
                    lowerIf(conditional, label: label)
                }
                else if let switchExpression = expression.expression.as(SwitchExprSyntax.self) {
                    lowerSwitch(switchExpression, label: label)
                }
                else {
                    walk(Syntax(expression.expression))
                }
            }
            else if let labeled = statement.as(LabeledStmtSyntax.self) {
                lowerStatement(labeled.statement, label: labeled.label.text)
            }
            else if let returned = statement.as(ReturnStmtSyntax.self) {
                if let expression = returned.expression {
                    walk(Syntax(expression))
                }
                blocks[current].isFinal = true
                startUnreachable()
            }
            else if let thrown = statement.as(ThrowStmtSyntax.self) {
                walk(Syntax(thrown.expression))
                if let frame = catchFrames.last {
                    link(current, frame.dispatch)
                }
                startUnreachable()
            }
            else if let breakStatement = statement.as(BreakStmtSyntax.self) {
                let target =
                    breakStatement.label.map { label in jumps.last { $0.label == label.text } }
                    ?? jumps.last { $0.takesUnlabeledBreak }
                if let target {
                    link(current, target.breakBlock)
                }
                startUnreachable()
            }
            else if let continueStatement = statement.as(ContinueStmtSyntax.self) {
                let target =
                    continueStatement.label.map { label in jumps.last { $0.label == label.text } }
                    ?? jumps.last { $0.continueBlock != nil }
                if let block = target?.continueBlock {
                    link(current, block)
                }
                startUnreachable()
            }
            else if statement.is(FallThroughStmtSyntax.self) {
                if let target = fallthroughTargets.last, let target {
                    link(current, target)
                }
                startUnreachable()
            }
            else if let guardStatement = statement.as(GuardStmtSyntax.self) {
                lowerGuard(guardStatement)
            }
            else if let loop = statement.as(WhileStmtSyntax.self) {
                lowerWhile(loop, label: label)
            }
            else if let loop = statement.as(RepeatStmtSyntax.self) {
                lowerRepeat(loop, label: label)
            }
            else if let loop = statement.as(ForStmtSyntax.self) {
                lowerFor(loop, label: label)
            }
            else if let doStatement = statement.as(DoStmtSyntax.self) {
                lowerDo(doStatement, label: label)
            }
            else {
                /* `defer` runs at the scope's exit, which this layout does not model; anything else is not laid out at all. */
                opaque(Syntax(statement))
            }
        }

        func lowerVariable(_ declaration: VariableDeclSyntax) {
            let isLet = declaration.bindingSpecifier.tokenKind == .keyword(.let)
            let deferred = CorrectnessRequireResponseStatusCheck.isDeferred(declaration)
            for binding in declaration.bindings {
                if let accessors = binding.accessorBlock {
                    opaque(Syntax(accessors))
                }
                guard let initializer = binding.initializer, !deferred else {
                    if let initializer = binding.initializer {
                        opaque(Syntax(initializer))
                    }
                    bindPattern(Syntax(binding.pattern))
                    continue
                }
                walk(Syntax(initializer.value))
                if isLet, let pair = pairNames(binding.pattern), pair.body != nil || pair.response != nil,
                    let call = CorrectnessRequireResponseStatusCheck.stripped(initializer.value).as(
                        FunctionCallExprSyntax.self
                    ), producers.kind(of: call) == .async
                {
                    track(producer: Syntax(initializer.value), body: pair.body, response: pair.response)
                    continue
                }
                if isLet, let identifier = binding.pattern.as(IdentifierPatternSyntax.self),
                    let alias = aliasTarget(initializer.value, inCondition: false)
                {
                    bind(
                        CorrectnessRequireResponseStatusCheck.plainName(identifier.identifier),
                        .tracked(request: alias.request, role: alias.role),
                    )
                    continue
                }
                bindPattern(Syntax(binding.pattern))
            }
        }

        /* `(data, response)`, `(data, _)`, `(_, response)`: a two-element tuple of names or `_`, unlabeled. */
        func pairNames(_ pattern: PatternSyntax) -> (body: String?, response: String?)? {
            guard let tuple = pattern.as(TuplePatternSyntax.self), tuple.elements.count == 2 else { return nil }
            var names: [String?] = []
            for element in tuple.elements {
                guard element.label == nil else { return nil }
                if let identifier = element.pattern.as(IdentifierPatternSyntax.self) {
                    names.append(CorrectnessRequireResponseStatusCheck.plainName(identifier.identifier))
                }
                else if element.pattern.is(WildcardPatternSyntax.self) {
                    names.append(nil)
                }
                else {
                    return nil
                }
            }
            return (names[0], names[1])
        }

        /*
         The tracked name an initializer makes a new name stand for. In a condition, a tracked name itself
         (`if let data = data`); anywhere, a cast of a tracked response to `HTTPURLResponse`.
         */
        func aliasTarget(_ expression: ExprSyntax, inCondition: Bool) -> (request: Int, role: Role, cast: Bool)? {
            let inner = CorrectnessRequireResponseStatusCheck.stripped(expression)
            if inCondition, let reference = inner.as(DeclReferenceExprSyntax.self), reference.argumentNames == nil,
                case .tracked(let request, let role)? = lookup(
                    CorrectnessRequireResponseStatusCheck.plainName(reference.baseName)
                )
            {
                return (request, role, false)
            }
            if let cast = inner.as(AsExprSyntax.self), producers.isHttpResponse(cast.type),
                let reference = CorrectnessRequireResponseStatusCheck.stripped(cast.expression).as(
                    DeclReferenceExprSyntax.self
                ), reference.argumentNames == nil,
                case .tracked(let request, .response)? = lookup(
                    CorrectnessRequireResponseStatusCheck.plainName(reference.baseName)
                )
            {
                return (request, .response, true)
            }
            return nil
        }

        /* Every name a pattern binds, shadowing whatever it named before. */
        func bindPattern(_ pattern: Syntax) {
            for name in Self.boundNames(pattern) {
                bind(name, .other)
            }
        }

        /*
         The names a pattern binds: its identifier patterns, and the names inside a `let` or `var` written around an
         expression pattern (`case let .some(value)`), which swift-syntax keeps as references.
         */
        static func boundNames(_ node: Syntax, binding: Bool = false) -> [String] {
            if let identifier = node.as(IdentifierPatternSyntax.self) {
                return [CorrectnessRequireResponseStatusCheck.plainName(identifier.identifier)]
            }
            if binding, let reference = node.as(DeclReferenceExprSyntax.self) {
                return [CorrectnessRequireResponseStatusCheck.plainName(reference.baseName)]
            }
            let inside = binding || node.is(ValueBindingPatternSyntax.self)
            return node.children(viewMode: .sourceAccurate).flatMap { boundNames($0, binding: inside) }
        }

        /* The expressions a pattern evaluates (`case limit:` compares with `limit`), never the names it binds. */
        func walkPattern(_ node: Syntax, binding: Bool = false) {
            if let reference = node.as(DeclReferenceExprSyntax.self) {
                if binding {
                    visited.insert(reference.id)
                }
                else {
                    visitReference(reference)
                }
                return
            }
            if node.is(ExprSyntax.self), !node.is(PatternExprSyntax.self), !node.is(DeclReferenceExprSyntax.self),
                !binding
            {
                walk(node)
                return
            }
            let inside = binding || node.is(ValueBindingPatternSyntax.self)
            for child in node.children(viewMode: .sourceAccurate) {
                walkPattern(child, binding: inside)
            }
        }

        /* Leaves the branch the conditions failed at for `failure`, after a check when the failure is a cast of the response. */
        func fail(to failure: Int, checking request: Int?) {
            if let request {
                let checked = newBlock()
                link(current, checked)
                blocks[checked].events.append(Event(request: request, use: .check, report: nil))
                link(checked, failure)
            }
            else {
                link(current, failure)
            }
            let next = newBlock()
            link(current, next)
            current = next
        }

        func lowerConditions(_ conditions: ConditionElementListSyntax, failure: Int) {
            for element in conditions {
                switch element.condition {
                    case .expression(let expression):
                        walk(Syntax(expression))
                        if expression.as(BooleanLiteralExprSyntax.self)?.literal.tokenKind != .keyword(.true) {
                            fail(to: failure, checking: nil)
                        }
                    case .availability:
                        fail(to: failure, checking: nil)
                    case .optionalBinding(let optionalBinding):
                        var alias: (request: Int, role: Role, cast: Bool)?
                        if let initializer = optionalBinding.initializer {
                            walk(Syntax(initializer.value))
                            alias = aliasTarget(initializer.value, inCondition: true)
                        }
                        else if let identifier = optionalBinding.pattern.as(IdentifierPatternSyntax.self),
                            case .tracked(let request, let role)? = lookup(
                                CorrectnessRequireResponseStatusCheck.plainName(identifier.identifier)
                            )
                        {
                            /* `if let data`: a read of the outer name that binds the same body to the inner one. */
                            requests[request].referenced = true
                            alias = (request, role, false)
                        }
                        fail(to: failure, checking: alias?.cast == true ? alias?.request : nil)
                        if let alias, let identifier = optionalBinding.pattern.as(IdentifierPatternSyntax.self) {
                            bind(
                                CorrectnessRequireResponseStatusCheck.plainName(identifier.identifier),
                                .tracked(request: alias.request, role: alias.role),
                            )
                        }
                        else {
                            bindPattern(Syntax(optionalBinding.pattern))
                        }
                    case .matchingPattern(let matching):
                        walk(Syntax(matching.initializer.value))
                        walkPattern(Syntax(matching.pattern))
                        fail(to: failure, checking: nil)
                        bindPattern(Syntax(matching.pattern))
                }
            }
        }

        func lowerIf(_ conditional: IfExprSyntax, label: String?) {
            let elseEntry = newBlock()
            let join = newBlock()
            jumps.append(Jump(label: label, breakBlock: join, continueBlock: nil, takesUnlabeledBreak: false))
            scopes.append([:])
            lowerConditions(conditional.conditions, failure: elseEntry)
            lowerBlock(conditional.body)
            scopes.removeLast()
            link(current, join)
            current = elseEntry
            switch conditional.elseBody {
                case .ifExpr(let nested):
                    lowerIf(nested, label: nil)
                case .codeBlock(let block):
                    lowerBlock(block)
                case nil:
                    break
            }
            jumps.removeLast()
            link(current, join)
            current = join
        }

        func lowerGuard(_ guardStatement: GuardStmtSyntax) {
            let elseEntry = newBlock()
            scopes.append([:])
            lowerConditions(guardStatement.conditions, failure: elseEntry)
            let bound = scopes.removeLast()
            let continuation = current
            current = elseEntry
            /* The compiler holds `else` to leaving the scope, so its end leads nowhere. */
            lowerBlock(guardStatement.body)
            current = continuation
            for (name, binding) in bound {
                bind(name, binding)
            }
        }

        func lowerWhile(_ loop: WhileStmtSyntax, label: String?) {
            let header = newBlock()
            link(current, header)
            current = header
            let exit = newBlock()
            jumps.append(Jump(label: label, breakBlock: exit, continueBlock: header, takesUnlabeledBreak: true))
            scopes.append([:])
            lowerConditions(loop.conditions, failure: exit)
            lowerBlock(loop.body)
            link(current, header)
            scopes.removeLast()
            jumps.removeLast()
            current = exit
        }

        func lowerRepeat(_ loop: RepeatStmtSyntax, label: String?) {
            let bodyEntry = newBlock()
            link(current, bodyEntry)
            current = bodyEntry
            let conditionBlock = newBlock()
            let exit = newBlock()
            jumps.append(
                Jump(label: label, breakBlock: exit, continueBlock: conditionBlock, takesUnlabeledBreak: true)
            )
            lowerBlock(loop.body)
            jumps.removeLast()
            link(current, conditionBlock)
            current = conditionBlock
            walk(Syntax(loop.condition))
            link(current, bodyEntry)
            if loop.condition.as(BooleanLiteralExprSyntax.self)?.literal.tokenKind != .keyword(.true) {
                link(current, exit)
            }
            current = exit
        }

        func lowerFor(_ loop: ForStmtSyntax, label: String?) {
            walk(Syntax(loop.sequence))
            let header = newBlock()
            link(current, header)
            current = header
            if loop.tryKeyword != nil {
                throwingPoint()
            }
            let exit = newBlock()
            link(current, exit)
            let bodyEntry = newBlock()
            link(current, bodyEntry)
            current = bodyEntry
            jumps.append(Jump(label: label, breakBlock: exit, continueBlock: header, takesUnlabeledBreak: true))
            scopes.append([:])
            walkPattern(Syntax(loop.pattern))
            bindPattern(Syntax(loop.pattern))
            if let whereClause = loop.whereClause {
                walk(Syntax(whereClause.condition))
                fail(to: header, checking: nil)
            }
            lowerBlock(loop.body)
            link(current, header)
            scopes.removeLast()
            jumps.removeLast()
            current = exit
        }

        func lowerSwitch(_ switchExpression: SwitchExprSyntax, label: String?) {
            var cases: [SwitchCaseSyntax] = []
            for element in switchExpression.cases {
                guard case .switchCase(let switchCase) = element else {
                    /* A `#if` among the cases: which cases exist is the build's decision, not the source's. */
                    opaque(Syntax(switchExpression))
                    return
                }
                cases.append(switchCase)
            }
            walk(Syntax(switchExpression.subject))
            let join = newBlock()
            jumps.append(Jump(label: label, breakBlock: join, continueBlock: nil, takesUnlabeledBreak: true))
            let bodyEntries = cases.map { _ in newBlock() }
            var test = current
            for (index, switchCase) in cases.enumerated() {
                current = test
                scopes.append([:])
                switch switchCase.label {
                    case .default:
                        link(current, bodyEntries[index])
                    case .case(let caseLabel):
                        for item in caseLabel.caseItems {
                            walkPattern(Syntax(item.pattern))
                            bindPattern(Syntax(item.pattern))
                            if let whereClause = item.whereClause {
                                walk(Syntax(whereClause.condition))
                            }
                        }
                        link(current, bodyEntries[index])
                        /* The cases are exhaustive, so failing the last one leads nowhere. */
                        if index + 1 < cases.count {
                            let next = newBlock()
                            link(current, next)
                            test = next
                        }
                }
                current = bodyEntries[index]
                fallthroughTargets.append(index + 1 < cases.count ? bodyEntries[index + 1] : nil)
                lowerItems(switchCase.statements)
                fallthroughTargets.removeLast()
                link(current, join)
                scopes.removeLast()
            }
            jumps.removeLast()
            current = join
        }

        func lowerDo(_ doStatement: DoStmtSyntax, label: String?) {
            let join = newBlock()
            jumps.append(Jump(label: label, breakBlock: join, continueBlock: nil, takesUnlabeledBreak: false))
            defer { jumps.removeLast() }
            guard !doStatement.catchClauses.isEmpty else {
                lowerBlock(doStatement.body)
                link(current, join)
                current = join
                return
            }
            let dispatch = newBlock()
            catchFrames.append(CatchFrame(dispatch: dispatch))
            lowerBlock(doStatement.body)
            catchFrames.removeLast()
            link(current, join)
            for clause in doStatement.catchClauses {
                let entry = newBlock()
                link(dispatch, entry)
                current = entry
                scopes.append([:])
                if clause.catchItems.isEmpty {
                    bind("error", .other)
                }
                for item in clause.catchItems {
                    if let pattern = item.pattern {
                        walkPattern(Syntax(pattern))
                        bindPattern(Syntax(pattern))
                    }
                    if let whereClause = item.whereClause {
                        walk(Syntax(whereClause.condition))
                    }
                }
                lowerBlock(clause.body)
                scopes.removeLast()
                link(current, join)
            }
            current = join
        }

        /* A `try` completed inside a `do` with `catch` clauses: the first one forks to them, as ESLint's layout does. */
        func throwingPoint() {
            guard let frame = catchFrames.last, !frame.forked else { return }
            frame.forked = true
            link(current, frame.dispatch)
            let next = newBlock()
            link(current, next)
            current = next
        }

        /* An expression, in evaluation order. */
        func walk(_ node: Syntax) {
            if node.is(ClosureExprSyntax.self) {
                opaque(node)
            }
            else if let conditional = node.as(IfExprSyntax.self) {
                lowerIf(conditional, label: nil)
            }
            else if let switchExpression = node.as(SwitchExprSyntax.self) {
                lowerSwitch(switchExpression, label: nil)
            }
            else if let access = node.as(MemberAccessExprSyntax.self) {
                if let base = access.base {
                    walk(Syntax(base))
                }
                visited.insert(access.declName.id)
            }
            else if let component = node.as(KeyPathPropertyComponentSyntax.self) {
                visited.insert(component.declName.id)
            }
            else if let reference = node.as(DeclReferenceExprSyntax.self) {
                visitReference(reference)
            }
            else if let attempt = node.as(TryExprSyntax.self) {
                walk(Syntax(attempt.expression))
                if attempt.questionOrExclamationMark == nil {
                    throwingPoint()
                }
            }
            else if let call = node.as(FunctionCallExprSyntax.self) {
                for child in node.children(viewMode: .sourceAccurate) {
                    walk(child)
                }
                /* `fatalError` and `preconditionFailure` never return: the path ends here, and not at an exit. */
                if let callee = call.calledExpression.as(DeclReferenceExprSyntax.self),
                    ["fatalError", "preconditionFailure"].contains(callee.baseName.text)
                {
                    startUnreachable()
                }
            }
            else {
                for child in node.children(viewMode: .sourceAccurate) {
                    walk(child)
                }
            }
        }

        /*
         Code this layout does not walk: a closure, a nested function or type, a `defer`, a lazy initializer. A tracked
         name read in it drops its pair, because when that code runs is not something the graph can say.
         */
        func opaque(_ node: Syntax) {
            if let reference = node.as(DeclReferenceExprSyntax.self) {
                visited.insert(reference.id)
                if case .tracked(let request, _)? = lookup(
                    CorrectnessRequireResponseStatusCheck.plainName(reference.baseName)
                ) {
                    requests[request].referenced = true
                    requests[request].dropped = true
                }
            }
            for child in node.children(viewMode: .sourceAccurate) {
                opaque(child)
            }
        }

        func visitReference(_ reference: DeclReferenceExprSyntax) {
            visited.insert(reference.id)
            guard reference.argumentNames == nil,
                case .tracked(let request, let role)? = lookup(
                    CorrectnessRequireResponseStatusCheck.plainName(reference.baseName)
                )
            else { return }
            requests[request].referenced = true
            if let use = classify(reference, role: role) {
                emit(request, use, report: use == .bodyRead ? Syntax(reference) : nil)
            }
        }

        // MARK: Reading one use

        /* What one reference does to its pair; nil when it neither reads the body nor tells anything of the status. */
        func classify(_ reference: DeclReferenceExprSyntax, role: Role) -> Use? {
            let node = CorrectnessRequireResponseStatusCheck.outward(Syntax(reference))
            if Self.isConditionInitializer(node) {
                return nil
            }
            switch role {
                case .body:
                    return Self.isNilComparison(node) ? nil : .bodyRead
                case .response:
                    if let cast = node.parent?.as(AsExprSyntax.self), Syntax(cast.expression).id == node.id {
                        guard producers.isHttpResponse(cast.type) else { return .escape }
                        let castNode = CorrectnessRequireResponseStatusCheck.outward(Syntax(cast))
                        if Self.isConditionInitializer(castNode) || Self.isLetInitializer(castNode) {
                            return nil
                        }
                        return Self.member(of: castNode)
                    }
                    return Self.member(of: node)
            }
        }

        /* The initializer of an optional binding in a condition: `if let data = data`, `if let http = response as? HTTPURLResponse`. */
        static func isConditionInitializer(_ node: Syntax) -> Bool {
            node.parent?.as(InitializerClauseSyntax.self)?.parent?.is(OptionalBindingConditionSyntax.self) == true
        }

        /* The initializer of `let name = ...`, the one plain declaration a cast is followed through. */
        static func isLetInitializer(_ node: Syntax) -> Bool {
            guard let binding = node.parent?.as(InitializerClauseSyntax.self)?.parent?.as(PatternBindingSyntax.self),
                binding.pattern.is(IdentifierPatternSyntax.self),
                let declaration = binding.parent?.parent?.as(VariableDeclSyntax.self)
            else { return false }
            return declaration.bindingSpecifier.tokenKind == .keyword(.let)
                && !CorrectnessRequireResponseStatusCheck.isDeferred(declaration)
        }

        static func isNilComparison(_ node: Syntax) -> Bool {
            guard let comparison = node.parent?.as(InfixOperatorExprSyntax.self),
                let operation = comparison.operator.as(BinaryOperatorExprSyntax.self),
                ["==", "!="].contains(operation.operator.text)
            else { return false }
            let other = Syntax(comparison.leftOperand).id == node.id ? comparison.rightOperand : comparison.leftOperand
            return other.is(NilLiteralExprSyntax.self)
        }

        /* A member read off the response: `statusCode` checks, the neutral ones say nothing, anything else hands it on. */
        static func member(of node: Syntax) -> Use? {
            guard let access = node.parent?.as(MemberAccessExprSyntax.self),
                access.base.map({ Syntax($0).id == node.id }) == true
            else { return .escape }
            let member = CorrectnessRequireResponseStatusCheck.plainName(access.declName.baseName)
            if member == "statusCode" {
                return .check
            }
            if CorrectnessRequireResponseStatusCheck.neutralProperties.contains(member),
                access.declName.argumentNames == nil
            {
                return nil
            }
            if member == "value", let call = access.parent?.as(FunctionCallExprSyntax.self),
                Syntax(call.calledExpression).id == Syntax(access).id
            {
                return nil
            }
            return .escape
        }

        // MARK: Analysis

        func analyze(_ root: Root, completion: FunctionCallExprSyntax?) -> [Report] {
            if let completion, let closure = CorrectnessRequireResponseStatusCheck.handler(of: completion) {
                let names = Self.parameterNames(closure)
                track(
                    producer: Syntax(completion),
                    body: names.first ?? nil,
                    response: names.count > 1 ? names[1] : nil,
                )
            }
            lowerItems(root.items)
            blocks[current].isFinal = true
            guard !requests.isEmpty else { return [] }

            /* A tracked name the walk never reached is a use it cannot place, so its pair is not judged. */
            var unvisited: [String] = []
            Self.forEachReference(Syntax(root.items)) { reference in
                if !visited.contains(reference.id) {
                    unvisited.append(CorrectnessRequireResponseStatusCheck.plainName(reference.baseName))
                }
            }
            for name in unvisited {
                for index in requests.indices where requests[index].names.contains(name) {
                    requests[index].dropped = true
                }
            }

            var reachable = Array(repeating: false, count: blocks.count)
            var queue = [0]
            reachable[0] = true
            while let block = queue.popLast() {
                for successor in blocks[block].successors where !reachable[successor] {
                    reachable[successor] = true
                    queue.append(successor)
                }
            }

            var reports: [Report] = []
            for (index, request) in requests.enumerated() where !request.dropped {
                if !request.referenced {
                    reports.append(Report(node: request.producer, discarded: true))
                    continue
                }
                reports += uncheckedReads(index, reachable: reachable).map { Report(node: $0, discarded: false) }
            }
            return reports
        }

        /* The names a completion handler gives its body and its response: written, `_` (nil), or `$0` and `$1`. */
        static func parameterNames(_ closure: ClosureExprSyntax) -> [String?] {
            guard let signature = closure.signature else { return ["$0", "$1"] }
            func named(_ token: TokenSyntax) -> String? {
                token.tokenKind == .wildcard ? nil : CorrectnessRequireResponseStatusCheck.plainName(token)
            }
            switch signature.parameterClause {
                case .simpleInput(let parameters):
                    return parameters.map { named($0.name) }
                case .parameterClause(let clause):
                    return clause.parameters.map { named($0.secondName ?? $0.firstName) }
                case nil:
                    return []
            }
        }

        static func forEachReference(_ node: Syntax, _ body: (DeclReferenceExprSyntax) -> Void) {
            if let reference = node.as(DeclReferenceExprSyntax.self) {
                body(reference)
            }
            for child in node.children(viewMode: .sourceAccurate) {
                forEachReference(child, body)
            }
        }

        typealias Position = (block: Int, start: Int)

        /*
         The body reads of one pair that lie on an unchecked path: reachable from the declaration without passing a
         check or an escape, and leading on to a normal exit, or to the declaration again, without passing one either.
         */
        func uncheckedReads(_ request: Int, reachable: [Bool]) -> [Syntax] {
            var starts: [Position] = []
            for (index, block) in blocks.enumerated() where reachable[index] {
                for (position, event) in block.events.enumerated()
                where event.request == request && event.use == .declare {
                    starts.append((index, position + 1))
                }
            }

            var candidates: [Position] = []
            walkPaths(starts, reachable: reachable) { block, position in
                let event = blocks[block].events[position]
                guard event.request == request else { return true }
                guard event.use == .bodyRead else { return false }
                candidates.append((block, position))
                return true
            } completeBlock: { _ in
            }

            var reports: [Syntax] = []
            var reported: Set<SyntaxIdentifier> = []
            for candidate in candidates {
                guard let report = blocks[candidate.block].events[candidate.start].report, !reported.contains(report.id)
                else { continue }
                var unchecked = false
                walkPaths([(candidate.block, candidate.start + 1)], reachable: reachable) { block, position in
                    let event = blocks[block].events[position]
                    if event.request != request || unchecked {
                        return !unchecked
                    }
                    if event.use == .declare {
                        /* The next turn of a loop binds a fresh pair; this one was used and dropped. */
                        unchecked = true
                        return false
                    }
                    return event.use == .bodyRead
                } completeBlock: { block in
                    if blocks[block].isFinal {
                        unchecked = true
                    }
                }
                if unchecked {
                    reported.insert(report.id)
                    reports.append(report)
                }
            }
            return reports
        }

        /*
         Visits the events reachable from the starts, in order along each path. `visitEvent` says whether the path goes
         on past an event; `completeBlock` runs for every block a path walks to its end. Each block is entered from its
         top at most once, which is enough because the state a path carries is only whether it is still going.
         */
        func walkPaths(
            _ starts: [Position],
            reachable: [Bool],
            visitEvent: (Int, Int) -> Bool,
            completeBlock: (Int) -> Void,
        ) {
            var queue = starts
            var entered: Set<Int> = []
            var next = 0
            while next < queue.count {
                let position = queue[next]
                next += 1
                var stopped = false
                var index = position.start
                while index < blocks[position.block].events.count {
                    if !visitEvent(position.block, index) {
                        stopped = true
                        break
                    }
                    index += 1
                }
                if stopped {
                    continue
                }
                completeBlock(position.block)
                for successor in blocks[position.block].successors
                where reachable[successor] && !entered.contains(successor) {
                    entered.insert(successor)
                    queue.append((successor, 0))
                }
            }
        }
    }
}
