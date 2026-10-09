/*
 What happened to one phase. Every run writes exactly one of these per phase, in pipeline order,
 including the phases that did not run, because a phase missing from the report reads the same as one
 the reporter forgot, and only a stated outcome separates "found nothing" from "never looked".
 */
public struct PhaseRecord: Codable, Equatable, Sendable {
    /* The pipeline's phases, in the order they run. */
    public enum Name: String, Codable, CaseIterable, Sendable {
        case fix
        case types
        case lint
        case unused
    }

    /* The outcomes `phases.go` defines, spelled the way the contract spells them. */
    public enum Outcome: String, Codable, Sendable {
        case ran
        case skipped
        case notReached
        case reused
    }

    public var kind = "phase"
    public var name: Name
    public var outcome: Outcome
    public var elapsedMilliseconds: Int
    public var detail: String

    public init(name: Name, outcome: Outcome, elapsedMilliseconds: Int = 0, detail: String = "") {
        self.name = name
        self.outcome = outcome
        self.elapsedMilliseconds = elapsedMilliseconds
        self.detail = detail
    }
}
