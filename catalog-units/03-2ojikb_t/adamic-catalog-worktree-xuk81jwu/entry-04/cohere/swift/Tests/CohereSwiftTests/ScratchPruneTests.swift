import Foundation
import Testing

@testable import CohereSwift

/*
 The scratch prune's keep sets. Local package scratches on a made directory: one the package model lists, one
 unlisted but used within the hour, one unlisted and stale, which alone goes, every file of it. Index records on a
 real store: a tiny package built, edited and built again, so its first record is left behind named by no unit.
 That record goes once it is an hour old, the records the units name never go, nothing goes when a unit was
 written after the plan, and the build descriptions are never touched.
 */
@Suite(.serialized)
struct ScratchPruneTests {
    static func makeDirectory(_ name: String) throws -> URL {
        let directory = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-prune-\(name)-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        return directory
    }

    static func age(_ url: URL, by seconds: TimeInterval) throws {
        try FileManager.default.setAttributes(
            [.modificationDate: Date().addingTimeInterval(-seconds)],
            ofItemAtPath: url.path,
        )
    }

    @Test func onlyAnUnlistedLocalScratchUnusedForAnHourGoes() throws {
        let scratch = try Self.makeDirectory("local")
        let local = scratch.appendingPathComponent("local", isDirectory: true)
        for name in ["Listed", "Fresh", "Gone"] {
            let nested = local.appendingPathComponent("\(name)/out/debug", isDirectory: true)
            try FileManager.default.createDirectory(at: nested, withIntermediateDirectories: true)
            try Data(repeating: 7, count: 1_000).write(to: nested.appendingPathComponent("artifact.o"))
            try Data(repeating: 7, count: 500).write(to: local.appendingPathComponent("\(name)/build.db"))
        }
        try Self.age(local.appendingPathComponent("Listed"), by: 7_200)
        try Self.age(local.appendingPathComponent("Gone"), by: 7_200)

        let stale = ScratchPrune.staleLocalScratches(scratchPath: scratch, listed: ["Listed"], now: Date())
        #expect(stale.map(\.lastPathComponent) == ["Gone"], "listed, and used within the hour, both stay: \(stale)")

        let outcome = ScratchPrune.run(stores: [], scratchPath: scratch, listedLocalPackages: ["Listed"])
        #expect(outcome.localScratchesRemoved == ["Gone"])
        #expect(outcome.bytesRemoved == 1_500)
        #expect(outcome.sentence.hasPrefix("the scratch prune removed 2 KB: "), "\(outcome.sentence)")
        #expect(!FileManager.default.fileExists(atPath: local.appendingPathComponent("Gone").path))
        #expect(
            FileManager.default.fileExists(atPath: local.appendingPathComponent("Listed/out/debug/artifact.o").path)
        )
        #expect(FileManager.default.fileExists(atPath: local.appendingPathComponent("Fresh/out/debug/artifact.o").path))
        #expect(outcome.sentence.contains("the scratch of Gone, no longer a local package"), "\(outcome.sentence)")
    }

    /* The prune runs at most once an hour per scratch: a stale scratch waits while the stamp is fresh, and goes once it is not. */
    @Test func thePruneRunsAtMostOnceAnHour() throws {
        let scratch = try Self.makeDirectory("stamp")
        let gone = scratch.appendingPathComponent("local/Gone", isDirectory: true)
        try FileManager.default.createDirectory(at: gone, withIntermediateDirectories: true)
        try Self.age(gone, by: 7_200)
        let stamp = scratch.appendingPathComponent(ScratchPrune.stampName)
        FileManager.default.createFile(atPath: stamp.path, contents: nil)
        #expect(ScratchPrune.run(stores: [], scratchPath: scratch, listedLocalPackages: []) == ScratchPrune.Outcome())
        #expect(FileManager.default.fileExists(atPath: gone.path))
        try Self.age(stamp, by: 7_200)
        #expect(
            ScratchPrune.run(stores: [], scratchPath: scratch, listedLocalPackages: []).localScratchesRemoved == [
                "Gone"
            ]
        )
        #expect(
            Date().timeIntervalSince(
                try #require(try FileManager.default.attributesOfItem(atPath: stamp.path)[.modificationDate] as? Date)
            ) < 60,
            "the run stamps its time",
        )
    }

    /* A scratch used between the plan and the removal stays: its time is asked again first. */
    @Test func aLocalScratchUsedAfterThePlanStays() throws {
        let scratch = try Self.makeDirectory("used")
        let gone = scratch.appendingPathComponent("local/Gone", isDirectory: true)
        try FileManager.default.createDirectory(at: gone, withIntermediateDirectories: true)
        try Self.age(gone, by: 7_200)
        let stale = ScratchPrune.staleLocalScratches(scratchPath: scratch, listed: [], now: Date())
        #expect(stale.count == 1)
        ScratchPrune.markUsed(gone)
        #expect(ScratchPrune.removeDirectory(gone) == nil)
        #expect(FileManager.default.fileExists(atPath: gone.path))
    }

    static let firstSource = "public struct Gauge { public init() {} }\npublic func reading() -> Int { 1 }\n"
    static let secondSource =
        "public struct Gauge { public init() {} }\npublic func reading() -> Int { 1 }\npublic func spare() -> Int { 2 }\n"

    /* A package built, edited so its file declares something new, and built again. */
    static func editedBuild() throws -> (scratch: URL, store: IndexStore, storePath: URL) {
        let root = try makeDirectory("records")
        let manifest =
            "// swift-tools-version:6.0\nimport PackageDescription\nlet package = Package(name: \"Gauge\", targets: [.target(name: \"Gauge\")])\n"
        try manifest.write(to: root.appendingPathComponent("Package.swift"), atomically: true, encoding: .utf8)
        let source = root.appendingPathComponent("Sources/Gauge/Gauge.swift")
        try FileManager.default.createDirectory(
            at: source.deletingLastPathComponent(),
            withIntermediateDirectories: true,
        )
        let scratch = root.appendingPathComponent(".cache/cohere/swift", isDirectory: true)
        let runner = ProcessRunner()
        for text in [firstSource, secondSource] {
            try text.write(to: source, atomically: true, encoding: .utf8)
            let built = try runner.run(
                "swift",
                ["build", "--build-system", "swiftbuild", "--scratch-path", scratch.path],
                in: root,
            )
            try #require(built.succeeded, "the fixture package must build: \(built.standardError)")
        }
        let storePath = IndexStore.storePath(scratchPath: scratch)
        return (
            scratch, try IndexStore(libraryPath: IndexStore.toolchainLibraryPath(), storePath: storePath), storePath,
        )
    }

    static func allRecords(_ storePath: URL) -> [URL] {
        let shards =
            (try? FileManager.default.contentsOfDirectory(
                at: storePath.appendingPathComponent("v5/records"),
                includingPropertiesForKeys: nil,
            )) ?? []
        return shards.flatMap {
            (try? FileManager.default.contentsOfDirectory(at: $0, includingPropertiesForKeys: nil)) ?? []
        }
    }

    @Test func aRecordNoUnitNamesGoesOnceAnHourOld() throws {
        let (scratch, store, storePath) = try Self.editedBuild()
        let named = Set(store.units().flatMap(\.allRecords))
        let orphans = Self.allRecords(storePath).filter { !named.contains($0.lastPathComponent) }
        try #require(!orphans.isEmpty, "the edit must leave the first record behind, or this test checks nothing")

        /* Written moments ago: kept, as a build's record would be. */
        #expect(ScratchPrune.planRecords(store: store, storePath: storePath, now: Date()).records.isEmpty)

        for record in Self.allRecords(storePath) {
            try Self.age(record, by: 7_200)
        }
        let plan = ScratchPrune.planRecords(store: store, storePath: storePath, now: Date())
        #expect(
            Set(plan.records.map(\.lastPathComponent)) == Set(orphans.map(\.lastPathComponent)),
            "exactly the records no unit names",
        )

        let descriptions = scratch.appendingPathComponent("out/Intermediates.noindex/XCBuildData", isDirectory: true)
        let before = try FileManager.default.contentsOfDirectory(atPath: descriptions.path).sorted()
        let removed = ScratchPrune.apply(plan, storePath: storePath)
        #expect(removed.count == orphans.count)
        #expect(removed.bytes > 0)
        let remaining = Set(Self.allRecords(storePath).map(\.lastPathComponent))
        #expect(named.isSubset(of: remaining), "every record a unit names stays")
        #expect(
            try FileManager.default.contentsOfDirectory(atPath: descriptions.path).sorted() == before,
            "build descriptions are the build system's to keep",
        )
    }

    /* A unit written after the plan means a build is writing into the store: nothing goes. */
    @Test func nothingGoesWhenAUnitIsWrittenAfterThePlan() throws {
        let (_, store, storePath) = try Self.editedBuild()
        for record in Self.allRecords(storePath) {
            try Self.age(record, by: 7_200)
        }
        var plan = ScratchPrune.planRecords(store: store, storePath: storePath, now: Date())
        try #require(!plan.records.isEmpty)
        plan.decidedAt = Date().addingTimeInterval(-60)
        let removed = ScratchPrune.apply(plan, storePath: storePath)
        #expect(removed.count == 0)
        #expect(plan.records.allSatisfy { FileManager.default.fileExists(atPath: $0.path) })
    }
}
