import CohereSwift
import Foundation
import Testing

/* The upcoming-features rule, from `dump-package` output shaped exactly as SwiftPM 6.4 writes it. */
struct ToolchainRequireUpcomingFeaturesTests {
    static let root = URL(fileURLWithPath: "/package", isDirectory: true)

    static func model(
        _ targets: [(name: String, type: String, features: [String])],
        fileExtension: String = "swift",
    ) throws -> PackageModel {
        let described = targets.map {
            #"{"name":"\#($0.name)","type":"\#($0.type)","path":"Sources/\#($0.name)","sources":["\#($0.name).\#(fileExtension)"]}"#
        }
        let dumped = targets.map { target in
            let settings = target.features.map { #"{"kind":{"enableUpcomingFeature":{"_0":"\#($0)"}},"tool":"swift"}"# }
            return #"{"name":"\#(target.name)","settings":[\#(settings.joined(separator: ","))]}"#
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
        ToolchainRequireUpcomingFeatures().findings(in: model, manifest: nil).map {
            String($0.message.prefix { $0 != " " })
        }
    }

    @Test func aTargetWithBothFeaturesPasses() throws {
        #expect(
            Self.flagged(try Self.model([("Core", "library", ["ExistentialAny", "MemberImportVisibility"])])).isEmpty
        )
    }

    @Test func aTargetMissingAFeatureIsNamedWithWhatToAdd() throws {
        let model = try Self.model([("Core", "library", ["ExistentialAny"]), ("App", "executable", [])])
        let findings = ToolchainRequireUpcomingFeatures().findings(in: model, manifest: nil)
        #expect(findings.map { String($0.message.prefix { $0 != " " }) } == ["App", "Core"])
        #expect(findings.last?.message.contains(#".enableUpcomingFeature("MemberImportVisibility")"#) == true)
        #expect(
            findings.last?.message.contains(#"("ExistentialAny")"#) == false,
            "a feature already enabled must not be asked for",
        )
    }

    /* Settings on a plugin or macro target serve the build tool, not the code being judged. */
    @Test func pluginAndMacroTargetsAreSkipped() throws {
        #expect(Self.flagged(try Self.model([("Generate", "plugin", []), ("Macros", "macro", [])])).isEmpty)
    }

    /* A C or C++ target has no Swift for the features to mean anything to. */
    @Test func aTargetWithNoSwiftIsSkipped() throws {
        #expect(Self.flagged(try Self.model([("MotionCorrection", "library", [])], fileExtension: "cpp")).isEmpty)
    }
}
