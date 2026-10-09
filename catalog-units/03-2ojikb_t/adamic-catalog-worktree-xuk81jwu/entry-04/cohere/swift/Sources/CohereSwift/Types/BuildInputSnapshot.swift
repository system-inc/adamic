import CryptoKit
import Foundation

/*
 A fingerprint of everything a `swift build` reads, so the types phase can tell that nothing changed since the
 last build that succeeded and read that build's compiler records instead of building again. The no-op build
 costs 5.1s on ahraos-macos and 7.0s on ahraos-presence warm, most of a warm run; the walk below costs 7ms
 and 39ms.

 What goes in, and why each is enough:

 - every file under each package root and each path dependency, by path, size and modification time, hidden
   directories and `.build` left out. That is the signal SwiftPM's own incremental build trusts, so this is
   exactly as safe as the build it replaces, and it catches what a source-only walk would miss: a header
   outside a target, a resource, `Packages/` from `swift package edit`;
 - the root's `Package.resolved` by content, which pins every remote dependency's checkout;
 - the toolchain and the build's own arguments, so a toolchain change or `--no-fix`'s disabled resolution is
   a different build.

 The fingerprint is taken before a build starts and stored only after every build succeeded, so an edit made
 while a build ran, or a build that failed or was killed, always means building next time. A cache that
 never invalidates reports the last answer forever, which on a measurement looks exactly like a clean tree,
 so the test that matters is warm-after-a-change (`BuildInputSnapshotTests`, and the pipeline control that
 plants a type error after a warm run).
 */
struct BuildInputSnapshot: Equatable, Codable {
    let fingerprint: String
    /* The toolchain, the build's arguments and the resolved pins, one line each. */
    let settings: [String]
    /* Every file the build reads, by path, to its size and modification time. */
    let files: [String: String]
    /*
     Each owned Swift file's `InterfaceFingerprint` as the build about to run compiles it, by resolved path. The
     types phase fills it in before building; it is what lets a later run tell that an edit stayed inside
     function bodies.
     */
    var interfaces: [String: String] = [:]

    /* Equal when the build would read the same inputs. The interfaces are derived from those inputs, so they never decide. */
    static func == (left: BuildInputSnapshot, right: BuildInputSnapshot) -> Bool {
        left.fingerprint == right.fingerprint
    }

    static func take(roots: [URL], resolved: URL, toolchain: String, arguments: [String]) -> BuildInputSnapshot {
        let settings = [
            "toolchain\t\(toolchain)", "arguments\t\(arguments.joined(separator: " "))",
            "resolved\t\(PackageDescriptionCache.fingerprint(of: resolved))",
        ]
        var files: [String: String] = [:]
        let keys: [URLResourceKey] = [.isDirectoryKey, .fileSizeKey, .contentModificationDateKey]
        for root in roots {
            let walker = FileManager.default.enumerator(
                at: root,
                includingPropertiesForKeys: keys,
                options: [.skipsHiddenFiles],
            )
            while let item = walker?.nextObject() as? URL {
                guard let values = try? item.resourceValues(forKeys: Set(keys)) else {
                    files[item.path] = "unreadable"
                    continue
                }
                if values.isDirectory == true {
                    if item.lastPathComponent == ".build" {
                        walker?.skipDescendants()
                    }
                    continue
                }
                let modified = values.contentModificationDate?.timeIntervalSinceReferenceDate ?? -1
                files[item.path] = "\(values.fileSize ?? -1)\t\(modified)"
            }
        }
        let lines = settings + files.map { "\($0.key)\t\($0.value)" }.sorted()
        let digest = SHA256.hash(data: Data(lines.joined(separator: "\n").utf8))
        return BuildInputSnapshot(
            fingerprint: digest.map { String($0 >> 4, radix: 16) + String($0 & 0x0f, radix: 16) }.joined(),
            settings: settings,
            files: files,
        )
    }

    /*
     The files that changed since `earlier`, when nothing else did: the same settings and the same set of
     files, only some of them edited. Nil when anything else moved (a file added or removed, a pin, the
     toolchain), because then only a build can say what the compiler thinks.
     */
    func filesChanged(since earlier: BuildInputSnapshot) -> Set<String>? {
        guard settings == earlier.settings, files.count == earlier.files.count else { return nil }
        var changed = Set<String>()
        for (path, entry) in files {
            guard let before = earlier.files[path] else { return nil }
            if before != entry {
                changed.insert(path)
            }
        }
        return changed
    }

    /* Where the last successful build's snapshot is kept, beside that build in the engine's scratch. */
    static func storeFile(scratchPath: URL) -> URL {
        scratchPath.appendingPathComponent("build-inputs.json")
    }

    /* The last successful build's snapshot. An older engine's store, a bare digest, does not decode and reads as none: the next run builds. */
    static func stored(scratchPath: URL) -> BuildInputSnapshot? {
        guard let data = try? Data(contentsOf: storeFile(scratchPath: scratchPath)), !data.isEmpty else { return nil }
        return try? JSONDecoder().decode(BuildInputSnapshot.self, from: data)
    }

    func store(scratchPath: URL) {
        do {
            try FileManager.default.createDirectory(at: scratchPath, withIntermediateDirectories: true)
            try JSONEncoder().encode(self).write(to: Self.storeFile(scratchPath: scratchPath), options: .atomic)
        }
        catch {
            /* Not stored, so the next run builds: slower, and still right. */
        }
    }

    /*
     Forgotten after a build that did not succeed, so the next run cannot read a failed build's records as
     current. Overwritten with nothing rather than removed: an empty store reads as no snapshot, and a removal
     that failed would have left the old fingerprint standing.
     */
    static func forget(scratchPath: URL) {
        let file = storeFile(scratchPath: scratchPath)
        guard FileManager.default.fileExists(atPath: file.path) else { return }
        do {
            try Data().write(to: file, options: .atomic)
        }
        catch {
            /* Only an unwritable scratch reaches here, and then nothing can be stored either, so the next run builds. */
        }
    }
}
