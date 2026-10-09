/*
 What one `--unused` analysis looked at, with the population beside the result, so a report that checked
 nothing never reads like a report that found nothing.

 `filesNotChecked` counts files by why the index could not vouch for them: a file the build has not compiled
 as it stands, a file with `#if` (the index describes one configuration), a file with a reference no module
 claims. `skipped` counts what is never reported by design, such as an `@_exported` import, which is API.
 */
public struct UnusedCoverageRecord: Codable, Equatable, Sendable {
    public var kind = "unusedCoverage"
    public var rule: String
    public var filesChecked: Int
    public var filesNotChecked: [String: Int]
    public var checked: Int
    public var skipped: [String: Int]
    public var found: Int
    public var elapsedMilliseconds: Int

    public init(
        rule: String,
        filesChecked: Int,
        filesNotChecked: [String: Int],
        checked: Int,
        skipped: [String: Int],
        found: Int,
        elapsedMilliseconds: Int,
    ) {
        self.rule = rule
        self.filesChecked = filesChecked
        self.filesNotChecked = filesNotChecked
        self.checked = checked
        self.skipped = skipped
        self.found = found
        self.elapsedMilliseconds = elapsedMilliseconds
    }
}
