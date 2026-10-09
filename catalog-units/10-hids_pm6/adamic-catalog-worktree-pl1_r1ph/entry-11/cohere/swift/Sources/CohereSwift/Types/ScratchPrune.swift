import Foundation

/*
 What the engine's build scratch keeps that nothing reads any more, removed after a build succeeds. One project
 has one scratch (`<root>/.cache/cohere/swift`), so the scratch itself never multiplies; what grows is inside it.
 The rule is Kirk's, the one the release cache keeps (#ve5gps3, bc85c90): keep what is current and anything used
 within the last hour, re-check right before removing, and remove one named file at a time, a directory only once
 its files are gone.

 - The index store's records. A compile writes its file's record under a name made from the record's content and
   points the file's unit at it, so every edit that changes what a file declares or references leaves the old
   record behind, named by no unit. The store reaches a record only through a unit, so a record no unit names is
   one nothing can read. Measured on the proving-ground worktrees after a day of edits: 226 of 2,969 records
   (6.9 MB of 38 MB) on ahraos-macos and 300 of 6,747 (12.9 MB of 93 MB) on Presence; fresh copies have none. A
   record written within the hour stays, and if any unit is written after the plan was made, a build is running
   against the store and nothing is removed.
 - A local package's scratch (`local/<package>`), built for that package's tests. It is kept while the package
   model lists the package as local, and a build refreshes its time; one whose package is no longer local goes
   once it has not been used for an hour.

 Not pruned: the build descriptions under `out/Intermediates.noindex/XCBuildData`. The build system keeps them
 in an on-disk cache with a size limit of its own (`BuildDescriptionOnDiskCacheSize` in swift-build, evicted
 through `prior-build-descriptions.txt`), and both proving grounds hold exactly four however many builds they
 ran. Removing one it still lists would cost a rebuilt description and gain nothing the limit does not.
 */
struct ScratchPrune {
    /* Anything touched this recently may belong to a build still running, or one about to read it again. */
    static let inUseWindow: TimeInterval = 60 * 60

    /* The file in the scratch whose time says when the prune last ran there. */
    static let stampName = "last-prune"

    struct Outcome: Equatable {
        var recordsRemoved = 0
        var localScratchesRemoved: [String] = []
        var bytesRemoved = 0

        /* The words the types record's build line carries, empty when nothing went. */
        var sentence: String {
            guard recordsRemoved > 0 || !localScratchesRemoved.isEmpty else { return "" }
            var parts: [String] = []
            if recordsRemoved > 0 {
                parts.append("\(recordsRemoved) index records no unit names")
            }
            if !localScratchesRemoved.isEmpty {
                parts.append(
                    "the scratch of \(localScratchesRemoved.sorted().joined(separator: ", ")), no longer a local package"
                )
            }
            let size = ByteCountFormatter.string(fromByteCount: Int64(bytesRemoved), countStyle: .file)
            return "the scratch prune removed \(size): \(parts.joined(separator: "; "))"
        }
    }

    /* The index records to remove from one store, decided without touching anything, and when the decision was made. */
    struct RecordPlan {
        var records: [URL]
        var decidedAt: Date
    }

    /*
     The records under `<storePath>/v5/records` that no unit of the store names and that were not written within the
     window. Exact: a unit lists every record it reads, and a reader finds records only through units.
     */
    static func planRecords(store: IndexStore, storePath: URL, now: Date) -> RecordPlan {
        let named = Set(store.units().flatMap(\.allRecords))
        let recordsDirectory = storePath.appendingPathComponent("v5/records", isDirectory: true)
        var records: [URL] = []
        for shard in contents(of: recordsDirectory) {
            for record in contents(of: shard) where !named.contains(record.lastPathComponent) {
                if let modified = modificationDate(of: record), now.timeIntervalSince(modified) >= inUseWindow {
                    records.append(record)
                }
            }
        }
        return RecordPlan(records: records.sorted { $0.path < $1.path }, decidedAt: now)
    }

    /*
     Removes a plan's records, unless a unit was written after the plan was decided: that is a build writing into the
     store, and a record the plan thought unnamed may be named again by it. Each record's time is asked again right
     before it goes. Returns how many went and their bytes.
     */
    static func apply(_ plan: RecordPlan, storePath: URL) -> (count: Int, bytes: Int) {
        let units = storePath.appendingPathComponent("v5/units", isDirectory: true)
        let newestUnit = contents(of: units).compactMap(modificationDate(of:)).max() ?? .distantPast
        guard newestUnit <= plan.decidedAt else { return (0, 0) }
        var count = 0
        var bytes = 0
        for record in plan.records {
            guard let modified = modificationDate(of: record), Date().timeIntervalSince(modified) >= inUseWindow else {
                continue
            }
            let size = fileSize(of: record)
            if (try? FileManager.default.removeItem(at: record)) != nil {
                count += 1
                bytes += size
            }
        }
        return (count, bytes)
    }

