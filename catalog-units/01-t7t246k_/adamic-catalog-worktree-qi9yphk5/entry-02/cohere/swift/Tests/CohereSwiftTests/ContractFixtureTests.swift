import CohereSwift
import Foundation
import Testing

/*
 The contract fixtures in `swift/Contract/` are read by this test and by the Go renderer's, so an encoder
 change that would break the front door breaks here first.

 Every record in every fixture is decoded into the engine's own type and encoded back, and the result must
 equal the original as JSON. A field the fixture has and the type lacks, a renamed key, or a wrong value
 spelling fails the comparison rather than being silently dropped by the decoder.
 */
struct ContractFixtureTests {
    static let contractDirectory = URL(fileURLWithPath: #filePath)
        .deletingLastPathComponent()
        .deletingLastPathComponent()
        .deletingLastPathComponent()
        .appendingPathComponent("Contract", isDirectory: true)

    static let fixtureNames = [
        "Clean", "Findings", "TypesBail", "CrashWithoutSummary", "Unreadable", "Unused", "LintOnly",
    ]

    @Test(arguments: fixtureNames)
    func everyRecordRoundTrips(fixture: String) throws {
        let lines = try Self.lines(of: fixture)
        #expect(!lines.isEmpty, "fixture \(fixture) is empty, so this test would check nothing")
        for line in lines {
            let original = try JSONSerialization.jsonObject(with: Data(line.utf8)) as? NSDictionary
            let kind = try #require(original?["kind"] as? String)
            let reencoded = try Self.roundTrip(line, kind: kind)
            let reparsed = try JSONSerialization.jsonObject(with: reencoded) as? NSDictionary
            #expect(reparsed == original, "\(fixture): a \(kind) record did not survive decode and encode unchanged")
        }
    }

    /* The control: a record with a field the contract does not define must fail the comparison, or the test above proves nothing. */
    @Test func anUnknownFieldIsCaught() throws {
        let line = #"{"kind":"summary","findings":0,"complete":true,"exitCode":0,"extra":1}"#
        let original = try JSONSerialization.jsonObject(with: Data(line.utf8)) as? NSDictionary
        let reparsed = try JSONSerialization.jsonObject(with: Self.roundTrip(line, kind: "summary")) as? NSDictionary
        #expect(reparsed != original)
    }

    static func lines(of fixture: String) throws -> [String] {
        let url = contractDirectory.appendingPathComponent("\(fixture).jsonl")
        return try String(contentsOf: url, encoding: .utf8).split(separator: "\n").map(String.init)
    }

    static func roundTrip(_ line: String, kind: String) throws -> Data {
        let data = Data(line.utf8)
        let decoder = JSONDecoder()
        let encoder = JSONEncoder()
        switch kind {
            case "provenance": return try encoder.encode(decoder.decode(ProvenanceRecord.self, from: data))
            case "project": return try encoder.encode(decoder.decode(ProjectRecord.self, from: data))
            case "finding": return try encoder.encode(decoder.decode(FindingRecord.self, from: data))
            case "fix": return try encoder.encode(decoder.decode(FixRecord.self, from: data))
            case "types": return try encoder.encode(decoder.decode(TypesRecord.self, from: data))
            case "lint": return try encoder.encode(decoder.decode(LintRecord.self, from: data))
            case "phase": return try encoder.encode(decoder.decode(PhaseRecord.self, from: data))
            case "rule": return try encoder.encode(decoder.decode(RuleRecord.self, from: data))
            case "summary": return try encoder.encode(decoder.decode(SummaryRecord.self, from: data))
            case "unreadable": return try encoder.encode(decoder.decode(UnreadableRecord.self, from: data))
            case "unused": return try encoder.encode(decoder.decode(UnusedRecord.self, from: data))
            case "unusedCoverage": return try encoder.encode(decoder.decode(UnusedCoverageRecord.self, from: data))
            default:
                Issue.record("a fixture holds a record kind the engine has no type for: \(kind)")
                return Data()
        }
    }
}
