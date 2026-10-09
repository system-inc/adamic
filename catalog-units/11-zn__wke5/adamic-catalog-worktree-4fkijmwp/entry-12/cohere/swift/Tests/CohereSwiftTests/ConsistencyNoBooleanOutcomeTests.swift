import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 `consistency-no-boolean-outcome` both ways. Each case returns its findings as the text each one covers, so a finding
 on the wrong field or the wrong declaration fails rather than passing on its count. The real sites come first:
 macOS's `DeveloperSettingsView` fetch results (an inline tuple, left alone as the original leaves an inline literal,
 then declared as a typealias and flagged, then repaired), Presence's `BodyRemoval.Outcome` (a flag with no companion),
 SwiftTerm's `createSharedMemory` tuple, and Presence's `SilhouetteMask.Counts` (a `failed` that is a count). Then the
 original's own cases, each one ported, then the places Swift needs its own exactness.
 */
struct ConsistencyNoBooleanOutcomeTests {
    /* Every finding as the text of its span, for a file the rule agrees to read, as the pipeline runs it. */
    static func findings(
        _ source: String,
        path: String = "/fixture/Subject.swift",
        targetKind: String = "library",
    ) -> [String] {
        let file = ParsedFile(
            url: URL(fileURLWithPath: path),
            targetName: "Fixture",
            targetKind: targetKind,
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let rule = ConsistencyNoBooleanOutcome()
        let found = rule.findings(in: file)
        #expect(found.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        let lines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return found.map { finding in
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
            #expect(finding.messageId == "booleanOutcome")
            guard let endLine = finding.endLine, let endColumn = finding.endColumn, endLine == finding.line else {
                return "spans lines"
            }
            return String(decoding: lines[finding.line - 1][(finding.column - 1)..<(endColumn - 1)], as: UTF8.self)
        }
    }

    static func messages(_ source: String) -> [String] {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        return ConsistencyNoBooleanOutcome().findings(in: file).map(\.message)
    }

    // MARK: The real sites

    /*
     `ahraos-macos` `Sources/AhraOs/DeveloperSettingsView.swift:35` on 2026-10-03: whether the daemon confirmed a pushed
     value, with the daemon's value, or the placeholder "no daemon" when there was none to ask. Three outcomes in one bit.
     */
    static func developerSettings(declaration: String, results: String) -> String {
        """
        import SwiftUI

        \(declaration)
        struct DeveloperSettingsView: View {
            /// Per-knob fetch-confirm results: defaultsKey → (matches, daemonValue).
            @State private var fetchResults: \(results) = [:]

            var body: some View { Text("Developer") }

            private func fetch(key: String, expected: String) {
                DevFlagsBridge.shared.fetchDevConfiguration { values, build in
                    guard let values, let daemon = values[key] else {
                        fetchResults[key] = (false, "no daemon")
                        return
                    }
                    fetchResults[key] = (daemon == expected, daemon)
                }
            }
        }

        """
    }

