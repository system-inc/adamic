import CohereSwift
import Foundation
import Testing

/*
 Language mode resolution, from output shaped exactly like SwiftPM 6.4's (measured: the package list
 serializes as `swiftLanguageVersions`, a target's as `{"swiftLanguageMode":{"_0":"5"}}`).
 */
struct PackageModelTests {
    static let root = URL(fileURLWithPath: "/package", isDirectory: true)

    static func describe(targets: [String]) -> Data {
        let entries = targets.map {
            #"{"name":"\#($0)","type":"library","path":"Sources/\#($0)","sources":["\#($0).swift","Notes.md"]}"#
        }
        return Data(#"{"name":"Example","targets":[\#(entries.joined(separator: ","))]}"#.utf8)
    }

    static func dump(tools: String, packageModes: String = "null", fiveTarget: String? = nil) -> Data {
        let settings = fiveTarget.map {
            #"{"name":"\#($0)","settings":[{"kind":{"swiftLanguageMode":{"_0":"5"}},"tool":"swift"}]}"#
        }
        return Data(
            #"{"toolsVersion":{"_version":"\#(tools)"},"swiftLanguageVersions":\#(packageModes),"targets":[\#(settings ?? "")]}"#
                .utf8
        )
    }

    @Test func aSixManifestDefaultsToSix() throws {
        let model = try PackageModel(
            root: Self.root,
            describeJson: Self.describe(targets: ["A"]),
            dumpPackageJson: Self.dump(tools: "6.4.0"),
        )
        #expect(model.targets.map(\.languageMode) == ["6"])
        #expect(
            model.targets.first?.sources.map(\.lastPathComponent) == ["A.swift"],
            "only Swift sources are ours to check",
        )
    }

    @Test func aFiveManifestDefaultsToFive() throws {
        let model = try PackageModel(
            root: Self.root,
            describeJson: Self.describe(targets: ["A"]),
            dumpPackageJson: Self.dump(tools: "5.9.0"),
        )
        #expect(model.targets.map(\.languageMode) == ["5"])
    }

    @Test func theHighestPackageModeWinsOverOrder() throws {
        let model = try PackageModel(
            root: Self.root,
            describeJson: Self.describe(targets: ["A"]),
            dumpPackageJson: Self.dump(tools: "5.9.0", packageModes: #"["6","5"]"#),
        )
        #expect(model.targets.map(\.languageMode) == ["6"])
    }

    @Test func aTargetSettingOverridesThePackage() throws {
        let model = try PackageModel(
            root: Self.root,
            describeJson: Self.describe(targets: ["A", "B"]),
            dumpPackageJson: Self.dump(tools: "6.4.0", fiveTarget: "B"),
        )
        #expect(model.targets.map(\.languageMode) == ["6", "5"])
    }

    @Test func theSwiftSixRuleReportsOnlyTheFiveTarget() throws {
        let model = try PackageModel(
            root: Self.root,
            describeJson: Self.describe(targets: ["A", "B"]),
            dumpPackageJson: Self.dump(tools: "6.4.0", fiveTarget: "B"),
        )
        let findings = ToolchainRequireSwiftSixLanguageMode().findings(in: model, manifest: nil)
        #expect(findings.count == 1)
        #expect(findings.first?.message.hasPrefix("B compiles in Swift 5") == true)
    }

    /* `Vendor/` marks third-party code by path; a local package anywhere else under the root is ours. */
    @Test func onlyPackagesUnderVendorAreVendored() {
        let root = PackageModel(name: "Root", root: URL(fileURLWithPath: "/repo"), toolsVersion: "6.4.0", targets: [])
        let vendored = PackageModel(
            name: "SwiftTerm",
            root: URL(fileURLWithPath: "/repo/Vendor/SwiftTerm"),
            toolsVersion: "6.0.0",
            targets: [],
        )
        let owned = PackageModel(
            name: "VRMKit",
            root: URL(fileURLWithPath: "/repo/Libraries/VRMKit"),
            toolsVersion: "6.0.0",
            targets: [],
        )
        let vendorAboveRoot = PackageModel(
            name: "Root",
            root: URL(fileURLWithPath: "/Vendor/repo"),
            toolsVersion: "6.4.0",
            targets: [],
        )
        let insideThatRoot = PackageModel(
            name: "Child",
            root: URL(fileURLWithPath: "/Vendor/repo/Libraries/Child"),
            toolsVersion: "6.4.0",
            targets: [],
        )
        #expect(root.isVendored(vendored))
        #expect(!root.isVendored(owned))
        #expect(
            !vendorAboveRoot.isVendored(insideThatRoot),
            "a Vendor directory above the root says nothing about what is inside it",
        )
    }

    @Test func versionsCompareAsVersions() {
        #expect(PackageModel.isOlderLanguageMode("5", "6"))
        #expect(PackageModel.isOlderLanguageMode("4.2", "5"))
        #expect(!PackageModel.isOlderLanguageMode("10", "6"))
    }
}
