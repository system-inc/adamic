import Foundation

/*
 One of swift-format's own lint rules, run inside this process over the tree the engine already parsed,
 reported under a cohere name.

 Where the incumbent's rule is exactly the house rule, running it is the strongest parity there is: the
 findings are swift-format's by construction, and nothing can drift. `always-use-lower-camel-case` is
 `AlwaysUseLowerCamelCase` (with its exemptions: `override`s, and underscores in XCTest and `@Test` method
 names), and `no-leading-underscores` is `NoLeadingUnderscores`. The house adds the reason to each message;
 the location and the name are swift-format's. Each wrapped rule renders its own message and its own failure
 to run: `always-use-lower-camel-case` from policy/messages/, and `no-leading-underscores`, which has no
 message file yet, from its text here.

 The swift-format pass itself lives in `SwiftFormatPass`, run once per file for every wrapped rule, with
 only those rules enabled and the pretty-printer off, so it is a rule walk, not a second format. Neither
 rule follows the house format's rule table: a house rule applies whatever the formatter's table turns
 off.
 */
public struct SwiftFormatRule: FileRule {
    public let name: String
    /*
     Declared per wrapped rule rather than implied by the wrapper: running swift-format's rule is how the
     house checks a thing, and a house rule could be checked this way too.
     */
    public let origin: RuleOrigin
    public let upstreamName: String?
    let incumbentRule: String
    /* The finding's message, from swift-format's own text for it, capitalized as a sentence. */
    let message: @Sendable (_ swiftFormatFinding: String) -> RuleMessages.Message
    /* The message when swift-format's rule could not run on a file. */
    let failure: @Sendable (_ incumbentRule: String, _ error: String) -> RuleMessages.Message

    init(
        name: String,
        origin: RuleOrigin,
        upstreamName: String?,
        incumbentRule: String,
        message: @escaping @Sendable (_ swiftFormatFinding: String) -> RuleMessages.Message,
        failure: @escaping @Sendable (_ incumbentRule: String, _ error: String) -> RuleMessages.Message,
    ) {
        self.name = name
        self.origin = origin
        self.upstreamName = upstreamName
        self.incumbentRule = incumbentRule
        self.message = message
        self.failure = failure
    }

    public static let requireLowerCamelCase = SwiftFormatRule(
        name: "cohere-swift/always-use-lower-camel-case",
        origin: .swiftFormat,
        upstreamName: "AlwaysUseLowerCamelCase",
        incumbentRule: "AlwaysUseLowerCamelCase",
        message: { swiftFormatFinding in
            RuleMessages.AlwaysUseLowerCamelCase.alwaysUseLowerCamelCase(swiftFormatFinding: swiftFormatFinding)
        },
        failure: { incumbentRule, error in
            RuleMessages.AlwaysUseLowerCamelCase.incumbentFailed(incumbentRule: incumbentRule, error: error)
        },
    )

    public static let noLeadingUnderscores = SwiftFormatRule(
        name: "cohere-swift/no-leading-underscores",
        origin: .swiftFormat,
        upstreamName: "NoLeadingUnderscores",
        incumbentRule: "NoLeadingUnderscores",
        message: { swiftFormatFinding in
            RuleMessages.Message(
                id: "NoLeadingUnderscores",
                text:
                    "\(swiftFormatFinding). A leading underscore is a convention for \"private\", and access control says that in a way the compiler checks.",
            )
        },
        failure: { incumbentRule, error in
            RuleMessages.Message(
                id: "incumbentFailed",
                text: "swift-format's \(incumbentRule) could not run on this file, so it was not checked: \(error)",
            )
        },
    )

    /* Every incumbent rule this wrapper runs, so one swift-format pass answers for all of them. */
    static let incumbentRules = [requireLowerCamelCase.incumbentRule, noLeadingUnderscores.incumbentRule]

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let reported: SwiftFormatPass.Outcome
        do {
            reported = try SwiftFormatPass.shared.findings(in: file)
        }
        catch {
            /*
             Reported rather than dropped: a rule that threw checked nothing in this file, and an empty list
             would read as a clean file.
             */
            let rendered = failure(incumbentRule, "\(error)")
            return [
                FindingRecord(
                    source: .rule,
                    file: file.url.path,
                    line: 1,
                    column: 1,
                    severity: .error,
                    rule: name,
                    messageId: rendered.id,
                    message: rendered.text,
                )
            ]
        }
        return reported.filter { $0.rule == incumbentRule }.map { finding in
            let rendered = message(finding.text.prefix(1).uppercased() + finding.text.dropFirst())
            return FindingRecord(
                source: .rule,
                file: file.url.path,
                line: finding.line,
                column: finding.column,
                severity: .error,
                rule: name,
                messageId: rendered.id,
                message: rendered.text,
            )
        }
    }
}
