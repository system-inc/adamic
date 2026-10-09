import CohereSwift
import Foundation
import Testing

/* A flag accepted and ignored reads as a run that did what was asked, so every flag either means something or is refused. */
struct CommandOptionsTests {
    static let here = URL(fileURLWithPath: "/work", isDirectory: true)

    @Test func aBareRunRunsEveryPhaseAndMutates() throws {
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", "/work"],
            workingDirectory: Self.here,
        )
        #expect(options.runFix && options.runTypes && options.runLint && options.mutate)
    }

    @Test func oneDashIsTheSameAsTwo() throws {
        let options = try CommandOptions.parse(["-lint", "-no-fix"], workingDirectory: Self.here)
        #expect(options.runLint && !options.runTypes && !options.runFix && options.noFix)
    }

    @Test(arguments: [
        ["--timing"], ["--explain", "A.swift"], ["--tsconfig", "x"], ["--bogus"], ["--changed"], ["--fix", "--no-fix"],
        ["--contract", "1"], ["--fix-passes", "0"],
    ])
    func refusedCommandLines(arguments: [String]) {
        #expect(throws: CommandOptions.UsageFailure.self) {
            try CommandOptions.parse(arguments, workingDirectory: Self.here)
        }
    }

    @Test func pathsAreCollected() throws {
        let options = try CommandOptions.parse(["Sources/A", "--no-fix", "B.swift"], workingDirectory: Self.here)
        #expect(options.paths == ["Sources/A", "B.swift"])
    }
}
