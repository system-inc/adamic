/*
 One rule this binary implements, written under `--rules` and `--rules-enabled`. A binary that
 implements no rule and a binary whose rules found nothing print the same empty finding list, and this
 is how the outside asks which one it is talking to.
 */
public struct RuleRecord: Codable, Equatable, Sendable {
    public var kind = "rule"
    public var name: String
    public var severity: String

    public init(name: String, severity: String) {
        self.name = name
        self.severity = severity
    }
}