    /* As it stands, an inline tuple in a property's type: the original reads declarations only, and leaves `Record<string, { ok: boolean; value: string }>` alone too. */
    @Test func theFetchResultsInlineTupleIsLeftAlone() {
        #expect(
            Self.findings(Self.developerSettings(declaration: "", results: "[String: (ok: Bool, value: String)]"))
                .isEmpty
        )
        let outsideAView =
            "struct DaemonProbe {\n    var fetchResults: [String: (ok: Bool, value: String)] = [:]\n    func confirm() -> (ok: Bool, value: String) { (true, \"\") }\n}\n"
        #expect(Self.findings(outsideAView).isEmpty)
    }

    /* The same shape given a name is the original's `type` alias of an object literal. */
    @Test func theFetchResultDeclaredAsATypealiasIsFlagged() {
        let source = Self.developerSettings(
            declaration: "typealias FetchResult = (ok: Bool, value: String)",
            results: "[String: FetchResult]",
        )
        #expect(Self.findings(source) == ["ok: Bool"])
        #expect(Self.messages(source).first?.hasSuffix("Suggested name: FetchOutcome.") == true)
    }

    /* The repair names the three outcomes the bit collapsed. */
    @Test func theFetchResultRepairedAsAnEnumIsClean() {
        let declaration =
            "enum FetchOutcome {\n    case confirmed(String)\n    case mismatch(daemon: String)\n    case noDaemon\n}"
        #expect(
            Self.findings(Self.developerSettings(declaration: declaration, results: "[String: FetchOutcome]")).isEmpty
        )
    }

    /* `ahraos-presence` `Sources/AhraOSPresence/Studio/Bodies/BodyRemoval.swift:31`: `isFailure` beside `said`, which names no result. */
    static func bodyRemoval(_ text: String) -> String {
        """
        @MainActor
        enum BodyRemoval {
            /// What happened, in words for the alert after, and whether any of it went wrong.
            struct Outcome {
                let \(text): String
                let isFailure: Bool
            }
        }

        """
    }

    @Test func theBodyRemovalOutcomeHasNoCompanionAndIsClean() {
        #expect(Self.findings(Self.bodyRemoval("said")).isEmpty)
    }

    @Test func theBodyRemovalOutcomeWithAMessageIsFlagged() {
        #expect(Self.findings(Self.bodyRemoval("message")) == ["isFailure: Bool"])
    }

    /* `ahraos-macos` `Vendor/SwiftTerm/Tests/SwiftTermTests/KittyTransmissionTests.swift:41`: a return tuple, inline, whose companion is `errorCode`. */
    @Test func theSharedMemoryTupleIsClean() {
        let inline =
            "func createSharedMemory(name: String, bytes: [UInt8]) -> (ok: Bool, errorCode: Int32) {\n    (false, errno)\n}\n"
        #expect(Self.findings(inline, targetKind: "test").isEmpty)
        /* Named, it is still clean: `errorCode` is not one of the original's companions. */
        #expect(Self.findings("typealias SharedMemory = (ok: Bool, errorCode: Int32)\n", targetKind: "test").isEmpty)
        /* With `error`, it is the envelope, and a test target is read like any other. */
        #expect(
            Self.findings("typealias SharedMemory = (ok: Bool, error: Int32)\n", targetKind: "test") == ["ok: Bool"]
        )
    }

    /* `ahraos-presence` `Sources/AhraOSPresence/Stage/SilhouetteMask.swift:176`: a `failed` that counts, beside a `reason`, is no flag. */
    @Test func theSilhouetteCountsAreClean() {
        let source =
            "struct Counts {\n    let captures: Int\n    let publications: Int\n    let skipped: Int\n    let failed: Int\n    let reason: String\n}\n"
        #expect(Self.findings(source).isEmpty)
    }

    // MARK: The original's cases

    @Test func everyFlagWithACompanionIsFlagged() {
        #expect(
            Self.findings("struct LookupResult {\n    let success: Bool\n    let error: String\n}\n") == [
                "success: Bool"
            ]
        )
        #expect(Self.findings("typealias LookupResult = (ok: Bool, value: String)\n") == ["ok: Bool"])
        #expect(
            Self.findings("struct CallResult {\n    let failed: Bool\n    let reason: String\n}\n") == ["failed: Bool"]
        )
        #expect(
            Self.findings("struct CallResult {\n    let succeeded: Bool\n    let data: Data\n}\n") == [
                "succeeded: Bool"
            ]
        )
        #expect(
            Self.findings("struct CallResult {\n    let isError: Bool\n    let message: String\n}\n") == [
                "isError: Bool"
            ]
        )
        #expect(Self.findings("struct CallResult {\n    var isOk: Bool\n    var result: Int\n}\n") == ["isOk: Bool"])
        #expect(
            Self.findings("struct CallResult {\n    let isSuccess: Bool\n    let errors: [String]\n}\n") == [
                "isSuccess: Bool"
            ]
        )
        #expect(
            Self.findings("struct CallResult {\n    let isFailure: Bool\n    let value: Int\n}\n") == [
                "isFailure: Bool"
            ]
        )
    }

    /* A third-party envelope exempts only on its whole field set, so three of Cloudflare v4's four are still ours. */
    @Test func threeOfCloudflaresFourFieldsAreFlagged() {
        #expect(
            Self.findings(
                "struct SyncResult {\n    let success: Bool\n    let errors: [String]\n    let result: String\n}\n"
            ) == ["success: Bool"]
        )
    }

    @Test func theOriginalsExemptionsAreClean() {
        #expect(Self.findings("struct ToggleState {\n    let isVisible: Bool\n    let succeeded: Bool\n}\n").isEmpty)
        #expect(Self.findings("struct JobRow {\n    let id: String\n    let succeeded: Bool\n}\n").isEmpty)
        #expect(Self.findings("struct BannerProperties {\n    let isError: Bool\n    let message: String\n}\n").isEmpty)
        #expect(
            Self.findings(
                "struct QueryResult {\n    let isError: Bool\n    let isLoading: Bool\n    let error: String\n}\n"
            ).isEmpty
        )
        #expect(Self.findings("struct Reply {\n    let ok: Bool\n    let status: Int\n    let data: Data\n}\n").isEmpty)
        let cloudflare =
            "struct R2CloudflareApiResponse: Decodable {\n    let success: Bool\n    let errors: [Message]\n    let messages: [Message]\n    let result: Bucket?\n}\n"
        #expect(Self.findings(cloudflare).isEmpty)
        #expect(
            Self.findings(
                "struct QueryResult {\n    let isSuccess: Bool\n    let isError: Bool\n    let error: String\n}\n"
            ).isEmpty
        )
        #expect(Self.findings("struct Point {\n    let x: Double\n    let y: Double\n}\n").isEmpty)
    }

    /* The sum type is the shape the rule steers toward, so it never fires on one. */
    @Test func theRepairsAreClean() {
        #expect(
            Self.findings("enum PostOutcome {\n    case found(value: Post)\n    case notFound(message: String)\n}\n")
                .isEmpty
        )
        #expect(Self.findings("enum Arm {\n    case done(ok: Bool, error: String?)\n}\n").isEmpty)
        #expect(
            Self.findings(
                "typealias LookupOutcome = Result<Post, LookupFailure>\nfunc lookup() throws -> Post { post }\n"
            ).isEmpty
        )
    }

    /* Generated code gets every rule, as Kirk ruled for both languages. */
    @Test func aGeneratedFileIsHeldToTheRule() {
        #expect(
            Self.findings(
                "struct LookupResult {\n    let success: Bool\n    let error: String\n}\n",
                path: "/fixture/Api.generated.swift",
            ) == ["success: Bool"]
        )
    }

    @Test func theSuggestionStripsRoleSuffixes() {
        let message = Self.messages(
            "struct ClaudeCallResultInterface {\n    let success: Bool\n    let error: String\n}\n"
        )
        #expect(
            message == [
                "success: Bool on struct ClaudeCallResultInterface collapses every outcome into one bit, at the moment the distinction is cheapest to keep. It also leaves the reader to know which other fields hold for which value of the flag. Return a named outcome instead: an enum with a case for each way the operation can turn out and the payload on the case that carries it, or Result, or a throw for the failure. Suggested name: ClaudeCallOutcome."
            ]
        )
        #expect(
            Self.messages("typealias Result = (ok: Bool, value: Int)\n").first?.hasSuffix(
                "Suggested name: ResultOutcome."
            ) == true
        )
    }

    // MARK: Swift's own exactness

    /* `Swift.Bool` is the same type, and a binding without a type takes the one written after it. */
    @Test func theFlagsTypeIsReadAsTheCompilerReadsIt() {
        #expect(Self.findings("struct Reply {\n    let ok: Swift.Bool\n    let value: Int\n}\n") == ["ok: Swift.Bool"])
        #expect(Self.findings("struct Reply {\n    var ok, failed: Bool\n    var message: String\n}\n") == ["ok"])
        #expect(
            Self.findings("struct Reply {\n    var count = 0, ok: Bool\n    var message: String\n}\n") == ["ok: Bool"]
        )
    }

    /* Only the bare two-valued flag: an optional, an inferred type, a type of the same name in another module, an array. */
    @Test func anythingButABareBoolIsClean() {
        #expect(Self.findings("struct Reply {\n    let ok: Bool?\n    let value: Int\n}\n").isEmpty)
        #expect(Self.findings("struct Reply {\n    var ok = false\n    var value = 0\n}\n").isEmpty)
        #expect(Self.findings("struct Reply {\n    let ok: Flags.Bool\n    let value: Int\n}\n").isEmpty)
        #expect(Self.findings("struct Reply {\n    let ok: [Bool]\n    let value: Int\n}\n").isEmpty)
        #expect(
            Self.findings("struct Reply {\n    var ok = false, value: Bool\n    var message: String\n}\n") == [],
            "an initializer stops the trailing type",
        )
    }

    /* The first flag in declaration order that is a bare `Bool` is the one reported. */
    @Test func theFirstBareFlagIsReported() {
        #expect(
            Self.findings("struct Reply {\n    let failed: Bool?\n    let ok: Bool\n    let error: String?\n}\n") == [
                "ok: Bool"
            ]
        )
    }

    /* A field is a stored instance property: computed and static properties are not the value's fields. */
    @Test func onlyStoredPropertiesAreFields() {
        #expect(
            Self.findings("struct Reply {\n    let code: Int\n    var ok: Bool { code < 300 }\n    let data: Data\n}\n")
                .isEmpty
        )
        #expect(
            Self.findings(
                "struct Reply {\n    let ok: Bool\n    var message: String { ok ? \"fine\" : \"failed\" }\n}\n"
            ).isEmpty
        )
        #expect(
            Self.findings("struct Reply {\n    var ok: Bool {\n        get { true }\n    }\n    let value: Int\n}\n")
                .isEmpty
        )
        #expect(Self.findings("struct Reply {\n    static let ok: Bool = true\n    let value: Int\n}\n").isEmpty)
        #expect(
            Self.findings(
                "struct Reply {\n    var ok: Bool {\n        didSet { print(ok) }\n    }\n    var value: Int\n}\n"
            ) == ["ok: Bool"],
            "observers still store",
        )
    }

    /* The original listens to interfaces and object-literal aliases alone. */
    @Test func otherDeclarationsAreClean() {
        #expect(Self.findings("final class Reply {\n    let ok: Bool = true\n    let value: Int = 0\n}\n").isEmpty)
        #expect(Self.findings("actor Reply {\n    var ok: Bool = true\n    var value: Int = 0\n}\n").isEmpty)
        #expect(Self.findings("protocol Reply {\n    var ok: Bool { get }\n    var value: Int { get }\n}\n").isEmpty)
        #expect(Self.findings("typealias Reply = (ok: Bool, value: Int)?\n").isEmpty)
        #expect(Self.findings("typealias Reply = [(ok: Bool, value: Int)]\n").isEmpty)
        #expect(Self.findings("typealias Reply = (Bool, Int)\n").isEmpty)
    }

    /* A generic alias and a struct nested in a function are declarations too. */
    @Test func everyDeclarationIsRead() {
        #expect(Self.findings("typealias Reply<Value> = (ok: Bool, value: Value)\n") == ["ok: Bool"])
        #expect(
            Self.findings(
                "func probe() {\n    struct Step {\n        let failed: Bool\n        let reason: String\n    }\n}\n"
            ) == ["failed: Bool"]
        )
        #expect(
            Self.findings(
                "enum Outer {\n    struct Inner: Codable {\n        let ok: Bool\n        let error: String?\n    }\n}\n"
            ) == ["ok: Bool"]
        )
    }

    /* A SwiftUI component's fields are display state its parent hands down, as React properties are. */
    @Test func aSwiftUIComponentIsClean() {
        let banner = "    let isError: Bool\n    let message: String\n    var body: some View { Text(message) }\n}\n"
        #expect(Self.findings("struct Banner: View {\n" + banner).isEmpty)
        #expect(Self.findings("struct Banner: SwiftUI.View {\n" + banner).isEmpty)
        #expect(Self.findings("struct Banner: Equatable, View & Sendable {\n" + banner).isEmpty)
        #expect(Self.findings("struct Banner {\n" + banner + "extension Banner: View {}\n").isEmpty)
        #expect(
            Self.findings("struct Tint: ViewModifier {\n    let isError: Bool\n    let message: String\n}\n").isEmpty
        )
        #expect(
            Self.findings("struct Terminal: NSViewRepresentable {\n    let failed: Bool\n    let reason: String\n}\n")
                .isEmpty
        )
        /* Another conformance is not a component, and a component's nested result type is still read. */
        #expect(
            Self.findings("struct Banner: Equatable {\n    let isError: Bool\n    let message: String\n}\n") == [
                "isError: Bool"
            ]
        )
        #expect(
            Self.findings(
                "struct Banner: View {\n    struct Load {\n        let ok: Bool\n        let value: Int\n    }\n}\n"
            ) == ["ok: Bool"]
        )
    }

    /* Swift's acronym spelling is outside the original's names: a known miss, kept so the two languages agree. */
    @Test func namesOutsideTheOriginalsListAreClean() {
        #expect(Self.findings("struct Reply {\n    let isOK: Bool\n    let message: String\n}\n").isEmpty)
        #expect(Self.findings("struct Reply {\n    let isDone: Bool\n    let error: String?\n}\n").isEmpty)
    }
}
