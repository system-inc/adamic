/*
 The part of `swift package dump-package` the engine reads: the tools version, the package's language
 versions, and each target's settings, which is the only place a target's `swiftLanguageMode` appears.

 The spellings are SwiftPM's serialized ones, measured on Swift 6.4 rather than read from the manifest
 API: the package-level list serializes as `swiftLanguageVersions` even when the manifest spells it
 `swiftLanguageModes`, and a target's mode is `{"kind":{"swiftLanguageMode":{"_0":"6"}}}`.
 */
struct DumpPackageOutput: Decodable {
    /* `{"_version": "6.4.0"}` */
    struct ToolsVersion: Decodable {
        var version: String

        enum CodingKeys: String, CodingKey {
            case version = "_version"
        }
    }

    /* One target's name and settings. */
    struct Target: Decodable {
        var name: String
        var settings: [Setting]?
    }

    /* One build setting. Only the kinds the engine reads are decoded; every other kind decodes as empty. */
    struct Setting: Decodable {
        var kind: Kind
    }

    /* The setting's kind, keyed by its name. */
    struct Kind: Decodable {
        var swiftLanguageMode: LanguageMode?
        /* `{"enableUpcomingFeature": {"_0": "ExistentialAny"}}`, measured on Swift 6.4. */
        var enableUpcomingFeature: Feature?
        /* `{"strictMemorySafety": {}}`: present means on. */
        var strictMemorySafety: Empty?
    }

    /* `{"_0": "ExistentialAny"}` */
    struct Feature: Decodable {
        var name: String

        enum CodingKeys: String, CodingKey {
            case name = "_0"
        }
    }

    struct Empty: Decodable {}

    /* `{"_0": "6"}` */
    struct LanguageMode: Decodable {
        var mode: String

        enum CodingKeys: String, CodingKey {
            case mode = "_0"
        }
    }

    var toolsVersion: ToolsVersion
    var swiftLanguageVersions: [String]?
    var targets: [Target]
}
