/*
 The last record of every run. The front door cross-checks it against everything it received (the
 findings it counted, the phases it saw, the exit code the process returned), so an engine that stopped
 early or miscounted cannot end a run that reads as clean.
 */
public struct SummaryRecord: Codable, Equatable, Sendable {
    public var kind = "summary"
    public var findings: Int
    public var complete: Bool
    public var exitCode: Int

    public init(findings: Int, complete: Bool, exitCode: Int) {
        self.findings = findings
        self.complete = complete
        self.exitCode = exitCode
    }
}
