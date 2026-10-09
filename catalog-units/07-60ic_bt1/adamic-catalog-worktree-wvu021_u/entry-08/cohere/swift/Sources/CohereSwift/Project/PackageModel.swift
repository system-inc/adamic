import Foundation

/*
 The package as SwiftPM sees it: its root targets, what kind each is, which Swift files each compiles,
 and the language mode each compiles in.

 Read from SwiftPM rather than from a walk of the directory, so a file belongs to the run because a
 target compiles it, never because it happens to sit under `Sources/`.

 Local packages the root depends on by path, and that live inside the root's own directory, are loaded
 too: ahraos-presence's `Libraries/VRMKit` is a fork we own and format as ours, and leaving it out would
 call Presence clean while a fifth of its Swift went unread. Whether such a package is ours or vendored
 is the file set's decision, not this model's. Remote dependencies are never loaded.

 Two commands, because neither holds everything: `describe` gives each target's kind and resolved
 sources, and `dump-package` gives the settings, which is the only place a target's
 `swiftLanguageMode` appears. Both run against the engine's own scratch path, never the project's
 `.build`, so a run never contends on SwiftPM's lock with the developer's build and `--no-fix` writes
 nothing to the project.
 */
public struct PackageModel: Equatable, Sendable {
    /* One root target. */
    public struct Target: Equatable, Sendable {
        public var name: String
        /* SwiftPM's own word: library, executable, test, plugin, macro, snippet. */
        public var kind: String
        public var directory: URL
        public var sources: [URL]
        /* The Swift language mode the target compiles in, as SwiftPM resolves it: "6", "5", "4.2". */
        public var languageMode: String
        /* The upcoming features its settings enable, by name: `ExistentialAny`, `MemberImportVisibility`. */
        public var upcomingFeatures: Set<String>
        /* Whether its settings turn on strict memory safety (SE-0458). */
        public var strictMemorySafety: Bool

        public init(
            name: String,
            kind: String,
            directory: URL,
            sources: [URL],
            languageMode: String,
            upcomingFeatures: Set<String> = [],
            strictMemorySafety: Bool = false,
        ) {
            self.name = name
            self.kind = kind
            self.directory = directory
            self.sources = sources
            self.languageMode = languageMode
            self.upcomingFeatures = upcomingFeatures
            self.strictMemorySafety = strictMemorySafety
        }
    }

    /* SwiftPM could not describe the package, so there is nothing this run can check. */
    public struct DescriptionFailure: Error, CustomStringConvertible {
        public var command: String
        public var standardError: String

        public var description: String {
            "swift package \(command) failed, so the package could not be read and nothing was checked:\n\(standardError)"
        }
    }

    public var name: String
    public var root: URL
    public var toolsVersion: String
    public var targets: [Target]
    /* Paths of the local packages this one depends on that sit inside its directory. */
    public var localDependencyRoots: [URL]
    /* Every local package this one depends on by path, inside its directory or not: what a build reads beyond the root. */
    public var pathDependencyRoots: [URL] = []
    /* Those packages, loaded. Filled by `load`; empty when built from output alone. */
    public var localPackages: [PackageModel]
    /*
     Each single-target product's name, to its target's. The build names an executable's directory for its
     product (`cohere-swift-p.build` for target `CohereSwiftCommand`), so this is how its records are found.
     */
    public var productTargets: [String: String] = [:]
    /* The targets a library product is built from: the package's API to whoever depends on it, so what they declare public is used by someone this build never sees. */
    public var libraryProductTargets: Set<String> = []

    public init(
        name: String,
        root: URL,
        toolsVersion: String,
        targets: [Target],
        localDependencyRoots: [URL] = [],
        localPackages: [PackageModel] = [],
    ) {
        self.name = name
        self.root = root
        self.toolsVersion = toolsVersion
        self.targets = targets
        self.localDependencyRoots = localDependencyRoots
        self.localPackages = localPackages
    }

    /* This package and every local package under it, the root first. */
    public var allPackages: [PackageModel] {
        [self] + localPackages
    }

    /* One package's two answers from SwiftPM, wherever they came from. */
    typealias Answers = (describe: Data, dump: Data)

