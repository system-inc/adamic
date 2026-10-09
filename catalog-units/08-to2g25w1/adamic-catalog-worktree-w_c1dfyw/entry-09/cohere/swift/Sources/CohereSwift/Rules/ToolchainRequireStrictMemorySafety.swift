import Foundation
import SwiftSyntax

/*
 Every target we own compiles under strict memory safety (SE-0458, `.strictMemorySafety()` in its
 `swiftSettings`), so every use of an unsafe construct is visible and marked where it happens.

 Kirk ruled it on 2026-10-02 (#fzaq352), house-wide, staged a repository at a time. With the setting on, the
 compiler warns at each expression that uses an unsafe pointer, an unchecked conversion or an `@unsafe`
 declaration without saying `unsafe`. The purpose is not the marking. Each flagged site is first asked whether
 a safe API does the job (a span, a typed buffer, a scope narrowed to the one line that needs the pointer), and
 only what truly must stay unsafe is marked. The warnings themselves are compiler findings, so the types phase
 reports them; this rule only requires the setting that makes the compiler say them.

 Read from SwiftPM's own answer (`dump-package`), as `toolchain-require-upcoming-features` reads its features. Plugin
 and macro targets, and targets with no Swift sources, are skipped for that rule's reasons.
 */
public struct ToolchainRequireStrictMemorySafety: PackageRule {
    public let name = "cohere-swift/toolchain-require-strict-memory-safety"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    public func findings(in package: PackageModel, manifest: ParsedFile?) -> [FindingRecord] {
        package.targets
            .filter { $0.kind != "plugin" && $0.kind != "macro" && !$0.sources.isEmpty && !$0.strictMemorySafety }
            .map { target in
                let message = RuleMessages.ToolchainRequireStrictMemorySafety.strictMemorySafetyMissing(
                    target: target.name
                )
                if let manifest, let declaration = TargetNameFinder.find(target.name, in: manifest.tree) {
                    return manifest.finding(
                        at: declaration,
                        rule: name,
                        message: message,
                    )
                }
                /* No declaration found (a target named by a computed string): point at the manifest's first line rather than drop the finding. */
                return FindingRecord(
                    source: .rule,
                    file: package.root.appendingPathComponent("Package.swift").path,
                    line: 1,
                    column: 1,
                    severity: .error,
                    rule: name,
                    messageId: message.id,
                    message: message.text,
                )
            }
    }
}
