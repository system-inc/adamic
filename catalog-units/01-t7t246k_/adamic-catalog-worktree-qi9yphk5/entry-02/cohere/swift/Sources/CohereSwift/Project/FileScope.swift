import Foundation

/*
 Which of our files this run is about: all of them, or the ones named on the command line.

 A narrowed scope narrows fix, format and lint. It never narrows types: in Swift a change in one file
 changes what the rest of its module means, and the compiler checks the whole module regardless, so the
 types phase always reports the whole package and the coverage line says so.
 */
public struct FileScope: Equatable, Sendable {
    public var files: [FileSet.OwnedFile]
    public var everything: Bool
    /* What was asked for, in words, printed in the parenthesis of `N in scope (…)`. Empty for the whole package. */
    public var description: String

    public init(files: [FileSet.OwnedFile], everything: Bool, description: String) {
        self.files = files
        self.everything = everything
        self.description = description
    }

    public static func resolve(options: CommandOptions, fileSet: FileSet, workingDirectory: URL) throws -> FileScope {
        if !options.paths.isEmpty {
            return try named(options.paths, fileSet: fileSet, workingDirectory: workingDirectory)
        }
        return FileScope(files: fileSet.owned, everything: true, description: "")
    }

    static func named(_ paths: [String], fileSet: FileSet, workingDirectory: URL) throws -> FileScope {
        var prefixes: [String] = []
        for path in paths {
            let url = URL(fileURLWithPath: path, relativeTo: workingDirectory).standardizedFileURL
                .resolvingSymlinksInPath()
            guard FileManager.default.fileExists(atPath: url.path) else {
                throw CommandOptions.UsageFailure(
                    description: "\(path) does not exist, so there is nothing named to check"
                )
            }
            prefixes.append(url.path)
        }
        let files = fileSet.owned.filter { owned in
            let path = owned.url.resolvingSymlinksInPath().path
            return prefixes.contains { path == $0 || path.hasPrefix($0 + "/") }
        }
        return FileScope(files: files, everything: false, description: paths.joined(separator: ", "))
    }
}