    /* The cache had no answer for a package the traversal reached, so the run describes cold. */
    private struct CacheMiss: Error {}

    /*
     With a known `toolchain`, answers come from `PackageDescriptionCache` when every fingerprint still holds,
     and a cold description refills it. Without one (the parity harness, or a toolchain `swift --version`
     could not name) every run describes cold, because an entry keyed on an unknown toolchain could survive
     a toolchain change.
     */
    public static func load(
        root: URL,
        scratchPath: URL,
        runner: ProcessRunner = ProcessRunner(),
        toolchain: String? = nil,
    ) throws -> PackageModel {
        let cache = PackageDescriptionCache(scratchPath: scratchPath)
        let cacheable = toolchain.map { !$0.hasPrefix("unknown") } ?? false
        if cacheable, let toolchain, let cached = cache.members(root: root, toolchain: toolchain) {
            do {
                return try traverse(root: root, scratchPath: scratchPath) { packageRoot, _ in
                    guard let member = cached[packageRoot.path] else { throw CacheMiss() }
                    return (member.describe, member.dump)
                }.model
            }
            catch {
                /* A member the cache never held: describe cold below, which also refills it. */
            }
        }
        let cold = try traverse(root: root, scratchPath: scratchPath) { packageRoot, packageScratch in
            (
                try runPackageCommand(
                    ["describe", "--type", "json"],
                    root: packageRoot,
                    scratchPath: packageScratch,
                    runner: runner,
                ),
                try runPackageCommand(
                    ["dump-package"],
                    root: packageRoot,
                    scratchPath: packageScratch,
                    runner: runner,
                ),
            )
        }
        if cacheable, let toolchain {
            cache.store(root: root, toolchain: toolchain, answers: cold.answers)
        }
        return cold.model
    }

    /* The root and every local package inside it, each described by `answer`, which either asks SwiftPM or reads the cache. */
    private static func traverse(
        root: URL,
        scratchPath: URL,
        answer: (URL, URL) throws -> Answers,
    ) throws -> (model: PackageModel, answers: [(root: URL, describe: Data, dump: Data, model: PackageModel)]) {
        var answers: [(root: URL, describe: Data, dump: Data, model: PackageModel)] = []
        func loadOne(_ packageRoot: URL, _ packageScratch: URL) throws -> PackageModel {
            let given = try answer(packageRoot, packageScratch)
            let model = try PackageModel(root: packageRoot, describeJson: given.describe, dumpPackageJson: given.dump)
            answers.append((packageRoot, given.describe, given.dump, model))
            return model
        }
        var seen: Set<String> = [root.resolvingSymlinksInPath().path]
        var model = try loadOne(root, scratchPath)
        var pending = model.localDependencyRoots
        while let next = pending.popLast() {
            guard seen.insert(next.resolvingSymlinksInPath().path).inserted else { continue }
            let dependency = try loadOne(next, scratchPath.appendingPathComponent(next.lastPathComponent))
            /* A local package's own local packages count only while they stay inside the root, the same test the root applied. */
            pending.append(contentsOf: dependency.localDependencyRoots.filter { isInside($0, root) })
            model.localPackages.append(dependency)
        }
        model.localPackages.sort { $0.root.path < $1.root.path }
        return (model, answers)
    }

    /*
     Whether a local package is vendored third-party code rather than ours: it sits under a directory named
     `Vendor`, the convention both proving-ground repositories follow (ahraos-macos's `Vendor/SwiftTerm`).
     Decided by that path rather than by config, so the decision is the same in every run and readable in
     the coverage line. A fork we own, like Presence's `Libraries/VRMKit`, lives outside `Vendor/` for
     exactly that reason.
     */
    public func isVendored(_ local: PackageModel) -> Bool {
        let rootComponents = root.resolvingSymlinksInPath().pathComponents
        let localComponents = local.root.resolvingSymlinksInPath().pathComponents
        return localComponents.dropFirst(rootComponents.count).contains("Vendor")
    }

    static func isInside(_ candidate: URL, _ root: URL) -> Bool {
        let candidatePath = candidate.resolvingSymlinksInPath().path
        let rootPath = root.resolvingSymlinksInPath().path
        return candidatePath.hasPrefix(rootPath + "/")
    }