    /* The local package scratches under `<scratch>/local` whose package is not listed and that were not used within the window. */
    static func staleLocalScratches(scratchPath: URL, listed: Set<String>, now: Date) -> [URL] {
        contents(of: scratchPath.appendingPathComponent("local", isDirectory: true)).filter { scratch in
            guard !listed.contains(scratch.lastPathComponent), let modified = modificationDate(of: scratch) else {
                return false
            }
            return now.timeIntervalSince(modified) >= inUseWindow
        }
    }

    /* Marks a local package's scratch as used now, so the hour counts from the last build into it. */
    static func markUsed(_ scratch: URL) {
        do {
            try FileManager.default.setAttributes([.modificationDate: Date()], ofItemAtPath: scratch.path)
        }
        catch {
            /*
             The scratch of a package the model lists is kept whatever its time, so a mark that failed matters only once
             the package stops being local, and then it costs at most a rebuild of a scratch nothing builds into.
             */
        }
    }

    /*
     Removes one directory: its time asked again first, then every file by name, then each directory once it is
     empty, deepest first. Never one recursive call, so a path that is not what the plan believed costs at most the
     files named under it. Returns the bytes removed, or nil when the directory was used in the meantime.
     */
    static func removeDirectory(_ directory: URL) -> Int? {
        guard let modified = modificationDate(of: directory), Date().timeIntervalSince(modified) >= inUseWindow else {
            return nil
        }
        var files: [URL] = []
        var directories: [URL] = [directory]
        let walker = FileManager.default.enumerator(
            at: directory,
            includingPropertiesForKeys: [.isDirectoryKey, .isSymbolicLinkKey],
            options: [],
        )
        while let item = walker?.nextObject() as? URL {
            let values = try? item.resourceValues(forKeys: [.isDirectoryKey, .isSymbolicLinkKey])
            if values?.isDirectory == true && values?.isSymbolicLink != true {
                directories.append(item)
            }
            else {
                files.append(item)
            }
        }
        var bytes = 0
        for file in files {
            let size = fileSize(of: file)
            if (try? FileManager.default.removeItem(at: file)) != nil {
                bytes += size
            }
        }
        /*
         rmdir, not removeItem: it removes only an empty directory, so a file that could not go keeps its parents.
         Unsafe because the path crosses into C as a pointer; correct because Swift lends the C string only for the
         call, and rmdir reads it and keeps nothing.
         */
        for emptied in directories.sorted(by: { $0.pathComponents.count > $1.pathComponents.count }) {
            _ = unsafe Darwin.rmdir(emptied.path)
        }
        return bytes
    }

    /*
     The whole prune for one run: each scratch's index records, and the root scratch's stale local package
     scratches. Called only after every build of the run succeeded, so no build of this run is in flight. At most
     once an hour per scratch, recorded by the stamp's time: nothing becomes removable sooner than an hour after
     its last use, and reading every unit of every store costs 0.4s after a build on ahraos-macos and 3.4s on
     Presence, measured.
     */
    static func run(
        stores: [(store: IndexStore, storePath: URL)],
        scratchPath: URL,
        listedLocalPackages: Set<String>,
        now: Date = Date(),
    ) -> Outcome {
        var outcome = Outcome()
        let stamp = scratchPath.appendingPathComponent(stampName)
        if let last = modificationDate(of: stamp), now.timeIntervalSince(last) < inUseWindow {
            return outcome
        }
        FileManager.default.createFile(atPath: stamp.path, contents: nil)
        for (store, storePath) in stores {
            let removed = apply(planRecords(store: store, storePath: storePath, now: now), storePath: storePath)
            outcome.recordsRemoved += removed.count
            outcome.bytesRemoved += removed.bytes
        }
        for scratch in staleLocalScratches(scratchPath: scratchPath, listed: listedLocalPackages, now: now) {
            if let bytes = removeDirectory(scratch) {
                outcome.localScratchesRemoved.append(scratch.lastPathComponent)
                outcome.bytesRemoved += bytes
            }
        }
        return outcome
    }

    private static func contents(of directory: URL) -> [URL] {
        (try? FileManager.default.contentsOfDirectory(
            at: directory,
            includingPropertiesForKeys: [.contentModificationDateKey],
            options: [],
        )) ?? []
    }

    private static func modificationDate(of url: URL) -> Date? {
        (try? FileManager.default.attributesOfItem(atPath: url.path))?[.modificationDate] as? Date
    }

    private static func fileSize(of url: URL) -> Int {
        ((try? FileManager.default.attributesOfItem(atPath: url.path))?[.size] as? NSNumber)?.intValue ?? 0
    }
}
