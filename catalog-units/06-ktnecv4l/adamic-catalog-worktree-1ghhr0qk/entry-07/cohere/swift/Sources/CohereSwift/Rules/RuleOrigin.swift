/*
 Where a rule came from: our own idea, or a port of an incumbent tool's check. Every rule declares it,
 with `upstreamName`, the incumbent's own id for a port, spelled as that tool spells it (`force_unwrapping`,
 `AlwaysUseLowerCamelCase`). A house rule has no upstream name, even when the parity harness pairs it with
 an incumbent for comparison: a pairing says what to compare against, not where the rule came from.

 Declared rather than read off the name, because the rule catalog checks every house name against the
 naming scheme, and that check would be circular if the name decided what counts as house. `swift/Rules.json`
 is this declaration as data, one row per registered rule, and `RulesJsonTests` holds the two together.
 */
public enum RuleOrigin: String, Codable, Sendable {
    case house
    case swiftLint = "swiftlint"
    case swiftFormat = "swift-format"
}
