/*
 A file the engine could not read, so nothing in it was checked (contract 2).

 Written after `project` and before the first `phase`, one per file. Before contract 2 these were notes on
 stderr and the summary said only that the run fell short: the front door believed it, but could not say
 which files. As a record, each prints as `not checked: <file> (could not be read: <error>)` beside the
 excluded files, and the front door refuses a summary that calls the run complete over one.
 */
public struct UnreadableRecord: Codable, Equatable, Sendable {
    public var kind = "unreadable"
    public var file: String
    /* A sentence, printed after "could not be read:". */
    public var error: String

    public init(file: String, error: String) {
        self.file = file
        self.error = error
    }
}
