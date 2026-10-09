import Foundation

/*
 Which of the package's Swift files this run treats as ours, and why each of the others is not.

 A file is ours when a target of the package, or of a local package inside it that is not vendored,
 compiles it, and git does not ignore it. A generated file is ours too: Kirk's convention (2026-10-03,
 shared with TypeScript) names it `*.generated.swift` and gives it every rule and the formatter, exempting
 it only from file-length, by that name. What git ignores is read from the ignore files themselves
 (`IgnoreRules`), with no git process, and it includes untracked files that are not ignored. A new file someone is still writing is exactly the one a gate must see, and TypeScript's program
 already includes it, because a tsconfig reads the disk rather than the index.

 Every file a target compiles and this run does not check is listed with a reason, so the coverage line
 can say what was left out rather than letting a smaller number pass for a complete one.
 */
public struct FileSet: Equatable, Sendable {
    /* A file this run checks, with the target that compiles it. */
    public struct OwnedFile: Equatable, Sendable {
        public var url: URL
        public var targetName: String
        public var targetKind: String
        /* The root of the package whose target compiles it. */
        public var packageRoot: URL?

        public init(url: URL, targetName: String, targetKind: String, packageRoot: URL? = nil) {
            self.packageRoot = packageRoot
            self.url = url
            self.targetName = targetName
            self.targetKind = targetKind
        }
    }

    public var filesInPackage: [URL]
    public var owned: [OwnedFile]
    public var excluded: [ProjectRecord.ExcludedFile]
    /* Empty inside a git repository. Otherwise why every target source was treated as ours. */
    public var note: String

    public init(filesInPackage: [URL], owned: [OwnedFile], excluded: [ProjectRecord.ExcludedFile], note: String) {
        self.filesInPackage = filesInPackage
        self.owned = owned
        self.excluded = excluded
        self.note = note
    }

    /* `environment` is where the global git config and excludes file are found; tests pass their own. */
    public static func build(
        package: PackageModel,
        environment: [String: String] = ProcessInfo.processInfo.environment,
    ) throws -> FileSet {
        /* Nil outside a git repository, which is an answer ("there are no ignore rules to apply") rather than an empty set. */
        let ignoreRules = IgnoreRules.Repository.containing(package.root).map {
            IgnoreRules(repository: $0, environment: environment)
        }
        var filesInPackage: [URL] = []
        var owned: [OwnedFile] = []
        var excluded: [ProjectRecord.ExcludedFile] = []
        var seen = Set<String>()

        for member in package.allPackages {
            let vendored = member.root != package.root && package.isVendored(member)
            for target in member.targets {
                let kind = try Self.isApplication(target) ? "application" : target.kind
                for source in target.sources where seen.insert(source.path).inserted {
                    filesInPackage.append(source)
                    if vendored {
                        excluded.append(.init(file: source.path, reason: "vendored under Vendor/"))
                        continue
                    }
                    /* Resolved, as the repository root is: a root reached through a symlink (`/tmp` is `/private/tmp`) would otherwise sit outside the repository and exclude everything. */
                    if let ignoreRules, !ignoreRules.isVisible(source.resolvingSymlinksInPath().path) {
                        excluded.append(.init(file: source.path, reason: "ignored by git"))
                        continue
                    }
                    owned.append(
                        OwnedFile(url: source, targetName: target.name, targetKind: kind, packageRoot: member.root)
                    )
                }
            }
        }

        /*
         Swift files git would list under the package that no target compiles: standalone scripts, the manifests, an
         app built by an Xcode project, a separate package beside this one. None is checked by this run, and
         each is named, so a clean verdict over the targets cannot read as a clean verdict over the directory.
         Measured on Presence: 20 such files, holding 56 force unwraps the target files do not.
         */
        if let ignoreRules {
            let resolvedRoot = package.root.resolvingSymlinksInPath().path
            let rootPath = resolvedRoot + "/"
            let compiled = Set(filesInPackage.map { $0.resolvingSymlinksInPath().path })
            let visible = Set(ignoreRules.visibleFiles(under: resolvedRoot))
            for path in visible.sorted()
            where path.hasPrefix(rootPath) && path.hasSuffix(".swift") && !compiled.contains(path) {
                let underVendor = path.dropFirst(rootPath.count).split(separator: "/").dropLast().contains("Vendor")
                let reason =
                    underVendor
                    ? "vendored under Vendor/"
                    : path.hasSuffix("/Package.swift") ? "a package manifest" : "no target of this package compiles it"
                excluded.append(.init(file: path, reason: reason))
            }
        }

        let note =
            ignoreRules == nil
            ? "not a git repository, so every file a target compiles was treated as ours"
            : ""
        return FileSet(filesInPackage: filesInPackage, owned: owned, excluded: excluded, note: note)
    }

    /*
     Whether an executable target is an app rather than a command-line tool. SwiftPM spells both
     `executableTarget`, and the rules that care (print, where stdout means something) need to know which.
     An app is a target where any file imports SwiftUI, AppKit or UIKit. Decided per target, not per file:
     an app's model files import only Foundation and are still app code, and a per-file test would call
     their prints tool output.
     */
    static func isApplication(_ target: PackageModel.Target) throws -> Bool {
        guard target.kind == "executable" else { return false }
        for source in target.sources {
            let text = try String(contentsOf: source, encoding: .utf8)
            if text.range(of: #"(?m)^\s*(@\w+\s+)*import\s+(SwiftUI|AppKit|UIKit)\b"#, options: .regularExpression)
                != nil
            {
                return true
            }
        }
        return false
    }
}
