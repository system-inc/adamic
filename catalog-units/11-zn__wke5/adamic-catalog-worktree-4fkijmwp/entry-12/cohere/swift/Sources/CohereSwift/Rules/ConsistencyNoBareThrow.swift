import SwiftOperators
import SwiftParser
import SwiftSyntax

/*
 No throwing an error made up at the throw site: `throw NSError(domain: "DanceImport", code: 1, userInfo:
 [NSLocalizedDescriptionKey: "no video track"])`. The domain is a string nothing else declares and the code a
 number nothing else names, so the failure has no identity a caller can catch: matching it means repeating
 both, `(error as NSError).domain == "DanceImport" && code == 1`, and nothing keeps that in step with the throw.
 The repair declares the failure as a case of an error type of our own (an enum conforming to `Error`, and to
 `LocalizedError` for its text) and throws that case, so a caller catches it by case and the compiler checks
 the match.

 It ports `base/no-bare-throw`, which refuses `throw new Error('x')` and the other built-in constructors that
 carry a message and no identifier. Swift cannot throw a bare string or a plain message, since only a type
 conforming to `Error` can be thrown, so the shape that survives is the one error built from a message and
 nothing else: `NSError` whose domain is a string literal written at the throw site. The pieces map one to one:
 - The anchor is the thrown expression of a `throw` statement, and the finding covers the `NSError(...)` call,
   as the original reports the `new` expression and leaves the `throw` keyword outside it.
 - The constructor is `NSError`, `Foundation.NSError`, or either one's `.init`. The original reads only a bare
   identifier, so `Errors.Validation` is not its business; `Foundation.NSError` is the same built-in spelled with
   its module, so it is read, and a type of ours named `NSError` inside another namespace is not.
 - The original's anchor is exactly a `new` expression, which leaves `throw cause ?? new Error('x')` unread. Swift
   reaches for that shape far more often, to throw a framework's optional error or a fallback
   (`throw reader.error ?? NSError(domain: "DanceImport", code: 2)`), and the fallback is the same ad-hoc error
   the rule exists for, built at the same throw site. So the right side of a `??` and both branches of a ternary
   are read too, through parentheses. That is the one place this rule reaches past the original's syntax, and
   it is the original's reasoning, not a wider idea.
 - The original exempts tests and the scaffolding they run on, by path, because a test's thrown message is the
   diagnosis and nothing branches on it. Here that is the target kind: a file of a test target is not read,
   decided inside the rule as `consistency-no-print` decides it, never by a path list. The original's other exemptions (the
   capture path, declaration-time and below-the-vocabulary directories) name places in one TypeScript
   repository that have no Swift counterpart.

 What it accepts as safe, and why:
 - A domain that is not a string literal: `NSError(domain: NSPOSIXErrorDomain, code: Int(result))`. A domain
   the SDK declares names a catalog of codes the caller can match (`POSIXError.Code`, `CocoaError.Code`), so the
   failure is declared, only by Apple.
 - A literal spelling of such a domain: a value starting `NS`, `kCF` or `com.apple.`, or ending `ErrorDomain`,
   the conventions every SDK domain constant follows (`"NSCocoaErrorDomain"`, `"AVFoundationErrorDomain"`,
   `"com.apple.LocalAuthentication"`).
 - Every typed error, whatever it carries: `CocoaError(.fileReadCorruptFile, userInfo: [...])`,
   `URLError(.badServerResponse)`, `POSIXError(...)`, `CancellationError()`, an error type of ours. Each has a
   code or a case a caller can match, as the original leaves a declared subclass alone.
 - An `NSError` built for an Objective-C API that requires one and not thrown: a completion handler's argument,
   an `NSErrorPointer`'s `pointee`, a delegate callback. Those are a third-party contract, which the original
   also leaves legal, and none of them is a `throw`.

 Known misses, every one a finding not made and never one invented:
 - An error built anywhere but the throw statement and thrown by name, `let error = NSError(...)` then
   `throw error`, or returned by a helper and thrown as `throw fail("...")` (Presence's `MToonLockerSky`), as the
   original reads neither a rethrow nor a call.
 - An ad-hoc error delivered by anything but `throw`: `continuation.resume(throwing:)`, `Result.failure(...)`,
   `promise(.failure(...))`. The original reads only throw statements and misses `reject(new Error())` the same
   way.
 - A domain held in a constant of ours (`NSError(domain: Self.errorDomain, ...)`): the rule cannot tell it from
   an SDK constant without resolving the name. And an ad-hoc literal that follows the SDK's spelling
   (`"com.ahra.ImportErrorDomain"`), which is exempt with the real ones.
 - An `NSError` subclass of ours, `NSError` wrapped in a cast (`throw NSError(...) as Error`), and one on the left
   of a `??`, where it is never optional and the compiler already warns.
 - A test-support file compiled into a library target rather than a test target, which is read and may report.
 */
