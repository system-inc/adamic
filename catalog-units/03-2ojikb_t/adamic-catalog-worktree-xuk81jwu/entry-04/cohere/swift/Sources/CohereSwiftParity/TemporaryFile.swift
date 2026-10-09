import Foundation

/* The incumbents' generated configurations, written to the temporary directory and removed after each run. */
enum TemporaryFile {
    static func url(extension pathExtension: String) -> URL {
        FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-parity-\(UUID().uuidString).\(pathExtension)"
        )
    }

    static func remove(_ url: URL) {
        do {
            try FileManager.default.removeItem(at: url)
        }
        catch {
            /* Ignored on purpose: a leftover configuration costs a few bytes in a directory the system sweeps, and failing a finished comparison over it would trade a real answer for tidiness. */
        }
    }
}
