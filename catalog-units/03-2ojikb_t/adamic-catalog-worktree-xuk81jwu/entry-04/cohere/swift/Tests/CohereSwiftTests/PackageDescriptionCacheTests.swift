import Foundation
import Testing

@testable import CohereSwift

/*
 The package-description cache, both ways, on a real package and a real `swift package describe`.

 A hit is proven rather than assumed: after a cold run fills the cache, the cached `describe` answer is
 doctored to name a target that does not exist, with every fingerprint left intact. A load that reports the
 doctored target read the cache. Then each change that alters SwiftPM's answer (a file added, a file
 removed, the manifest edited, a different toolchain) must make the doctored answer disappear, which is the
 warm-after-a-change test a cache that never invalidates would fail.
 */
@Suite(.serialized)
struct PackageDescriptionCacheTests {
    static let toolchain = "swiftlang-test.1"

    static func package() throws -> (root: URL, scratch: URL) {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-describe-\(UUID().uuidString)",
            isDirectory: true,
        )
        let sources = root.appendingPathComponent("Sources/Control", isDirectory: true)
        try FileManager.default.createDirectory(at: sources, withIntermediateDirectories: true)
        try manifest(targets: ["Control"]).write(
            to: root.appendingPathComponent("Package.swift"),
            atomically: true,
            encoding: .utf8,
        )
        try "struct Control {}\n".write(
            to: sources.appendingPathComponent("Control.swift"),
            atomically: true,
            encoding: .utf8,
        )
        return (root, root.appendingPathComponent(".scratch", isDirectory: true))
    }

    static func manifest(targets: [String]) -> String {
        "// swift-tools-version:6.0\nimport PackageDescription\nlet package = Package(name: \"Control\", targets: [\(targets.map { ".target(name: \"\($0)\")" }.joined(separator: ", "))])\n"
    }

    /* Fill the cache cold, then rename its target in the cached answer so a load that read it can be told apart. */
    static func doctor(scratch: URL) throws {
        let cache = PackageDescriptionCache(scratchPath: scratch)
        var entry = try JSONDecoder().decode(PackageDescriptionCache.Entry.self, from: Data(contentsOf: cache.file))
        let renamed = String(decoding: entry.members[0].describe, as: UTF8.self).replacingOccurrences(
            of: "\"name\" : \"Control\"",
            with: "\"name\" : \"FromTheCache\"",
        )
        .replacingOccurrences(of: "\"name\":\"Control\"", with: "\"name\":\"FromTheCache\"")
        #expect(
            renamed.contains("FromTheCache"),
            "the doctoring found no target name to rename, so this test would prove nothing",
        )
        entry.members[0].describe = Data(renamed.utf8)
        try JSONEncoder().encode(entry).write(to: cache.file)
    }

    static func targetNames(_ root: URL, _ scratch: URL, toolchain: String = toolchain) throws -> [String] {
        try PackageModel.load(root: root, scratchPath: scratch, toolchain: toolchain).targets.map(\.name)
    }

    @Test func aWarmRunReadsTheCache() throws {
        let (root, scratch) = try Self.package()
        #expect(try Self.targetNames(root, scratch) == ["Control"])
        try Self.doctor(scratch: scratch)
        #expect(try Self.targetNames(root, scratch).contains("FromTheCache"), "the second run described cold")
    }

    @Test(arguments: ["file added", "file removed", "manifest edited", "toolchain changed"])
    func aChangeSwiftPMWouldSeeIsNotAnsweredFromTheCache(change: String) throws {
        let (root, scratch) = try Self.package()
        let sources = root.appendingPathComponent("Sources/Control", isDirectory: true)
        try "struct Second {}\n".write(
            to: sources.appendingPathComponent("Second.swift"),
            atomically: true,
            encoding: .utf8,
        )
        _ = try Self.targetNames(root, scratch)
        try Self.doctor(scratch: scratch)

        var toolchain = Self.toolchain
        switch change {
            case "file added":
                try "struct Third {}\n".write(
                    to: sources.appendingPathComponent("Third.swift"),
                    atomically: true,
                    encoding: .utf8,
                )
            case "file removed":
                try FileManager.default.removeItem(at: sources.appendingPathComponent("Second.swift"))
            case "manifest edited":
                try FileManager.default.createDirectory(
                    at: root.appendingPathComponent("Sources/Extra"),
                    withIntermediateDirectories: true,
                )
                try "struct Extra {}\n".write(
                    to: root.appendingPathComponent("Sources/Extra/Extra.swift"),
                    atomically: true,
                    encoding: .utf8,
                )
                try Self.manifest(targets: ["Control", "Extra"]).write(
                    to: root.appendingPathComponent("Package.swift"),
                    atomically: true,
                    encoding: .utf8,
                )
            default:
                toolchain = "swiftlang-test.2"
        }

        let model = try PackageModel.load(root: root, scratchPath: scratch, toolchain: toolchain)
        #expect(
            !model.targets.map(\.name).contains("FromTheCache"),
            "after a \(change), the stale cached answer was used",
        )
        let files = Set(model.targets.flatMap(\.sources).map(\.lastPathComponent))
        switch change {
            case "file added": #expect(files.contains("Third.swift"))
            case "file removed": #expect(!files.contains("Second.swift"))
            case "manifest edited": #expect(model.targets.map(\.name) == ["Control", "Extra"])
            default: break
        }
    }

    /* An edit inside a file changes nothing SwiftPM answers, so it must not cost a cold description. */
    @Test func editingAFileKeepsTheCache() throws {
        let (root, scratch) = try Self.package()
        _ = try Self.targetNames(root, scratch)
        try Self.doctor(scratch: scratch)
        try "struct Control { let edited = 1 }\n".write(
            to: root.appendingPathComponent("Sources/Control/Control.swift"),
            atomically: false,
            encoding: .utf8,
        )
        #expect(try Self.targetNames(root, scratch).contains("FromTheCache"))
    }

    /* Without a known toolchain nothing is cached, so nothing can outlive a toolchain change. */
    @Test func anUnknownToolchainNeverCaches() throws {
        let (root, scratch) = try Self.package()
        _ = try Self.targetNames(root, scratch, toolchain: "unknown (no swift)")
        #expect(!FileManager.default.fileExists(atPath: PackageDescriptionCache(scratchPath: scratch).file.path))
    }
}