public struct ConsistencyNoBareThrow: FileRule {
    public let name = "cohere-swift/consistency-no-bare-throw"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* The prefixes and suffix every domain constant the SDK declares is spelled with, so a literal with one names Apple's catalog. */
    static let platformDomainPrefixes = ["NS", "kCF", "com.apple."]
    static let platformDomainSuffix = "ErrorDomain"

    /* A test target's throws are its diagnosis, and every flagged shape spells `throw` and `NSError`. */
    public func applies(to file: ParsedFile) -> Bool {
        file.targetKind != "test" && file.source.contains("throw") && file.source.contains("NSError")
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        /* Folded, so `cause ?? NSError(...)` is one `??` and a ternary one node. Folding moves no token. */
        let folded = OperatorTable.standardOperators.foldAll(file.tree) { _ in }
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(folded)
        return visitor.found.map { construction, domain in
            file.finding(
                at: construction,
                rule: name,
                message: RuleMessages.ConsistencyNoBareThrow.bareThrow(domain: domain.trimmedDescription),
            )
        }
    }

    /* Collects every ad-hoc `NSError(...)` a `throw` statement throws, with its domain literal for the message. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [(FunctionCallExprSyntax, StringLiteralExprSyntax)] = []

        override func visit(_ node: ThrowStmtSyntax) -> SyntaxVisitorContinueKind {
            found += Self.adHocErrors(thrownBy: node.expression)
            return .visitChildren
        }

        /* The thrown expression itself, the right side of a `??`, or a ternary's branches, through parentheses. */
        static func adHocErrors(thrownBy expression: ExprSyntax) -> [(FunctionCallExprSyntax, StringLiteralExprSyntax)]
        {
            if let tuple = expression.as(TupleExprSyntax.self), tuple.elements.count == 1,
                let only = tuple.elements.first, only.label == nil
            {
                return adHocErrors(thrownBy: only.expression)
            }
            if let infix = expression.as(InfixOperatorExprSyntax.self) {
                guard infix.operator.as(BinaryOperatorExprSyntax.self)?.operator.text == "??" else { return [] }
                return adHocErrors(thrownBy: infix.rightOperand)
            }
            if let ternary = expression.as(TernaryExprSyntax.self) {
                return adHocErrors(thrownBy: ternary.thenExpression) + adHocErrors(thrownBy: ternary.elseExpression)
            }
            guard let call = expression.as(FunctionCallExprSyntax.self), isNSError(call.calledExpression),
                let domain = adHocDomain(call)
            else {
                return []
            }
            return [(call, domain)]
        }

        /* `NSError`, `Foundation.NSError`, or either one's `.init`. */
        static func isNSError(_ callee: ExprSyntax) -> Bool {
            if let reference = callee.as(DeclReferenceExprSyntax.self) {
                return reference.baseName.text == "NSError" && reference.argumentNames == nil
            }
            guard let member = callee.as(MemberAccessExprSyntax.self), member.declName.argumentNames == nil,
                let base = member.base
            else { return false }
            if member.declName.baseName.text == "NSError" {
                return base.as(DeclReferenceExprSyntax.self)?.baseName.text == "Foundation"
            }
            return member.declName.baseName.tokenKind == .keyword(.`init`) && isNSError(base)
        }

        /* The `domain:` argument when it is a string literal that does not spell one of the SDK's domains. */
        static func adHocDomain(_ call: FunctionCallExprSyntax) -> StringLiteralExprSyntax? {
            guard let first = call.arguments.first, first.label?.text == "domain",
                let literal = first.expression.as(StringLiteralExprSyntax.self)
            else {
                return nil
            }
            /* An interpolated domain has no single value and is assembled here, so it is ad hoc whatever it spells. */
            if let value = literal.representedLiteralValue,
                ConsistencyNoBareThrow.platformDomainPrefixes.contains(where: { value.hasPrefix($0) })
                    || value.hasSuffix(ConsistencyNoBareThrow.platformDomainSuffix)
            {
                return nil
            }
            return literal
        }
    }
}
