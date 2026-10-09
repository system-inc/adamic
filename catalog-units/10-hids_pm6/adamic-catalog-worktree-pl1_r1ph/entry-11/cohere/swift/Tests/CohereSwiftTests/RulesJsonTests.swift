import Foundation
import Testing

@testable import CohereSwift

/*
 `swift/Rules.json` is the registry as data, for readers that cannot link Swift: the Go generator of the rule
 catalog (#gawdsvx) reads it. One row per rule `RuleRegistry` registers, sorted by name, each carrying what
 the rule declares (`origin`, `upstreamName`) and whether it is a typed rule, which is which registry list
 it sits in rather than anything it declares.

 The rows here are built from the registry itself, never from a grep of the sources: the abbreviation rule
 is built with its vocabulary and names itself through a static, so a grep for `name = "` finds 53 of the 54.
 The registry is read the way `allNames` reads it, with an empty vocabulary.

 The file is generated, never edited by hand. To regenerate it after adding, renaming or redeclaring a rule:

     COHERE_SWIFT_WRITE_RULES=1 swift test --filter RulesJsonTests/regenerate

 then read the diff. Every other run leaves the file alone and fails when it and the registry disagree.
 */
struct RulesJsonTests {
    struct Row: Decodable, Equatable {
        var name: String
        var origin: RuleOrigin
        var upstreamName: String?
        var typeAware: Bool
    }

    static let file = URL(fileURLWithPath: #filePath)
        .deletingLastPathComponent().deletingLastPathComponent().deletingLastPathComponent()
        .appendingPathComponent("Rules.json")

    /* Every registered rule as it declares itself, sorted by name. */
    static func declaredRows() -> [Row] {
        let fileRows = RuleRegistry.fileRules(vocabulary: AbbreviationVocabulary()).map { rule in
            Row(name: rule.name, origin: rule.origin, upstreamName: rule.upstreamName, typeAware: false)
        }
        let typedRows = RuleRegistry.typedRules.map { rule in
            Row(name: rule.name, origin: rule.origin, upstreamName: rule.upstreamName, typeAware: true)
        }
        let packageRows = RuleRegistry.packageRules.map { rule in
            Row(name: rule.name, origin: rule.origin, upstreamName: rule.upstreamName, typeAware: false)
        }
        return (fileRows + typedRows + packageRows).sorted { $0.name < $1.name }
    }

    static func fileRows() throws -> [Row] {
        try JSONDecoder().decode([Row].self, from: Data(contentsOf: file))
    }

    /*
     The file's text: two-space JSON, one key per line in the row's order, as `HouseRuleVerdicts.json` is
     written. A house row has no `upstreamName` key at all, as the brief the Go reader is written against says.
     */
    static func rendered(_ rows: [Row]) throws -> String {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.withoutEscapingSlashes]
        func literal(_ value: some Encodable) throws -> String {
            String(decoding: try encoder.encode(value), as: UTF8.self)
        }
        let objects = try rows.map { row in
            [
                "  {",
                "    \"name\": \(try literal(row.name)),",
                "    \"origin\": \(try literal(row.origin)),",
                try row.upstreamName.map { "    \"upstreamName\": \(try literal($0))," },
                "    \"typeAware\": \(row.typeAware)",
                "  }",
            ]
            .compactMap(\.self)
            .joined(separator: "\n")
        }
        return "[\n" + objects.joined(separator: ",\n") + "\n]\n"
    }

    @Test(.enabled(if: ProcessInfo.processInfo.environment["COHERE_SWIFT_WRITE_RULES"] == "1"))
    func regenerate() throws {
        try Self.rendered(Self.declaredRows()).write(to: Self.file, atomically: true, encoding: .utf8)
    }

    /* The positive control: a file that failed to decode, or decoded to nothing, must not read as agreement below. */
    @Test func theFileDecodesAndHasRows() throws {
        #expect(!(try Self.fileRows()).isEmpty, "Rules.json decoded to no rows")
    }

    /* The rows built here cover exactly what the registry registers, so nothing below checks a partial list. */
    @Test func theDeclaredRowsAreTheRegistry() {
        let declared = Self.declaredRows().map(\.name)
        #expect(declared == RuleRegistry.allNames)
        #expect(Set(declared).count == declared.count, "a rule is registered twice")
    }

    @Test func everyRegisteredRuleHasARowAndEveryRowARule() throws {
        let fileNames = try Self.fileRows().map(\.name)
        let registered = Set(RuleRegistry.allNames)
        #expect(registered.subtracting(fileNames).sorted() == [], "registered but missing from Rules.json")
        #expect(Set(fileNames).subtracting(registered).sorted() == [], "in Rules.json but not registered")
        #expect(fileNames == fileNames.sorted(), "Rules.json is sorted by name")
        #expect(Set(fileNames).count == fileNames.count, "Rules.json names a rule twice")
    }

    @Test func everyRowSaysWhatItsRuleDeclares() throws {
        let declared = Dictionary(uniqueKeysWithValues: Self.declaredRows().map { ($0.name, $0) })
        for row in try Self.fileRows() {
            guard let rule = declared[row.name] else {
                continue
            }
            #expect(row.origin == rule.origin, "\(row.name): origin")
            #expect(row.upstreamName == rule.upstreamName, "\(row.name): upstreamName")
            #expect(row.typeAware == rule.typeAware, "\(row.name): typeAware")
        }
    }

    /* A port names the check it ports, and a house rule names none. */
    @Test func aRuleHasAnUpstreamNameExactlyWhenItIsAPort() {
        for rule in Self.declaredRows() {
            #expect((rule.origin == .house) == (rule.upstreamName == nil), "\(rule.name)")
        }
    }
}
