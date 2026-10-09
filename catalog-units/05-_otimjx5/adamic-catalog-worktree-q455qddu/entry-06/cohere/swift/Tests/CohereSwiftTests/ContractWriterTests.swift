import CohereSwift
import Foundation
import Testing

/* The writer builds the summary from what it actually wrote, so the engine cannot disagree with itself in a way only the front door would catch. */
struct ContractWriterTests {
    final class Capture {
        var lines: [String] = []
    }

    static func writer(into capture: Capture) -> ContractWriter {
        ContractWriter { data in capture.lines.append(String(decoding: data, as: UTF8.self)) }
    }

    static func allPhases(_ writer: ContractWriter) throws {
        for name in PhaseRecord.Name.allCases {
            try writer.write(PhaseRecord(name: name, outcome: .ran))
        }
    }

    @Test func aCleanCompleteRunExitsZero() throws {
        let capture = Capture()
        let writer = Self.writer(into: capture)
        try Self.allPhases(writer)
        #expect(try writer.finish(complete: true) == 0)
        #expect(capture.lines.last?.contains(#""exitCode":0"#) == true)
    }

    @Test func findingsAreCountedFromWhatWasWritten() throws {
        let capture = Capture()
        let writer = Self.writer(into: capture)
        let finding = FindingRecord(
            source: .rule,
            file: "/a.swift",
            line: 1,
            column: 1,
            severity: .error,
            rule: "r",
            messageId: "m",
            message: "x",
        )
        try writer.write(finding)
        try writer.write(finding)
        try Self.allPhases(writer)
        #expect(try writer.finish(complete: true) == 1)
        #expect(capture.lines.last?.contains(#""findings":2"#) == true)
    }

    @Test func anIncompleteRunWithNoFindingsStillExitsOne() throws {
        let capture = Capture()
        let writer = Self.writer(into: capture)
        try Self.allPhases(writer)
        #expect(try writer.finish(complete: false) == 1)
    }

    @Test func everyLineIsOneJsonObject() throws {
        let capture = Capture()
        let writer = Self.writer(into: capture)
        try writer.write(
            FindingRecord(
                source: .rule,
                file: "/a.swift",
                line: 1,
                column: 1,
                severity: .error,
                rule: "r",
                messageId: "m",
                message: "two\nlines",
            )
        )
        try Self.allPhases(writer)
        _ = try writer.finish(complete: true)
        for line in capture.lines {
            #expect(line.hasSuffix("\n"))
            #expect(line.dropLast().contains("\n") == false, "a record spans more than one line: \(line)")
        }
    }
}
