import CryptoKit
import Foundation

/*
 What `swift package describe` and `dump-package` said last time, kept in the engine's scratch so a warm
 run does not ask again. Asking costs 1.7s on ahraos-macos and 3.4s on ahraos-presence, warm, on every run,
 and the answer changes only when the package does.

 The answer depends on three things, and the entry holds a fingerprint of each:

 - the toolchain, since a new SwiftPM can describe the same package differently;
 - every owned `Package.swift` and the root's `Package.resolved`, by content;
 - the shape of every target's directory. `describe` resolves a target's sources from the disk, so adding,
   removing or renaming a file changes its answer while the manifest stays the same. Each of those changes
   the modification time of the directory holding the file, and editing a file's contents changes neither
   the directory nor the answer, so the entry records every directory under each target with its time.

 Any difference, a missing entry, or an entry that does not decode means describing cold. A cache that
 kept answering after the package changed would check the wrong files and still print a verdict, which is
 the stale-cache failure `internal/types/program/incremental.go` warns about, so the test that matters is
 warm-after-a-change (`PackageDescriptionCacheTests`).
 */
struct PackageDescriptionCache {
    /* One package's two answers, with the fingerprint they were given under. */
    struct Member: Codable, Equatable {
        var root: String
        var manifest: String
        var describe: Data
        var dump: Data
        var directories: [DirectoryState]
    }

    struct DirectoryState: Codable, Equatable {
        var path: String
        var modified: Double
    }

    struct Entry: Codable {
        var toolchain: String
        var resolved: String
        var members: [Member]
    }

    let file: URL

    init(scratchPath: URL) {
        file = scratchPath.appendingPathComponent("package-description.json")
    }

    /* The cached answers for this root, only if every fingerprint still holds. */
    func members(root: URL, toolchain: String) -> [String: Member]? {
        guard let data = try? Data(contentsOf: file), let entry = try? JSONDecoder().decode(Entry.self, from: data)
        else {
            return nil
        }
        guard entry.toolchain == toolchain,
            entry.resolved == Self.fingerprint(of: root.appendingPathComponent("Package.resolved"))
        else {
            return nil
        }
        var byRoot: [String: Member] = [:]
        for member in entry.members {
            let memberRoot = URL(fileURLWithPath: member.root, isDirectory: true)
            guard member.manifest == Self.fingerprint(of: memberRoot.appendingPathComponent("Package.swift")),
                let model = try? PackageModel(
                    root: memberRoot,
                    describeJson: member.describe,
                    dumpPackageJson: member.dump,
                ),
                Self.directoryStates(of: model) == member.directories
            else {
                return nil
            }
            byRoot[member.root] = member
        }
        return byRoot
    }

    func store(root: URL, toolchain: String, answers: [(root: URL, describe: Data, dump: Data, model: PackageModel)]) {
        let entry = Entry(
            toolchain: toolchain,
            resolved: Self.fingerprint(of: root.appendingPathComponent("Package.resolved")),
            members: answers.map { answer in
                Member(
                    root: answer.root.path,
                    manifest: Self.fingerprint(of: answer.root.appendingPathComponent("Package.swift")),
                    describe: answer.describe,
                    dump: answer.dump,
                    directories: Self.directoryStates(of: answer.model),
                )
            },
        )
        do {
            try FileManager.default.createDirectory(
                at: file.deletingLastPathComponent(),
                withIntermediateDirectories: true,
            )
            try JSONEncoder().encode(entry).write(to: file, options: .atomic)
        }
        catch {
            /* Not written, so the next run describes cold: slower, and still right. Failing a run that already has its answer over a cache would trade the answer for speed. */
        }
    }

    /* A file's content hash, or a fixed word when it is absent, so "no Package.resolved" is a fingerprint too. */
    static func fingerprint(of file: URL) -> String {
        guard let data = try? Data(contentsOf: file) else { return "absent" }
        return SHA256.hash(data: data).map { String($0 >> 4, radix: 16) + String($0 & 0x0f, radix: 16) }.joined()
    }

    /* Every directory under every target, with its modification time, sorted so two walks compare equal. */
    static func directoryStates(of model: PackageModel) -> [DirectoryState] {
        var states: [DirectoryState] = []
        let keys: [URLResourceKey] = [.isDirectoryKey, .contentModificationDateKey]
        for target in model.targets {
            states.append(state(of: target.directory))
            guard let walker = FileManager.default.enumerator(at: target.directory, includingPropertiesForKeys: keys)
            else { continue }
            for case let item as URL in walker {
                guard (try? item.resourceValues(forKeys: [.isDirectoryKey]))?.isDirectory == true else { continue }
                states.append(state(of: item))
            }
        }
        return states.sorted { $0.path < $1.path }
    }

    private static func state(of directory: URL) -> DirectoryState {
        let modified = (try? directory.resourceValues(forKeys: [.contentModificationDateKey]))?.contentModificationDate
        return DirectoryState(path: directory.path, modified: modified?.timeIntervalSinceReferenceDate ?? -1)
    }
}
