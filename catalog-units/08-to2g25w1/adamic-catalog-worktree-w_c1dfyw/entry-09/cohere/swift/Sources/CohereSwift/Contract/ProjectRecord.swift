/*
 The package this run checked and how much of it was in reach: the Swift counterpart of TypeScript's
 `graph built in` line. It carries both the population and the narrowing, because a run that checked
 three files and a run that checked the whole package otherwise print the same verdict.
 */
public struct ProjectRecord: Codable, Equatable, Sendable {
    /* One root target, as the package describes it. */
    public struct Target: Codable, Equatable, Sendable {
        public var name: String
        public var kind: String
        public var files: Int
        public var languageMode: String

        public init(name: String, kind: String, files: Int, languageMode: String) {
            self.name = name
            self.kind = kind
            self.files = files
            self.languageMode = languageMode
        }
    }

    /* A file in a root target that this run deliberately did not check, with the reason in words. */
    public struct ExcludedFile: Codable, Equatable, Sendable {
        public var file: String
        public var reason: String

        public init(file: String, reason: String) {
            self.file = file
            self.reason = reason
        }
    }

    public var kind = "project"
    public var root: String
    public var package: String
    public var elapsedMilliseconds: Int
    public var filesInPackage: Int
    public var filesOurs: Int
    public var filesInScope: Int
    public var scopeDescription: String
    public var targets: [Target]
    public var excluded: [ExcludedFile]

    public init(
        root: String,
        package: String,
        elapsedMilliseconds: Int,
        filesInPackage: Int,
        filesOurs: Int,
        filesInScope: Int,
        scopeDescription: String,
        targets: [Target],
        excluded: [ExcludedFile],
    ) {
        self.root = root
        self.package = package
        self.elapsedMilliseconds = elapsedMilliseconds
        self.filesInPackage = filesInPackage
        self.filesOurs = filesOurs
        self.filesInScope = filesInScope
        self.scopeDescription = scopeDescription
        self.targets = targets
        self.excluded = excluded
    }
}
