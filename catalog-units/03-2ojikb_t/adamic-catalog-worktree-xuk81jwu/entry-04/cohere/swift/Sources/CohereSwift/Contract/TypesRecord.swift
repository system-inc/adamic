/*
 What the compiler said about the package, and which files it could not vouch for.

 `filesWithoutRecord` is the part that keeps this honest. The phase reads each file's serialized
 diagnostics rather than the build's output, because an incremental build does not reprint warnings in
 files it did not recompile; a file with no record is therefore a file nobody can say anything about,
 and the run is incomplete while any exist.
 */
public struct TypesRecord: Codable, Equatable, Sendable {
    public var kind = "types"
    public var diagnostics: Int
    public var files: Int
    public var elapsedMilliseconds: Int
    public var filesWithoutRecord: [String]
    public var build: String

    public init(diagnostics: Int, files: Int, elapsedMilliseconds: Int, filesWithoutRecord: [String], build: String) {
        self.diagnostics = diagnostics
        self.files = files
        self.elapsedMilliseconds = elapsedMilliseconds
        self.filesWithoutRecord = filesWithoutRecord
        self.build = build
    }
}
