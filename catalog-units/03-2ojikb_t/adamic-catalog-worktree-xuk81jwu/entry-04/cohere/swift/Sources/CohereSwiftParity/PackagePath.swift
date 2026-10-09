import Foundation

/* One spelling of a path for all three tools: relative to the package root, symlinks resolved (`/tmp` is `/private/tmp`). */
enum PackagePath {
    static func relative(_ path: String, to root: URL) -> String {
        let resolved = URL(fileURLWithPath: path).resolvingSymlinksInPath().path
        let prefix = root.resolvingSymlinksInPath().path + "/"
        return resolved.hasPrefix(prefix) ? String(resolved.dropFirst(prefix.count)) : resolved
    }
}
