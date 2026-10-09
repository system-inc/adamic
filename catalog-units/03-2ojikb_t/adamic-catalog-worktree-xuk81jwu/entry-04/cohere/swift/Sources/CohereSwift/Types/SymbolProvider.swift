import Foundation

/*
 Symbols for the files typed rules need, from the cheapest source that can vouch for each file as it stands.

 First the build's index store, which costs a record read: it holds every file the last build compiled, and a
 unit newer than the file vouches for it. Then sourcekitd in process, about half a second a file, for a file
 the index cannot vouch for: one the bodies-only path checked without building, or one edited since. A file
 neither can describe gets a reason instead, which the linter reports as unchecked.

 The package's local packages build into their own scratch paths, so each has its own store and its own
 compile commands; every one is consulted.
 */
struct SymbolProvider {
    struct Result: Sendable {
        var symbols: [String: FileSymbols]
        /* Files no source could describe, by path, with why. */
        var unavailable: [String: String]
        var fromIndex: Int
        var fromSourcekitd: Int
    }

    let scratchPaths: [URL]
    let runner: ProcessRunner
    /* Handed to every file's symbols, so a rule can tell our declarations from a dependency's. */
    var ownedModules: Set<String> = []

    func symbols(for files: [ParsedFile]) -> Result {
        var result = Result(symbols: [:], unavailable: [:], fromIndex: 0, fromSourcekitd: 0)
        guard !files.isEmpty else { return result }
        var stores: [IndexStore] = []
        if let library = try? IndexStore.toolchainLibraryPath(runner: runner) {
            stores = scratchPaths.compactMap {
                try? IndexStore(libraryPath: library, storePath: IndexStore.storePath(scratchPath: $0))
            }
        }
        var needSourcekitd: [ParsedFile] = []
        for file in files {
            if var found = stores.lazy.compactMap({ $0.symbols(of: file.url) }).first {
                found.ownedModules = ownedModules
                result.symbols[file.url.path] = found
                result.fromIndex += 1
            }
            else {
                needSourcekitd.append(file)
            }
        }
        guard !needSourcekitd.isEmpty else { return result }

        let tables = scratchPaths.compactMap { try? CompileCommands.read(scratchPath: $0) }
        let session: Sourcekitd
        do {
            session = try Sourcekitd.shared(runner: runner)
        }
        catch {
            for file in needSourcekitd {
                result.unavailable[file.url.path] =
                    "the build's index does not describe it as it stands, and sourcekitd could not load: \(error)"
            }
            return result
        }
        for file in needSourcekitd {
            guard let command = tables.lazy.compactMap({ $0.command(for: file.url) }).first else {
                result.unavailable[file.url.path] =
                    "the build's index does not describe it as it stands, and no build recorded how to compile it"
                continue
            }
            do {
                var found = try session.symbols(
                    file: file.url.resolvingSymlinksInPath().path,
                    arguments: command.arguments,
                )
                found.ownedModules = ownedModules
                result.symbols[file.url.path] = found
                result.fromSourcekitd += 1
            }
            catch {
                result.unavailable[file.url.path] =
                    "the build's index does not describe it as it stands, and sourcekitd could not index it: \(error)"
            }
        }
        return result
    }
}
