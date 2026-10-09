/* One finding from any tool, reduced to what the comparison reads: where, and under the tool's own rule id. */
struct RuleFinding: Hashable, Sendable {
    /* Relative to the package root, so the three tools' spellings of one path compare equal. */
    var file: String
    var line: Int
    var rule: String
    /* The cohere finding's message id; empty for an incumbent's. Lets one cohere rule that covers several incumbent rules be compared family by family. */
    var messageId = ""
}
