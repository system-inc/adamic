import Foundation

/*
 No folder named `utils` or `_utils` inside the package. The repair is to rename the folder `utilities`
 (or `_utilities`), keeping the casing the folder already has: `Utils` becomes `Utilities`.

 The abbreviation is the smaller half of the problem. A folder named for nothing in particular collects
 everything, and "utilities" is not meaningfully better at that, but the spelling is what the kingdom
 settled on and one spelling searched is worth two spellings guessed. The Swift form of our TypeScript
 `nexus/consistency-no-utils-folder`, with its judgment and its two messages carried over.

 It reads only the path, never the tree, so there is no syntax visitor: the decision is made on the file's
 directory names alone.

 - A folder, never a file. `Utils.swift` is a file name, which the Go rule's separator-on-both-sides test
   also leaves alone.
 - The whole name, never a piece of it. `utilities`, `myutils`, `MyUtils` and `utilsomething` all pass,
   because a rule that fires on the spelling it is asking for is worse than no rule.
 - Two casings of each spelling: `utils` and `_utils`, which the Go rule matches exactly, and `Utils` and
   `_Utils`, the same names in the capitalized form Swift folders take (they are named for modules and
   types). `UTILS` and other casings no convention produces are left alone rather than guessed at.
 - `Helpers` and `Utility` are not banned. The Go rule bans exactly the abbreviation, and this port does
   not widen the ruling it ports.
 - One finding per file per spelling, as the Go rule reports: a file under both `_utils` and `utils` gets
   two, a file under `utils` twice gets one. Each sits at line 1, column 1, zero width, where the Go rule
   puts it (the start of the source file, before any header comment), because the file as a whole is what
   is misplaced.
 - The rule applies to every file and answers with nothing for most of them, rather than declining them,
   so a clean package reads as watched and quiet, not as a rule that never looked.

 Only folders inside the package count, judged from `ParsedFile.packageRoot`, the root of the package whose
 target compiles the file (a local package's own root for its files). The Go rule judges the whole absolute
 path, so a checkout under `~/utils/` would flag every file in it, a finding nobody inside the package can
 act on. Every folder from the root down is judged, the target's own directory included, so a target
 directory named `Utils` is a utils folder like any other. A file whose package root is unknown is not
 judged at all, rather than judged against a guess.
 */
public struct ConsistencyNoUtilsFolder: FileRule {
    public let name = "cohere-swift/consistency-no-utils-folder"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    /* Each banned folder name, mapped to the name to use instead, in the same casing. */
    static let underscoreReplacements = ["_utils": "_utilities", "_Utils": "_Utilities"]
    static let replacements = ["utils": "utilities", "Utils": "Utilities"]

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let folders = Self.foldersInsidePackage(of: file)
        var findings: [FindingRecord] = []
        if let folder = folders.first(where: { Self.underscoreReplacements[$0] != nil }),
            let replacement = Self.underscoreReplacements[folder]
        {
            findings.append(
                finding(
                    in: file,
                    message: RuleMessages.ConsistencyNoUtilsFolder.noUnderscoreUtils(
                        folder: folder,
                        replacement: replacement,
                    ),
                )
            )
        }
        if let folder = folders.first(where: { Self.replacements[$0] != nil }),
            let replacement = Self.replacements[folder]
        {
            findings.append(
                finding(
                    in: file,
                    message: RuleMessages.ConsistencyNoUtilsFolder.noUtils(folder: folder, replacement: replacement),
                )
            )
        }
        return findings
    }

    /* The directories between the package root and the file, outermost first; empty when the root is unknown or the file is not under it. */
    static func foldersInsidePackage(of file: ParsedFile) -> ArraySlice<String> {
        guard let root = file.packageRoot else { return [] }
        let rootComponents = root.resolvingSymlinksInPath().standardizedFileURL.pathComponents
        let directories = file.url.resolvingSymlinksInPath().standardizedFileURL.pathComponents.dropLast()
        guard directories.starts(with: rootComponents) else { return [] }
        return directories.dropFirst(rootComponents.count)
    }

    /* A zero-width finding at the very start of the file, the place the Go rule reports a misplaced file. */
    private func finding(in file: ParsedFile, message: RuleMessages.Message) -> FindingRecord {
        FindingRecord(
            source: .rule,
            file: file.url.path,
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 1,
            severity: .error,
            rule: name,
            messageId: message.id,
            message: message.text,
        )
    }
}
