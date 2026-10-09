import Foundation

/*
 What this binary is and what it was built from, for the provenance record.

 The swift-syntax and swift-format versions are written down here and in `Package.swift`, and a test
 reads `Package.resolved` and fails if they disagree, because a provenance line that names the wrong
 parser is worse than none: it sends a reader to the wrong source when a finding looks wrong.

 The commit is `dev` unless the release stamped it. Go stamps the TypeScript engine at link time, and
 Swift has no equivalent flag, and a manifest can pass only boolean defines, never a string. So the
 release writes `EngineReleaseStamp.generated.swift` beside this file, holding the 40-hex commit inside
 `#if COHERE_RELEASE_STAMP`, builds with `-Xswiftc -DCOHERE_RELEASE_STAMP`, and removes the file. The file
 is gitignored, so the stamp never marks the tree modified (measured: SwiftPM's `hasUncommittedChanges`
 skips ignored files). Neither half can fail quietly: a build without the define says `dev` even with a
 stale file lying there, and a build with the define and no file does not compile.
 */
public enum EngineVersion {
    public static let contract = 3
    public static let engine = "cohere-swift"
    public static let version = "0.1.0"
    #if COHERE_RELEASE_STAMP
        public static let commit = EngineReleaseStamp.commit
    #else
        public static let commit = "dev"
    #endif
    public static let swiftSyntax = "604.0.0"
    public static let swiftFormat = "604.0.0"

    public static var sourceTreeModified: Bool {
        #if COHERE_SOURCE_TREE_MODIFIED
            true
        #else
            false
        #endif
    }

    /* The toolchain, as `swift --version` names it: `swiftlang-6.4.0.34.1`. Asked at run time, because it is also the toolchain that will build the checked package. */
    public static func toolchain(runner: ProcessRunner = ProcessRunner()) -> String {
        do {
            let result = try runner.run("swift", ["--version"], in: URL(fileURLWithPath: "/"))
            let text = String(decoding: result.standardOutput, as: UTF8.self) + result.standardError
            if let range = text.range(of: #"swiftlang-[0-9.]+"#, options: .regularExpression) {
                return String(text[range])
            }
            return "unknown (swift --version said: \(text.split(separator: "\n").first ?? ""))"
        }
        catch {
            return "unknown (\(error))"
        }
    }

    public static func provenance(toolchain: String) -> ProvenanceRecord {
        ProvenanceRecord(
            contract: contract,
            engine: engine,
            version: version,
            commit: commit,
            sourceTreeModified: sourceTreeModified,
            toolchain: toolchain,
            swiftSyntax: swiftSyntax,
            swiftFormat: swiftFormat,
        )
    }
}
