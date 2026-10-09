import Foundation
import SwiftSyntax

/*
 Every root target compiles in Swift 6 language mode.

 Kirk's ruling: Swift 6 is required, with each project migrated as it comes under `c`. Swift 6 mode is
 what makes strict concurrency checking an error rather than a suggestion, so a target in Swift 5 mode is
 a target where the compiler, the first gate, is quietly running weaker.

 The mode is SwiftPM's resolved answer (the target's own setting, else the package's list, else the
 tools version's default), so a target that says nothing in a 6.x manifest is correctly in Swift 6.
 */
public struct ToolchainRequireSwiftSixLanguageMode: PackageRule {
    public let name = "cohere-swift/toolchain-require-swift-6-language-mode"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    public func findings(in package: PackageModel, manifest: ParsedFile?) -> [FindingRecord] {
        package.targets
            .filter { PackageModel.isOlderLanguageMode($0.languageMode, "6") }
            .map { target in
                let message = RuleMessages.ToolchainRequireSwift6LanguageMode.languageModeBelowSix(
                    target: target.name,
                    languageMode: target.languageMode,
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
