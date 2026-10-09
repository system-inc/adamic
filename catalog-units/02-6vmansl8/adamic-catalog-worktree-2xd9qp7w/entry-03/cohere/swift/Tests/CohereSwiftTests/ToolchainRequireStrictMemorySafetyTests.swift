import CohereSwift
import Foundation
import Testing

/* The strict-memory-safety rule, from `dump-package` output shaped as SwiftPM 6.4 writes the setting: `{"kind":{"strictMemorySafety":{}}}`. */
struct ToolchainRequireStrictMemorySafetyTests {
    static let root = URL(fileURLWithPath: "/package", isDirectory: true)

    static func model(_ targets: [(name: String, type: String, strict: Bool)]) throws -> PackageModel {
        let described = targets.map {
            #"{"name":"\#($0.name)","type":"\#($0.type)","path":"Sources/\#($0.name)","sources":["\#($0.name).swift"]}"#
        }
        let dumped = targets.map { target in
            let settings = target.strict ? #"{"kind":{"strictMemorySafety":{}},"tool":"swift"}"# : ""
            return #"{"name":"\#(target.name)","settings":[\#(settings)]}"#
        }
        return try PackageModel(
            root: root,
            describeJson: Data(#"{"name":"Example","targets":[\#(described.joined(separator: ","))]}"#.utf8),
            dumpPackageJson: Data(
                #"{"toolsVersion":{"_version":"6.2.0"},"swiftLanguageVersions":null,"targets":[\#(dumped.joined(separator: ","))]}"#
                    .utf8
            ),
        )
    }

    static func flagged(_ model: PackageModel) -> [String] {
        ToolchainRequireStrictMemorySafety().findings(in: model, manifest: nil).map {
            String($0.message.prefix { $0 != " " })
        }
    }

    @Test func aTargetWithTheSettingPasses() throws {
        #expect(Self.flagged(try Self.model([("Core", "library", true)])).isEmpty)
    }

    /* Both directions on one package, so a rule that flagged everything, or nothing, fails here. */
    @Test func onlyTheTargetWithoutItIsNamed() throws {
        let findings = ToolchainRequireStrictMemorySafety().findings(
            in: try Self.model([("Core", "library", true), ("App", "executable", false)]),
            manifest: nil,
        )
        #expect(findings.map { String($0.message.prefix { $0 != " " }) } == ["App"])
        #expect(findings.first?.message.contains(".strictMemorySafety()") == true)
        #expect(findings.first?.messageId == "strictMemorySafetyMissing")
    }

    @Test func pluginAndMacroTargetsAreSkipped() throws {
        #expect(Self.flagged(try Self.model([("Generate", "plugin", false), ("Macros", "macro", false)])).isEmpty)
    }
}
