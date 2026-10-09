import CohereSwift
import Foundation
import Testing

/*
 The provenance names the parser and formatter that produced every finding, so it must name the ones
 actually resolved. This reads `Package.resolved` and fails when the constants drift from it.
 */
struct EngineVersionTests {
    static let resolvedFile = URL(fileURLWithPath: #filePath)
        .deletingLastPathComponent()
        .deletingLastPathComponent()
        .deletingLastPathComponent()
        .appendingPathComponent("Package.resolved")

    static func resolvedVersion(of identity: String) throws -> String? {
        let document = try JSONSerialization.jsonObject(with: Data(contentsOf: resolvedFile)) as? [String: Any]
        let pins = document?["pins"] as? [[String: Any]] ?? []
        let pin = pins.first { ($0["identity"] as? String) == identity }
        return (pin?["state"] as? [String: Any])?["version"] as? String
    }

    @Test func swiftSyntaxMatchesWhatWasResolved() throws {
        #expect(try Self.resolvedVersion(of: "swift-syntax") == EngineVersion.swiftSyntax)
    }

    @Test func swiftFormatMatchesWhatWasResolved() throws {
        #expect(try Self.resolvedVersion(of: "swift-format") == EngineVersion.swiftFormat)
    }
}
