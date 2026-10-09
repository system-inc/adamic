import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for consistency-no-utils-folder, ported from the Go rule's own cases (`consistency_no_utils_folder_test.go`).
 Each path's package root is the directory named `project` on it, as the pipeline would set it. The
 silent cases are the near misses: the spelling the rule asks for, names that contain `utils` without being
 it, a file rather than a folder named `Utils`, and a `utils` above the package.
 */
struct ConsistencyNoUtilsFolderTests {
    static let source = "/* A header comment, so the finding is seen to sit before it. */\nlet value = 1\n"

    static func findings(_ path: String, targetName: String = "Fixture") -> [FindingRecord] {
        let file = ParsedFile(
            url: URL(fileURLWithPath: path),
            targetName: targetName,
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
            packageRoot: packageRoot(of: path),
        )
        let rule = ConsistencyNoUtilsFolder()
        guard rule.applies(to: file) else { return [] }
        return rule.findings(in: file)
    }

    /* The path up to and including its `project` directory, or nil when it has none: an unknown root. */
    static func packageRoot(of path: String) -> URL? {
        let components = URL(fileURLWithPath: path).pathComponents
        guard let index = components.firstIndex(of: "project") else { return nil }
        return URL(fileURLWithPath: NSString.path(withComponents: Array(components[...index])), isDirectory: true)
    }

    static func described(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)-\($0.endLine ?? 0):\($0.endColumn ?? 0) \($0.messageId) \($0.message)" }
    }

    @Test(arguments: [
        (
            "/project/Sources/Fixture/utils/Thing.swift",
            ["1:1-1:1 noUtils Folder name \"utils\" is not allowed. Use \"utilities\" instead."],
        ),
        (
            "/project/Sources/Fixture/_utils/Thing.swift",
            ["1:1-1:1 noUnderscoreUtils Folder name \"_utils\" is not allowed. Use \"_utilities\" instead."],
        ),
        (
            "/project/Sources/Fixture/utils/nested/deep/Thing.swift",
            ["1:1-1:1 noUtils Folder name \"utils\" is not allowed. Use \"utilities\" instead."],
        ),
        /* Swift's capitalized folder names: the same ban, the replacement in the folder's own casing. */
        (
            "/project/Sources/Fixture/Utils/Thing.swift",
            ["1:1-1:1 noUtils Folder name \"Utils\" is not allowed. Use \"Utilities\" instead."],
        ),
        (
            "/project/Tests/Fixture/_Utils/Thing.swift",
            ["1:1-1:1 noUnderscoreUtils Folder name \"_Utils\" is not allowed. Use \"_Utilities\" instead."],
        ),
        /* Both spellings on one path: one finding each, the underscore first, as the Go rule reports them. */
        (
            "/project/Sources/Fixture/_utils/utils/Thing.swift",
            [
                "1:1-1:1 noUnderscoreUtils Folder name \"_utils\" is not allowed. Use \"_utilities\" instead.",
                "1:1-1:1 noUtils Folder name \"utils\" is not allowed. Use \"utilities\" instead.",
            ],
        ),
        /* The same spelling twice is one finding: the file is misplaced once. */
        (
            "/project/Sources/Fixture/utils/More/utils/Thing.swift",
            ["1:1-1:1 noUtils Folder name \"utils\" is not allowed. Use \"utilities\" instead."],
        ),
        /* A directory inside the target that repeats the target's name: folders below it are still judged. */
        (
            "/project/Sources/Fixture/Fixture/Utils/Thing.swift",
            ["1:1-1:1 noUtils Folder name \"Utils\" is not allowed. Use \"Utilities\" instead."],
        ),
        /* A utils folder between the package root and the target's directory is inside the package too. */
        (
            "/project/utils/Sources/Fixture/Thing.swift",
            ["1:1-1:1 noUtils Folder name \"utils\" is not allowed. Use \"utilities\" instead."],
        ),
    ])
    func utilsFoldersAreFound(path: String, expected: [String]) {
        #expect(Self.described(Self.findings(path)) == expected)
    }

    @Test(arguments: [
        /* The spelling the rule asks for, in every casing it suggests. */
        "/project/Sources/Fixture/utilities/Thing.swift",
        "/project/Sources/Fixture/_utilities/Thing.swift",
        "/project/Sources/Fixture/Utilities/Thing.swift",
        "/project/Sources/Fixture/_Utilities/Thing.swift",
        "/project/Sources/Fixture/components/Thing.swift",
        /* Names that contain the abbreviation without being it, which a substring match would flag. */
        "/project/Sources/Fixture/utilsomething/Thing.swift",
        "/project/Sources/Fixture/myutils/Thing.swift",
        "/project/Sources/Fixture/MyUtils/Thing.swift",
        "/project/Sources/Fixture/StringUtils/Thing.swift",
        /* Not the abbreviation the Go rule bans, so not banned here. */
        "/project/Sources/Fixture/Helpers/Thing.swift",
        "/project/Sources/Fixture/Utility/Thing.swift",
        "/project/Sources/Fixture/util/Thing.swift",
        /* A casing no convention produces is left alone rather than guessed at. */
        "/project/Sources/Fixture/UTILS/Thing.swift",
        /* A file named Utils is not a folder named utils. */
        "/project/Sources/Fixture/Utils.swift",
        "/project/Sources/Fixture/utils.swift",
        /* A folder above the package is never the package's to rename. */
        "/Users/someone/utils/project/Sources/Fixture/Thing.swift",
        "/Users/someone/Utils/project/Sources/Fixture/Nested/Thing.swift",
        /* A custom target path inside the package, beneath a utils folder above it. */
        "/Users/someone/utils/project/Code/Thing.swift",
        /* No known package root: not judged, rather than judged against a guess. */
        "/Users/someone/utils/elsewhere/Thing.swift",
    ])
    func nearMissesAreNot(path: String) {
        #expect(Self.findings(path).isEmpty)
    }

    /* A target whose directory is named Utils is itself a utils folder inside the package, and is named first. */
    @Test func aTargetDirectoryNamedUtilsIsJudged() {
        #expect(
            Self.described(Self.findings("/project/Sources/Utils/utils/Thing.swift", targetName: "Utils")) == [
                "1:1-1:1 noUtils Folder name \"Utils\" is not allowed. Use \"Utilities\" instead."
            ]
        )
    }
}
