import Foundation

/*
 Writes records to the front door, one JSON object per line, and keeps the counts the summary must agree
 with.

 The front door refuses a summary whose finding count disagrees with the findings it received, or a run
 missing a phase record. Rather than trust every call site to keep those numbers straight, the writer
 counts what actually went out and builds the summary from that, so the engine cannot contradict itself
 in a way only the front door would catch.
 */
public final class ContractWriter {
    private let output: (Data) -> Void
    private let encoder: JSONEncoder

    public private(set) var findingsWritten = 0
    public private(set) var phasesWritten: [PhaseRecord.Name] = []
    public private(set) var summaryWritten = false

    /* The sink is a closure so tests can capture the stream; the command passes standard output. */
    public init(output: @escaping (Data) -> Void) {
        self.output = output
        self.encoder = JSONEncoder()
        /* Sorted so two runs over the same tree print byte-identical streams, which is what lets a fixture compare them. */
        self.encoder.outputFormatting = [.sortedKeys, .withoutEscapingSlashes]
    }

    public static func standardOutput() -> ContractWriter {
        ContractWriter { data in FileHandle.standardOutput.write(data) }
    }

    public func write(_ record: some Encodable) throws {
        try send(record)
    }

    public func write(_ finding: FindingRecord) throws {
        try send(finding)
        findingsWritten += 1
    }

    public func write(_ phase: PhaseRecord) throws {
        precondition(!phasesWritten.contains(phase.name), "phase \(phase.name.rawValue) was recorded twice")
        try send(phase)
        phasesWritten.append(phase.name)
    }

    private func send(_ record: some Encodable) throws {
        precondition(!summaryWritten, "a record was written after the summary, which the contract makes the last line")
        var line = try encoder.encode(record)
        line.append(UInt8(ascii: "\n"))
        output(line)
    }

    /*
     Ends the run. The finding count comes from what was written, never from the caller, and the exit code
     follows from it: 0 only for a complete run with no findings. Returns the exit code so the command
     exits with exactly what the summary claims.
     */
    public func finish(complete: Bool) throws -> Int32 {
        let missing = PhaseRecord.Name.allCases.filter { !phasesWritten.contains($0) }
        precondition(missing.isEmpty, "phases never recorded: \(missing.map(\.rawValue))")
        let exitCode: Int32 = findingsWritten == 0 && complete ? 0 : 1
        /* Always empty from this engine: every run describes the package, so none finishes with nothing to look at. */
        try send(SummaryRecord(findings: findingsWritten, complete: complete, exitCode: Int(exitCode)))
        summaryWritten = true
        return exitCode
    }
}
