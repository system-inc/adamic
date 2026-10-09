import Foundation

/*
 Which rules run and at what severity, from the `swift` block of `CohereSettings.json`:

     { "swift": { "rules": { "cohere-swift/force-unwrapping": "error" } } }

 A value is a severity (`error`, `warning`, `off`) or an array whose first element is one, the shape the
 TypeScript config uses. Severities and strictness only, never allow lists or ignore names.

 Formatting is not configured here. Swift has one house format built into the engine (`HouseSwiftFormat`),
 so a `format` key, at the top level or in the `swift` block, is refused by name rather than read or
 ignored: a setting that looks chosen and does nothing is worse than none.

 When the package has no config, every house rule runs at `error`. That is not the permissive default the
 TypeScript loader refuses, because every rule runs, and the note this carries says so on every run, so a
 missing config never reads as a configured clean tree. A config that exists and cannot be read is an
 error, never a fallback to the defaults.
 */
public struct RuleConfiguration: Equatable, Sendable {
    /* How strongly a rule's findings count. `off` means the rule runs on no file. */
    public enum Severity: String, Sendable {
        case error
        case warning
        case off
    }

    /* The config named a rule or a severity this binary does not understand. */
    public struct ReadFailure: Error, CustomStringConvertible {
        public var path: String
        public var reason: String

        public var description: String { "reading \(path): \(reason)" }
    }

    public var severities: [String: Severity]
    /* Empty when a config was read. Otherwise the sentence the lint line prints under itself. */
    public var note: String

    public init(severities: [String: Severity], note: String) {
        self.severities = severities
        self.note = note
    }

    /* The severity a rule runs at: what the config says, or `error` for a house rule the config does not mention. */
    public func severity(of ruleName: String) -> Severity {
        severities[ruleName] ?? .error
    }

    public static func load(packageRoot: URL, explicitPath: URL?) throws -> RuleConfiguration {
        let path = explicitPath ?? packageRoot.appendingPathComponent("CohereSettings.json")
        guard FileManager.default.fileExists(atPath: path.path) else {
            if explicitPath != nil {
                throw ReadFailure(
                    path: path.path,
                    reason: "no such file, and a config named by flag is never replaced by the defaults",
                )
            }
            return RuleConfiguration(
                severities: [:],
                note:
                    "config: no CohereSettings.json beside Package.swift, so every house rule ran at error — put one at \(path.path) to change that",
            )
        }
        return try parse(Data(contentsOf: path), path: path.path)
    }

    static func parse(_ data: Data, path: String) throws -> RuleConfiguration {
        guard let document = try JSONSerialization.jsonObject(with: data) as? [String: Any] else {
            throw ReadFailure(path: path, reason: "the top level is not an object")
        }
        if document["format"] != nil {
            throw ReadFailure(path: path, reason: formatKeyRefusal(named: "\"format\""))
        }
        if (document["swift"] as? [String: Any])?["format"] != nil {
            throw ReadFailure(path: path, reason: formatKeyRefusal(named: "the swift block's \"format\""))
        }
        guard let swiftBlock = document["swift"] else {
            return RuleConfiguration(
                severities: [:],
                note: "config: \(path) has no swift block, so every house rule ran at error",
            )
        }
        guard let rules = (swiftBlock as? [String: Any])?["rules"] as? [String: Any] else {
            throw ReadFailure(path: path, reason: "the swift block has no rules object")
        }
        var severities: [String: Severity] = [:]
        for (name, value) in rules {
            let spelled = (value as? String) ?? ((value as? [Any])?.first as? String)
            guard let spelled, let severity = Severity(rawValue: spelled) else {
                throw ReadFailure(
                    path: path,
                    reason: "\(name) has severity \(value), which is not error, warning or off",
                )
            }
            severities[name] = severity
        }
        return RuleConfiguration(severities: severities, note: "")
    }

    /* Why a format key is refused, and what to do instead. */
    static func formatKeyRefusal(named key: String) -> String {
        "\(key) configures formatting, and Swift has one house format built into cohere-swift that no project's settings change; remove the key, and bring any setting it needs to the house format (HouseSwiftFormat.swift)"
    }
}