    /* Built from the two commands' output, separately from running them, so tests can hold real output without a toolchain. */
    public init(root: URL, describeJson: Data, dumpPackageJson: Data) throws {
        let description = try JSONDecoder().decode(DescribeOutput.self, from: describeJson)
        let manifest = try JSONDecoder().decode(DumpPackageOutput.self, from: dumpPackageJson)

        let toolsVersion = manifest.toolsVersion.version
        /* SwiftPM compiles with the highest version in the package's list, not the first, so the list is read the same way. */
        let packageDefault =
            manifest.swiftLanguageVersions?.max(by: PackageModel.isOlderLanguageMode)
            ?? PackageModel.defaultLanguageMode(toolsVersion: toolsVersion)
        var modesByTarget: [String: String] = [:]
        var featuresByTarget: [String: Set<String>] = [:]
        var strictMemorySafetyTargets: Set<String> = []
        for target in manifest.targets {
            for setting in target.settings ?? [] {
                if let mode = setting.kind.swiftLanguageMode?.mode {
                    modesByTarget[target.name] = mode
                }
                if let feature = setting.kind.enableUpcomingFeature?.name {
                    featuresByTarget[target.name, default: []].insert(feature)
                }
                if setting.kind.strictMemorySafety != nil {
                    strictMemorySafetyTargets.insert(target.name)
                }
            }
        }

        self.name = description.name
        self.root = root
        self.toolsVersion = toolsVersion
        self.targets = description.targets.map { target in
            let directory = root.appendingPathComponent(target.path, isDirectory: true)
            let sources = (target.sources ?? [])
                .filter { $0.hasSuffix(".swift") }
                .map { directory.appendingPathComponent($0).standardizedFileURL }
            return Target(
                name: target.name,
                kind: target.type,
                directory: directory,
                sources: sources,
                languageMode: modesByTarget[target.name] ?? packageDefault,
                upcomingFeatures: featuresByTarget[target.name] ?? [],
                strictMemorySafety: strictMemorySafetyTargets.contains(target.name),
            )
        }
        .sorted { $0.name < $1.name }
        self.pathDependencyRoots = (description.dependencies ?? [])
            .filter { $0.type == "fileSystem" }
            .compactMap { $0.path.map { URL(fileURLWithPath: $0, isDirectory: true) } }
        self.localDependencyRoots = pathDependencyRoots.filter { PackageModel.isInside($0, root) }
        self.localPackages = []
        self.productTargets = Dictionary(
            (description.products ?? []).compactMap { product in
                product.targets.count == 1 ? product.targets.first.map { (product.name, $0) } : nil
            },
            uniquingKeysWith: { first, _ in first },
        )
        self.libraryProductTargets = Set(
            (description.products ?? []).filter { $0.type?.name == "library" }.flatMap(\.targets)
        )
    }

    /*
     The mode a target gets when nothing names one: Swift 6 for tools version 6.0 and later, Swift 5 before.
     That is SwiftPM's own rule, restated here because the manifest is silent exactly when it applies.
     */
    static func defaultLanguageMode(toolsVersion: String) -> String {
        let major = Int(toolsVersion.split(separator: ".").first ?? "") ?? 0
        return major >= 6 ? "6" : "5"
    }

    /* Compares "4.2", "5" and "6" as versions rather than as strings, where "10" would sort before "6". */
    public static func isOlderLanguageMode(_ first: String, _ second: String) -> Bool {
        let firstParts = first.split(separator: ".").map { Int($0) ?? 0 }
        let secondParts = second.split(separator: ".").map { Int($0) ?? 0 }
        return firstParts.lexicographicallyPrecedes(secondParts)
    }

    private static func runPackageCommand(
        _ arguments: [String],
        root: URL,
        scratchPath: URL,
        runner: ProcessRunner,
    ) throws -> Data {
        let result = try runner.run("swift", ["package", "--scratch-path", scratchPath.path] + arguments, in: root)
        guard result.succeeded else {
            throw DescriptionFailure(command: arguments.joined(separator: " "), standardError: result.standardError)
        }
        return result.standardOutput
    }
}
