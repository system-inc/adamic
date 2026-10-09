import Foundation
import Testing

@testable import CohereSwift

/*
 `HouseRuleVerdicts.json` gives every TypeScript house rule (`nexus/`, `structure/`, `base/`) one Swift verdict,
 and the rule catalog (#gawdsvx) is generated from it, so the file must not drift from either registry.
 The TypeScript names are read from the Go sources the front door registers them from, every `Name:` of a
 house rule, so a rule added there without a verdict here fails this test, and so does a verdict for a rule
 that no longer exists.
 */
struct HouseRuleVerdictsTests {
    enum Verdict: String, Decodable {
        case ported = "Ported"
        case port = "Port"
        case compiler = "Compiler"
        case formatter = "Formatter"
        case notApplicable = "NotApplicable"
    }

    struct Row: Decodable {
        var typeScriptRule: String
        var verdict: Verdict
        var swiftRule: String?
        var reason: String?
    }

    static let packageRoot = URL(fileURLWithPath: #filePath)
        .deletingLastPathComponent().deletingLastPathComponent().deletingLastPathComponent()

    /* Decoding refuses a verdict outside the five, which is the unknown-verdict check. */
    static func rows() throws -> [Row] {
        try JSONDecoder().decode(
            [Row].self,
            from: Data(contentsOf: packageRoot.appendingPathComponent("HouseRuleVerdicts.json")),
        )
    }

    /* Every house rule the Go engine registers, by the `Name:` its rule value declares. */
    static func registeredHouseRules() throws -> Set<String> {
        let rules = packageRoot.deletingLastPathComponent().appendingPathComponent("internal/lint/rules")
        let pattern = try Regex(#"Name:\s+"((?:nexus|structure|base)/[a-z0-9-]+)""#)
        var names = Set<String>()
        guard let walker = FileManager.default.enumerator(at: rules, includingPropertiesForKeys: nil) else {
            return names
        }
        for case let url as URL in walker
        where url.pathExtension == "go" && !url.lastPathComponent.hasSuffix("_test.go") {
            for match in try String(contentsOf: url, encoding: .utf8).matches(of: pattern) {
                if let name = match.output[1].substring {
                    names.insert(String(name))
                }
            }
        }
        return names
    }

    @Test func everyHouseRuleHasExactlyOneVerdictAndNothingElse() throws {
        let rows = try Self.rows()
        let names = rows.map(\.typeScriptRule)
        let duplicates = Dictionary(grouping: names, by: \.self).filter { $0.value.count > 1 }.keys.sorted()
        #expect(duplicates.isEmpty, "named twice: \(duplicates)")
        let registered = try Self.registeredHouseRules()
        #expect(!registered.isEmpty, "no house rule was found under internal/lint/rules")
        #expect(registered.subtracting(names).sorted() == [], "registered but given no verdict")
        #expect(Set(names).subtracting(registered).sorted() == [], "given a verdict but not registered")
        #expect(names == names.sorted(), "rows are sorted by typeScriptRule")
    }

    /* A Ported row names the Swift rule that carries it, and that rule exists; every other row says why in a sentence. */
    @Test func eachVerdictCarriesWhatItClaims() throws {
        let swiftRules = Set(RuleRegistry.allNames)
        for row in try Self.rows() {
            if let swiftRule = row.swiftRule {
                #expect(
                    swiftRules.contains(swiftRule),
                    "\(row.typeScriptRule) names \(swiftRule), which cohere-swift does not register",
                )
            }
            if row.verdict == .ported {
                #expect(row.swiftRule != nil, "\(row.typeScriptRule) is Ported but names no Swift rule")
            }
            else {
                #expect(
                    !(row.reason ?? "").trimmingCharacters(in: .whitespaces).isEmpty,
                    "\(row.typeScriptRule) needs a reason",
                )
            }
        }
    }
}
